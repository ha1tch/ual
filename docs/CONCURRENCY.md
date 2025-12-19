# ual Concurrency

This document provides a comprehensive technical reference for concurrency in ual. It covers the spawn mechanism, stack semantics, synchronisation primitives, and common patterns.

## Overview

ual's concurrency model is built on two key principles:

1. **Explicit coordination through stacks** — Communication between concurrent tasks happens through shared, thread-safe stacks. There are no hidden channels or implicit synchronisation points.

2. **Isolation of operational state** — Each spawned task gets its own operational stacks (`@dstack`, `@rstack`, `@bool`, `@error`), preventing race conditions in Forth-style stack manipulation while allowing deliberate sharing of user-defined stacks.

This design makes data flow visible and predictable. When you see `@results take:val` in a spawned task, you know exactly what's being shared and where synchronisation occurs.

## The Spawn Mechanism

### Task Queuing

Tasks are queued onto the `@spawn` stack as closures:

```ual
@spawn < {
    -- code to run concurrently
    @results push(100)
}
```

The `@spawn <` syntax pushes a closure (the code block) onto the spawn stack. The task does not execute immediately — it waits until explicitly started.

### Task Execution

Tasks are started with `@spawn pop play`:

```ual
@spawn pop play    -- pop task from spawn stack, run in new goroutine
```

This pops the most recently queued task and executes it in a new goroutine (Go) or thread (Rust). Multiple tasks can be queued and started:

```ual
@spawn < { @results < 100 }
@spawn < { @results < 200 }
@spawn < { @results < 300 }

@spawn pop play    -- starts task pushing 300
@spawn pop play    -- starts task pushing 200
@spawn pop play    -- starts task pushing 100
```

Note the LIFO order: the last task queued is the first to be popped and started.

### Spawn Operations

| Operation | Effect |
|-----------|--------|
| `@spawn < { ... }` | Queue a task closure onto the spawn stack |
| `@spawn pop play` | Pop and run task in new goroutine |
| `@spawn peek play` | Run top task without removing (limited support) |
| `@spawn pop` | Remove task without running |
| `@spawn len` | Push task count to `@dstack` |
| `@spawn clear` | Remove all queued tasks |

See `examples/016_spawn.ual` for basic spawning and `examples/017_spawn_chain.ual` for chained task execution.

## Stack Semantics in Concurrent Contexts

### Operational Stacks: Per-Goroutine

When a task executes via `@spawn pop play`, it receives **private copies** of the operational stacks:

| Stack | Scope | Purpose |
|-------|-------|---------|
| `@dstack` | Per-goroutine | Data stack for Forth-style operations |
| `@rstack` | Per-goroutine | Return stack for temporary storage |
| `@bool` | Per-goroutine | Boolean/comparison results |
| `@error` | Per-goroutine | Error state |

This isolation is critical for correctness. Consider what would happen if `@dstack` were shared:

```ual
-- HYPOTHETICAL: if @dstack were shared (IT IS NOT)
@spawn < {
    push:1         -- dstack = [1]
    inc            -- pops 1, pushes 2
    pop:result     -- result = 2
}
@spawn < {
    push:10        -- dstack = [10] or [1, 10]?
    inc            -- race condition!
    pop:result     -- unpredictable
}
```

With shared `@dstack`, concurrent `push`/`pop`/`inc` operations would interleave unpredictably. By giving each goroutine its own operational stacks, ual ensures that Forth-style stack code behaves identically whether run sequentially or concurrently.

### User-Defined Stacks: Shared

User-defined stacks are **shared** between all goroutines and are thread-safe:

```ual
@results = stack.new(i64)    -- shared, thread-safe

@spawn < {
    @results push(100)       -- safe concurrent access
}
@spawn < {
    @results push(200)       -- safe concurrent access
}
```

The runtime ensures that `push`, `pop`, and `take` on user-defined stacks are atomic. This is how tasks communicate — by pushing to and taking from shared stacks.

### Summary Table

| Stack Type | Sharing | Thread Safety | Use Case |
|------------|---------|---------------|----------|
| `@dstack`, `@rstack` | Per-goroutine | N/A (not shared) | Local computation |
| `@bool`, `@error` | Per-goroutine | N/A (not shared) | Local state |
| User-defined (`@foo`) | Shared | Yes (mutex-protected) | Inter-task communication |

## Synchronisation Primitives

### Blocking Take

The `take` operation blocks until data is available:

```ual
@results take:val    -- blocks until @results has data, pops into val
```

This is the primary synchronisation mechanism. A task waiting on `take` yields its goroutine until another task pushes data.

```ual
@signal = stack.new(i64)

@spawn < {
    -- do work...
    @signal push(1)    -- signal completion
}

@spawn pop play
@signal take:done      -- blocks until signal arrives
```

See `examples/018_take_sync.ual` for a complete example.

### Take with Timeout

`take` accepts an optional timeout in milliseconds:

```ual
@results take(1000):val    -- wait up to 1 second
```

If the timeout expires before data arrives, `take` returns a zero/default value. Check stack state or use error handling for timeout detection.

See `examples/019_take_timeout.ual` and `examples/021_take_timeout_var.ual`.

### Select: Multi-Stack Waiting

The `.select()` construct waits on multiple stacks simultaneously:

```ual
@dstack {
    -- setup code (optional)
}.select(
    @inbox {|msg|
        process_message(msg)
    }
    @commands {|cmd|
        execute_command(cmd)
    }
    _: {
        -- default case (non-blocking)
        handle_idle()
    }
)
```

Select semantics:

- **Without default (`_:`)**: Blocks until one of the named stacks has data
- **With default**: Non-blocking; runs default case if no data available
- **Binding**: The `|msg|` syntax binds the popped value to a variable
- **Single match**: Only one case executes per select

See `examples/031_select_basic.ual`, `examples/032_select_blocking.ual`, and `examples/033_select_timeout.ual`.

### Timeout in Select

Select can include a timeout case:

```ual
@dstack {}.select(
    @data {|val| process(val) }
    timeout(5000) {
        handle_timeout()
    }
)
```

## Common Concurrency Patterns

### Producer-Consumer

One or more producers push data; one or more consumers take it.

```ual
@work = stack.new(i64)
@done = stack.new(i64)

-- Producer
@spawn < {
    var i i64 = 0
    while (i < 10) {
        @work push(i * 10)
        push:i inc let:i
    }
    @done push(1)
}

-- Consumer
@spawn < {
    var count i64 = 0
    while (count < 10) {
        var item i64 = 0
        @work take:item
        -- process item
        push:count inc let:count
    }
    @done push(1)
}

@spawn pop play
@spawn pop play

-- Wait for both
@done take:x
@done take:y
```

See `examples/072_multi_producer.ual` for multiple producers.

### Bounded Buffer

Semaphore-style coordination limits buffer size:

```ual
@buffer = stack.new(i64)
@slots = stack.new(i64)     -- available space
@items = stack.new(i64)     -- available items

-- Initialize with N slots
@slots push(1)
@slots push(1)
@slots push(1)

-- Producer: wait for slot, push item, signal item
@spawn < {
    var slot i64 = 0
    @slots take:slot        -- blocks if buffer full
    @buffer push(value)
    @items push(1)          -- signal item available
}

-- Consumer: wait for item, take from buffer, signal slot
@spawn < {
    var signal i64 = 0
    @items take:signal      -- blocks if buffer empty
    @buffer take:value
    @slots push(1)          -- signal slot available
}
```

This pattern ensures the producer blocks when the buffer is full and the consumer blocks when empty. See `examples/079_bounded_buffer.ual`.

### Request-Response (Ping-Pong)

Bidirectional communication between tasks:

```ual
@request = stack.new(i64)
@response = stack.new(i64)

-- Worker: receive request, send response
@spawn < {
    while (running) {
        var req i64 = 0
        @request take:req
        @response push(req + 1)
    }
}

-- Main: send request, await response
@request push(current)
@response take:current
```

See `examples/075_ping_pong.ual`.

### Fan-Out / Fan-In

Distribute work across multiple workers, collect results:

```ual
@work = stack.new(i64)
@results = stack.new(i64)
@done = stack.new(i64)

-- Spawn N workers
var i i64 = 0
while (i < N) {
    @spawn < {
        while (has_work) {
            var item i64 = 0
            @work take(100):item    -- timeout prevents deadlock
            if (item != 0) {
                @results push(process(item))
            }
        }
        @done push(1)
    }
    @spawn pop play
    push:i inc let:i
}

-- Fan out: push work items
-- ...

-- Fan in: collect results
-- ...

-- Wait for all workers
var finished i64 = 0
while (finished < N) {
    @done take:x
    push:finished inc let:finished
}
```

See `examples/073_fan_out_in.ual`.

### Barrier Synchronisation

Wait for multiple tasks to reach a synchronisation point:

```ual
@barrier = stack.new(i64)

-- Each worker signals when ready
@spawn < {
    -- do work...
    @barrier push(1)    -- signal ready
}

-- Wait for all N workers
var count i64 = 0
while (count < N) {
    @barrier take:x
    push:count inc let:count
}
-- All workers have reached the barrier
```

See `examples/074_barrier_sync.ual`.

### Work Queue with Multiple Workers

```ual
@queue = stack.new(i64)
@queue perspective(FIFO)    -- process in order
@results = stack.new(i64)

-- Workers compete for work items
@spawn < {
    while (true) {
        var item i64 = 0
        @queue take(100):item
        if (item == 0) { break }
        @results push(process(item))
    }
}
```

The FIFO perspective ensures work is processed in submission order. See `examples/076_work_queue.ual`.

### MapReduce Pattern

```ual
@input = stack.new(i64)
@mapped = stack.new(i64)
@reduced = stack.new(i64)

-- Mapper workers
@spawn < {
    while (has_input) {
        var val i64 = 0
        @input take:val
        @mapped push(map_fn(val))
    }
}

-- Reducer
@spawn < {
    var acc i64 = 0
    while (has_mapped) {
        var val i64 = 0
        @mapped take:val
        push:(acc + val) let:acc
    }
    @reduced push(acc)
}
```

See `examples/077_mapreduce.ual`.

## Implementation Details

### Go Backend

Spawned tasks become Go goroutines. Each closure captures its environment and creates local operational stacks:

```go
spawn_tasks = append(spawn_tasks, func() {
    stack_dstack := ual.NewStack(ual.LIFO, ual.TypeInt64)
    stack_rstack := ual.NewStack(ual.LIFO, ual.TypeInt64)
    // ... closure body
})

// Later:
go task()  // runs in new goroutine
```

User-defined stacks use mutex-protected operations with condition variables for blocking `take`.

### Rust Backend

Spawned tasks become `std::thread::spawn` threads. Local operational stacks shadow the global statics:

```rust
tasks.push(Box::new(move || {
    let _dstack: Stack<i64> = Stack::new(Perspective::LIFO);
    let _rstack: Stack<i64> = Stack::new(Perspective::LIFO);
    // ... closure body uses _dstack, _rstack
}));

// Later:
std::thread::spawn(move || { task(); });
```

User-defined stacks use `Mutex<Vec<T>>` with `Condvar` for blocking operations.

### Interpreter (iual)

The interpreter uses real Go goroutines for `@spawn pop play`, matching compiled semantics. Each spawned goroutine gets fresh operational stacks while sharing references to user-defined stacks:

```go
childStacks := make(map[string]*runtime.ValueStack)
for name, stack := range i.stacks {
    switch name {
    case "dstack", "rstack", "bool", "error":
        childStacks[name] = runtime.NewValueStack(runtime.LIFO)
    default:
        childStacks[name] = stack  // share user stacks
    }
}
```

## Best Practices

### Always Synchronise Before Exit

The main program may exit before spawned goroutines complete. Use `take` or barriers to ensure work finishes:

```ual
@done = stack.new(i64)

@spawn < {
    -- work...
    @done push(1)
}
@spawn pop play

@done take:x    -- wait for completion
```

### Use Timeouts to Prevent Deadlock

In complex coordination patterns, use timeouts to detect and handle deadlock conditions:

```ual
@data take(5000):val
if (val == 0) {
    -- timeout occurred, handle gracefully
}
```

### Prefer Explicit Stacks Over Operational Stacks

For inter-task communication, always use named user-defined stacks:

```ual
-- Good: explicit, visible communication
@results = stack.new(i64)
@spawn < { @results push(42) }

-- Problematic: relies on dstack which is per-goroutine
@spawn < { push:42 }    -- this 42 is isolated to the spawned task
```

### Keep Spawn Blocks Small

Spawn blocks should be focused. Complex logic belongs in functions:

```ual
-- Prefer this:
@spawn < {
    var result i64 = complex_computation(input)
    @results push(result)
}

-- Over this:
@spawn < {
    -- 100 lines of inline code
}
```

### Use FIFO for Ordered Processing

When order matters, set the stack perspective:

```ual
@queue = stack.new(i64)
@queue perspective(FIFO)    -- first in, first out
```

## Debugging Concurrent Code

### Tracing

Use `iual --trace` to see execution flow:

```bash
iual --trace examples/075_ping_pong.ual
```

### Stress Testing

The stress test script helps find race conditions:

```bash
./tests/stress_concurrency.sh -n 10000 -s 075,079 --go --rust --iual
```

Run thousands of iterations to surface timing-dependent bugs.

### Common Issues

| Symptom | Likely Cause | Solution |
|---------|--------------|----------|
| Deadlock (hangs) | Missing `push` to unblock `take` | Ensure all `take` calls have matching `push` |
| Wrong values | Shared operational stack (old bug) | Update to v0.7.4+; operational stacks are now per-goroutine |
| Race condition | Unsynchronised access | Use stacks for all inter-task communication |
| Early exit | Main exits before workers | Add barrier/done synchronisation |

## Related Examples

| Example | Pattern | Key Concepts |
|---------|---------|--------------|
| `016_spawn.ual` | Basic spawning | `@spawn <`, `@spawn pop play` |
| `017_spawn_chain.ual` | Task chaining | Sequential task execution |
| `018_take_sync.ual` | Blocking sync | `take` as synchronisation |
| `019_take_timeout.ual` | Timeout handling | `take(ms)` |
| `022_pipeline.ual` | Data pipeline | Producer-consumer chain |
| `032_select_blocking.ual` | Multi-wait | `.select()` without default |
| `033_select_timeout.ual` | Timed wait | `.select()` with timeout |
| `072_multi_producer.ual` | Multiple producers | Fan-in pattern |
| `073_fan_out_in.ual` | Work distribution | Fan-out/fan-in |
| `074_barrier_sync.ual` | Barrier | Wait for N tasks |
| `075_ping_pong.ual` | Request-response | Bidirectional communication |
| `076_work_queue.ual` | Work stealing | FIFO queue with workers |
| `077_mapreduce.ual` | MapReduce | Parallel map, sequential reduce |
| `078_semaphore.ual` | Semaphore | Resource counting |
| `079_bounded_buffer.ual` | Bounded buffer | Back-pressure with semaphores |
| `081_pipeline_stages.ual` | Multi-stage pipeline | Three concurrent stages |
| `082_competing_workers.ual` | Competing workers | Shared queue with sentinel termination |
| `083_load_balancer.ual` | Load balancer | Round-robin request distribution |
| `084_graceful_shutdown.ual` | Graceful shutdown | Clean termination with in-flight work |
| `085_resource_pool.ual` | Resource pool | Connection pool pattern |
| `086_local_stack_basic.ual` | Local stacks | Spawn-local stacks that don't conflict |
| `087_compute_in_spawn.ual` | Compute in spawn | Compute blocks inside spawn blocks |
| `088_local_stack_compute.ual` | Local stack + compute | Per-worker accumulation with compute reduction |
| `089_parallel_reduction.ual` | Parallel reduction | Partial sums with final combination |

## Further Reading

- `MANUAL.md` — Language reference with spawn and select syntax
- `CHANGELOG.md` — Version history including concurrency fixes in v0.7.4
- `ERROR_PHILOSOPHY.md` — Error handling in concurrent contexts
