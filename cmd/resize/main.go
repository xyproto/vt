package main

import (
	"fmt"
	"github.com/xyproto/vt"
	"time"
)

func main() {
	fmt.Println("Try resizing the terminal")
	for {
		w, h := vt.MustTermSize()
		fmt.Printf("%dx%d\n", w, h)
		time.Sleep(time.Millisecond * 500)
	}
}
