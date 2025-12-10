package sys

import (
	"machine"
	"os"
	"runtime"
	"time"
)

// Exit terminates program execution with the given code.
func Exit(code int) int {
	os.Exit(code)
	return 0 // Never reached, just to satisfy the ual function signature
}

// Millis returns the number of milliseconds since system startup.
func Millis() int {
	return int(time.Now().UnixMilli() % (1 << 31))
}

// Reboot resets or reboots the system.
func Reboot() int {
	// TinyGo-specific reset functionality
	machine.CPUReset()
	
	// This code should never be reached, but just in case...
	runtime.Breakpoint()
	return 0
}