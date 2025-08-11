package main

import (
	"fmt"
	"github.com/xyproto/vt"
)

// This program demonstrates several different ways of outputting colored text
// Use `./color | cat -v` to see the color codes that are used.

func main() {
	vt.Blue.Output("This is in blue")

	vt.Println("<green>hi</green>")
	vt.Println("<blue>done</blue>")

	vt.Println("<lightgreen>process: <lightred>ERROR<off>")

	vt.LightYellow.Output("jk")

	blue := vt.BackgroundBlue.Get
	green := vt.LightGreen.Get

	fmt.Printf("%s: %s\n", blue("status"), green("good"))

	combined := vt.Blue.Background().Combine(vt.Yellow).Combine(vt.Reverse)
	combined.Output("DONE")
}
