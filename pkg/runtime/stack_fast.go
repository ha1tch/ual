package runtime

import (
	"sync/atomic"
)

// FastInt64Stack is a lock-free stack optimised for single-producer patterns.
// Uses atomic operations for the hot path, falling back to CAS for contention.
// LIFO only - use Int64Stack for FIFO.
type FastInt64Stack struct {
	elements []int64
	top      int64 // atomic: index of next push slot
	capacity int64
}

// NewFastInt64Stack creates a preallocated lock-free stack.
// Capacity is fixed - no growth after creation.
func NewFastInt64Stack(capacity int) *FastInt64Stack {
	return &FastInt64Stack{
		elements: make([]int64, capacity),
		top:      0,
		capacity: int64(capacity),
	}
}

// Push adds a value. Returns false if full. O(1), lock-free.
func (s *FastInt64Stack) Push(value int64) bool {
	for {
		top := atomic.LoadInt64(&s.top)
		if top >= s.capacity {
			return false
		}
		if atomic.CompareAndSwapInt64(&s.top, top, top+1) {
			s.elements[top] = value
			return true
		}
		// CAS failed, retry
	}
}

// Pop removes and returns top value. Returns 0, false if empty. O(1), lock-free.
func (s *FastInt64Stack) Pop() (int64, bool) {
	for {
		top := atomic.LoadInt64(&s.top)
		if top == 0 {
			return 0, false
		}
		newTop := top - 1
		if atomic.CompareAndSwapInt64(&s.top, top, newTop) {
			return s.elements[newTop], true
		}
		// CAS failed, retry
	}
}

// Peek returns top value without removing. O(1), lock-free.
func (s *FastInt64Stack) Peek() (int64, bool) {
	top := atomic.LoadInt64(&s.top)
	if top == 0 {
		return 0, false
	}
	return s.elements[top-1], true
}

// Len returns number of elements. O(1).
func (s *FastInt64Stack) Len() int {
	return int(atomic.LoadInt64(&s.top))
}

// Steal removes and returns bottom value (FIFO). For work-stealing pattern.
// O(1), lock-free.
func (s *FastInt64Stack) Steal() (int64, bool) {
	// This is trickier - we need to track both head and tail
	// For now, simple version: just pop from top
	// TODO: proper work-stealing deque with head/tail atomics
	return s.Pop()
}

// WorkStealingDeque is a proper Chase-Lev work-stealing deque.
// Owner pushes/pops from top (LIFO), thieves steal from bottom (FIFO).
type WorkStealingDeque struct {
	elements []int64
	top      int64 // atomic: owner's end (LIFO)
	bottom   int64 // atomic: thieves' end (FIFO)
	capacity int64
}

// NewWorkStealingDeque creates a fixed-capacity work-stealing deque.
func NewWorkStealingDeque(capacity int) *WorkStealingDeque {
	return &WorkStealingDeque{
		elements: make([]int64, capacity),
		top:      0,
		bottom:   0,
		capacity: int64(capacity),
	}
}

// Push adds a value at the top (owner only). O(1).
func (d *WorkStealingDeque) Push(value int64) bool {
	top := atomic.LoadInt64(&d.top)
	bottom := atomic.LoadInt64(&d.bottom)
	
	if top-bottom >= d.capacity {
		return false // full
	}
	
	d.elements[top%d.capacity] = value
	atomic.AddInt64(&d.top, 1)
	return true
}

// Pop removes from top (owner only). O(1).
func (d *WorkStealingDeque) Pop() (int64, bool) {
	top := atomic.AddInt64(&d.top, -1)
	bottom := atomic.LoadInt64(&d.bottom)
	
	if top < bottom {
		// Empty, restore top
		atomic.StoreInt64(&d.top, bottom)
		return 0, false
	}
	
	value := d.elements[top%d.capacity]
	
	if top == bottom {
		// Last element - race with thieves
		if !atomic.CompareAndSwapInt64(&d.bottom, bottom, bottom+1) {
			// Thief got it
			atomic.StoreInt64(&d.top, bottom+1)
			return 0, false
		}
		atomic.StoreInt64(&d.top, bottom+1)
	}
	
	return value, true
}

// Steal removes from bottom (thieves). O(1).
func (d *WorkStealingDeque) Steal() (int64, bool) {
	bottom := atomic.LoadInt64(&d.bottom)
	top := atomic.LoadInt64(&d.top)
	
	if bottom >= top {
		return 0, false // empty
	}
	
	value := d.elements[bottom%d.capacity]
	
	if !atomic.CompareAndSwapInt64(&d.bottom, bottom, bottom+1) {
		// Another thief got it
		return 0, false
	}
	
	return value, true
}

// Len returns approximate size.
func (d *WorkStealingDeque) Len() int {
	top := atomic.LoadInt64(&d.top)
	bottom := atomic.LoadInt64(&d.bottom)
	size := top - bottom
	if size < 0 {
		return 0
	}
	return int(size)
}
