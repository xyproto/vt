package vt

import "strings"

// ColorSplit splits on the first sep in line. It returns two parts: left and right.
// The right part includes the sep itself (so subsequent splits see it).
// nil color funcs are skipped. reverse=true swaps which side gets the fallback when sep is absent.
func ColorSplit(line, sep string, colorHead, colorSep, colorTail func(string) string, reverse bool) (string, string) {
	if sep == "" {
		if reverse {
			return "", line
		}
		return line, ""
	}
	idx := strings.Index(line, sep)
	if idx == -1 {
		if reverse {
			return "", line
		}
		return line, ""
	}
	head := line[:idx]
	sepAndTail := line[idx:] // includes sep
	var left, right strings.Builder
	// Head
	if colorHead != nil {
		left.WriteString(colorHead(head))
	} else {
		left.WriteString(head)
	}
	// Separator
	if colorSep != nil {
		right.WriteString(colorSep(sep))
	} else {
		right.WriteString(sep)
	}
	// Tail
	tail := sepAndTail[len(sep):]
	if colorTail != nil {
		right.WriteString(colorTail(tail))
	} else {
		right.WriteString(tail)
	}
	return left.String(), right.String()
}
