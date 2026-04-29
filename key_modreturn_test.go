package vt

import (
	"bytes"
	"testing"
)

func TestReadKey_AltReturn_ESC_CR(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte{0x1B, 0x0D}))
	if k := tty.ReadKey(); k != KeyAltReturnString {
		t.Errorf("expected %q, got %q", KeyAltReturnString, k)
	}
}

func TestReadKey_AltReturn_ESC_LF(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte{0x1B, 0x0A}))
	if k := tty.ReadKey(); k != KeyAltReturnString {
		t.Errorf("expected %q, got %q", KeyAltReturnString, k)
	}
}

func TestReadKey_ShiftReturn_Kitty(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte("\x1b[13;2u")))
	if k := tty.ReadKey(); k != KeyShiftReturnString {
		t.Errorf("expected %q, got %q", KeyShiftReturnString, k)
	}
}

func TestReadKey_ShiftReturn_Xterm(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte("\x1b[27;2;13~")))
	if k := tty.ReadKey(); k != KeyShiftReturnString {
		t.Errorf("expected %q, got %q", KeyShiftReturnString, k)
	}
}

func TestReadKey_AltReturn_Kitty(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte("\x1b[13;3u")))
	if k := tty.ReadKey(); k != KeyAltReturnString {
		t.Errorf("expected %q, got %q", KeyAltReturnString, k)
	}
}

// Lone Escape must still be reported as "c:27".
func TestReadKey_LoneEscape(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte{0x1B}))
	if k := tty.ReadKey(); k != "c:27" {
		t.Errorf("expected c:27, got %q", k)
	}
}

// Existing 6-byte modifier sequences must still parse after the long-CSI
// lookup was added.
func TestReadKey_ShiftUp_Regression(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte{27, 91, 49, 59, 50, 65}))
	if k := tty.ReadKey(); k != "shift↑" {
		t.Errorf("expected shift↑, got %q", k)
	}
}

func TestReadKey_PlainCR(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte{0x0D}))
	if k := tty.ReadKey(); k != "c:13" {
		t.Errorf("expected c:13, got %q", k)
	}
}

// ESC followed by a printable letter is two distinct keystrokes (Escape,
// then the letter) and must not be merged.
func TestReadKey_EscapeThenLetter(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte{0x1B, 'a'}))
	if k := tty.ReadKey(); k != "c:27" {
		t.Fatalf("first key: expected c:27, got %q", k)
	}
	if k := tty.ReadKey(); k != "a" {
		t.Errorf("second key: expected a, got %q", k)
	}
}
