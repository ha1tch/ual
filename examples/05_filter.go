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
	stack_source := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(0, intToBytes(int64(0))) // var k
	stack_i64.PushAt(1, intToBytes(int64(0))) // var val
	for func() int64 { v, _ := stack_i64.PeekAt(0); return bytesToInt(v) }() < int64(500) {
		{ v, _ := stack_i64.PeekAt(0); stack_dstack.Push(v) } // push k
		stack_dstack.Push(intToBytes(7))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) * bytesToInt(b))) }
		stack_dstack.Push(intToBytes(100))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) % bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(1, v) } // val = ...
		{ v, _ := stack_i64.PeekAt(1); stack_source.Push(v) } // push val
		{ v, _ := stack_i64.PeekAt(0); stack_dstack.Push(v) } // push k
		{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
		{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(0, v) } // k = ...
	}
	stack_filtered := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(2, intToBytes(int64(50))) // var threshold
	stack_i64.PushAt(3, intToBytes(int64(0))) // var count
	{ // for @source
		_forLen := stack_source.Len()
		for _forIdx := _forLen - 1; _forIdx >= 0; _forIdx-- {
			_forVal, _ := stack_source.PeekAt(_forIdx)
			stack_i64.PushAt(4, _forVal) // v
			if func() int64 { v, _ := stack_i64.PeekAt(4); return bytesToInt(v) }() > func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }() {
				{ v, _ := stack_i64.PeekAt(4); stack_filtered.Push(v) } // push v
				{ v, _ := stack_i64.PeekAt(3); stack_dstack.Push(v) } // push count
				{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
				{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(3, v) } // count = ...
			}
		}
	}
	{ v, _ := stack_i64.PeekAt(3); stack_dstack.Push(v) } // push count
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
