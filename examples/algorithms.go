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

func main() {
	stack_i64.PushAt(0, intToBytes(int64(10))) // var n
	stack_i64.PushAt(1, intToBytes(int64(0))) // var a
	stack_i64.PushAt(2, intToBytes(int64(1))) // var b
	stack_fib := ual.NewStack(ual.LIFO, ual.TypeInt64)
	{ v, _ := stack_i64.PeekAt(1); stack_fib.Push(v) } // push a
	stack_i64.PushAt(3, intToBytes(int64(1))) // var i
	for func() int64 { v, _ := stack_i64.PeekAt(3); return bytesToInt(v) }() < func() int64 { v, _ := stack_i64.PeekAt(0); return bytesToInt(v) }() {
		{ v, _ := stack_i64.PeekAt(2); stack_fib.Push(v) } // push b
		{ v, _ := stack_i64.PeekAt(1); stack_dstack.Push(v) } // push a
		{ v, _ := stack_i64.PeekAt(2); stack_dstack.Push(v) } // push b
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(4, v) } // let temp
		{ v, _ := stack_i64.PeekAt(2); stack_dstack.Push(v) } // push b
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(1, v) } // a = ...
		{ v, _ := stack_i64.PeekAt(4); stack_dstack.Push(v) } // push temp
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(2, v) } // b = ...
		{ v, _ := stack_i64.PeekAt(3); stack_dstack.Push(v) } // push i
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(3, v) } // i = ...
	}
	stack_i64.PushAt(5, intToBytes(int64(0))) // var temp
	{ // for @fib
		_forLen := stack_fib.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_fib.PeekAt(_forIdx)
			stack_i64.PushAt(6, _forVal) // v
			{ v, _ := stack_i64.PeekAt(6); stack_dstack.Push(v) } // push v
			{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		}
	}
	stack_i64.PushAt(7, intToBytes(int64(7))) // var num
	stack_i64.PushAt(8, intToBytes(int64(1))) // var fact
	stack_i64.PushAt(9, intToBytes(int64(1))) // var j
	for func() int64 { v, _ := stack_i64.PeekAt(9); return bytesToInt(v) }() <= func() int64 { v, _ := stack_i64.PeekAt(7); return bytesToInt(v) }() {
		{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push fact
		{ v, _ := stack_i64.PeekAt(9); stack_dstack.Push(v) } // push j
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) * bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(8, v) } // fact = ...
		{ v, _ := stack_i64.PeekAt(9); stack_dstack.Push(v) } // push j
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(9, v) } // j = ...
	}
	{ v, _ := stack_i64.PeekAt(8); stack_dstack.Push(v) } // push fact
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_data := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_data.Push(intToBytes(42))
	stack_data.Push(intToBytes(17))
	stack_data.Push(intToBytes(93))
	stack_data.Push(intToBytes(8))
	stack_data.Push(intToBytes(56))
	stack_data.Push(intToBytes(71))
	stack_data.Push(intToBytes(29))
	stack_i64.PushAt(10, intToBytes(int64(0))) // var total
	{ // for @data
		_forLen := stack_data.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_data.PeekAt(_forIdx)
			stack_i64.PushAt(11, _forVal) // v
			{ v, _ := stack_i64.PeekAt(10); stack_dstack.Push(v) } // push total
			{ v, _ := stack_i64.PeekAt(11); stack_dstack.Push(v) } // push v
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(10, v) } // total = ...
		}
	}
	{ v, _ := stack_i64.PeekAt(10); stack_dstack.Push(v) } // push total
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(12, intToBytes(int64(999999))) // var minimum
	{ // for @data
		_forLen := stack_data.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_data.PeekAt(_forIdx)
			stack_i64.PushAt(13, _forVal) // v
			if func() int64 { v, _ := stack_i64.PeekAt(13); return bytesToInt(v) }() < func() int64 { v, _ := stack_i64.PeekAt(12); return bytesToInt(v) }() {
				{ v, _ := stack_i64.PeekAt(13); stack_dstack.Push(v) } // push v
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(12, v) } // minimum = ...
			}
		}
	}
	{ v, _ := stack_i64.PeekAt(12); stack_dstack.Push(v) } // push minimum
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(14, intToBytes(int64(0))) // var maximum
	{ // for @data
		_forLen := stack_data.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_data.PeekAt(_forIdx)
			stack_i64.PushAt(15, _forVal) // v
			if func() int64 { v, _ := stack_i64.PeekAt(15); return bytesToInt(v) }() > func() int64 { v, _ := stack_i64.PeekAt(14); return bytesToInt(v) }() {
				{ v, _ := stack_i64.PeekAt(15); stack_dstack.Push(v) } // push v
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(14, v) } // maximum = ...
			}
		}
	}
	{ v, _ := stack_i64.PeekAt(14); stack_dstack.Push(v) } // push maximum
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_haystack := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_haystack.Push(intToBytes(5))
	stack_haystack.Push(intToBytes(12))
	stack_haystack.Push(intToBytes(3))
	stack_haystack.Push(intToBytes(18))
	stack_haystack.Push(intToBytes(7))
	stack_haystack.Push(intToBytes(25))
	stack_haystack.Push(intToBytes(9))
	stack_i64.PushAt(16, intToBytes(int64(18))) // var needle
	stack_i64.PushAt(17, intToBytes(int64(0))) // var found
	stack_i64.PushAt(18, intToBytes(int64(0))) // var position
	{ // for @haystack
		_forLen := stack_haystack.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_haystack.PeekAt(_forIdx)
			stack_i64.PushAt(19, intToBytes(int64(_forIdx))) // idx
			stack_i64.PushAt(20, _forVal) // val
			if func() int64 { v, _ := stack_i64.PeekAt(20); return bytesToInt(v) }() == func() int64 { v, _ := stack_i64.PeekAt(16); return bytesToInt(v) }() {
				stack_dstack.Push(intToBytes(1))
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(17, v) } // found = ...
				{ v, _ := stack_i64.PeekAt(19); stack_dstack.Push(v) } // push idx
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(18, v) } // position = ...
			}
		}
	}
	if func() int64 { v, _ := stack_i64.PeekAt(17); return bytesToInt(v) }() > int64(0) {
		{ v, _ := stack_i64.PeekAt(18); stack_dstack.Push(v) } // push position
		{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	} else {
		stack_dstack.Push(intToBytes(999))
		{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	}
	stack_original := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_original.Push(intToBytes(1))
	stack_original.Push(intToBytes(2))
	stack_original.Push(intToBytes(3))
	stack_original.Push(intToBytes(4))
	stack_original.Push(intToBytes(5))
	stack_reversed := ual.NewStack(ual.LIFO, ual.TypeInt64)
	{ // for @original
		_forLen := stack_original.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_original.PeekAt(_forIdx)
			stack_i64.PushAt(21, _forVal) // v
			{ v, _ := stack_i64.PeekAt(21); stack_reversed.Push(v) } // push v
		}
	}
	{ // for @reversed
		_forLen := stack_reversed.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_reversed.PeekAt(_forIdx)
			stack_i64.PushAt(22, _forVal) // v
			{ v, _ := stack_i64.PeekAt(22); stack_dstack.Push(v) } // push v
			{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		}
	}
	stack_nums := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_nums.Push(intToBytes(10))
	stack_nums.Push(intToBytes(25))
	stack_nums.Push(intToBytes(30))
	stack_nums.Push(intToBytes(15))
	stack_nums.Push(intToBytes(45))
	stack_nums.Push(intToBytes(20))
	stack_nums.Push(intToBytes(35))
	stack_i64.PushAt(23, intToBytes(int64(0))) // var countOver20
	{ // for @nums
		_forLen := stack_nums.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_nums.PeekAt(_forIdx)
			stack_i64.PushAt(24, _forVal) // v
			if func() int64 { v, _ := stack_i64.PeekAt(24); return bytesToInt(v) }() > int64(20) {
				{ v, _ := stack_i64.PeekAt(23); stack_dstack.Push(v) } // push countOver20
				{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(23, v) } // countOver20 = ...
			}
		}
	}
	{ v, _ := stack_i64.PeekAt(23); stack_dstack.Push(v) } // push countOver20
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_source := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_source.Push(intToBytes(5))
	stack_source.Push(intToBytes(15))
	stack_source.Push(intToBytes(8))
	stack_source.Push(intToBytes(22))
	stack_source.Push(intToBytes(3))
	stack_source.Push(intToBytes(19))
	stack_source.Push(intToBytes(12))
	stack_filtered := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(25, intToBytes(int64(10))) // var threshold
	{ // for @source
		_forLen := stack_source.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_source.PeekAt(_forIdx)
			stack_i64.PushAt(26, _forVal) // v
			if func() int64 { v, _ := stack_i64.PeekAt(26); return bytesToInt(v) }() > func() int64 { v, _ := stack_i64.PeekAt(25); return bytesToInt(v) }() {
				{ v, _ := stack_i64.PeekAt(26); stack_filtered.Push(v) } // push v
			}
		}
	}
	{ // for @filtered
		_forLen := stack_filtered.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_filtered.PeekAt(_forIdx)
			stack_i64.PushAt(27, _forVal) // v
			{ v, _ := stack_i64.PeekAt(27); stack_dstack.Push(v) } // push v
			{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		}
	}
	stack_input := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_input.Push(intToBytes(2))
	stack_input.Push(intToBytes(4))
	stack_input.Push(intToBytes(6))
	stack_input.Push(intToBytes(8))
	stack_input.Push(intToBytes(10))
	stack_doubled := ual.NewStack(ual.LIFO, ual.TypeInt64)
	{ // for @input
		_forLen := stack_input.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_input.PeekAt(_forIdx)
			stack_i64.PushAt(28, _forVal) // v
			{ v, _ := stack_i64.PeekAt(28); stack_dstack.Push(v) } // push v
			stack_dstack.Push(intToBytes(2))
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) * bytesToInt(b))) }
			{ v, _ := stack_doubled.Pop(); stack_dstack.Push(v) }
		}
	}
	stack_input2 := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_input2.Push(intToBytes(1))
	stack_input2.Push(intToBytes(2))
	stack_input2.Push(intToBytes(3))
	stack_input2.Push(intToBytes(4))
	stack_input2.Push(intToBytes(5))
	stack_mapped := ual.NewStack(ual.LIFO, ual.TypeInt64)
	{ // for @input2
		_forLen := stack_input2.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_input2.PeekAt(_forIdx)
			stack_i64.PushAt(29, _forVal) // v
			{ v, _ := stack_i64.PeekAt(29); stack_dstack.Push(v) } // push v
			stack_dstack.Push(intToBytes(2))
			{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) * bytesToInt(b))) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(30, v) } // let doubled2
			{ v, _ := stack_i64.PeekAt(30); stack_mapped.Push(v) } // push doubled2
		}
	}
	stack_i64.PushAt(31, intToBytes(int64(0))) // var doubled2
	{ // for @mapped
		_forLen := stack_mapped.Len()
		for _forIdx := 0; _forIdx < _forLen; _forIdx++ {
			_forVal, _ := stack_mapped.PeekAt(_forIdx)
			stack_i64.PushAt(32, _forVal) // v
			{ v, _ := stack_i64.PeekAt(32); stack_dstack.Push(v) } // push v
			{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		}
	}
	stack_i64.PushAt(33, intToBytes(int64(48))) // var gcd_a
	stack_i64.PushAt(34, intToBytes(int64(18))) // var gcd_b
	for func() int64 { v, _ := stack_i64.PeekAt(34); return bytesToInt(v) }() > int64(0) {
		{ v, _ := stack_i64.PeekAt(33); stack_dstack.Push(v) } // push gcd_a
		{ v, _ := stack_i64.PeekAt(34); stack_dstack.Push(v) } // push gcd_b
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) % bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(35, v) } // let temp2
		{ v, _ := stack_i64.PeekAt(34); stack_dstack.Push(v) } // push gcd_b
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(33, v) } // gcd_a = ...
		{ v, _ := stack_i64.PeekAt(35); stack_dstack.Push(v) } // push temp2
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(34, v) } // gcd_b = ...
	}
	stack_i64.PushAt(36, intToBytes(int64(0))) // var temp2
	{ v, _ := stack_i64.PeekAt(33); stack_dstack.Push(v) } // push gcd_a
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(37, intToBytes(int64(17))) // var checkNum
	stack_i64.PushAt(38, intToBytes(int64(1))) // var isPrime
	stack_i64.PushAt(39, intToBytes(int64(2))) // var divisor
	for func() int64 { v, _ := stack_i64.PeekAt(39); return bytesToInt(v) }() < func() int64 { v, _ := stack_i64.PeekAt(37); return bytesToInt(v) }() {
		{ v, _ := stack_i64.PeekAt(37); stack_dstack.Push(v) } // push checkNum
		{ v, _ := stack_i64.PeekAt(39); stack_dstack.Push(v) } // push divisor
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) % bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(40, v) } // let remainder
		if func() int64 { v, _ := stack_i64.PeekAt(40); return bytesToInt(v) }() == int64(0) {
			stack_dstack.Push(intToBytes(0))
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(38, v) } // isPrime = ...
			break
		}
		{ v, _ := stack_i64.PeekAt(39); stack_dstack.Push(v) } // push divisor
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(39, v) } // divisor = ...
	}
	stack_i64.PushAt(41, intToBytes(int64(0))) // var remainder
	{ v, _ := stack_i64.PeekAt(38); stack_dstack.Push(v) } // push isPrime
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(42, intToBytes(int64(2))) // var base
	stack_i64.PushAt(43, intToBytes(int64(10))) // var exp
	stack_i64.PushAt(44, intToBytes(int64(1))) // var result
	for func() int64 { v, _ := stack_i64.PeekAt(43); return bytesToInt(v) }() > int64(0) {
		{ v, _ := stack_i64.PeekAt(44); stack_dstack.Push(v) } // push result
		{ v, _ := stack_i64.PeekAt(42); stack_dstack.Push(v) } // push base
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) * bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(44, v) } // result = ...
		{ v, _ := stack_i64.PeekAt(43); stack_dstack.Push(v) } // push exp
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) - 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(43, v) } // exp = ...
	}
	{ v, _ := stack_i64.PeekAt(44); stack_dstack.Push(v) } // push result
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
