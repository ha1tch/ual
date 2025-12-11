package ual

import (
	"runtime"
	"testing"
	"time"
)

func TestLIFOStack(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	
	// Push some values
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	if s.Len() != 3 {
		t.Errorf("expected len 3, got %d", s.Len())
	}
	
	// Pop should return last pushed (30)
	val, err := s.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 30 {
		t.Errorf("expected 30, got %d", bytesToInt(val))
	}
	
	// Next pop should be 20
	val, err = s.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 20 {
		t.Errorf("expected 20, got %d", bytesToInt(val))
	}
}

func TestFIFOStack(t *testing.T) {
	s := NewStack(FIFO, TypeInt64)
	
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	// Pop should return first pushed (10)
	val, err := s.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 10 {
		t.Errorf("expected 10, got %d", bytesToInt(val))
	}
	
	// Next should be 20
	val, err = s.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 20 {
		t.Errorf("expected 20, got %d", bytesToInt(val))
	}
}

func TestIndexedStack(t *testing.T) {
	s := NewStack(Indexed, TypeInt64)
	
	s.Push(intToBytes(100))
	s.Push(intToBytes(200))
	s.Push(intToBytes(300))
	
	// Pop at index 1 should return 200
	val, err := s.Pop(intToBytes(1))
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 200 {
		t.Errorf("expected 200, got %d", bytesToInt(val))
	}
	
	// Stack is now [100, 300]
	// Pop without index removes last element (array semantics)
	val, err = s.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 300 {
		t.Errorf("expected 300, got %d", bytesToInt(val))
	}
	
	// Stack is now [100]
	// Pop again should return 100
	val, err = s.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 100 {
		t.Errorf("expected 100, got %d", bytesToInt(val))
	}
	
	// Stack is empty, pop should fail
	_, err = s.Pop()
	if err == nil {
		t.Error("expected error for pop on empty stack")
	}
}

func TestHashStack(t *testing.T) {
	s := NewStack(Hash, TypeInt64)
	
	// Push with keys
	s.Push(intToBytes(100), []byte("foo"))
	s.Push(intToBytes(200), []byte("bar"))
	s.Push(intToBytes(300), []byte("baz"))
	
	// Pop by key
	val, err := s.Pop([]byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 200 {
		t.Errorf("expected 200, got %d", bytesToInt(val))
	}
	
	// Pop without key should fail
	_, err = s.Pop()
	if err == nil {
		t.Error("expected error for hash pop without key")
	}
	
	// Pop missing key should fail
	_, err = s.Pop([]byte("missing"))
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestBringSameType(t *testing.T) {
	src := NewStack(LIFO, TypeInt64)
	dst := NewStack(LIFO, TypeInt64)
	
	src.Push(intToBytes(42))
	src.Push(intToBytes(99))
	
	// Bring should move top element (99) from src to dst
	err := dst.Bring(src)
	if err != nil {
		t.Fatal(err)
	}
	
	if src.Len() != 1 {
		t.Errorf("source should have 1 element, got %d", src.Len())
	}
	if dst.Len() != 1 {
		t.Errorf("dest should have 1 element, got %d", dst.Len())
	}
	
	val, _ := dst.Pop()
	if bytesToInt(val) != 99 {
		t.Errorf("expected 99, got %d", bytesToInt(val))
	}
}

func TestBringStringToInt(t *testing.T) {
	src := NewStack(LIFO, TypeString)
	dst := NewStack(LIFO, TypeInt64)
	
	src.Push([]byte("42"))
	
	// Bring with base 10 conversion
	err := dst.Bring(src, intToBytes(10))
	if err != nil {
		t.Fatal(err)
	}
	
	val, _ := dst.Pop()
	if bytesToInt(val) != 42 {
		t.Errorf("expected 42, got %d", bytesToInt(val))
	}
}

func TestBringStringToIntHex(t *testing.T) {
	src := NewStack(LIFO, TypeString)
	dst := NewStack(LIFO, TypeInt64)
	
	src.Push([]byte("ff"))
	
	// Bring with base 16 conversion
	err := dst.Bring(src, intToBytes(16))
	if err != nil {
		t.Fatal(err)
	}
	
	val, _ := dst.Pop()
	if bytesToInt(val) != 255 {
		t.Errorf("expected 255, got %d", bytesToInt(val))
	}
}

func TestBringFailsAtomically(t *testing.T) {
	src := NewStack(LIFO, TypeString)
	dst := NewStack(LIFO, TypeInt64)
	
	src.Push([]byte("not_a_number"))
	
	// Bring should fail
	err := dst.Bring(src, intToBytes(10))
	if err == nil {
		t.Fatal("expected error for invalid conversion")
	}
	
	// Source should still have the element (atomic failure)
	if src.Len() != 1 {
		t.Errorf("source should still have 1 element after failed bring, got %d", src.Len())
	}
	
	// Dest should be empty
	if dst.Len() != 0 {
		t.Errorf("dest should be empty after failed bring, got %d", dst.Len())
	}
}

func TestPerspectiveSwitch(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	// LIFO: pop returns 30
	val, _ := s.Peek()
	if bytesToInt(val) != 30 {
		t.Errorf("LIFO peek expected 30, got %d", bytesToInt(val))
	}
	
	// Switch to FIFO
	s.SetPerspective(FIFO)
	
	// FIFO: peek returns 10
	val, _ = s.Peek()
	if bytesToInt(val) != 10 {
		t.Errorf("FIFO peek expected 10, got %d", bytesToInt(val))
	}
	
	// Switch to Indexed
	s.SetPerspective(Indexed)
	
	// Index 1 should be 20
	val, _ = s.Peek(intToBytes(1))
	if bytesToInt(val) != 20 {
		t.Errorf("Indexed[1] expected 20, got %d", bytesToInt(val))
	}
}

func TestIntToFloat(t *testing.T) {
	src := NewStack(LIFO, TypeInt64)
	dst := NewStack(LIFO, TypeFloat64)
	
	src.Push(intToBytes(42))
	
	err := dst.Bring(src)
	if err != nil {
		t.Fatal(err)
	}
	
	val, _ := dst.Pop()
	f := bytesToFloat64(val)
	if f != 42.0 {
		t.Errorf("expected 42.0, got %f", f)
	}
}

func TestFloatToInt(t *testing.T) {
	src := NewStack(LIFO, TypeFloat64)
	dst := NewStack(LIFO, TypeInt64)
	
	src.Push(float64ToBytes(3.7))
	
	err := dst.Bring(src)
	if err != nil {
		t.Fatal(err)
	}
	
	val, _ := dst.Pop()
	i := bytesToInt(val)
	if i != 3 { // truncation
		t.Errorf("expected 3 (truncated), got %d", i)
	}
}

func TestFreeze(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	
	s.Push(intToBytes(10))
	s.Push(intToBytes(20))
	s.Push(intToBytes(30))
	
	// Freeze the stack
	s.Freeze()
	
	if !s.IsFrozen() {
		t.Error("expected stack to be frozen")
	}
	
	// Peek should still work
	val, err := s.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 30 {
		t.Errorf("expected 30, got %d", bytesToInt(val))
	}
	
	// Push should fail
	err = s.Push(intToBytes(40))
	if err == nil {
		t.Error("expected error on push to frozen stack")
	}
	
	// Pop should fail
	_, err = s.Pop()
	if err == nil {
		t.Error("expected error on pop from frozen stack")
	}
	
	// Len should still work
	if s.Len() != 3 {
		t.Errorf("expected len 3, got %d", s.Len())
	}
	
	// SetPerspective should still work
	s.SetPerspective(FIFO)
	val, _ = s.Peek()
	if bytesToInt(val) != 10 {
		t.Errorf("expected 10 after FIFO switch, got %d", bytesToInt(val))
	}
}

func TestTakeBlocking(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	
	// Push in goroutine after delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.Push(intToBytes(42))
	}()
	
	// Take should block until data arrives
	val, err := s.Take()
	if err != nil {
		t.Fatalf("Take failed: %v", err)
	}
	if bytesToInt(val) != 42 {
		t.Errorf("expected 42, got %d", bytesToInt(val))
	}
}

func TestTakeTimeout(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	
	// Take with short timeout should fail
	_, err := s.Take(10) // 10ms
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestTakeTimeoutGoroutineLeak(t *testing.T) {
	// This test checks that goroutines don't leak when Take times out.
	// The goroutine spawned for timed wait should eventually be cleaned up.
	
	s := NewStack(LIFO, TypeInt64)
	
	initialGoroutines := runtime.NumGoroutine()
	
	// Do several Take timeouts
	for i := 0; i < 10; i++ {
		s.Take(1) // 1ms timeout
	}
	
	// Close the stack to signal any lingering goroutines
	s.Close()
	
	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)
	
	finalGoroutines := runtime.NumGoroutine()
	
	// Allow some slack (other goroutines may exist)
	leaked := finalGoroutines - initialGoroutines
	if leaked > 2 {
		t.Errorf("potential goroutine leak: started with %d, ended with %d (leaked %d)",
			initialGoroutines, finalGoroutines, leaked)
	}
}

func TestTakeWithClose(t *testing.T) {
	s := NewStack(LIFO, TypeInt64)
	
	// Close in goroutine
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.Close()
	}()
	
	// Take should return error when closed
	_, err := s.Take()
	if err == nil {
		t.Error("expected error when stack closed")
	}
}
