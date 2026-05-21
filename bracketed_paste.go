package vt

import "fmt"

// Bracketed paste mode escape sequences.
const (
	bracketedPasteEnable  = "\x1b[?2004h"
	bracketedPasteDisable = "\x1b[?2004l"
)

// Bracketed paste delimiters (received from terminal).
const (
	PasteStart = "\x1b[200~"
	PasteEnd   = "\x1b[201~"
)

// EnableBracketedPaste enables bracketed paste mode.
// When enabled, pasted text is wrapped in PasteStart/PasteEnd sequences,
// allowing the application to distinguish pastes from typed input.
func EnableBracketedPaste() {
	fmt.Print(bracketedPasteEnable)
}

// DisableBracketedPaste disables bracketed paste mode.
func DisableBracketedPaste() {
	fmt.Print(bracketedPasteDisable)
}
