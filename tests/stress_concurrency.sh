#!/bin/sh
# =============================================================================
# ual Concurrency Stress Test
# =============================================================================
#
# Runs the most stringent concurrency tests N times across specified backends.
# Builds each binary once, then runs it N times.
#
# Usage:
#   ./stress_concurrency.sh [OPTIONS]
#
# Options:
#   -n NUM      Number of iterations (default: 100)
#   -s TESTS    Comma-separated list of tests to run (e.g., "075,079")
#   --go        Test Go backend
#   --rust      Test Rust backend
#   --iual      Test iual interpreter
#   --all       Test all backends (default if none specified)
#   -q, --quiet Only show failures and summary
#   -h, --help  Show this help
#
# Examples:
#   ./stress_concurrency.sh                       # 100 iterations, all backends
#   ./stress_concurrency.sh -n 500                # 500 iterations, all backends
#   ./stress_concurrency.sh --go --rust           # 100 iterations, Go + Rust only
#   ./stress_concurrency.sh -n 50 --iual -q       # 50 iterations, iual only, quiet
#   ./stress_concurrency.sh -n 3000 -s 075,079    # 3000 iterations, only tests 075 and 079
#   ./stress_concurrency.sh -s ping_pong --go     # test only ping_pong on Go
#
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

# Configuration
ITERATIONS=100
TEST_GO=false
TEST_RUST=false
TEST_IUAL=false
QUIET=false
SELECTED_TESTS=""

# All concurrency test examples (most stringent)
ALL_CONCURRENCY_TESTS="
072_multi_producer
073_fan_out_in
074_barrier_sync
075_ping_pong
076_work_queue
077_mapreduce
078_semaphore
079_bounded_buffer
081_pipeline_stages
082_competing_workers
083_load_balancer
084_graceful_shutdown
085_resource_pool
086_local_stack_basic
087_compute_in_spawn
088_local_stack_compute
089_parallel_reduction
"

# Parse arguments
while [ $# -gt 0 ]; do
    case $1 in
        -n)         ITERATIONS="$2"; shift 2 ;;
        -s|--s)     SELECTED_TESTS="$2"; shift 2 ;;
        --go)       TEST_GO=true; shift ;;
        --rust)     TEST_RUST=true; shift ;;
        --iual)     TEST_IUAL=true; shift ;;
        --all)      TEST_GO=true; TEST_RUST=true; TEST_IUAL=true; shift ;;
        -q|--quiet) QUIET=true; shift ;;
        -h|--help)
            head -32 "$0" | tail -29
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Default: all backends
if ! $TEST_GO && ! $TEST_RUST && ! $TEST_IUAL; then
    TEST_GO=true
    TEST_RUST=true
    TEST_IUAL=true
fi

# Build test list based on selection
CONCURRENCY_TESTS=""
if [ -n "$SELECTED_TESTS" ]; then
    # Parse comma-separated list (e.g., "075,079" or "075_ping_pong,079_bounded_buffer")
    # Convert commas to newlines and process each
    for sel in $(echo "$SELECTED_TESTS" | tr ',' ' '); do
        # Find matching test(s)
        for test in $ALL_CONCURRENCY_TESTS; do
            case "$test" in
                *"$sel"*) CONCURRENCY_TESTS="$CONCURRENCY_TESTS $test" ;;
            esac
        done
    done
    if [ -z "$CONCURRENCY_TESTS" ]; then
        echo "Error: No tests matched selection '$SELECTED_TESTS'"
        echo "Available tests:"
        for t in $ALL_CONCURRENCY_TESTS; do echo "  $t"; done
        exit 1
    fi
else
    CONCURRENCY_TESTS="$ALL_CONCURRENCY_TESTS"
fi

# Colours (disabled if not tty)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BOLD=''
    NC=''
fi

log() { $QUIET || printf "%s\n" "$1"; }
log_pass() { printf "${GREEN}✓${NC} %s\n" "$1"; }
log_fail() { printf "${RED}✗${NC} %s\n" "$1"; }

# Temp directory
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Build tools if needed
ensure_tools() {
    [ -x "./ual" ] || go build -o ual ./cmd/ual/ 2>/dev/null
    [ -x "./iual" ] || go build -o iual ./cmd/iual/ 2>/dev/null
}

# Set up Rust project for compiled tests
RUST_PROJECT=""
setup_rust() {
    if ! command -v rustc >/dev/null 2>&1; then
        return 1
    fi
    if ! command -v cargo >/dev/null 2>&1; then
        return 1
    fi
    
    RUST_PROJECT="$TMPDIR/rust_proj"
    mkdir -p "$RUST_PROJECT/src"
    
    cat > "$RUST_PROJECT/Cargo.toml" << EOF
[package]
name = "ual_stress"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "$PROJECT_DIR/rual" }
lazy_static = "1.4"

[profile.release]
opt-level = 3
EOF
    
    # Pre-compile dependencies
    echo "fn main() {}" > "$RUST_PROJECT/src/main.rs"
    (cd "$RUST_PROJECT" && cargo build --release 2>/dev/null) || return 1
    return 0
}

# Build a single Go binary
# Returns path to binary or empty string on failure
build_go() {
    local name="$1"
    local ual_file="examples/${name}.ual"
    local go_file="$TMPDIR/${name}.go"
    local binary="$TMPDIR/${name}_go"
    
    ./ual compile --target go "$ual_file" -o "$go_file" 2>/dev/null || return 1
    go build -o "$binary" "$go_file" 2>/dev/null || return 1
    echo "$binary"
}

# Build a single Rust binary
# Returns path to binary or empty string on failure
build_rust() {
    local name="$1"
    local ual_file="examples/${name}.ual"
    local binary="$TMPDIR/${name}_rust"
    
    [ -z "$RUST_PROJECT" ] && return 1
    
    ./ual compile --target rust "$ual_file" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null || return 1
    (cd "$RUST_PROJECT" && cargo build --release 2>/dev/null) || return 1
    cp "$RUST_PROJECT/target/release/ual_stress" "$binary"
    echo "$binary"
}

# Run a pre-built binary N times
# Returns: "pass" or "fail:COUNT"
run_stress() {
    local binary="$1"
    local expected="$2"
    local backend="$3"
    
    local pass=0
    local fail=0
    local i=1
    
    while [ $i -le $ITERATIONS ]; do
        local output=""
        
        if [ "$backend" = "iual" ]; then
            output=$(./iual -q "$binary" 2>&1) || true
        else
            output=$("$binary" 2>&1) || true
        fi
        
        if [ "$output" = "$expected" ]; then
            pass=$((pass + 1))
        else
            fail=$((fail + 1))
            if ! $QUIET; then
                printf "    iteration %d: expected '%s', got '%s'\n" "$i" "$expected" "$output" >&2
            fi
        fi
        
        i=$((i + 1))
    done
    
    if [ $fail -eq 0 ]; then
        echo "pass"
    else
        echo "fail:$fail"
    fi
}

# Main
main() {
    printf "${BOLD}ual Concurrency Stress Test${NC}\n"
    printf "============================\n"
    printf "Iterations: %d\n" "$ITERATIONS"
    printf "Backends:  "
    $TEST_GO && printf "Go "
    $TEST_RUST && printf "Rust "
    $TEST_IUAL && printf "iual "
    printf "\n\n"
    
    ensure_tools
    
    # Set up Rust if needed
    if $TEST_RUST; then
        log "Setting up Rust backend..."
        if ! setup_rust; then
            printf "${YELLOW}Warning: Rust not available, skipping Rust tests${NC}\n"
            TEST_RUST=false
        fi
    fi
    
    # Counters
    local total_tests=0
    local total_pass=0
    local total_fail=0
    local failed_list=""
    
    # Phase 1: Build all binaries
    printf "\n${BOLD}Building binaries...${NC}\n"
    
    for name in $CONCURRENCY_TESTS; do
        [ -f "examples/${name}.ual" ] || continue
        [ -f "tests/correctness/expected/${name}.txt" ] || continue
        
        log "  Building $name..."
        
        if $TEST_GO; then
            go_bin=$(build_go "$name")
            if [ -n "$go_bin" ]; then
                eval "GO_BIN_${name}=\"$go_bin\""
            fi
        fi
        
        if $TEST_RUST; then
            rust_bin=$(build_rust "$name")
            if [ -n "$rust_bin" ]; then
                eval "RUST_BIN_${name}=\"$rust_bin\""
            fi
        fi
    done
    
    # Phase 2: Run stress tests
    printf "\n${BOLD}Running tests...${NC}\n\n"
    
    for name in $CONCURRENCY_TESTS; do
        [ -f "examples/${name}.ual" ] || continue
        
        local expected_file="tests/correctness/expected/${name}.txt"
        [ -f "$expected_file" ] || continue
        local expected=$(cat "$expected_file")
        
        log "Testing $name..."
        
        # Test Go
        if $TEST_GO; then
            eval "go_bin=\"\$GO_BIN_${name}\""
            if [ -n "$go_bin" ] && [ -x "$go_bin" ]; then
                total_tests=$((total_tests + 1))
                result=$(run_stress "$go_bin" "$expected" "go")
                if [ "$result" = "pass" ]; then
                    total_pass=$((total_pass + 1))
                    log_pass "$name (Go): $ITERATIONS/$ITERATIONS"
                else
                    fail_count=$(echo "$result" | cut -d: -f2)
                    pass_count=$((ITERATIONS - fail_count))
                    total_fail=$((total_fail + 1))
                    log_fail "$name (Go): $pass_count/$ITERATIONS ($fail_count failures)"
                    failed_list="$failed_list $name:go"
                fi
            else
                log "  $name (Go): skipped (build failed)"
            fi
        fi
        
        # Test Rust
        if $TEST_RUST; then
            eval "rust_bin=\"\$RUST_BIN_${name}\""
            if [ -n "$rust_bin" ] && [ -x "$rust_bin" ]; then
                total_tests=$((total_tests + 1))
                result=$(run_stress "$rust_bin" "$expected" "rust")
                if [ "$result" = "pass" ]; then
                    total_pass=$((total_pass + 1))
                    log_pass "$name (Rust): $ITERATIONS/$ITERATIONS"
                else
                    fail_count=$(echo "$result" | cut -d: -f2)
                    pass_count=$((ITERATIONS - fail_count))
                    total_fail=$((total_fail + 1))
                    log_fail "$name (Rust): $pass_count/$ITERATIONS ($fail_count failures)"
                    failed_list="$failed_list $name:rust"
                fi
            else
                log "  $name (Rust): skipped (build failed)"
            fi
        fi
        
        # Test iual (no build needed, pass ual file directly)
        if $TEST_IUAL; then
            total_tests=$((total_tests + 1))
            result=$(run_stress "examples/${name}.ual" "$expected" "iual")
            if [ "$result" = "pass" ]; then
                total_pass=$((total_pass + 1))
                log_pass "$name (iual): $ITERATIONS/$ITERATIONS"
            else
                fail_count=$(echo "$result" | cut -d: -f2)
                pass_count=$((ITERATIONS - fail_count))
                total_fail=$((total_fail + 1))
                log_fail "$name (iual): $pass_count/$ITERATIONS ($fail_count failures)"
                failed_list="$failed_list $name:iual"
            fi
        fi
    done
    
    # Summary
    printf "\n${BOLD}=== Summary ===${NC}\n"
    printf "Total: %d test configurations\n" "$total_tests"
    printf "Passed: ${GREEN}%d${NC}\n" "$total_pass"
    printf "Failed: ${RED}%d${NC}\n" "$total_fail"
    printf "Iterations per test: %d\n" "$ITERATIONS"
    
    if [ -n "$failed_list" ]; then
        printf "\n${RED}Failed tests:${NC}%s\n" "$failed_list"
        exit 1
    fi
    
    printf "\n${GREEN}All concurrency tests passed.${NC}\n"
    exit 0
}

main "$@"