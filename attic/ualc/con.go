package con

import (
	"fmt"
)

// Cls clears the console or screen and resets the cursor to (0,0).
func Cls() int {
	fmt.Print("\033[2J\033[H") // ANSI escape sequence to clear screen and reset cursor
	return 0
}

// Print writes the string s at the current cursor location.
func Print(s string) int {
	fmt.Print(s)
	return 0
}

// Printat moves the cursor to coordinate (x, y), then writes the string s.
func Printat(x, y int, s string) int {
	At(x, y)
	Print(s)
	return 0
}

// At moves the console cursor to coordinate (x, y).
func At(x, y int) int {
	// ANSI escape sequence for cursor positioning (1-based)
	fmt.Printf("\033[%d;%dH", y+1, x+1)
	return 0
}