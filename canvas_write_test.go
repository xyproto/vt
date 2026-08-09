package vt

import "testing"

// runeAt is At without the error, for terser assertions
func runeAt(t *testing.T, c *Canvas, x, y uint) rune {
	t.Helper()
	r, err := c.At(x, y)
	if err != nil {
		t.Fatalf("At(%d,%d): %v", x, y, err)
	}
	return r
}

// Writes must stop at the right edge. Running past it used to overwrite the
// start of the row below, and past the last cell it ran off the buffer.
func TestWritesStopAtTheEndOfTheRow(t *testing.T) {
	t.Run("WriteString", func(t *testing.T) {
		c := NewCanvasWithSize(10, 2)
		c.WriteString(5, 0, White, Black, "0123456789")
		if r := runeAt(t, c, 9, 0); r != '4' {
			t.Errorf("last cell of row 0: got %q, want '4'", r)
		}
		if r := runeAt(t, c, 0, 1); r != 0 {
			t.Errorf("row 1 was written to: got %q", r)
		}
	})
	t.Run("WriteRunesB", func(t *testing.T) {
		c := NewCanvasWithSize(10, 2)
		c.WriteRunesB(5, 1, White, Black, 'x', 10) // would run off the end
		if r := runeAt(t, c, 9, 1); r != 'x' {
			t.Errorf("last cell: got %q, want 'x'", r)
		}
	})
	t.Run("out of range is a no-op", func(t *testing.T) {
		c := NewCanvasWithSize(4, 2)
		c.WriteRunesB(9, 9, White, Black, 'x', 4)
		c.WriteBackground(9, 9, Black)
		c.WriteBackgroundAddRuneIfEmpty(9, 9, Black, 'x')
	})
}

// A double-width rune occupies two cells. Overwriting one of them has to blank
// the other, or the row renders with the wrong number of columns and wraps.
func TestOverwritingAWideRuneClearsItsOtherHalf(t *testing.T) {
	t.Run("overwrite the wide half", func(t *testing.T) {
		c := NewCanvasWithSize(6, 1)
		c.WriteWideRuneB(0, 0, White, Black, '中')
		c.WriteRuneB(0, 0, White, Black, 'x')
		if r := runeAt(t, c, 1, 0); r != ' ' {
			t.Errorf("continuation cell: got %q, want a space", r)
		}
	})
	t.Run("overwrite the continuation half", func(t *testing.T) {
		c := NewCanvasWithSize(6, 1)
		c.WriteWideRuneB(0, 0, White, Black, '中')
		c.WriteRuneB(1, 0, White, Black, 'x')
		if r := runeAt(t, c, 0, 0); r != ' ' {
			t.Errorf("wide cell: got %q, want a space", r)
		}
	})
	t.Run("a string written over a wide rune", func(t *testing.T) {
		c := NewCanvasWithSize(6, 1)
		c.WriteWideRuneB(2, 0, White, Black, '中')
		c.WriteString(0, 0, White, Black, "abc") // ends on the wide half
		if r := runeAt(t, c, 3, 0); r != ' ' {
			t.Errorf("continuation cell: got %q, want a space", r)
		}
	})
}
