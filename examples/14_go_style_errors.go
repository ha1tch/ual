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

func safeDivide(a int64, b int64) int64 {
	stack_i64.PushAt(0, intToBytes(int64(a))) // param a
	stack_i64.PushAt(1, intToBytes(int64(b))) // param b
	if func() int64 { v, _ := stack_i64.PeekAt(1); return bytesToInt(v) }() == int64(0) {
		stack_error.Push([]byte("division by zero"))
		return 0
	}
	return (func() int64 { v, _ := stack_i64.PeekAt(0); return bytesToInt(v) }() / func() int64 { v, _ := stack_i64.PeekAt(1); return bytesToInt(v) }())
}

func mustBePositive(n int64) int64 {
	stack_i64.PushAt(2, intToBytes(int64(n))) // param n
	if func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }() <= int64(0) {
		panic("value must be positive")
	}
	return func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }()
}

func processWithCleanup() int64 {
	defer func() {
		stack_dstack.Push(intToBytes(777))
		{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	}()
	stack_i64.PushAt(3, intToBytes(int64(100))) // var value
	{ v, _ := stack_i64.PeekAt(3); stack_dstack.Push(v) } // push value
	stack_dstack.Push(intToBytes(50))
	{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
	{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(3, v) } // value = ...
	return func() int64 { v, _ := stack_i64.PeekAt(3); return bytesToInt(v) }()
}

func main() {
	stack_i64.PushAt(4, intToBytes(int64(safeDivide(100, 5)))) // var x
	{ v, _ := stack_i64.PeekAt(4); stack_dstack.Push(v) } // push x
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(5, intToBytes(int64(safeDivide(100, 0)))) // var y
	{ v, _ := stack_i64.PeekAt(5); stack_dstack.Push(v) } // push y
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_bool.Push(boolToBytes(stack_error.Len() > 0))
	for stack_error.Len() > 0 { stack_error.Pop() }
	stack_i64.PushAt(6, intToBytes(int64(safeDivide(50, 2)))) // var z
	{ v, _ := stack_i64.PeekAt(6); stack_dstack.Push(v) } // push z
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(7, intToBytes(int64(mustBePositive(42)))) // var good
	{ v, _ := stack_i64.PeekAt(7); stack_dstack.Push(v) } // push good
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_i64.PushAt(8, intToBytes(int64((0 - 10)))) // var negVal
	func() {
		var _recovered interface{}
		defer func() {
			if r := recover(); r != nil {
				_recovered = r
				err := fmt.Sprintf("%v", r)
				_ = err // suppress unused warning
				stack_dstack.Push(intToBytes(0))
				{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
			}
			_ = _recovered
		}()
		stack_i64.PushAt(9, intToBytes(int64(mustBePositive(func() int64 { v, _ := stack_i64.PeekAt(8); return bytesToInt(v) }())))) // var result
		{ v, _ := stack_i64.PeekAt(9); stack_dstack.Push(v) } // push result
		{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	}()
	stack_i64.PushAt(10, intToBytes(int64(processWithCleanup()))) // var final
	{ v, _ := stack_i64.PeekAt(10); stack_dstack.Push(v) } // push final
	{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
	stack_dstack.Push(intToBytes(888))
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
