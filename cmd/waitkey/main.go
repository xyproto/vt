package main

import (
	"fmt"
	"github.com/xyproto/vt"
)

func main() {
	fmt.Println("Waiting for either one of these: Ctrl-C, Space, Return or Escape...")
	vt.WaitForKey()
	fmt.Println("There you go!")
}
