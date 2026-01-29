package main

import (
	"fmt"
	"unicode"

	"github.com/xyproto/vt"
)

func main() {
	escCount := 0
	tty, err := vt.NewTTY()
	if err != nil {
		panic(err)
	}
	for {
		key := tty.Rune()
		if key != rune(0) {
			if unicode.IsPrint(key) {
				fmt.Print(string(key) + "\r\n")
			} else {
				fmt.Printf("%U\r\n", key)
			}
		}
		if key == rune(27) {
			if escCount == 0 {
				fmt.Print("Press ESC again to exit\r\n")
			} else {
				fmt.Print("bye!\r\n")
			}
			escCount++
		}
		if escCount > 1 {
			break
		}
	}
	tty.Close()
}
