#!/bin/bash
# ual Go Microbenchmarks
# 
# Usage: ./bench.sh [compute|pipeline|overhead|c|all]
#
# Categories:
#   compute  - Pure computation comparison (Go vs ual-style code)
#   pipeline - Full ual pattern (lock + bytes + compute)
#   overhead - Isolated overhead measurements
#   c        - C reference benchmarks
#   all      - Run everything (default)

set -e

cd "$(dirname "$0")"

MODE="${1:-all}"
COUNT="${2:-3}"

echo "=============================================================================="
echo "ual Go Microbenchmarks"
echo "=============================================================================="
echo ""

run_bench() {
    local pattern="$1"
    local desc="$2"
    echo "--- $desc ---"
    go test -bench="$pattern" -benchmem -count=$COUNT 2>/dev/null | grep -E "^Benchmark|^pkg:|^goos:|^goarch:"
    echo ""
}

run_c_bench() {
    echo "--- C Reference Benchmarks (gcc -O2) ---"
    if [ ! -f c/c_bench ]; then
        echo "Compiling C benchmarks..."
        (cd c && gcc -O2 -o c_bench c_bench.c -lm)
    fi
    ./c/c_bench
}

case "$MODE" in
    compute)
        echo "COMPUTE-ONLY BENCHMARKS"
        echo "Measures pure computation quality (no stack overhead)"
        echo ""
        run_bench "BenchmarkCompute_" "Go vs ual Compute Comparison"
        ;;
    
    pipeline)
        echo "FULL PIPELINE BENCHMARKS"
        echo "Measures complete ual pattern (lock + bytes + compute)"
        echo ""
        run_bench "BenchmarkPipeline_" "All Pipeline Benchmarks"
        ;;
    
    overhead)
        echo "OVERHEAD ISOLATION BENCHMARKS"
        echo "Measures individual sources of overhead"
        echo ""
        run_bench "BenchmarkOverhead_" "Overhead Components"
        ;;
    
    c)
        echo "C REFERENCE BENCHMARKS"
        echo ""
        run_c_bench
        ;;
    
    all|*)
        echo "COMPUTE-ONLY BENCHMARKS (Go vs ual)"
        echo "=============================================================================="
        run_bench "BenchmarkCompute_" "Go vs ual Compute"
        
        echo ""
        echo "C REFERENCE BENCHMARKS"
        echo "=============================================================================="
        run_c_bench
        
        echo ""
        echo "FULL PIPELINE BENCHMARKS"
        echo "=============================================================================="
        run_bench "BenchmarkPipeline_" "Pipeline (lock + bytes + compute)"
        
        echo ""
        echo "OVERHEAD ISOLATION"
        echo "=============================================================================="
        run_bench "BenchmarkOverhead_" "Overhead Components"
        ;;
esac

echo "=============================================================================="
echo "INTERPRETATION"
echo "=============================================================================="
echo ""
echo "Compute benchmarks: ual should match Go (compiles to Go)."
echo "Pipeline benchmarks: Shows overhead from lock/unlock + byte conversion."
echo "Overhead benchmarks: Isolated costs of individual operations."
echo ""
