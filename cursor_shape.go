package vt

import "fmt"

// CursorShape is the shape of the terminal cursor.
type CursorShape int

// Cursor shapes (DECSCUSR values).
const (
	CursorDefault        CursorShape = 0
	CursorBlockBlink     CursorShape = 1
	CursorBlock          CursorShape = 2
	CursorUnderlineBlink CursorShape = 3
	CursorUnderline      CursorShape = 4
	CursorBeamBlink      CursorShape = 5
	CursorBeam           CursorShape = 6
)

// SetCursorShape changes the terminal cursor shape.
// Use CursorDefault to restore the terminal's default cursor.
func SetCursorShape(shape CursorShape) {
	fmt.Printf("\x1b[%d q", int(shape))
}
