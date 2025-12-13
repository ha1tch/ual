package runtime

import (
	"errors"
	"sync"
)

// View is a decoupled perspective attached to a stack.
// Multiple views can attach to the same stack with independent cursors.
type View struct {
	mu          sync.Mutex
	stack       *Stack
	perspective Perspective
	
	// Cursor position - interpretation depends on perspective
	// LIFO: offset from end (0 = last element)
	// FIFO: offset from logical start (0 = first element)
	// Indexed: absolute position
	// Hash: not used (hash uses key lookup)
	cursor int
	
	// Hash index - built on attach for Hash perspective
	hashIdx map[string]int
}

// NewView creates a view with the given perspective, not yet attached
func NewView(p Perspective) *View {
	return &View{
		perspective: p,
	}
}

// Attach connects this view to a stack and initializes cursor state
func (v *View) Attach(s *Stack) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.stack = s
	v.cursor = 0
	
	if v.perspective == Hash {
		v.rebuildHashIndex()
	}
	
	return nil
}

// rebuildHashIndex builds hash lookup from stack's current state
func (v *View) rebuildHashIndex() {
	v.stack.mu.RLock()
	defer v.stack.mu.RUnlock()
	
	v.hashIdx = make(map[string]int)
	for i := v.stack.head; i < len(v.stack.elements); i++ {
		if v.stack.keys[i] != nil {
			v.hashIdx[string(v.stack.keys[i])] = i
		}
	}
}

// Detach disconnects this view from its stack
func (v *View) Detach() {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.stack = nil
	v.hashIdx = nil
	v.cursor = 0
}

// Stack returns the attached stack (or nil)
func (v *View) Stack() *Stack {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.stack
}

// Perspective returns this view's access mode
func (v *View) Perspective() Perspective {
	return v.perspective
}

// SetPerspective changes the access mode and resets cursor
func (v *View) SetPerspective(p Perspective) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.perspective = p
	v.cursor = 0
	
	if p == Hash && v.stack != nil {
		v.rebuildHashIndex()
	} else {
		v.hashIdx = nil
	}
}

// Peek returns element at cursor without advancing
func (v *View) Peek(param ...[]byte) ([]byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.stack == nil {
		return nil, errors.New("view not attached")
	}
	
	v.stack.mu.RLock()
	defer v.stack.mu.RUnlock()
	
	idx, err := v.resolveIndex(param)
	if err != nil {
		return nil, err
	}
	
	return v.stack.elements[idx].data, nil
}

// Advance moves cursor forward in perspective order
func (v *View) Advance() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.stack == nil {
		return errors.New("view not attached")
	}
	
	if v.perspective == Hash {
		return errors.New("hash perspective has no cursor to advance")
	}
	
	v.cursor++
	return nil
}

// Reset moves cursor back to start
func (v *View) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cursor = 0
}

// Cursor returns current cursor position
func (v *View) Cursor() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.cursor
}

// SetCursor sets cursor to specific position
func (v *View) SetCursor(pos int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cursor = pos
}

// Remaining returns count of elements from cursor to end (in perspective order)
func (v *View) Remaining() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.stack == nil {
		return 0
	}
	
	v.stack.mu.RLock()
	defer v.stack.mu.RUnlock()
	
	size := len(v.stack.elements) - v.stack.head
	
	switch v.perspective {
	case LIFO, FIFO, Indexed:
		remaining := size - v.cursor
		if remaining < 0 {
			return 0
		}
		return remaining
	case Hash:
		return len(v.hashIdx)
	}
	
	return 0
}

// resolveIndex converts perspective + cursor + optional param to actual slice index
// Must be called with v.mu and v.stack.mu held
func (v *View) resolveIndex(param [][]byte) (int, error) {
	size := len(v.stack.elements) - v.stack.head
	if size == 0 {
		return 0, errors.New("stack empty")
	}
	
	var idx int
	
	switch v.perspective {
	case LIFO:
		// Count from end, cursor is offset from end
		offset := v.cursor
		if len(param) > 0 {
			offset = int(bytesToInt(param[0]))
		}
		idx = len(v.stack.elements) - 1 - offset
		
	case FIFO:
		// Count from head, cursor is offset from head
		offset := v.cursor
		if len(param) > 0 {
			offset = int(bytesToInt(param[0]))
		}
		idx = v.stack.head + offset
		
	case Indexed:
		// Direct position, cursor or param
		pos := v.cursor
		if len(param) > 0 {
			pos = int(bytesToInt(param[0]))
		}
		idx = v.stack.head + pos
		
	case Hash:
		if len(param) == 0 {
			return 0, errors.New("hash perspective requires key")
		}
		keyStr := string(param[0])
		var exists bool
		idx, exists = v.hashIdx[keyStr]
		if !exists {
			return 0, errors.New("key not found")
		}
	}
	
	if idx < v.stack.head || idx >= len(v.stack.elements) {
		return 0, errors.New("index out of bounds")
	}
	
	// Check for tombstone (hash perspective)
	if v.stack.elements[idx].data == nil && v.perspective != Hash {
		// For non-hash, this shouldn't happen, but check anyway
	}
	
	return idx, nil
}

// Pop removes and returns element at cursor position.
// For LIFO: removes from end (tail)
// For FIFO: removes from head
// Other views attached to the same stack may need to adjust.
func (v *View) Pop(param ...[]byte) ([]byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.stack == nil {
		return nil, errors.New("view not attached")
	}
	
	v.stack.mu.Lock()
	defer v.stack.mu.Unlock()
	
	if v.stack.frozen {
		return nil, errors.New("stack is frozen")
	}
	
	size := len(v.stack.elements) - v.stack.head
	if size == 0 {
		return nil, errors.New("stack empty")
	}
	
	var elem Element
	
	switch v.perspective {
	case LIFO:
		// Pop from end - O(1)
		// Cursor is offset from end; default (cursor=0) means last element
		offset := v.cursor
		if len(param) > 0 {
			offset = int(bytesToInt(param[0]))
		}
		idx := len(v.stack.elements) - 1 - offset
		
		if idx < v.stack.head || idx >= len(v.stack.elements) {
			return nil, errors.New("index out of bounds")
		}
		
		elem = v.stack.elements[idx]
		
		if offset == 0 {
			// Fast path: just shrink slice
			v.stack.elements = v.stack.elements[:idx]
			v.stack.keys = v.stack.keys[:idx]
		} else {
			// Slow path: shift
			v.stack.elements = append(v.stack.elements[:idx], v.stack.elements[idx+1:]...)
			v.stack.keys = append(v.stack.keys[:idx], v.stack.keys[idx+1:]...)
		}
		
	case FIFO:
		// Pop from head - O(1)
		offset := v.cursor
		if len(param) > 0 {
			offset = int(bytesToInt(param[0]))
		}
		idx := v.stack.head + offset
		
		if idx < v.stack.head || idx >= len(v.stack.elements) {
			return nil, errors.New("index out of bounds")
		}
		
		elem = v.stack.elements[idx]
		
		if offset == 0 {
			// Fast path: just advance head
			v.stack.head++
			if v.stack.head > len(v.stack.elements)/2 && v.stack.head > 100 {
				v.stack.compact()
			}
		} else {
			// Slow path: shift
			v.stack.elements = append(v.stack.elements[:idx], v.stack.elements[idx+1:]...)
			v.stack.keys = append(v.stack.keys[:idx], v.stack.keys[idx+1:]...)
		}
		
	case Indexed:
		if len(param) == 0 {
			return nil, errors.New("indexed perspective requires position")
		}
		idx := v.stack.head + int(bytesToInt(param[0]))
		if idx < v.stack.head || idx >= len(v.stack.elements) {
			return nil, errors.New("index out of bounds")
		}
		elem = v.stack.elements[idx]
		v.stack.elements = append(v.stack.elements[:idx], v.stack.elements[idx+1:]...)
		v.stack.keys = append(v.stack.keys[:idx], v.stack.keys[idx+1:]...)
		
	case Hash:
		if len(param) == 0 {
			return nil, errors.New("hash perspective requires key")
		}
		keyStr := string(param[0])
		idx, exists := v.hashIdx[keyStr]
		if !exists {
			return nil, errors.New("key not found")
		}
		elem = v.stack.elements[idx]
		
		// Remove from both view's index and stack's index if present
		delete(v.hashIdx, keyStr)
		if v.stack.hashIdx != nil {
			delete(v.stack.hashIdx, keyStr)
		}
		
		// Tombstone
		v.stack.elements[idx] = Element{}
		v.stack.keys[idx] = nil
	}
	
	return elem.data, nil
}

// Walk traverses stack from cursor position in perspective order
func (v *View) Walk(fn WalkFunc, dest *Stack, errStack *Stack) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.stack == nil {
		return
	}
	
	v.stack.mu.RLock()
	indices := v.walkIndices()
	elements := make([]Element, len(indices))
	keys := make([][]byte, len(indices))
	for i, idx := range indices {
		elements[i] = v.stack.elements[idx]
		keys[i] = v.stack.keys[idx]
	}
	v.stack.mu.RUnlock()
	
	// Process outside of stack lock
	if dest != nil {
		dest.mu.Lock()
		defer dest.mu.Unlock()
	}
	
	for i, elem := range elements {
		result, err := fn(elem.data)
		if err != nil {
			if errStack != nil {
				errStack.Push([]byte(err.Error()))
			}
			continue
		}
		
		if dest != nil {
			if dest.perspective == Hash {
				key := keys[i]
				if key == nil {
					key = intToBytes(int64(i))
				}
				dest.elements = append(dest.elements, Element{data: result})
				dest.keys = append(dest.keys, key)
				if dest.hashIdx == nil {
					dest.hashIdx = make(map[string]int)
				}
				dest.hashIdx[string(key)] = len(dest.elements) - 1
			} else {
				dest.elements = append(dest.elements, Element{data: result})
				dest.keys = append(dest.keys, nil)
			}
		}
	}
}

// walkIndices returns indices to walk based on perspective and cursor
// Must be called with v.stack.mu held
func (v *View) walkIndices() []int {
	size := len(v.stack.elements) - v.stack.head
	if size == 0 {
		return nil
	}
	
	indices := make([]int, 0, size)
	
	switch v.perspective {
	case LIFO:
		// From cursor position (offset from end) to head
		start := len(v.stack.elements) - 1 - v.cursor
		for i := start; i >= v.stack.head; i-- {
			indices = append(indices, i)
		}
		
	case FIFO, Indexed:
		// From cursor position to end
		start := v.stack.head + v.cursor
		for i := start; i < len(v.stack.elements); i++ {
			indices = append(indices, i)
		}
		
	case Hash:
		// All non-tombstone entries
		for i := v.stack.head; i < len(v.stack.elements); i++ {
			if v.stack.keys[i] != nil {
				indices = append(indices, i)
			}
		}
	}
	
	return indices
}
