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
	stack_mandel := ual.NewStack(ual.LIFO, ual.TypeFloat64)
	stack_mandel.Push(floatToBytes(0.250000))
	stack_mandel.Push(floatToBytes(0.500000))
	func() {
		stack_mandel.Lock()
		defer stack_mandel.Unlock()
		_bytes_ci, _err_ci := stack_mandel.PopRaw()
		if _err_ci != nil { panic(_err_ci) }
		var ci float64 = bytesToFloat(_bytes_ci)
		_bytes_cr, _err_cr := stack_mandel.PopRaw()
		if _err_cr != nil { panic(_err_cr) }
		var cr float64 = bytesToFloat(_bytes_cr)
		var zr float64 = 0.0
		var zi float64 = 0.0
		var zr2 float64 = 0.0
		var zi2 float64 = 0.0
		var iter float64 = 0.0
		var max_iter float64 = 1000.0
		var escape float64 = 4.0
		for (iter < max_iter) {
			zr2 = (zr * zr)
			zi2 = (zi * zi)
			if ((zr2 + zi2) > escape) {
				stack_mandel.PushRaw(floatToBytes(iter))
				return
			}
			zi = (((2.0 * zr) * zi) + ci)
			zr = ((zr2 - zi2) + cr)
			iter = (iter + 1.0)
		}
		stack_mandel.PushRaw(floatToBytes(max_iter))
		return
	}()
	func() {
		stack_mandel.Lock()
		defer stack_mandel.Unlock()
		_bytes_result, _err_result := stack_mandel.PopRaw()
		if _err_result != nil { panic(_err_result) }
		var result float64 = bytesToFloat(_bytes_result)
		if (result > 0.0) {
			if (result < 1001.0) {
				stack_mandel.PushRaw(floatToBytes(1.0))
				return
			}
		}
		stack_mandel.PushRaw(floatToBytes(0.0))
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
