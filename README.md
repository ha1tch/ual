# ual

**Version 0.7.4**

A coordination-first language for orchestration and embedded computation, presented with a scripting-style surface.

**ual** comes in three forms: an **interpreter** for development and exploration, and **two compilers** (Go and Rust backends) for production. All three share the same semantics and produce identical results — write once, run on any path.

## Documentation

- [MANUAL.md](docs/MANUAL.md) — Language reference
- [CONCURRENCY.md](docs/CONCURRENCY.md) — Spawn, select, synchronisation patterns
- [PERFORMANCE.md](docs/PERFORMANCE.md) — Benchmarks and methodology

## Philosophy

**ual** is built on a specific philosophical foundation: that coordination is the primary problem of programming, and computation is a subordinate activity within coordinated contexts.

- [Part One: Why Philosophy Matters for Language Design](ual-philosophy-01.md)
- [Part Two: The Ground and What Emerges](ual-philosophy-02.md)
- [Part Three: Boundaries, Time, and Acknowledgment](ual-philosophy-03.md)
- [Part Four: Coordination Precedes Computation](ual-philosophy-04.md)

```
Surface Feel              Actual Semantics
─────────────────         ─────────────────────────────────
Forth-like stack ops      Explicit data flow, typed containers
Erlang-like select        Deterministic scheduling, structured
Inline DSL blocks         Native codegen, zero overhead
```

## Quick Start

```bash
# Build the compiler and interpreter
make build

# Run with the interpreter (development)
./iual examples/001_fibonacci.ual

# Run with the compiler (production)
./ual run examples/001_fibonacci.ual

# Run with Rust backend (smallest binaries)
./ual run --target rust examples/001_fibonacci.ual
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
    while (i <= n) {
        result = result * i
        i = i + 1
    }
    return result
})

@numbers dot
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

The `.compute()` block is ual's "optimisation island" — arithmetic runs on native CPU types with zero serialisation overhead:

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

**Performance**: Compute blocks compile to native loops. Benchmarks show compiled ual matches C performance, and the interpreter beats Python by 2-20x:

| Benchmark | C | ual-Go | ual-Rust | iual | Python |
|-----------|---|--------|----------|------|--------|
| Leibniz π (1M) | 7-10ms | 10-11ms | 11-12ms | 37-47ms | 89-229ms |
| Mandelbrot 50×50 | 7-8ms | 9-10ms | 9-10ms | 11-13ms | 44-187ms |
| Newton sqrt ×1000 | 7ms | 8-12ms | 11ms | 9ms | 39-183ms |

*Ranges show M1 and Xeon results. All implementations compute identical workloads with verified matching outputs.*

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

### Implemented (v0.7.4)

- **Four perspectives**: LIFO, FIFO, Indexed, Hash
- **Compute blocks**: Threaded code compilation (interpreter), native codegen (compiler)
- **Container access**: `self.property` (Hash), `self[i]` (Indexed)
- **Select construct**: Multi-stack waiting with timeouts
- **Consider construct**: Structured error handling with status
- **Forth-style operations**: push, pop, dup, swap, rot, over, nip, tuck, arithmetic
- **Type checking**: Compile-time type validation
- **Control flow**: if/elseif/else, while, for, break, continue
- **Functions**: With typed parameters and returns
- **Spawn**: Goroutine-based concurrency with per-task operational stacks
- **Views**: Decoupled perspectives on shared data
- **Work-stealing**: LIFO owner + FIFO thieves pattern
- **Bring**: Atomic transfer with type conversion
- **Three backends**: Interpreter (iual), Go compiler, Rust compiler — 100% output parity
- **Build profiles**: `--small`, `--strip`, `--release`

### Not Yet Implemented

- Module system
- Struct types
- Spans (borrowed ranges)

## How It Works

### The Compiler

The ual compiler doesn't target machine code directly. Instead, it generates **Go or Rust source code**, then invokes the respective toolchain:

```
┌─────────┐     ┌─────────────┐     ┌──────────┐     ┌──────────┐
│ .ual    │ ──▶ │ ual compiler│ ──▶ │ .go/.rs  │ ──▶ │ binary   │
│ source  │     │ (codegen)   │     │ source   │     │          │
└─────────┘     └─────────────┘     └──────────┘     └──────────┘
                                          │
                                          ▼
                                    go build / rustc
```

**Why generate source code rather than bytecode or machine code?**

1. **Mature optimisation** — Go and Rust compilers have years of optimisation work. ual gets this for free.
2. **Native concurrency** — Go's goroutines and Rust's threads map directly to ual's spawn model.
3. **Debuggable output** — The generated code is readable. When something goes wrong, you can inspect it.
4. **No runtime to ship** — The compiled binary is self-contained.

The generated code is deliberately straightforward. Compute blocks become native loops. Stack operations become slice operations. Spawn blocks become goroutines (Go) or `std::thread::spawn` (Rust).

### The Rust Runtime (rual)

For the Rust backend, ual ships with `rual/` — a runtime library providing:

- `Stack<T>` with perspective switching (LIFO/FIFO/Indexed/Hash)
- Thread-safe operations with `Mutex` and `Condvar` for blocking `take`
- `View` for decoupled perspectives on shared data
- Work-stealing primitives

The compiler generates Rust code that links against `rual`. The library is ~1,200 lines of Rust and compiles to ~50KB in the final binary.

### The Interpreter (iual)

The interpreter takes a different approach. It shares the lexer, parser, and AST with the compiler (via `pkg/`), but executes directly rather than generating code.

For most constructs, `iual` uses **tree-walking interpretation** — it traverses the AST and executes nodes. This is simple and correct, but slow for tight loops.

For **compute blocks**, `iual` uses **threaded code compilation**:

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ Compute AST │ ──▶ │ Compile to      │ ──▶ │ []func(*Env)    │
│             │     │ closures        │     │ (cached)        │
└─────────────┘     └─────────────────┘     └─────────────────┘
                                                   │
                           ┌───────────────────────┘
                           ▼
              Execute closures in sequence
              (no AST dispatch, direct slot access)
```

On first execution, the compute block's AST is compiled to a slice of closures. Variables become direct slot access (`env.floats[3]`) rather than map lookups. The compiled form is cached, so subsequent invocations skip compilation entirely.

This is why `iual` achieves 4-13x faster performance than Python on numeric workloads — the hot path avoids interpretation overhead entirely.

### Shared Infrastructure

Both compiler and interpreter use the same packages:

| Package | Purpose |
|---------|---------|
| `pkg/lexer` | Tokenisation |
| `pkg/parser` | AST construction |
| `pkg/ast` | Node definitions |
| `pkg/runtime` | Stack, Value, View, Scope (interpreter only) |

This shared infrastructure ensures the compiler and interpreter agree on syntax and semantics. The 92 correctness tests verify identical output across all three backends.

## Performance

Three performance tiers emerge from cross-platform benchmarks (Apple M1 and Intel Xeon):

**1. Compiled ual vs C:**
| Benchmark | ual-Go / C | ual-Rust / C |
|-----------|------------|--------------|
| Leibniz | 1.0-1.1x | 1.1-1.2x |
| Mandelbrot | 1.1-1.4x | 1.1-1.3x |
| Newton | 1.1-1.7x | 1.4-1.6x |

Compiled ual is within 1.0-1.7x of C — the overhead is Go/Rust runtime characteristics, not ual abstractions.

**2. iual interpreter vs compiled:**
| Benchmark | iual / ual-Go |
|-----------|---------------|
| Leibniz | 3.7-4.7x slower |
| Mandelbrot | 1.1-1.4x slower |
| Newton | 0.75-1.1x (matches or beats!) |

The threaded code compiler makes iual competitive on structured loops.

**3. iual vs Python:**
| Benchmark | iual speedup |
|-----------|--------------|
| Leibniz | **1.9-6.2x faster** |
| Mandelbrot | **3.4-17x faster** |
| Newton | **4.3-20x faster** |

iual beats Python on every benchmark, on both platforms.

```
Performance tiers (range across benchmarks, both platforms):

        C |=====|                                       7-10ms
     Rust |=====|                                       7-11ms
   ual-Go |======|                                      8-12ms
 ual-Rust |======|                                      9-12ms
     iual |      |===========|                          9-47ms
   Python |                              |==============| 39-229ms
          0         25        50        100       150    200ms
```

Compiled ual belongs with C and Rust. The interpreter sits in its own tier — always faster than Python (2-20x), sometimes matching compiled.

## Project Structure

```
ual/
├── cmd/
│   ├── ual/                 # Compiler (Go + Rust backends)
│   │   ├── main.go
│   │   ├── codegen_go.go
│   │   └── codegen_rust.go
│   └── iual/                # Interpreter
│       ├── main.go
│       ├── interp.go
│       ├── interp_control.go
│       ├── interp_expr.go
│       └── compute_compile.go  # Threaded code compiler
├── pkg/                     # Shared packages
│   ├── ast/                 # Abstract syntax tree
│   ├── lexer/               # Lexical analysis
│   ├── parser/              # Parser
│   ├── runtime/             # Stack, Value, Views
│   └── version/             # Version management
├── rual/                    # Rust runtime library
├── tests/
│   ├── correctness/         # 92 tests × 3 backends
│   ├── negative/            # Error detection tests
│   └── benchmarks/          # Cross-language benchmarks
├── docs/                    # Technical documentation
├── examples/                # 92 example programs
└── ual-philosophy-*.md      # Philosophy essays (4)
```

## Prerequisites

**Required (at least one):**
- Go 1.22+ (for `ual` compiler and `iual` interpreter)
- Rust 1.75+ (for `--target rust` backend)

```bash
go version       # Check Go
rustc --version  # Check Rust (optional)
```

## Building

```bash
make build     # Build compiler and interpreter
make test      # Run all tests
make benchmark # Run benchmarks
make clean     # Remove build artefacts
```

## Usage

### Compiler (ual)

```bash
ual run <file.ual>              # Compile and run (Go backend)
ual run --target rust <file>    # Compile and run (Rust backend)
ual build <file.ual>            # Build executable
ual build --small --target rust # Small Rust binary (~343KB)
ual compile <file.ual>          # Generate source only
ual tokens <file.ual>           # Show lexer tokens
ual ast <file.ual>              # Show parse tree

# Options
-o, --output <path>    # Specify output file
-q, --quiet            # Suppress non-error output
-v, --verbose          # Show detailed progress
```

### Interpreter (iual)

```bash
iual <file.ual>           # Interpret directly
iual --trace <file.ual>   # Trace execution
iual -q <file.ual>        # Quiet mode
```

### When to Use Which

| Scenario | Tool | Why |
|----------|------|-----|
| Development | `iual` | Instant feedback, `--trace` for debugging |
| Production | `ual` → Go | Native performance, standard deployment |
| Minimal footprint | `ual` → Rust | Smallest binaries (~343KB stripped) |
| Scripting | `iual` | 4-13x faster than Python, no compile step |

All three produce identical results — they share the same runtime semantics.

## Design Philosophy

ual is a **systems language disguised as a scripting language**:

- **Surface**: Concise, Forth-like syntax for rapid prototyping
- **Semantics**: Explicit data flow, deterministic scheduling, predictable memory
- **Performance**: Native codegen where it matters (compute blocks)
- **Target**: Orchestration logic, data pipelines, embedded computation

> "Write high-level orchestration without losing low-level control."

## Authors

Copyright (C) 2025 haitch

h@ual.fi

https://oldbytes.space/@haitchfive

## License

Apache 2.0 — https://www.apache.org/licenses/LICENSE-2.0

---

*ual v0.7.4 — Coordination-first programming*