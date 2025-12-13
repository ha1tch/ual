# ual Benchmark Results

Platform: Intel Xeon @ 2.60GHz, Linux, Go 1.22.2

## Executive Summary

ual's stack abstraction adds overhead for simple operations but provides competitive performance for concurrent access patterns where its perspective model shines.

## Core Operations

| Operation | Time | Notes |
|-----------|------|-------|
| Push LIFO | 374 ns | Includes element allocation |
| Push FIFO | 295 ns | Slightly faster (append) |
| Push Hash | 780 ns | Key hashing overhead |
| Pop LIFO | 34 ns | No allocation |
| Pop FIFO | 34 ns | Same performance |
| Pop Hash | 253 ns | Lookup overhead |
| Peek LIFO | 20 ns | Read-only, fast |
| Peek Hash | 44 ns | Map lookup |
| Indexed Lookup | 19 ns | Array access |

## Work-Stealing Pattern

The main use case - work-stealing with decoupled perspectives:

| Operation | Time | Allocations |
|-----------|------|-------------|
| Owner Pop (LIFO) | 101 ns | 0 |
| Thief Steal (FIFO) | 128 ns | 0 |
| Contention (alternating) | 231 ns | 0 |
| 4 Concurrent Thieves | 137 ns | 0 |

**Comparison to baseline:**
- Atomic increment: 16 ns
- Mutex increment: 52 ns
- ual dual-view peek: 73 ns ← *competitive with mutex!*

## ual vs Native Go

Direct comparisons for common patterns:

### Sum 10,000 Integers
| Implementation | Time | Ratio |
|----------------|------|-------|
| ual Stack | 1.63 ms | 55x slower |
| Native Slice | 29 µs | baseline |

*Expected: ual pays for byte conversion on every element.*

### Fibonacci(30)
| Implementation | Time | Ratio |
|----------------|------|-------|
| ual Stack | 5.7 µs | 560x slower |
| Native Variables | 10 ns | baseline |

*Stack operations vs register variables - not a fair fight.*

### RPN Calculator (100 expressions)
| Implementation | Time | Ratio |
|----------------|------|-------|
| ual Stack | 58 µs | 336x slower |
| Native Slice | 174 ns | baseline |

*Same story: byte conversion dominates.*

### Partition (1000 elements)
| Implementation | Time | Ratio |
|----------------|------|-------|
| ual Stack | 208 µs | 87x slower |
| Native Slice | 2.4 µs | baseline |

## Traversal Patterns

Same algorithm, different perspectives:

| Traversal | Time | Notes |
|-----------|------|-------|
| DFS (LIFO) | 169 µs | 1023 nodes |
| BFS (FIFO) | 225 µs | Same nodes, +33% time |

*FIFO is slower due to index management for queue semantics.*

## Memory Allocation

| Pattern | Time | Allocs |
|---------|------|--------|
| 10K push/pop (growing) | 1.75 ms | 10,037 |
| 10K push/pop (preallocated) | 1.01 ms | 10,003 |

Preallocation saves ~42% time.

## Concurrent Patterns

### Producer-Consumer
| Implementation | Time | Ratio |
|----------------|------|-------|
| ual Stack | 176 µs | 27x slower |
| Go Channels | 6.6 µs | baseline |

*Different abstractions - channels have runtime optimisation.*

### Traditional vs ual Work-Stealing
| Operation | Traditional | ual | Capped ual |
|-----------|-------------|-----|------------|
| Push | 25 ns | 330 ns | 60 ns |
| Pop | 12 ns | 72 ns | 87 ns |
| Steal | 24 ns | 75 ns | 80 ns |
| Push+Pop | 44 ns | 130 ns | 109 ns |
| Concurrent | 150 ns | 168 ns | 158 ns |
| 1 Owner + Thieves | 462 ns | 525 ns | 341 ns |

**Key insight:** Concurrent access is *only 1.1x slower* than traditional. The abstraction cost disappears under contention.

## Perspective Switching

| Approach | Time |
|----------|------|
| SetPerspective (3 switches) | 10.2 µs |
| Dual Views (no switch) | 73 ns |

**Recommendation:** Create views for different access patterns rather than switching stack perspective.

## Key Takeaways

1. **Don't use ual for simple arithmetic.** The byte conversion overhead kills performance (50-500x slower).

2. **ual shines for concurrent access.** Only 1.1x overhead vs hand-rolled work-stealing under contention.

3. **Perspectives are cheap.** Dual-view access (73ns) is competitive with mutex operations (52ns).

4. **Preallocate when possible.** `NewCappedStack` saves 42% on allocation-heavy workloads.

5. **LIFO > FIFO > Hash** for raw speed. FIFO pays for queue semantics; Hash pays for key management.

6. **The abstraction pays for itself** when you need multiple access patterns over shared data. One stack, multiple views, no synchronisation code.

## When to Use ual

✓ Work-stealing schedulers  
✓ Multi-consumer queues with different priorities  
✓ Concurrent data structures needing LIFO/FIFO/random access  
✓ Stack-based interpreters or VMs  

✗ Hot loops with simple arithmetic  
✗ Single-threaded batch processing  
✗ Memory-constrained environments (byte[] allocation)
