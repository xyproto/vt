package vt

// Key constants
const (
	ctrlA        = 1
	ctrlC        = 3
	ctrlD        = 4
	ctrlE        = 5
	ctrlF        = 6
	ctrlH        = 8
	keyTab       = 9
	ctrlL        = 12
	keyEnter     = 13
	ctrlQ        = 17
	ctrlS        = 19
	keyEsc       = 27
	keyBackspace = 127

	arrowLeft  = 1000
	arrowRight = 1001
	arrowUp    = 1002
	arrowDown  = 1003
	delKey     = 1004
	homeKey    = 1005
	endKey     = 1006
	pageUp     = 1007
	pageDown   = 1008

	// Function keys
	keyF1  = 1009
	keyF2  = 1010
	keyF3  = 1011
	keyF4  = 1012
	keyF5  = 1013
	keyF6  = 1014
	keyF7  = 1015
	keyF8  = 1016
	keyF9  = 1017
	keyF10 = 1018
	keyF11 = 1019
	keyF12 = 1020
)

// Modifiers
const (
	ModNone  = 0
	ModCtrl  = 1 << 0
	ModAlt   = 1 << 1
	ModShift = 1 << 2
)

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key       int    // One of the key constants or a rune value
	Rune      rune   // The actual rune if applicable
	Modifiers int    // Bitmask of modifiers
	Name      string // Human readable name for debugging
}
