package vt

import "testing"

func ExamplePrintln() {
	o := NewTextOutput(true, true)
	o.Println("hello")
	// Output:
	// hello
}

func TestTags(t *testing.T) {
	o := NewTextOutput(true, true)
	a := o.LightTags("<blue>hi</blue>")
	b := o.LightBlue("hi")
	if a != b {
		t.Fatal(a + " != " + b)
	}
}

// TestNoColorRespected verifies that when EnvNoColor is true, no ANSI escape
// sequences are emitted by any color-producing code path.
func TestNoColorRespected(t *testing.T) {
	if EnvNoColor {
		t.Skip("NO_COLOR already set in environment; test is not meaningful")
	}

	// Temporarily set EnvNoColor and the derived envResetSeq so the code
	// paths see NO_COLOR=1 without relying on os.Setenv.
	origEnvNoColor := EnvNoColor
	origEnvResetSeq := envResetSeq
	EnvNoColor = true
	envResetSeq = ""
	defer func() {
		EnvNoColor = origEnvNoColor
		envResetSeq = origEnvResetSeq
	}()

	// AttributeColor.String() must return ""
	for _, ac := range []AttributeColor{Red, Green, Blue, White, LightBlue, Color256(42), TrueColor(255, 100, 0)} {
		if s := ac.String(); s != "" {
			t.Errorf("AttributeColor.String() = %q, want \"\" when NO_COLOR is set", s)
		}
	}

	// Wrap / Get / StartStop must return the bare text
	if got := Red.Wrap("hello"); got != "hello" {
		t.Errorf("Red.Wrap(\"hello\") = %q, want \"hello\"", got)
	}
	if got := Blue.Get("world"); got != "world" {
		t.Errorf("Blue.Get(\"world\") = %q, want \"world\"", got)
	}

	// Stop (method) must not append a reset sequence
	if got := Green.Stop("text"); got != "text" {
		t.Errorf("Green.Stop(\"text\") = %q, want \"text\"", got)
	}

	// Stop (package function) must return ""
	if got := Stop(); got != "" {
		t.Errorf("Stop() = %q, want \"\"", got)
	}

	// BestColor must return Default, not a computed color
	if got := BestColor(255, 128, 0); got != Default {
		t.Errorf("BestColor() = %v, want Default when NO_COLOR is set", got)
	}
	if got := BestBackground(255, 128, 0); got != DefaultBackground {
		t.Errorf("BestBackground() = %v, want DefaultBackground when NO_COLOR is set", got)
	}
}
