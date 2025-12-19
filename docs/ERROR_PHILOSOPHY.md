# The Philosophy of Error Handling in ual

**Version 0.8 Design Rationale**

---

## 1. The Central Claim

Errors are not values to be passed around. They are facts about what happened — evidence of failed state transitions. Treating errors as ordinary return values, as most languages do, creates a fundamental category error that leads to systematic mishandling.

ual's error system is built on a different premise: **errors must be acknowledged before computation can proceed**. This is not a restriction on the programmer's freedom; it is a recognition that ignoring errors does not make them disappear — it merely defers their consequences to a place where they become harder to diagnose and fix.

---

## 2. The Problem with Error-as-Value

### 2.1 The Go Pattern

Go popularised explicit error returns:

```go
result, err := doSomething()
if err != nil {
    return err
}
// use result
```

This has virtues: errors are visible, local, and explicit. But Go's approach has a critical flaw — the check is optional:

```go
result, _ := doSomething()  // error silently discarded
```

The compiler accepts this without complaint. The programmer might intend to handle the error later, or might have forgotten, or might be prototyping quickly with plans to "fix it properly" that never materialise. The language provides no mechanism to distinguish intentional discard from negligent omission.

The consequences compound. A discarded error means:

- `result` may be a zero value, invalid, or partially initialised
- Subsequent operations proceed on potentially corrupt state
- When failure eventually manifests, the root cause is obscured
- Debugging requires tracing back through code paths that looked "successful"

### 2.2 The Exception Pattern

Languages with exceptions (Java, Python, C++) take the opposite approach: errors propagate automatically unless caught.

```python
try:
    result = do_something()
    use(result)
except SomeError as e:
    handle(e)
```

This solves the "forgotten check" problem — unhandled errors propagate rather than disappear. But exceptions introduce their own pathologies:

**Non-locality**: An exception can be thrown anywhere and caught anywhere. Control flow becomes invisible. Reading a function, you cannot know which lines might not execute.

**The catch-all temptation**:

```python
try:
    complex_operation()
except Exception:
    pass  # "handle" all errors by ignoring them
```

This is worse than Go's `_` discard because it catches errors the programmer didn't anticipate and may not understand.

**Resource management complexity**: Code must be written defensively, assuming any line might not execute. This leads to `try`/`finally` nesting, context managers, RAII patterns — all mechanisms to cope with a control flow model that fights against local reasoning.

**Performance unpredictability**: Exception handling typically involves stack unwinding, which has non-trivial cost and makes performance analysis difficult.

### 2.3 The Rust Pattern

Rust's `Result` type represents the current state of the art:

```rust
fn do_something() -> Result<Value, Error> { ... }

// Must handle both cases
match do_something() {
    Ok(value) => use(value),
    Err(e) => handle(e),
}

// Or propagate explicitly
let value = do_something()?;
```

Rust makes it a compile-time error to ignore a `Result`. You cannot access the success value without acknowledging the error case. The `?` operator provides ergonomic propagation, but propagation is still explicit — you see it in the code.

If you truly want to ignore an error, you must write `.unwrap()` or `.expect()` — explicit markers that say "I accept this will panic if wrong." The intentionality is captured in the source.

Rust's approach is excellent. ual's differs in mechanism but shares the core principle: **errors must be explicitly acknowledged**.

---

## 3. ual's Model: The Error Stack

### 3.1 Separation of Channels

In ual, errors flow through a separate channel from data:

```
Data channel:    @source → operation → @destination
Error channel:   operation → @error (if failed)
```

This separation reflects a conceptual truth: the error is metadata *about* the operation, not a value *from* the operation. When `file.read()` fails, the error isn't alternative content — it's information about why content wasn't obtained.

The `@error` stack accumulates errors as they occur:

```ual
operation1()    // might fail
operation2()    // might also fail
operation3()    // and this

// @error now contains 0, 1, 2, or 3 errors
```

### 3.2 Forced Acknowledgment

Here is ual's key mechanism: **any stack operation checks whether `@error` is non-empty. If unacknowledged errors exist, the program panics.**

```ual
risky_operation()      // pushes to @error on failure
@data push:42          // PANIC: unhandled error exists
```

This is not a crash caused by the error itself. It is a crash caused by attempting to continue without acknowledging the error. The distinction matters.

### 3.3 Acknowledgment Options

The programmer has full flexibility in how to acknowledge:

**Handle and recover:**

```ual
risky_operation()
@error {}.consider(
    ok: { 
        // no error occurred
        proceed_normally() 
    }
    error |e|: { 
        // error occurred, e contains message
        log(e)
        use_fallback()
    }
)
```

**Propagate to caller:**

```ual
func do_work() {
    risky_operation()
    @error {}.consider(
        ok: { continue_work() }
        error |e|: { 
            status:error(e)  // propagate via status
            return 
        }
    )
}
```

**Explicitly discard (when truly appropriate):**

```ual
optional_operation()   // we don't care if this fails
@error clear           // explicit: "I acknowledge and discard"
continue_regardless()
```

**Accumulate and handle in batch:**

```ual
step1()  // might fail
step2()  // might fail  
step3()  // might fail

@error {}.consider(
    ok: { commit_results() }
    error |e|: {
        // handle first error
        log(e)
        // drain remaining errors
        while (@error: len() > 0) {
            @error pop:next
            log(next)
        }
        rollback()
    }
)
```

### 3.4 What You Cannot Do

You cannot:

```ual
risky_operation()
// pretend nothing happened
@data push:42      // PANIC
more_operations()
even_more()
```

The program will not allow you to ignore an error and continue mutating state. This is the restriction, and it is the point.

---

## 4. Why This Design

### 4.1 Errors Do Not Disappear

When an error is ignored, the error condition doesn't vanish — only the programmer's awareness of it does. The consequences remain:

- Network request failed → data is stale or missing
- File write failed → data is lost
- Allocation failed → pointer is null or uninitialised
- Parse failed → structure contains garbage

Continuing computation on corrupt state produces corrupt results. If those results are persisted, corruption spreads. If they affect subsequent decisions, the corruption compounds.

Silent error propagation is technical debt with compound interest.

### 4.2 Fail-Fast vs Fail-Gracefully

"Fail gracefully" is often invoked as a virtue, contrasted with crashes. But examine what "graceful" failure typically means:

```go
result, err := fetchData()
if err != nil {
    log.Printf("warning: %v", err)
    result = defaultValue  // or worse: leave result uninitialised
}
// continue with result
```

This pattern:
- Logs a message that will be lost in thousands of other log lines
- Substitutes data that may be inappropriate for the actual situation
- Continues execution as if the operation succeeded
- Surfaces as mysterious bugs far from the actual failure

This is not graceful. It is error laundering — converting explicit failure into implicit corruption.

True graceful degradation requires *understanding* the error and making an informed decision:

```ual
fetch_data(@result)
@error {}.consider(
    ok: { process(@result) }
    error |e|: {
        if is_transient(e) {
            retry_with_backoff()
        } elseif has_cached_version() {
            use_cached_data()
            notify_staleness()
        } else {
            return_service_unavailable()
        }
    }
)
```

This is graceful — the error is acknowledged, its nature is examined, and appropriate action is taken. ual's model encourages this; error-as-value models permit it but don't encourage it.

### 4.3 The Transaction Analogy

Database systems understood this decades ago:

```sql
BEGIN;
INSERT INTO orders (...);      -- fails: constraint violation
INSERT INTO audit_log (...);   -- should this execute?
COMMIT;                        -- should this succeed?
```

No reasonable database allows a failed transaction to continue and commit. The failure must be handled — rollback, retry, or explicitly acknowledge and proceed with a new transaction.

ual treats computation similarly: the program is a sequence of state transitions, and a failed transition must be addressed before subsequent transitions occur. This is not novel; it is applying established principles from data management to general computation.

### 4.4 Concurrency Demands It

In concurrent systems, error handling becomes critical. Consider:

```ual
@spawn {
    while running {
        task = @queue take
        process(task)      // might fail
        @done push:task
    }
}
```

If `process(task)` fails and errors are silently ignored:
- The task appears completed (pushed to `@done`)
- Other workers may depend on its results
- The failure is invisible to coordination logic
- System state diverges from assumed state

Distributed systems literature is full of catastrophic failures caused by ignored errors in concurrent code. ual's model makes such oversights impossible — not through programmer discipline, but through language mechanics.

---

## 5. Objections and Responses

### 5.1 "This is too strict for prototyping"

During rapid prototyping, programmers want to focus on the happy path. Forced error handling seems burdensome.

Response: `@error clear` exists. Write your prototype:

```ual
experimental_thing()
@error clear
next_experiment()
@error clear
```

The explicit clears document that error handling is deferred. When the prototype matures, these clears become TODO markers. They're visible in code review, searchable with grep, impossible to forget.

This is better than Go's `_` because the acknowledgment is at the point of potential failure, not buried in a return value assignment.

### 5.2 "Performance overhead of checking @error"

Every stack operation checks `@error.Len() > 0`. Is this expensive?

Response: This is an integer comparison — effectively free. The check is:

```go
if globalErrorStack.Len() > 0 {
    panic("unhandled error: " + ...)
}
```

This is one memory read, one comparison, one predictable branch (almost always not-taken). Modern CPUs execute this in nanoseconds with no pipeline disruption.

If profiling reveals this matters (it won't), the check could be disabled in release builds. But the check has near-zero cost, and its benefit is substantial.

### 5.3 "What about best-effort operations?"

Some operations are genuinely optional — logging, metrics, cache warming. Failure doesn't matter.

Response: Express that explicitly:

```ual
warm_cache()        // might fail
@error clear        // we don't care

send_metrics()      // might fail
@error clear        // still don't care

do_actual_work()    // this one matters
@error {}.consider(
    ok: { ... }
    error |e|: { ... }
)
```

The clears document your intent. A reader knows immediately which operations are critical and which are optional. This is *more* information than silent error ignoring provides.

### 5.4 "I want to handle errors later, not immediately"

Sometimes you want to attempt several operations and handle failures in aggregate.

Response: This is supported:

```ual
attempt1()
attempt2()
attempt3()

// Handle all accumulated errors
@error {}.consider(
    ok: { all_succeeded() }
    error |e|: { 
        at_least_one_failed(e)
        // drain remaining if desired
    }
)
```

The only restriction is that you must handle before the next stack operation. If you need to do non-stack computation between attempts, that's fine — the check triggers on stack operations, not on every line.

### 5.5 "Rust does this at compile time, isn't that better?"

Rust's compile-time checking catches errors earlier. ual's runtime checking is strictly less powerful.

Response: Compile-time checking is indeed preferable when available. ual's checking is a pragmatic choice given the language's design:

- ual compiles to Go, which lacks Rust's type system
- ual's stack-based model makes static error tracking complex
- Runtime checking still catches 100% of error-ignoring bugs (at runtime)

The comparison isn't "ual vs Rust" but "ual vs Go/Python/JavaScript" — languages where ignoring errors is trivially easy. Against that baseline, ual's runtime checking is a significant improvement.

---

## 6. Scope of @error

### 6.1 Global vs Per-Goroutine

The design document identifies an open question: should `@error` be global or per-goroutine?

**Global @error:**
- Simpler implementation
- Errors from spawned tasks visible to parent
- But: concurrent pushes create race conditions
- And: errors from one goroutine could block another's stack operations

**Per-goroutine @error:**
- Each spawn gets its own error stack
- No race conditions on @error itself
- Matches Go's goroutine-local error handling
- But: errors must be explicitly communicated to parent

The per-goroutine model is almost certainly correct. Consider:

```ual
@spawn {
    operation_a()  // fails, pushes to @error
    // this spawn should handle its own errors
}

@spawn {
    operation_b()  // succeeds
    @data push:1   // should this panic because spawn-1 failed?
}
```

With global @error, spawn-2 would be blocked by spawn-1's error. This makes no sense — they're independent computation.

Recommendation: **@error is per-goroutine**. Errors must be explicitly communicated:

```ual
@results = stack.new(bytes)
@errors = stack.new(bytes)

@spawn {
    operation()
    @error {}.consider(
        ok: { @results push:value }
        error |e|: { @errors push:e }  // explicit communication
    )
}
```

This maintains the forced-acknowledgment property while respecting concurrency boundaries.

### 6.2 Error Stack Lifecycle

Each execution context has its own @error:
- Main program: one @error, exists for program lifetime
- Each spawn: one @error, exists for spawn lifetime
- Spawn termination with non-empty @error: panic (error was never handled)

This last point is important: you cannot "exit your way out" of error handling. A spawned task that terminates with unhandled errors is a bug, caught at runtime.

---

## 7. Comparison Summary

| Aspect | Go | Exceptions | Rust | ual |
|--------|----|-----------:|------|-----|
| Error visibility | Explicit (return value) | Hidden (throw site) | Explicit (Result type) | Explicit (separate stack) |
| Ignoring errors | Easy (`_`) | Easy (empty catch) | Hard (must `.unwrap()`) | Hard (must `clear` or handle) |
| Propagation | Manual | Automatic | Manual (but ergonomic `?`) | Manual (via status) |
| Non-local control flow | No | Yes | No | No |
| Compile-time checking | No | No | Yes | No |
| Runtime checking | No | N/A | Panic on unwrap | Yes (every stack op) |
| Error accumulation | No | No | No | Yes (@error stack) |

---

## 8. Conclusion

ual's error handling is not about restricting programmers. It is about aligning language mechanics with computational reality.

Errors represent failed state transitions. Ignoring failed transitions and continuing computation produces corrupt state. Corrupt state propagates. Systems built on ignored errors are systems with latent failures waiting to manifest.

The forced-acknowledgment model makes a simple demand: if something failed, say what you're doing about it. Handle it, propagate it, discard it explicitly — any of these is acceptable. What is not acceptable is pretending it didn't happen.

This is not a burden. It is a service. The language is telling you: something went wrong here. You must decide what that means.

That decision is the programmer's job. Forcing the decision to be made is the language's job. ual does its job.

---

*This document accompanies DESIGN_v0.8.md and describes the philosophical basis for ual's error handling architecture.*
