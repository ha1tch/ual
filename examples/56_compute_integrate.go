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
	stack_integrate := ual.NewStack(ual.LIFO, ual.TypeFloat64)
	stack_integrate.Push(floatToBytes(0.000000))
	stack_integrate.Push(floatToBytes(1.000000))
	stack_integrate.Push(floatToBytes(1000.000000))
	func() {
		stack_integrate.Lock()
		defer stack_integrate.Unlock()
		_bytes_n, _err_n := stack_integrate.PopRaw()
		if _err_n != nil { panic(_err_n) }
		var n float64 = bytesToFloat(_bytes_n)
		_bytes_b, _err_b := stack_integrate.PopRaw()
		if _err_b != nil { panic(_err_b) }
		var b float64 = bytesToFloat(_bytes_b)
		_bytes_a, _err_a := stack_integrate.PopRaw()
		if _err_a != nil { panic(_err_a) }
		var a float64 = bytesToFloat(_bytes_a)
		var h float64 = ((b - a) / n)
		var sum float64 = 0.0
		var x float64 = a
		var i float64 = 0.0
		var fa float64 = (a * a)
		sum = (fa / 2.0)
		i = 1.0
		for (i < n) {
			x = (a + (i * h))
			var fx float64 = (x * x)
			sum = (sum + fx)
			i = (i + 1.0)
		}
		var fb float64 = (b * b)
		sum = (sum + (fb / 2.0))
		var result float64 = (h * sum)
		stack_integrate.PushRaw(floatToBytes(result))
		return
	}()
	stack_integrate.Push(floatToBytes(0.333333))
	func() {
		stack_integrate.Lock()
		defer stack_integrate.Unlock()
		_bytes_expected, _err_expected := stack_integrate.PopRaw()
		if _err_expected != nil { panic(_err_expected) }
		var expected float64 = bytesToFloat(_bytes_expected)
		_bytes_actual, _err_actual := stack_integrate.PopRaw()
		if _err_actual != nil { panic(_err_actual) }
		var actual float64 = bytesToFloat(_bytes_actual)
		var diff float64 = (actual - expected)
		if (diff < 0.0) {
			diff = (0.0 - diff)
		}
		if (diff < 0.001) {
			stack_integrate.PushRaw(floatToBytes(1.0))
			return
		}
		stack_integrate.PushRaw(floatToBytes(0.0))
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
