package main

import (
	"fmt"
	"github.com/xyproto/vt"
)

func main() {
	// Define four functions that takes text and returns text together with VT100 codes
	df := vt.DarkGray.Start
	rf := vt.Blink.Combine(vt.Red).StartStop
	yf := vt.Blink.Combine(vt.Yellow).StartStop
	gf := vt.Blink.Combine(vt.Green).StartStop

	// Define a colored traffic light, with blinking lights
	trafficLight := df(`
	.-----.
	|  `) + rf("O") + df(`  |
	|     |
	|  `) + yf("O") + df(`  |
	|     |
	|  `) + gf("O") + df(`  |
	'-----'
	  | |
	  | |`+vt.Stop())

	// Output the amazing artwork
	fmt.Printf("%s\n\n", trafficLight)
}
