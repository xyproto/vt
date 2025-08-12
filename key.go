//go:build !windows

package vt

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/term"
	"github.com/xyproto/env/v2"
)

var (
	defaultTimeout = 2 * time.Millisecond
	lastKey        int
)

// Key codes for 3-byte sequences (arrows, Home, End)
var keyCodeLookup = map[[3]byte]int{
	{27, 91, 65}:  253, // Up Arrow
	{27, 91, 66}:  255, // Down Arrow
	{27, 91, 67}:  254, // Right Arrow
	{27, 91, 68}:  252, // Left Arrow
	{27, 91, 'H'}: 1,   // Home
	{27, 91, 'F'}: 5,   // End
}

// Key codes for 4-byte sequences (Page Up, Page Down, Home, End, Insert, Delete)
var pageNavLookup = map[[4]byte]int{
	{27, 91, 49, 126}: 1,   // Home
	{27, 91, 50, 126}: 259, // Insert
	{27, 91, 51, 126}: 127, // Delete
	{27, 91, 52, 126}: 5,   // End
	{27, 91, 53, 126}: 251, // Page Up
	{27, 91, 54, 126}: 250, // Page Down
}

// Function key codes (5-byte sequences)
var functionKeyLookup = map[[5]byte]int{
	{27, 91, 49, 49, 126}: 260, // F1
	{27, 91, 49, 50, 126}: 261, // F2
	{27, 91, 49, 51, 126}: 262, // F3
	{27, 91, 49, 52, 126}: 263, // F4
	{27, 91, 49, 53, 126}: 264, // F5
	{27, 91, 49, 55, 126}: 265, // F6
	{27, 91, 49, 56, 126}: 266, // F7
	{27, 91, 49, 57, 126}: 267, // F8
	{27, 91, 50, 48, 126}: 268, // F9
	{27, 91, 50, 49, 126}: 269, // F10
	{27, 91, 50, 51, 126}: 270, // F11
	{27, 91, 50, 52, 126}: 271, // F12
}

// Key codes for 6-byte sequences (Ctrl-Insert)
var ctrlInsertLookup = map[[6]byte]int{
	{27, 91, 50, 59, 53, 126}: 258, // Ctrl-Insert (ESC [2;5~)
}

// Ctrl key combinations
var ctrlKeyLookup = map[[6]byte]int{
	{27, 91, 49, 59, 53, 65}: 280, // Ctrl-Up
	{27, 91, 49, 59, 53, 66}: 281, // Ctrl-Down
	{27, 91, 49, 59, 53, 67}: 282, // Ctrl-Right
	{27, 91, 49, 59, 53, 68}: 283, // Ctrl-Left
	{27, 91, 49, 59, 53, 72}: 284, // Ctrl-Home
	{27, 91, 49, 59, 53, 70}: 285, // Ctrl-End
}

// Alt key combinations
var altKeyLookup = map[[6]byte]int{
	{27, 91, 49, 59, 51, 65}: 290, // Alt-Up
	{27, 91, 49, 59, 51, 66}: 291, // Alt-Down
	{27, 91, 49, 59, 51, 67}: 292, // Alt-Right
	{27, 91, 49, 59, 51, 68}: 293, // Alt-Left
}

// Shift key combinations
var shiftKeyLookup = map[[6]byte]int{
	{27, 91, 49, 59, 50, 65}:  300, // Shift-Up
	{27, 91, 49, 59, 50, 66}:  301, // Shift-Down
	{27, 91, 49, 59, 50, 67}:  302, // Shift-Right
	{27, 91, 49, 59, 50, 68}:  303, // Shift-Left
	{27, 91, 50, 59, 50, 126}: 304, // Shift-Insert
}

// Bracketed paste mode sequences
var pasteModeLookup = map[[6]byte]int{
	{27, 91, 50, 48, 48, 126}: 400, // Paste start (ESC [200~)
	{27, 91, 50, 48, 49, 126}: 401, // Paste end (ESC [201~)
}

// Paste operation constants
const (
	PasteStart  = 400
	PasteEnd    = 401
	ShiftInsert = 304
)

// String representations for 3-byte sequences
var keyStringLookup = map[[3]byte]string{
	{27, 91, 65}:  "↑", // Up Arrow
	{27, 91, 66}:  "↓", // Down Arrow
	{27, 91, 67}:  "→", // Right Arrow
	{27, 91, 68}:  "←", // Left Arrow
	{27, 91, 'H'}: "⇱", // Home
	{27, 91, 'F'}: "⇲", // End
}

// String representations for 4-byte sequences
var pageStringLookup = map[[4]byte]string{
	{27, 91, 49, 126}: "⇱", // Home
	{27, 91, 50, 126}: "⎀", // Insert
	{27, 91, 51, 126}: "⌦", // Delete
	{27, 91, 52, 126}: "⇲", // End
	{27, 91, 53, 126}: "⇞", // Page Up
	{27, 91, 54, 126}: "⇟", // Page Down
}

// String representations for 5-byte sequences (function keys)
var functionStringLookup = map[[5]byte]string{
	{27, 91, 49, 49, 126}: "F1",  // F1
	{27, 91, 49, 50, 126}: "F2",  // F2
	{27, 91, 49, 51, 126}: "F3",  // F3
	{27, 91, 49, 52, 126}: "F4",  // F4
	{27, 91, 49, 53, 126}: "F5",  // F5
	{27, 91, 49, 55, 126}: "F6",  // F6
	{27, 91, 49, 56, 126}: "F7",  // F7
	{27, 91, 49, 57, 126}: "F8",  // F8
	{27, 91, 50, 48, 126}: "F9",  // F9
	{27, 91, 50, 49, 126}: "F10", // F10
	{27, 91, 50, 51, 126}: "F11", // F11
	{27, 91, 50, 52, 126}: "F12", // F12
}

// String representations for Ctrl-Insert
var ctrlInsertStringLookup = map[[6]byte]string{
	{27, 91, 50, 59, 53, 126}: "⎘", // Ctrl-Insert
}

// String representations for Ctrl key combinations
var ctrlKeyStringLookup = map[[6]byte]string{
	{27, 91, 49, 59, 53, 65}: "C-↑", // Ctrl-Up
	{27, 91, 49, 59, 53, 66}: "C-↓", // Ctrl-Down
	{27, 91, 49, 59, 53, 67}: "C-→", // Ctrl-Right
	{27, 91, 49, 59, 53, 68}: "C-←", // Ctrl-Left
	{27, 91, 49, 59, 53, 72}: "C-⇱", // Ctrl-Home
	{27, 91, 49, 59, 53, 70}: "C-⇲", // Ctrl-End
}

// String representations for Alt key combinations
var altKeyStringLookup = map[[6]byte]string{
	{27, 91, 49, 59, 51, 65}: "M-↑", // Alt-Up
	{27, 91, 49, 59, 51, 66}: "M-↓", // Alt-Down
	{27, 91, 49, 59, 51, 67}: "M-→", // Alt-Right
	{27, 91, 49, 59, 51, 68}: "M-←", // Alt-Left
}

// String representations for Shift key combinations
var shiftKeyStringLookup = map[[6]byte]string{
	{27, 91, 49, 59, 50, 65}:  "S-↑", // Shift-Up
	{27, 91, 49, 59, 50, 66}:  "S-↓", // Shift-Down
	{27, 91, 49, 59, 50, 67}:  "S-→", // Shift-Right
	{27, 91, 49, 59, 50, 68}:  "S-←", // Shift-Left
	{27, 91, 50, 59, 50, 126}: "S-⎀", // Shift-Insert
}

// String representations for paste mode
var pasteModeStringLookup = map[[6]byte]string{
	{27, 91, 50, 48, 48, 126}: "[PASTE_START]", // Paste start
	{27, 91, 50, 48, 49, 126}: "[PASTE_END]",   // Paste end
}

type TTY struct {
	t       *term.Term
	timeout time.Duration
}

// NewTTY opens a terminal device in raw mode
func NewTTY() (*TTY, error) {
	// Get appropriate TTY path and open in raw mode
	ttyPath := getTTYPath()
	t, err := term.Open(ttyPath, term.RawMode, term.CBreakMode, term.ReadTimeout(defaultTimeout))
	if err != nil {
		return nil, err
	}
	return &TTY{t, defaultTimeout}, nil
}

// getTTYPath returns the appropriate TTY path
func getTTYPath() string {
	// Check for tmux pane TTY
	if tmuxTTY := env.Str("TMUX_PANE_TTY"); tmuxTTY != "" {
		return tmuxTTY
	}

	// Check for SSH TTY
	if sshTTY := env.Str("SSH_TTY"); sshTTY != "" {
		return sshTTY
	}

	// Default to /dev/tty
	defaultTTY := "/dev/tty"
	if _, err := os.Stat(defaultTTY); err == nil {
		return defaultTTY
	}

	// Fallback to stdin if /dev/tty unavailable
	return "/dev/stdin"
}

// normalizeKeyCode handles terminal-specific key differences
func normalizeKeyCode(ascii, keyCode int) (int, int) {
	info := GetTerminalInfo()

	if info.IsPutty {
		switch ascii {
		case 13: // CR
			return 10, 0 // Normalize to LF
		case 127: // DEL
			return 8, 0 // Normalize to Backspace
		}
	}

	if info.InTmux {
		switch ascii {
		case 127: // DEL
			return 8, 0 // Normalize to Backspace
		case 27: // ESC - handle tmux delay
			// In tmux, ESC might be part of escape sequence
			// Return as-is, let timeout handle it
			return ascii, keyCode
		}

		// Handle tmux-specific modified keys
		if keyCode >= 260 && keyCode <= 271 {
			// Function keys in tmux may need adjustment
			return ascii, keyCode
		}
	}

	return ascii, keyCode
}

// SetTimeout sets a timeout for reading a key
func (tty *TTY) SetTimeout(d time.Duration) {
	tty.timeout = d
	tty.t.SetReadTimeout(tty.timeout)
}

// Close will restore and close the raw terminal
func (tty *TTY) Close() {
	tty.t.Restore()
	tty.t.Close()
}

// asciiAndKeyCode processes input into ASCII or key codes, handling multi-byte sequences
func asciiAndKeyCode(tty *TTY) (ascii, keyCode int, err error) {
	bytes := make([]byte, 6) // Use 6 bytes to cover longer sequences like Ctrl-Insert
	var numRead int

	// Set the terminal into raw mode and non-blocking mode with a timeout
	tty.RawMode()
	tty.NoBlock()
	tty.SetTimeout(tty.timeout)
	// Read bytes from the terminal
	numRead, err = tty.t.Read(bytes)

	if err != nil {
		// Restore the terminal settings
		tty.Restore()
		// Clear the key buffer
		tty.t.Flush()
		return
	}

	// Handle multi-byte sequences
	switch {
	case numRead == 1:
		ascii = int(bytes[0])
	case numRead == 3:
		seq := [3]byte{bytes[0], bytes[1], bytes[2]}
		if code, found := keyCodeLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
		// Not found, check if it's a printable character
		r, _ := utf8.DecodeRune(bytes[:numRead])
		if unicode.IsPrint(r) {
			ascii = int(r)
		}
	case numRead == 4:
		seq := [4]byte{bytes[0], bytes[1], bytes[2], bytes[3]}
		if code, found := pageNavLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
	case numRead == 5:
		// Handle function keys (F1-F12)
		seq := [5]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4]}
		if code, found := functionKeyLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
	case numRead == 6:
		seq := [6]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5]}
		// Check Ctrl combinations
		if code, found := ctrlKeyLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
		// Check Alt combinations
		if code, found := altKeyLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
		// Check Shift combinations
		if code, found := shiftKeyLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
		// Check bracketed paste mode
		if code, found := pasteModeLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
		// Check Ctrl-Insert sequences
		if code, found := ctrlInsertLookup[seq]; found {
			keyCode = code
			// Restore the terminal settings
			tty.Restore()
			// Clear the key buffer
			tty.t.Flush()
			return
		}
	default:
		// Attempt to decode as UTF-8
		r, _ := utf8.DecodeRune(bytes[:numRead])
		if unicode.IsPrint(r) {
			ascii = int(r)
		}
	}

	// Restore the terminal settings
	tty.Restore()
	// Clear the key buffer
	tty.t.Flush()
	return
}

// Key reads the keycode or ASCII code and avoids repeated keys
func (tty *TTY) Key() int {
	ascii, keyCode, err := asciiAndKeyCode(tty)
	if err != nil {
		lastKey = 0
		return 0
	}

	// Normalize for terminal compatibility
	ascii, keyCode = normalizeKeyCode(ascii, keyCode)

	var key int
	if keyCode != 0 {
		key = keyCode
	} else {
		key = ascii
	}
	if key == lastKey {
		lastKey = 0
		return 0
	}
	lastKey = key
	return key
}

// String reads a string, handling key sequences and printable characters
func (tty *TTY) String() string {
	bytes := make([]byte, 6)
	var numRead int
	// Set the terminal into raw mode with a timeout
	tty.RawMode()
	tty.SetTimeout(0)
	// Read bytes from the terminal
	numRead, err := tty.t.Read(bytes)
	defer func() {
		// Restore the terminal settings
		tty.Restore()
		tty.t.Flush()
	}()
	if err != nil || numRead == 0 {
		return ""
	}
	switch {
	case numRead == 1:
		r := rune(bytes[0])
		if unicode.IsPrint(r) {
			return string(r)
		}
		return "c:" + strconv.Itoa(int(r))
	case numRead == 3:
		seq := [3]byte{bytes[0], bytes[1], bytes[2]}
		if str, found := keyStringLookup[seq]; found {
			return str
		}
		// Attempt to interpret as UTF-8 string
		return string(bytes[:numRead])
	case numRead == 4:
		seq := [4]byte{bytes[0], bytes[1], bytes[2], bytes[3]}
		if str, found := pageStringLookup[seq]; found {
			return str
		}
		return string(bytes[:numRead])
	case numRead == 5:
		// Handle function keys
		seq := [5]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4]}
		if str, found := functionStringLookup[seq]; found {
			return str
		}
		return string(bytes[:numRead])
	case numRead == 6:
		seq := [6]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5]}
		if str, found := ctrlKeyStringLookup[seq]; found {
			return str
		}
		if str, found := altKeyStringLookup[seq]; found {
			return str
		}
		if str, found := shiftKeyStringLookup[seq]; found {
			return str
		}
		if str, found := pasteModeStringLookup[seq]; found {
			return str
		}
		if str, found := ctrlInsertStringLookup[seq]; found {
			return str
		}
		fallthrough
	default:
		bytesLeftToRead, err := tty.t.Available()
		if err == nil { // success
			bytes2 := make([]byte, bytesLeftToRead)
			numRead2, err := tty.t.Read(bytes2)
			if err != nil { // error
				// Just read the first read bytes
				return string(bytes[:numRead])
			}
			return string(append(bytes[:numRead], bytes2[:numRead2]...))
		}
	}
	return string(bytes[:numRead])
}

// Rune reads a rune, handling special sequences for arrows, Home, End, etc.
func (tty *TTY) Rune() rune {
	bytes := make([]byte, 6)
	var numRead int

	// Set the terminal into raw mode with a timeout
	tty.RawMode()
	tty.SetTimeout(0)
	// Read bytes from the terminal
	numRead, err := tty.t.Read(bytes)
	// Restore the terminal settings
	tty.Restore()
	tty.t.Flush()

	if err != nil || numRead == 0 {
		return rune(0)
	}

	switch {
	case numRead == 1:
		return rune(bytes[0])
	case numRead == 3:
		seq := [3]byte{bytes[0], bytes[1], bytes[2]}
		if str, found := keyStringLookup[seq]; found {
			return []rune(str)[0]
		}
		// Attempt to interpret as UTF-8 rune
		r, _ := utf8.DecodeRune(bytes[:numRead])
		return r
	case numRead == 4:
		seq := [4]byte{bytes[0], bytes[1], bytes[2], bytes[3]}
		if str, found := pageStringLookup[seq]; found {
			return []rune(str)[0]
		}
		r, _ := utf8.DecodeRune(bytes[:numRead])
		return r
	case numRead == 5:
		// Handle function keys
		seq := [5]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4]}
		if str, found := functionStringLookup[seq]; found {
			return []rune(str)[0]
		}
		r, _ := utf8.DecodeRune(bytes[:numRead])
		return r
	case numRead == 6:
		seq := [6]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5]}
		if str, found := ctrlKeyStringLookup[seq]; found {
			return []rune(str)[0]
		}
		if str, found := altKeyStringLookup[seq]; found {
			return []rune(str)[0]
		}
		if str, found := shiftKeyStringLookup[seq]; found {
			return []rune(str)[0]
		}
		if str, found := pasteModeStringLookup[seq]; found {
			return []rune(str)[0]
		}
		if str, found := ctrlInsertStringLookup[seq]; found {
			return []rune(str)[0]
		}
		r, _ := utf8.DecodeRune(bytes[:numRead])
		return r
	default:
		// Attempt to interpret as UTF-8 rune
		r, _ := utf8.DecodeRune(bytes[:numRead])
		return r
	}
}

// RawMode switches the terminal to raw mode
func (tty *TTY) RawMode() {
	term.RawMode(tty.t)
}

// NoBlock sets the terminal to cbreak mode (non-blocking)
func (tty *TTY) NoBlock() {
	tty.t.SetCbreak()
}

// Restore the terminal to its original state
func (tty *TTY) Restore() {
	tty.t.Restore()
}

// Flush flushes the terminal output
func (tty *TTY) Flush() {
	tty.t.Flush()
}

// WriteString writes a string to the terminal
func (tty *TTY) WriteString(s string) error {
	if n, err := tty.t.Write([]byte(s)); err != nil || n == 0 {
		return errors.New("no bytes written to the TTY")
	}
	return nil
}

// ReadString reads a string from the TTY with timeout
func (tty *TTY) ReadString() (string, error) {
	// Set up a timeout channel
	timeout := time.After(100 * time.Millisecond)
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		// Set raw mode temporarily
		tty.RawMode()
		defer tty.Restore()
		defer tty.Flush()

		var result []byte
		buffer := make([]byte, 1)

		for {
			n, err := tty.t.Read(buffer)
			if err != nil {
				errorChan <- err
				return
			}
			if n > 0 {
				// For terminal responses, look for bell character (0x07) which terminates OSC sequences
				if buffer[0] == 0x07 || buffer[0] == '\a' {
					resultChan <- string(result)
					return
				}
				// Also break on ESC sequence end for some terminals
				if len(result) > 0 && buffer[0] == '\\' && result[len(result)-1] == 0x1b {
					resultChan <- string(result)
					return
				}
				result = append(result, buffer[0])

				// Prevent infinite reading - limit response size
				if len(result) > 512 {
					resultChan <- string(result)
					return
				}
			}
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	case <-timeout:
		// Timeout - return empty string (no error, just no response from terminal)
		return "", nil
	}
}

// ReadPasteData reads pasted text between bracketed paste markers
func (tty *TTY) ReadPasteData() (string, error) {
	var result strings.Builder
	inPaste := false

	for {
		key := tty.Key()
		if key == 0 {
			continue
		}

		if IsPasteStart(key) {
			inPaste = true
			continue
		}

		if IsPasteEnd(key) {
			break
		}

		if inPaste {
			// Read the actual character for paste data
			if key < 128 && key >= 32 { // Printable ASCII
				result.WriteByte(byte(key))
			} else if key == 10 || key == 13 { // Newlines
				result.WriteByte('\n')
			}
		}
	}

	return result.String(), nil
}

// PrintRawBytes for debugging raw byte sequences
func (tty *TTY) PrintRawBytes() {
	bytes := make([]byte, 6)
	var numRead int

	// Set the terminal into raw mode with a timeout
	tty.RawMode()
	tty.SetTimeout(0)
	// Read bytes from the terminal
	numRead, err := tty.t.Read(bytes)
	// Restore the terminal settings
	tty.Restore()
	tty.t.Flush()

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Raw bytes: %v\n", bytes[:numRead])
}

// Term will return the underlying term.Term
func (tty *TTY) Term() *term.Term {
	return tty.t
}

// ASCII returns the ASCII code of the key pressed
func (tty *TTY) ASCII() int {
	ascii, _, err := asciiAndKeyCode(tty)
	if err != nil {
		return 0
	}
	return ascii
}

// KeyCode returns the key code of the key pressed
func (tty *TTY) KeyCode() int {
	_, keyCode, err := asciiAndKeyCode(tty)
	if err != nil {
		return 0
	}
	return keyCode
}

// WaitForKey waits for ctrl-c, Return, Esc, Space, or 'q' to be pressed
func WaitForKey() {
	// Get a new TTY and start reading keypresses in a loop
	r, err := NewTTY()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	for {
		switch r.Key() {
		case 3, 13, 27, 32, 113:
			return
		}
	}
}
