package vt

import (
	"math"
	"testing"
)

func TestIsTrueColor(t *testing.T) {
	if !IsTrueColor(TrueColor(255, 128, 0)) {
		t.Error("TrueColor should report IsTrueColor=true")
	}
	if !IsTrueColor(TrueBackground(0, 128, 255)) {
		t.Error("TrueBackground should report IsTrueColor=true")
	}
	if IsTrueColor(Color256(42)) {
		t.Error("Color256 should not report IsTrueColor=true")
	}
	if IsTrueColor(Red) {
		t.Error("Red should not report IsTrueColor=true")
	}
}

func TestIs256Color(t *testing.T) {
	if !Is256Color(Color256(200)) {
		t.Error("Color256 should report Is256Color=true")
	}
	if !Is256Color(Background256(50)) {
		t.Error("Background256 should report Is256Color=true")
	}
	if Is256Color(TrueColor(1, 2, 3)) {
		t.Error("TrueColor should not report Is256Color=true")
	}
	if Is256Color(Blue) {
		t.Error("Blue should not report Is256Color=true")
	}
}

func TestToRGB(t *testing.T) {
	tests := []struct {
		color       AttributeColor
		r, g, b     uint8
		ok          bool
		description string
	}{
		{TrueColor(100, 150, 200), 100, 150, 200, true, "TrueColor"},
		{TrueBackground(10, 20, 30), 10, 20, 30, true, "TrueBackground"},
		{Color256(196), 255, 0, 0, true, "Color256(196) = bright red in cube"},
		{Red, 205, 0, 0, true, "ANSI Red"},
		{White, 255, 255, 255, true, "ANSI White (97)"},
		{Default, 0, 0, 0, false, "Default has no RGB"},
		{Bright, 0, 0, 0, false, "Bright attribute has no RGB"},
	}
	for _, tc := range tests {
		r, g, b, ok := ToRGB(tc.color)
		if ok != tc.ok {
			t.Errorf("%s: ok=%v, want %v", tc.description, ok, tc.ok)
			continue
		}
		if ok && (r != tc.r || g != tc.g || b != tc.b) {
			t.Errorf("%s: got (%d,%d,%d), want (%d,%d,%d)", tc.description, r, g, b, tc.r, tc.g, tc.b)
		}
	}
}

func TestToHex(t *testing.T) {
	if got := ToHex(TrueColor(255, 0, 127)); got != "#ff007f" {
		t.Errorf("ToHex TrueColor: got %q, want #ff007f", got)
	}
	if got := ToHex(Red); got != "#cd0000" {
		t.Errorf("ToHex Red: got %q, want #cd0000", got)
	}
}

func TestLightenDarken(t *testing.T) {
	base := TrueColor(100, 100, 100)

	lighter := Lighten(base, 0.5)
	lr, lg, lb, _ := ToRGB(lighter)
	if lr <= 100 || lg <= 100 || lb <= 100 {
		t.Errorf("Lighten: expected brighter, got (%d,%d,%d)", lr, lg, lb)
	}

	darker := Darken(base, 0.5)
	dr, dg, db, _ := ToRGB(darker)
	if dr >= 100 || dg >= 100 || db >= 100 {
		t.Errorf("Darken: expected darker, got (%d,%d,%d)", dr, dg, db)
	}

	// Lighten to full white
	white := Lighten(base, 1.0)
	wr, wg, wb, _ := ToRGB(white)
	if wr != 255 || wg != 255 || wb != 255 {
		t.Errorf("Lighten(1.0): expected (255,255,255), got (%d,%d,%d)", wr, wg, wb)
	}

	// Darken to full black
	black := Darken(base, 1.0)
	br, bg, bb, _ := ToRGB(black)
	if br != 0 || bg != 0 || bb != 0 {
		t.Errorf("Darken(1.0): expected (0,0,0), got (%d,%d,%d)", br, bg, bb)
	}

	// Non-color attribute passes through
	if Lighten(Bright, 0.5) != Bright {
		t.Error("Lighten of non-color attribute should return it unchanged")
	}
}

func TestBlend(t *testing.T) {
	black := TrueColor(0, 0, 0)
	white := TrueColor(255, 255, 255)

	mid := Blend(black, white, 0.5)
	r, g, b, ok := ToRGB(mid)
	if !ok {
		t.Fatal("Blend result has no RGB")
	}
	// Allow rounding of ±1
	if abs(int(r)-128) > 1 || abs(int(g)-128) > 1 || abs(int(b)-128) > 1 {
		t.Errorf("Blend(black,white,0.5): got (%d,%d,%d), want ~(128,128,128)", r, g, b)
	}

	// t=0 returns a, t=1 returns b
	atZero := Blend(black, white, 0.0)
	r0, g0, b0, _ := ToRGB(atZero)
	if r0 != 0 || g0 != 0 || b0 != 0 {
		t.Errorf("Blend(t=0): got (%d,%d,%d), want (0,0,0)", r0, g0, b0)
	}

	atOne := Blend(black, white, 1.0)
	r1, g1, b1, _ := ToRGB(atOne)
	if r1 != 255 || g1 != 255 || b1 != 255 {
		t.Errorf("Blend(t=1): got (%d,%d,%d), want (255,255,255)", r1, g1, b1)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestLuminance(t *testing.T) {
	// Black should have luminance 0
	if l := Luminance(TrueColor(0, 0, 0)); l != 0 {
		t.Errorf("Luminance(black) = %f, want 0", l)
	}

	// White should have luminance 1
	if l := Luminance(TrueColor(255, 255, 255)); math.Abs(l-1.0) > 1e-9 {
		t.Errorf("Luminance(white) = %f, want 1.0", l)
	}

	// Mid-grey should be between 0 and 1
	l := Luminance(TrueColor(128, 128, 128))
	if l <= 0 || l >= 1 {
		t.Errorf("Luminance(mid-grey) = %f, expected in (0,1)", l)
	}

	// Non-color attribute should return 0
	if l := Luminance(Bright); l != 0 {
		t.Errorf("Luminance(Bright) = %f, want 0", l)
	}
}

func TestContrastRatio(t *testing.T) {
	black := TrueColor(0, 0, 0)
	white := TrueColor(255, 255, 255)

	// Black on white = maximum contrast (21:1)
	cr := ContrastRatio(black, white)
	if math.Abs(cr-21.0) > 0.01 {
		t.Errorf("ContrastRatio(black,white) = %f, want ~21.0", cr)
	}

	// Same color = ratio 1
	cr2 := ContrastRatio(black, black)
	if math.Abs(cr2-1.0) > 0.01 {
		t.Errorf("ContrastRatio(black,black) = %f, want 1.0", cr2)
	}

	// Order should not matter
	cr3 := ContrastRatio(white, black)
	if math.Abs(cr3-21.0) > 0.01 {
		t.Errorf("ContrastRatio(white,black) = %f, want ~21.0", cr3)
	}
}

func TestHasSufficientContrast(t *testing.T) {
	black := TrueColor(0, 0, 0)
	white := TrueColor(255, 255, 255)
	darkGrey := TrueColor(80, 80, 80)

	if !HasSufficientContrast(black, white) {
		t.Error("black on white should have sufficient contrast")
	}
	if HasSufficientContrast(darkGrey, black) {
		t.Error("dark grey on black should not have sufficient contrast")
	}
}

func TestBrightBackground(t *testing.T) {
	// LightBlue (94) should map to BackgroundBrightBlue (104)
	bg := LightBlue.Background()
	if bg != BackgroundBrightBlue {
		t.Errorf("LightBlue.Background() = %d, want BackgroundBrightBlue (%d)", uint32(bg), uint32(BackgroundBrightBlue))
	}

	// White (97) should map to BackgroundBrightWhite (107)
	bg2 := White.Background()
	if bg2 != BackgroundBrightWhite {
		t.Errorf("White.Background() = %d, want BackgroundBrightWhite (%d)", uint32(bg2), uint32(BackgroundBrightWhite))
	}
}

func TestBestColorFromHex(t *testing.T) {
	// Just verify it doesn't panic and returns a non-Default value for valid input
	c := BestColorFromHex("#ff0000")
	if c == Default {
		t.Error("BestColorFromHex(#ff0000): got Default, expected a real color")
	}

	// Invalid hex should return Default
	if got := BestColorFromHex("nothex"); got != Default {
		t.Errorf("BestColorFromHex(invalid): got %v, want Default", got)
	}
}

func TestBestBackgroundFromHex(t *testing.T) {
	c := BestBackgroundFromHex("#0000ff")
	if c == DefaultBackground {
		t.Error("BestBackgroundFromHex(#0000ff): got DefaultBackground, expected a real color")
	}

	if got := BestBackgroundFromHex("nothex"); got != DefaultBackground {
		t.Errorf("BestBackgroundFromHex(invalid): got %v, want DefaultBackground", got)
	}
}

func TestBestColorFromHexNoColor(t *testing.T) {
	if EnvNoColor {
		t.Skip("NO_COLOR already set")
	}
	origNoColor := EnvNoColor
	origResetSeq := envResetSeq
	EnvNoColor = true
	envResetSeq = ""
	defer func() {
		EnvNoColor = origNoColor
		envResetSeq = origResetSeq
	}()

	if got := BestColorFromHex("#ff0000"); got != Default {
		t.Errorf("BestColorFromHex with NO_COLOR: got %v, want Default", got)
	}
	if got := BestBackgroundFromHex("#ff0000"); got != DefaultBackground {
		t.Errorf("BestBackgroundFromHex with NO_COLOR: got %v, want DefaultBackground", got)
	}
}
