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

	// Text input state
	text := ""
	cursorPos := 0
	colorIdx := uint8(0)

	render := func() {
		c.Clear()
		w, h := c.W(), c.H()

		// Centered box
		boxW := uint(40)
		boxH := uint(3)
		if boxW > w-2 {
			boxW = w - 2
		}
		x := w/2 - boxW/2
		y := h/2 - boxH/2

		// Border
		borderFg := vt.Color256(colorIdx)
		c.PlotColor(x-1, y-1, borderFg, '╭')
		c.PlotColor(x+boxW, y-1, borderFg, '╮')
		c.PlotColor(x-1, y+boxH-2, borderFg, '╰')
		c.PlotColor(x+boxW, y+boxH-2, borderFg, '╯')
		for i := x; i < x+boxW; i++ {
			c.PlotColor(i, y-1, borderFg, '─')
			c.PlotColor(i, y+boxH-2, borderFg, '─')
		}
		c.PlotColor(x-1, y, borderFg, '│')
		c.PlotColor(x+boxW, y, borderFg, '│')

		// Text content (with horizontal scrolling)
		innerW := int(boxW)
		displayText := text
		displayOffset := 0
		if cursorPos > innerW-1 {
			displayOffset = cursorPos - innerW + 1
		}
		if displayOffset > 0 {
			displayText = text[displayOffset:]
		}
		if len(displayText) > innerW {
			displayText = displayText[:innerW]
		}

		c.Write(x, y, vt.White, vt.DefaultBackground, displayText)

		// Cursor
		cursorScreenX := x + uint(cursorPos-displayOffset)
		if cursorScreenX < x+boxW {
			c.PlotColor(cursorScreenX, y, vt.White.Bold(), '_')
		}

		// Help text
		help := "Type text, Enter to clear, Ctrl-C to quit"
		c.Write(w/2-uint(len(help))/2, h-1, vt.LightGray, vt.DefaultBackground, help)

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
			continue
		default:
		}

		key := tty.ReadKey()
		switch {
		case key == "c:3" || key == "c:17":
			return
		case key == "c:13": // Enter
			text = ""
			cursorPos = 0
		case key == "c:8" || key == "c:127": // Backspace
			if cursorPos > 0 {
				runes := []rune(text)
				runes = append(runes[:cursorPos-1], runes[cursorPos:]...)
				text = string(runes)
				cursorPos--
			}
		case key == "←":
			if cursorPos > 0 {
				cursorPos--
			}
		case key == "→":
			if cursorPos < len([]rune(text)) {
				cursorPos++
			}
		case key == "c:1": // Ctrl-A (home)
			cursorPos = 0
		case key == "c:5": // Ctrl-E (end)
			cursorPos = len([]rune(text))
		case !strings.HasPrefix(key, "c:") && len(key) > 0:
			runes := []rune(text)
			newRunes := make([]rune, 0, len(runes)+1)
			newRunes = append(newRunes, runes[:cursorPos]...)
			newRunes = append(newRunes, []rune(key)...)
			newRunes = append(newRunes, runes[cursorPos:]...)
			text = string(newRunes)
			cursorPos += len([]rune(key))
		}
		colorIdx++
		render()
	}
}
