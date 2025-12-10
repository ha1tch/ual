package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"
	"unsafe"
	
	ual "github.com/ha1tch/ual"
)

// Helper functions
func intToBytes(n int64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(n & 0xff)
		n >>= 8
	}
	return b
}

func bytesToInt(b []byte) int64 {
	var n int64
	for _, v := range b {
		n = (n << 8) | int64(v)
	}
	return n
}

func uintToBytes(n uint64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(n & 0xff)
		n >>= 8
	}
	return b
}

func floatToBytes(f float64) []byte {
	bits := *(*uint64)(unsafe.Pointer(&f))
	return intToBytes(int64(bits))
}

func bytesToFloat(b []byte) float64 {
	bits := uint64(bytesToInt(b))
	return *(*float64)(unsafe.Pointer(&bits))
}

func boolToBytes(v bool) []byte {
	if v { return []byte{1} }
	return []byte{0}
}

func bytesToBool(b []byte) bool {
	return len(b) > 0 && b[0] != 0
}

func absInt(n int64) int64 {
	if n < 0 { return -n }
	return n
}

func minInt(a, b int64) int64 {
	if a < b { return a }
	return b
}

func maxInt(a, b int64) int64 {
	if a > b { return a }
	return b
}

// Select helper: creates cancellable context
func _selectContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

var _ = time.Second // suppress unused import
var _ = math.Pi // suppress unused import
var _ = binary.LittleEndian // suppress unused import

// Global stacks
var stack_dstack = ual.NewStack(ual.LIFO, ual.TypeInt64)
var stack_rstack = ual.NewStack(ual.LIFO, ual.TypeInt64)
var stack_bool = ual.NewStack(ual.LIFO, ual.TypeBool)
var stack_error = ual.NewStack(ual.LIFO, ual.TypeBytes)

// Spawn task queue
var spawn_tasks []func()
var spawn_mu sync.Mutex

// Global status for consider blocks
var _consider_status = "ok"
var _consider_value interface{}

// Type stacks for variables
var stack_i64 = ual.NewStack(ual.Hash, ual.TypeInt64)
var stack_u64 = ual.NewStack(ual.Hash, ual.TypeUint64)
var stack_f64 = ual.NewStack(ual.Hash, ual.TypeFloat64)
var stack_string = ual.NewStack(ual.Hash, ual.TypeString)
var stack_bytes = ual.NewStack(ual.Hash, ual.TypeBytes)

func addOne(x int64) int64 {
	stack_i64.PushAt(0, intToBytes(int64(x))) // param x
	{ v, _ := stack_i64.PeekAt(0); stack_dstack.Push(v) } // push x
	stack_dstack.Push(intToBytes(1))
	{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
	{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(1, v) } // let r
	// Error: variable r already declared in this scope
	return func() int64 { v, _ := stack_i64.PeekAt(1); return bytesToInt(v) }()
}

func addTen(x int64) int64 {
	stack_i64.PushAt(2, intToBytes(int64(x))) // param x
	{ v, _ := stack_i64.PeekAt(2); stack_dstack.Push(v) } // push x
	stack_dstack.Push(intToBytes(10))
	{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
	{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(3, v) } // let r
	// Error: variable r already declared in this scope
	return func() int64 { v, _ := stack_i64.PeekAt(3); return bytesToInt(v) }()
}

func main() {
	stack_bench1 := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(4, intToBytes(int64(0))) // var b1
	for func() int64 { v, _ := stack_i64.PeekAt(4); return bytesToInt(v) }() < int64(10000) {
		{ v, _ := stack_i64.PeekAt(4); stack_bench1.Push(v) } // push b1
		{ v, _ := stack_i64.PeekAt(4); stack_dstack.Push(v) } // push b1
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(4, v) } // b1 = ...
	}
	stack_i64.PushAt(5, intToBytes(int64(0))) // var sum1
	{ // for @bench1
		_forLen := stack_bench1.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_bench1.PeekAt(_forIdx)
			stack_i64.PushAt(6, _forVal) // v
			{ v, _ := stack_i64.PeekAt(5); stack_dstack.Push(v) } // push sum1
			{ v, _ := stack_i64.PeekAt(6); stack_dstack.Push(v) } // push v
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(5, v) } // sum1 = ...
		}
	}
	{ v, _ := stack_i64.PeekAt(5); stack_dstack.Push(v) } // push sum1
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(7, intToBytes(int64(0))) // var b2
	for func() int64 { v, _ := stack_i64.PeekAt(7); return bytesToInt(v) }() < int64(100) {
		stack_dstack.Push(intToBytes(1))
		stack_dstack.Push(intToBytes(2))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		stack_dstack.Push(intToBytes(3))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) * bytesToInt(b))) }
		stack_dstack.Push(intToBytes(4))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) - bytesToInt(b))) }
		stack_dstack.Push(intToBytes(5))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) / bytesToInt(b))) }
		stack_dstack.Pop()
		{ v, _ := stack_i64.PeekAt(7); stack_dstack.Push(v) } // push b2
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(7, v) } // b2 = ...
	}
	stack_i64.PushAt(8, intToBytes(int64(0))) // var acc
	stack_i64.PushAt(9, intToBytes(int64(0))) // var b3
	for func() int64 { v, _ := stack_i64.PeekAt(9); return bytesToInt(v) }() < int64(100) {
		{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push acc
		stack_dstack.Push(intToBytes(1))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(8, v) } // acc = ...
		{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push acc
		stack_dstack.Push(intToBytes(2))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(8, v) } // acc = ...
		{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push acc
		stack_dstack.Push(intToBytes(3))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(8, v) } // acc = ...
		{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push acc
		stack_dstack.Push(intToBytes(4))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(8, v) } // acc = ...
		{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push acc
		stack_dstack.Push(intToBytes(5))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(8, v) } // acc = ...
		{ v, _ := stack_i64.PeekAt(9); stack_dstack.Push(v) } // push b3
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(9, v) } // b3 = ...
	}
	{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push acc
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_nums := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(10, intToBytes(int64(0))) // var b4
	for func() int64 { v, _ := stack_i64.PeekAt(10); return bytesToInt(v) }() < int64(1000) {
		{ v, _ := stack_i64.PeekAt(10); stack_nums.Push(v) } // push b4
		{ v, _ := stack_i64.PeekAt(10); stack_dstack.Push(v) } // push b4
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(10, v) } // b4 = ...
	}
	stack_i64.PushAt(11, intToBytes(int64(0))) // var sumLifo
	{ // for @nums
		_forLen := stack_nums.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_nums.PeekAt(_forIdx)
			stack_i64.PushAt(12, _forVal) // v
			{ v, _ := stack_i64.PeekAt(11); stack_dstack.Push(v) } // push sumLifo
			{ v, _ := stack_i64.PeekAt(12); stack_dstack.Push(v) } // push v
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(11, v) } // sumLifo = ...
		}
	}
	stack_i64.PushAt(13, intToBytes(int64(0))) // var sumFifo
	{ // for @nums
		_forLen := stack_nums.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_nums.PeekAt(_forIdx)
			stack_i64.PushAt(14, _forVal) // v
			{ v, _ := stack_i64.PeekAt(13); stack_dstack.Push(v) } // push sumFifo
			{ v, _ := stack_i64.PeekAt(14); stack_dstack.Push(v) } // push v
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(13, v) } // sumFifo = ...
		}
	}
	{ v, _ := stack_i64.PeekAt(11); stack_dstack.Push(v) } // push sumLifo
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	{ v, _ := stack_i64.PeekAt(13); stack_dstack.Push(v) } // push sumFifo
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(15, intToBytes(int64(0))) // var fcall
	stack_i64.PushAt(16, intToBytes(int64(0))) // var b5
	for func() int64 { v, _ := stack_i64.PeekAt(16); return bytesToInt(v) }() < int64(1000) {
		{ v, _ := stack_i64.PeekAt(15); stack_dstack.Push(v) } // push fcall
		stack_dstack.Push(intToBytes(1))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(15, v) } // fcall = ...
		{ v, _ := stack_i64.PeekAt(16); stack_dstack.Push(v) } // push b5
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(16, v) } // b5 = ...
	}
	{ v, _ := stack_i64.PeekAt(15); stack_dstack.Push(v) } // push fcall
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(17, intToBytes(int64(addOne(99)))) // var result1
	{ v, _ := stack_i64.PeekAt(17); stack_dstack.Push(v) } // push result1
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(18, intToBytes(int64(addTen(addOne(0))))) // var result2
	{ v, _ := stack_i64.PeekAt(18); stack_dstack.Push(v) } // push result2
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_data := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(19, intToBytes(int64(1))) // var b7
	for func() int64 { v, _ := stack_i64.PeekAt(19); return bytesToInt(v) }() <= int64(100) {
		{ v, _ := stack_i64.PeekAt(19); stack_data.Push(v) } // push b7
		{ v, _ := stack_i64.PeekAt(19); stack_dstack.Push(v) } // push b7
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(19, v) } // b7 = ...
	}
	stack_i64.PushAt(20, intToBytes(int64(0))) // var manualSum
	{ // for @data
		_forLen := stack_data.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_data.PeekAt(_forIdx)
			stack_i64.PushAt(21, _forVal) // v
			{ v, _ := stack_i64.PeekAt(20); stack_dstack.Push(v) } // push manualSum
			{ v, _ := stack_i64.PeekAt(21); stack_dstack.Push(v) } // push v
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(20, v) } // manualSum = ...
		}
	}
	{ v, _ := stack_i64.PeekAt(20); stack_dstack.Push(v) } // push manualSum
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_small := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_small.Push(intToBytes(1))
	stack_small.Push(intToBytes(2))
	stack_small.Push(intToBytes(3))
	stack_small.Push(intToBytes(4))
	stack_small.Push(intToBytes(5))
	stack_i64.PushAt(22, intToBytes(int64(1))) // var manualProd
	{ // for @small
		_forLen := stack_small.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_small.PeekAt(_forIdx)
			stack_i64.PushAt(23, _forVal) // v
			{ v, _ := stack_i64.PeekAt(22); stack_dstack.Push(v) } // push manualProd
			{ v, _ := stack_i64.PeekAt(23); stack_dstack.Push(v) } // push v
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) * bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(22, v) } // manualProd = ...
		}
	}
	{ v, _ := stack_i64.PeekAt(22); stack_dstack.Push(v) } // push manualProd
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_source := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(24, intToBytes(int64(0))) // var b8
	for func() int64 { v, _ := stack_i64.PeekAt(24); return bytesToInt(v) }() < int64(100) {
		{ v, _ := stack_i64.PeekAt(24); stack_source.Push(v) } // push b8
		{ v, _ := stack_i64.PeekAt(24); stack_dstack.Push(v) } // push b8
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(24, v) } // b8 = ...
	}
	stack_evens := ual.NewStack(ual.LIFO, ual.TypeInt64)
	{ // for @source
		_forLen := stack_source.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_source.PeekAt(_forIdx)
			stack_i64.PushAt(25, _forVal) // v
			{ v, _ := stack_i64.PeekAt(25); stack_dstack.Push(v) } // push v
			stack_dstack.Push(intToBytes(2))
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) % bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(26, v) } // let rem
			// Error: variable rem already declared in this scope
			if func() int64 { v, _ := stack_i64.PeekAt(26); return bytesToInt(v) }() == int64(0) {
				{ v, _ := stack_i64.PeekAt(25); stack_evens.Push(v) } // push v
			}
		}
	}
	stack_i64.PushAt(27, intToBytes(int64(0))) // var evenCount
	{ // for @evens
		_forLen := stack_evens.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_evens.PeekAt(_forIdx)
			stack_i64.PushAt(28, _forVal) // v
			{ v, _ := stack_i64.PeekAt(27); stack_dstack.Push(v) } // push evenCount
			{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(27, v) } // evenCount = ...
		}
	}
	{ v, _ := stack_i64.PeekAt(27); stack_dstack.Push(v) } // push evenCount
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(29, intToBytes(int64(0))) // var nested
	stack_i64.PushAt(30, intToBytes(int64(0))) // var i
	for func() int64 { v, _ := stack_i64.PeekAt(30); return bytesToInt(v) }() < int64(100) {
		stack_i64.PushAt(31, intToBytes(int64(0))) // var j
		for func() int64 { v, _ := stack_i64.PeekAt(31); return bytesToInt(v) }() < int64(100) {
			{ v, _ := stack_i64.PeekAt(29); stack_dstack.Push(v) } // push nested
			{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(29, v) } // nested = ...
			{ v, _ := stack_i64.PeekAt(31); stack_dstack.Push(v) } // push j
			{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(31, v) } // j = ...
		}
		{ v, _ := stack_i64.PeekAt(30); stack_dstack.Push(v) } // push i
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(30, v) } // i = ...
	}
	{ v, _ := stack_i64.PeekAt(29); stack_dstack.Push(v) } // push nested
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(32, intToBytes(int64(0))) // var b9
	for func() int64 { v, _ := stack_i64.PeekAt(32); return bytesToInt(v) }() < int64(1000) {
		{ v, _ := stack_i64.PeekAt(32); stack_dstack.Push(v) } // push b9
		stack_dstack.Push(intToBytes(255))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) & bytesToInt(b))) }
		{ v, _ := stack_i64.PeekAt(32); stack_dstack.Push(v) } // push b9
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) ^ bytesToInt(b))) }
		stack_dstack.Push(intToBytes(3))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) << uint(bytesToInt(b)))) }
		stack_dstack.Push(intToBytes(1))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) >> uint(bytesToInt(b)))) }
		stack_dstack.Pop()
		{ v, _ := stack_i64.PeekAt(32); stack_dstack.Push(v) } // push b9
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(32, v) } // b9 = ...
	}
	stack_dstack.Push(intToBytes(1))
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_vals := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_vals.Push(intToBytes(42))
	stack_vals.Push(intToBytes(17))
	stack_vals.Push(intToBytes(99))
	stack_vals.Push(intToBytes(3))
	stack_vals.Push(intToBytes(88))
	stack_vals.Push(intToBytes(56))
	stack_vals.Push(intToBytes(71))
	stack_vals.Push(intToBytes(23))
	stack_vals.Push(intToBytes(45))
	stack_vals.Push(intToBytes(12))
	stack_i64.PushAt(33, intToBytes(int64(999999))) // var minVal
	stack_i64.PushAt(34, intToBytes(int64(0))) // var maxVal
	{ // for @vals
		_forLen := stack_vals.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_vals.PeekAt(_forIdx)
			stack_i64.PushAt(35, _forVal) // v
			if func() int64 { v, _ := stack_i64.PeekAt(35); return bytesToInt(v) }() < func() int64 { v, _ := stack_i64.PeekAt(33); return bytesToInt(v) }() {
				{ v, _ := stack_i64.PeekAt(35); stack_dstack.Push(v) } // push v
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(33, v) } // minVal = ...
			}
			if func() int64 { v, _ := stack_i64.PeekAt(35); return bytesToInt(v) }() > func() int64 { v, _ := stack_i64.PeekAt(34); return bytesToInt(v) }() {
				{ v, _ := stack_i64.PeekAt(35); stack_dstack.Push(v) } // push v
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(34, v) } // maxVal = ...
			}
		}
	}
	{ v, _ := stack_i64.PeekAt(33); stack_dstack.Push(v) } // push minVal
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	{ v, _ := stack_i64.PeekAt(34); stack_dstack.Push(v) } // push maxVal
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_tasks := ual.NewCappedStack(ual.LIFO, ual.TypeInt64, 1000)
	stack_i64.PushAt(36, intToBytes(int64(0))) // var t
	for func() int64 { v, _ := stack_i64.PeekAt(36); return bytesToInt(v) }() < int64(500) {
		{ v, _ := stack_i64.PeekAt(36); stack_tasks.Push(v) } // push t
		{ v, _ := stack_i64.PeekAt(36); stack_dstack.Push(v) } // push t
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(36, v) } // t = ...
	}
	stack_i64.PushAt(37, intToBytes(int64(0))) // var ownerWork
	stack_i64.PushAt(38, intToBytes(int64(0))) // var ow
	for func() int64 { v, _ := stack_i64.PeekAt(38); return bytesToInt(v) }() < int64(100) {
		{ v, _ := stack_tasks.Pop(); stack_dstack.Push(v) }
		{ v, _ := stack_tasks.Pop(); stack_i64.PushAt(39, v) } // let task
		// Error: variable task already declared in this scope
		{ v, _ := stack_i64.PeekAt(37); stack_dstack.Push(v) } // push ownerWork
		{ v, _ := stack_i64.PeekAt(39); stack_dstack.Push(v) } // push task
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(37, v) } // ownerWork = ...
		{ v, _ := stack_i64.PeekAt(38); stack_dstack.Push(v) } // push ow
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(38, v) } // ow = ...
	}
	stack_i64.PushAt(40, intToBytes(int64(0))) // var thiefWork
	stack_i64.PushAt(41, intToBytes(int64(0))) // var tw
	{ // for @tasks
		_forLen := stack_tasks.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_tasks.PeekAt(_forIdx)
			stack_i64.PushAt(42, _forVal) // v
			if func() int64 { v, _ := stack_i64.PeekAt(41); return bytesToInt(v) }() < int64(100) {
				{ v, _ := stack_i64.PeekAt(40); stack_dstack.Push(v) } // push thiefWork
				{ v, _ := stack_i64.PeekAt(42); stack_dstack.Push(v) } // push v
				{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(40, v) } // thiefWork = ...
				{ v, _ := stack_i64.PeekAt(41); stack_dstack.Push(v) } // push tw
				{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(41, v) } // tw = ...
			}
		}
	}
	{ v, _ := stack_i64.PeekAt(37); stack_dstack.Push(v) } // push ownerWork
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	{ v, _ := stack_i64.PeekAt(40); stack_dstack.Push(v) } // push thiefWork
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	
	_ = ual.LIFO
	var _ = unsafe.Pointer(nil)
	_ = stack_dstack
	_ = stack_rstack
	_ = stack_bool
	_ = stack_error
	_ = stack_i64
	_ = stack_u64
	_ = stack_f64
	_ = stack_string
	_ = stack_bytes
}
