//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package vt

import (
	"bytes"
	"testing"
)

func TestKeyString(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bytes []byte
		want  string
	}{
		{"up", []byte("\x1b[A"), "↑"},
		{"ss3 up", []byte("\x1bOA"), "↑"},
		{"ss3 down", []byte("\x1bOB"), "↓"},
		{"ss3 right", []byte("\x1bOC"), "→"},
		{"ss3 left", []byte("\x1bOD"), "←"},
		{"ss3 home", []byte("\x1bOH"), "⇱"},
		{"ss3 end", []byte("\x1bOF"), "⇲"},
		{"ss3 F1", []byte("\x1bOP"), "F1"},
		{"page up", []byte("\x1b[5~"), "⇞"},
		{"F5", []byte("\x1b[15~"), "F5"},
		{"F1 linux console", []byte("\x1b[[A"), "F1"},
		{"ctrl right", []byte("\x1b[1;5C"), "ctrl→"},
		{"bare esc", []byte("\x1b"), "c:27"},
		{"letter", []byte("a"), "a"},
		{"utf-8", []byte("ø"), "ø"},
		{"return", []byte("\r"), "c:13"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tty := NewTTYFromReader(bytes.NewReader(tc.bytes))
			if k := tty.KeyString(); k != tc.want {
				t.Errorf("expected %q, got %q", tc.want, k)
			}
		})
	}
}

func TestKeyString_DoesNotReadAhead(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte("\x1bOA\x1b[Bx")))
	for _, want := range []string{"↑", "↓", "x"} {
		if k := tty.KeyString(); k != want {
			t.Errorf("expected %q, got %q", want, k)
		}
	}
}
