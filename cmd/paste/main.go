package main

import (
	"fmt"
	"os"

	"github.com/xyproto/vt"
)

func main() {
	// Initialize terminal with bracketed paste support
	vt.Init()
	vt.EnableBracketedPaste()
	defer vt.Close()

	// Create TTY for key input
	tty, err := vt.NewTTY()
	if err != nil {
		fmt.Printf("Error creating TTY: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	fmt.Println("Paste Demo - Linux terminal and tmux compatibility")
	fmt.Println("==================================================")
	fmt.Println("Try these operations:")
	fmt.Println("- Regular typing")
	fmt.Println("- Ctrl+C to exit")
	fmt.Println("- Shift+Insert to paste")
	fmt.Println("- Middle-click to paste (X11)")
	fmt.Println("- Ctrl+Arrow keys for word navigation")
	fmt.Println("- Function keys (F1-F12)")
	fmt.Println("- Bracketed paste mode")
	fmt.Println("")
	fmt.Print("Input: ")

	var buffer []rune

	for {
		key := tty.Key()
		if key == 0 {
			continue
		}

		switch {
		case key == 3: // Ctrl+C
			fmt.Println("\nExiting...")
			return

		case key == 13 || key == 10: // Enter/Return (normalized)
			fmt.Printf("\nYou entered: %s\n", string(buffer))
			buffer = buffer[:0]
			fmt.Print("Input: ")

		case key == 8 || key == 127: // Backspace/Delete (normalized)
			if len(buffer) > 0 {
				buffer = buffer[:len(buffer)-1]
				fmt.Print("\b \b")
			}

		case vt.IsShiftInsert(key): // Shift+Insert paste
			fmt.Print("[PASTE]")
			pasteData, err := tty.ReadPasteData()
			if err == nil && pasteData != "" {
				for _, r := range pasteData {
					buffer = append(buffer, r)
					fmt.Printf("%c", r)
				}
			}

		case vt.IsPasteStart(key): // Bracketed paste start
			fmt.Print("[BRACKETED_PASTE]")
			pasteData, err := tty.ReadPasteData()
			if err == nil && pasteData != "" {
				for _, r := range pasteData {
					buffer = append(buffer, r)
					fmt.Printf("%c", r)
				}
			}

		case key >= 280 && key <= 285: // Ctrl combinations
			switch key {
			case 280:
				fmt.Print("[Ctrl+Up]")
			case 281:
				fmt.Print("[Ctrl+Down]")
			case 282:
				fmt.Print("[Ctrl+Right]")
			case 283:
				fmt.Print("[Ctrl+Left]")
			case 284:
				fmt.Print("[Ctrl+Home]")
			case 285:
				fmt.Print("[Ctrl+End]")
			}

		case key >= 32 && key < 127: // Printable characters
			buffer = append(buffer, rune(key))
			fmt.Printf("%c", rune(key))

		case key >= 260 && key <= 271: // Function keys
			fmt.Printf("[F%d]", key-259)

		default:
			// Show other special keys
			str := tty.String()
			if str != "" {
				fmt.Printf("[%s]", str)
			}
		}
	}
}