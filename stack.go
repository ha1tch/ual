package ual

import (
	"context"
	"encoding/binary"
	"errors"
	"hash/fnv"
	"sync"
	"time"
)

// Perspective determines how access parameters are interpreted
type Perspective int

const (
	LIFO Perspective = iota
	FIFO
	Indexed
	Hash
)

// ElementType represents the type property of a container
type ElementType int

const (
	TypeInt64 ElementType = iota
	TypeUint64
	TypeFloat64
	TypeString
	TypeBytes
	TypeBool
)

// Element wraps raw bytes with type awareness
type Element struct {
	data []byte
}

// Stack is a container that confers type to its elements
type Stack struct {
	mu          sync.RWMutex
	cond        *sync.Cond   // for blocking take
	perspective Perspective
	elementType ElementType
	frozen      bool
	capacity    int // 0 = unlimited
	closed      bool // when true, take returns immediately
	
	// Unified storage: all perspectives use this
	// For positional: sequential access
	// For hash: key -> index mapping
	elements []Element
	keys     [][]byte           // parallel to elements for hash perspective
	hashIdx  map[string]int     // key string -> index (hash perspective only)
	
	// Position tracking for FIFO (head points to first valid element)
	head int
}

// NewStack creates a stack with given perspective and element type
func NewStack(p Perspective, t ElementType) *Stack {
	s := &Stack{
		perspective: p,
		elementType: t,
		elements:    make([]Element, 0),
		keys:        make([][]byte, 0),
	}
	s.cond = sync.NewCond(&s.mu)
	if p == Hash {
		s.hashIdx = make(map[string]int)
	}
	return s
}

// NewCappedStack creates a stack with fixed capacity (no allocations after creation)
func NewCappedStack(p Perspective, t ElementType, capacity int) *Stack {
	s := &Stack{
		perspective: p,
		elementType: t,
		capacity:    capacity,
		elements:    make([]Element, 0, capacity),
		keys:        make([][]byte, 0, capacity),
	}
	s.cond = sync.NewCond(&s.mu)
	if p == Hash {
		s.hashIdx = make(map[string]int, capacity)
	}
	return s
}

// Capacity returns the stack's capacity (0 = unlimited)
func (s *Stack) Capacity() int {
	return s.capacity
}

// IsFull returns true if stack is at capacity
func (s *Stack) IsFull() bool {
	if s.capacity == 0 {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.elements)-s.head >= s.capacity
}

// Push adds an element. For hash perspective, requires a key.
func (s *Stack) Push(value []byte, key ...[]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.frozen {
		return errors.New("stack is frozen")
	}
	
	if s.capacity > 0 && len(s.elements)-s.head >= s.capacity {
		return errors.New("stack is full")
	}
	
	elem := Element{data: value}
	
	switch s.perspective {
	case LIFO, FIFO, Indexed:
		s.elements = append(s.elements, elem)
		s.keys = append(s.keys, nil) // no key for positional
		
	case Hash:
		if len(key) == 0 {
			return errors.New("hash perspective requires key")
		}
		k := key[0]
		keyStr := string(k)
		
		// Check if key exists - update in place
		if idx, exists := s.hashIdx[keyStr]; exists {
			s.elements[idx] = elem
			s.cond.Signal() // wake one waiter
			return nil
		}
		
		// New key
		idx := len(s.elements)
		s.elements = append(s.elements, elem)
		s.keys = append(s.keys, k)
		s.hashIdx[keyStr] = idx
	}
	
	s.cond.Signal() // wake one waiter
	return nil
}

// Pop removes and returns an element. Parameter interpretation depends on perspective.
func (s *Stack) Pop(param ...[]byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.frozen {
		return nil, errors.New("stack is frozen")
	}
	
	size := len(s.elements) - s.head
	if size == 0 {
		return nil, errors.New("stack empty")
	}
	
	var elem Element
	
	switch s.perspective {
	case LIFO:
		// Pop from end - O(1)
		idx := len(s.elements) - 1
		if len(param) > 0 {
			offset := bytesToInt(param[0])
			idx = len(s.elements) - 1 - int(offset)
			if idx < s.head || idx >= len(s.elements) {
				return nil, errors.New("index out of bounds")
			}
			// Non-default pop requires shift
			elem = s.elements[idx]
			s.elements = append(s.elements[:idx], s.elements[idx+1:]...)
			s.keys = append(s.keys[:idx], s.keys[idx+1:]...)
		} else {
			elem = s.elements[idx]
			s.elements = s.elements[:idx]
			s.keys = s.keys[:idx]
		}
		
	case FIFO:
		// Pop from head - O(1)
		idx := s.head
		if len(param) > 0 {
			offset := bytesToInt(param[0])
			idx = s.head + int(offset)
			if idx < s.head || idx >= len(s.elements) {
				return nil, errors.New("index out of bounds")
			}
			// Non-default pop requires shift
			elem = s.elements[idx]
			s.elements = append(s.elements[:idx], s.elements[idx+1:]...)
			s.keys = append(s.keys[:idx], s.keys[idx+1:]...)
		} else {
			elem = s.elements[idx]
			s.head++
			// Compact if too much slack
			if s.head > len(s.elements)/2 && s.head > 100 {
				s.compact()
			}
		}
		
	case Indexed:
		// Default: pop from end (like removing last array element)
		// With param: pop from specific index
		var idx int
		if len(param) == 0 {
			idx = len(s.elements) - 1
		} else {
			idx = s.head + int(bytesToInt(param[0]))
		}
		if idx < s.head || idx >= len(s.elements) {
			return nil, errors.New("index out of bounds")
		}
		elem = s.elements[idx]
		s.elements = append(s.elements[:idx], s.elements[idx+1:]...)
		s.keys = append(s.keys[:idx], s.keys[idx+1:]...)
		
	case Hash:
		// No default, key required
		if len(param) == 0 {
			return nil, errors.New("hash perspective requires key")
		}
		keyStr := string(param[0])
		idx, exists := s.hashIdx[keyStr]
		if !exists {
			return nil, errors.New("key not found")
		}
		elem = s.elements[idx]
		// Remove from hash index
		delete(s.hashIdx, keyStr)
		// Mark slot as empty (tombstone) instead of shifting
		s.elements[idx] = Element{}
		s.keys[idx] = nil
		// Could compact periodically, but for now just leave tombstones
	}
	
	return elem.data, nil
}

// Peek returns element without removing it
func (s *Stack) Peek(param ...[]byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	size := len(s.elements) - s.head
	if size == 0 {
		return nil, errors.New("stack empty")
	}
	
	var idx int
	
	switch s.perspective {
	case LIFO:
		idx = len(s.elements) - 1
		if len(param) > 0 {
			offset := bytesToInt(param[0])
			idx = len(s.elements) - 1 - int(offset)
		}
		
	case FIFO:
		idx = s.head
		if len(param) > 0 {
			offset := bytesToInt(param[0])
			idx = s.head + int(offset)
		}
		
	case Indexed:
		if len(param) == 0 {
			return nil, errors.New("indexed perspective requires position")
		}
		idx = s.head + int(bytesToInt(param[0]))
		
	case Hash:
		if len(param) == 0 {
			return nil, errors.New("hash perspective requires key")
		}
		keyStr := string(param[0])
		var exists bool
		idx, exists = s.hashIdx[keyStr]
		if !exists {
			return nil, errors.New("key not found")
		}
	}
	
	if idx < s.head || idx >= len(s.elements) {
		return nil, errors.New("index out of bounds")
	}
	
	return s.elements[idx].data, nil
}

// =============================================================================
// Locking Methods (for compute blocks)
// =============================================================================

// Lock acquires exclusive write lock on the stack.
// Used by generated compute block code for atomic operations.
func (s *Stack) Lock() {
	s.mu.Lock()
}

// Unlock releases the exclusive write lock.
// Used by generated compute block code.
func (s *Stack) Unlock() {
	s.mu.Unlock()
}

// =============================================================================
// Raw Methods (for compute blocks - caller MUST hold s.mu)
// =============================================================================

// PopRaw removes and returns the top element without acquiring the mutex.
// UNSAFE: Caller must hold s.mu.Lock() before calling.
// Used by generated compute block code.
func (s *Stack) PopRaw() ([]byte, error) {
	size := len(s.elements) - s.head
	if size == 0 {
		return nil, errors.New("stack underflow in compute")
	}

	var idx int
	switch s.perspective {
	case LIFO:
		idx = len(s.elements) - 1
	case FIFO:
		idx = s.head
	default:
		// Indexed/Hash: default to LIFO behavior for raw pop
		idx = len(s.elements) - 1
	}

	elem := s.elements[idx]

	if s.perspective == FIFO {
		s.elements[s.head] = Element{}
		s.head++
		if s.head > len(s.elements)/2 && s.head > 16 {
			s.compact()
		}
	} else {
		s.elements = s.elements[:idx]
	}

	return elem.data, nil
}

// PushRaw appends an element without acquiring the mutex.
// UNSAFE: Caller must hold s.mu.Lock() before calling.
// Used by generated compute block code.
func (s *Stack) PushRaw(value []byte) error {
	if s.capacity > 0 && len(s.elements)-s.head >= s.capacity {
		return errors.New("stack full in compute")
	}
	s.elements = append(s.elements, Element{data: value})
	s.keys = append(s.keys, nil) // maintain key slice alignment
	return nil
}

// SetRaw sets a hash key to a value without acquiring the mutex.
// UNSAFE: Caller must hold s.mu.Lock() before calling.
// Used by generated compute block code for returning values in Hash stacks.
// Only valid for Hash perspective.
func (s *Stack) SetRaw(key string, value []byte) error {
	if s.perspective != Hash {
		return errors.New("SetRaw only valid for Hash perspective")
	}
	elem := Element{data: value}
	
	// Check if key exists - update in place
	if idx, exists := s.hashIdx[key]; exists {
		s.elements[idx] = elem
		return nil
	}
	
	// New key
	idx := len(s.elements)
	s.elements = append(s.elements, elem)
	s.keys = append(s.keys, []byte(key))
	s.hashIdx[key] = idx
	return nil
}

// GetRaw retrieves a value by key without acquiring the mutex.
// UNSAFE: Caller must hold s.mu.Lock() (or s.mu.RLock()) before calling.
// Used by generated compute block code for self.key access.
// Only valid for Hash perspective.
func (s *Stack) GetRaw(key string) ([]byte, bool) {
	if s.perspective != Hash {
		return nil, false
	}
	idx, exists := s.hashIdx[key]
	if !exists {
		return nil, false
	}
	if idx < s.head || idx >= len(s.elements) {
		return nil, false
	}
	return s.elements[idx].data, true
}

// GetAtRaw retrieves a value by index without acquiring the mutex.
// UNSAFE: Caller must hold s.mu.Lock() (or s.mu.RLock()) before calling.
// Used by generated compute block code for self[i] access.
// Only valid for Indexed perspective.
func (s *Stack) GetAtRaw(index int) ([]byte, bool) {
	if s.perspective != Indexed {
		return nil, false
	}
	idx := s.head + index
	if idx < s.head || idx >= len(s.elements) {
		return nil, false
	}
	return s.elements[idx].data, true
}

// Len returns number of elements
func (s *Stack) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.elements) - s.head
}

// PushAt stores a value at a specific index (for hash/indexed stacks)
// Used for variable storage in type stacks
func (s *Stack) PushAt(index int, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.frozen {
		return errors.New("stack is frozen")
	}
	
	// Extend if needed
	for len(s.elements) <= index {
		s.elements = append(s.elements, Element{})
		s.keys = append(s.keys, nil)
	}
	
	s.elements[index] = Element{data: value}
	return nil
}

// PeekAt retrieves a value at a specific index without removing it
// Used for variable access in type stacks
func (s *Stack) PeekAt(index int) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if index < 0 || index >= len(s.elements) {
		return nil, errors.New("index out of bounds")
	}
	
	return s.elements[index].data, nil
}

// SetPerspective changes how the stack is accessed
func (s *Stack) SetPerspective(p Perspective) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	oldPerspective := s.perspective
	s.perspective = p
	
	// If switching to hash from non-hash, we need keys
	// This is a problem - elements don't have keys yet
	// For now: use position as key
	if p == Hash && oldPerspective != Hash {
		s.hashIdx = make(map[string]int)
		for i := range s.elements {
			if s.keys[i] == nil {
				// Generate key from position
				s.keys[i] = intToBytes(int64(i))
			}
			s.hashIdx[string(s.keys[i])] = i
		}
	}
}

// reindex rebuilds hash index after removal
func (s *Stack) reindex() {
	s.hashIdx = make(map[string]int)
	for i, k := range s.keys {
		if k != nil {
			s.hashIdx[string(k)] = i
		}
	}
}

// compact removes slack from FIFO head
func (s *Stack) compact() {
	if s.head == 0 {
		return
	}
	s.elements = s.elements[s.head:]
	s.keys = s.keys[s.head:]
	s.head = 0
}

// Freeze makes the stack immutable. Peek, Walk still work. Push, Pop will error.
func (s *Stack) Freeze() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.compact()
	s.frozen = true
}

// IsFrozen returns whether the stack is immutable
func (s *Stack) IsFrozen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.frozen
}

// Helper: bytes to int64
func bytesToInt(b []byte) int64 {
	if len(b) >= 8 {
		return int64(binary.BigEndian.Uint64(b))
	}
	// Pad if needed
	padded := make([]byte, 8)
	copy(padded[8-len(b):], b)
	return int64(binary.BigEndian.Uint64(padded))
}

// Helper: int64 to bytes
func intToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

// Helper: hash bytes to uint64 (for internal use)
func hashBytes(b []byte) uint64 {
	h := fnv.New64a()
	h.Write(b)
	return h.Sum64()
}

// Take removes and returns an element, blocking until one is available.
// Optional timeout in milliseconds (0 = wait forever).
// Returns nil, error if stack is closed or timeout.
func (s *Stack) Take(timeoutMs ...int64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	timeout := int64(0)
	if len(timeoutMs) > 0 {
		timeout = timeoutMs[0]
	}
	
	// Wait for data
	for len(s.elements)-s.head == 0 && !s.closed {
		if timeout > 0 {
			// Timed wait using a goroutine
			done := make(chan bool, 1)
			go func() {
				s.mu.Lock()
				s.cond.Wait()
				s.mu.Unlock()
				done <- true
			}()
			s.mu.Unlock()
			
			select {
			case <-done:
				s.mu.Lock()
			case <-timeAfter(timeout):
				s.mu.Lock()
				return nil, errors.New("take timeout")
			}
		} else {
			s.cond.Wait()
		}
	}
	
	if s.closed && len(s.elements)-s.head == 0 {
		return nil, errors.New("stack closed")
	}
	
	// Now pop based on perspective
	var elem Element
	
	switch s.perspective {
	case LIFO:
		idx := len(s.elements) - 1
		elem = s.elements[idx]
		s.elements = s.elements[:idx]
		s.keys = s.keys[:idx]
		
	case FIFO:
		idx := s.head
		elem = s.elements[idx]
		s.head++
		if s.head > len(s.elements)/2 && s.head > 100 {
			s.compact()
		}
		
	default:
		// For indexed/hash, behave like LIFO
		idx := len(s.elements) - 1
		elem = s.elements[idx]
		s.elements = s.elements[:idx]
		s.keys = s.keys[:idx]
	}
	
	return elem.data, nil
}

// TakeWithContext removes and returns an element, blocking until one is available
// or the context is cancelled. Optional timeout in milliseconds (0 = no timeout, just context).
// Returns nil, error if context is cancelled, stack is closed, or timeout.
func (s *Stack) TakeWithContext(ctx context.Context, timeoutMs int64) ([]byte, error) {
	// Create a channel to receive the result
	type result struct {
		data []byte
		err  error
	}
	resultCh := make(chan result, 1)
	
	go func() {
		if timeoutMs > 0 {
			data, err := s.Take(timeoutMs)
			resultCh <- result{data, err}
		} else {
			// Use a very long timeout to allow context cancellation to work
			// Check periodically
			for {
				// Try take with short timeout
				data, err := s.Take(100) // 100ms check interval
				if err == nil {
					resultCh <- result{data, nil}
					return
				}
				if err.Error() != "take timeout" {
					// Real error (closed, etc)
					resultCh <- result{nil, err}
					return
				}
				// It was a timeout, check if context is done
				select {
				case <-ctx.Done():
					resultCh <- result{nil, errors.New("cancelled")}
					return
				default:
					// Continue waiting
				}
			}
		}
	}()
	
	select {
	case <-ctx.Done():
		return nil, errors.New("cancelled")
	case r := <-resultCh:
		if r.err != nil && r.err.Error() == "take timeout" {
			return nil, errors.New("timeout")
		}
		return r.data, r.err
	}
}

// Close signals that no more data will be pushed.
// Wakes all waiting Take calls.
func (s *Stack) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.cond.Broadcast()
}

// IsClosed returns whether the stack has been closed.
func (s *Stack) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// timeAfter returns a channel that receives after ms milliseconds
func timeAfter(ms int64) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		close(ch)
	}()
	return ch
}
