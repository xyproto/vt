//go:build plan9

package vt

// CustomString is not yet supported on Plan 9, since the github.com/pkg/term
// package is not available. Returns an empty string.
func (tty *TTY) CustomString() string { return "" }
