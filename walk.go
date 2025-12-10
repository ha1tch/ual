package ual

// WalkFunc is applied to each element during walk
type WalkFunc func(data []byte) ([]byte, error)

// Walk traverses source in perspective order, applies fn to each element,
// pushes results to destination. Errors go to errStack if provided.
// Source is NOT consumed (unlike bring).
func (dest *Stack) Walk(source *Stack, fn WalkFunc, errStack *Stack) {
	source.mu.RLock()
	defer source.mu.RUnlock()
	dest.mu.Lock()
	defer dest.mu.Unlock()
	
	// Determine traversal order based on source perspective
	indices := walkOrder(source)
	
	for _, idx := range indices {
		elem := source.elements[idx]
		result, err := fn(elem.data)
		
		if err != nil {
			if errStack != nil {
				// Push error info to error stack
				// Push handles its own locking
				errStack.Push([]byte(err.Error()))
			}
			continue // skip this element, continue with others
		}
		
		// Push result to dest
		if dest.perspective == Hash {
			// For hash dest during walk, use source key if available
			var key []byte
			if source.keys[idx] != nil {
				key = source.keys[idx]
			} else {
				key = intToBytes(int64(idx))
			}
			dest.elements = append(dest.elements, Element{data: result})
			dest.keys = append(dest.keys, key)
			dest.hashIdx[string(key)] = len(dest.elements) - 1
		} else {
			dest.elements = append(dest.elements, Element{data: result})
			dest.keys = append(dest.keys, nil)
		}
	}
}

// walkOrder returns indices in perspective order
func walkOrder(s *Stack) []int {
	n := len(s.elements) - s.head
	if n == 0 {
		return nil
	}
	
	indices := make([]int, 0, n)
	
	switch s.perspective {
	case LIFO:
		// Last to first (from end to head)
		for i := len(s.elements) - 1; i >= s.head; i-- {
			indices = append(indices, i)
		}
	case FIFO, Indexed:
		// First to last (from head to end)
		for i := s.head; i < len(s.elements); i++ {
			indices = append(indices, i)
		}
	case Hash:
		// Only non-tombstone entries
		for i := s.head; i < len(s.elements); i++ {
			if s.keys[i] != nil {
				indices = append(indices, i)
			}
		}
	}
	
	return indices
}

// Filter walks source, keeping only elements where predicate returns true
func (dest *Stack) Filter(source *Stack, pred func([]byte) bool, errStack *Stack) {
	// Custom walk that skips elements not matching predicate
	source.mu.RLock()
	defer source.mu.RUnlock()
	dest.mu.Lock()
	defer dest.mu.Unlock()
	
	indices := walkOrder(source)
	
	for _, idx := range indices {
		elem := source.elements[idx]
		if pred(elem.data) {
			if dest.perspective == Hash {
				var key []byte
				if source.keys[idx] != nil {
					key = source.keys[idx]
				} else {
					key = intToBytes(int64(idx))
				}
				dest.elements = append(dest.elements, Element{data: elem.data})
				dest.keys = append(dest.keys, key)
				dest.hashIdx[string(key)] = len(dest.elements) - 1
			} else {
				dest.elements = append(dest.elements, Element{data: elem.data})
				dest.keys = append(dest.keys, nil)
			}
		}
	}
}

// Map is a convenience wrapper: walk with transform, results to new stack
func Map(source *Stack, fn WalkFunc, destType ElementType, errStack *Stack) *Stack {
	dest := NewStack(source.perspective, destType)
	dest.Walk(source, fn, errStack)
	return dest
}

// Reduce walks source and accumulates a result
func Reduce(source *Stack, initial []byte, fn func(acc, elem []byte) []byte) []byte {
	source.mu.RLock()
	defer source.mu.RUnlock()
	
	acc := initial
	indices := walkOrder(source)
	
	for _, idx := range indices {
		acc = fn(acc, source.elements[idx].data)
	}
	
	return acc
}
