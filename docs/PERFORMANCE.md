# ual Performance Summary

> **Full details:** See [BENCHMARKS.md](BENCHMARKS.md) for complete methodology, raw data, and analysis.

## Quick Reference

### Compiled ual vs C

| Benchmark | ual-Go / C | ual-Rust / C |
|-----------|------------|--------------|
| Leibniz | 1.0-1.1x | 1.1-1.2x |
| Mandelbrot | 1.1-1.4x | 1.1-1.3x |
| Newton | 1.1-1.7x | 1.4-1.6x |

**Compiled ual is within 1.0-1.7x of C.**

### iual vs Compiled

| Benchmark | iual / ual-Go |
|-----------|---------------|
| Leibniz | 3.7-4.7x slower |
| Mandelbrot | 1.1-1.4x slower |
| Newton | 0.75-1.1x (matches or beats) |

**Threaded code compilation makes iual competitive on structured loops.**

### iual vs Python

| Benchmark | iual speedup |
|-----------|--------------|
| Leibniz | 1.9-6.2x faster |
| Mandelbrot | 3.4-17x faster |
| Newton | 4.3-20x faster |

**iual beats Python on every benchmark, by 2-20x.**

## Performance Tiers

```
        C |=====|                                       7-10ms
     Rust |=====|                                       7-11ms
   ual-Go |======|                                      8-12ms
 ual-Rust |======|                                      9-12ms
     iual |      |===========|                          9-47ms
   Python |                              |==============| 39-229ms
          0         25        50        100       150    200ms
```

## Binary Sizes

| Target | Unstripped | Stripped |
|--------|------------|----------|
| ual-Go | 2.0 MB | 1.5 MB |
| ual-Rust | 13 MB | 344 KB |
| iual | 3.1 MB | 2.1 MB |

## Running Benchmarks

```bash
make benchmark              # Full suite with HTML report
./verify_benchmarks.sh      # Verify outputs match
```

See [BENCHMARKS.md](BENCHMARKS.md) for detailed methodology and platform-specific results.