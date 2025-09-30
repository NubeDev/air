package main

import (
	"fmt"
)

func main() {
	// Test different color codes
	colors := map[string]int{
		"cyan":    36,
		"yellow":  33,
		"magenta": 35,
		"green":   32,
		"red":     31,
		"blue":    34,
		"white":   37,
	}

	fmt.Println("Testing colors:")
	for name, code := range colors {
		fmt.Printf("\033[%dm[%s]\033[0m ", code, name)
	}
	fmt.Println()

	// Test our specific format
	fmt.Printf("\033[36m[SVR]\033[0m \033[36m[INF]\033[0m Test message\n")
	fmt.Printf("\033[33m[RST]\033[0m \033[36m[INF]\033[0m HTTP request\n")
	fmt.Printf("\033[31m[AUTH]\033[0m \033[33m[WRN]\033[0m Authentication disabled\n")
}
