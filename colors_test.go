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
