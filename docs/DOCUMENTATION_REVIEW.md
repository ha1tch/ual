# UAL Documentation Review — December 2025

## Overview

This document identifies gaps in the current UAL documentation following the Phase 1 and Phase 2 runtime unification work completed on 2025-12-13.

### Recent Major Changes Not Documented

1. **Interpreter (iual)** — A complete tree-walking interpreter now exists at `cmd/iual/`
2. **Runtime Unification** — `pkg/runtime/` now contains shared types (Value, ValueStack, ScopeStack)
3. **Interpreter Concurrency** — Interpreter now uses true goroutines matching compiler semantics
4. **Package Restructure** — Shared packages at `pkg/ast/`, `pkg/lexer/`, `pkg/parser/`, `pkg/runtime/`, `pkg/version/`
5. **71/71 Tests Pass** — Both compiler and interpreter pass all example tests

---

## Document-by-Document Analysis

### 1. README.md

**Current State:** Describes compiler only, outdated project structure

**Required Updates:**

#### Section: Quick Start
Add interpreter usage:
```bash
# Using the interpreter (for development)
./iual examples/001_fibonacci.ual

# Using the compiler (for production)
./ual run examples/001_fibonacci.ual
```

#### Section: Project Structure
Replace current structure with:
```
ual/
├── cmd/
│   ├── ual/                 # Compiler
│   │   ├── main.go
│   │   └── codegen.go
│   └── iual/                # Interpreter
│       ├── main.go
│       ├── interp.go
│       ├── interp_control.go
│       └── interp_expr.go
├── pkg/                     # Shared packages
│   ├── ast/                 # Abstract syntax tree
│   ├── lexer/               # Lexical analysis
│   ├── parser/              # Parser
│   ├── runtime/             # Stack, Value, Views
│   └── version/             # Version management
├── examples/                # 71 example programs
├── benchmarks/              # Performance tests
├── Makefile
├── MANUAL.md
├── CHANGELOG.md
└── DESIGN_v0.8.md
```

Remove references to root-level `stack.go`, `view.go`, `bring.go`, `walk.go`, `worksteal.go` — these are now in `pkg/runtime/`.

#### Section: Usage
Add interpreter commands:
```bash
# Compiler commands
ual compile <file.ual>      # Compile to Go source
ual build <file.ual>        # Compile to binary
ual run <file.ual>          # Compile and run

# Interpreter commands
iual <file.ual>             # Interpret directly
iual run <file.ual>         # Same as above
iual --trace <file.ual>     # Trace execution
```

#### Section: When to Use Which
Add new section:
```markdown
## Compiler vs Interpreter

| Aspect | ual (compiler) | iual (interpreter) |
|--------|----------------|-------------------|
| Speed | Native Go performance | 10-50x slower |
| Startup | Compile + run | Immediate |
| Debugging | Limited | --trace flag |
| Use case | Production | Development, testing |

Both tools produce identical results — they share the same runtime
types and concurrency model.
```

---

### 2. MANUAL.md

**Current State:** Comprehensive language reference, no interpreter mention

**Required Updates:**

#### Section: Installation
Add:
```markdown
## Building

```bash
# Build both compiler and interpreter
make build

# Or individually
go build -o ual ./cmd/ual
go build -o iual ./cmd/iual
```

#### Section: Usage
Add interpreter section after compiler usage:
```markdown
## Interpreter (iual)

The interpreter runs UAL programs directly without compilation.
Useful for development and testing.

```bash
iual program.ual            # Run program
iual --trace program.ual    # Trace execution
iual -q program.ual         # Quiet mode
```

**Performance Note:** The interpreter is approximately 10-50x slower
than compiled code. For production workloads, use `ual build`.
```

---

### 3. CHANGELOG.md

**Current State:** Last entry is 0.7.3 from 2025-12-12

**Required Updates:**

Add new section at top:
```markdown
## [0.7.3] - 2025-12-13 (Runtime Unification)

### Added

**Interpreter (iual)**

The UAL interpreter is now a first-class tool with full feature parity:

- Tree-walking interpreter at `cmd/iual/`
- True goroutine-based concurrency (matches compiler semantics)
- Shared runtime types with compiler via `pkg/runtime/`
- All 71 example programs pass in both compiler and interpreter
- `--trace` flag for execution tracing
- Quiet mode (`-q`) for scripting

```bash
iual examples/001_fibonacci.ual    # Run with interpreter
iual --trace examples/022_pipeline.ual  # Trace concurrent execution
```

**Runtime Package (pkg/runtime/)**

Unified runtime types shared between compiler and interpreter:

- `Value` — Type-safe value wrapper (int64, float64, string, bytes, bool)
- `ValueStack` — Stack of Values with type checking
- `ScopeStack` — Variable scope management
- Existing types: `Stack`, `View`, `Bring`, `Walk`

**Package Restructure**

Shared packages extracted to `pkg/`:

- `pkg/ast/` — Abstract syntax tree definitions
- `pkg/lexer/` — Lexical analysis
- `pkg/parser/` — Parser
- `pkg/runtime/` — Stack, Value, Views, Walk, Bring
- `pkg/version/` — Version management

### Changed

**Interpreter Concurrency**

The interpreter now uses true goroutines for `@spawn pop play`:

```ual
@spawn < { producer() }
@spawn < { consumer() }
@spawn pop play    -- Now launches real goroutine
@spawn pop play    -- Now launches real goroutine
```

Previously, spawn tasks ran synchronously. Now they run concurrently,
matching the compiler's `go _task()` semantics.

**Blocking Take**

The `take` operation now blocks correctly in the interpreter, waiting
for concurrent producers:

```ual
@data take:value   -- Blocks until value available
```

### Fixed

- Interpreter `take` no longer uses artificial timeouts
- Variable scoping in interpreter matches compiler
- Stack perspective changes work in interpreter

### Tests

- 71/71 example programs pass with compiler
- 71/71 example programs pass with interpreter
- Runtime package has comprehensive unit tests
```

---

### 4. DESIGN_v0.8.md

**Current State:** References v0.7.2

**Required Updates:**

#### Header
Change:
```markdown
**Document Purpose:** This specification captures design decisions made for UAL v0.8, intended for implementors continuing development. It assumes familiarity with UAL v0.7.3 architecture...
```

#### Section 10: References
Update:
```markdown
## 10. References

- UAL v0.7.3 source: `cmd/ual/` (compiler), `cmd/iual/` (interpreter)
- Shared packages: `pkg/ast/`, `pkg/lexer/`, `pkg/parser/`, `pkg/runtime/`
- Runtime: `pkg/runtime/` (Stack, Value, View, Bring, Walk)
- Existing constructs: `MANUAL.md`, `COMPUTE_SPEC_V2.md`
- Examples: `examples/` directory (71 programs)
- Benchmarks: `benchmarks/` directory
```

#### Add Note After Section 7
```markdown
### Implementation Note (v0.7.3)

As of v0.7.3, the interpreter has achieved parity with the compiler:

- Both use `pkg/runtime` types
- Both execute spawned tasks as goroutines
- Both pass all 71 example tests

This provides a foundation for implementing v0.8 features in both tools
simultaneously.
```

---

### 5. BENCHMARKS.md

**Current State:** Brief overview

**Required Updates:**

Add interpreter benchmark comparison:
```markdown
## Compiler vs Interpreter Performance

| Algorithm | ual (compiled) | iual (interpreted) | Ratio |
|-----------|----------------|-------------------|-------|
| Fibonacci | 61 ns | ~3,000 ns | ~50x |
| Mandelbrot | 4,170 ns | ~200,000 ns | ~48x |
| Pipeline | 920 μs | ~12,000 μs | ~13x |

The interpreter is suitable for:
- Development and testing
- Short-lived scripts
- Programs where startup time dominates

Use the compiler for:
- Production workloads
- Performance-critical code
- Long-running computations
```

---

### 6. New Document: ARCHITECTURE.md (Proposed)

Create new architecture document:

```markdown
# UAL Architecture

## Package Structure

```
github.com/ha1tch/ual/
├── cmd/
│   ├── ual/         # Compiler (Go codegen)
│   └── iual/        # Interpreter (tree-walking)
└── pkg/
    ├── ast/         # Abstract syntax tree
    ├── lexer/       # Tokenisation
    ├── parser/      # Parsing
    ├── runtime/     # Shared runtime
    └── version/     # Version info
```

## Compilation Pipeline

```
Source (.ual)
    │
    ▼
┌─────────┐
│  Lexer  │ pkg/lexer
└────┬────┘
     │ tokens
     ▼
┌─────────┐
│ Parser  │ pkg/parser
└────┬────┘
     │ AST
     ▼
┌─────────────────────────────────┐
│                                 │
│  ┌─────────┐    ┌───────────┐  │
│  │ CodeGen │    │ Interpret │  │
│  │cmd/ual  │    │ cmd/iual  │  │
│  └────┬────┘    └─────┬─────┘  │
│       │               │        │
│       ▼               ▼        │
│   Go Source      Direct Exec   │
│       │                        │
│       ▼                        │
│   go build                     │
│       │                        │
│       ▼                        │
│    Binary                      │
└─────────────────────────────────┘
```

## Runtime Types (pkg/runtime)

### Stack
Core container with four perspectives:
- LIFO (stack)
- FIFO (queue)
- Indexed (array)
- Hash (map)

### Value
Type-safe wrapper:
- Int64, Float64, String, Bytes, Bool
- Runtime type checking
- Conversion methods

### ValueStack
Stack of Values with type enforcement.

### ScopeStack
Variable scope management for interpreter.

## Concurrency Model

Both compiler and interpreter use Go goroutines:

```
@spawn < { task1() }
@spawn < { task2() }
@spawn pop play  →  go task()
@spawn pop play  →  go task()
```

The `take` operation blocks waiting for data from concurrent producers.
```

---

## Makefile Updates

Current Makefile may need updates:

```makefile
# Add interpreter targets
.PHONY: build-interpreter
build-interpreter:
	go build -o iual ./cmd/iual

.PHONY: test-interpreter
test-interpreter: build-interpreter
	@for f in examples/*.ual; do \
		./iual -q "$$f" > /dev/null 2>&1 || echo "FAIL: $$f"; \
	done
	@echo "Interpreter tests complete"

.PHONY: test-all
test-all: test test-interpreter
	@echo "All tests complete"
```

---

## Summary of Required Actions

| Document | Priority | Changes Required |
|----------|----------|------------------|
| CHANGELOG.md | High | Add 0.7.3 runtime unification entry |
| README.md | High | Add interpreter, update structure |
| MANUAL.md | Medium | Add interpreter usage section |
| DESIGN_v0.8.md | Low | Update version references |
| BENCHMARKS.md | Low | Add compiler vs interpreter comparison |
| ARCHITECTURE.md | Medium | Create new document |
| Makefile | Medium | Add interpreter targets |

---

## Validation Checklist

After updates, verify:

- [ ] `make build` builds both ual and iual
- [ ] `make test` runs all tests
- [ ] All 71 examples work with both tools
- [ ] README quick start works for new users
- [ ] CHANGELOG accurately reflects changes
- [ ] Project structure in README matches reality

---

*Review completed: 2025-12-13*
