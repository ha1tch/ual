#!/bin/bash
# benchmark_backends.sh - Compare Go, Rust, and iual performance
#
# Measures:
#   - Execution time (average over multiple runs)
#   - Peak memory usage
#   - Binary size (compiled backends)
#   - Startup time

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Configuration
WARMUP_RUNS=2
TIMED_RUNS=5
TIMEOUT_SEC=30

# Temp directory
WORK_DIR=$(mktemp -d)
trap "rm -rf $WORK_DIR" EXIT

# Colours
if [ -t 1 ]; then
    BOLD='\033[1m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    CYAN='\033[0;36m'
    NC='\033[0m'
else
    BOLD='' GREEN='' YELLOW='' CYAN='' NC=''
fi

echo_header() { echo -e "\n${BOLD}=== $1 ===${NC}"; }
echo_info() { echo -e "${CYAN}$1${NC}"; }

# Get file size in bytes (portable)
file_size() {
    local file="$1"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        stat -f%z "$file" 2>/dev/null || echo "N/A"
    else
        stat -c%s "$file" 2>/dev/null || echo "N/A"
    fi
}

# Check prerequisites
check_prereqs() {
    echo_header "Prerequisites"
    
    local have_go=false have_rust=false have_iual=false
    
    if command -v go &>/dev/null; then
        echo "Go: $(go version | sed -E 's/.*go([0-9]+\.[0-9]+).*/go\1/')"
        have_go=true
    else
        echo "Go: not found"
    fi
    
    if command -v rustc &>/dev/null; then
        echo "Rust: $(rustc --version | sed -E 's/.*([0-9]+\.[0-9]+\.[0-9]+).*/\1/')"
        have_rust=true
    else
        echo "Rust: not found"
    fi
    
    if [ -x "./ual" ]; then
        echo "ual compiler: found"
    else
        echo "Building ual compiler..."
        go build -o ual ./cmd/ual/
    fi
    
    if [ -x "./iual" ]; then
        echo "iual interpreter: found"
        have_iual=true
    else
        echo "Building iual interpreter..."
        go build -o iual ./cmd/iual/
        have_iual=true
    fi
    
    # Store capabilities
    echo "$have_go" > "$WORK_DIR/have_go"
    echo "$have_rust" > "$WORK_DIR/have_rust"
    echo "$have_iual" > "$WORK_DIR/have_iual"
}

# Set up Rust project once
setup_rust() {
    if [ "$(cat $WORK_DIR/have_rust)" != "true" ]; then
        return 1
    fi
    
    mkdir -p "$WORK_DIR/rust_proj/src"
    cat > "$WORK_DIR/rust_proj/Cargo.toml" << EOF
[package]
name = "ual_bench"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "$SCRIPT_DIR/rual" }
lazy_static = "1.4"

[profile.release]
opt-level = 3
lto = true
codegen-units = 1

[profile.dev]
opt-level = 0
EOF
    
    # Pre-compile dependencies
    echo "fn main() {}" > "$WORK_DIR/rust_proj/src/main.rs"
    (cd "$WORK_DIR/rust_proj" && cargo build --release 2>/dev/null) || return 1
    return 0
}

# Get peak memory usage (RSS in KB)
get_peak_memory() {
    local cmd="$1"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS: /usr/bin/time -l reports in bytes, convert to KB
        local bytes=$(/usr/bin/time -l $cmd 2>&1 | grep "maximum resident set size" | awk '{print $1}')
        if [ -n "$bytes" ] && [ "$bytes" -gt 0 ] 2>/dev/null; then
            echo $((bytes / 1024))
        else
            echo "N/A"
        fi
    elif command -v /usr/bin/time &>/dev/null; then
        # Linux: /usr/bin/time -v reports in KB directly
        /usr/bin/time -v $cmd 2>&1 | grep "Maximum resident set size" | awk '{print $6}'
    else
        echo "N/A"
    fi
}

# Get time in nanoseconds (portable)
get_time_ns() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS: use perl for nanosecond precision
        perl -MTime::HiRes -e 'printf("%.0f\n", Time::HiRes::time()*1000000000)'
    else
        # Linux: use date with nanoseconds
        date +%s%N
    fi
}

# Time a command (returns milliseconds)
time_cmd() {
    local cmd="$1"
    local start=$(get_time_ns)
    eval "$cmd" >/dev/null 2>&1
    local end=$(get_time_ns)
    echo $(( (end - start) / 1000000 ))
}

# Average of array
average() {
    local sum=0
    local count=0
    for val in "$@"; do
        if [[ "$val" =~ ^[0-9]+$ ]]; then
            sum=$((sum + val))
            count=$((count + 1))
        fi
    done
    if [ $count -gt 0 ]; then
        echo $((sum / count))
    else
        echo "N/A"
    fi
}

# Benchmark a single example across all backends
benchmark_example() {
    local ual_file="$1"
    local name=$(basename "$ual_file" .ual)
    
    local go_time="N/A" rust_time="N/A" iual_time="N/A"
    local go_mem="N/A" rust_mem="N/A" iual_mem="N/A"
    local go_size="N/A" rust_size="N/A"
    
    # === Go Backend ===
    if [ "$(cat $WORK_DIR/have_go)" = "true" ]; then
        # Compile
        if ./ual compile --target go "$ual_file" -o "$WORK_DIR/${name}.go" 2>/dev/null; then
            if go build -o "$WORK_DIR/${name}_go" "$WORK_DIR/${name}.go" 2>/dev/null; then
                go_size=$(file_size "$WORK_DIR/${name}_go" 2>/dev/null || echo "N/A")
                
                # Warmup
                for i in $(seq 1 $WARMUP_RUNS); do
                    timeout ${TIMEOUT_SEC}s "$WORK_DIR/${name}_go" >/dev/null 2>&1 || true
                done
                
                # Timed runs
                local times=()
                for i in $(seq 1 $TIMED_RUNS); do
                    local t=$(time_cmd "timeout ${TIMEOUT_SEC}s $WORK_DIR/${name}_go")
                    times+=($t)
                done
                go_time=$(average "${times[@]}")
                
                # Memory (single run)
                go_mem=$(get_peak_memory "timeout ${TIMEOUT_SEC}s $WORK_DIR/${name}_go")
            fi
        fi
    fi
    
    # === Rust Backend ===
    if [ "$(cat $WORK_DIR/have_rust)" = "true" ]; then
        if ./ual compile --target rust "$ual_file" -o "$WORK_DIR/rust_proj/src/main.rs" 2>/dev/null; then
            if (cd "$WORK_DIR/rust_proj" && cargo build --release 2>/dev/null); then
                local rust_bin="$WORK_DIR/rust_proj/target/release/ual_bench"
                rust_size=$(file_size "$rust_bin" 2>/dev/null || echo "N/A")
                
                # Warmup
                for i in $(seq 1 $WARMUP_RUNS); do
                    timeout ${TIMEOUT_SEC}s "$rust_bin" >/dev/null 2>&1 || true
                done
                
                # Timed runs
                local times=()
                for i in $(seq 1 $TIMED_RUNS); do
                    local t=$(time_cmd "timeout ${TIMEOUT_SEC}s $rust_bin")
                    times+=($t)
                done
                rust_time=$(average "${times[@]}")
                
                # Memory
                rust_mem=$(get_peak_memory "timeout ${TIMEOUT_SEC}s $rust_bin")
            fi
        fi
    fi
    
    # === iual Interpreter ===
    if [ "$(cat $WORK_DIR/have_iual)" = "true" ]; then
        # Warmup
        for i in $(seq 1 $WARMUP_RUNS); do
            timeout ${TIMEOUT_SEC}s ./iual -q "$ual_file" >/dev/null 2>&1 || true
        done
        
        # Timed runs
        local times=()
        for i in $(seq 1 $TIMED_RUNS); do
            local t=$(time_cmd "timeout ${TIMEOUT_SEC}s ./iual -q $ual_file")
            times+=($t)
        done
        iual_time=$(average "${times[@]}")
        
        # Memory
        iual_mem=$(get_peak_memory "timeout ${TIMEOUT_SEC}s ./iual -q $ual_file")
    fi
    
    # Output CSV line
    echo "$name,$go_time,$rust_time,$iual_time,$go_mem,$rust_mem,$iual_mem,$go_size,$rust_size"
}

# Format size for display
format_size() {
    local size=$1
    if [[ "$size" == "N/A" ]]; then
        echo "N/A"
    elif [ "$size" -gt 1048576 ]; then
        echo "$(echo "scale=1; $size/1048576" | bc)M"
    elif [ "$size" -gt 1024 ]; then
        echo "$(echo "scale=0; $size/1024" | bc)K"
    else
        echo "${size}B"
    fi
}

# Main benchmark suite
run_benchmarks() {
    echo_header "Running Benchmarks"
    echo "Warmup runs: $WARMUP_RUNS, Timed runs: $TIMED_RUNS"
    echo ""
    
    # Select benchmark examples (compute-heavy ones)
    local bench_examples=(
        "examples/001_fibonacci.ual"
        "examples/008_primes.ual"
        "examples/038_compute_newton.ual"
        "examples/039_compute_mandelbrot.ual"
        "examples/040_compute_integrate.ual"
        "examples/041_compute_leibniz.ual"
        "examples/050_compute_dp.ual"
        "examples/059_algorithms.ual"
    )
    
    # CSV output
    local csv_file="$WORK_DIR/results.csv"
    echo "Example,Go_ms,Rust_ms,iual_ms,Go_KB,Rust_KB,iual_KB,Go_bytes,Rust_bytes" > "$csv_file"
    
    for ual_file in "${bench_examples[@]}"; do
        if [ -f "$ual_file" ]; then
            echo -n "  $(basename $ual_file .ual)... "
            local result=$(benchmark_example "$ual_file")
            echo "$result" >> "$csv_file"
            echo "done"
        fi
    done
    
    # Generate report
    echo_header "Results"
    echo ""
    
    printf "%-25s %10s %10s %10s %8s\n" "Example" "Go (ms)" "Rust (ms)" "iual (ms)" "Ratio"
    printf "%-25s %10s %10s %10s %8s\n" "-------" "-------" "--------" "--------" "-----"
    
    tail -n +2 "$csv_file" | while IFS=, read -r name go_t rust_t iual_t go_m rust_m iual_m go_s rust_s; do
        local ratio="N/A"
        if [[ "$go_t" =~ ^[0-9]+$ ]] && [[ "$iual_t" =~ ^[0-9]+$ ]] && [ "$go_t" -gt 0 ]; then
            ratio=$(echo "scale=1; $iual_t / $go_t" | bc)x
        fi
        printf "%-25s %10s %10s %10s %8s\n" "$name" "$go_t" "$rust_t" "$iual_t" "$ratio"
    done
    
    echo ""
    echo_header "Memory Usage (Peak RSS in KB)"
    printf "%-25s %10s %10s %10s\n" "Example" "Go" "Rust" "iual"
    printf "%-25s %10s %10s %10s\n" "-------" "----" "------" "------"
    
    tail -n +2 "$csv_file" | while IFS=, read -r name go_t rust_t iual_t go_m rust_m iual_m go_s rust_s; do
        printf "%-25s %10s %10s %10s\n" "$name" "$go_m" "$rust_m" "$iual_m"
    done
    
    echo ""
    echo_header "Binary Sizes"
    printf "%-25s %10s %10s\n" "Example" "Go" "Rust"
    printf "%-25s %10s %10s\n" "-------" "----" "------"
    
    tail -n +2 "$csv_file" | while IFS=, read -r name go_t rust_t iual_t go_m rust_m iual_m go_s rust_s; do
        printf "%-25s %10s %10s\n" "$name" "$(format_size $go_s)" "$(format_size $rust_s)"
    done
    
    # Copy results
    cp "$csv_file" "./benchmark_results.csv"
    echo ""
    echo "Raw results saved to: benchmark_results.csv"
}

# Quick smoke test
smoke_test() {
    echo_header "Smoke Test (single example)"
    local test_file="examples/001_fibonacci.ual"
    
    echo "Testing: $test_file"
    echo ""
    
    local result=$(benchmark_example "$test_file")
    IFS=, read -r name go_t rust_t iual_t go_m rust_m iual_m go_s rust_s <<< "$result"
    
    echo "Execution Time:"
    echo "  Go:   ${go_t}ms"
    echo "  Rust: ${rust_t}ms"
    echo "  iual: ${iual_t}ms"
    echo ""
    echo "Memory (Peak RSS KB):"
    echo "  Go:   $go_m"
    echo "  Rust: $rust_m"
    echo "  iual: $iual_m"
    echo ""
    echo "Binary Size:"
    echo "  Go:   $(format_size $go_s)"
    echo "  Rust: $(format_size $rust_s)"
}

# Main
main() {
    echo -e "${BOLD}ual Three-Way Backend Benchmark${NC}"
    echo "================================="
    
    check_prereqs
    
    if [ "$(cat $WORK_DIR/have_rust)" = "true" ]; then
        echo_header "Setting up Rust project"
        if ! setup_rust; then
            echo "Warning: Rust setup failed, skipping Rust benchmarks"
            echo "false" > "$WORK_DIR/have_rust"
        else
            echo "Done"
        fi
    fi
    
    case "${1:-full}" in
        smoke)
            smoke_test
            ;;
        full|"")
            run_benchmarks
            ;;
        *)
            echo "Usage: $0 [smoke|full]"
            exit 1
            ;;
    esac
}

main "$@"
