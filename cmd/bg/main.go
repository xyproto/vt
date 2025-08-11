package main

import (
	"github.com/xyproto/vt"
)

func main() {
	vt.Init()

	c := vt.NewCanvas()
	c.FillBackground(vt.Blue)
	c.Draw()

	vt.WaitForKey()

	vt.Close()
}
