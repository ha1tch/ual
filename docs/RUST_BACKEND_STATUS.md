# ual Rust Backend Status

**Version:** 0.7.4  
**Date:** 2025-12-18  
**Status:** Production Ready (100% output equivalence)

## Current State

| Metric | Count | Percentage |
|--------|-------|------------|
| Codegen success | 92/92 | 100% |
| Rust compilation | 92/92 | 100% |
| Output equivalence (Go = Rust) | 92/92 | 100% |
| Output equivalence (Go = iual) | 92/92 | 100% |
| All three implementations match | 92/92 | 100% |

## Binary Size Comparison

| Build Profile | Go | Rust |
|---------------|-----|------|
| Default | 1.9M | 13M |
| Stripped | 1.3M | 403K |
| Small (`--small`) | 1.3M | 343K |

Rust with `--small` produces binaries ~4x smaller than Go.

## Usage

```bash
# Compile to Rust
ual compile --target rust program.ual -o program.rs

# Compile with size optimization
ual compile --target rust --small program.ual -o program.rs
```

## Implementation Notes

The Rust backend generates code that depends on the `rual` crate (located in the `rual/` directory), which provides:

- `Stack<T>` with perspectives (LIFO, FIFO, Indexed, Hash)
- `Value` for dynamic typing
- `View` for borrowed perspectives
- `BlockingStack<T>` for blocking operations
- Work-stealing primitives (`WSDeque`, `WSStack`)

All ual semantics including `consider`, `select`, `spawn`, `take`, and compute blocks are fully supported with output identical to the Go backend.
