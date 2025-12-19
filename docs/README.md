# ual Documentation

**Version 0.7.4**

## Execution Modes

ual provides three ways to run the same source code:

| Mode | Command | Use Case |
|------|---------|----------|
| Interpreted | `iual program.ual` | Development, scripting |
| Compiled (Go) | `ual run program.ual` | Production |
| Compiled (Rust) | `ual run --target rust program.ual` | Systems integration |

All three produce identical output. The interpreter uses threaded code compilation for compute blocks, achieving 4-13x faster performance than Python.

## Core Documentation

| Document | Description |
|----------|-------------|
| [MANUAL.md](MANUAL.md) | Comprehensive language manual |
| [CONCURRENCY.md](CONCURRENCY.md) | Concurrency model, spawn, synchronisation, patterns |
| [CHANGELOG.md](CHANGELOG.md) | Version history |
| [BENCHMARKS.md](BENCHMARKS.md) | Comprehensive benchmark results, methodology, and analysis |
| [PERFORMANCE.md](PERFORMANCE.md) | Quick reference performance summary |

## Specifications

| Document | Description |
|----------|-------------|
| [COMPUTE_SPEC_V2.md](COMPUTE_SPEC_V2.md) | Compute block specification (includes interpreter details) |
| [ERROR_PHILOSOPHY.md](ERROR_PHILOSOPHY.md) | Error handling philosophy |
| [DESIGN_v0.8.md](DESIGN_v0.8.md) | Design document for v0.8 |

## Backend Status

| Document | Description |
|----------|-------------|
| [RUST_BACKEND_STATUS.md](RUST_BACKEND_STATUS.md) | Rust backend (100% parity with Go) |

Both backends compile to native code and produce identical output across all 92 test programs.

## Performance Summary

Cross-language benchmarks (all times in milliseconds):

| Benchmark | C | Python | ual-Go | iual |
|-----------|---|--------|--------|------|
| Leibniz π (1M) | 12 | 196 | 12 | 50 |
| Mandelbrot 50×50 | 12 | 172 | 10 | 13 |
| Newton sqrt ×1000 | 10 | 44 | 10 | 11 |

Compiled ual matches C. Interpreted ual (`iual`) is 4-13x faster than Python.

---

*ual v0.7.4 — Coordination-first programming*