// Compute-Only Benchmarks
//
// These benchmarks isolate the COMPUTATION ONLY, excluding:
// - Stack creation
// - Push/Pop operations
// - Lock/Unlock overhead
// - Byte conversion (floatToBytes, bytesToFloat)
//
// This measures the quality of the generated compute block code
// compared to equivalent idiomatic Go.

package main

import (
	"math"
	"testing"
)

// =============================================================================
// MANDELBROT - Escape iteration for a single point
// =============================================================================

// Pure Go: idiomatic implementation
func computeGo_Mandelbrot(cr, ci float64) float64 {
	var zr, zi float64 = 0, 0
	var zr2, zi2 float64
	const maxIter = 1000.0
	const escape = 4.0

	for iter := 0.0; iter < maxIter; iter++ {
		zr2 = zr * zr
		zi2 = zi * zi
		if zr2+zi2 > escape {
			return iter
		}
		zi = 2*zr*zi + ci
		zr = zr2 - zi2 + cr
	}
	return maxIter
}

// ual-style: mimics generated compute block interior
// (same algorithm, same variable pattern as codegen produces)
func computeUal_Mandelbrot(cr, ci float64) float64 {
	var zr float64 = 0
	var zi float64 = 0
	var zr2 float64 = 0
	var zi2 float64 = 0
	var iter float64 = 0
	var max_iter float64 = 1000
	var escape float64 = 4

	for iter < max_iter {
		zr2 = zr * zr
		zi2 = zi * zi
		if (zr2 + zi2) > escape {
			return iter
		}
		zi = ((2 * zr) * zi) + ci
		zr = (zr2 - zi2) + cr
		iter = iter + 1
	}
	return max_iter
}

func BenchmarkCompute_Mandelbrot_Go(b *testing.B) {
	cr, ci := 0.25, 0.5
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeGo_Mandelbrot(cr, ci)
	}
	_ = result
}

func BenchmarkCompute_Mandelbrot_Ual(b *testing.B) {
	cr, ci := 0.25, 0.5
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeUal_Mandelbrot(cr, ci)
	}
	_ = result
}

// =============================================================================
// INTEGRATE - Trapezoidal integration of x²
// =============================================================================

// Pure Go: idiomatic implementation
func computeGo_Integrate(a, bval, n float64) float64 {
	h := (bval - a) / n
	sum := (a * a) / 2.0

	for i := 1.0; i < n; i++ {
		x := a + i*h
		sum += x * x
	}

	sum += (bval * bval) / 2.0
	return h * sum
}

// ual-style: mimics generated compute block interior
func computeUal_Integrate(a, bval, n float64) float64 {
	var h float64 = (bval - a) / n
	var sum float64 = 0
	var x float64 = a
	var idx float64 = 0
	var fa float64 = a * a
	sum = fa / 2
	idx = 1
	for idx < n {
		x = a + (idx * h)
		var fx float64 = x * x
		sum = sum + fx
		idx = idx + 1
	}
	var fb float64 = bval * bval
	sum = sum + (fb / 2)
	var result float64 = h * sum
	return result
}

func BenchmarkCompute_Integrate_Go(b *testing.B) {
	a, bval, n := 0.0, 1.0, 1000.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeGo_Integrate(a, bval, n)
	}
	_ = result
}

func BenchmarkCompute_Integrate_Ual(b *testing.B) {
	a, bval, n := 0.0, 1.0, 1000.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeUal_Integrate(a, bval, n)
	}
	_ = result
}

// =============================================================================
// LEIBNIZ - π calculation via Leibniz series
// =============================================================================

// Pure Go: idiomatic implementation
func computeGo_Leibniz(terms float64) float64 {
	sum := 0.0
	sign := 1.0
	denom := 1.0

	for i := 0.0; i < terms; i++ {
		sum += sign / denom
		sign = -sign
		denom += 2.0
	}
	return 4.0 * sum
}

// ual-style: mimics generated compute block interior
func computeUal_Leibniz(terms float64) float64 {
	var sum float64 = 0
	var sign float64 = 1
	var denom float64 = 1
	var idx float64 = 0

	for idx < terms {
		sum = sum + (sign / denom)
		sign = 0 - sign
		denom = denom + 2
		idx = idx + 1
	}
	var pi float64 = 4 * sum
	return pi
}

func BenchmarkCompute_Leibniz_Go(b *testing.B) {
	terms := 100000.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeGo_Leibniz(terms)
	}
	_ = result
}

func BenchmarkCompute_Leibniz_Ual(b *testing.B) {
	terms := 100000.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeUal_Leibniz(terms)
	}
	_ = result
}

// =============================================================================
// NEWTON-RAPHSON - Square root via iteration
// =============================================================================

// Pure Go: idiomatic implementation
func computeGo_Newton(x float64) float64 {
	guess := x / 2
	for i := 0; i < 20; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}

// ual-style: mimics generated compute block interior
func computeUal_Newton(x float64) float64 {
	var guess float64 = x / 2
	var i float64 = 0
	for i < 20 {
		guess = (guess + (x / guess)) / 2
		i = i + 1
	}
	return guess
}

func BenchmarkCompute_Newton_Go(b *testing.B) {
	x := 2.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeGo_Newton(x)
	}
	_ = result
}

func BenchmarkCompute_Newton_Ual(b *testing.B) {
	x := 2.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeUal_Newton(x)
	}
	_ = result
}

// =============================================================================
// LOCAL ARRAY - Sum with buffer (tests array codegen quality)
// =============================================================================

// Pure Go: idiomatic implementation
func computeGo_ArraySum(n int) int64 {
	buf := make([]int64, 100)
	for i := 0; i < n; i++ {
		buf[i] = int64(i + 1)
	}
	var sum int64 = 0
	for i := 0; i < n; i++ {
		sum += buf[i]
	}
	return sum
}

// Pure Go with fixed array (closer to ual codegen)
func computeGo_FixedArraySum(n int64) int64 {
	var buf [100]int64
	var i int64 = 0
	for i < n {
		buf[i] = i + 1
		i = i + 1
	}
	var sum int64 = 0
	i = 0
	for i < n {
		sum = sum + buf[i]
		i = i + 1
	}
	return sum
}

// ual-style: exact pattern from codegen
func computeUal_ArraySum(n int64) int64 {
	var buf [100]int64
	var i int64 = 0

	for i < n {
		buf[int(i)] = (i + 1)
		i = (i + 1)
	}

	var sum int64 = 0
	i = 0
	for i < n {
		sum = (sum + buf[int(i)])
		i = (i + 1)
	}
	return sum
}

func BenchmarkCompute_ArraySum_Go(b *testing.B) {
	var result int64
	for i := 0; i < b.N; i++ {
		result = computeGo_ArraySum(50)
	}
	_ = result
}

func BenchmarkCompute_ArraySum_GoFixed(b *testing.B) {
	var result int64
	for i := 0; i < b.N; i++ {
		result = computeGo_FixedArraySum(50)
	}
	_ = result
}

func BenchmarkCompute_ArraySum_Ual(b *testing.B) {
	var result int64
	for i := 0; i < b.N; i++ {
		result = computeUal_ArraySum(50)
	}
	_ = result
}

// =============================================================================
// DP FIBONACCI - Dynamic programming with array
// =============================================================================

// Pure Go: idiomatic implementation
func computeGo_DPFib(n int) int64 {
	dp := make([]int64, n+1)
	dp[0] = 0
	dp[1] = 1
	for i := 2; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]
	}
	return dp[n]
}

// Pure Go with fixed array
func computeGo_FixedDPFib(n int64) int64 {
	var dp [100]int64
	dp[0] = 0
	dp[1] = 1
	var i int64 = 2
	for i <= n {
		dp[i] = dp[i-1] + dp[i-2]
		i++
	}
	return dp[n]
}

// ual-style: exact pattern from codegen
func computeUal_DPFib(n int64) int64 {
	var dp [100]int64
	dp[0] = 0
	dp[1] = 1
	var i int64 = 2
	for i <= n {
		dp[int(i)] = (dp[int((i - 1))] + dp[int((i - 2))])
		i = (i + 1)
	}
	return dp[int(n)]
}

func BenchmarkCompute_DPFib_Go(b *testing.B) {
	var result int64
	for i := 0; i < b.N; i++ {
		result = computeGo_DPFib(40)
	}
	_ = result
}

func BenchmarkCompute_DPFib_GoFixed(b *testing.B) {
	var result int64
	for i := 0; i < b.N; i++ {
		result = computeGo_FixedDPFib(40)
	}
	_ = result
}

func BenchmarkCompute_DPFib_Ual(b *testing.B) {
	var result int64
	for i := 0; i < b.N; i++ {
		result = computeUal_DPFib(40)
	}
	_ = result
}

// =============================================================================
// MATH FUNCTIONS - Tests math.* call overhead
// =============================================================================

// Pure Go: idiomatic
func computeGo_MathOps(x float64) float64 {
	return math.Sqrt(x) + math.Sin(x) + math.Cos(x) + math.Log(x+1)
}

// ual-style: same pattern
func computeUal_MathOps(x float64) float64 {
	var result float64 = math.Sqrt(x) + math.Sin(x) + math.Cos(x) + math.Log((x + 1))
	return result
}

func BenchmarkCompute_MathOps_Go(b *testing.B) {
	x := 2.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeGo_MathOps(x)
	}
	_ = result
}

func BenchmarkCompute_MathOps_Ual(b *testing.B) {
	x := 2.0
	var result float64
	for i := 0; i < b.N; i++ {
		result = computeUal_MathOps(x)
	}
	_ = result
}
