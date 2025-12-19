# ual Test Suite

This directory contains the testing infrastructure for ual v0.7.4+.

## Directory Structure

```
tests/
├── README.md                 # This file
├── correctness/              # Output correctness testing
│   ├── run_all.sh           # Master test runner
│   ├── expected/            # Expected outputs (92 files)
│   └── results/             # Test results (gitignored)
├── negative/                # Error detection tests
│   ├── parser/              # Parser error tests
│   └── runtime/             # Runtime error tests
├── benchmarks/              # End-to-end benchmarks (ms-level)
│   ├── run_unified.sh       # Main benchmark runner
│   ├── generate_report.py   # HTML report generator
│   ├── programs/            # ual benchmark programs
│   ├── cross_language/      # C, Python, Rust comparisons
│   ├── results/             # JSON benchmark results
│   └── reports/             # HTML benchmark reports
└── go-microbenchmarks/      # Go microbenchmarks (ns-level)
    ├── bench.sh             # Microbenchmark runner
    ├── compute_bench_test.go # Codegen quality benchmarks
    ├── pipeline_bench_test.go # Full pattern benchmarks
    └── c/                   # C reference implementations
```

## Quick Start

```bash
# Run all correctness tests (Go, Rust, iual)
./tests/correctness/run_all.sh --all

# Run negative tests (parser/runtime errors)
./tests/negative/run_negative_tests.sh

# Run unit tests
go test ./...

# Test specific backend
./tests/correctness/run_all.sh --go
./tests/correctness/run_all.sh --rust
./tests/correctness/run_all.sh --iual

# Show only failures
./tests/correctness/run_all.sh --all --quiet

# Show diffs for failures
./tests/correctness/run_all.sh --all --verbose

# Test single example
./tests/correctness/run_all.sh --example 041_compute_leibniz --verbose

# JSON output (for automation)
./tests/correctness/run_all.sh --all --json

# Update expected outputs (after intentional changes)
./tests/correctness/run_all.sh --update
```

## Using Make

```bash
make test              # Run all correctness tests
make test-go           # Go backend only
make test-rust         # Rust backend only
make test-iual         # iual interpreter only
make test-update       # Regenerate expected outputs
```

## Correctness Testing

The correctness test suite verifies that all three backends (Go, Rust, iual) produce identical output for all 92 example programs.

## Negative Testing

Negative tests verify that invalid programs produce appropriate errors.

```bash
./tests/negative/run_negative_tests.sh
```

### Parser Error Tests

| Test | Verifies |
|------|----------|
| `err_unclosed_brace` | Missing closing brace detection |
| `err_invalid_token` | Invalid token handling |
| `err_missing_paren` | Missing parenthesis detection |
| `err_compute_no_bindings` | Compute block syntax validation |
| `err_function_no_body` | Function body requirement |
| `err_while_no_condition` | While condition requirement |

### Runtime Error Tests

| Test | Verifies |
|------|----------|
| `err_type_mismatch` | Type safety on pop operations |
| `err_undefined_var` | Undefined variable detection |
| `err_undefined_func` | Undefined function detection |
| `err_undefined_stack` | Undefined stack detection |
| `err_array_bounds` | Array bounds checking |

## Unit Testing

Unit tests for specific modules:

```bash
# All unit tests
go test ./...

# Runtime package (40 tests)
go test -v ./pkg/runtime/

# Compute compiler (15 tests)
go test -v ./cmd/iual/
```

### Compute Compiler Tests (`cmd/iual/compute_compile_test.go`)

| Test | Verifies |
|------|----------|
| `TestNewComputeCompiler` | Constructor initialisation |
| `TestSlotAllocation` | Parameter slot allocation |
| `TestVarDeclSlots` | Variable slot allocation by type |
| `TestArrayDeclSlots` | Array slot and size tracking |
| `TestComputeEnvExecution` | Basic expression execution |
| `TestParameterBinding` | Parameter passing |
| `TestWhileLoop` | While loop compilation |
| `TestIfStatement` | If/else compilation |
| `TestArrayAccess` | Local array read/write |
| `TestMathFunctions` | Math function compilation |
| `TestBreakStatement` | Break statement handling |
| `TestArithmeticOps` | +, -, *, / operators |
| `TestComparisonOps` | <, >, <=, >=, ==, != operators |
| `TestUnaryMinus` | Unary minus compilation |
| `TestNestedExpressions` | Complex expression nesting |

### Expected Outputs

The `expected/` directory contains the canonical output for each example, generated from the Go backend. These files are version-controlled.

To update expected outputs after intentional changes:

```bash
./tests/correctness/run_all.sh --update
git add tests/correctness/expected/
git commit -m "Update expected outputs for <reason>"
```

### Test Results

Results are written to `results/` (gitignored) when using `--save`:

```bash
./tests/correctness/run_all.sh --all --save
```

### Exit Codes

- `0` - All tests passed
- `1` - One or more tests failed

### JSON Output Format

```json
{
  "timestamp": "2025-12-16T20:00:00+00:00",
  "total": 92,
  "backends": {
    "go": {"pass": 92, "fail": 0, "skip": 0},
    "rust": {"pass": 92, "fail": 0, "skip": 0},
    "iual": {"pass": 92, "fail": 0, "skip": 0}
  },
  "results": [
    {"name": "001_fibonacci", "go": "pass", "rust": "pass", "iual": "pass"},
    ...
  ]
}
```

## Benchmarks

The `benchmarks/` directory contains the unified benchmark infrastructure.

```bash
# Run full benchmarks with HTML report
./tests/benchmarks/run_unified.sh

# Quick benchmarks (1 iteration, faster)
./tests/benchmarks/run_unified.sh --quick --backends

# JSON output for CI/automation
./tests/benchmarks/run_unified.sh --json

# Cross-language comparison
./tests/benchmarks/run_unified.sh --cross-lang
```

### Benchmark Results (v0.7.4)

| Benchmark | C | Rust | Python | ual-Go | ual-Rust | iual |
|-----------|---|------|--------|--------|----------|------|
| Leibniz π (1M) | 12ms | 13ms | 196ms | 12ms | 11ms | 50ms |
| Mandelbrot 50×50 | 12ms | 14ms | 172ms | 10ms | 10ms | 13ms |
| Newton sqrt ×1000 | 10ms | 12ms | 44ms | 10ms | 10ms | 11ms |

**Key findings:**
- Compiled ual matches native C/Rust performance
- Interpreted ual (`iual`) is **4-13x faster than Python**
- `iual` achieves near-C on some benchmarks (Mandelbrot: 1.0x)

### Benchmark Programs

| File | Tests |
|------|-------|
| `bench_compute_leibniz.ual` | Pure floating-point arithmetic |
| `bench_compute_mandelbrot.ual` | Nested loops with early exit |
| `bench_compute_newton.ual` | Iterative convergence |

### Output

- **JSON results**: `tests/benchmarks/results/latest.json`
- **HTML reports**: `tests/benchmarks/reports/latest.html`

See `tests/benchmarks/README.md` for full documentation.

## Go Microbenchmarks

The `go-microbenchmarks/` directory contains nanosecond-level benchmarks for analysing codegen quality and overhead.

```bash
# Run all microbenchmarks
make bench-micro

# Or directly
cd tests/go-microbenchmarks
go test -bench=. -benchmem

# Specific categories
go test -bench="BenchmarkCompute_" -benchmem    # Codegen quality
go test -bench="BenchmarkPipeline_" -benchmem   # Full pattern
go test -bench="BenchmarkOverhead_" -benchmem   # Isolated costs
```

**Purpose:** These benchmarks verify that ual-generated code matches handwritten Go performance. Use for deep performance analysis; use `tests/benchmarks/` for regression testing.

See `tests/go-microbenchmarks/README.md` for full documentation.

## Adding New Tests

1. Create the ual program in `examples/NNN_name.ual`
2. Run `./tests/correctness/run_all.sh --update` to generate expected output
3. Verify with `./tests/correctness/run_all.sh --example NNN_name --verbose`
4. Commit both the example and expected output

## Troubleshooting

### "No expected outputs found"

Run `--update` first:
```bash
./tests/correctness/run_all.sh --update
```

### Rust tests skipped

Ensure Rust 1.75+ is installed:
```bash
rustc --version
```

### Test fails but output looks correct

Check for trailing whitespace or newline differences:
```bash
./tests/correctness/run_all.sh --example NAME --verbose
```

Or compare directly:
```bash
./ual -q run examples/NAME.ual > /tmp/actual.txt
diff tests/correctness/expected/NAME.txt /tmp/actual.txt
```
