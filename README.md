# ual

**Version 0.7.2**

A systems language for orchestration and embedded computation, presented with a scripting-style surface.

```
Surface Feel              Actual Semantics
─────────────────         ─────────────────────────────────
Forth-like stack ops      Explicit data flow, typed containers
Erlang-like select        Deterministic scheduling, structured
Inline DSL blocks         Native Go codegen, zero overhead
```

## Quick Start

```bash
# Build the compiler
make build

# Run a program
./ual run examples/01_fibonacci.ual
```

## Hello ual

```ual
-- Compute factorial of 5
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

## Core Concepts

### Stacks with Perspectives

Data lives in typed stacks. The same stack can be accessed through different perspectives:

| Perspective | Behaviour | Use Case |
|-------------|-----------|----------|
| **LIFO** | Last-in, first-out | Call stacks, undo |
| **FIFO** | First-in, first-out | Queues, pipelines |
| **Indexed** | Random access | Arrays, vectors |
| **Hash** | Key-value access | Records, objects |

```ual
@tasks = stack.new(i64, LIFO)     -- stack
@queue = stack.new(i64, FIFO)     -- queue
@array = stack.new(i64, Indexed)  -- array
@record = stack.new(f64, Hash)    -- key-value
```

### The Compute Construct

The `.compute()` block is ual's "optimization island" — arithmetic runs on native CPU types with zero serialisation overhead:

```ual
@physics = stack.new(f64)
@physics push(10.0)   -- mass
@physics push(5.0)    -- velocity

@physics {
}.compute({|v, m|
    var ke = 0.5 * m * v * v
    return ke
})
-- Result: 125.0
```

**Performance**: Compute blocks compile to native Go loops. Benchmarks show ual matches Go and C performance:

| Algorithm | C (-O2) | Go | ual | Python |
|-----------|---------|-----|-----|--------|
| Mandelbrot | 4,078 ns | 4,161 ns | 4,170 ns | 116,461 ns |
| Leibniz π | 127 μs | 120 μs | 120 μs | 6,773 μs |

### Concurrency with Select

Wait on multiple stacks concurrently:

```ual
@inbox {}.select(
    @inbox {|msg| process(msg) }
    @commands {|cmd| execute(cmd) }
    _: { idle() }  -- default case
)
```

### Structured Error Handling

Pattern match on outcomes with `.consider()`:

```ual
@data {
    risky_operation()
}.consider(
    ok: { process_result() }
    error |e|: { handle_error(e) }
)
```

## Features

### Implemented (v0.7.2)

- **Four perspectives**: LIFO, FIFO, Indexed, Hash
- **Compute blocks**: Native codegen with local arrays, math functions
- **Container access**: `self.property` (Hash), `self[i]` (Indexed)
- **Select construct**: Multi-stack waiting with timeouts
- **Consider construct**: Structured error handling with status
- **Forth-style operations**: push, pop, dup, swap, rot, arithmetic
- **Negative literals**: `push:-42`, `push:-3.14`
- **Type checking**: Compile-time type validation for push operations
- **Control flow**: if/elseif/else, while, break, continue
- **Functions**: With typed parameters and returns
- **Spawn**: Goroutine-based concurrency
- **Views**: Decoupled perspectives on shared data
- **Work-stealing**: LIFO owner + FIFO thieves pattern
- **Bring**: Atomic transfer with type conversion

### Not Yet Implemented

- Module system
- Struct field access
- Spans (borrowed ranges)

## Performance

ual occupies the same performance tier as C and Go:

```
Speed Scale (log):

     C =========|==========|
       Go ======|==========|
         ual ===|==========|
                |          |          | Python ==============
           10ns      100ns      1μs        10μs       100μs
```

Compute blocks have ~33ns fixed overhead per invocation. For computations >1μs, overhead is <5%.

## Project Structure

```
ual/
├── Makefile                 # Build system
├── build.sh                 # Alternative build script
├── cmd/ual/                 # Compiler source
│   ├── lexer.go
│   ├── parser.go
│   ├── codegen.go
│   └── main.go
├── examples/                # ual examples (60+)
├── benchmarks/              # Performance tests
│   ├── c/                   # C reference
│   ├── python/              # Python reference
│   └── *.go                 # Go benchmarks
├── stack.go                 # Core stack implementation
├── view.go                  # Decoupled views
├── bring.go                 # Atomic transfer
├── walk.go                  # Traversal operations
├── worksteal.go             # Work-stealing
├── MANUAL.md                # Comprehensive manual
├── CHANGELOG.md             # Version history
└── VERSION                  # Current version
```

## Documentation

| Document | Description |
|----------|-------------|
| `MANUAL.md` | Comprehensive language manual |
| `CHANGELOG.md` | Version history with examples |
| `COMPUTE_SPEC_V2.md` | Compute block specification |
| `benchmarks/RESULTS.md` | Performance analysis |

## Prerequisites

Go 1.22 or later must be installed. This is the only dependency.

```bash
# Check Go installation
go version
```

## Building

```bash
# Using Make (recommended)
make              # Build compiler
make test         # Run all tests
make install      # Install to $GOPATH/bin
make clean        # Remove build artifacts

# Using build.sh
./build.sh              # Build compiler
./build.sh --test       # Build and test
./build.sh --install    # Install to $GOPATH/bin
./build.sh --all        # Clean, build, test, install
```

## Usage

```bash
ual compile <file.ual>      # Compile to Go source (.go)
ual build <file.ual>        # Compile to executable binary
ual run <file.ual>          # Compile and run immediately
ual tokens <file.ual>       # Show lexer tokens
ual ast <file.ual>          # Show parse tree
ual version                 # Show version
ual help                    # Show help

# Options
-o, --output <path>         # Specify output file
-q, --quiet                 # Suppress non-error output
-v, --verbose               # Show detailed progress
-vv, --debug                # Show debug information
-O, --optimize              # Use optimized dstack
--version                   # Show version and exit
```

## Examples

```bash
# Compile to Go source
ual compile examples/01_fibonacci.ual
# Creates: examples/01_fibonacci.go

# Build executable
ual build examples/01_fibonacci.ual
# Creates: 01_fibonacci (binary)

# Run directly
ual run examples/01_fibonacci.ual
# Output: 4181

# Run quietly (only program output)
ual -q run examples/01_fibonacci.ual

# Build with custom name
ual build -o fib examples/01_fibonacci.ual
./fib
```

## Design Philosophy

ual is a **systems language disguised as a scripting language**:

- **Surface**: Concise, Forth-like syntax for rapid prototyping
- **Semantics**: Explicit data flow, deterministic scheduling, predictable memory
- **Performance**: Native codegen where it matters (compute blocks)
- **Target**: Orchestration logic, data pipelines, embedded computation

> "Write high-level orchestration without losing low-level control."

## Authors
Copyright (C)2025 haitch

h@ual.fi

https://oldbytes.space/@haitchfive


## License

Apache 2.0

https://www.apache.org/licenses/LICENSE-2.0

---

*ual v0.7.2*
