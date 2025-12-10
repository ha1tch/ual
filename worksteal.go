package ual

import (
	"sync"
	"sync/atomic"
)

// ============================================================================
// Traditional Work-Stealing (Chase-Lev style deque)
// As taught in computer science - lock-free owner, locked steal
// ============================================================================

// Task represents a unit of work
type Task struct {
	ID   int64
	Data []byte
}

// WSDeque is a work-stealing deque (traditional implementation)
// Owner pushes/pops from bottom (LIFO), thieves steal from top (FIFO)
type WSDeque struct {
	tasks  []Task
	bottom int64 // atomic - owner's end
	top    int64 // atomic - thief's end
	mu     sync.Mutex // for steal synchronization
}

// NewWSDeque creates a traditional work-stealing deque
func NewWSDeque(capacity int) *WSDeque {
	return &WSDeque{
		tasks:  make([]Task, capacity),
		bottom: 0,
		top:    0,
	}
}

// Push adds a task (owner only, LIFO end)
func (d *WSDeque) Push(t Task) bool {
	b := atomic.LoadInt64(&d.bottom)
	top := atomic.LoadInt64(&d.top)
	
	if b-top >= int64(len(d.tasks)) {
		return false // full
	}
	
	d.tasks[b%int64(len(d.tasks))] = t
	atomic.StoreInt64(&d.bottom, b+1)
	return true
}

// Pop removes a task (owner only, LIFO end)
func (d *WSDeque) Pop() (Task, bool) {
	b := atomic.LoadInt64(&d.bottom) - 1
	atomic.StoreInt64(&d.bottom, b)
	
	top := atomic.LoadInt64(&d.top)
	
	if top <= b {
		// Non-empty
		t := d.tasks[b%int64(len(d.tasks))]
		if top == b {
			// Last element - race with steal
			if !atomic.CompareAndSwapInt64(&d.top, top, top+1) {
				// Lost race
				atomic.StoreInt64(&d.bottom, b+1)
				return Task{}, false
			}
			atomic.StoreInt64(&d.bottom, b+1)
		}
		return t, true
	}
	
	// Empty
	atomic.StoreInt64(&d.bottom, top)
	return Task{}, false
}

// Steal takes a task (thief, FIFO end)
func (d *WSDeque) Steal() (Task, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	top := atomic.LoadInt64(&d.top)
	b := atomic.LoadInt64(&d.bottom)
	
	if top >= b {
		return Task{}, false // empty
	}
	
	t := d.tasks[top%int64(len(d.tasks))]
	if !atomic.CompareAndSwapInt64(&d.top, top, top+1) {
		return Task{}, false // lost race
	}
	
	return t, true
}

// Len returns approximate size
func (d *WSDeque) Len() int {
	b := atomic.LoadInt64(&d.bottom)
	t := atomic.LoadInt64(&d.top)
	size := b - t
	if size < 0 {
		return 0
	}
	return int(size)
}

// ============================================================================
// ual Work-Stealing (using decoupled views)
// ============================================================================

// WSStack is a work-stealing stack using ual's decoupled views
type WSStack struct {
	stack     *Stack
	ownerView *View // LIFO - owner pops newest
	thiefView *View // FIFO - thief steals oldest
	
	// Pre-allocated buffers for capped stacks (avoids allocation on push)
	buffers   [][]byte
	bufSize   int
	bufHead   int // next buffer for FIFO steal to free
	bufTail   int // next buffer for LIFO push to use
}

// NewWSStack creates a ual-based work-stealing stack (unlimited)
func NewWSStack() *WSStack {
	s := NewStack(LIFO, TypeBytes)
	
	owner := NewView(LIFO)
	owner.Attach(s)
	
	thief := NewView(FIFO)
	thief.Attach(s)
	
	return &WSStack{
		stack:     s,
		ownerView: owner,
		thiefView: thief,
	}
}

// NewWSStackCapped creates a ual-based work-stealing stack with fixed capacity
// Pre-allocates buffers to avoid allocation on push
func NewWSStackCapped(capacity int) *WSStack {
	return NewWSStackCappedWithBufSize(capacity, 64) // default 64 bytes per task
}

// NewWSStackCappedWithBufSize creates a capped stack with custom buffer size
func NewWSStackCappedWithBufSize(capacity, bufSize int) *WSStack {
	s := NewCappedStack(LIFO, TypeBytes, capacity)
	
	owner := NewView(LIFO)
	owner.Attach(s)
	
	thief := NewView(FIFO)
	thief.Attach(s)
	
	// Pre-allocate all buffers
	buffers := make([][]byte, capacity)
	for i := range buffers {
		buffers[i] = make([]byte, bufSize)
	}
	
	return &WSStack{
		stack:     s,
		ownerView: owner,
		thiefView: thief,
		buffers:   buffers,
		bufSize:   bufSize,
		bufHead:   0,
		bufTail:   0,
	}
}

// Push adds a task (owner)
func (ws *WSStack) Push(t Task) bool {
	var data []byte
	
	if ws.buffers != nil {
		// Capped: use pre-allocated buffer at tail
		cap := len(ws.buffers)
		if (ws.bufTail+1)%cap == ws.bufHead && ws.stack.Len() > 0 {
			return false // full
		}
		data = ws.buffers[ws.bufTail%cap]
		encodeTaskInto(t, data)
		ws.bufTail++
	} else {
		// Unlimited: allocate
		data = encodeTask(t)
	}
	
	return ws.stack.Push(data) == nil
}

// Pop removes a task (owner, LIFO)
func (ws *WSStack) Pop() (Task, bool) {
	data, err := ws.ownerView.Pop()
	if err != nil {
		return Task{}, false
	}
	
	// LIFO pop frees tail buffer
	if ws.buffers != nil {
		ws.bufTail--
	}
	
	return decodeTask(data), true
}

// Steal takes a task (thief, FIFO)
func (ws *WSStack) Steal() (Task, bool) {
	data, err := ws.thiefView.Pop()
	if err != nil {
		return Task{}, false
	}
	
	// FIFO steal frees head buffer
	if ws.buffers != nil {
		ws.bufHead++
	}
	
	return decodeTask(data), true
}

// Len returns size
func (ws *WSStack) Len() int {
	return ws.stack.Len()
}

// Simple task encoding (ID as 8 bytes + data)
func encodeTask(t Task) []byte {
	b := make([]byte, 8+len(t.Data))
	encodeTaskInto(t, b)
	return b
}

// encodeTaskInto writes task into existing buffer (no allocation)
func encodeTaskInto(t Task, b []byte) {
	id := t.ID
	for i := 7; i >= 0; i-- {
		b[i] = byte(id & 0xff)
		id >>= 8
	}
	copy(b[8:], t.Data)
}

func decodeTask(b []byte) Task {
	if len(b) < 8 {
		return Task{}
	}
	var id int64
	for i := 0; i < 8; i++ {
		id = (id << 8) | int64(b[i])
	}
	return Task{ID: id, Data: b[8:]}
}
