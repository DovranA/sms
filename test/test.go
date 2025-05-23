package main

import (
	"fmt"
	"regexp"
)

func main() {
	input := "Sizin balansynyz 2,96 manat"

	// Define a regex pattern to match numbers (including decimal and commas)
	re := regexp.MustCompile(`[\d]+(?:,[\d]{1,2})?`)

	// Find the first match
	match := re.FindString(input)

	if match != "" {
		fmt.Printf("Extracted number: %s\n", match)
	} else {
		fmt.Println("No number found.")
	}
}
