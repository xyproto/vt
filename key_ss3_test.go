package vt

import (
	"bytes"
	"testing"
)

// Terminals in application cursor keys mode (DECCKM) report arrows, Home and
// End with SS3 (ESC O X) instead of CSI (ESC [ X).
func TestReadKey_SS3CursorKeys(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bytes []byte
		want  string
	}{
		{"up", []byte("\x1bOA"), "↑"},
		{"down", []byte("\x1bOB"), "↓"},
		{"right", []byte("\x1bOC"), "→"},
		{"left", []byte("\x1bOD"), "←"},
		{"home", []byte("\x1bOH"), "⇱"},
		{"end", []byte("\x1bOF"), "⇲"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tty := NewTTYFromReader(bytes.NewReader(tc.bytes))
			if k := tty.ReadKey(); k != tc.want {
				t.Errorf("expected %q, got %q", tc.want, k)
			}
		})
	}
}

func TestKeyCodeLookup_SS3CursorKeys(t *testing.T) {
	for _, tc := range []struct {
		seq  [3]byte
		want int
	}{
		{[3]byte{27, 'O', 'A'}, KeyUp},
		{[3]byte{27, 'O', 'B'}, KeyDown},
		{[3]byte{27, 'O', 'C'}, KeyRight},
		{[3]byte{27, 'O', 'D'}, KeyLeft},
		{[3]byte{27, 'O', 'H'}, 1},
		{[3]byte{27, 'O', 'F'}, 5},
	} {
		if got := keyCodeLookup[tc.seq]; got != tc.want {
			t.Errorf("%v: expected %d, got %d", tc.seq, tc.want, got)
		}
	}
}
