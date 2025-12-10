package io

import (
	"machine"
)

// Constants for pin modes
const (
	OUTPUT = 1
	INPUT  = 0
)

// PinMode sets the mode of the pin (e.g. INPUT or OUTPUT).
func PinMode(pin, mode int) int {
	if pin < 0 || pin >= len(machine.GPIO) {
		return -1 // Invalid pin
	}
	
	p := machine.GPIO{pin}
	
	if mode == OUTPUT {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	} else {
		p.Configure(machine.PinConfig{Mode: machine.PinInput})
	}
	
	return 0
}

// WritePin writes a digital value (0 or 1) to an output pin.
func WritePin(pin, value int) int {
	if pin < 0 || pin >= len(machine.GPIO) {
		return -1 // Invalid pin
	}
	
	p := machine.GPIO{pin}
	p.Set(value != 0)
	
	return 0
}

// ReadPin reads the current digital value (0 or 1) from the pin.
func ReadPin(pin int) int {
	if pin < 0 || pin >= len(machine.GPIO) {
		return -1 // Invalid pin
	}
	
	p := machine.GPIO{pin}
	if p.Get() {
		return 1
	}
	return 0
}