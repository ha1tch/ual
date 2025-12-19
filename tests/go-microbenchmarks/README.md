# Go Microbenchmarks

Go benchmarks measuring compute block codegen quality and runtime overhead at the nanosecond level.

## Purpose

These benchmarks answer: **"Is the generated code efficient?"**

They compare ual-style code patterns against idiomatic Go to verify that compilation introduces no performance penalty.

## Running

```bash
# From project root
make bench-micro

# Or directly
cd tests/go-microbenchmarks
go test -bench=. -benchmem

# Specific categories
go test -bench="BenchmarkCompute_" -benchmem    # Codegen quality
go test -bench="BenchmarkPipeline_" -benchmem   # Full pattern
go test -bench="BenchmarkOverhead_" -benchmem   # Isolated costs
```

## Benchmark Categories

### Compute Benchmarks (`BenchmarkCompute_*`)

Pure computation comparison — no stack operations, just the generated code vs handwritten Go.

| Algorithm | Go (ns) | ual (ns) | Ratio |
|-----------|---------|----------|-------|
| Mandelbrot | 4,165 | 4,173 | 1.00x |
| Leibniz | 119,870 | 119,829 | 1.00x |
| Newton | 7.99 | 9.73 | 1.22x |
| ArraySum | 34.6 | 34.6 | 1.00x |
| DPFib | 60.0 | 59.4 | 0.99x |

**Verdict:** ual-generated code matches handwritten Go within noise.

### Pipeline Benchmarks (`BenchmarkPipeline_*`)

Full ual pattern: lock stack → convert bytes → compute → convert back → unlock.

| Algorithm | Time (ns) | Overhead vs Compute |
|-----------|-----------|---------------------|
| Mandelbrot | 4,300 | +3% |
| Leibniz | 145,691 | +22% |
| Newton | 129.7 | +16x (short computation) |

**Insight:** Overhead is negligible for computations >1μs, significant for <100ns.

### Overhead Benchmarks (`BenchmarkOverhead_*`)

Isolated cost measurements:

| Operation | Time (ns) | Notes |
|-----------|-----------|-------|
| Lock/Unlock | 13.2 | sync.Mutex overhead |
| Byte Convert | 0.33 | Float↔bytes, negligible |
| Push/Pop Raw | 3.3 | Slice operations |
| Full Cycle | 34.2 | Complete 1-value roundtrip |

## Relationship to End-to-End Benchmarks

| Location | Measures | Unit |
|----------|----------|------|
| `tests/go-microbenchmarks/` (here) | Codegen quality, overhead | ns/op |
| `tests/benchmarks/` | Cross-backend, cross-language | ms total |

Use this directory for **performance analysis**. Use `tests/benchmarks/` for **regression testing**.

## Files

| File | Description |
|------|-------------|
| `compute_bench_test.go` | Codegen comparison benchmarks |
| `pipeline_bench_test.go` | Full pattern + overhead benchmarks |
| `c/c_bench.c` | C reference implementations |
| `RESULTS.md` | Detailed analysis |
