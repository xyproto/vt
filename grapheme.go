package vt

// Grapheme-aware string width calculation.
// Ported from gaxis — provides accurate terminal column widths for
// strings containing CJK, emoji, combining characters, and ZWJ sequences.

// StringWidth returns the display width of s in terminal columns.
// It uses per-codepoint wcwidth-style width calculation.
func StringWidth(s string) int {
	total := 0
	for _, r := range s {
		w := RuneWidth(r)
		if w > 0 {
			total += w
		}
	}
	return total
}

// RuneWidth returns the display width of a single rune.
// Returns 2 for wide (CJK/emoji), 0 for zero-width (combining/format),
// -1 for control characters, and 1 otherwise.
func RuneWidth(r rune) int {
	if r == 0 {
		return 0
	}
	if r < 32 || (r >= 0x7f && r < 0xa0) {
		return -1
	}
	if IsWideRune(r) {
		return 2
	}
	if isZeroWidth(r) {
		return 0
	}
	return 1
}

// IsWideRune reports whether r occupies two terminal columns.
func IsWideRune(r rune) bool {
	// CJK and fullwidth ranges
	if r >= 0x1100 && r <= 0x115f {
		return true
	}
	if r >= 0x2e80 && r <= 0xa4cf && r != 0x303f {
		return true
	}
	if r >= 0xac00 && r <= 0xd7a3 {
		return true
	}
	if r >= 0xf900 && r <= 0xfaff {
		return true
	}
	if r >= 0xfe10 && r <= 0xfe19 {
		return true
	}
	if r >= 0xfe30 && r <= 0xfe6f {
		return true
	}
	if r >= 0xff01 && r <= 0xff60 {
		return true
	}
	if r >= 0xffe0 && r <= 0xffe6 {
		return true
	}
	if r >= 0x20000 && r <= 0x2fffd {
		return true
	}
	if r >= 0x30000 && r <= 0x3fffd {
		return true
	}
	// Emoji presentation default
	if r >= 0x1f300 && r <= 0x1f9ff {
		return true
	}
	if r >= 0x1fa00 && r <= 0x1faff {
		return true
	}
	return false
}

func isZeroWidth(r rune) bool {
	if r == 0x00ad || r == 0x200b || r == 0x200c || r == 0x200d ||
		r == 0x2060 || r == 0x034f || r == 0xfeff {
		return true
	}
	if r >= 0x0300 && r <= 0x036f {
		return true
	}
	if r >= 0x180b && r <= 0x180d {
		return true
	}
	if r >= 0xfe00 && r <= 0xfe0f {
		return true
	}
	if r >= 0xe0100 && r <= 0xe01ef {
		return true
	}
	return false
}

// StringWidthZWJ returns the display width handling ZWJ sequences correctly.
// ZWJ-joined sequences (like family emoji 👨‍👩‍👧) are split on ZWJ
// and each part's width is summed, giving the correct width on terminals
// that don't support ZWJ composition.
func StringWidthZWJ(s string) int {
	total := 0
	for {
		idx := indexZWJ(s)
		if idx < 0 {
			total += StringWidth(s)
			break
		}
		total += StringWidth(s[:idx])
		s = s[idx+3:] // ZWJ is 3 bytes (U+200D)
	}
	return total
}

func indexZWJ(s string) int {
	for i := 0; i < len(s)-2; i++ {
		if s[i] == 0xE2 && s[i+1] == 0x80 && s[i+2] == 0x8D {
			return i
		}
	}
	return -1
}
