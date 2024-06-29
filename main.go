package main

import (
	"fmt"
	"github.com/2asm/glox/src"
)

func main() {
	var s = `
let i = 0;
let sum = 0;
while i < 1_000_000 {
    i += 1;
    sum += i;
}
print sum;
`
	fmt.Println("-----------------------")
	fmt.Println(s)
	fmt.Println("-----------------------")
	glox.Interpret(s)

}
