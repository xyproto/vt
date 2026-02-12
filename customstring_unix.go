//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package vt

import (
	"strconv"
	"time"
	"unicode"
)

// CustomString reads a key string like String(), but preserves any pending
// input in the kernel's tty buffer. This is critical for operations like
// shift-insert paste, where the key escape sequence is immediately followed
// by paste data — flushing the buffer after reading the key would lose it.
//
// Differences from String():
//   - Saves and restores timeout via SetTimeout instead of Restore()+Flush(),
//     so the terminal stays in raw mode and pending input is preserved.
//   - When the first byte is ESC, a short timed follow-up read collects the
//     rest of the escape sequence, preventing partial-sequence splits.
//   - For unrecognized 6-byte sequences, returns only the bytes read
//     instead of consuming additional available input.
func (tty *TTY) CustomString() string {
	buf := make([]byte, 6)

	// Save timeout before SetTimeout(0) overwrites it
	savedTimeout := tty.timeout

	tty.RawMode()
	tty.SetTimeout(0) // VMIN=1: block until at least 1 byte

	n, err := tty.t.Read(buf)

	// Restore the saved timeout without flushing pending input
	defer tty.SetTimeout(savedTimeout)

	if err != nil || n == 0 {
		return ""
	}

	// If only the ESC byte arrived, do a short timed read to collect
	// the remaining bytes of a multi-byte escape sequence.
	if n == 1 && buf[0] == 27 {
		tty.SetTimeout(50 * time.Millisecond)
		n2, _ := tty.t.Read(buf[1:])
		n += n2
	}

	switch {
	case n == 1:
		r := rune(buf[0])
		if unicode.IsPrint(r) {
			return string(r)
		}
		return "c:" + strconv.Itoa(int(r))

	case n == 3:
		seq := [3]byte{buf[0], buf[1], buf[2]}
		if s, found := keyStringLookup[seq]; found {
			return s
		}
		return string(buf[:n])

	case n == 4:
		seq := [4]byte{buf[0], buf[1], buf[2], buf[3]}
		if s, found := pageStringLookup[seq]; found {
			return s
		}
		return string(buf[:n])

	case n == 6:
		seq := [6]byte{buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]}
		if s, found := ctrlInsertStringLookup[seq]; found {
			return s
		}
		// Return just what was read; do NOT consume additional
		// available bytes, since they may be paste data.
		return string(buf[:n])
	}

	return string(buf[:n])
}
