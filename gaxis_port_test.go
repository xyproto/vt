package vt

import "testing"

func TestStringWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"", 0},
		{"日本語", 6},     // 3 CJK chars × 2
		{"abc日本", 7},   // 3 + 2×2
		{"café", 4},    // e with combining accent counts as 1 char
		{"a\u0301", 1}, // a + combining acute = 1 column
		{"🎉", 2},       // emoji = 2
		{"\t", 0},      // control chars have width -1, skipped
	}
	for _, tt := range tests {
		got := StringWidth(tt.input)
		if got != tt.want {
			t.Errorf("StringWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestRuneWidth(t *testing.T) {
	if w := RuneWidth('A'); w != 1 {
		t.Errorf("RuneWidth('A') = %d, want 1", w)
	}
	if w := RuneWidth('日'); w != 2 {
		t.Errorf("RuneWidth('日') = %d, want 2", w)
	}
	if w := RuneWidth('\x00'); w != 0 {
		t.Errorf("RuneWidth(0) = %d, want 0", w)
	}
	if w := RuneWidth('\x1b'); w != -1 {
		t.Errorf("RuneWidth(ESC) = %d, want -1", w)
	}
	// Combining mark
	if w := RuneWidth('\u0300'); w != 0 {
		t.Errorf("RuneWidth(U+0300) = %d, want 0", w)
	}
}

func TestIsWideRune(t *testing.T) {
	if !IsWideRune('日') {
		t.Error("expected '日' to be wide")
	}
	if IsWideRune('A') {
		t.Error("expected 'A' to not be wide")
	}
	if !IsWideRune('🎉') {
		t.Error("expected '🎉' to be wide")
	}
}

func TestStringWidthZWJ(t *testing.T) {
	// Family emoji: 👨‍👩‍👧 = man ZWJ woman ZWJ girl
	family := "👨\u200D👩\u200D👧"
	w := StringWidthZWJ(family)
	// Each part is 2-wide, 3 parts = 6
	if w != 6 {
		t.Errorf("StringWidthZWJ(%q) = %d, want 6", family, w)
	}
}

func TestParseMouseSGR(t *testing.T) {
	// Left button press at col 15, row 3
	ev, ok := ParseMouseSGR("0;15;3", 'M')
	if !ok {
		t.Fatal("ParseMouseSGR failed")
	}
	if ev.Button != MouseLeft {
		t.Errorf("button = %d, want MouseLeft", ev.Button)
	}
	if ev.Action != MousePress {
		t.Errorf("action = %d, want MousePress", ev.Action)
	}
	if ev.Col != 14 || ev.Row != 2 { // 0-based
		t.Errorf("pos = (%d,%d), want (14,2)", ev.Col, ev.Row)
	}

	// Release
	ev, ok = ParseMouseSGR("0;10;5", 'm')
	if !ok {
		t.Fatal("ParseMouseSGR release failed")
	}
	if ev.Action != MouseRelease {
		t.Errorf("action = %d, want MouseRelease", ev.Action)
	}

	// Wheel up
	ev, ok = ParseMouseSGR("64;1;1", 'M')
	if !ok {
		t.Fatal("ParseMouseSGR wheel failed")
	}
	if ev.Button != MouseWheelUp {
		t.Errorf("button = %d, want MouseWheelUp(%d)", ev.Button, MouseWheelUp)
	}

	// Invalid
	_, ok = ParseMouseSGR("bad", 'M')
	if ok {
		t.Error("expected ParseMouseSGR to fail on bad input")
	}
}

func TestCursorShape(t *testing.T) {
	// Just verify constants are correct DECSCUSR values
	if CursorDefault != 0 {
		t.Error("CursorDefault should be 0")
	}
	if CursorBlock != 2 {
		t.Error("CursorBlock should be 2")
	}
	if CursorBeam != 6 {
		t.Error("CursorBeam should be 6")
	}
}

func TestBorderStyle(t *testing.T) {
	tl, h, tr, v, br, bl := BorderRounded.Glyphs()
	if tl != '╭' || h != '─' || tr != '╮' || v != '│' || br != '╯' || bl != '╰' {
		t.Error("BorderRounded has wrong glyphs")
	}
}

func TestKittyFlags(t *testing.T) {
	if KittyDefault&KittyDisambiguate == 0 {
		t.Error("KittyDefault should include KittyDisambiguate")
	}
	if KittyDefault&KittyReportText == 0 {
		t.Error("KittyDefault should include KittyReportText")
	}
}
