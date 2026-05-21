package vt

import (
	"encoding/base64"
	"fmt"
)

// Alt screen escape sequences.
const (
	altScreenEnable  = "\x1b[?1049h"
	altScreenDisable = "\x1b[?1049l"
)

// EnableAltScreen switches to the alternate screen buffer.
func EnableAltScreen() {
	fmt.Print(altScreenEnable)
}

// DisableAltScreen switches back to the main screen buffer.
func DisableAltScreen() {
	fmt.Print(altScreenDisable)
}

// SetTitle sets the terminal window title via OSC 2.
func SetTitle(title string) {
	fmt.Printf("\x1b]2;%s\x1b\\", title)
}

// CopyToClipboard sends text to the system clipboard via OSC 52.
func CopyToClipboard(text string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	fmt.Printf("\x1b]52;c;%s\x1b\\", encoded)
}

// RequestClipboard requests clipboard content from the terminal via OSC 52.
// The response arrives as an OSC 52 sequence in the input stream.
func RequestClipboard() {
	fmt.Print("\x1b]52;c;?\x1b\\")
}

// Notify sends a desktop notification via OSC 9 (iTerm2/ConEmu style).
func Notify(message string) {
	fmt.Printf("\x1b]9;%s\x1b\\", message)
}

// Hyperlink wraps text in an OSC 8 hyperlink.
// Returns the escaped string suitable for printing.
func Hyperlink(uri, text string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", uri, text)
}

// FocusEvents escape sequences.
const (
	focusEventsEnable  = "\x1b[?1004h"
	focusEventsDisable = "\x1b[?1004l"
)

// EnableFocusEvents enables focus-in/focus-out event reporting.
// Focus in: CSI I, Focus out: CSI O.
func EnableFocusEvents() {
	fmt.Print(focusEventsEnable)
}

// DisableFocusEvents disables focus event reporting.
func DisableFocusEvents() {
	fmt.Print(focusEventsDisable)
}
