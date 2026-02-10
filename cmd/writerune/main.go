package main

import (
	"github.com/xyproto/vt"
)

func main() {
	// Initialize vt terminal settings
	vt.Init()

	// Prepare a canvas
	c := vt.NewCanvas()

	// Draw things on the canvas
	c.Plot(0, 0, '!')
	c.Plot(0, 1, '!')
	c.Plot(1, 0, '!')
	c.Plot(1, 1, '!')

	c.WriteString(9, 20, vt.Red, vt.BackgroundBlue, "# Thank you code_nomad: http://9m.no/ꪯ鵞")

	bg := vt.BackgroundBlue
	fg := vt.LightYellow

	// Draw the contents of the canvas
	c.Draw()

	// Wait for a keypress
	vt.WaitForKey()

	c.WriteRuneB(0, 0, fg, bg, 'A')
	c.WriteRuneB(0, 1, fg, bg, 'B')
	c.WriteRuneB(1, 0, fg, bg, 'C')
	c.WriteRuneB(1, 1, fg, bg, 'D')

	c.SetRunewise(true)

	// Draw the contents of the canvas
	c.Draw()

	// Wait for a keypress
	vt.WaitForKey()

	// Reset the vt terminal settings
	vt.Close()
}
