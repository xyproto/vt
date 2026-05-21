package vt

import (
	"fmt"
	"strconv"
	"strings"
)

// MouseButton identifies which mouse button was used.
type MouseButton uint8

// Mouse buttons.
const (
	MouseLeft MouseButton = iota
	MouseMiddle
	MouseRight
	MouseNone
	MouseWheelUp    MouseButton = 64
	MouseWheelDown  MouseButton = 65
	MouseWheelRight MouseButton = 66
	MouseWheelLeft  MouseButton = 67
	Mouse8          MouseButton = 128
	Mouse9          MouseButton = 129
	Mouse10         MouseButton = 130
	Mouse11         MouseButton = 131
)

// MouseAction describes the type of mouse event.
type MouseAction uint8

// Mouse actions.
const (
	MousePress MouseAction = iota
	MouseRelease
	MouseMotion
	MouseDrag
)

// MouseEvent represents a parsed mouse event from the terminal.
type MouseEvent struct {
	Col    int
	Row    int
	Button MouseButton
	Action MouseAction
	Mod    uint8 // shift=4, alt=8, ctrl=16
}

// Mouse mode escape sequences.
const (
	mouseEnableSGR    = "\x1b[?1006h"                // SGR mouse mode
	mouseEnableAll    = "\x1b[?1002;1003;1004;1006h" // all events + SGR
	mouseDisableAll   = "\x1b[?1002;1003;1004;1006l"
	mouseEnablePixels = "\x1b[?1002;1003;1004;1016h" // SGR pixel mode
	mouseDisable      = "\x1b[?1000;1002;1003;1006;1016l"
)

// EnableMouse enables SGR mouse reporting for all events (press, release, motion, drag).
func EnableMouse() {
	fmt.Print(mouseEnableAll)
}

// DisableMouse disables all mouse reporting.
func DisableMouse() {
	fmt.Print(mouseDisable)
}

// EnableMousePixels enables SGR pixel mouse mode (sub-cell precision).
func EnableMousePixels() {
	fmt.Print(mouseEnablePixels)
}

// ParseMouseSGR parses an SGR mouse sequence.
// Input should be the parameter string after CSI < and before M or m.
// terminator is 'M' for press/motion or 'm' for release.
// Example: CSI < 0;15;3 M → button 0, col 15, row 3, press.
func ParseMouseSGR(params string, terminator byte) (MouseEvent, bool) {
	parts := strings.Split(params, ";")
	if len(parts) != 3 {
		return MouseEvent{}, false
	}
	code, err1 := strconv.Atoi(parts[0])
	col, err2 := strconv.Atoi(parts[1])
	row, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return MouseEvent{}, false
	}

	// SGR reports 1-based coordinates
	col--
	row--

	ev := MouseEvent{
		Col: col,
		Row: row,
		Mod: uint8((code >> 2) & 0x07),
	}

	// Determine button
	low := code & 0x03
	high := code & 0xC0
	switch {
	case high == 64:
		ev.Button = MouseButton(64 + low) // wheel
	case high == 128:
		ev.Button = MouseButton(128 + low) // extended
	case low == 3 && terminator == 'M':
		ev.Button = MouseNone // motion with no button
	default:
		ev.Button = MouseButton(low)
	}

	// Determine action
	switch {
	case terminator == 'm':
		ev.Action = MouseRelease
	case code&32 != 0:
		if ev.Button == MouseNone {
			ev.Action = MouseMotion
		} else {
			ev.Action = MouseDrag
		}
	default:
		ev.Action = MousePress
	}

	return ev, true
}
