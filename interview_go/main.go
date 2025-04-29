package main

import (
	"fmt"
)

// Add takes two integers and returns their sum
func Add(a, b int) int {
	return a + b
}

func main() {

	a := 5
	b := 10
	// Calculate and print the result
	sum := Add(a, b)
	fmt.Printf("%d + %d = %d\n", a, b, sum)

}
