package vt

import (
	"bytes"
	"testing"
)

// The Linux console reports F1-F5 as ESC [ [ A .. ESC [ [ E, while xterm-class
// terminals use SS3 for F1-F4 and ESC [ nn ~ for F5 and up.
func TestReadKey_FunctionKeys(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bytes []byte
		want  string
	}{
		{"F1 linux console", []byte("\x1b[[A"), "F1"},
		{"F2 linux console", []byte("\x1b[[B"), "F2"},
		{"F3 linux console", []byte("\x1b[[C"), "F3"},
		{"F4 linux console", []byte("\x1b[[D"), "F4"},
		{"F5 linux console", []byte("\x1b[[E"), "F5"},
		{"F1 SS3", []byte("\x1bOP"), "F1"},
		{"F4 SS3", []byte("\x1bOS"), "F4"},
		{"F5 xterm", []byte("\x1b[15~"), "F5"},
		{"F6 xterm", []byte("\x1b[17~"), "F6"},
		{"F12 xterm", []byte("\x1b[24~"), "F12"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tty := NewTTYFromReader(bytes.NewReader(tc.bytes))
			if k := tty.ReadKey(); k != tc.want {
				t.Errorf("expected %q, got %q", tc.want, k)
			}
		})
	}
}

// A Linux console F-key that arrives split across two reads must not be
// classified before its final byte shows up.
func TestParseFirstKey_IncompleteLinuxConsoleFKey(t *testing.T) {
	if key, consumed := parseFirstKey([]byte("\x1b[[")); consumed != 0 {
		t.Errorf("expected the sequence to be incomplete, got %q after %d bytes", key, consumed)
	}
	if key, consumed := parseFirstKey([]byte("\x1b[[A")); key != "F1" || consumed != 4 {
		t.Errorf("expected (F1, 4), got (%q, %d)", key, consumed)
	}
}
