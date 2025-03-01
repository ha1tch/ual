package math

// Abs returns the absolute value of n.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Min returns the smaller of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of a and b.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Pow performs integer exponentiation. For example, Pow(2,3) yields 8.
func Pow(base, exponent int) int {
	if exponent < 0 {
		return 0 // Integer division would round to zero
	}
	
	result := 1
	for i := 0; i < exponent; i++ {
		result *= base
	}
	
	return result
}