package main

import (
	"fmt"

	"github.com/xyproto/vt"
)

func main() {
	escCount := 0
	tty, err := vt.NewTTY()
	if err != nil {
		panic(err)
	}
	for {
		key := tty.String()
		if key != "" {
			fmt.Print(key + "\r\n")
		}
		if key == "c:27" {
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
