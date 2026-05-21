package vt

import "fmt"

// KittyFlags configures the kitty keyboard protocol enhancement level.
type KittyFlags uint8

// Kitty keyboard protocol flags.
const (
	KittyDisambiguate        KittyFlags = 1 << iota // report disambiguated keys
	KittyReportEvents                               // report key press and release
	KittyReportAlternateKeys                        // report shifted/base codepoints
	KittyReportAllAsCtlSeqs                         // all keys as CSI sequences
	KittyReportText                                 // include text payload
)

// KittyDefault is the recommended set of flags for full key reporting.
const KittyDefault = KittyDisambiguate | KittyReportAlternateKeys | KittyReportAllAsCtlSeqs | KittyReportText

// EnableKittyKeyboard pushes kitty keyboard mode with the given flags onto the stack.
func EnableKittyKeyboard(flags KittyFlags) {
	fmt.Printf("\x1b[>%du", int(flags))
}

// DisableKittyKeyboard pops the top entry from the kitty keyboard mode stack.
func DisableKittyKeyboard() {
	fmt.Print("\x1b[<u")
}
