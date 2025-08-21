package vt

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode"

	"github.com/eiannone/keyboard"
)

var (
	defaultTimeout = 50 * time.Millisecond
)

// Map keyboard.Key constants to vt key codes
var keyboardToVTKeyMap = map[keyboard.Key]int{
	keyboard.KeyArrowUp:    253, // Up Arrow
	keyboard.KeyArrowDown:  255, // Down Arrow
	keyboard.KeyArrowRight: 254, // Right Arrow
	keyboard.KeyArrowLeft:  252, // Left Arrow
	keyboard.KeyHome:       1,   // Home (Ctrl-A)
	keyboard.KeyEnd:        5,   // End (Ctrl-E)
	keyboard.KeyPgup:       251, // Page Up
	keyboard.KeyPgdn:       250, // Page Down
	keyboard.KeyInsert:     258, // Insert (mapped to Ctrl-Insert)
}

// Map keyboard.Key constants to Unicode symbols for string representation
var keyboardToStringMap = map[keyboard.Key]string{
	keyboard.KeyArrowUp:    "↑", // Up Arrow
	keyboard.KeyArrowDown:  "↓", // Down Arrow
	keyboard.KeyArrowRight: "→", // Right Arrow
	keyboard.KeyArrowLeft:  "←", // Left Arrow
	keyboard.KeyHome:       "⇱", // Home
	keyboard.KeyEnd:        "⇲", // End
	keyboard.KeyPgup:       "⇞", // Page Up
	keyboard.KeyPgdn:       "⇟", // Page Down
	keyboard.KeyInsert:     "⎘", // Insert (Copy symbol)
}

type TTY struct {
	timeout time.Duration
	open    bool
}

// NewTTY creates a new TTY instance using the keyboard package
func NewTTY() (*TTY, error) {
	return &TTY{
		timeout: defaultTimeout,
		open:    false,
	}, nil
}

// SetTimeout sets a timeout for reading a key
func (tty *TTY) SetTimeout(d time.Duration) {
	tty.timeout = d
}

// Close will close the keyboard
func (tty *TTY) Close() {
	if tty.open {
		keyboard.Close()
		tty.open = false
	}
}

// asciiAndKeyCode processes input using the keyboard package
func asciiAndKeyCode(tty *TTY) (ascii, keyCode int, err error) {
	if !tty.open {
		err = keyboard.Open()
		if err != nil {
			return 0, 0, err
		}
		tty.open = true
		defer func() {
			keyboard.Close()
			tty.open = false
		}()
	}

	// Use GetSingleKey for timeout behavior
	done := make(chan bool, 1)
	var r rune
	var key keyboard.Key

	go func() {
		r, key, err = keyboard.GetKey()
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

	// Map keyboard keys to vt key codes
	if vtKey, found := keyboardToVTKeyMap[key]; found {
		keyCode = vtKey
	} else if r != 0 {
		ascii = int(r)
	} else {
		// Handle special keys that map directly to ASCII values
		ascii = int(key)
	}

	return ascii, keyCode, nil
}

// Key reads the keycode or ASCII code and avoids repeated keys
func (tty *TTY) Key() int {
	ascii, keyCode, err := asciiAndKeyCode(tty)
	if err != nil {
		return 0
	}
	var key int
	if keyCode != 0 {
		key = keyCode
	} else {
		key = ascii
	}
	// Don't filter repeated keys - let the application handle key repeats
	return key
}

// String reads a string using the keyboard package
func (tty *TTY) String() string {
	if !tty.open {
		err := keyboard.Open()
		if err != nil {
			return ""
		}
		tty.open = true
		defer func() {
			keyboard.Close()
			tty.open = false
		}()
	}

	r, key, err := keyboard.GetKey()
	if err != nil {
		return ""
	}

	// Check for special key mappings first
	if str, found := keyboardToStringMap[key]; found {
		return str
	}

	// Return printable rune as string
	if r != 0 && unicode.IsPrint(r) {
		return string(r)
	}

	// Return control character representation
	if key != 0 {
		return "c:" + strconv.Itoa(int(key))
	}

	return ""
}

// Rune reads a rune using the keyboard package
func (tty *TTY) Rune() rune {
	if !tty.open {
		err := keyboard.Open()
		if err != nil {
			return rune(0)
		}
		tty.open = true
		defer func() {
			keyboard.Close()
			tty.open = false
		}()
	}

	r, key, err := keyboard.GetKey()
	if err != nil {
		return rune(0)
	}

	// Check for special key mappings first
	if str, found := keyboardToStringMap[key]; found {
		// Return the first rune of the Unicode symbol
		return []rune(str)[0]
	}

	// Return the actual rune if available
	if r != 0 {
		return r
	}

	// For control characters, return the key code as a rune
	return rune(key)
}

// RawMode is a no-op since keyboard package handles terminal mode
func (tty *TTY) RawMode() {
	// No-op: keyboard package handles terminal mode
}

// NoBlock is a no-op since keyboard package handles non-blocking mode
func (tty *TTY) NoBlock() {
	// No-op: keyboard package handles non-blocking mode
}

// Restore is a no-op since keyboard package handles restoration
func (tty *TTY) Restore() {
	// No-op: keyboard package handles restoration
}

// Flush is a no-op since keyboard package handles flushing
func (tty *TTY) Flush() {
	// No-op: keyboard package handles flushing
}

// WriteString writes a string to stdout
func (tty *TTY) WriteString(s string) error {
	_, err := fmt.Print(s)
	return err
}

// ReadString reads characters until a termination sequence using keyboard package
func (tty *TTY) ReadString() (string, error) {
	if !tty.open {
		err := keyboard.Open()
		if err != nil {
			return "", err
		}
		tty.open = true
		defer func() {
			keyboard.Close()
			tty.open = false
		}()
	}

	timeout := time.After(100 * time.Millisecond)
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		var result []rune
		for {
			r, key, err := keyboard.GetKey()
			if err != nil {
				errorChan <- err
				return
			}

			// For terminal responses, look for bell character or ESC sequences
			if key == 0x07 || r == '\a' {
				resultChan <- string(result)
				return
			}
			// Break on ESC sequence end
			if len(result) > 0 && r == '\\' && result[len(result)-1] == 0x1b {
				resultChan <- string(result)
				return
			}

			if r != 0 {
				result = append(result, r)
			} else if key != 0 {
				result = append(result, rune(key))
			}

			// Prevent infinite reading - limit response size
			if len(result) > 512 {
				resultChan <- string(result)
				return
			}
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	case <-timeout:
		return "", nil
	}
}

// PrintRawBytes for debugging keyboard events
func (tty *TTY) PrintRawBytes() {
	if !tty.open {
		err := keyboard.Open()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		tty.open = true
		defer func() {
			keyboard.Close()
			tty.open = false
		}()
	}

	r, key, err := keyboard.GetKey()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Key event - Rune: %v (%c), Key: %v\n", int(r), r, key)
}

// Term returns nil since we no longer use term.Term
func (tty *TTY) Term() interface{} {
	return nil
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
