//go:build windows

package vt

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/xyproto/env/v2"
	"golang.org/x/term"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
)

const (
	STD_INPUT_HANDLE                   = ^uintptr(10) // -11
	STD_OUTPUT_HANDLE                  = ^uintptr(11) // -12
	STD_ERROR_HANDLE                   = ^uintptr(12) // -13
	ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	ENABLE_VIRTUAL_TERMINAL_INPUT      = 0x0200
	DISABLE_NEWLINE_AUTO_RETURN        = 0x0008
	ENABLE_PROCESSED_INPUT             = 0x0001
	ENABLE_LINE_INPUT                  = 0x0002
	ENABLE_ECHO_INPUT                  = 0x0004
	ENABLE_WINDOW_INPUT                = 0x0008
	ENABLE_MOUSE_INPUT                 = 0x0010
	ENABLE_INSERT_MODE                 = 0x0020
	ENABLE_QUICK_EDIT_MODE             = 0x0040
	ENABLE_EXTENDED_FLAGS              = 0x0080
)

var (
	defaultTimeout = 50 * time.Millisecond
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

// Key codes for 6-byte sequences (Ctrl combinations)
var ctrlInsertLookup = map[[6]byte]int{
	{27, 91, 50, 59, 53, 126}: 258, // Ctrl-Insert
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
	{27, 91, 50, 48, 48, 126}: 400, // Paste start
	{27, 91, 50, 48, 49, 126}: 401, // Paste end
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
	fd                 int
	originalState      *term.State
	timeout            time.Duration
	originalInputMode  uint32
	originalOutputMode uint32
	info               *WindowsTerminalInfo
}

// WindowsTerminalInfo holds Windows-specific terminal information
type WindowsTerminalInfo struct {
	IsWindowsTerminal bool
	IsConHost         bool
	SupportsVT        bool
	Version           string
	BuildNumber       int
}

// detectWindowsTerminal detects Windows Terminal
func detectWindowsTerminal() *WindowsTerminalInfo {
	info := &WindowsTerminalInfo{}

	// Check for Windows Terminal
	wtSession := env.Str("WT_SESSION")
	wtProfile := env.Str("WT_PROFILE_ID")
	info.IsWindowsTerminal = wtSession != "" || wtProfile != ""

	// Check for ConHost
	info.IsConHost = !info.IsWindowsTerminal

	// Get terminal version
	if termProgram := env.Str("TERM_PROGRAM"); termProgram != "" {
		info.Version = termProgram
	}

	return info
}

// enableVTMode enables VT100/ANSI processing
func enableVTMode() error {
	// Enable VT processing for stdout
	stdout, _, _ := procGetStdHandle.Call(STD_OUTPUT_HANDLE)
	if stdout == 0 {
		return errors.New("could not get stdout handle")
	}

	var outputMode uint32
	ret, _, _ := procGetConsoleMode.Call(stdout, uintptr(unsafe.Pointer(&outputMode)))
	if ret == 0 {
		return errors.New("could not get console output mode")
	}

	outputMode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING | DISABLE_NEWLINE_AUTO_RETURN
	ret, _, _ = procSetConsoleMode.Call(stdout, uintptr(outputMode))
	if ret == 0 {
		return errors.New("could not set console output mode")
	}

	// Enable VT processing for stdin
	stdin, _, _ := procGetStdHandle.Call(STD_INPUT_HANDLE)
	if stdin == 0 {
		return errors.New("could not get stdin handle")
	}

	var inputMode uint32
	ret, _, _ = procGetConsoleMode.Call(stdin, uintptr(unsafe.Pointer(&inputMode)))
	if ret == 0 {
		return errors.New("could not get console input mode")
	}

	inputMode |= ENABLE_VIRTUAL_TERMINAL_INPUT
	ret, _, _ = procSetConsoleMode.Call(stdin, uintptr(inputMode))
	if ret == 0 {
		return errors.New("could not set console input mode")
	}

	return nil
}

// isVTSupported checks if VT processing is supported
func isVTSupported() bool {
	// Windows 10 TH2 and later support VT processing
	// Test VT mode enablement
	stdout, _, _ := procGetStdHandle.Call(STD_OUTPUT_HANDLE)
	if stdout == 0 {
		return false
	}

	var outputMode uint32
	ret, _, _ := procGetConsoleMode.Call(stdout, uintptr(unsafe.Pointer(&outputMode)))
	if ret == 0 {
		return false
	}

	// Test VT processing flag
	testMode := outputMode | ENABLE_VIRTUAL_TERMINAL_PROCESSING
	ret, _, _ = procSetConsoleMode.Call(stdout, uintptr(testMode))
	if ret == 0 {
		return false
	}

	// Restore original mode
	procSetConsoleMode.Call(stdout, uintptr(outputMode))
	return true
}

// NewTTY opens terminal for input/output on Windows
func NewTTY() (*TTY, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, errors.New("not a terminal")
	}

	// Detect terminal type
	info := detectWindowsTerminal()
	info.SupportsVT = isVTSupported()

	// Enable VT100/ANSI mode for Windows Console
	err := enableVTMode()
	if err != nil && info.SupportsVT {
		// Warn if VT should be supported but failed
		fmt.Fprintf(os.Stderr, "Warning: Could not enable VT100 mode: %v\n", err)
	}

	originalState, err := term.GetState(fd)
	if err != nil {
		return nil, err
	}

	return &TTY{
		fd:            fd,
		originalState: originalState,
		timeout:       defaultTimeout,
		info:          info,
	}, nil
}

// SetTimeout sets a timeout for reading a key
func (tty *TTY) SetTimeout(d time.Duration) {
	tty.timeout = d
}

// Close will restore the terminal state
func (tty *TTY) Close() {
	if tty.originalState != nil {
		term.Restore(tty.fd, tty.originalState)
	}
}

// normalizeKeyCode handles Windows key differences
func normalizeKeyCode(ascii, keyCode int, info *WindowsTerminalInfo) (int, int) {
	if info.IsConHost {
		// Traditional console normalization
		switch ascii {
		case 13: // CR
			return 10, 0 // Normalize to LF
		case 127: // DEL
			return 8, 0 // Normalize to Backspace
		}
	}

	return ascii, keyCode
}

// asciiAndKeyCode processes input into ASCII or key codes, handling multi-byte sequences
func asciiAndKeyCode(tty *TTY) (ascii, keyCode int, err error) {
	bytes := make([]byte, 6) // Use 6 bytes to cover longer sequences like Ctrl-Insert
	var numRead int

	// Set the terminal into raw mode
	rawState, err := term.MakeRaw(tty.fd)
	if err != nil {
		return 0, 0, err
	}
	defer term.Restore(tty.fd, rawState)

	// Read bytes from stdin with timeout
	done := make(chan bool, 1)
	go func() {
		numRead, err = os.Stdin.Read(bytes)
		done <- true
	}()

	select {
	case <-done:
		if err != nil {
			return 0, 0, err
		}
	case <-time.After(tty.timeout):
		return 0, 0, errors.New("read timeout")
	}

	// Handle multi-byte sequences
	switch {
	case numRead == 1:
		ascii = int(bytes[0])
	case numRead == 3:
		seq := [3]byte{bytes[0], bytes[1], bytes[2]}
		if code, found := keyCodeLookup[seq]; found {
			keyCode = code
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
			return
		}
	case numRead == 5:
		// Handle function keys (F1-F12)
		seq := [5]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4]}
		if code, found := functionKeyLookup[seq]; found {
			keyCode = code
			return
		}
	case numRead == 6:
		seq := [6]byte{bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5]}
		// Check Ctrl combinations
		if code, found := ctrlKeyLookup[seq]; found {
			keyCode = code
			return
		}
		// Check Alt combinations
		if code, found := altKeyLookup[seq]; found {
			keyCode = code
			return
		}
		// Check Shift combinations
		if code, found := shiftKeyLookup[seq]; found {
			keyCode = code
			return
		}
		// Check bracketed paste mode
		if code, found := pasteModeLookup[seq]; found {
			keyCode = code
			return
		}
		// Check Ctrl-Insert
		if code, found := ctrlInsertLookup[seq]; found {
			keyCode = code
			return
		}
	default:
		// Attempt to decode as UTF-8
		r, _ := utf8.DecodeRune(bytes[:numRead])
		if unicode.IsPrint(r) {
			ascii = int(r)
		}
	}

	return
}

// Key reads the keycode or ASCII code and avoids repeated keys
func (tty *TTY) Key() int {
	ascii, keyCode, err := asciiAndKeyCode(tty)
	if err != nil {
		return 0
	}

	// Normalize for Windows compatibility
	ascii, keyCode = normalizeKeyCode(ascii, keyCode, tty.info)

	var key int
	if keyCode != 0 {
		key = keyCode
	} else {
		key = ascii
	}
	// Don't filter repeated keys - let the application handle key repeats
	return key
}

// String reads a string, handling key sequences and printable characters
func (tty *TTY) String() string {
	bytes := make([]byte, 6)
	var numRead int
	var err error

	// Set the terminal into raw mode
	rawState, err := term.MakeRaw(tty.fd)
	if err != nil {
		return ""
	}
	defer term.Restore(tty.fd, rawState)

	// Read bytes from stdin with timeout
	done := make(chan bool, 1)
	go func() {
		numRead, err = os.Stdin.Read(bytes)
		done <- true
	}()

	select {
	case <-done:
		if err != nil || numRead == 0 {
			return ""
		}
	case <-time.After(tty.timeout):
		return "" // timeout
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
		// For simplicity, just return what we read
		return string(bytes[:numRead])
	}
	return string(bytes[:numRead])
}

// Rune reads a rune, handling special sequences for arrows, Home, End, etc.
func (tty *TTY) Rune() rune {
	bytes := make([]byte, 6)
	var numRead int
	var err error

	// Set the terminal into raw mode
	rawState, err := term.MakeRaw(tty.fd)
	if err != nil {
		return rune(0)
	}
	defer term.Restore(tty.fd, rawState)

	// Read bytes from stdin with timeout
	done := make(chan bool, 1)
	go func() {
		numRead, err = os.Stdin.Read(bytes)
		done <- true
	}()

	select {
	case <-done:
		if err != nil || numRead == 0 {
			return rune(0)
		}
	case <-time.After(tty.timeout):
		return rune(0) // timeout
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
	_, _ = term.MakeRaw(tty.fd)
}

// NoBlock sets the terminal to cbreak mode (no-op for golang.org/x/term)
func (tty *TTY) NoBlock() {
	// No-op for golang.org/x/term - raw mode handles this
}

// Restore the terminal to its original state
func (tty *TTY) Restore() {
	if tty.originalState != nil {
		term.Restore(tty.fd, tty.originalState)
	}
}

// Flush flushes the terminal output (no-op)
func (tty *TTY) Flush() {
	// No-op for golang.org/x/term
}

// WriteString writes a string to stdout
func (tty *TTY) WriteString(s string) error {
	_, err := os.Stdout.WriteString(s)
	return err
}

// ReadString reads a string from the TTY with timeout
func (tty *TTY) ReadString() (string, error) {
	// Adjust timeout for terminal type
	timeoutDuration := 100 * time.Millisecond
	if tty.info.IsWindowsTerminal {
		// Windows Terminal is faster
		timeoutDuration = 50 * time.Millisecond
	}

	timeout := time.After(timeoutDuration)
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		// Set raw mode temporarily
		_, err := term.MakeRaw(tty.fd)
		if err != nil {
			errorChan <- err
			return
		}
		defer term.Restore(tty.fd, tty.originalState)

		var result []byte
		buffer := make([]byte, 1)

		for {
			n, err := os.Stdin.Read(buffer)
			if err != nil {
				errorChan <- err
				return
			}
			if n > 0 {
				// Look for bell character terminating OSC sequences
				if buffer[0] == 0x07 || buffer[0] == '\a' {
					resultChan <- string(result)
					return
				}
				// Break on ESC sequence end
				if len(result) > 0 && buffer[0] == '\\' && result[len(result)-1] == 0x1b {
					resultChan <- string(result)
					return
				}
				// ST (String Terminator) for OSC
				if len(result) > 0 && buffer[0] == 0x9c {
					resultChan <- string(result)
					return
				}
				result = append(result, buffer[0])

				// Limit response size
				if len(result) > 1024 {
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
		// Timeout - no response from terminal
		return "", nil
	}
}

// PrintRawBytes for debugging raw byte sequences
func (tty *TTY) PrintRawBytes() {
	bytes := make([]byte, 6)
	var numRead int

	// Set the terminal into raw mode
	_, err := term.MakeRaw(tty.fd)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer term.Restore(tty.fd, tty.originalState)

	// Read bytes from stdin
	numRead, err = os.Stdin.Read(bytes)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Raw bytes: %v\n", bytes[:numRead])
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
			// Read character for paste data
			if key < 128 && key >= 32 { // Printable ASCII
				result.WriteByte(byte(key))
			} else if key == 10 || key == 13 { // Newlines
				result.WriteByte('\n')
			}
		}
	}

	return result.String(), nil
}

// IsPasteStart checks if key code indicates paste start
func IsPasteStart(keyCode int) bool {
	return keyCode == PasteStart
}

// IsPasteEnd checks if key code indicates paste end
func IsPasteEnd(keyCode int) bool {
	return keyCode == PasteEnd
}

// IsShiftInsert checks if key code is Shift+Insert
func IsShiftInsert(keyCode int) bool {
	return keyCode == ShiftInsert
}

// GetTerminalInfo returns Windows terminal information
func GetTerminalInfo() *WindowsTerminalInfo {
	return detectWindowsTerminal()
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
