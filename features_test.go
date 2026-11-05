package vt

import (
	"bytes"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"testing/iotest"
)

func readKeys(t *testing.T, input string, n int) []string {
	t.Helper()
	tty := NewTTYFromReader(bytes.NewReader([]byte(input)))
	keys := make([]string, n)
	for i := range keys {
		keys[i] = tty.ReadKey()
	}
	return keys
}

func TestReadKey_ModifiedNamedKeys(t *testing.T) {
	for _, tc := range []struct {
		seq  string
		want string
	}{
		{"\x1b[1;6A", "ctrlshift↑"},
		{"\x1b[1;4D", "altshift←"},
		{"\x1b[1;7C", "ctrlalt→"},
		{"\x1b[1;9B", "super↓"},
		{"\x1b[1;5P", "ctrlF1"},
		{"\x1b[1;2S", "shiftF4"},
		{"\x1bO5P", "ctrlF1"},
		{"\x1bO2A", "shift↑"},
		{"\x1b[15;5~", "ctrlF5"},
		{"\x1b[24;2~", "shiftF12"},
		{"\x1b[3;3~", "alt⌦"},
		{"\x1b[6;6~", "ctrlshift⇟"},
	} {
		if got := readKeys(t, tc.seq, 1)[0]; got != tc.want {
			t.Errorf("%q: expected %q, got %q", tc.seq, tc.want, got)
		}
	}
}

// The generic parser must agree with the fixed lookup table, so that the key
// strings do not depend on which of them recognizes a sequence
func TestModifiedNamedKeysMatchTheLookupTable(t *testing.T) {
	for seq, want := range modKeyStringLookup {
		s := string(seq[:])
		params := strings.Split(s[2:len(s)-1], ";")
		if got, ok := modifiedNamedKeyString(s[len(s)-1], params); !ok || got != want {
			t.Errorf("%q: expected %q, got %q", s, want, got)
		}
	}
}

func TestReadKey_KittyAltIsEscapeThenKey(t *testing.T) {
	for _, tc := range []struct {
		seq  string
		want []string
	}{
		{"\x1b[97;3uz", []string{"c:27", "a", "z"}},
		{"\x1b[97;7u", []string{"c:27", "c:1"}},
		{"\x1b[27;3;97~", []string{"c:27", "a"}},
		{"\x1b[9;3u", []string{"c:27", "c:9"}},
		{"\x1b[13;3u", []string{KeyAltReturnString, ""}},
	} {
		got := readKeys(t, tc.seq, len(tc.want))
		if strings.Join(got, " ") != strings.Join(tc.want, " ") {
			t.Errorf("%q: expected %q, got %q", tc.seq, tc.want, got)
		}
	}
}

func TestKeyString_KittyAlt(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte("\x1b[97;3u")))
	if got := tty.KeyString(); got != "a" {
		t.Errorf("expected %q, got %q", "a", got)
	}
}

func TestReadPaste(t *testing.T) {
	input := KeyPasteStartString + "hello\r\x1b[Aworld\x1b[20" + KeyPasteEndString + "x"
	for name, r := range map[string]io.Reader{
		"whole":        bytes.NewReader([]byte(input)),
		"byte by byte": iotest.OneByteReader(bytes.NewReader([]byte(input))),
	} {
		t.Run(name, func(t *testing.T) {
			tty := NewTTYFromReader(r)
			if key := tty.ReadKey(); key != KeyPasteStartString {
				t.Fatalf("expected the paste start marker, got %q", key)
			}
			if text := tty.ReadPaste(0); text != "hello\r\x1b[Aworld\x1b[20" {
				t.Errorf("unexpected pasted text %q", text)
			}
			if key := tty.ReadKey(); key != "x" {
				t.Errorf("expected %q after the paste, got %q", "x", key)
			}
		})
	}
}

func TestReadPaste_WithoutEndMarker(t *testing.T) {
	tty := NewTTYFromReader(bytes.NewReader([]byte("abc\x1b[20")))
	if text := tty.ReadPaste(0); text != "abc\x1b[20" {
		t.Errorf("unexpected pasted text %q", text)
	}
}

func TestReadKey_FocusEvents(t *testing.T) {
	got := readKeys(t, KeyFocusInString+KeyFocusOutString+"a", 3)
	want := []string{KeyFocusInString, KeyFocusOutString, "a"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestMouseEvents(t *testing.T) {
	for _, tc := range []struct {
		seq  string
		want MouseEvent
	}{
		{"\x1b[<0;10;5M", MouseEvent{X: 9, Y: 4, Button: MouseLeft}},
		{"\x1b[<0;10;5m", MouseEvent{X: 9, Y: 4, Button: MouseLeft, Release: true}},
		{"\x1b[<2;1;1M", MouseEvent{Button: MouseRight}},
		{"\x1b[<64;3;4M", MouseEvent{X: 2, Y: 3, Button: MouseWheelUp}},
		{"\x1b[<65;3;4M", MouseEvent{X: 2, Y: 3, Button: MouseWheelDown}},
		{"\x1b[<32;7;8M", MouseEvent{X: 6, Y: 7, Button: MouseLeft, Motion: true}},
		{"\x1b[<20;1;1M", MouseEvent{Button: MouseLeft, Shift: true, Ctrl: true}},
		{"\x1b[M !!", MouseEvent{Button: MouseLeft}},
		{"\x1b[M#*+", MouseEvent{X: 9, Y: 10, Button: MouseNone, Release: true}},
	} {
		key := readKeys(t, tc.seq+"q", 2)
		if key[1] != "q" {
			t.Errorf("%q: the mouse report was not read as one key: %q", tc.seq, key)
		}
		got, ok := ParseMouseEvent(key[0])
		if !ok || got != tc.want {
			t.Errorf("%q: expected %+v, got %+v (%v)", tc.seq, tc.want, got, ok)
		}
	}
	for _, key := range []string{"a", "\x1b[A", "\x1b[<0;0;1M", "\x1b[<0;1M"} {
		if _, ok := ParseMouseEvent(key); ok {
			t.Errorf("%q should not parse as a mouse event", key)
		}
	}
}

func TestParseOSCColor(t *testing.T) {
	for _, tc := range []struct {
		reply   string
		r, g, b float64
	}{
		{"\x1b]11;rgb:ffff/8000/0000\x07", 1, 0x8000 / 65535.0, 0},
		{"\x1b]11;rgb:ff/80/00\x1b\\", 1, 0x80 / 255.0, 0},
		{"\x1b]10;rgb:0000/0000/0000\x07\x1b]11;rgb:ffff/ffff/ffff\x07", 1, 1, 1},
	} {
		r, g, b, ok := parseOSCColor(tc.reply, "]11;rgb:")
		if !ok || math.Abs(r-tc.r) > 1e-9 || math.Abs(g-tc.g) > 1e-9 || math.Abs(b-tc.b) > 1e-9 {
			t.Errorf("%q: expected %v %v %v, got %v %v %v (%v)", tc.reply, tc.r, tc.g, tc.b, r, g, b, ok)
		}
	}
}

func TestHasDeviceAttributesReply(t *testing.T) {
	for s, want := range map[string]bool{
		"\x1b[?62;22c":                    true,
		"\x1b]11;rgb:0/0/0\x07\x1b[?1;2c": true,
		"\x1b[?62;22":                     false,
		"\x1b]11;rgb:0/0/0\x07":           false,
	} {
		if got := hasDeviceAttributesReply(s); got != want {
			t.Errorf("%q: expected %v, got %v", s, want, got)
		}
	}
}

// captureDraw returns what draw writes to stdout
func captureDraw(t *testing.T, draw func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	draw()
	os.Stdout = stdout
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestDrawOnlyTheBottomRightCellChanged(t *testing.T) {
	c := NewCanvasWithSize(4, 2)
	captureDraw(t, c.Draw)
	c.WriteRune(3, 1, White, Black, 'Z')
	if out := captureDraw(t, c.Draw); !strings.Contains(out, "Z") {
		t.Errorf("the changed bottom right cell was not drawn: %q", out)
	}
}

func TestRedrawRepaintsUnchangedCells(t *testing.T) {
	c := NewCanvasWithSize(4, 2)
	c.WriteString(0, 0, White, Black, "abcd")
	captureDraw(t, c.Draw)
	if out := captureDraw(t, c.Draw); strings.Contains(out, "abcd") {
		t.Errorf("Draw repainted an unchanged canvas: %q", out)
	}
	if out := captureDraw(t, c.Redraw); !strings.Contains(out, "abcd") {
		t.Errorf("Redraw did not repaint the canvas: %q", out)
	}
	if out := captureDraw(t, c.HideCursorAndRedraw); !strings.Contains(out, "abcd") {
		t.Errorf("HideCursorAndRedraw did not repaint the canvas: %q", out)
	}
}

func TestWriteStringWideRunes(t *testing.T) {
	c := NewCanvasWithSize(6, 1)
	c.WriteString(0, 0, White, Black, "a日本b")
	var sb strings.Builder
	c.Snapshot(&sb)
	if got := strings.Split(sb.String(), "\n")[1]; got != "a日 本 b" {
		t.Errorf("unexpected row %q", got)
	}
	if c.chars[1].cw != 2 || c.chars[2].cw != 1 {
		t.Errorf("the wide rune does not span two cells")
	}
	// A wide rune that does not fit at the right edge is replaced by a space
	c = NewCanvasWithSize(3, 1)
	c.WriteString(0, 0, White, Black, "ab日")
	if r := runeAt(t, c, 2, 0); r != ' ' {
		t.Errorf("expected a space at the right edge, got %q", r)
	}
}

func TestCellCount(t *testing.T) {
	for s, want := range map[string]uint{"": 0, "abc": 3, "日本": 4, "a日b": 4, "ø": 1} {
		if got := CellCount(s); got != want {
			t.Errorf("%q: expected %d, got %d", s, want, got)
		}
	}
}

func TestClearRemovesWideCells(t *testing.T) {
	c := NewCanvasWithSize(4, 1)
	c.WriteString(0, 0, White, Black, "日")
	c.Clear()
	if c.chars[0].cw != 0 || c.chars[1].cw != 0 {
		t.Errorf("cleared cells are still marked as wide")
	}
}

func TestWriteStringOverAHalfOfAWideRune(t *testing.T) {
	c := NewCanvasWithSize(6, 1)
	c.WriteString(0, 0, White, Black, "日本語")
	c.WriteString(1, 0, White, Black, "ab") // starts on the right half of 日, ends on the left half of 本
	var sb strings.Builder
	c.Snapshot(&sb)
	if got := strings.Split(sb.String(), "\n")[1]; got != " ab 語 " {
		t.Errorf("unexpected row %q", got)
	}
	for i, want := range []uint8{0, 0, 0, 0, 2, 1} {
		if c.chars[i].cw != want {
			t.Errorf("cell %d: expected cw %d, got %d", i, want, c.chars[i].cw)
		}
	}
}
