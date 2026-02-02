package vt

import "testing"

func TestLowContrastLightGrayOnGray(t *testing.T) {
	if !LowContrast(LightGray, DarkGray.Background(), false) {
		t.Fatalf("expected light gray on gray background to be low contrast")
	}
}
