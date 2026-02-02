package vt

import "testing"

func TestLowContrastLightGrayOnGray(t *testing.T) {
	if !LowContrast(LightGray, DarkGray.Background(), false) {
		t.Fatalf("expected light gray on gray background to be low contrast")
	}
}

func TestLowContrastGrayOnGray(t *testing.T) {
	// Black on white (both are gray - R==G==B)
	if !LowContrast(Black, White.Background(), false) {
		t.Fatalf("expected black on white to be low contrast (gray on gray)")
	}
	// Default colors are gray too
	if !LowContrast(Default, DefaultBackground, false) {
		t.Fatalf("expected default on default to be low contrast (gray on gray)")
	}
	if !LowContrast(Default, DefaultBackground, true) {
		t.Fatalf("expected default on default (light bg) to be low contrast (gray on gray)")
	}
}
