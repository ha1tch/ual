package ual

import (
	"sync"
	"sync/atomic"
	"testing"
)

// Push benchmarks

func BenchmarkPushLIFO(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	data := intToBytes(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(data)
	}
}

func BenchmarkPushFIFO(b *testing.B) {
	s := NewStack(FIFO, TypeInt64)
	data := intToBytes(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(data)
	}
}

func BenchmarkPushHash(b *testing.B) {
	s := NewStack(Hash, TypeInt64)
	data := intToBytes(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := intToBytes(int64(i))
		s.Push(data, key)
	}
}

// Pop benchmarks

func BenchmarkPopLIFO(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	data := intToBytes(42)
	for i := 0; i < b.N; i++ {
		s.Push(data)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkPopFIFO(b *testing.B) {
	s := NewStack(FIFO, TypeInt64)
	data := intToBytes(42)
	for i := 0; i < b.N; i++ {
		s.Push(data)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkPopHash(b *testing.B) {
	s := NewStack(Hash, TypeInt64)
	data := intToBytes(42)
	keys := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = intToBytes(int64(i))
		s.Push(data, keys[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop(keys[i])
	}
}

// Peek benchmarks

func BenchmarkPeekLIFO(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Peek()
	}
}

func BenchmarkPeekHash(b *testing.B) {
	s := NewStack(Hash, TypeInt64)
	key := []byte("target")
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)), intToBytes(int64(i)))
	}
	s.Push(intToBytes(999), key)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Peek(key)
	}
}

// Bring benchmarks

func BenchmarkBringSameType(b *testing.B) {
	src := NewStack(LIFO, TypeInt64)
	dst := NewStack(LIFO, TypeInt64)
	data := intToBytes(42)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src.Push(data)
		dst.Bring(src)
	}
}

func BenchmarkBringStringToInt(b *testing.B) {
	src := NewStack(LIFO, TypeString)
	dst := NewStack(LIFO, TypeInt64)
	data := []byte("12345")
	base := intToBytes(10)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src.Push(data)
		dst.Bring(src, base)
	}
}

func BenchmarkBringIntToFloat(b *testing.B) {
	src := NewStack(LIFO, TypeInt64)
	dst := NewStack(LIFO, TypeFloat64)
	data := intToBytes(42)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src.Push(data)
		dst.Bring(src)
	}
}

// Walk benchmarks

func BenchmarkWalk100(b *testing.B) {
	src := NewStack(FIFO, TypeInt64)
	for i := 0; i < 100; i++ {
		src.Push(intToBytes(int64(i)))
	}
	
	square := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		return intToBytes(n * n), nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := NewStack(FIFO, TypeInt64)
		dst.Walk(src, square, nil)
	}
}

func BenchmarkWalk1000(b *testing.B) {
	src := NewStack(FIFO, TypeInt64)
	for i := 0; i < 1000; i++ {
		src.Push(intToBytes(int64(i)))
	}
	
	square := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		return intToBytes(n * n), nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := NewStack(FIFO, TypeInt64)
		dst.Walk(src, square, nil)
	}
}

func BenchmarkWalk10000(b *testing.B) {
	src := NewStack(FIFO, TypeInt64)
	for i := 0; i < 10000; i++ {
		src.Push(intToBytes(int64(i)))
	}
	
	square := func(data []byte) ([]byte, error) {
		n := bytesToInt(data)
		return intToBytes(n * n), nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := NewStack(FIFO, TypeInt64)
		dst.Walk(src, square, nil)
	}
}

// Reduce benchmarks

func BenchmarkReduce1000(b *testing.B) {
	src := NewStack(FIFO, TypeInt64)
	for i := 0; i < 1000; i++ {
		src.Push(intToBytes(int64(i)))
	}
	
	sum := func(acc, elem []byte) []byte {
		return intToBytes(bytesToInt(acc) + bytesToInt(elem))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reduce(src, intToBytes(0), sum)
	}
}

// Filter benchmarks

func BenchmarkFilter1000(b *testing.B) {
	src := NewStack(FIFO, TypeInt64)
	for i := 0; i < 1000; i++ {
		src.Push(intToBytes(int64(i)))
	}
	
	isEven := func(data []byte) bool {
		return bytesToInt(data)%2 == 0
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := NewStack(FIFO, TypeInt64)
		dst.Filter(src, isEven, nil)
	}
}

// Perspective switch benchmark

func BenchmarkPerspectiveSwitch(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	for i := 0; i < 100; i++ {
		s.Push(intToBytes(int64(i)))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetPerspective(FIFO)
		s.SetPerspective(Hash)
		s.SetPerspective(LIFO)
	}
}

// Memory allocation benchmarks

func BenchmarkPushPopCycle(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	data := intToBytes(42)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(data)
		s.Pop()
	}
}

// Comparison: Hash lookup vs Indexed lookup

func BenchmarkIndexedLookup(b *testing.B) {
	s := NewStack(Indexed, TypeInt64)
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	idx := intToBytes(500)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Peek(idx)
	}
}

func BenchmarkHashLookup(b *testing.B) {
	s := NewStack(Hash, TypeInt64)
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)), intToBytes(int64(i)))
	}
	key := intToBytes(500)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Peek(key)
	}
}

// View benchmarks

func BenchmarkViewPeekLIFO(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	v := NewView(LIFO)
	v.Attach(s)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Peek()
	}
}

func BenchmarkViewPeekFIFO(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	v := NewView(FIFO)
	v.Attach(s)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Peek()
	}
}

func BenchmarkWorkStealingPopOwner(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	owner := NewView(LIFO)
	owner.Attach(s)
	data := intToBytes(42)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(data)
		owner.Pop()
	}
}

func BenchmarkWorkStealingPopThief(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	thief := NewView(FIFO)
	thief.Attach(s)
	data := intToBytes(42)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(data)
		thief.Pop()
	}
}

func BenchmarkWorkStealingContention(b *testing.B) {
	// Simulate owner and thief alternating
	s := NewStack(LIFO, TypeInt64)
	owner := NewView(LIFO)
	thief := NewView(FIFO)
	owner.Attach(s)
	thief.Attach(s)
	data := intToBytes(42)
	
	// Pre-fill
	for i := 0; i < 100; i++ {
		s.Push(data)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(data)
		s.Push(data)
		owner.Pop()  // owner takes from end
		thief.Pop()  // thief steals from beginning
	}
}

// ============================================================
// ual vs Native Go: Direct Comparisons
// ============================================================

// Sum 10,000 integers
func BenchmarkSumUalStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewStack(LIFO, TypeInt64)
		for j := 0; j < 10000; j++ {
			s.Push(intToBytes(int64(j)))
		}
		sum := int64(0)
		for s.Len() > 0 {
			v, _ := s.Pop()
			sum += bytesToInt(v)
		}
		_ = sum
	}
}

func BenchmarkSumNativeSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice := make([]int64, 0, 10000)
		for j := 0; j < 10000; j++ {
			slice = append(slice, int64(j))
		}
		sum := int64(0)
		for _, v := range slice {
			sum += v
		}
		_ = sum
	}
}

// Fibonacci: stack-based vs variables
func BenchmarkFibUalStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewStack(LIFO, TypeInt64)
		n := 30
		s.Push(intToBytes(0))
		s.Push(intToBytes(1))
		
		for j := 2; j <= n; j++ {
			b1, _ := s.Pop()
			b2, _ := s.Peek()
			s.Push(b1)
			next := bytesToInt(b1) + bytesToInt(b2)
			s.Push(intToBytes(next))
		}
		result, _ := s.Peek()
		_ = bytesToInt(result)
	}
}

func BenchmarkFibNativeVars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := 30
		a, bb := int64(0), int64(1)
		for j := 2; j <= n; j++ {
			a, bb = bb, a+bb
		}
		_ = bb
	}
}

// ============================================================
// Concurrent Work-Stealing: The Main Event
// ============================================================

func BenchmarkConcurrent4Thieves(b *testing.B) {
	s := NewCappedStack(LIFO, TypeInt64, 100000)
	
	// Pre-fill
	for i := 0; i < 50000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		thief := NewView(FIFO)
		thief.Attach(s)
		for pb.Next() {
			thief.Pop()
		}
	})
}

func BenchmarkConcurrentChannelDistribute(b *testing.B) {
	ch := make(chan int64, 50000)
	for i := 0; i < 50000; i++ {
		ch <- int64(i)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			select {
			case <-ch:
			default:
			}
		}
	})
}

// Producer-consumer pattern
func BenchmarkProducerConsumerUal(b *testing.B) {
	var wg sync.WaitGroup
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewCappedStack(LIFO, TypeInt64, 200)
		
		wg.Add(2)
		go func() {
			for j := 0; j < 100; j++ {
				s.Push(intToBytes(int64(j)))
			}
			wg.Done()
		}()
		
		go func() {
			v := NewView(LIFO)
			v.Attach(s)
			consumed := 0
			for consumed < 100 {
				if _, err := v.Pop(); err == nil {
					consumed++
				}
			}
			wg.Done()
		}()
		
		wg.Wait()
	}
}

func BenchmarkProducerConsumerChannel(b *testing.B) {
	var wg sync.WaitGroup
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan int64, 100)
		
		wg.Add(2)
		go func() {
			for j := 0; j < 100; j++ {
				ch <- int64(j)
			}
			close(ch)
			wg.Done()
		}()
		
		go func() {
			for range ch {
			}
			wg.Done()
		}()
		
		wg.Wait()
	}
}

// ============================================================
// Algorithm Patterns
// ============================================================

// RPN calculator (stack-based expression evaluation)
func BenchmarkRPNCalculatorUal(b *testing.B) {
	// Evaluate: ((3 + 4) * 5 - 2) repeatedly
	for i := 0; i < b.N; i++ {
		s := NewStack(LIFO, TypeInt64)
		for j := 0; j < 100; j++ {
			s.Push(intToBytes(3))
			s.Push(intToBytes(4))
			a, _ := s.Pop()
			bb, _ := s.Pop()
			s.Push(intToBytes(bytesToInt(a) + bytesToInt(bb)))
			
			s.Push(intToBytes(5))
			a, _ = s.Pop()
			bb, _ = s.Pop()
			s.Push(intToBytes(bytesToInt(a) * bytesToInt(bb)))
			
			s.Push(intToBytes(2))
			a, _ = s.Pop()
			bb, _ = s.Pop()
			s.Push(intToBytes(bytesToInt(bb) - bytesToInt(a)))
			
			s.Pop()
		}
	}
}

func BenchmarkRPNCalculatorNative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stack := make([]int64, 0, 10)
		for j := 0; j < 100; j++ {
			stack = append(stack, 3, 4)
			a, bb := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a+bb)
			
			stack = append(stack, 5)
			a, bb = stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a*bb)
			
			stack = append(stack, 2)
			a, bb = stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, bb-a)
			
			stack = stack[:len(stack)-1]
		}
	}
}

// DFS vs BFS: Same traversal, different perspective
func BenchmarkDFSTraversal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		frontier := NewStack(LIFO, TypeInt64) // LIFO = depth-first
		visited := make(map[int64]bool)
		
		frontier.Push(intToBytes(0))
		
		for frontier.Len() > 0 {
			nodeBytes, _ := frontier.Pop()
			node := bytesToInt(nodeBytes)
			
			if visited[node] {
				continue
			}
			visited[node] = true
			
			if node < 512 {
				frontier.Push(intToBytes(node*2 + 1))
				frontier.Push(intToBytes(node*2 + 2))
			}
		}
	}
}

func BenchmarkBFSTraversal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		frontier := NewStack(FIFO, TypeInt64) // FIFO = breadth-first
		visited := make(map[int64]bool)
		
		frontier.Push(intToBytes(0))
		
		for frontier.Len() > 0 {
			nodeBytes, _ := frontier.Pop()
			node := bytesToInt(nodeBytes)
			
			if visited[node] {
				continue
			}
			visited[node] = true
			
			if node < 512 {
				frontier.Push(intToBytes(node*2 + 1))
				frontier.Push(intToBytes(node*2 + 2))
			}
		}
	}
}

// Partition (quicksort building block)
func BenchmarkPartitionUal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		input := NewStack(LIFO, TypeInt64)
		less := NewStack(LIFO, TypeInt64)
		greater := NewStack(LIFO, TypeInt64)
		
		for j := 0; j < 1000; j++ {
			input.Push(intToBytes(int64((j * 17) % 1000)))
		}
		
		pivot := int64(500)
		for input.Len() > 0 {
			val, _ := input.Pop()
			if bytesToInt(val) < pivot {
				less.Push(val)
			} else {
				greater.Push(val)
			}
		}
		_ = less.Len() + greater.Len()
	}
}

func BenchmarkPartitionNative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		input := make([]int64, 1000)
		for j := 0; j < 1000; j++ {
			input[j] = int64((j * 17) % 1000)
		}
		
		less := make([]int64, 0, 500)
		greater := make([]int64, 0, 500)
		pivot := int64(500)
		
		for _, v := range input {
			if v < pivot {
				less = append(less, v)
			} else {
				greater = append(greater, v)
			}
		}
		_ = len(less) + len(greater)
	}
}

// ============================================================
// Memory & Allocation Patterns
// ============================================================

func BenchmarkAlloc10K(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := NewStack(LIFO, TypeInt64)
		for j := 0; j < 10000; j++ {
			s.Push(intToBytes(int64(j)))
		}
		for j := 0; j < 10000; j++ {
			s.Pop()
		}
	}
}

func BenchmarkAllocPrealloc10K(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := NewCappedStack(LIFO, TypeInt64, 10000)
		for j := 0; j < 10000; j++ {
			s.Push(intToBytes(int64(j)))
		}
		for j := 0; j < 10000; j++ {
			s.Pop()
		}
	}
}

// ============================================================
// Perspective Overhead
// ============================================================

func BenchmarkPerspectiveSwitchWithPeek(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetPerspective(FIFO)
		s.Peek()
		s.SetPerspective(LIFO)
		s.Peek()
	}
}

func BenchmarkDualViewNoPerspectiveSwitch(b *testing.B) {
	s := NewStack(LIFO, TypeInt64)
	lifo := NewView(LIFO)
	fifo := NewView(FIFO)
	lifo.Attach(s)
	fifo.Attach(s)
	
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fifo.Peek()
		lifo.Peek()
	}
}

// ============================================================
// Baseline Reference: Atomic Operations
// ============================================================

func BenchmarkAtomicIncrement(b *testing.B) {
	var counter int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&counter, 1)
		}
	})
}

func BenchmarkMutexIncrement(b *testing.B) {
	var counter int64
	var mu sync.Mutex
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			counter++
			mu.Unlock()
		}
	})
}
