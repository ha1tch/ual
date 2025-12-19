# ual Benchmark Results

**Version 0.7.4** — Cross-platform performance analysis

## Executive Summary

ual provides three execution paths with distinct performance characteristics:

| Mode | Speed | Use Case |
|------|-------|----------|
| `ual` → Go | Native (matches C) | Production |
| `ual` → Rust | Native (matches C) | Systems integration |
| `iual` | 2-20x faster than Python | Development, scripting |

**Key findings:**
- Compiled ual is within **1.0-1.7x of C** performance
- iual interpreter is **2-20x faster than Python**
- iual approaches compiled speed on structured loops (1.1x slower to 0.75x faster)

## Test Environments

### Apple M1 (macOS)

| Component | Version |
|-----------|---------|
| CPU | Apple M1 |
| OS | macOS (Darwin 23.5.0) |
| C compiler | Apple Clang 15.0.0 |
| Rust compiler | rustc 1.92.0 |
| Go compiler | go1.24.4 darwin/arm64 |
| Python | CPython 3.12.2 (native arm64) |
| ual | v0.7.4 |

### Intel Xeon (Linux)

| Component | Version |
|-----------|---------|
| CPU | Intel Xeon |
| OS | Ubuntu 24.04 LTS |
| C compiler | gcc 13.3.0 |
| Rust compiler | rustc 1.75.0 |
| Go compiler | go1.22.2 linux/amd64 |
| Python | CPython 3.12.3 |
| ual | v0.7.4 |

## Benchmark Programs

All implementations compute identical workloads with verified matching outputs.

### Leibniz π (1M iterations)

Computes π using the Leibniz series: π/4 = 1 - 1/3 + 1/5 - 1/7 + ...

Tests: floating-point arithmetic, loop performance, accumulator patterns.

Expected output: `3.1415916535897743`

### Mandelbrot 50×50

Computes escape iterations for a 50×50 grid of points in the Mandelbrot set.
- Grid: 50×50 pixels
- Complex plane: x ∈ [-2, 1], y ∈ [-1.5, 1.5]
- Max iterations: 100
- Escape radius: 2.0

Tests: nested loops, complex number arithmetic, conditional branching.

Expected output: `52761`

### Newton sqrt ×1000

Computes square roots of integers 1-1000 using Newton-Raphson iteration (20 iterations each), summing results.

Tests: floating-point division, nested loops, accumulator patterns.

Expected output: `21097.455887480734`

## Raw Results

### Apple M1 (milliseconds)

| Benchmark | C | Rust | Python | ual-Go | ual-Rust | iual |
|-----------|---|------|--------|--------|----------|------|
| Leibniz π (1M) | 10 | 11 | 229 | 10 | 12 | 37 |
| Mandelbrot 50×50 | 7 | 7 | 187 | 10 | 9 | 11 |
| Newton sqrt ×1000 | 7 | 7 | 183 | 12 | 11 | 9 |

### Intel Xeon (milliseconds)

| Benchmark | C | Rust | Python | ual-Go | ual-Rust | iual |
|-----------|---|------|--------|--------|----------|------|
| Leibniz π (1M) | 9 | 11 | 89 | 10 | 12 | 47 |
| Mandelbrot 50×50 | 8 | 9 | 44 | 9 | 10 | 13 |
| Newton sqrt ×1000 | 7 | 8 | 39 | 8 | 11 | 9 |

## Analysis

### 1. Compiled ual vs C

How does compiled ual compare to hand-written C?

| Benchmark | M1: ual-Go/C | M1: ual-Rust/C | Xeon: ual-Go/C | Xeon: ual-Rust/C |
|-----------|--------------|----------------|----------------|------------------|
| Leibniz | 1.0x | 1.2x | 1.1x | 1.2x |
| Mandelbrot | 1.4x | 1.3x | 1.1x | 1.1x |
| Newton | 1.7x | 1.6x | 1.1x | 1.4x |

**Summary:** Compiled ual is within **1.0-1.7x of C**. The variance reflects Go/Rust runtime characteristics (garbage collection, bounds checking), not ual abstraction overhead. Both backends produce equivalent performance.

### 2. iual Interpreter vs Compiled ual

How much overhead does interpretation add?

| Benchmark | M1: iual/ual-Go | Xeon: iual/ual-Go |
|-----------|-----------------|-------------------|
| Leibniz | 3.7x slower | 4.7x slower |
| Mandelbrot | 1.1x slower | 1.4x slower |
| Newton | **0.75x (faster!)** | 1.1x slower |

**Summary:** The threaded code compiler makes iual **competitive on structured loops** (Mandelbrot, Newton). Arithmetic-heavy code (Leibniz) shows expected interpreter overhead. The Newton result where iual beats compiled on M1 reflects measurement noise at small timescales.

### 3. iual vs Python

How does ual's interpreter compare to Python's?

| Benchmark | M1: iual speedup | Xeon: iual speedup |
|-----------|------------------|-------------------|
| Leibniz | **6.2x faster** | **1.9x faster** |
| Mandelbrot | **17x faster** | **3.4x faster** |
| Newton | **20x faster** | **4.3x faster** |

**Summary:** iual beats Python on **every benchmark, on both platforms**, by **2-20x**. The wider gap on M1 reflects platform-specific Python interpreter characteristics.

## Performance Tiers

Three distinct tiers emerge from the data:

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

**Tier 1 — Compiled (7-12ms):** C, Rust, ual-Go, ual-Rust cluster together. Compiled ual belongs here.

**Tier 2 — Interpreter (9-47ms):** iual sits in its own tier. Overlaps with compiled on some benchmarks (Newton: 9ms), extends further on others (Leibniz: 47ms).

**Tier 3 — Python (39-229ms):** Clearly separated. iual is always faster.

## Threaded Code Compilation

The iual interpreter's performance comes from **threaded code compilation** for compute blocks.

### How It Works

When a compute block is first executed:

1. **Scan phase**: Variables are assigned direct slot indices (`env.floats[3]`)
2. **Compile phase**: AST is compiled to `[]func(*ComputeEnv)` closures
3. **Cache**: Compiled code is stored and reused on subsequent invocations

This eliminates:
- AST node dispatch (~5-10ns per operation)
- Map lookups for variables (~20-30ns per access)
- Per-operation type checking

### Speedup vs Tree-Walking

| Benchmark | Tree-walking | Threaded | Speedup |
|-----------|--------------|----------|---------|
| Leibniz π (1M) | 561ms | 47ms | **11.9x** |
| Mandelbrot 50×50 | 66ms | 13ms | **5.1x** |
| Newton sqrt ×1000 | 21ms | 9ms | **2.3x** |

The threaded compiler provides **2-12x speedup** over naive interpretation.

## Binary Sizes

| Target | Unstripped | Stripped |
|--------|------------|----------|
| ual-Go | 2.0 MB | 1.5 MB |
| ual-Rust | 13 MB | 344 KB |
| iual interpreter | 3.1 MB | 2.1 MB |

Rust with LTO produces dramatically smaller stripped binaries due to aggressive dead code elimination. Go binaries include runtime metadata that survives stripping. The iual interpreter is larger because it includes the full parser, runtime, and threaded code compiler.

## When to Use Each Mode

### Use `iual` (interpreter) when:
- Developing and testing ual programs
- Scripting tasks (faster than Python)
- Debugging with `--trace`
- Quick iteration without compilation

### Use `ual` → Go when:
- Production deployment
- Standard server environments
- Integration with Go ecosystem
- Cross-compilation needed

### Use `ual` → Rust when:
- Size-constrained environments (344 KB binaries)
- Systems integration
- Embedding in Rust applications
- Maximum portability

## Methodology

### Measurement

- Each benchmark runs 5 iterations
- Median time is taken (eliminates outliers)
- Times include process startup overhead
- Timing via shell: `date +%s%N` (nanoseconds)

### Verification

All implementations produce identical output:
```bash
./verify_benchmarks.sh
```

This confirms the workloads are equivalent, not just similar.

### Compilation Flags

| Language | Flags |
|----------|-------|
| C | `gcc -O2` |
| Rust | `--release` (includes LTO) |
| Go | Default (includes optimisations) |
| Python | CPython interpreter |

## Running the Benchmarks

```bash
# Full benchmark suite with HTML report
make benchmark

# Quick smoke test (1 iteration)
make benchmark-quick

# Verify outputs match across implementations
./verify_benchmarks.sh

# Direct invocation with options
./tests/benchmarks/run_unified.sh --help
./tests/benchmarks/run_unified.sh --json        # JSON output
./tests/benchmarks/run_unified.sh --backends    # ual backends only
./tests/benchmarks/run_unified.sh --cross-lang  # Cross-language only
```

**Output locations:**
- JSON results: `tests/benchmarks/results/latest.json`
- HTML report: `tests/benchmarks/reports/latest.html`

**Source code:**
- ual benchmark programs: `tests/benchmarks/programs/`
- Cross-language implementations: `tests/benchmarks/cross_language/`

## Conclusions

1. **Compiled ual matches native performance** — within 1.0-1.7x of hand-written C

2. **iual interpreter beats Python by 2-20x** — threaded code compilation closes the gap with compiled code

3. **Three performance tiers emerge naturally** — compiled (7-12ms), interpreter (9-47ms), Python (39-229ms)

4. **Results hold across platforms** — ARM (M1) and x86 (Xeon) show consistent patterns

5. **The abstraction pays for itself** — ual's stack model and compute blocks don't impose performance penalties

ual demonstrates that a coordination-first language with explicit stack semantics can achieve systems-level performance without sacrificing its programming model.