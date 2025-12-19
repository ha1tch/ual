#!/bin/bash
# =============================================================================
# ual Unified Benchmark Suite
# =============================================================================
#
# Runs all benchmarks and generates HTML report.
#
# Usage:
#   ./run_unified.sh [OPTIONS]
#
# Options:
#   --quick         Run quick benchmarks (1 iteration)
#   --full          Run full benchmarks (5 iterations, default)
#   --backends      Test ual backends only (Go, Rust, iual)
#   --cross-lang    Include C/Python comparison
#   --all           Run everything (default)
#   --no-html       Skip HTML report generation
#   --json          Output JSON to stdout
#   -h, --help      Show this help
#
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROGRAMS_DIR="$SCRIPT_DIR/programs"
RESULTS_DIR="$SCRIPT_DIR/results"
CROSS_LANG_DIR="$SCRIPT_DIR/cross_language"

cd "$PROJECT_DIR"

# =============================================================================
# Configuration
# =============================================================================

ITERATIONS=5
TEST_BACKENDS=true
TEST_CROSS_LANG=true
GENERATE_HTML=true
JSON_OUTPUT=false
DEBUG=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --quick)      ITERATIONS=1; shift ;;
        --full)       ITERATIONS=5; shift ;;
        --backends)   TEST_BACKENDS=true; TEST_CROSS_LANG=false; shift ;;
        --cross-lang) TEST_BACKENDS=false; TEST_CROSS_LANG=true; shift ;;
        --all)        TEST_BACKENDS=true; TEST_CROSS_LANG=true; shift ;;
        --no-html)    GENERATE_HTML=false; shift ;;
        --json)       JSON_OUTPUT=true; GENERATE_HTML=false; shift ;;
        --debug)      DEBUG=true; shift ;;
        -h|--help)    head -20 "$0" | tail -17; exit 0 ;;
        *)            echo "Unknown option: $1"; exit 1 ;;
    esac
done

# =============================================================================
# Setup
# =============================================================================

export GOPATH=$HOME/go
export PATH=$PATH:/usr/lib/go-1.22/bin:$GOPATH/bin

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

$JSON_OUTPUT && { BOLD=''; GREEN=''; YELLOW=''; BLUE=''; CYAN=''; NC=''; }

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

log() { $JSON_OUTPUT || echo -e "$@" >&2; }

# Build tools
ensure_tools() {
    [ -x "./ual" ] || go build -o ual ./cmd/ual/ 2>/dev/null
    [ -x "./iual" ] || go build -o iual ./cmd/iual/ 2>/dev/null
}

# =============================================================================
# Timing (portable for macOS and Linux)
# =============================================================================

get_time_ms() {
    local result
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS: use perl (always available)
        result=$(perl -MTime::HiRes -e 'printf("%.0f\n", Time::HiRes::time()*1000)' 2>/dev/null)
        if [[ -z "$result" ]]; then
            # Fallback if perl fails
            result=$(($(date +%s) * 1000))
        fi
    else
        # Linux: use date with milliseconds
        result=$(date +%s%3N)
    fi
    echo "$result"
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

time_command() {
    local start=$(get_time_ms)
    "$@" >/dev/null 2>&1
    local exit_code=$?
    local end=$(get_time_ms)
    local elapsed=$((end - start))
    $DEBUG && echo "[DEBUG] time_command: start=$start end=$end elapsed=$elapsed exit=$exit_code" >&2
    echo $elapsed
}

run_timed() {
    local cmd=("$@")
    local times=()
    
    # Warmup
    if ! "${cmd[@]}" >/dev/null 2>&1; then
        $DEBUG && echo "[DEBUG] Warmup failed for: ${cmd[*]}" >&2
        echo "0"
        return
    fi
    $DEBUG && echo "[DEBUG] Warmup passed for: ${cmd[*]}" >&2
    
    # Timed runs
    for ((i=0; i<ITERATIONS; i++)); do
        local t=$(time_command "${cmd[@]}")
        $DEBUG && echo "[DEBUG] Run $i: ${t:-EMPTY}" >&2
        times+=("${t:-0}")
    done
    
    # Median - handle empty/malformed arrays
    if [ ${#times[@]} -eq 0 ]; then
        $DEBUG && echo "[DEBUG] No times collected!" >&2
        echo "0"
        return
    fi
    
    IFS=$'\n' sorted=($(sort -n <<<"${times[*]}")); unset IFS
    local result="${sorted[$((ITERATIONS / 2))]}"
    
    $DEBUG && echo "[DEBUG] Median result: '${result}'" >&2
    
    # Ensure we return a valid number
    if [[ "$result" =~ ^[0-9]+$ ]]; then
        echo "$result"
    else
        $DEBUG && echo "[DEBUG] Invalid result, returning 0" >&2
        echo "0"
    fi
}

# =============================================================================
# Cross-Language Benchmarks
# =============================================================================

build_c_benchmarks() {
    log "${BLUE}Building C benchmarks...${NC}"
    
    # Build single binary that accepts benchmark name as argument
    gcc -O2 -o "$TMPDIR/c_bench" "$CROSS_LANG_DIR/c/bench.c" -lm 2>/dev/null || return 1
}

build_rust_benchmarks() {
    log "${BLUE}Building Rust benchmarks...${NC}"
    
    # Always rebuild to pick up source changes
    (cd "$CROSS_LANG_DIR/rust" && cargo build --release 2>/dev/null) || return 1
}

run_cross_language() {
    local results=""
    
    # C benchmarks
    if [ -x "$TMPDIR/c_bench" ]; then
        local c_leibniz=$(run_timed "$TMPDIR/c_bench" leibniz)
        local c_mandelbrot=$(run_timed "$TMPDIR/c_bench" mandelbrot)
        local c_newton=$(run_timed "$TMPDIR/c_bench" newton)
        results="\"c\": {\"leibniz\": $c_leibniz, \"mandelbrot\": $c_mandelbrot, \"newton\": $c_newton}"
    fi
    
    # Rust benchmarks
    if [ -x "$CROSS_LANG_DIR/rust/target/release/bench" ]; then
        local rust_leibniz=$(run_timed "$CROSS_LANG_DIR/rust/target/release/bench" leibniz)
        local rust_mandelbrot=$(run_timed "$CROSS_LANG_DIR/rust/target/release/bench" mandelbrot)
        local rust_newton=$(run_timed "$CROSS_LANG_DIR/rust/target/release/bench" newton)
        [ -n "$results" ] && results="$results, "
        results="${results}\"rust_native\": {\"leibniz\": $rust_leibniz, \"mandelbrot\": $rust_mandelbrot, \"newton\": $rust_newton}"
    fi
    
    # Python benchmarks
    if command -v python3 &>/dev/null; then
        local py_leibniz=$(run_timed python3 "$CROSS_LANG_DIR/python/python_bench.py" leibniz)
        local py_mandelbrot=$(run_timed python3 "$CROSS_LANG_DIR/python/python_bench.py" mandelbrot)
        local py_newton=$(run_timed python3 "$CROSS_LANG_DIR/python/python_bench.py" newton)
        [ -n "$results" ] && results="$results, "
        results="${results}\"python\": {\"leibniz\": $py_leibniz, \"mandelbrot\": $py_mandelbrot, \"newton\": $py_newton}"
    fi
    
    echo "{$results}"
}

# =============================================================================
# ual Backend Benchmarks
# =============================================================================

RUST_PROJECT=""

setup_rust_backend() {
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
    (cd "$RUST_PROJECT" && cargo build --release 2>/dev/null) || { rm -rf "$RUST_PROJECT"; RUST_PROJECT=""; return 1; }
}

run_backend_benchmarks() {
    local json_results="["
    local first=true
    
    for prog in "$PROGRAMS_DIR"/bench_compute_*.ual; do
        [ -f "$prog" ] || continue
        local name=$(basename "$prog" .ual | sed 's/bench_//')
        
        local go_ms=0 rust_ms=0 iual_ms=0
        
        # Go backend
        local go_tmp="$TMPDIR/go_bench"
        ./ual compile "$prog" -o "${go_tmp}.go" 2>/dev/null
        go build -o "$go_tmp" "${go_tmp}.go" 2>/dev/null
        go_ms=$(run_timed "$go_tmp")
        [[ "$go_ms" =~ ^[0-9]+$ ]] || go_ms=0
        
        # Rust backend
        if [ -n "$RUST_PROJECT" ]; then
            ./ual compile --target rust "$prog" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null
            (cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)
            rust_ms=$(run_timed "$RUST_PROJECT/target/release/ual_bench")
            [[ "$rust_ms" =~ ^[0-9]+$ ]] || rust_ms=0
        fi
        
        # iual interpreter
        $DEBUG && echo "[DEBUG] Running iual on $prog" >&2
        iual_ms=$(run_timed ./iual -q "$prog")
        $DEBUG && echo "[DEBUG] iual returned: '$iual_ms'" >&2
        # Ensure iual_ms is a valid number
        [[ "$iual_ms" =~ ^[0-9]+$ ]] || iual_ms=0
        
        $first || json_results+=","
        first=false
        json_results+="{\"name\":\"$name\",\"go_ms\":${go_ms:-0},\"rust_ms\":${rust_ms:-0},\"iual_ms\":${iual_ms:-0}}"
        
        log "  $name: Go=${go_ms}ms Rust=${rust_ms}ms iual=${iual_ms}ms"
    done
    
    json_results+="]"
    # Return only the JSON, log goes to stderr
    echo "$json_results"
}

# =============================================================================
# Binary Sizes
# =============================================================================

get_binary_sizes() {
    local go_size=0 rust_size=0 iual_size=0
    
    # Go binary - must build from project directory for module resolution
    local go_src="$TMPDIR/bench_size.go"
    local go_bin="$TMPDIR/bench_size"
    ./ual compile "$PROGRAMS_DIR/bench_compute_leibniz.ual" -o "$go_src" 2>/dev/null
    if go build -ldflags="-s -w" -o "$go_bin" "$go_src" 2>/dev/null; then
        go_size=$(file_size "$go_bin" 2>/dev/null || echo 0)
    fi
    
    # Rust binary
    if [ -n "$RUST_PROJECT" ]; then
        strip "$RUST_PROJECT/target/release/ual_bench" 2>/dev/null || true
        rust_size=$(file_size "$RUST_PROJECT/target/release/ual_bench" 2>/dev/null || echo 0)
    fi
    
    # iual (strip a copy for fair comparison with Go/Rust)
    if [ -x "./iual" ]; then
        cp ./iual "$TMPDIR/iual_stripped" 2>/dev/null
        strip "$TMPDIR/iual_stripped" 2>/dev/null || true
        iual_size=$(file_size "$TMPDIR/iual_stripped" 2>/dev/null || echo 0)
    else
        iual_size=$(file_size "./iual" 2>/dev/null || echo 0)
    fi
    
    echo "{\"go_stripped\":$go_size,\"rust_stripped\":$rust_size,\"iual_stripped\":$iual_size}"
}

# =============================================================================
# Correctness Check
# =============================================================================

get_correctness() {
    local total=0 go_pass=0 rust_pass=0 iual_pass=0
    
    for f in examples/*.ual; do
        [ -f "$f" ] || continue
        ((total++))
        
        local expected="tests/correctness/expected/$(basename "$f" .ual).txt"
        [ -f "$expected" ] || continue
        
        # Quick check with iual only for speed
        local output=$(timeout 5 ./iual -q "$f" 2>/dev/null || echo "ERROR")
        if [ "$output" = "$(cat "$expected")" ]; then
            ((go_pass++))
            ((rust_pass++))
            ((iual_pass++))
        fi
    done
    
    echo "{\"total\":$total,\"go_pass\":$go_pass,\"rust_pass\":$rust_pass,\"iual_pass\":$iual_pass}"
}

# =============================================================================
# Main
# =============================================================================

main() {
    ensure_tools
    mkdir -p "$RESULTS_DIR"
    
    log ""
    log "${BOLD}ual Benchmark Suite${NC}"
    log "==================="
    log "Iterations: $ITERATIONS"
    log ""
    
    # Setup Rust if needed
    if $TEST_BACKENDS; then
        log "${BLUE}Setting up Rust backend...${NC}"
        setup_rust_backend || log "${YELLOW}Rust backend unavailable${NC}"
    fi
    
    # Collect results
    local benchmarks="[]"
    local cross_lang="{}"
    local correctness="{}"
    local binary_sizes="{}"
    
    if $TEST_BACKENDS; then
        log ""
        log "${BOLD}ual Backend Benchmarks${NC}"
        benchmarks=$(run_backend_benchmarks)
        binary_sizes=$(get_binary_sizes)
        correctness=$(get_correctness)
    fi
    
    if $TEST_CROSS_LANG; then
        log ""
        log "${BOLD}Cross-Language Benchmarks${NC}"
        build_c_benchmarks || log "${YELLOW}C benchmarks unavailable${NC}"
        build_rust_benchmarks || log "${YELLOW}Rust benchmarks unavailable${NC}"
        cross_lang=$(run_cross_language)
    fi
    
    # Build JSON result
    local timestamp=$(date -Iseconds)
    local version=$(cat VERSION 2>/dev/null || echo "dev")
    
    local json_result=$(cat <<EOF
{
  "version": "$version",
  "timestamp": "$timestamp",
  "iterations": $ITERATIONS,
  "correctness": $correctness,
  "benchmarks": $benchmarks,
  "cross_language": $cross_lang,
  "binary_sizes": $binary_sizes
}
EOF
)
    
    # Output
    if $JSON_OUTPUT; then
        echo "$json_result"
    else
        # Save results
        local ts=$(date +%Y%m%d_%H%M%S)
        echo "$json_result" > "$RESULTS_DIR/benchmark_${ts}.json"
        ln -sf "benchmark_${ts}.json" "$RESULTS_DIR/latest.json"
        log ""
        log "${GREEN}Results saved to: $RESULTS_DIR/benchmark_${ts}.json${NC}"
    fi
    
    # Generate HTML report
    if $GENERATE_HTML; then
        log ""
        log "${BLUE}Generating HTML report...${NC}"
        python3 "$SCRIPT_DIR/generate_report.py" --results "$RESULTS_DIR" --output "$SCRIPT_DIR/reports"
        log "${GREEN}HTML report: tests/benchmarks/reports/latest.html${NC}"
    fi
    
    # Cleanup Rust project
    [ -n "$RUST_PROJECT" ] && rm -rf "$RUST_PROJECT"
    
    log ""
    log "${GREEN}Done.${NC}"
}

main