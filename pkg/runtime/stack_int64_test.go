package runtime

import (
	"testing"
)

// ============================================================
// Generic Stack vs Int64Stack: Direct Comparison
// ============================================================

func BenchmarkGenericStack_Push(b *testing.B) {
	s := NewCappedStack(LIFO, TypeInt64, b.N+100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(intToBytes(int64(i)))
	}
}

func BenchmarkInt64Stack_Push(b *testing.B) {
	s := NewCappedInt64Stack(LIFO, b.N+100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(int64(i))
	}
}

func BenchmarkGenericStack_Pop(b *testing.B) {
	s := NewCappedStack(LIFO, TypeInt64, b.N+100)
	for i := 0; i < b.N; i++ {
		s.Push(intToBytes(int64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkInt64Stack_Pop(b *testing.B) {
	s := NewCappedInt64Stack(LIFO, b.N+100)
	for i := 0; i < b.N; i++ {
		s.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkGenericStack_Peek(b *testing.B) {
	s := NewCappedStack(LIFO, TypeInt64, 1000)
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Peek()
	}
}

func BenchmarkInt64Stack_Peek(b *testing.B) {
	s := NewCappedInt64Stack(LIFO, 1000)
	for i := 0; i < 1000; i++ {
		s.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Peek()
	}
}

func BenchmarkGenericStack_PushPop(b *testing.B) {
	s := NewCappedStack(LIFO, TypeInt64, 1000)
	for i := 0; i < 500; i++ {
		s.Push(intToBytes(int64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(intToBytes(42))
		s.Pop()
	}
}

func BenchmarkInt64Stack_PushPop(b *testing.B) {
	s := NewCappedInt64Stack(LIFO, 1000)
	for i := 0; i < 500; i++ {
		s.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(42)
		s.Pop()
	}
}

// ============================================================
// Work-Stealing Pattern Comparison
// ============================================================

func BenchmarkGenericStack_WorkSteal(b *testing.B) {
	s := NewCappedStack(LIFO, TypeInt64, 10000)
	owner := NewView(LIFO)
	thief := NewView(FIFO)
	owner.Attach(s)
	thief.Attach(s)
	
	// Pre-fill
	for i := 0; i < 1000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(intToBytes(42))
		s.Push(intToBytes(43))
		owner.Pop()
		thief.Pop()
	}
}

func BenchmarkInt64Stack_WorkSteal(b *testing.B) {
	s := NewCappedInt64Stack(LIFO, 10000)
	owner := NewInt64View(LIFO)
	thief := NewInt64View(FIFO)
	owner.Attach(s)
	thief.Attach(s)
	
	// Pre-fill
	for i := 0; i < 1000; i++ {
		s.Push(int64(i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(42)
		s.Push(43)
		owner.Pop()
		thief.Pop()
	}
}

// ============================================================
// Sum Reduction Comparison
// ============================================================

func BenchmarkGenericStack_Sum10K(b *testing.B) {
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

func BenchmarkInt64Stack_Sum10K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewInt64Stack(LIFO)
		for j := 0; j < 10000; j++ {
			s.Push(int64(j))
		}
		sum := int64(0)
		for s.Len() > 0 {
			v, _ := s.Pop()
			sum += v
		}
		_ = sum
	}
}

func BenchmarkInt64Stack_Sum10K_Preallocated(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewCappedInt64Stack(LIFO, 10000)
		for j := 0; j < 10000; j++ {
			s.Push(int64(j))
		}
		sum := int64(0)
		for s.Len() > 0 {
			v, _ := s.Pop()
			sum += v
		}
		_ = sum
	}
}

func BenchmarkNativeSlice_Sum10K(b *testing.B) {
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

// ============================================================
// Parallel Access Comparison
// ============================================================

func BenchmarkGenericStack_Parallel(b *testing.B) {
	s := NewCappedStack(LIFO, TypeInt64, 100000)
	for i := 0; i < 50000; i++ {
		s.Push(intToBytes(int64(i)))
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		v := NewView(FIFO)
		v.Attach(s)
		for pb.Next() {
			v.Pop()
		}
	})
}

func BenchmarkInt64Stack_Parallel(b *testing.B) {
	s := NewCappedInt64Stack(LIFO, 100000)
	for i := 0; i < 50000; i++ {
		s.Push(int64(i))
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		v := NewInt64View(FIFO)
		v.Attach(s)
		for pb.Next() {
			v.Pop()
		}
	})
}

// ============================================================
// Allocation Comparison
// ============================================================

func BenchmarkGenericStack_Allocs(b *testing.B) {
	b.ReportAllocs()
	s := NewCappedStack(LIFO, TypeInt64, 100)
	for i := 0; i < 50; i++ {
		s.Push(intToBytes(int64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(intToBytes(42))
		s.Pop()
	}
}

func BenchmarkInt64Stack_Allocs(b *testing.B) {
	b.ReportAllocs()
	s := NewCappedInt64Stack(LIFO, 100)
	for i := 0; i < 50; i++ {
		s.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(42)
		s.Pop()
	}
}

// ============================================================
// Lock-Free Stack Comparison
// ============================================================

func BenchmarkFastInt64Stack_Push(b *testing.B) {
	s := NewFastInt64Stack(b.N + 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(int64(i))
	}
}

func BenchmarkFastInt64Stack_Pop(b *testing.B) {
	s := NewFastInt64Stack(b.N + 100)
	for i := 0; i < b.N; i++ {
		s.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkFastInt64Stack_PushPop(b *testing.B) {
	s := NewFastInt64Stack(1000)
	for i := 0; i < 500; i++ {
		s.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(42)
		s.Pop()
	}
}

// ============================================================
// Work-Stealing Deque Comparison
// ============================================================

func BenchmarkWorkStealingDeque_Push(b *testing.B) {
	d := NewWorkStealingDeque(b.N + 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(int64(i))
	}
}

func BenchmarkWorkStealingDeque_Pop(b *testing.B) {
	d := NewWorkStealingDeque(b.N + 100)
	for i := 0; i < b.N; i++ {
		d.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Pop()
	}
}

func BenchmarkWorkStealingDeque_Steal(b *testing.B) {
	d := NewWorkStealingDeque(b.N + 100)
	for i := 0; i < b.N; i++ {
		d.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Steal()
	}
}

func BenchmarkWorkStealingDeque_OwnerThief(b *testing.B) {
	d := NewWorkStealingDeque(10000)
	for i := 0; i < 1000; i++ {
		d.Push(int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(42)
		d.Push(43)
		d.Pop()   // owner
		d.Steal() // thief
	}
}

func BenchmarkWorkStealingDeque_Parallel(b *testing.B) {
	d := NewWorkStealingDeque(100000)
	for i := 0; i < 50000; i++ {
		d.Push(int64(i))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d.Steal()
		}
	})
}

// ============================================================
// Full Comparison: All Implementations
// ============================================================

func BenchmarkAllImpls_PushPop(b *testing.B) {
	b.Run("Generic", func(b *testing.B) {
		s := NewCappedStack(LIFO, TypeInt64, 1000)
		for i := 0; i < 500; i++ {
			s.Push(intToBytes(int64(i)))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Push(intToBytes(42))
			s.Pop()
		}
	})
	
	b.Run("Int64Stack", func(b *testing.B) {
		s := NewCappedInt64Stack(LIFO, 1000)
		for i := 0; i < 500; i++ {
			s.Push(int64(i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Push(42)
			s.Pop()
		}
	})
	
	b.Run("FastInt64", func(b *testing.B) {
		s := NewFastInt64Stack(1000)
		for i := 0; i < 500; i++ {
			s.Push(int64(i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Push(42)
			s.Pop()
		}
	})
	
	b.Run("Deque", func(b *testing.B) {
		d := NewWorkStealingDeque(1000)
		for i := 0; i < 500; i++ {
			d.Push(int64(i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d.Push(42)
			d.Pop()
		}
	})
}

func BenchmarkAllImpls_WorkSteal(b *testing.B) {
	b.Run("Generic", func(b *testing.B) {
		s := NewCappedStack(LIFO, TypeInt64, 10000)
		owner := NewView(LIFO)
		thief := NewView(FIFO)
		owner.Attach(s)
		thief.Attach(s)
		for i := 0; i < 1000; i++ {
			s.Push(intToBytes(int64(i)))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Push(intToBytes(42))
			s.Push(intToBytes(43))
			owner.Pop()
			thief.Pop()
		}
	})
	
	b.Run("Int64Stack", func(b *testing.B) {
		s := NewCappedInt64Stack(LIFO, 10000)
		owner := NewInt64View(LIFO)
		thief := NewInt64View(FIFO)
		owner.Attach(s)
		thief.Attach(s)
		for i := 0; i < 1000; i++ {
			s.Push(int64(i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Push(42)
			s.Push(43)
			owner.Pop()
			thief.Pop()
		}
	})
	
	b.Run("Deque", func(b *testing.B) {
		d := NewWorkStealingDeque(10000)
		for i := 0; i < 1000; i++ {
			d.Push(int64(i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d.Push(42)
			d.Push(43)
			d.Pop()
			d.Steal()
		}
	})
}

// Sustained throughput: producer keeps queue filled
func BenchmarkWorkStealingDeque_SustainedSteal(b *testing.B) {
	d := NewWorkStealingDeque(10000)
	
	// Producer goroutine keeps it filled
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				if d.Len() < 5000 {
					for i := 0; i < 1000; i++ {
						d.Push(42)
					}
				}
			}
		}
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Steal()
	}
	b.StopTimer()
	close(done)
}

// Compare: channel receive with producer
func BenchmarkChannel_SustainedReceive(b *testing.B) {
	ch := make(chan int64, 10000)
	
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				select {
				case ch <- 42:
				default:
				}
			}
		}
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-ch
	}
	b.StopTimer()
	close(done)
}
