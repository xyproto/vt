package main

import (
	"github.com/xyproto/vt"
)

var (
	red   = vt.Red
	green = vt.Green
	blue  = vt.Blue
	none  = vt.None
)

const (
	TL = '╭' // top left
	TR = '╮' // top right
	BL = '╰' // bottom left
	BR = '╯' // bottom right
	VL = '│' // vertical line, left side
	VR = '│' // vertical line, right side
	HT = '─' // horizontal line
	HB = '─' // horizontal bottom line
)

func main() {
	vt.Init()
	defer vt.Close()

	c := vt.NewCanvas()

	c.WriteRune(12, 14, green, none, TL)
	c.WriteRune(13, 14, green, none, HT)
	c.WriteRune(14, 14, green, none, HT)
	c.WriteRune(15, 14, green, none, HT)
	c.WriteRune(16, 14, green, none, HT)
	c.WriteRune(17, 14, green, none, TR)

	c.WriteRune(12, 15, green, none, VL)
	c.Write(14, 15, green, none, "OK")
	c.WriteRune(17, 15, green, none, VR)

	c.WriteRune(12, 16, green, none, BL)
	c.WriteRune(13, 16, green, none, HB)
	c.WriteRune(14, 16, green, none, HB)
	c.WriteRune(15, 16, green, none, HB)
	c.WriteRune(16, 16, green, none, HB)
	c.WriteRune(17, 16, green, none, BR)

	c.Draw()
	vt.WaitForKey()
}
