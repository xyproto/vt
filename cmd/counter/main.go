package main

import (
	"fmt"
	"os"
	"os/signal"
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

	count := 0
	label := "Press Enter"

	render := func() {
		c.Clear()
		w, h := c.W(), c.H()

		// Centered button
		btnW := uint(len(label) + 4)
		btnH := uint(3)
		x := w/2 - btnW/2
		y := h/2 - btnH/2

		// Draw border
		borderFg := vt.TrueColor(100, 200, 100)
		c.PlotColor(x, y, borderFg, '╭')
		c.PlotColor(x+btnW-1, y, borderFg, '╮')
		c.PlotColor(x, y+btnH-1, borderFg, '╰')
		c.PlotColor(x+btnW-1, y+btnH-1, borderFg, '╯')
		for i := uint(1); i < btnW-1; i++ {
			c.PlotColor(x+i, y, borderFg, '─')
			c.PlotColor(x+i, y+btnH-1, borderFg, '─')
		}
		for i := uint(1); i < btnH-1; i++ {
			c.PlotColor(x, y+i, borderFg, '│')
			c.PlotColor(x+btnW-1, y+i, borderFg, '│')
		}

		// Draw label
		c.Write(x+2, y+1, vt.White.Bold(), vt.DefaultBackground, label)

		// Status bar
		status := " Press q to quit"
		c.Write(0, h-1, vt.LightGray, vt.DefaultBackground, status)

		c.HideCursorAndDraw()
	}

	render()

	for {
		select {
		case <-sigCh:
			if nc := c.Resized(); nc != nil {
				c = nc
			}
			render()
		default:
		}

		key := tty.ReadKey()
		switch key {
		case "q", "c:3", "c:17":
			return
		case "c:13": // Enter
			count++
			label = fmt.Sprintf("Presses: %d", count)
		}
		render()
	}
}
