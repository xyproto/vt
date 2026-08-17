package vt

import (
	"bytes"
	"testing"
)

// rxvt / urxvt (and PuTTY for F1-F4) use encodings that differ from xterm:
// lowercase finals for modified arrows and '$' / '^' finals for modified
// navigation keys.
func TestReadKey_RxvtSequences(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bytes []byte
		want  string
	}{
		{"shift-up", []byte("\x1b[a"), "shift↑"},
		{"shift-down", []byte("\x1b[b"), "shift↓"},
		{"shift-right", []byte("\x1b[c"), "shift→"},
		{"shift-left", []byte("\x1b[d"), "shift←"},
		{"ctrl-up", []byte("\x1bOa"), "ctrl↑"},
		{"ctrl-down", []byte("\x1bOb"), "ctrl↓"},
		{"ctrl-right", []byte("\x1bOc"), "ctrl→"},
		{"ctrl-left", []byte("\x1bOd"), "ctrl←"},
		{"shift-home", []byte("\x1b[7$"), "shift⇱"},
		{"shift-end", []byte("\x1b[8$"), "shift⇲"},
		{"shift-pgup", []byte("\x1b[5$"), "shift⇞"},
		{"shift-pgdn", []byte("\x1b[6$"), "shift⇟"},
		{"shift-delete", []byte("\x1b[3$"), "shift⌦"},
		{"shift-insert", []byte("\x1b[2$"), "shift⎀"},
		{"ctrl-home", []byte("\x1b[7^"), "ctrl⇱"},
		{"ctrl-end", []byte("\x1b[8^"), "ctrl⇲"},
		{"ctrl-pgup", []byte("\x1b[5^"), "ctrl⇞"},
		{"ctrl-pgdn", []byte("\x1b[6^"), "ctrl⇟"},
		{"ctrl-delete", []byte("\x1b[3^"), "ctrl⌦"},
		{"ctrl-insert", []byte("\x1b[2^"), "ctrl⎀"},
		{"F1 rxvt", []byte("\x1b[11~"), "F1"},
		{"F2 rxvt", []byte("\x1b[12~"), "F2"},
		{"F3 rxvt", []byte("\x1b[13~"), "F3"},
		{"F4 rxvt", []byte("\x1b[14~"), "F4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tty := NewTTYFromReader(bytes.NewReader(tc.bytes))
			if k := tty.ReadKey(); k != tc.want {
				t.Errorf("expected %q, got %q", tc.want, k)
			}
		})
	}
}

// The three Insert variants share one glyph, prefixed by the modifier.
func TestReadKey_InsertVariants(t *testing.T) {
	for _, tc := range []struct {
		bytes []byte
		want  string
	}{
		{[]byte("\x1b[2~"), "⎀"},
		{[]byte("\x1b[2;2~"), "shift⎀"},
		{[]byte("\x1b[2;5~"), "ctrl⎀"},
	} {
		tty := NewTTYFromReader(bytes.NewReader(tc.bytes))
		if k := tty.ReadKey(); k != tc.want {
			t.Errorf("%q: expected %q, got %q", tc.bytes, tc.want, k)
		}
	}
	if code := pageNavLookup[[4]byte{27, 91, 50, 126}]; code != KeyInsert {
		t.Errorf("expected KeyInsert, got %d", code)
	}
}

// An unrecognised '$'-terminated rxvt sequence must still be consumed rather
// than leaving parseFirstKey waiting forever for a 0x40-0x7E final byte.
func TestParseFirstKey_UnknownDollarTerminatedSequence(t *testing.T) {
	key, consumed := parseFirstKey([]byte("\x1b[4$x"))
	if consumed != 4 {
		t.Errorf("expected 4 bytes consumed, got %d (%q)", consumed, key)
	}
}

// Every string lookup entry should have a matching key code entry, and vice
// versa, so that ReadKey and Key/KeyCode agree on which sequences are known.
func TestLookupTablesAgree(t *testing.T) {
	for seq := range keyStringLookup {
		if _, ok := keyCodeLookup[seq]; !ok {
			t.Errorf("keyCodeLookup is missing %v", seq)
		}
	}
	for seq := range keyCodeLookup {
		if _, ok := keyStringLookup[seq]; !ok {
			t.Errorf("keyStringLookup is missing %v", seq)
		}
	}
	for seq := range pageStringLookup {
		if _, ok := pageNavLookup[seq]; !ok {
			t.Errorf("pageNavLookup is missing %v", seq)
		}
	}
	for seq := range pageNavLookup {
		if _, ok := pageStringLookup[seq]; !ok {
			t.Errorf("pageStringLookup is missing %v", seq)
		}
	}
	for seq := range fKeyStringLookup {
		if _, ok := fKeyLookup[seq]; !ok {
			t.Errorf("fKeyLookup is missing %v", seq)
		}
	}
	for seq := range fKeyLookup {
		if _, ok := fKeyStringLookup[seq]; !ok {
			t.Errorf("fKeyStringLookup is missing %v", seq)
		}
	}
	for seq := range modKeyStringLookup {
		if _, ok := modKeyLookup[seq]; !ok {
			t.Errorf("modKeyLookup is missing %v", seq)
		}
	}
	for seq := range modKeyLookup {
		if _, ok := modKeyStringLookup[seq]; !ok {
			t.Errorf("modKeyStringLookup is missing %v", seq)
		}
	}
}
