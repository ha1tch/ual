package runtime

import (
	"errors"
	"testing"
)

func TestWalkLIFO(t *testing.T) {
	src := NewStack(LIFO, TypeInt64)
	dst := NewStack(LIFO, TypeInt64)
	
	src.Push(intToBytes(1))
	src.Push(intToBytes(2))
	src.Push(intToBytes(3))
	
	// Square each element
	square := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		return intToBytes(n * n), nil
	}
	
	dst.Walk(src, square, nil)
	
	// Source unchanged
	if src.Len() != 3 {
		t.Errorf("source should be unchanged, got len %d", src.Len())
	}
	
	// Dest has squared values
	if dst.Len() != 3 {
		t.Errorf("dest should have 3 elements, got %d", dst.Len())
	}
	
	// LIFO walk order: 3, 2, 1 -> 9, 4, 1
	// But pushed in that order, so popping LIFO gives: 1, 4, 9
	val, _ := dst.Pop()
	if bytesToInt(val) != 1 {
		t.Errorf("expected 1, got %d", bytesToInt(val))
	}
	val, _ = dst.Pop()
	if bytesToInt(val) != 4 {
		t.Errorf("expected 4, got %d", bytesToInt(val))
	}
	val, _ = dst.Pop()
	if bytesToInt(val) != 9 {
		t.Errorf("expected 9, got %d", bytesToInt(val))
	}
}

func TestWalkFIFO(t *testing.T) {
	src := NewStack(FIFO, TypeInt64)
	dst := NewStack(FIFO, TypeInt64)
	
	src.Push(intToBytes(1))
	src.Push(intToBytes(2))
	src.Push(intToBytes(3))
	
	double := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		return intToBytes(n * 2), nil
	}
	
	dst.Walk(src, double, nil)
	
	// FIFO order: 1, 2, 3 -> 2, 4, 6
	// Popping FIFO: 2, 4, 6
	val, _ := dst.Pop()
	if bytesToInt(val) != 2 {
		t.Errorf("expected 2, got %d", bytesToInt(val))
	}
	val, _ = dst.Pop()
	if bytesToInt(val) != 4 {
		t.Errorf("expected 4, got %d", bytesToInt(val))
	}
}

func TestWalkWithErrors(t *testing.T) {
	src := NewStack(LIFO, TypeInt64)
	dst := NewStack(LIFO, TypeInt64)
	errStack := NewStack(LIFO, TypeString)
	
	src.Push(intToBytes(10))
	src.Push(intToBytes(0))  // will cause error
	src.Push(intToBytes(20))
	
	// Function that errors on zero
	noZero := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		if n == 0 {
			return nil, errors.New("zero not allowed")
		}
		return intToBytes(n * 10), nil
	}
	
	dst.Walk(src, noZero, errStack)
	
	// Dest should have 2 elements (skipped zero)
	if dst.Len() != 2 {
		t.Errorf("dest should have 2 elements, got %d", dst.Len())
	}
	
	// Error stack should have 1 error
	if errStack.Len() != 1 {
		t.Errorf("error stack should have 1 error, got %d", errStack.Len())
	}
	
	errMsg, _ := errStack.Pop()
	if string(errMsg) != "zero not allowed" {
		t.Errorf("unexpected error message: %s", string(errMsg))
	}
}

func TestFilter(t *testing.T) {
	src := NewStack(FIFO, TypeInt64)
	dst := NewStack(FIFO, TypeInt64)
	
	src.Push(intToBytes(1))
	src.Push(intToBytes(2))
	src.Push(intToBytes(3))
	src.Push(intToBytes(4))
	src.Push(intToBytes(5))
	
	// Keep only even numbers
	isEven := func(data []byte) bool {
		return bytesToInt(data)%2 == 0
	}
	
	dst.Filter(src, isEven, nil)
	
	if dst.Len() != 2 {
		t.Errorf("expected 2 even numbers, got %d", dst.Len())
	}
	
	val, _ := dst.Pop()
	if bytesToInt(val) != 2 {
		t.Errorf("expected 2, got %d", bytesToInt(val))
	}
	val, _ = dst.Pop()
	if bytesToInt(val) != 4 {
		t.Errorf("expected 4, got %d", bytesToInt(val))
	}
}

func TestReduce(t *testing.T) {
	src := NewStack(FIFO, TypeInt64)
	
	src.Push(intToBytes(1))
	src.Push(intToBytes(2))
	src.Push(intToBytes(3))
	src.Push(intToBytes(4))
	
	// Sum all elements
	sum := func(acc, elem []byte) []byte {
		a := bytesToInt(acc)
		e := bytesToInt(elem)
		return intToBytes(a + e)
	}
	
	result := Reduce(src, intToBytes(0), sum)
	
	if bytesToInt(result) != 10 {
		t.Errorf("expected sum 10, got %d", bytesToInt(result))
	}
}

func TestMap(t *testing.T) {
	src := NewStack(FIFO, TypeInt64)
	errStack := NewStack(LIFO, TypeString)
	
	src.Push(intToBytes(1))
	src.Push(intToBytes(2))
	src.Push(intToBytes(3))
	
	// Convert to strings
	toString := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		s := []byte{'0' + byte(n)}
		return s, nil
	}
	
	dst := Map(src, toString, TypeString, errStack)
	
	if dst.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", dst.Len())
	}
	
	// Should preserve source perspective (FIFO)
	val, _ := dst.Pop()
	if string(val) != "1" {
		t.Errorf("expected '1', got '%s'", string(val))
	}
}

func TestWalkHashPerspective(t *testing.T) {
	src := NewStack(Hash, TypeInt64)
	dst := NewStack(Hash, TypeInt64)
	
	src.Push(intToBytes(100), []byte("a"))
	src.Push(intToBytes(200), []byte("b"))
	src.Push(intToBytes(300), []byte("c"))
	
	double := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		return intToBytes(n * 2), nil
	}
	
	dst.Walk(src, double, nil)
	
	// Should preserve keys
	val, err := dst.Pop([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	if bytesToInt(val) != 400 {
		t.Errorf("expected 400, got %d", bytesToInt(val))
	}
}

func TestWalkFrozenStack(t *testing.T) {
	src := NewStack(FIFO, TypeInt64)
	dst := NewStack(FIFO, TypeInt64)
	
	src.Push(intToBytes(1))
	src.Push(intToBytes(2))
	src.Push(intToBytes(3))
	
	// Freeze source
	src.Freeze()
	
	double := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		return intToBytes(n * 2), nil
	}
	
	// Walk should still work on frozen source
	dst.Walk(src, double, nil)
	
	if dst.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", dst.Len())
	}
	
	val, _ := dst.Pop()
	if bytesToInt(val) != 2 {
		t.Errorf("expected 2, got %d", bytesToInt(val))
	}
}
