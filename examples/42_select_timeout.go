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
	stack_inbox := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_commands := ual.NewStack(ual.LIFO, ual.TypeInt64)
	stack_i64.PushAt(0, intToBytes(int64(0))) // var attempts
	spawn_mu.Lock()
	spawn_tasks = append(spawn_tasks, func() {
		stack_i64.PushAt(1, intToBytes(int64(0))) // var i
		for func() int64 { v, _ := stack_i64.PeekAt(1); return bytesToInt(v) }() < int64(500000) {
			{ v, _ := stack_i64.PeekAt(1); stack_dstack.Push(v) } // push i
			{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
			{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(1, v) } // i = ...
		}
		stack_inbox.Push(intToBytes(42))
	})
	spawn_mu.Unlock()
	spawn_mu.Lock()
	if len(spawn_tasks) > 0 {
		_task := spawn_tasks[len(spawn_tasks)-1]
		spawn_tasks = spawn_tasks[:len(spawn_tasks)-1]
		spawn_mu.Unlock()
		go _task()
	} else {
		spawn_mu.Unlock()
	}
	// select block
	func() {
		// setup
		stack_dstack.Push(intToBytes(1))
		{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		
		type _selectResult struct {
			caseID int
			value  []byte
		}
		
		_ctx1, _cancel1 := _selectContext()
		defer _cancel1()
		
		_resultCh1 := make(chan _selectResult, 1)
		
		// Case 0: @inbox
		go func() {
			_retry1_0:
			_v, _err := stack_inbox.TakeWithContext(_ctx1, int64(50))
			if _err != nil {
				// Check if it was a timeout (not a cancel)
				if _err.Error() == "timeout" {
					{ v, _ := stack_i64.PeekAt(0); stack_dstack.Push(v) } // push attempts
					{ v, _ := stack_dstack.Pop(); stack_dstack.Push(intToBytes(bytesToInt(v) + 1)) }
					{ v, _ := stack_dstack.Pop(); stack_i64.PushAt(0, v) } // attempts = ...
					{ v, _ := stack_i64.PeekAt(0); stack_dstack.Push(v) } // push attempts
					{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
					goto _retry1_0
					return // timeout, select completes via handler or default
				}
				return // cancelled
			}
			select {
			case _resultCh1 <- _selectResult{0, _v}:
				_cancel1() // won the race
			default:
			}
		}()
		
		// Case 1: @commands
		go func() {
			_v, _err := stack_commands.TakeWithContext(_ctx1, 0)
			if _err != nil {
				return // cancelled
			}
			select {
			case _resultCh1 <- _selectResult{1, _v}:
				_cancel1() // won the race
			default:
			}
		}()
		
		// Blocking: wait for a result
		_result := <-_resultCh1
		switch _result.caseID {
		case 0: // @inbox
			msg := bytesToInt(_result.value)
			_ = msg // suppress unused warning
			stack_dstack.Push(intToBytes(msg))
			{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		case 1: // @commands
			cmd := bytesToInt(_result.value)
			_ = cmd // suppress unused warning
			stack_dstack.Push(intToBytes(cmd))
			{ v, _ := stack_dstack.Pop(); fmt.Println(bytesToInt(v)) }
		}
	}()
	stack_dstack.Push(intToBytes(999))
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
