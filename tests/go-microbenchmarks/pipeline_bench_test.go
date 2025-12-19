// Full Pipeline Benchmarks
//
// These benchmarks measure the COMPLETE ual pattern including:
// - Lock/Unlock
// - PopRaw + byte conversion (input)
// - Computation
// - PushRaw + byte conversion (output)
//
// This shows the total cost of using compute() blocks,
// helping identify when the "byte tax" matters.

package main

import (
	"encoding/binary"
	"math"
	"sync"
	"testing"
)

// =============================================================================
// HELPERS - Simulate ual runtime overhead
// =============================================================================

// Byte conversion (exact copy of generated code pattern)
func floatToBytes(f float64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(f))
	return b
}

func bytesToFloat(b []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

func intToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

func bytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

// Minimal stack simulation for benchmarking
type benchStack struct {
	mu   sync.Mutex
	data [][]byte
}

func newBenchStack() *benchStack {
	return &benchStack{data: make([][]byte, 0, 16)}
}

func (s *benchStack) Lock()   { s.mu.Lock() }
func (s *benchStack) Unlock() { s.mu.Unlock() }

func (s *benchStack) PushRaw(b []byte) {
	s.data = append(s.data, b)
}

func (s *benchStack) PopRaw() []byte {
	n := len(s.data)
	if n == 0 {
		return nil
	}
	b := s.data[n-1]
	s.data = s.data[:n-1]
	return b
}

func (s *benchStack) Clear() {
	s.data = s.data[:0]
}

// =============================================================================
// MANDELBROT - Full pipeline
// =============================================================================

func BenchmarkPipeline_Mandelbrot(b *testing.B) {
	stack := newBenchStack()
	cr_bytes := floatToBytes(0.25)
	ci_bytes := floatToBytes(0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Push inputs
		stack.PushRaw(cr_bytes)
		stack.PushRaw(ci_bytes)

		// Compute block
		func() {
			stack.Lock()
			defer stack.Unlock()

			ci := bytesToFloat(stack.PopRaw())
			cr := bytesToFloat(stack.PopRaw())

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
					stack.PushRaw(floatToBytes(iter))
					return
				}
				zi = ((2 * zr) * zi) + ci
				zr = (zr2 - zi2) + cr
				iter = iter + 1
			}
			stack.PushRaw(floatToBytes(max_iter))
		}()

		// Pop result
		_ = stack.PopRaw()
	}
}

// =============================================================================
// INTEGRATE - Full pipeline
// =============================================================================

func BenchmarkPipeline_Integrate(b *testing.B) {
	stack := newBenchStack()
	a_bytes := floatToBytes(0.0)
	b_bytes := floatToBytes(1.0)
	n_bytes := floatToBytes(1000.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.PushRaw(a_bytes)
		stack.PushRaw(b_bytes)
		stack.PushRaw(n_bytes)

		func() {
			stack.Lock()
			defer stack.Unlock()

			n := bytesToFloat(stack.PopRaw())
			bval := bytesToFloat(stack.PopRaw())
			a := bytesToFloat(stack.PopRaw())

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
			stack.PushRaw(floatToBytes(result))
		}()

		_ = stack.PopRaw()
	}
}

// =============================================================================
// LEIBNIZ - Full pipeline
// =============================================================================

func BenchmarkPipeline_Leibniz(b *testing.B) {
	stack := newBenchStack()
	terms_bytes := floatToBytes(100000.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.PushRaw(terms_bytes)

		func() {
			stack.Lock()
			defer stack.Unlock()

			terms := bytesToFloat(stack.PopRaw())

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
			stack.PushRaw(floatToBytes(pi))
		}()

		_ = stack.PopRaw()
	}
}

// =============================================================================
// NEWTON - Full pipeline
// =============================================================================

func BenchmarkPipeline_Newton(b *testing.B) {
	stack := newBenchStack()
	x_bytes := floatToBytes(2.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.PushRaw(x_bytes)

		func() {
			stack.Lock()
			defer stack.Unlock()

			x := bytesToFloat(stack.PopRaw())

			var guess float64 = x / 2
			var iter float64 = 0
			for iter < 20 {
				guess = (guess + (x / guess)) / 2
				iter = iter + 1
			}
			stack.PushRaw(floatToBytes(guess))
		}()

		_ = stack.PopRaw()
	}
}

// =============================================================================
// ARRAY SUM - Full pipeline with local array
// =============================================================================

func BenchmarkPipeline_ArraySum(b *testing.B) {
	stack := newBenchStack()
	n_bytes := intToBytes(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.PushRaw(n_bytes)

		func() {
			stack.Lock()
			defer stack.Unlock()

			n := bytesToInt(stack.PopRaw())

			var buf [100]int64
			var idx int64 = 0
			for idx < n {
				buf[int(idx)] = (idx + 1)
				idx = (idx + 1)
			}

			var sum int64 = 0
			idx = 0
			for idx < n {
				sum = (sum + buf[int(idx)])
				idx = (idx + 1)
			}
			stack.PushRaw(intToBytes(sum))
		}()

		_ = stack.PopRaw()
	}
}

// =============================================================================
// DP FIB - Full pipeline with local array
// =============================================================================

func BenchmarkPipeline_DPFib(b *testing.B) {
	stack := newBenchStack()
	n_bytes := intToBytes(40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.PushRaw(n_bytes)

		func() {
			stack.Lock()
			defer stack.Unlock()

			n := bytesToInt(stack.PopRaw())

			var dp [100]int64
			dp[0] = 0
			dp[1] = 1
			var i int64 = 2
			for i <= n {
				dp[int(i)] = (dp[int((i - 1))] + dp[int((i - 2))])
				i = (i + 1)
			}
			stack.PushRaw(intToBytes(dp[int(n)]))
		}()

		_ = stack.PopRaw()
	}
}

// =============================================================================
// OVERHEAD ISOLATION - Measure just the stack/lock/byte overhead
// =============================================================================

func BenchmarkOverhead_LockUnlock(b *testing.B) {
	stack := newBenchStack()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.Lock()
		stack.Unlock()
	}
}

func BenchmarkOverhead_ByteConvert_Float(b *testing.B) {
	x := 3.14159
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes := floatToBytes(x)
		_ = bytesToFloat(bytes)
	}
}

func BenchmarkOverhead_ByteConvert_Int(b *testing.B) {
	x := int64(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes := intToBytes(x)
		_ = bytesToInt(bytes)
	}
}

func BenchmarkOverhead_PushPopRaw(b *testing.B) {
	stack := newBenchStack()
	data := floatToBytes(3.14159)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.PushRaw(data)
		_ = stack.PopRaw()
	}
}

func BenchmarkOverhead_FullCycle(b *testing.B) {
	// Measures: lock + pop + convert + convert + push + unlock
	// This is the "byte tax" for a single input/output
	stack := newBenchStack()
	input := floatToBytes(3.14159)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.PushRaw(input)

		func() {
			stack.Lock()
			defer stack.Unlock()
			x := bytesToFloat(stack.PopRaw())
			stack.PushRaw(floatToBytes(x))
		}()

		_ = stack.PopRaw()
	}
}
