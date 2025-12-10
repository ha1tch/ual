package ual

import (
	"testing"
)

func TestViewAttachDetach(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	
	v := NewView(FIFO)
	
	if v.Stack() != nil {
		t.Error("expected nil stack before attach")
	}
	
	v.Attach(s)
	
	if v.Stack() != s {
		t.Error("expected attached stack")
	}
	
	v.Detach()
	
	if v.Stack() != nil {
		t.Error("expected nil stack after detach")
	}
}

func TestViewPeekLIFO(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	v := NewView(LIFO)
	v.Attach(s)
	
	// LIFO peek should return last element
	val, err := v.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 30 {
		t.Errorf("expected 30, got %d", bytesToInt(val))
	}
}

func TestViewPeekFIFO(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	v := NewView(FIFO)
	v.Attach(s)
	
	// FIFO peek should return first element
	val, err := v.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 10 {
		t.Errorf("expected 10, got %d", bytesToInt(val))
	}
}

func TestViewCursorAdvance(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	v := NewView(FIFO)
	v.Attach(s)
	
	// First peek
	val, _ := v.Peek()
	if bytesToInt(val) != 10 {
		t.Errorf("expected 10, got %d", bytesToInt(val))
	}
	
	// Advance cursor
	v.Advance()
	
	// Second peek
	val, _ = v.Peek()
	if bytesToInt(val) != 20 {
		t.Errorf("expected 20, got %d", bytesToInt(val))
	}
	
	// Advance again
	v.Advance()
	
	val, _ = v.Peek()
	if bytesToInt(val) != 30 {
		t.Errorf("expected 30, got %d", bytesToInt(val))
	}
}

func TestViewRemaining(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	v := NewView(FIFO)
	v.Attach(s)
	
	if v.Remaining() != 3 {
		t.Errorf("expected 3 remaining, got %d", v.Remaining())
	}
	
	v.Advance()
	if v.Remaining() != 2 {
		t.Errorf("expected 2 remaining, got %d", v.Remaining())
	}
	
	v.Advance()
	v.Advance()
	if v.Remaining() != 0 {
		t.Errorf("expected 0 remaining, got %d", v.Remaining())
	}
}

func TestMultipleViewsSameStack(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	lifoView := NewView(LIFO)
	fifoView := NewView(FIFO)
	
	lifoView.Attach(s)
	fifoView.Attach(s)
	
	// LIFO sees last
	lifoVal, _ := lifoView.Peek()
	if bytesToInt(lifoVal) != 30 {
		t.Errorf("LIFO expected 30, got %d", bytesToInt(lifoVal))
	}
	
	// FIFO sees first
	fifoVal, _ := fifoView.Peek()
	if bytesToInt(fifoVal) != 10 {
		t.Errorf("FIFO expected 10, got %d", bytesToInt(fifoVal))
	}
	
	// Cursors are independent
	fifoView.Advance()
	fifoView.Advance()
	
	fifoVal, _ = fifoView.Peek()
	if bytesToInt(fifoVal) != 30 {
		t.Errorf("FIFO after advance expected 30, got %d", bytesToInt(fifoVal))
	}
	
	// LIFO cursor unchanged
	lifoVal, _ = lifoView.Peek()
	if bytesToInt(lifoVal) != 30 {
		t.Errorf("LIFO should still be 30, got %d", bytesToInt(lifoVal))
	}
}

func TestWorkStealingPattern(t *testing.T) {
	// Classic work-stealing: owner uses LIFO (hot), thief uses FIFO (cold)
	tasks := NewStack(LIFO, TypeInt64)
	tasks.Push(intToBytes(1))
	tasks.Push(intToBytes(2))
	tasks.Push(intToBytes(3))
	tasks.Push(intToBytes(4))
	tasks.Push(intToBytes(5))
	
	ownerView := NewView(LIFO)  // pops from end
	thiefView := NewView(FIFO)  // steals from beginning
	
	ownerView.Attach(tasks)
	thiefView.Attach(tasks)
	
	// Owner pops task 5 (newest, cache-hot)
	ownerTask, _ := ownerView.Pop()
	if bytesToInt(ownerTask) != 5 {
		t.Errorf("owner expected 5, got %d", bytesToInt(ownerTask))
	}
	
	// Thief steals task 1 (oldest, cache-cold)
	thiefTask, _ := thiefView.Pop()
	if bytesToInt(thiefTask) != 1 {
		t.Errorf("thief expected 1, got %d", bytesToInt(thiefTask))
	}
	
	// Owner pops task 4
	ownerTask, _ = ownerView.Pop()
	if bytesToInt(ownerTask) != 4 {
		t.Errorf("owner expected 4, got %d", bytesToInt(ownerTask))
	}
	
	// Thief steals task 2
	thiefTask, _ = thiefView.Pop()
	if bytesToInt(thiefTask) != 2 {
		t.Errorf("thief expected 2, got %d", bytesToInt(thiefTask))
	}
	
	// Only task 3 remains
	if tasks.Len() != 1 {
		t.Errorf("expected 1 task remaining, got %d", tasks.Len())
	}
	
	// Either can take it
	lastTask, _ := ownerView.Pop()
	if bytesToInt(lastTask) != 3 {
		t.Errorf("expected 3, got %d", bytesToInt(lastTask))
	}
	
	if tasks.Len() != 0 {
		t.Errorf("expected 0 tasks remaining, got %d", tasks.Len())
	}
}

func TestViewHashPerspective(t *testing.T) {
	s := NewStack(Hash, TypeInt64)
	s.Push(intToBytes(100), []byte("a"))
	s.Push(intToBytes(200), []byte("b"))
	s.Push(intToBytes(300), []byte("c"))
	
	v := NewView(Hash)
	v.Attach(s)
	
	// Lookup by key
	val, err := v.Peek([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 200 {
		t.Errorf("expected 200, got %d", bytesToInt(val))
	}
	
	// Pop by key
	val, _ = v.Pop([]byte("a"))
	if bytesToInt(val) != 100 {
		t.Errorf("expected 100, got %d", bytesToInt(val))
	}
	
	// Key should be gone
	_, err = v.Peek([]byte("a"))
	if err == nil {
		t.Error("expected error for removed key")
	}
}

func TestViewWalk(t *testing.T) {
	s := NewStack(FIFO, TypeInt64)
	s.Push(intToBytes(1))
	s.Push(intToBytes(2))
	s.Push(intToBytes(3))
	
	v := NewView(FIFO)
	v.Attach(s)
	
	// Advance cursor past first element
	v.Advance()
	
	// Walk should start from cursor
	dest := NewStack(FIFO, TypeInt64)
	double := func(data []byte) ([]byte, error) {
		return intToBytes(bytesToInt(data) * 2), nil
	}
	
	v.Walk(double, dest, nil)
	
	// Should only have 2 elements (walked from cursor)
	if dest.Len() != 2 {
		t.Errorf("expected 2 elements, got %d", dest.Len())
	}
	
	val, _ := dest.Pop()
	if bytesToInt(val) != 4 { // 2 * 2
		t.Errorf("expected 4, got %d", bytesToInt(val))
	}
}

func TestViewPerspectiveSwitch(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	v := NewView(LIFO)
	v.Attach(s)
	
	// LIFO peek
	val, _ := v.Peek()
	if bytesToInt(val) != 30 {
		t.Errorf("LIFO expected 30, got %d", bytesToInt(val))
	}
	
	// Switch to FIFO
	v.SetPerspective(FIFO)
	
	// FIFO peek (cursor reset to 0)
	val, _ = v.Peek()
	if bytesToInt(val) != 10 {
		t.Errorf("FIFO expected 10, got %d", bytesToInt(val))
	}
}

func TestViewOnFrozenStack(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Freeze()
	
	v := NewView(FIFO)
	v.Attach(s)
	
	// Peek should work
	val, err := v.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 10 {
		t.Errorf("expected 10, got %d", bytesToInt(val))
	}
	
	// Pop should fail
	_, err = v.Pop()
	if err == nil {
		t.Error("expected error popping from frozen stack through view")
	}
}

func TestMoveViewBetweenStacks(t *testing.T) {
	s1 := NewStack(LIFO, TypeInt64)
	s1.Push(intToBytes(10))
	s1.Push(intToBytes(20))
	
	s2 := NewStack(LIFO, TypeInt64)
	s2.Push(intToBytes(100))
	s2.Push(intToBytes(200))
	
	v := NewView(FIFO)
	
	// Attach to first stack
	v.Attach(s1)
	val, _ := v.Peek()
	if bytesToInt(val) != 10 {
		t.Errorf("on s1 expected 10, got %d", bytesToInt(val))
	}
	
	// Move to second stack
	v.Attach(s2)
	val, _ = v.Peek()
	if bytesToInt(val) != 100 {
		t.Errorf("on s2 expected 100, got %d", bytesToInt(val))
	}
}
