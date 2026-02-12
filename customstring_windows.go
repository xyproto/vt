//go:build windows

package vt

// CustomString is not yet supported on Windows, since the github.com/pkg/term
// package is not available. Returns an empty string.
func (tty *TTY) CustomString() string { return "" }
