package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xyproto/vt"
)

func main() {
	tty, err := vt.NewTTY()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	vt.Init()
	defer vt.Close()

	c := vt.NewCanvas()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	splitCol := 0
	leftText := "This is the left pane.\n\nYou can resize the split with h/l or left/right arrows."
	rightText := "This is the right pane.\n\nPress q or Ctrl+C to quit."

	leftBg := vt.TrueBackground(16, 16, 32)
	leftFg := vt.TrueColor(200, 200, 255)
	rightBg := vt.TrueBackground(32, 16, 16)
	rightFg := vt.TrueColor(255, 200, 200)
	divFg := vt.TrueColor(80, 80, 120)

	render := func() {
		c.Clear()
		w, h := int(c.W()), int(c.H())

		split := splitCol
		if split <= 0 || split >= w-1 {
			split = w / 2
		}

		// Left pane
		leftLines := strings.Split(leftText, "\n")
		for row, line := range leftLines {
			if row >= h {
				break
			}
			// Fill background
			for col := 0; col < split-1; col++ {
				c.WriteRuneB(uint(col), uint(row), leftFg, leftBg, ' ')
			}
			for col, r := range line {
				if col >= split-1 {
					break
				}
				c.WriteRuneB(uint(col), uint(row), leftFg, leftBg, r)
			}
		}
		// Fill remaining left rows
		for row := len(leftLines); row < h; row++ {
			for col := 0; col < split-1; col++ {
				c.WriteRuneB(uint(col), uint(row), leftFg, leftBg, ' ')
			}
		}

		// Divider
		for row := range h {
			c.PlotColor(uint(split-1), uint(row), divFg, '│')
		}

		// Right pane
		rightLines := strings.Split(rightText, "\n")
		for row, line := range rightLines {
			if row >= h {
				break
			}
			for col := split; col < w; col++ {
				c.WriteRuneB(uint(col), uint(row), rightFg, rightBg, ' ')
			}
			for col, r := range line {
				if split+col >= w {
					break
				}
				c.WriteRuneB(uint(split+col), uint(row), rightFg, rightBg, r)
			}
		}
		for row := len(rightLines); row < h; row++ {
			for col := split; col < w; col++ {
				c.WriteRuneB(uint(col), uint(row), rightFg, rightBg, ' ')
			}
		}

		c.HideCursorAndDraw()
	}

	render()

	for {
		select {
		case <-sigCh:
			if nc := c.Resized(); nc != nil {
				c = nc
			}
			splitCol = 0
			render()
			continue
		default:
		}

		key := tty.ReadKey()
		w := int(c.W())
		if splitCol <= 0 {
			splitCol = w / 2
		}
		switch key {
		case "q", "c:3", "c:17":
			return
		case "h", "←":
			if splitCol > 3 {
				splitCol--
			}
		case "l", "→":
			if splitCol < w-3 {
				splitCol++
			}
		case "=":
			splitCol = w / 2
		}
		render()
	}
}
