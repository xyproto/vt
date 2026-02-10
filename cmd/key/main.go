package main

import (
	"fmt"
	"github.com/xyproto/vt"
	"time"
)

func main() {
	escCount := 0
	tty, err := vt.NewTTY()
	if err != nil {
		panic(err)
	}
	defer tty.Close()
	tty.SetTimeout(10 * time.Millisecond)
	tty.RawMode()
	defer tty.Restore()
	for {
		key := tty.Key()
		if key != 0 {
			fmt.Printf("%d\r\n", key)
		}
		if key == 27 {
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
