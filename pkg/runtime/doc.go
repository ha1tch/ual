// Package runtime provides the stack-based runtime library for compiled ual programs.
//
// This package implements:
//   - Stack: thread-safe stack with multiple perspectives (LIFO, FIFO, Indexed, Hash)
//   - View: decoupled perspective on a stack
//   - Walk: iteration operations (Filter, Reduce, Map)
//   - Bring: element transfer between stacks
//   - WorkSteal: work-stealing scheduler
//
// Compiled ual programs import this package as:
//
//	import ual "github.com/ha1tch/ual/pkg/runtime"
//
// The runtime provides perspective-based stack access:
//
//	stack := ual.NewStack(ual.LIFO, ual.Int64)
//	stack.Push(intToBytes(42))
//	val, _ := stack.Pop()
//
// Stacks support four perspectives:
//   - LIFO: Last-In-First-Out (traditional stack)
//   - FIFO: First-In-First-Out (queue)
//   - Indexed: Random access by index
//   - Hash: Key-value access
package runtime
