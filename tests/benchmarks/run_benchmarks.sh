#!/bin/bash
# =============================================================================
# ual Cross-Backend Benchmark Suite
# =============================================================================
#
# Benchmarks ual programs across Go, Rust, and iual backends.
# Measures execution time, binary size, and generates results.
#
# Usage:
#   ./run_benchmarks.sh [OPTIONS]
#
# Options:
#   --quick           Run quick smoke test (1 iteration)
#   --full            Run full benchmarks (5 iterations, default)
#   --go              Test Go backend only
#   --rust            Test Rust backend only  
#   --iual            Test interpreter only
#   --all             Test all backends (default)
#   --json            Output results as JSON
#   --output FILE     Write JSON results to file
#   -h, --help        Show this help
#
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROGRAMS_DIR="$SCRIPT_DIR/programs"
RESULTS_DIR="$SCRIPT_DIR/results"

cd "$PROJECT_DIR"

# =============================================================================
# Configuration
# =============================================================================

ITERATIONS=5
TEST_GO=false
TEST_RUST=false
TEST_Iual=false
JSON_OUTPUT=false
OUTPUT_FILE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --quick)    ITERATIONS=1; shift ;;
        --full)     ITERATIONS=5; shift ;;
        --go)       TEST_GO=true; shift ;;
        --rust)     TEST_RUST=true; shift ;;
        --iual)     TEST_Iual=true; shift ;;
        --all)      TEST_GO=true; TEST_RUST=true; TEST_Iual=true; shift ;;
        --json)     JSON_OUTPUT=true; shift ;;
        --output)   OUTPUT_FILE="$2"; shift 2 ;;
        -h|--help)  head -25 "$0" | tail -20; exit 0 ;;
        *)          echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Default to all backends
if ! $TEST_GO && ! $TEST_RUST && ! $TEST_Iual; then
    TEST_GO=true
    TEST_RUST=true
    TEST_Iual=true
fi

# =============================================================================
# Colours
# =============================================================================

if [ -t 1 ] && ! $JSON_OUTPUT; then
    BOLD='\033[1m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    BOLD='' GREEN='' YELLOW='' BLUE='' NC=''
fi

# =============================================================================
# Setup
# =============================================================================

ensure_tools() {
    if [ ! -x "./ual" ]; then
        echo "Building ual compiler..."
        go build -o ual ./cmd/ual/ 2>/dev/null
    fi
    if [ ! -x "./iual" ]; then
        echo "Building iual interpreter..."
        go build -o iual ./cmd/iual/ 2>/dev/null
    fi
}

# Rust project for compilation
RUST_PROJECT=""
RUST_AVAILABLE=false

setup_rust() {
    command -v rustc &>/dev/null || return 1
    
    RUST_PROJECT=$(mktemp -d)
    mkdir -p "$RUST_PROJECT/src"
    
    cat > "$RUST_PROJECT/Cargo.toml" << EOF
[package]
name = "ual_bench"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "$PROJECT_DIR/rual" }
lazy_static = "1.4"

[profile.release]
opt-level = 3
lto = true
EOF
    
    echo "fn main() {}" > "$RUST_PROJECT/src/main.rs"
    if (cd "$RUST_PROJECT" && cargo build --release 2>/dev/null); then
        RUST_AVAILABLE=true
        return 0
    fi
    rm -rf "$RUST_PROJECT"
    RUST_PROJECT=""
    return 1
}

cleanup_rust() {
    [ -n "$RUST_PROJECT" ] && [ -d "$RUST_PROJECT" ] && rm -rf "$RUST_PROJECT"
}
trap cleanup_rust EXIT

# =============================================================================
# Timing Functions (portable for macOS and Linux)
# =============================================================================

# Get time in milliseconds
get_time_ms() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        perl -MTime::HiRes -e 'printf("%.0f\n", Time::HiRes::time()*1000)'
    else
        date +%s%3N
    fi
}

# Get file size in bytes (portable)
file_size() {
    local file="$1"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        stat -f%z "$file" 2>/dev/null || echo 0
    else
        stat -c%s "$file" 2>/dev/null || echo 0
    fi
}

# Run command and return execution time in ms
time_command() {
    local start=$(get_time_ms)
    "$@" >/dev/null 2>&1
    local end=$(get_time_ms)
    echo $((end - start))
}

# Run multiple iterations and return median
run_benchmark() {
    local cmd=("$@")
    local times=()
    
    # Warmup
    "${cmd[@]}" >/dev/null 2>&1 || return 1
    
    # Timed runs
    for ((i=0; i<ITERATIONS; i++)); do
        local t=$(time_command "${cmd[@]}")
        times+=($t)
    done
    
    # Sort and get median
    IFS=$'\n' sorted=($(sort -n <<<"${times[*]}")); unset IFS
    local mid=$((ITERATIONS / 2))
    echo "${sorted[$mid]}"
}

# =============================================================================
# Binary Size
# =============================================================================

get_binary_sizes() {
    local sizes=""
    
    # Go binary (stripped)
    if [ -x "./ual" ]; then
        local go_tmp=$(mktemp)
        ./ual compile tests/benchmarks/programs/bench_compute_leibniz.ual -o "${go_tmp}.go" 2>/dev/null
        go build -ldflags="-s -w" -o "$go_tmp" "${go_tmp}.go" 2>/dev/null
        local go_size=$(file_size "$go_tmp" 2>/dev/null || echo 0)
        rm -f "$go_tmp" "${go_tmp}.go"
        sizes="$go_size"
    fi
    
    # Rust binary (stripped)
    if $RUST_AVAILABLE; then
        ./ual compile --target rust tests/benchmarks/programs/bench_compute_leibniz.ual -o "$RUST_PROJECT/src/main.rs" 2>/dev/null
        (cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)
        strip "$RUST_PROJECT/target/release/ual_bench" 2>/dev/null || true
        local rust_size=$(file_size "$RUST_PROJECT/target/release/ual_bench" 2>/dev/null || echo 0)
        sizes="$sizes $rust_size"
    else
        sizes="$sizes 0"
    fi
    
    # iual interpreter size
    local iual_size=$(file_size "./iual" 2>/dev/null || echo 0)
    sizes="$sizes $iual_size"
    
    echo $sizes
}

# =============================================================================
# Main Benchmark Run
# =============================================================================

run_benchmarks() {
    ensure_tools
    
    if $TEST_RUST; then
        ! $JSON_OUTPUT && echo -e "${BLUE}Setting up Rust...${NC}"
        setup_rust || TEST_RUST=false
    fi
    
    mkdir -p "$RESULTS_DIR"
    
    # Header
    if ! $JSON_OUTPUT; then
        echo ""
        echo -e "${BOLD}ual Cross-Backend Benchmarks${NC}"
        echo "=============================="
        echo "Iterations: $ITERATIONS"
        echo ""
        printf "%-30s" "Benchmark"
        $TEST_GO && printf "%12s" "Go (ms)"
        $TEST_RUST && printf "%12s" "Rust (ms)"
        $TEST_Iual && printf "%12s" "iual (ms)"
        echo ""
        printf "%-30s" "-----------------------------"
        $TEST_GO && printf "%12s" "--------"
        $TEST_RUST && printf "%12s" "--------"
        $TEST_Iual && printf "%12s" "--------"
        echo ""
    fi
    
    # JSON accumulator
    local json_benchmarks="["
    local json_first=true
    
    # Run each benchmark
    for prog in "$PROGRAMS_DIR"/*.ual; do
        [ -f "$prog" ] || continue
        local name=$(basename "$prog" .ual)
        
        local go_ms=0 rust_ms=0 iual_ms=0
        
        # Go backend
        if $TEST_GO; then
            local go_tmp=$(mktemp)
            ./ual compile "$prog" -o "${go_tmp}.go" 2>/dev/null
            go build -o "$go_tmp" "${go_tmp}.go" 2>/dev/null
            go_ms=$(run_benchmark "$go_tmp")
            rm -f "$go_tmp" "${go_tmp}.go"
        fi
        
        # Rust backend
        if $TEST_RUST && $RUST_AVAILABLE; then
            ./ual compile --target rust "$prog" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null
            (cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)
            rust_ms=$(run_benchmark "$RUST_PROJECT/target/release/ual_bench")
        fi
        
        # iual interpreter
        if $TEST_Iual; then
            iual_ms=$(run_benchmark ./iual -q "$prog")
        fi
        
        # Output
        if ! $JSON_OUTPUT; then
            printf "%-30s" "$name"
            $TEST_GO && printf "%12s" "$go_ms"
            $TEST_RUST && printf "%12s" "$rust_ms"
            $TEST_Iual && printf "%12s" "$iual_ms"
            echo ""
        fi
        
        # JSON
        $json_first || json_benchmarks+=","
        json_first=false
        json_benchmarks+="{\"name\":\"$name\",\"go_ms\":$go_ms,\"rust_ms\":$rust_ms,\"iual_ms\":$iual_ms}"
    done
    
    json_benchmarks+="]"
    
    # Binary sizes
    local sizes=($(get_binary_sizes))
    local go_size=${sizes[0]:-0}
    local rust_size=${sizes[1]:-0}
    local iual_size=${sizes[2]:-0}
    
    if ! $JSON_OUTPUT; then
        echo ""
        echo -e "${BOLD}Binary Sizes (stripped)${NC}"
        echo "-----------------------"
        $TEST_GO && echo "Go:   $(numfmt --to=iec-i --suffix=B $go_size 2>/dev/null || echo "${go_size}B")"
        $TEST_RUST && echo "Rust: $(numfmt --to=iec-i --suffix=B $rust_size 2>/dev/null || echo "${rust_size}B")"
        $TEST_Iual && echo "iual: $(numfmt --to=iec-i --suffix=B $iual_size 2>/dev/null || echo "${iual_size}B")"
    fi
    
    # Generate JSON output
    local json_result=$(cat <<EOF
{
  "version": "$(cat VERSION 2>/dev/null || echo 'dev')",
  "timestamp": "$(date -Iseconds)",
  "iterations": $ITERATIONS,
  "benchmarks": $json_benchmarks,
  "binary_sizes": {
    "go_stripped": $go_size,
    "rust_stripped": $rust_size,
    "iual": $iual_size
  }
}
EOF
)
    
    if $JSON_OUTPUT; then
        echo "$json_result"
    fi
    
    if [ -n "$OUTPUT_FILE" ]; then
        echo "$json_result" > "$OUTPUT_FILE"
        ! $JSON_OUTPUT && echo -e "\nResults saved to: $OUTPUT_FILE"
    fi
    
    # Also save to results directory
    local timestamp=$(date +%Y%m%d_%H%M%S)
    echo "$json_result" > "$RESULTS_DIR/benchmark_${timestamp}.json"
    ln -sf "benchmark_${timestamp}.json" "$RESULTS_DIR/latest.json"
}

# =============================================================================
# Entry Point
# =============================================================================

run_benchmarks
