//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly

package vt

import "time"

// Timeout returns the configured read timeout for the TTY.
func (tty *TTY) Timeout() time.Duration {
	return tty.timeout
}

// CustomString reads a key string like String(), but preserves any pending
// input in the kernel's tty buffer. On unsupported platforms this is a stub.
func (tty *TTY) CustomString() string { return "" }
