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
	c.Plot(10, 10, '!')
	c.Write(12, 12, vt.LightGreen, vt.BackgroundDefault, "hi")
	c.Write(15, 15, vt.White, vt.BackgroundMagenta, "floating")
	c.PlotColor(12, 17, vt.LightRed, '*')
	c.PlotColor(10, 20, vt.LightBlue, 'ø')
	c.PlotColor(11, 20, vt.LightBlue, 'l')

	c.WriteString(10, 21, vt.White, vt.BackgroundRed, "øl")

	// Draw the contents of the canvas
	c.Draw()

	// Wait for a keypress
	vt.WaitForKey()

	// Reset the vt terminal settings
	vt.Close()
}
