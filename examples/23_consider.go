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

func divide(a int64, b int64) int64 {
	stack_i64.PushAt(0, intToBytes(int64(a))) // param a
	stack_i64.PushAt(1, intToBytes(int64(b))) // param b
	if func() int64 { v, _ := stack_i64.PeekAt(1); return bytesToInt(v) }() == int64(0) {
		_consider_status = "error"
		_consider_value = "division by zero"
		return 0
	}
	return (func() int64 { v, _ := stack_i64.PeekAt(0); return bytesToInt(v) }() / func() int64 { v, _ := stack_i64.PeekAt(1); return bytesToInt(v) }())
}

func validate_age(age int64) int64 {
	stack_i64.PushAt(2, intToBytes(int64(age))) // param age
	if func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }() < int64(0) {
		_consider_status = "invalid"
		_consider_value = "age cannot be negative"
		return 0
	}
	if func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }() < int64(18) {
		_consider_status = "minor"
		_consider_value = func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }()
		return func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }()
	}
	if func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }() > int64(120) {
		_consider_status = "invalid"
		_consider_value = "age too high"
		return 0
	}
	_consider_status = "ok"
	return func() int64 { v, _ := stack_i64.PeekAt(2); return bytesToInt(v) }()
}

func main() {
	func() {
		_saved_status_1 := _consider_status
		_saved_value_1 := _consider_value
		_consider_status = "ok"
		_consider_value = nil
		
		stack_dstack.Push(intToBytes(42))
		stack_dstack.Push(intToBytes(10))
		{ b, _ := stack_dstack.Pop(); a, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(a) + bytesToInt(b))) }
		{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		
		// Check for errors (implicit from @error stack)
		if _consider_status == "ok" && stack_error.Len() > 0 {
			_consider_status = "error"
			if _v, _err := stack_error.Peek(); _err == nil { _consider_value = string(_v) }
		}
		
		switch _consider_status {
		case "ok":
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		case "error":
			var e int64
			switch _v := _consider_value.(type) {
			case int64:
				e = _v
			case int:
				e = int64(_v)
			case string:
				fmt.Sscanf(_v, "%d", &e)
			}
			_ = e // suppress unused
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		default:
			panic("unhandled status in consider: " + _consider_status)
		}
		
		_consider_status = _saved_status_1
		_consider_value = _saved_value_1
	}()
	func() {
		_saved_status_2 := _consider_status
		_saved_value_2 := _consider_value
		_consider_status = "ok"
		_consider_value = nil
		
		stack_i64.PushAt(3, intToBytes(int64(divide(100, 0)))) // var result
		{ v, _ := stack_i64.PeekAt(3); stack_dstack.Push(v) } // push result
		
		// Check for errors (implicit from @error stack)
		if _consider_status == "ok" && stack_error.Len() > 0 {
			_consider_status = "error"
			if _v, _err := stack_error.Peek(); _err == nil { _consider_value = string(_v) }
		}
		
		switch _consider_status {
		case "ok":
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
			{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		case "error":
			var msg int64
			switch _v := _consider_value.(type) {
			case int64:
				msg = _v
			case int:
				msg = int64(_v)
			case string:
				fmt.Sscanf(_v, "%d", &msg)
			}
			_ = msg // suppress unused
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		default:
			panic("unhandled status in consider: " + _consider_status)
		}
		
		_consider_status = _saved_status_2
		_consider_value = _saved_value_2
	}()
	func() {
		_saved_status_3 := _consider_status
		_saved_value_3 := _consider_value
		_consider_status = "ok"
		_consider_value = nil
		
		stack_i64.PushAt(4, intToBytes(int64(validate_age(15)))) // var age
		{ v, _ := stack_i64.PeekAt(4); stack_dstack.Push(v) } // push age
		
		// Check for errors (implicit from @error stack)
		if _consider_status == "ok" && stack_error.Len() > 0 {
			_consider_status = "error"
			if _v, _err := stack_error.Peek(); _err == nil { _consider_value = string(_v) }
		}
		
		switch _consider_status {
		case "ok":
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		case "minor":
			var age int64
			switch _v := _consider_value.(type) {
			case int64:
				age = _v
			case int:
				age = int64(_v)
			case string:
				fmt.Sscanf(_v, "%d", &age)
			}
			_ = age // suppress unused
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		case "invalid":
			var reason int64
			switch _v := _consider_value.(type) {
			case int64:
				reason = _v
			case int:
				reason = int64(_v)
			case string:
				fmt.Sscanf(_v, "%d", &reason)
			}
			_ = reason // suppress unused
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		default:
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		}
		
		_consider_status = _saved_status_3
		_consider_value = _saved_value_3
	}()
	func() {
		_saved_status_4 := _consider_status
		_saved_value_4 := _consider_value
		_consider_status = "ok"
		_consider_value = nil
		
		stack_dstack.Push(intToBytes(100))
		func() {
			_saved_status_5 := _consider_status
			_saved_value_5 := _consider_value
			_consider_status = "ok"
			_consider_value = nil
			
			stack_dstack.Push(intToBytes(50))
			_consider_status = "partial"
			
			// Check for errors (implicit from @error stack)
			if _consider_status == "ok" && stack_error.Len() > 0 {
				_consider_status = "error"
				if _v, _err := stack_error.Peek(); _err == nil { _consider_value = string(_v) }
			}
			
			switch _consider_status {
			case "ok":
				{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
			case "partial":
				{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
			default:
				{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
			}
			
			_consider_status = _saved_status_5
			_consider_value = _saved_value_5
		}()
		{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		
		// Check for errors (implicit from @error stack)
		if _consider_status == "ok" && stack_error.Len() > 0 {
			_consider_status = "error"
			if _v, _err := stack_error.Peek(); _err == nil { _consider_value = string(_v) }
		}
		
		switch _consider_status {
		case "ok":
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		default:
			{ v, _ := stack_dstack.Peek(); fmt.Println(bytesToInt(v)) }
		}
		
		_consider_status = _saved_status_4
		_consider_value = _saved_value_4
	}()
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
