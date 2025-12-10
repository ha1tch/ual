package ual

import (
	"errors"
	"sync"
)

// Int64Stack is a type-specialised stack for int64 values.
// Zero allocation on hot paths when preallocated.
type Int64Stack struct {
	mu          sync.RWMutex
	perspective Perspective
	frozen      bool
	capacity    int
	
	elements []int64
	head     int
}

// NewInt64Stack creates a stack with given perspective
func NewInt64Stack(p Perspective) *Int64Stack {
	return &Int64Stack{
		perspective: p,
		elements:    make([]int64, 0),
	}
}

// NewCappedInt64Stack creates a preallocated stack (zero allocation after creation)
func NewCappedInt64Stack(p Perspective, capacity int) *Int64Stack {
	return &Int64Stack{
		perspective: p,
		capacity:    capacity,
		elements:    make([]int64, 0, capacity),
	}
}

// Push adds a value. O(1) amortised.
func (s *Int64Stack) Push(value int64) error {
	s.mu.Lock()
	
	if s.frozen {
		s.mu.Unlock()
		return errors.New("stack is frozen")
	}
	
	if s.capacity > 0 && len(s.elements)-s.head >= s.capacity {
		s.mu.Unlock()
		return errors.New("stack is full")
	}
	
	s.elements = append(s.elements, value)
	s.mu.Unlock()
	return nil
}

// Pop removes and returns top (LIFO) or bottom (FIFO). O(1).
func (s *Int64Stack) Pop() (int64, error) {
	s.mu.Lock()
	
	if s.frozen {
		s.mu.Unlock()
		return 0, errors.New("stack is frozen")
	}
	
	size := len(s.elements) - s.head
	if size == 0 {
		s.mu.Unlock()
		return 0, errors.New("stack empty")
	}
	
	var value int64
	
	if s.perspective == FIFO {
		value = s.elements[s.head]
		s.head++
		// Compact if too much slack
		if s.head > len(s.elements)/2 && s.head > 100 {
			s.compactLocked()
		}
	} else {
		// LIFO (default)
		idx := len(s.elements) - 1
		value = s.elements[idx]
		s.elements = s.elements[:idx]
	}
	
	s.mu.Unlock()
	return value, nil
}

// Peek returns top (LIFO) or bottom (FIFO) without removing. O(1).
func (s *Int64Stack) Peek() (int64, error) {
	s.mu.RLock()
	
	size := len(s.elements) - s.head
	if size == 0 {
		s.mu.RUnlock()
		return 0, errors.New("stack empty")
	}
	
	var value int64
	if s.perspective == FIFO {
		value = s.elements[s.head]
	} else {
		value = s.elements[len(s.elements)-1]
	}
	
	s.mu.RUnlock()
	return value, nil
}

// PeekAt returns element at offset from access point. O(1).
func (s *Int64Stack) PeekAt(offset int) (int64, error) {
	s.mu.RLock()
	
	var idx int
	if s.perspective == FIFO {
		idx = s.head + offset
	} else {
		idx = len(s.elements) - 1 - offset
	}
	
	if idx < s.head || idx >= len(s.elements) {
		s.mu.RUnlock()
		return 0, errors.New("index out of bounds")
	}
	
	value := s.elements[idx]
	s.mu.RUnlock()
	return value, nil
}

// Len returns number of elements. O(1).
func (s *Int64Stack) Len() int {
	s.mu.RLock()
	n := len(s.elements) - s.head
	s.mu.RUnlock()
	return n
}

// SetPerspective changes access pattern.
func (s *Int64Stack) SetPerspective(p Perspective) {
	s.mu.Lock()
	s.perspective = p
	s.mu.Unlock()
}

// Perspective returns current access pattern.
func (s *Int64Stack) Perspective() Perspective {
	s.mu.RLock()
	p := s.perspective
	s.mu.RUnlock()
	return p
}

// Freeze makes the stack immutable.
func (s *Int64Stack) Freeze() {
	s.mu.Lock()
	s.compactLocked()
	s.frozen = true
	s.mu.Unlock()
}

// IsFrozen returns whether the stack is immutable.
func (s *Int64Stack) IsFrozen() bool {
	s.mu.RLock()
	f := s.frozen
	s.mu.RUnlock()
	return f
}

// compactLocked removes slack (caller must hold lock)
func (s *Int64Stack) compactLocked() {
	if s.head == 0 {
		return
	}
	copy(s.elements, s.elements[s.head:])
	s.elements = s.elements[:len(s.elements)-s.head]
	s.head = 0
}

// Int64View provides a perspective over an Int64Stack
type Int64View struct {
	mu          sync.RWMutex
	stack       *Int64Stack
	perspective Perspective
	cursor      int
}

// NewInt64View creates a view with given perspective
func NewInt64View(p Perspective) *Int64View {
	return &Int64View{
		perspective: p,
	}
}

// Attach connects view to a stack
func (v *Int64View) Attach(s *Int64Stack) {
	v.mu.Lock()
	v.stack = s
	v.cursor = 0
	v.mu.Unlock()
}

// Detach disconnects view from stack
func (v *Int64View) Detach() {
	v.mu.Lock()
	v.stack = nil
	v.cursor = 0
	v.mu.Unlock()
}

// Peek returns element at cursor without removing. O(1).
func (v *Int64View) Peek() (int64, error) {
	v.mu.RLock()
	if v.stack == nil {
		v.mu.RUnlock()
		return 0, errors.New("view not attached")
	}
	
	v.stack.mu.RLock()
	
	size := len(v.stack.elements) - v.stack.head
	if size == 0 {
		v.stack.mu.RUnlock()
		v.mu.RUnlock()
		return 0, errors.New("stack empty")
	}
	
	var idx int
	if v.perspective == FIFO {
		idx = v.stack.head + v.cursor
	} else {
		idx = len(v.stack.elements) - 1 - v.cursor
	}
	
	if idx < v.stack.head || idx >= len(v.stack.elements) {
		v.stack.mu.RUnlock()
		v.mu.RUnlock()
		return 0, errors.New("cursor out of bounds")
	}
	
	value := v.stack.elements[idx]
	v.stack.mu.RUnlock()
	v.mu.RUnlock()
	return value, nil
}

// Pop removes and returns element from view's perspective. O(1).
func (v *Int64View) Pop() (int64, error) {
	v.mu.Lock()
	if v.stack == nil {
		v.mu.Unlock()
		return 0, errors.New("view not attached")
	}
	
	v.stack.mu.Lock()
	
	if v.stack.frozen {
		v.stack.mu.Unlock()
		v.mu.Unlock()
		return 0, errors.New("stack is frozen")
	}
	
	size := len(v.stack.elements) - v.stack.head
	if size == 0 {
		v.stack.mu.Unlock()
		v.mu.Unlock()
		return 0, errors.New("stack empty")
	}
	
	var value int64
	
	if v.perspective == FIFO {
		// Steal from bottom
		value = v.stack.elements[v.stack.head]
		v.stack.head++
		if v.stack.head > len(v.stack.elements)/2 && v.stack.head > 100 {
			v.stack.compactLocked()
		}
	} else {
		// Pop from top
		idx := len(v.stack.elements) - 1
		value = v.stack.elements[idx]
		v.stack.elements = v.stack.elements[:idx]
	}
	
	v.stack.mu.Unlock()
	v.mu.Unlock()
	return value, nil
}

// Advance moves cursor forward. O(1).
func (v *Int64View) Advance() error {
	v.mu.Lock()
	if v.stack == nil {
		v.mu.Unlock()
		return errors.New("view not attached")
	}
	v.cursor++
	v.mu.Unlock()
	return nil
}

// Reset moves cursor to start.
func (v *Int64View) Reset() {
	v.mu.Lock()
	v.cursor = 0
	v.mu.Unlock()
}
