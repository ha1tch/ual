# ual Benchmark Suite

This directory contains the unified benchmark infrastructure for ual.

## Quick Start

```bash
# Run full benchmarks with HTML report
./run_unified.sh

# Quick benchmarks (1 iteration, no cross-language)
./run_unified.sh --quick --backends

# JSON output only
./run_unified.sh --json
```

## Benchmark Types

### 1. ual Backend Benchmarks (`programs/`)

Tests all three ual execution modes on compute-heavy programs:

| Program | Description |
|---------|-------------|
| `bench_compute_leibniz.ual` | Leibniz π series (1M iterations) |
| `bench_compute_mandelbrot.ual` | Mandelbrot set (50×50 grid) |
| `bench_compute_newton.ual` | Newton's method sqrt (1000 numbers) |

**Backends tested:**
- **ual→Go**: Compile to Go, then native binary
- **ual→Rust**: Compile to Rust, then native binary  
- **iual**: Direct interpretation with threaded code

### 2. Cross-Language Benchmarks (`cross_language/`)

Compares ual performance against native implementations:

| Language | Location | Notes |
|----------|----------|-------|
| C | `cross_language/c/bench.c` | Compiled with `-O2` |
| Rust | `cross_language/rust/` | Compiled with `--release` |
| Python | `cross_language/python/python_bench.py` | CPython 3.x |

## Output

### JSON Results (`results/`)

Each run produces a timestamped JSON file:

```json
{
  "version": "0.7.4",
  "timestamp": "2025-12-17T14:15:49+00:00",
  "iterations": 5,
  "correctness": {"total": 79, "go_pass": 79, "rust_pass": 79, "iual_pass": 79},
  "benchmarks": [
    {"name": "compute_leibniz", "go_ms": 12, "rust_ms": 11, "iual_ms": 50}
  ],
  "cross_language": {
    "c": {"leibniz": 12, "mandelbrot": 12, "newton": 10},
    "python": {"leibniz": 196, "mandelbrot": 172, "newton": 44}
  },
  "binary_sizes": {"go_stripped": 1261720, "rust_stripped": 407968, "iual": 3034771}
}
```

### HTML Reports (`reports/`)

Visual reports with charts are generated automatically:

- `reports/latest.html` — Most recent report
- `reports/report_YYYYMMDD_HHMMSS.html` — Timestamped archive

## Options

```
--quick         1 iteration (fast, for CI)
--full          5 iterations (default, for accuracy)
--backends      ual backends only (Go, Rust, iual)
--cross-lang    Cross-language comparison only
--all           Everything (default)
--no-html       Skip HTML report generation
--json          Output JSON to stdout
```

## Integration with Makefile

```bash
# From project root
make benchmark           # Full benchmark suite
make benchmark-quick     # Quick benchmark for CI
```

## Performance Baselines

The `BASELINE.json` file contains reference performance numbers. The benchmark suite compares against this baseline and flags regressions (>10% slower).

## Directory Structure

```
tests/benchmarks/
├── run_unified.sh          # Main benchmark runner
├── generate_report.py      # HTML report generator
├── BASELINE.json           # Reference performance
├── programs/               # ual benchmark programs
│   ├── bench_compute_leibniz.ual
│   ├── bench_compute_mandelbrot.ual
│   └── bench_compute_newton.ual
├── cross_language/         # Native implementations
│   ├── c/bench.c
│   ├── rust/
│   └── python/python_bench.py
├── results/                # JSON output
│   └── latest.json
└── reports/                # HTML reports
    └── latest.html
```

## See Also

- `/benchmarks/` — Go microbenchmarks for compute block efficiency
- `/docs/BENCHMARKS.md` — Performance analysis and methodology
