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
	stack_calc := ual.NewStack(ual.LIFO, ual.TypeFloat64)
	stack_calc.Push(floatToBytes(2.000000))
	func() {
		stack_calc.Lock()
		defer stack_calc.Unlock()
		_bytes_S, _err_S := stack_calc.PopRaw()
		if _err_S != nil { panic(_err_S) }
		var S float64 = bytesToFloat(_bytes_S)
		var x float64 = (S / 2.0)
		var prev float64 = 0.0
		var diff float64 = 1.0
		var epsilon float64 = 1e-07
		var half float64 = 0.5
		var iterations float64 = 0.0
		for (diff > epsilon) {
			prev = x
			x = (half * (x + (S / x)))
			diff = (x - prev)
			if (diff < 0.0) {
				diff = (0.0 - diff)
			}
			iterations = (iterations + 1.0)
		}
		stack_calc.PushRaw(floatToBytes(x))
		stack_calc.PushRaw(floatToBytes(iterations))
		return
	}()
	func() {
		stack_calc.Lock()
		defer stack_calc.Unlock()
		_bytes_iters, _err_iters := stack_calc.PopRaw()
		if _err_iters != nil { panic(_err_iters) }
		var iters float64 = bytesToFloat(_bytes_iters)
		_bytes_result, _err_result := stack_calc.PopRaw()
		if _err_result != nil { panic(_err_result) }
		var result float64 = bytesToFloat(_bytes_result)
		var expected float64 = 1.41421356
		var diff float64 = (result - expected)
		if (diff < 0.0) {
			diff = (0.0 - diff)
		}
		if (diff < 0.0001) {
			if (iters < 20.0) {
				stack_calc.PushRaw(floatToBytes(1.0))
				return
			}
		}
		stack_calc.PushRaw(floatToBytes(0.0))
		return
	}()
	stack_out := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_out.Push(intToBytes(1))
	{ v, _ := stack_out.Pop(); stack_dstack.Push(v) }
	{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
	
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
