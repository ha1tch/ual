# ual Manual

**Version 0.7.3**

A systems language for orchestration and embedded computation, presented with a scripting-style surface.

---

## What is ual?

ual is a stack-based language that compiles to Go. It occupies a unique position: the ergonomics of a scripting language with the performance and predictability of a systems language.

```
Surface Feel          Actual Semantics
─────────────────     ─────────────────────────────────
Forth-like ops        Explicit data flow, typed containers
Erlang-like select    Deterministic scheduling, structured
Inline DSL blocks     Native Go codegen, zero overhead
```

ual is designed for:

- **Data pipelines** where values flow through transformation stages
- **Concurrent coordination** where multiple data streams must be managed
- **Embedded computation** where predictable memory and performance matter
- **Orchestration logic** that glues computational kernels together

## Prerequisites

Go 1.22 or later must be installed. This is the only dependency.

```bash
# Check Go installation
go version
```

## Installation

```bash
# Using Make (recommended)
make build

# Or using build.sh
./build.sh

# Or manually:
go build -o ual ./cmd/ual
go build -o iual ./cmd/iual
```

## Usage

### Compiler (ual)

```bash
# Commands
ual compile program.ual     # Compile to Go source (.go)
ual build program.ual       # Build executable binary
ual run program.ual         # Compile and run immediately
ual tokens program.ual      # Show lexer tokens
ual ast program.ual         # Show parse tree
ual version                 # Show version
ual help                    # Show help

# Options
-o, --output <path>         # Output file path
-q, --quiet                 # Suppress non-error output
-v, --verbose               # Show detailed compilation info
-vv, --debug                # Show debug information
-O, --optimize              # Use optimised dstack
--version                   # Show version and exit

# Examples
ual compile program.ual           # Creates program.go
ual build -o myapp program.ual    # Creates myapp binary
ual -q run program.ual            # Run quietly
ual -v build program.ual          # Verbose build
```

### Interpreter (iual)

The interpreter runs UAL programs directly without compilation. Useful for development and testing.

```bash
# Commands
iual program.ual            # Run program directly
iual run program.ual        # Same as above
iual version                # Show version
iual help                   # Show help

# Options
-t, --trace                 # Trace execution
-q, --quiet                 # Suppress non-essential output
--verbose                   # Verbose output
--debug                     # Debug mode (implies --trace)

# Examples
iual program.ual            # Run directly
iual --trace program.ual    # Trace execution
iual -q program.ual         # Quiet mode
```

**Performance Note:** The interpreter is approximately 10-50x slower than compiled code. For production workloads, use `ual build`.

**Concurrency:** The interpreter uses real goroutines for `@spawn pop play`, matching the compiler's semantics. Both tools share the same runtime types from `pkg/runtime/`.

## Quick Start

```ual
-- Hello ual: compute factorial of 5
@numbers = stack.new(i64)
@numbers push(5)

@numbers {
}.compute({|n|
    var result = 1
    var i = 1
    while i <= n {
        result = result * i
        i = i + 1
    }
    return result
})

@numbers pop
print
-- Output: 120
```

---

## Part 1: Stacks and Perspectives

### Stacks are Containers

In ual, data lives in **stacks**. A stack is a typed container that can be accessed in different ways.

```ual
-- Create stacks with explicit types
@integers = stack.new(i64)      -- 64-bit integers
@floats = stack.new(f64)        -- 64-bit floats
@text = stack.new(string)       -- strings
@data = stack.new(bytes)        -- raw bytes
@flags = stack.new(bool)        -- booleans
```

### Perspectives

The same stack can be accessed through different **perspectives**:

| Perspective | Behaviour | Use Case |
|-------------|-----------|----------|
| **LIFO** | Last-in, first-out | Call stacks, undo history |
| **FIFO** | First-in, first-out | Queues, pipelines |
| **Indexed** | Random access by position | Arrays, vectors |
| **Hash** | Access by key | Records, dictionaries |

```ual
-- Same data, different access patterns
@tasks = stack.new(i64, LIFO)     -- stack behaviour
@queue = stack.new(i64, FIFO)     -- queue behaviour
@array = stack.new(i64, Indexed)  -- array behaviour
@record = stack.new(f64, Hash)    -- key-value behaviour
```

### Default Stacks

ual provides default stacks for common patterns (Forth-style):

| Stack | Type | Purpose |
|-------|------|---------|
| `@dstack` | i64 | Data stack (implicit target) |
| `@rstack` | i64 | Return stack (temporary storage) |
| `@bool` | bool | Boolean/comparison results |
| `@error` | bytes | Error handling |

```ual
-- Operations without @ use @dstack implicitly
push(10)
push(20)
add
print    -- Output: 30
```

---

## Part 2: Basic Operations

### Push and Pop

```ual
@numbers push(42)           -- push value
@numbers push(10)
val = @numbers pop()        -- pop into variable

-- Colon shorthand
@numbers push:100           -- same as push(100)

-- Negative literals
@numbers push:-42           -- negative integer
@floats push:-3.14          -- negative float
```

**Type Safety**: The compiler enforces type compatibility at compile time:

```ual
@integers = stack.new(i64)
@integers push:42           -- OK: integer to integer stack
@integers push:3.14         -- ERROR: cannot push float to i64 stack

@floats = stack.new(f64)
@floats push:3.14           -- OK: float to float stack
@floats push:42             -- OK: integer widened to float
```

### Stack Operators (Forth-Style)

Arithmetic:
```ual
push:10 push:3 add      -- 13
push:10 push:3 sub      -- 7
push:10 push:3 mul      -- 30
push:10 push:3 div      -- 3
push:10 push:3 mod      -- 1
```

Stack manipulation:
```ual
push:5 dup              -- 5 5
push:5 push:3 swap      -- 3 5
push:5 push:3 drop      -- 5
push:5 push:3 over      -- 5 3 5
push:1 push:2 push:3 rot -- 2 3 1
```

Comparison (results go to @bool):
```ual
push:5 push:3 gt        -- true  (5 > 3)
push:5 push:5 eq        -- true  (5 == 5)
push:5 push:3 lt        -- false (5 < 3)
```

### Output

```ual
print       -- peek and print (non-destructive)
dot         -- pop and print (destructive, Forth-style)
```

### Return Stack

```ual
push:42
tor         -- move to @rstack (>r in Forth)
push:99
fromr       -- move back from @rstack (r> in Forth)
dot         -- 42
dot         -- 99
```

---

## Part 3: Variables and Control Flow

### Variables

```ual
var x = 10              -- mutable variable
let pi = 3.14159        -- immutable constant

x = x + 1               -- assignment
```

### Control Flow

```ual
-- If/else
if x > 0 {
    push:1
} elseif x < 0 {
    push:-1
} else {
    push:0
}

-- While loop
var i = 0
while i < 10 {
    push:i
    i = i + 1
}

-- Break and continue
while true {
    if done {
        break
    }
    if skip {
        continue
    }
    process()
}
```

### Functions

```ual
func square(n i64) i64 {
    return n * n
}

func add(a i64, b i64) i64 {
    return a + b
}

-- Usage
push:5
square()
dot         -- 25
```

---

## Part 4: Stack Blocks

Stack blocks group operations on a specific stack:

```ual
@calculator {
    push:10
    push:20
    add
    push:5
    mul
}
-- Result: (10 + 20) * 5 = 150
```

### Hash Perspective with Set/Get

```ual
@person = stack.new(string, Hash)
@person set("name", "Alice")
@person set("city", "London")

@person get("name")     -- pushes "Alice" to @dstack
print                   -- Output: Alice
```

---

## Part 5: The Compute Construct

The `.compute()` construct is ual's "optimization island" — a block where arithmetic runs on native CPU types with zero serialisation overhead.

### Basic Compute

```ual
@physics = stack.new(f64)
@physics push(10.0)     -- mass
@physics push(5.0)      -- velocity

@physics {
}.compute({|v, m|
    -- Kinetic energy: ½mv²
    var ke = 0.5 * m * v * v
    return ke
})

@physics pop
dot     -- 125.0
```

### Bindings

The `{|a, b|}` syntax pops values from the stack into native variables:

```ual
{|top, second, third|}   -- pops 3 values (top is first popped)
{||}                     -- empty bindings (no pops)
```

### Self Access

For Hash perspective stacks, use `self.property`:

```ual
@body = stack.new(f64, Hash)
@body set("mass", 10.0)
@body set("velocity", 5.0)

@body {
}.compute({||
    var m = self.mass
    var v = self.velocity
    return 0.5 * m * v * v
})
```

For Indexed perspective, use `self[i]`:

```ual
@vec = stack.new(f64, Indexed)
@vec push(1.0)
@vec push(2.0)
@vec push(3.0)

@vec {
}.compute({||
    var sum = 0.0
    var i = 0
    while i < 3 {
        sum = sum + self[i] * self[i]
        i = i + 1
    }
    return sum
})
-- Result: 1 + 4 + 9 = 14
```

### Local Arrays

Compute blocks support fixed-size local arrays for algorithms:

```ual
@result = stack.new(i64)
@result push(20)

@result {
}.compute({|n|
    var dp[100]         -- local array
    dp[0] = 0
    dp[1] = 1
    
    var i = 2
    while i <= n {
        dp[i] = dp[i - 1] + dp[i - 2]
        i = i + 1
    }
    return dp[n]
})
-- Fibonacci(20) = 6765
```

### Math Functions

Standard math functions are available inside compute blocks:

```ual
sqrt(x)     sin(x)      cos(x)      tan(x)
log(x)      exp(x)      pow(x, y)
floor(x)    ceil(x)     round(x)
abs(x)      min(x, y)   max(x, y)
```

### Performance

Compute blocks compile to native Go loops. Benchmarks show:

| vs Go | vs C | vs Python |
|-------|------|-----------|
| 1.0x (identical) | 1.0-1.7x | 30-100x faster |

Use `.compute()` freely for any computation >1μs. For very short computations (<100ns), batch multiple operations into a single compute block.

---

## Part 6: Concurrency

### Select

The `.select()` construct waits on multiple stacks concurrently:

```ual
@inbox = stack.new(bytes)
@commands = stack.new(bytes)

@inbox {
    -- setup code
}.select(
    @inbox {|msg|
        process_message(msg)
    }
    @commands {|cmd|
        execute_command(cmd)
    }
    _: {
        -- default case (non-blocking)
        idle()
    }
)
```

### Timeouts

Each case can have its own timeout:

```ual
@inbox {|msg|
    process(msg)
    timeout(1000, {||
        log("No message for 1 second")
        retry()     -- restart waiting on this case
    })
}
```

### Spawn

Launch concurrent tasks:

```ual
@spawn worker()         -- runs in new goroutine
```

---

## Part 7: Error Handling

### Consider

The `.consider()` construct provides pattern matching on outcomes:

```ual
@data {
    risky_operation()
}.consider(
    ok: {
        process_result()
    }
    error |e|: {
        log_error(e)
        recover()
    }
    _: {
        -- default handler
    }
)
```

### Status Setting

Functions can set status explicitly:

```ual
func divide(a i64, b i64) i64 {
    if b == 0 {
        status:error("division by zero")
        return 0
    }
    status:ok
    return a / b
}
```

### Error Stack

The `@error` stack captures errors:

```ual
@error {
    push("Something went wrong")
}

-- Check if errors exist
if @error.len() > 0 {
    handle_errors()
}
```

---

## Part 8: Traversal Operations

### Reduce

Fold to single value:

```ual
sum = @numbers reduce(0, {|acc, x| return acc + x })
```

### Map

Transform to new stack:

```ual
@strings map(@lengths, {|s| return len(s) })
```

### Walk and Filter (Disabled)

The `walk()` and `filter()` operations are currently disabled pending design review. Use explicit loops or `reduce()` instead:

```ual
-- Instead of: @source walk(@dest, {|x| return x * 2 })
-- Use explicit iteration:
var i i64 = 0
while (i < @source: len()) {
    var x i64 = @source: get(i)
    @dest push:(x * 2)
    i = i + 1
}

-- Instead of: @numbers filter(@evens, {|x| return x % 2 == 0 })
-- Use explicit iteration with conditional:
var i i64 = 0
while (i < @numbers: len()) {
    var x i64 = @numbers: get(i)
    if (x % 2 == 0) {
        @evens push:x
    }
    i = i + 1
}
```

---

## Part 9: Views

Views provide independent perspectives on the same data:

```ual
@tasks = stack.new(Task)

owner = LIFO.on(@tasks)   -- pops newest (cache-hot)
thief = FIFO.on(@tasks)   -- steals oldest (cache-cold)

-- Work-stealing in 3 lines
```

This enables patterns like work-stealing where an owner works LIFO (cache-friendly) while thieves steal FIFO (minimize contention).

---

## Part 10: Bring

Atomic transfer between stacks with type conversion:

```ual
@source = stack.new(i64)
@dest = stack.new(f64)

@source push(42)
@source bring(@dest)      -- transfers and converts i64 → f64
```

Bring is atomic: if conversion fails, source is unchanged.

---

## Appendix A: Type Reference

| Type | Description | Size |
|------|-------------|------|
| `i64` | Signed 64-bit integer | 8 bytes |
| `u64` | Unsigned 64-bit integer | 8 bytes |
| `f64` | 64-bit float (IEEE 754) | 8 bytes |
| `bool` | Boolean | 1 byte |
| `string` | UTF-8 string | variable |
| `bytes` | Raw bytes | variable |

---

## Appendix B: Perspective Semantics

| Perspective | Push | Pop | Peek | Use Case |
|-------------|------|-----|------|----------|
| LIFO | End | End | End | Stacks, undo |
| FIFO | End | Head | Head | Queues |
| Indexed | End | By index | By index | Arrays |
| Hash | By key | By key | By key | Records |

---

## Appendix C: Compute Block Reference

### Bindings

```ual
{|a|}           -- pop 1 value
{|a, b|}        -- pop 2 values (a=top, b=second)
{|a, b, c|}     -- pop 3 values
{||}            -- no bindings (use self.property for Hash)
```

### Returns

```ual
return x            -- push 1 value
return a, b         -- push 2 values (b ends up on top)
-- no return        -- void compute (consumer pattern)
```

### Local Variables

```ual
var x = 0           -- mutable, type inferred
var y = 0.0         -- float
var buf[100]        -- fixed-size array
```

### Control Flow

```ual
if condition { }
while condition { }
break
continue
```

### Self Access

| Perspective | Read | Write |
|-------------|------|-------|
| Indexed | `self[i]` | Not in compute |
| Hash | `self.property` | Not in compute |
| LIFO/FIFO | Use bindings | Use return |

---

## Appendix D: Quick Reference Card

```
STACK CREATION
    @name = stack.new(type)
    @name = stack.new(type, perspective)

PUSH/POP
    @s push(value)      @s push:value
    @s push:-42         @s push:-3.14     -- negative literals
    x = @s pop()        @s pop

ARITHMETIC
    add sub mul div mod
    neg abs inc dec
    min max

STACK OPS
    dup drop swap over rot
    tor fromr

COMPARISON (→ @bool)
    eq ne lt gt le ge

BITWISE
    band bor bxor bnot shl shr

OUTPUT
    print (peek)    dot (pop)

CONTROL
    if { } elseif { } else { }
    while { }
    break continue

FUNCTIONS
    func name(args) rettype { }
    return value

COMPUTE
    @s {}.compute({|bindings| ... return value })
    self.property   self[i]
    var x = 0       var arr[N]

CONCURRENCY  
    @s {}.select( cases )
    timeout(ms, handler)
    retry() restart()
    @spawn task()

ERROR HANDLING
    @s {}.consider( ok: {} error: {} _: {} )
    status:label    status:label(value)

TRAVERSAL
    @s reduce(init, fn)
    @s map(@d, fn)
    -- walk/filter disabled, use explicit loops

BRING
    @source bring(@dest)
```

---

## Further Reading

- `CHANGELOG.md` — Version history and feature details
- `COMPUTE_SPEC_V2.md` — Compute block specification
- `DESIGN_v0.8.md` — Design document for v0.8 features
- `ERROR_PHILOSOPHY.md` — Error handling philosophy
- `BENCHMARK_SPECIFICATION.md` — Benchmark suite specification
- `../examples/` — Working code examples (71 programs)

---

*ual v0.7.3 — A systems language disguised as a scripting language.*
