# ual Design Decisions — Implementation Specification v0.8

**Document Purpose:** This specification captures design decisions made for ual v0.8, intended for implementors continuing development. It assumes familiarity with ual v0.7.4 architecture (lexer, parser, codegen in `cmd/ual/`, interpreter in `cmd/iual/`, runtime in `pkg/runtime/`, Rust runtime in `rual/`).

**Context:** These decisions emerged from analysis of ual's concurrency model and its relationship to distributed systems patterns. The core insight is that ual's stack-based primitives generalise naturally to I/O and network operations, and that the existing constructs (`select`, `consider`, `compute`) form a coherent system for handling time, outcomes, and interpretation.

---

## 1. Files and Sockets as Stack Sources/Sinks

### 1.1 Design Rationale

A socket is a stack you didn't fill yourself. The blocking `take` operation already exists. FIFO perspective already exists. Files and sockets are new sources and sinks that plug into existing machinery.

This unification means:
- Pipeline code works unchanged whether data comes from another goroutine or a network socket
- `select` across sockets uses the same construct as `select` across local stacks
- Error handling via `@error` stack applies uniformly

### 1.2 Streaming File Access (FIFO Pattern)

**Syntax:**
```ual
@lines = stack.new(string)
file.lines("path/to/file.txt", @lines)
```

**Semantics:**
- `file.lines(path, dest)` spawns a background reader
- Reader pushes lines to `dest` stack as they're read
- EOF signalled by closing the stack (not a sentinel value)
- Errors push to `@error` stack

**Implementation Notes:**

In `codegen.go`, generate:
```go
func _file_lines(path string, dest *ual.Stack) {
    go func() {
        f, err := os.Open(path)
        if err != nil {
            stack_error.Push([]byte(err.Error()))
            dest.Close()
            return
        }
        defer f.Close()
        scanner := bufio.NewScanner(f)
        for scanner.Scan() {
            dest.Push([]byte(scanner.Text()))
        }
        if err := scanner.Err(); err != nil {
            stack_error.Push([]byte(err.Error()))
        }
        dest.Close()
    }()
}
```

Consumer pattern:
```ual
var line string = ""
@lines take:line
while (!@lines.closed) {
    process(line)
    @lines take:line
}
```

**New Runtime Support Required (`stack.go`):**
- `Stack.Close()` already exists
- `Stack.IsClosed()` already exists
- `Take` already returns error on closed empty stack
- Need: way to distinguish "closed and empty" from "error" in take

### 1.3 Random Access File (mmap Pattern)

**Syntax:**
```ual
@data = file.mmap("matrix.bin", f64)
@data = file.mmap("output.bin", i64, "rw")  -- writable
```

**Semantics:**
- Returns stack with Indexed perspective
- Backed by memory-mapped file
- Read-only by default, "rw" flag for writable
- Compute blocks get zero-copy access via existing `unsafe.Slice` machinery

**Implementation Notes:**

New runtime type or Stack variant needed:
```go
type MmapStack struct {
    Stack
    file   *os.File
    mmap   []byte
    stride int  // bytes per element (8 for i64/f64)
}

func MmapFile(path string, elemType ElementType, writable bool) (*MmapStack, error) {
    // os.Open or os.OpenFile based on writable
    // syscall.Mmap
    // wrap in Stack interface with Indexed perspective
}
```

Compute block access uses existing `self[i]` codegen — the `unsafe.Slice` view over raw bytes works unchanged.

### 1.4 Socket Access

**Syntax:**
```ual
@conn = socket.open("tcp", "localhost:8080")
@conn perspective(FIFO)

-- or for server
@listener = socket.listen("tcp", ":8080")
@client = @listener.accept()
```

**Semantics:**
- `socket.open` returns a stack representing the connection
- Default perspective is FIFO (stream semantics)
- `push` to stack sends data
- `take` from stack receives data (blocking)
- Connection errors push to `@error`
- `close` closes the connection

**Implementation Notes:**

New file `socket.go` in runtime:
```go
type SocketStack struct {
    Stack
    conn net.Conn
}

func SocketOpen(network, address string) (*SocketStack, error) {
    conn, err := net.Dial(network, address)
    if err != nil {
        return nil, err
    }
    s := &SocketStack{
        Stack: *NewStack(FIFO, TypeBytes),
        conn:  conn,
    }
    // spawn reader goroutine that pushes to Stack
    go s.readLoop()
    return s, nil
}

func (s *SocketStack) readLoop() {
    buf := make([]byte, 4096)
    for {
        n, err := s.conn.Read(buf)
        if err != nil {
            if err != io.EOF {
                stack_error.Push([]byte(err.Error()))
            }
            s.Close()
            return
        }
        s.Push(buf[:n])
    }
}

// Override Push to write to socket
func (s *SocketStack) Push(data []byte, key ...[]byte) error {
    _, err := s.conn.Write(data)
    return err
}
```

**Lexer Changes (`lexer.go`):**
- Add keywords: `file`, `socket`, `listen`, `accept`, `mmap`
- Or treat as identifiers resolved in codegen (simpler)

**Parser Changes (`parser.go`):**
- `file.lines(...)` parses as method call on identifier `file`
- `socket.open(...)` same pattern
- May need new AST node or reuse existing FunctionCall with namespace

**Codegen Changes (`codegen.go`):**
- Recognise `file.lines`, `file.mmap`, `socket.open`, etc.
- Generate appropriate runtime calls
- Handle error pushing to `@error`

---

## 2. Error Stack Architecture

### 2.1 Design Rationale

Errors are not values. They are facts about what happened — evidence of failed state transitions. They should not pollute data flow.

Go's error handling treats errors as return values, forcing interleaved error checking:
```go
a, err := step1()
if err != nil { return err }
b, err := step2(a)
if err != nil { return err }
```

ual separates the error channel from the data channel:
```ual
step1(@a)
step2(@a, @b)
step3(@b, @c)

@error {}.consider(
    ok: { use(@c) }
    error |e|: { handle(e) }
)
```

### 2.2 @error Stack Specification

**Global Error Stack:**
- Pre-declared in every ual program (like `@dstack`, `@rstack`)
- Perspective: LIFO (most recent error on top)
- Element type: bytes (error messages as strings)

**Error Accumulation:**
- Errors push, not replace
- Multiple failures accumulate in order
- Full error trace available if needed

**Codegen (`codegen.go`):**

Already exists:
```go
var stack_error = ual.NewStack(ual.LIFO, ual.TypeBytes)
```

No changes needed for declaration.

### 2.3 Forcing Function: Unhandled Errors Cause Panic

**Semantics:**

Any operation that interacts with stacks (push, pop, take, etc.) must first check if `@error` is non-empty. If unhandled errors exist, panic immediately.

**Implementation:**

Option A — Runtime Check:

Add to `Stack` methods:
```go
var globalErrorStack *Stack  // set during init

func checkUnhandledErrors() {
    if globalErrorStack != nil && globalErrorStack.Len() > 0 {
        msg, _ := globalErrorStack.Peek()
        panic(fmt.Sprintf("unhandled error: %s", string(msg)))
    }
}

func (s *Stack) Push(value []byte, key ...[]byte) error {
    checkUnhandledErrors()
    // ... existing implementation
}

func (s *Stack) Pop(param ...[]byte) ([]byte, error) {
    checkUnhandledErrors()
    // ... existing implementation
}

// etc. for Take, Peek, etc.
```

Option B — Generated Check:

In `codegen.go`, emit check before each stack operation:
```go
if stack_error.Len() > 0 {
    _err, _ := stack_error.Peek()
    panic("unhandled error: " + string(_err))
}
```

**Recommendation:** Option A (runtime check) is cleaner and ensures no codegen path can forget.

### 2.4 Error Handling with Consider

**Syntax:**
```ual
@error {}.consider(
    ok: { proceed() }
    error |e|: { handle(e) }
)
```

**Semantics:**
- `ok` branch: `@error` is empty, block executes, nothing popped
- `error |e|` branch: `@error` non-empty, top error popped into `e`, block executes
- After consider completes: handled errors are gone (popped), program may proceed

**Multiple Errors:**
```ual
@error {}.consider(
    ok: { proceed() }
    error |e|: {
        log(e)
        -- optionally handle more:
        while (@error.len > 0) {
            @error pop:next
            log(next)
        }
    }
)
```

**Codegen Changes:**

Current consider codegen checks `_consider_status`. For `@error` consider, generate:
```go
if stack_error.Len() == 0 {
    // ok branch
} else {
    _err_bytes, _ := stack_error.Pop()
    e := string(_err_bytes)
    // error branch with e in scope
}
```

**Parser Changes:**

May need to recognise `@error` as special case in consider parsing, or unify with existing status-based consider by treating non-empty `@error` as implicit `error` status.

---

## 3. The expect(n) Primitive

### 3.1 Design Rationale

Go has:
- `select` — wait for 1 of N channels
- `WaitGroup` — wait for all of N goroutines

Two mechanisms because select can't express "all". And neither can express quorum ("k of N").

ual's `expect(n)` is the generalisation:
- `expect(1)` = select (any one)
- `expect(all)` = WaitGroup (barrier)
- `expect(k)` = quorum

### 3.2 Syntax

```ual
-- Wait for all (default)
@{a, b, c}.expect()

-- Wait for quorum
@{a, b, c}.expect(2)

-- With timeout
@{a, b, c}.expect(2, timeout: 5000)

-- With options as hash stack
@{a, b, c}.expect({ quorum: 2, timeout: 5000 })

-- Chained with consider for error handling
@{a, b, c}.expect(2).consider(
    ok |arrived|: { process(arrived) }
    timeout: { retry() }
    error |e|: { fail(e) }
)
```

### 3.3 Semantics

**Input:** Set of stacks `{s1, s2, ..., sN}`, count `k` (default N), optional timeout

**Behaviour:**
1. Wait until `k` of the N stacks have at least one item available
2. Pop one item from each ready stack (up to k items)
3. Return collected items (as a stack or tuple)
4. If timeout before k ready: set status to `timeout`
5. If any stack errors: push to `@error`, set status to `error`

**Binding in consider:**
- `ok |arrived|`: `arrived` is a stack containing the k items that arrived
- Items are in arrival order (FIFO), not declaration order
- Source identity available if needed via metadata (future extension)

### 3.4 Implementation

**Lexer (`lexer.go`):**
```go
// Add keyword
"expect": TokExpect,
```

**Parser (`parser.go`):**

New AST node:
```go
type ExpectNode struct {
    Sources   []string   // stack names: a, b, c
    Count     *ExprNode  // nil means all
    Options   *ExprNode  // optional hash literal or stack ref
}
```

Parse `@{a, b, c}.expect(...)`:
- `@{` introduces stack set literal
- `.expect` method call
- Arguments: optional count, optional options

**Codegen (`codegen.go`):**

Generate:
```go
func _expect(sources []*ual.Stack, count int, timeoutMs int64) (*ual.Stack, string) {
    results := ual.NewStack(ual.FIFO, ual.TypeBytes)
    
    if count <= 0 {
        count = len(sources)
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    if timeoutMs > 0 {
        ctx, cancel = context.WithTimeout(context.Background(), 
            time.Duration(timeoutMs)*time.Millisecond)
    }
    defer cancel()
    
    // Channel to receive results
    type result struct {
        index int
        data  []byte
        err   error
    }
    resultCh := make(chan result, len(sources))
    
    // Spawn taker for each source
    for i, src := range sources {
        go func(idx int, s *ual.Stack) {
            data, err := s.TakeWithContext(ctx, 0)
            resultCh <- result{idx, data, err}
        }(i, src)
    }
    
    // Collect until count reached or context done
    received := 0
    for received < count {
        select {
        case r := <-resultCh:
            if r.err != nil {
                if r.err.Error() == "cancelled" {
                    continue  // context cancelled, ignore
                }
                stack_error.Push([]byte(r.err.Error()))
                return nil, "error"
            }
            results.Push(r.data)
            received++
        case <-ctx.Done():
            if ctx.Err() == context.DeadlineExceeded {
                return results, "timeout"
            }
            return results, "cancelled"
        }
    }
    
    cancel()  // stop remaining takers
    return results, "ok"
}
```

**Usage in generated code:**
```go
// @{a, b, c}.expect(2, timeout: 5000)
_expect_sources := []*ual.Stack{stack_a, stack_b, stack_c}
_expect_results, _expect_status := _expect(_expect_sources, 2, 5000)
```

**Runtime Support (`stack.go`):**

`TakeWithContext` already exists. May need refinement for cleaner cancellation.

### 3.5 Stack Set Literal Syntax

**New syntax:** `@{a, b, c}` denotes a set of stacks for use with `expect` and `select`.

**Lexer:** 
- `@{` could be single token `TokStackSetStart`
- Or parse as `@` followed by `{` with special handling

**Parser:**
- After `@{`, parse comma-separated identifiers until `}`
- Result is list of stack references

**Note:** This is not a stack of stacks. It's syntactic grouping for multi-stack operations.

---

## 4. Clarification: select vs consider

### 4.1 Distinct Semantics

**select** — branches on *which source* provided input
- Question: "Where did this come from?"
- Labels are source identities
- Used after waiting on multiple stacks

**consider** — branches on *status/outcome*
- Question: "Did this succeed?"
- Labels are status values (ok, error, timeout, custom)
- Used after operations that may fail

These are not interchangeable. A source is an identity. A status is a judgment.

### 4.2 Composition

They chain naturally:
```ual
@{a, b, c}.expect(1).select(
    a |x|: { handle_a(x) }
    b |y|: { handle_b(y) }
    c |z|: { handle_c(z) }
).consider(
    ok: { commit() }
    error |e|: { rollback(e) }
)
```

- `expect(1)` waits for any one source
- `select` branches on which source fired
- `consider` handles any errors from the handler

### 4.3 Implementation Notes

Current `select` implementation inlines the branching. For cleaner separation:

1. `expect(1)` returns: results stack + which source fired
2. `select` is sugar that pattern-matches on source identity
3. `consider` pattern-matches on status

May require internal status variable to carry both source identity and outcome status through the chain.

---

## 5. Parameters as Hash Stacks

### 5.1 Design Rationale

`push:5` is `key:value` where key is implicit (push to data stack).
`timeout:5000` is `key:value` where key is explicit.

These are the same syntax. Parameters to operations are just hash stacks.

### 5.2 Inline Hash Literal Syntax

**Proposed syntax:**
```ual
{ key1: value1, key2: value2 }
-- or without commas:
{ key1: value1  key2: value2 }
```

Creates anonymous hash-perspective stack with given entries.

**Usage:**
```ual
@{a, b}.expect({ quorum: 2, timeout: 5000 })
```

### 5.3 Implementation

**Lexer:** No changes needed. `{`, `}`, `:` already tokenised.

**Parser:** 

In expression context, `{` followed by `ident` `:` indicates hash literal:
```go
type HashLiteralNode struct {
    Entries []HashEntry
}

type HashEntry struct {
    Key   string
    Value ExprNode
}
```

**Codegen:**

Generate temporary hash stack:
```go
_opts := ual.NewStack(ual.Hash, ual.TypeInt64)
_opts.Push(intToBytes(2), []byte("quorum"))
_opts.Push(intToBytes(5000), []byte("timeout"))
```

Operations that accept options read from this stack:
```go
func _expect(sources []*ual.Stack, opts *ual.Stack) (*ual.Stack, string) {
    count := len(sources)
    if v, ok := opts.GetRaw("quorum"); ok {
        count = int(bytesToInt(v))
    }
    timeoutMs := int64(0)
    if v, ok := opts.GetRaw("timeout"); ok {
        timeoutMs = bytesToInt(v)
    }
    // ...
}
```

---

## 6. Rejected Additions

### 6.1 dispatch

**Proposed:** Route output to different destinations based on condition.

**Rejected:** Already expressible with `consider`:
```ual
@data {}.consider(
    is_error: { @errors < value }
    is_priority: { @urgent < value }
    _: { @normal < value }
)
```

No new primitive needed.

### 6.2 gather

**Proposed:** Wait for all of N sources.

**Rejected:** Same as `expect()` with default count (all). Single primitive covers both.

### 6.3 collect

**Proposed:** Accumulate N items before proceeding.

**Rejected:** Same concept as gather/expect. For single stack:
```ual
while (@results.len < n) {
    @results take:item
}
```

Or extend `expect` to work on single stack with count:
```ual
@results.expect(n)
```

No separate primitive needed.

---

## 7. Work-Stealing and Decoupled Views

### 7.1 Design Rationale

ual's perspective system (`LIFO`, `FIFO`, `Indexed`, `Hash`) isn't just about access patterns — it enables **decoupled views** where different accessors see the same data structure through different lenses. The canonical application is work-stealing.

Traditional work-stealing (Go's scheduler, Rust's Rayon, Java's ForkJoinPool) uses a deque where:
- **Owner** accesses LIFO (push/pop from bottom) — cache-friendly, processes recent items
- **Thieves** access FIFO (steal from top) — load balancing, takes old items, reduces contention

In most languages, this requires a specialised data structure with careful lock-free implementation. In ual, it's a natural consequence of perspectives:

```ual
@work = stack.new(Task)

-- Owner's view (default LIFO)
@work push(task)
@work pop:next

-- Thief's view (FIFO perspective)
@work perspective(FIFO)
@work pop:stolen    -- takes from opposite end
```

Same stack, two perspectives, no explicit synchronisation code.

### 7.2 View Handles

To make work-stealing practical, we need stable handles to different perspectives on the same stack:

```ual
@work = stack.new(Task)

-- Create named views
owner = @work.view(LIFO)    -- returns view handle
thief = @work.view(FIFO)    -- different perspective, same data

-- Use views independently
owner.push(task)
owner.pop:next

thief.pop:stolen    -- steals from opposite end
```

Views are lightweight handles, not copies. All views share the underlying data and synchronisation.

### 7.3 Work-Stealing Pool Pattern

A complete work-stealing pool with multiple workers:

```ual
-- Create per-worker queues
@queues = stack.new(stack)    -- stack of stacks

var i i64 = 0
while (i < num_workers) {
    var q = stack.new(Task)
    @queues push(q)
    push:i inc let:i
}

-- Worker function: process own queue, steal when empty
func worker(id i64, own View, others []View) {
    while (running) {
        -- Try own queue first (LIFO, cache-friendly)
        var task Task = Task{}
        if (own.try_pop:task) {
            process(task)
        } else {
            -- Own queue empty, try stealing (FIFO from others)
            var stolen bool = false
            for other in others {
                if (other.try_pop:task) {
                    process(task)
                    push:true let:stolen
                    break
                }
            }
            if (!stolen) {
                yield()    -- nothing to do, yield CPU
            }
        }
    }
}

-- Spawn workers with their queues and steal targets
var worker_id i64 = 0
while (worker_id < num_workers) {
    var own = @queues[worker_id].view(LIFO)
    var others = steal_targets(worker_id, @queues)    -- FIFO views of other queues
    
    @spawn < {
        worker(worker_id, own, others)
    }
    @spawn pop play
    
    push:worker_id inc let:worker_id
}
```

### 7.4 Operations on Views

| Operation | Behaviour |
|-----------|-----------|
| `@s.view(perspective)` | Create view handle with given perspective |
| `view.push(val)` | Push according to view's perspective |
| `view.pop:var` | Pop according to view's perspective (blocking) |
| `view.try_pop:var` | Non-blocking pop, returns bool |
| `view.take:var` | Blocking pop (alias) |
| `view.take(timeout):var` | Pop with timeout |
| `view.len` | Current length (snapshot) |
| `view.perspective` | Get current perspective |

### 7.5 Semantics

**Perspective determines interpretation, not storage:**
- LIFO: `push` adds to top, `pop` removes from top
- FIFO: `push` adds to back, `pop` removes from front
- The underlying storage is always a deque; perspective determines which end you access

**Thread safety:**
- All view operations are thread-safe
- Owner operations (LIFO on own queue) can be lock-free
- Steal operations (FIFO on others' queues) use lightweight synchronisation
- The runtime chooses the optimal implementation based on access patterns

**View lifetime:**
- Views are valid as long as the underlying stack exists
- Multiple views can coexist on the same stack
- Dropping a view doesn't affect the stack or other views

### 7.6 Implementation

The Rust runtime (`rual/src/worksteal.rs`) provides two implementations:

**`WSDeque`** — Traditional Chase-Lev deque:
- Lock-free owner push/pop
- Atomic steal protocol
- Fixed capacity (can be extended)

**`WSStack`** — ual-native work-stealing using Stack + Views:
- Uses `parking_lot::Mutex` for synchronisation  
- Leverages existing Stack infrastructure
- More flexible, slightly higher overhead

The compiler should select the appropriate implementation based on usage patterns:
- Static work-stealing pool → `WSDeque`
- Dynamic/ad-hoc stealing → `WSStack`

### 7.7 Syntax Summary

```ual
-- Create stack and views
@work = stack.new(Task)
owner = @work.view(LIFO)
thief = @work.view(FIFO)

-- Non-blocking operations
if (owner.try_pop:task) { ... }
if (thief.try_pop:stolen) { ... }

-- Blocking operations
owner.take:task
thief.take(1000):stolen    -- 1 second timeout

-- Check state
if (owner.len > 0) { ... }
```

---

## 8. Implementation Priority

### Phase 1: Error Stack Architecture
1. Add unhandled error check to runtime Stack methods
2. Modify consider codegen for `@error` handling
3. Test error accumulation and forced handling

### Phase 2: expect(n) Primitive
1. Add lexer token for `expect`
2. Add parser support for `@{...}` stack set syntax
3. Add parser support for `expect(n)` with optional options
4. Implement `_expect` helper in codegen
5. Test: barrier, quorum, timeout cases

### Phase 3: File I/O
1. Add `file.lines(path, dest)` — streaming
2. Add `file.mmap(path, type)` — random access
3. Error handling integration with `@error`
4. Test with compute blocks for mmap case

### Phase 4: Socket I/O
1. Add `socket.open(network, address)` — client
2. Add `socket.listen(network, address)` — server
3. Add `accept()` for server sockets
4. Integration with `select` for multiplexing

### Phase 5: Syntax Refinements
1. Hash literal syntax `{ k: v, ... }`
2. Options-as-hash-stacks for all operations
3. Clean up parameter passing conventions

### Phase 6: Work-Stealing and Views
1. Add `@s.view(perspective)` syntax to parser
2. Add `View` type to AST and type system
3. Implement `try_pop` non-blocking operation
4. Wire up `WSDeque` and `WSStack` from rual to codegen
5. Add Go runtime equivalents (`pkg/runtime/worksteal.go`)
6. Test: single-owner LIFO, multi-thief FIFO, mixed workloads

---

## 9. Testing Strategy

### Unit Tests

**Error stack:**
- Errors accumulate correctly
- Unhandled error panics on next operation
- Consider clears handled errors
- Multiple errors can be drained

**expect(n):**
- expect() waits for all
- expect(k) waits for quorum
- Timeout fires correctly
- Partial results available on timeout
- Cancellation cleans up goroutines

**File I/O:**
- file.lines reads all lines
- EOF closes stack
- Read errors push to @error
- mmap provides indexed access
- mmap works in compute blocks

**Socket I/O:**
- Client connects and exchanges data
- Server accepts connections
- Network errors push to @error
- Select across multiple sockets works

**Work-stealing:**
- View creation returns valid handle
- LIFO view push/pop works correctly
- FIFO view accesses opposite end
- try_pop returns false when empty
- Multiple views on same stack are coherent
- Concurrent owner/thief access is safe
- Stress test: no lost or duplicated items under contention

### Integration Tests

**Pipeline with files:**
```ual
@lines = stack.new(string)
file.lines("input.txt", @lines)
@output = stack.new(string)
file.sink("output.txt", @output)

@spawn < {
    var line string = ""
    @lines take:line
    while (!@lines.closed) {
        @output < transform(line)
        @lines take:line
    }
    @output close
}
```

**Distributed quorum:**
```ual
@r1 = socket.open("tcp", "node1:8080")
@r2 = socket.open("tcp", "node2:8080")
@r3 = socket.open("tcp", "node3:8080")

@{r1, r2, r3}.expect(2, timeout: 5000).consider(
    ok |responses|: { consensus(responses) }
    timeout: { failover() }
)
```

**Work-stealing pool:**
```ual
-- Create work queues for 4 workers
@q0 = stack.new(i64)
@q1 = stack.new(i64)
@q2 = stack.new(i64)
@q3 = stack.new(i64)

@results = stack.new(i64)
@done = stack.new(i64)

-- Worker 0: owns q0, steals from q1,q2,q3
@spawn < {
    var own = @q0.view(LIFO)
    var victims = [@q1.view(FIFO), @q2.view(FIFO), @q3.view(FIFO)]
    
    while (running) {
        var task i64 = 0
        if (own.try_pop:task) {
            @results < (task * task)
        } else {
            -- Try stealing
            for v in victims {
                if (v.try_pop:task) {
                    @results < (task * task)
                    break
                }
            }
        }
    }
    @done < 1
}

-- ... workers 1, 2, 3 similarly ...
```

---

## 10. Open Questions

1. **Stack set syntax:** Is `@{a, b, c}` the right spelling? Alternatives: `@[a, b, c]`, `expect(a, b, c)` without grouping.

2. **Source identity in expect results:** When expect(2) of {a,b,c} completes, do we know which two fired? Current design: no, just get the values. May need metadata.

3. **mmap lifecycle:** When is the file unmapped? On stack garbage collection? Explicit close? Need clear semantics.

4. **Socket framing:** Raw bytes or message-based? Probably need both: raw for FIFO byte stream, message-based for discrete packets.

5. **Error stack scope:** Global `@error` or per-goroutine? Global is simpler but may cause confusion in concurrent code. Per-goroutine matches Go's error handling locality.

6. **View syntax:** Should `@s.view(LIFO)` return a new type (`View`) or just a configured reference to the stack? If a new type, how do views interact with generic stack operations?

7. **Work-stealing selection:** When a thief has multiple potential victims, what order should it try? Random (avoid contention), round-robin (fairness), or nearest-neighbour (cache locality)?

---

## 11. References

- ual v0.7.4 compiler: `../cmd/ual/` (lexer, parser, codegen_go, codegen_rust, main)
- ual v0.7.4 interpreter: `../cmd/iual/` (interp, interp_control, interp_expr)
- ual v0.7.4 Rust runtime: `../rual/` (stack, value, view, sync, worksteal)
- Shared packages: `../pkg/ast/`, `../pkg/lexer/`, `../pkg/parser/`, `../pkg/runtime/`, `../pkg/version/`
- Runtime: `../pkg/runtime/` (Stack, Value, ValueStack, ScopeStack, View, Bring, Walk)
- Existing constructs: `MANUAL.md`, `COMPUTE_SPEC_V2.md`, `CONCURRENCY.md`
- Examples: `../examples/` directory (92 programs)
- Benchmarks: `../tests/benchmarks/` directory

**Implementation Note (v0.7.4):** As of v0.7.4, all three execution modes (Go compiler, Rust compiler, iual interpreter) have achieved 100% parity. All 92 example tests pass identically across all three implementations. The Rust backend is production-ready. This provides a solid foundation for implementing v0.8 features across all execution modes simultaneously.

---

*End of specification.*