package main

import (
	"fmt"
	"os"

	glox "github.com/2asm/glox/src"
)

func main() {
	s, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := glox.Interpret(string(s)); err != nil {
		fmt.Println("ERROR: ", err.Error())
	}
}
