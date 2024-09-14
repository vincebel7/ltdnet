package main

import (
	"bufio"
	"fmt"
	"os"
)

// launchManual reads the USER-MANUAL.md file and prints its content to the screen.
func launchManual() {
	file, err := os.Open("USER-MANUAL.md")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
}
