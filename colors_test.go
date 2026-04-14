package vt

import (
	"fmt"
	"testing"
)

var ac = Bright.Combine(Blue)

func TestBackground(t *testing.T) {
	if BackgroundBlue.String() != Blue.Background().String() {
		fmt.Println("BLUE BG IS NOT BLUE BG")
		fmt.Println(BackgroundBlue.String() + "FIRST" + Stop())
		fmt.Println(Blue.Background().String() + "SECOND" + Stop())
		t.Fail()
	}
}

func TestInts(t *testing.T) {
	ai := BackgroundBlue.Ints()
	bi := Blue.Background().Ints()
	if len(ai) != len(bi) {
		fmt.Println("A", ai)
		fmt.Println("B", bi)
		fmt.Println("length mismatch")
		t.Fail()
	}
	for i := range ai {
		if ai[i] != bi[i] {
			fmt.Println("NO")
			t.Fail()
		}
	}
}

func BenchmarkNewAttributeColor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Bright.Combine(Blue)
	}
}

func BenchmarkHead(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Head()
	}
}

func BenchmarkTail(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Tail()
	}
}

func BenchmarkBackground(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Background()
	}
}

func BenchmarkStartStop(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.StartStop("test")
	}
}

func BenchmarkGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Get("test")
	}
}

func BenchmarkStart(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Start("test")
	}
}

func BenchmarkStop(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Stop("test")
	}
}

func BenchmarkCombine(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Combine(Red)
	}
}

func BenchmarkBright(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Bright()
	}
}

func BenchmarkInts(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ac.Ints()
	}
}

func TestColor256ToRGB(t *testing.T) {
	// Index 0 (Black): should be the ANSI black approximation
	r, g, b := Color256ToRGB(0)
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("Color256ToRGB(0): got (%d,%d,%d), want (0,0,0)", r, g, b)
	}
	// Index 231: last cube entry (5,5,5) → should be white (255,255,255)
	r, g, b = Color256ToRGB(231)
	if r != 255 || g != 255 || b != 255 {
		t.Errorf("Color256ToRGB(231): got (%d,%d,%d), want (255,255,255)", r, g, b)
	}
	// Index 232: first grayscale → 8,8,8
	r, g, b = Color256ToRGB(232)
	if r != 8 || g != 8 || b != 8 {
		t.Errorf("Color256ToRGB(232): got (%d,%d,%d), want (8,8,8)", r, g, b)
	}
	// Index 255: last grayscale → 238,238,238
	r, g, b = Color256ToRGB(255)
	if r != 238 || g != 238 || b != 238 {
		t.Errorf("Color256ToRGB(255): got (%d,%d,%d), want (238,238,238)", r, g, b)
	}
}

func TestGrayscale256(t *testing.T) {
	if Grayscale256(0) != Color256(232) {
		t.Error("Grayscale256(0) should be Color256(232)")
	}
	if Grayscale256(23) != Color256(255) {
		t.Error("Grayscale256(23) should be Color256(255)")
	}
	// Clamping
	if Grayscale256(100) != Color256(255) {
		t.Error("Grayscale256(100) should clamp to Color256(255)")
	}
}

func TestColorCube(t *testing.T) {
	if ColorCube(0, 0, 0) != Color256(16) {
		t.Error("ColorCube(0,0,0) should be Color256(16)")
	}
	if ColorCube(5, 5, 5) != Color256(231) {
		t.Error("ColorCube(5,5,5) should be Color256(231)")
	}
	if ColorCube(1, 0, 0) != Color256(52) {
		t.Errorf("ColorCube(1,0,0) should be Color256(52), got %v", ColorCube(1, 0, 0))
	}
	// Clamping
	if ColorCube(6, 6, 6) != Color256(231) {
		t.Error("ColorCube(6,6,6) should clamp to Color256(231)")
	}
}

func TestNearestColor256RoundTrip(t *testing.T) {
	// For every palette entry the nearest-256 of its own RGB should map back to itself
	for i := range 256 {
		r, g, b := Color256ToRGB(uint8(i))
		got := NearestColor256(r, g, b)
		// The round-trip may legitimately land on a different index with identical
		// RGB (duplicates exist in the low-16 range), so compare the RGB values.
		gr, gg, gb := Color256ToRGB(uint8(got & 0xFF))
		if gr != r || gg != g || gb != b {
			t.Errorf("round-trip failed for index %d (rgb %d,%d,%d): nearest gave index %d (rgb %d,%d,%d)",
				i, r, g, b, got&0xFF, gr, gg, gb)
		}
	}
}

func TestNearestANSI16(t *testing.T) {
	tests := []struct {
		hex      string
		expected AttributeColor
	}{
		{"#af0005", Red},        // dark red → Red
		{"#ff0000", LightRed},   // bright red → LightRed
		{"#00ff00", LightGreen}, // bright green → LightGreen
		{"#000000", Black},      // black → Black
		{"#ffffff", White},      // white → White
		{"#00aaff", Cyan},       // sky blue (0,170,255) → Cyan (0,205,205) is nearest
		{"#888888", DarkGray},   // mid gray → DarkGray
	}
	for _, tt := range tests {
		r, g, b, err := parseHexColor(tt.hex)
		if err != nil {
			t.Errorf("parseHexColor(%q) error: %v", tt.hex, err)
			continue
		}
		got := nearestANSI16(r, g, b)
		if got != tt.expected {
			t.Errorf("nearestANSI16 for %s: got %v, want %v", tt.hex, got, tt.expected)
		}
	}
}
