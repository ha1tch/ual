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
	stack_physics := ual.NewStack(ual.LIFO, ual.TypeFloat64)
	stack_physics.Push(floatToBytes(10.000000))
	stack_physics.Push(floatToBytes(5.000000))
	stack_physics.Push(floatToBytes(0.500000))
	func() {
		stack_physics.Lock()
		defer stack_physics.Unlock()
		_bytes_factor, _err_factor := stack_physics.PopRaw()
		if _err_factor != nil { panic(_err_factor) }
		var factor float64 = bytesToFloat(_bytes_factor)
		_bytes_v, _err_v := stack_physics.PopRaw()
		if _err_v != nil { panic(_err_v) }
		var v float64 = bytesToFloat(_bytes_v)
		_bytes_m, _err_m := stack_physics.PopRaw()
		if _err_m != nil { panic(_err_m) }
		var m float64 = bytesToFloat(_bytes_m)
		var ke float64 = (((factor * m) * v) * v)
		stack_physics.PushRaw(floatToBytes(ke))
		return
	}()
	stack_physics.Push(floatToBytes(125.000000))
	func() {
		stack_physics.Lock()
		defer stack_physics.Unlock()
		_bytes_expected, _err_expected := stack_physics.PopRaw()
		if _err_expected != nil { panic(_err_expected) }
		var expected float64 = bytesToFloat(_bytes_expected)
		_bytes_actual, _err_actual := stack_physics.PopRaw()
		if _err_actual != nil { panic(_err_actual) }
		var actual float64 = bytesToFloat(_bytes_actual)
		if (actual == expected) {
			stack_physics.PushRaw(floatToBytes(1.0))
			return
		}
		stack_physics.PushRaw(floatToBytes(0.0))
		return
	}()
	stack_check := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_check.Push(intToBytes(1))
	{ v, _ := stack_check.Pop(); stack_dstack.Push(v) }
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
