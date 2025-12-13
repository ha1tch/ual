package runtime

import (
	"sync"
	"testing"
)

// ============================================================================
// Correctness Tests
// ============================================================================

func TestTraditionalWorkStealing(t *testing.T) {
	d := NewWSDeque(100)
	
	// Owner pushes
	d.Push(Task{ID: 1})
	d.Push(Task{ID: 2})
	d.Push(Task{ID: 3})
	d.Push(Task{ID: 4})
	d.Push(Task{ID: 5})
	
	// Owner pops (LIFO - gets 5)
	task, ok := d.Pop()
	if !ok || task.ID != 5 {
		t.Errorf("owner expected 5, got %d", task.ID)
	}
	
	// Thief steals (FIFO - gets 1)
	task, ok = d.Steal()
	if !ok || task.ID != 1 {
		t.Errorf("thief expected 1, got %d", task.ID)
	}
	
	// Owner pops (gets 4)
	task, ok = d.Pop()
	if !ok || task.ID != 4 {
		t.Errorf("owner expected 4, got %d", task.ID)
	}
	
	// Thief steals (gets 2)
	task, ok = d.Steal()
	if !ok || task.ID != 2 {
		t.Errorf("thief expected 2, got %d", task.ID)
	}
	
	// Only 3 remains
	if d.Len() != 1 {
		t.Errorf("expected 1 remaining, got %d", d.Len())
	}
}

func TestUalWorkStealing(t *testing.T) {
	ws := NewWSStack()
	
	// Owner pushes
	ws.Push(Task{ID: 1})
	ws.Push(Task{ID: 2})
	ws.Push(Task{ID: 3})
	ws.Push(Task{ID: 4})
	ws.Push(Task{ID: 5})
	
	// Owner pops (LIFO - gets 5)
	task, ok := ws.Pop()
	if !ok || task.ID != 5 {
		t.Errorf("owner expected 5, got %d", task.ID)
	}
	
	// Thief steals (FIFO - gets 1)
	task, ok = ws.Steal()
	if !ok || task.ID != 1 {
		t.Errorf("thief expected 1, got %d", task.ID)
	}
	
	// Owner pops (gets 4)
	task, ok = ws.Pop()
	if !ok || task.ID != 4 {
		t.Errorf("owner expected 4, got %d", task.ID)
	}
	
	// Thief steals (gets 2)
	task, ok = ws.Steal()
	if !ok || task.ID != 2 {
		t.Errorf("thief expected 2, got %d", task.ID)
	}
	
	// Only 3 remains
	if ws.Len() != 1 {
		t.Errorf("expected 1 remaining, got %d", ws.Len())
	}
}

func TestCappedWorkStealing(t *testing.T) {
	ws := NewWSStackCapped(10)
	
	// Owner pushes
	ws.Push(Task{ID: 1})
	ws.Push(Task{ID: 2})
	ws.Push(Task{ID: 3})
	ws.Push(Task{ID: 4})
	ws.Push(Task{ID: 5})
	
	// Owner pops (LIFO - gets 5)
	task, ok := ws.Pop()
	if !ok || task.ID != 5 {
		t.Errorf("owner expected 5, got %d", task.ID)
	}
	
	// Thief steals (FIFO - gets 1)
	task, ok = ws.Steal()
	if !ok || task.ID != 1 {
		t.Errorf("thief expected 1, got %d", task.ID)
	}
	
	// Owner pops (gets 4)
	task, ok = ws.Pop()
	if !ok || task.ID != 4 {
		t.Errorf("owner expected 4, got %d", task.ID)
	}
	
	// Thief steals (gets 2)
	task, ok = ws.Steal()
	if !ok || task.ID != 2 {
		t.Errorf("thief expected 2, got %d", task.ID)
	}
	
	// Only 3 remains
	if ws.Len() != 1 {
		t.Errorf("expected 1 remaining, got %d", ws.Len())
	}
}

func TestCappedStackFull(t *testing.T) {
	ws := NewWSStackCapped(3)
	
	ws.Push(Task{ID: 1})
	ws.Push(Task{ID: 2})
	ws.Push(Task{ID: 3})
	
	// Should fail - stack is full
	ok := ws.Push(Task{ID: 4})
	if ok {
		t.Error("expected push to fail when stack is full")
	}
	
	// Pop one
	ws.Pop()
	
	// Now should succeed
	ok = ws.Push(Task{ID: 4})
	if !ok {
		t.Error("expected push to succeed after pop")
	}
}

// ============================================================================
// Single-threaded Benchmarks (baseline)
// ============================================================================

func BenchmarkTraditional_Push(b *testing.B) {
	d := NewWSDeque(b.N + 100)
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(task)
	}
}

func BenchmarkUal_Push(b *testing.B) {
	ws := NewWSStack()
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
}

func BenchmarkTraditional_Pop(b *testing.B) {
	d := NewWSDeque(b.N + 100)
	task := Task{ID: 42, Data: []byte("test")}
	for i := 0; i < b.N; i++ {
		d.Push(task)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Pop()
	}
}

func BenchmarkUal_Pop(b *testing.B) {
	ws := NewWSStack()
	task := Task{ID: 42, Data: []byte("test")}
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Pop()
	}
}

func BenchmarkTraditional_Steal(b *testing.B) {
	d := NewWSDeque(b.N + 100)
	task := Task{ID: 42, Data: []byte("test")}
	for i := 0; i < b.N; i++ {
		d.Push(task)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Steal()
	}
}

func BenchmarkUal_Steal(b *testing.B) {
	ws := NewWSStack()
	task := Task{ID: 42, Data: []byte("test")}
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Steal()
	}
}

func BenchmarkTraditional_PushPop(b *testing.B) {
	d := NewWSDeque(1000)
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(task)
		d.Pop()
	}
}

func BenchmarkUal_PushPop(b *testing.B) {
	ws := NewWSStack()
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
		ws.Pop()
	}
}

// ============================================================================
// Multi-threaded Benchmarks (realistic work-stealing)
// ============================================================================

func BenchmarkTraditional_Concurrent(b *testing.B) {
	d := NewWSDeque(10000)
	task := Task{ID: 42, Data: []byte("test")}
	
	// Pre-fill
	for i := 0; i < 1000; i++ {
		d.Push(task)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		isOwner := true // alternate
		for pb.Next() {
			if isOwner {
				d.Push(task)
				d.Pop()
			} else {
				d.Steal()
			}
			isOwner = !isOwner
		}
	})
}

func BenchmarkUal_Concurrent(b *testing.B) {
	ws := NewWSStack()
	task := Task{ID: 42, Data: []byte("test")}
	
	// Pre-fill
	for i := 0; i < 1000; i++ {
		ws.Push(task)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		isOwner := true
		for pb.Next() {
			if isOwner {
				ws.Push(task)
				ws.Pop()
			} else {
				ws.Steal()
			}
			isOwner = !isOwner
		}
	})
}

// Simulates work-stealing scheduler: 1 owner + N thieves
func BenchmarkTraditional_OneOwnerManyThieves(b *testing.B) {
	d := NewWSDeque(100000)
	task := Task{ID: 42, Data: []byte("test")}
	
	// Pre-fill
	for i := 0; i < 10000; i++ {
		d.Push(task)
	}
	
	var wg sync.WaitGroup
	stop := make(chan struct{})
	
	// Owner goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				d.Push(task)
				d.Pop()
			}
		}
	}()
	
	// Thief goroutines
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					d.Steal()
				}
			}
		}()
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(task)
	}
	
	close(stop)
	wg.Wait()
}

func BenchmarkUal_OneOwnerManyThieves(b *testing.B) {
	ws := NewWSStack()
	task := Task{ID: 42, Data: []byte("test")}
	
	// Pre-fill
	for i := 0; i < 10000; i++ {
		ws.Push(task)
	}
	
	var wg sync.WaitGroup
	stop := make(chan struct{})
	
	// Owner goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				ws.Push(task)
				ws.Pop()
			}
		}
	}()
	
	// Thief goroutines
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					ws.Steal()
				}
			}
		}()
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
	
	close(stop)
	wg.Wait()
}

// ============================================================================
// Memory allocation comparison
// ============================================================================

func BenchmarkTraditional_Allocs(b *testing.B) {
	d := NewWSDeque(10000)
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(task)
		d.Pop()
		d.Push(task)
		d.Steal()
	}
}

func BenchmarkUal_Allocs(b *testing.B) {
	ws := NewWSStack()
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
		ws.Pop()
		ws.Push(task)
		ws.Steal()
	}
}

// ============================================================================
// Capped ual benchmarks (fixed capacity, no slice growth allocations)
// ============================================================================

func BenchmarkCapped_Push(b *testing.B) {
	ws := NewWSStackCapped(b.N + 100)
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
}

func BenchmarkCapped_Pop(b *testing.B) {
	ws := NewWSStackCapped(b.N + 100)
	task := Task{ID: 42, Data: []byte("test")}
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Pop()
	}
}

func BenchmarkCapped_Steal(b *testing.B) {
	ws := NewWSStackCapped(b.N + 100)
	task := Task{ID: 42, Data: []byte("test")}
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Steal()
	}
}

func BenchmarkCapped_PushPop(b *testing.B) {
	ws := NewWSStackCapped(1000)
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
		ws.Pop()
	}
}

func BenchmarkCapped_Concurrent(b *testing.B) {
	ws := NewWSStackCapped(100000)
	task := Task{ID: 42, Data: []byte("test")}
	
	// Pre-fill
	for i := 0; i < 1000; i++ {
		ws.Push(task)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		isOwner := true
		for pb.Next() {
			if isOwner {
				ws.Push(task)
				ws.Pop()
			} else {
				ws.Steal()
			}
			isOwner = !isOwner
		}
	})
}

func BenchmarkCapped_OneOwnerManyThieves(b *testing.B) {
	ws := NewWSStackCapped(100000)
	task := Task{ID: 42, Data: []byte("test")}
	
	// Pre-fill
	for i := 0; i < 10000; i++ {
		ws.Push(task)
	}
	
	var wg sync.WaitGroup
	stop := make(chan struct{})
	
	// Owner goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				ws.Push(task)
				ws.Pop()
			}
		}
	}()
	
	// Thief goroutines
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					ws.Steal()
				}
			}
		}()
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
	}
	
	close(stop)
	wg.Wait()
}

func BenchmarkCapped_Allocs(b *testing.B) {
	ws := NewWSStackCapped(10000)
	task := Task{ID: 42, Data: []byte("test")}
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Push(task)
		ws.Pop()
		ws.Push(task)
		ws.Steal()
	}
}
