#!/bin/bash
# =============================================================================
# ual Correctness Test Suite
# =============================================================================
#
# Tests all 92 ual examples across Go, Rust, and iual backends.
# Compares output against expected results in expected/ directory.
#
# Usage:
#   ./run_all.sh [OPTIONS]
#
# Options:
#   --go              Test Go backend only
#   --rust            Test Rust backend only
#   --iual            Test interpreter only
#   --all             Test all backends (default if none specified)
#   --update          Update expected outputs from Go backend
#   --json            Output results as JSON
#   --quiet           Only show failures
#   --verbose         Show diffs for failures
#   --save            Save results to results/ directory
#   --example NAME    Test single example (e.g., --example 001_fibonacci)
#   -h, --help        Show this help
#
# Examples:
#   ./run_all.sh --all                    # Test all backends
#   ./run_all.sh --go --rust --quiet      # Test Go+Rust, show failures only
#   ./run_all.sh --iual --verbose         # Test iual with diffs
#   ./run_all.sh --update                 # Regenerate expected outputs
#   ./run_all.sh --example 041_compute_leibniz --verbose
#
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
EXPECTED_DIR="$SCRIPT_DIR/expected"
RESULTS_DIR="$SCRIPT_DIR/results"

cd "$PROJECT_DIR"

# =============================================================================
# Configuration
# =============================================================================

TEST_GO=false
TEST_RUST=false
TEST_IUAL=false
UPDATE_EXPECTED=false
JSON_OUTPUT=false
QUIET=false
VERBOSE=false
SAVE_RESULTS=false
SINGLE_EXAMPLE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --go)       TEST_GO=true; shift ;;
        --rust)     TEST_RUST=true; shift ;;
        --iual)     TEST_IUAL=true; shift ;;
        --all)      TEST_GO=true; TEST_RUST=true; TEST_IUAL=true; shift ;;
        --update)   UPDATE_EXPECTED=true; shift ;;
        --json)     JSON_OUTPUT=true; shift ;;
        --quiet)    QUIET=true; shift ;;
        --verbose)  VERBOSE=true; shift ;;
        --save)     SAVE_RESULTS=true; shift ;;
        --example)  SINGLE_EXAMPLE="$2"; shift 2 ;;
        -h|--help)
            head -35 "$0" | tail -30
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Default to all if none specified
if ! $TEST_GO && ! $TEST_RUST && ! $TEST_IUAL && ! $UPDATE_EXPECTED; then
    TEST_GO=true
    TEST_RUST=true
    TEST_IUAL=true
fi

# =============================================================================
# Colours (disabled for non-TTY or JSON output)
# =============================================================================

if [ -t 1 ] && ! $JSON_OUTPUT; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' BOLD='' NC=''
fi

# =============================================================================
# Utility Functions
# =============================================================================

log_info() {
    if ! $JSON_OUTPUT && ! $QUIET; then
        echo -e "${BLUE}$1${NC}"
    fi
}

log_header() {
    if ! $JSON_OUTPUT && ! $QUIET; then
        echo -e "\n${BOLD}=== $1 ===${NC}\n"
    fi
}

ensure_tools() {
    if [ ! -x "./ual" ]; then
        log_info "Building ual compiler..."
        go build -o ual ./cmd/ual/ 2>/dev/null || {
            echo "Error: Failed to build ual compiler" >&2
            exit 1
        }
    fi
    if [ ! -x "./iual" ]; then
        log_info "Building iual interpreter..."
        go build -o iual ./cmd/iual/ 2>/dev/null || {
            echo "Error: Failed to build iual interpreter" >&2
            exit 1
        }
    fi
}

# =============================================================================
# Rust Project Setup
# =============================================================================

RUST_PROJECT=""
RUST_AVAILABLE=false

setup_rust() {
    if ! command -v rustc &>/dev/null; then
        log_info "Rust not installed, skipping Rust tests"
        return 1
    fi
    
    local rust_version=$(rustc --version | sed -E 's/.*([0-9]+\.[0-9]+)\.[0-9]+.*/\1/')
    local rust_major=$(echo "$rust_version" | cut -d. -f1)
    local rust_minor=$(echo "$rust_version" | cut -d. -f2)
    
    if [ "$rust_major" -lt 1 ] || ([ "$rust_major" -eq 1 ] && [ "$rust_minor" -lt 75 ]); then
        log_info "Rust 1.75+ required (found $rust_version), skipping Rust tests"
        return 1
    fi
    
    RUST_PROJECT=$(mktemp -d)
    mkdir -p "$RUST_PROJECT/src"
    
    cat > "$RUST_PROJECT/Cargo.toml" << EOF
[package]
name = "ual_test"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "$PROJECT_DIR/rual" }
lazy_static = "1.4"

[profile.dev]
opt-level = 0
debug = false
EOF
    
    # Pre-compile dependencies (silent)
    echo "fn main() {}" > "$RUST_PROJECT/src/main.rs"
    if (cd "$RUST_PROJECT" && cargo build 2>/dev/null); then
        RUST_AVAILABLE=true
        return 0
    else
        log_info "Failed to set up Rust project, skipping Rust tests"
        rm -rf "$RUST_PROJECT"
        RUST_PROJECT=""
        return 1
    fi
}

cleanup_rust() {
    if [ -n "$RUST_PROJECT" ] && [ -d "$RUST_PROJECT" ]; then
        rm -rf "$RUST_PROJECT"
    fi
}

trap cleanup_rust EXIT

# =============================================================================
# Test Execution
# =============================================================================

# Global variables for test output
TEST_OUTPUT=""
TEST_EXPECTED=""

# Run a single test and return status
# Args: $1=ual_file, $2=backend (go|rust|iual)
# Returns: pass|fail:reason|skip:reason
# Sets: TEST_OUTPUT (actual output), TEST_EXPECTED (expected output)
run_single_test() {
    local ual_file="$1"
    local backend="$2"
    local name=$(basename "$ual_file" .ual)
    local expected_file="$EXPECTED_DIR/${name}.txt"
    
    TEST_OUTPUT=""
    TEST_EXPECTED=""
    
    # Check expected exists
    if [ ! -f "$expected_file" ]; then
        echo "skip:no_expected"
        return
    fi
    
    TEST_EXPECTED=$(cat "$expected_file")
    
    case "$backend" in
        go)
            if TEST_OUTPUT=$(./ual -q run "$ual_file" 2>&1); then
                if [ "$TEST_OUTPUT" = "$TEST_EXPECTED" ]; then
                    echo "pass"
                else
                    echo "fail:output_mismatch"
                fi
            else
                echo "fail:execution_error"
            fi
            ;;
            
        rust)
            if [ -z "$RUST_PROJECT" ] || ! $RUST_AVAILABLE; then
                echo "skip:no_rust"
                return
            fi
            
            # Generate Rust code
            if ! ./ual compile --target rust "$ual_file" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null; then
                echo "fail:codegen_error"
                return
            fi
            
            # Compile
            if ! (cd "$RUST_PROJECT" && cargo build 2>/dev/null); then
                echo "fail:compile_error"
                return
            fi
            
            # Run
            if TEST_OUTPUT=$("$RUST_PROJECT/target/debug/ual_test" 2>&1); then
                if [ "$TEST_OUTPUT" = "$TEST_EXPECTED" ]; then
                    echo "pass"
                else
                    echo "fail:output_mismatch"
                fi
            else
                echo "fail:execution_error"
            fi
            ;;
            
        iual)
            if TEST_OUTPUT=$(./iual -q "$ual_file" 2>&1); then
                if [ "$TEST_OUTPUT" = "$TEST_EXPECTED" ]; then
                    echo "pass"
                else
                    echo "fail:output_mismatch"
                fi
            else
                echo "fail:execution_error"
            fi
            ;;
            
        *)
            echo "skip:unknown_backend"
            ;;
    esac
}

# Show diff between expected and actual
show_diff() {
    local name="$1"
    local expected="$2"
    local actual="$3"
    
    echo -e "${YELLOW}--- Expected${NC}"
    echo "$expected" | head -10
    [ $(echo "$expected" | wc -l) -gt 10 ] && echo "  ... (truncated)"
    echo -e "${YELLOW}+++ Actual${NC}"
    echo "$actual" | head -10
    [ $(echo "$actual" | wc -l) -gt 10 ] && echo "  ... (truncated)"
    echo ""
}

# =============================================================================
# Update Expected Outputs
# =============================================================================

update_expected() {
    log_header "Updating Expected Outputs"
    ensure_tools
    mkdir -p "$EXPECTED_DIR"
    
    local count=0
    local failed=0
    
    for ual_file in examples/*.ual; do
        name=$(basename "$ual_file" .ual)
        
        if ! $QUIET; then
            echo -n "  $name... "
        fi
        
        if ./ual -q run "$ual_file" > "$EXPECTED_DIR/${name}.txt" 2>&1; then
            count=$((count + 1))
            ! $QUIET && echo -e "${GREEN}ok${NC}"
        else
            failed=$((failed + 1))
            ! $QUIET && echo -e "${RED}failed${NC}"
        fi
    done
    
    echo ""
    echo "Updated: $count"
    [ $failed -gt 0 ] && echo "Failed: $failed"
    exit 0
}

# =============================================================================
# Main Test Run
# =============================================================================

run_tests() {
    ensure_tools
    
    # Set up Rust if needed
    if $TEST_RUST; then
        log_info "Setting up Rust environment..."
        setup_rust || TEST_RUST=false
    fi
    
    # Check for expected outputs
    if [ ! -d "$EXPECTED_DIR" ] || [ -z "$(ls -A "$EXPECTED_DIR" 2>/dev/null)" ]; then
        echo "Error: No expected outputs found in $EXPECTED_DIR"
        echo "Run with --update first to generate expected outputs."
        exit 1
    fi
    
    # Determine which examples to test
    local examples=()
    if [ -n "$SINGLE_EXAMPLE" ]; then
        if [ -f "examples/${SINGLE_EXAMPLE}.ual" ]; then
            examples=("examples/${SINGLE_EXAMPLE}.ual")
        else
            echo "Error: Example not found: examples/${SINGLE_EXAMPLE}.ual"
            exit 1
        fi
    else
        for f in examples/*.ual; do
            examples+=("$f")
        done
    fi
    
    # Results tracking
    local total=0
    local go_pass=0 go_fail=0 go_skip=0
    local rust_pass=0 rust_fail=0 rust_skip=0
    local iual_pass=0 iual_fail=0 iual_skip=0
    local failed_tests=()
    
    # JSON accumulator
    local json_results="["
    local json_first=true
    
    # Header
    if ! $JSON_OUTPUT && ! $QUIET; then
        echo ""
        echo "ual Correctness Test Suite"
        echo "=========================="
        echo ""
        printf "%-35s" "Example"
        $TEST_GO && printf "%6s" "Go"
        $TEST_RUST && printf "%8s" "Rust"
        $TEST_IUAL && printf "%8s" "iual"
        echo ""
        printf "%-35s" "-------"
        $TEST_GO && printf "%6s" "----"
        $TEST_RUST && printf "%8s" "------"
        $TEST_IUAL && printf "%8s" "------"
        echo ""
    fi
    
    # Run tests
    for ual_file in "${examples[@]}"; do
        local name=$(basename "$ual_file" .ual)
        total=$((total + 1))
        
        local go_status="" rust_status="" iual_status=""
        local go_output="" rust_output="" iual_output=""
        local expected=""
        
        # Test each backend
        if $TEST_GO; then
            go_status=$(run_single_test "$ual_file" "go")
            go_output="$TEST_OUTPUT"
            expected="$TEST_EXPECTED"
            
            case "$go_status" in
                pass) go_pass=$((go_pass + 1)) ;;
                fail*) go_fail=$((go_fail + 1)); failed_tests+=("$name:go") ;;
                skip*) go_skip=$((go_skip + 1)) ;;
            esac
        fi
        
        if $TEST_RUST; then
            rust_status=$(run_single_test "$ual_file" "rust")
            rust_output="$TEST_OUTPUT"
            [ -z "$expected" ] && expected="$TEST_EXPECTED"
            
            case "$rust_status" in
                pass) rust_pass=$((rust_pass + 1)) ;;
                fail*) rust_fail=$((rust_fail + 1)); failed_tests+=("$name:rust") ;;
                skip*) rust_skip=$((rust_skip + 1)) ;;
            esac
        fi
        
        if $TEST_IUAL; then
            iual_status=$(run_single_test "$ual_file" "iual")
            iual_output="$TEST_OUTPUT"
            [ -z "$expected" ] && expected="$TEST_EXPECTED"
            
            case "$iual_status" in
                pass) iual_pass=$((iual_pass + 1)) ;;
                fail*) iual_fail=$((iual_fail + 1)); failed_tests+=("$name:iual") ;;
                skip*) iual_skip=$((iual_skip + 1)) ;;
            esac
        fi
        
        # JSON output
        if $JSON_OUTPUT; then
            $json_first || json_results+=","
            json_first=false
            json_results+="{\"name\":\"$name\""
            $TEST_GO && json_results+=",\"go\":\"$go_status\""
            $TEST_RUST && json_results+=",\"rust\":\"$rust_status\""
            $TEST_IUAL && json_results+=",\"iual\":\"$iual_status\""
            json_results+="}"
        fi
        
        # Console output
        if ! $JSON_OUTPUT; then
            local show_line=true
            local has_failure=false
            
            [[ "$go_status" == fail* ]] && has_failure=true
            [[ "$rust_status" == fail* ]] && has_failure=true
            [[ "$iual_status" == fail* ]] && has_failure=true
            
            $QUIET && ! $has_failure && show_line=false
            
            if $show_line; then
                printf "%-35s" "$name"
                
                if $TEST_GO; then
                    case "$go_status" in
                        pass)  printf "    ${GREEN}✓${NC} " ;;
                        skip*) printf "    ${YELLOW}○${NC} " ;;
                        fail*) printf "    ${RED}✗${NC} " ;;
                    esac
                fi
                
                if $TEST_RUST; then
                    case "$rust_status" in
                        pass)  printf "      ${GREEN}✓${NC} " ;;
                        skip*) printf "      ${YELLOW}○${NC} " ;;
                        fail*) printf "      ${RED}✗${NC} " ;;
                    esac
                fi
                
                if $TEST_IUAL; then
                    case "$iual_status" in
                        pass)  printf "      ${GREEN}✓${NC} " ;;
                        skip*) printf "      ${YELLOW}○${NC} " ;;
                        fail*) printf "      ${RED}✗${NC} " ;;
                    esac
                fi
                
                echo ""
                
                # Show diffs for failures in verbose mode
                if $VERBOSE && $has_failure; then
                    if [[ "$go_status" == fail:output_mismatch ]]; then
                        echo -e "  ${RED}Go output mismatch:${NC}"
                        show_diff "$name" "$expected" "$go_output" | sed 's/^/    /'
                    fi
                    if [[ "$rust_status" == fail:output_mismatch ]]; then
                        echo -e "  ${RED}Rust output mismatch:${NC}"
                        show_diff "$name" "$expected" "$rust_output" | sed 's/^/    /'
                    fi
                    if [[ "$iual_status" == fail:output_mismatch ]]; then
                        echo -e "  ${RED}iual output mismatch:${NC}"
                        show_diff "$name" "$expected" "$iual_output" | sed 's/^/    /'
                    fi
                fi
            fi
        fi
    done
    
    # JSON output
    if $JSON_OUTPUT; then
        json_results+="]"
        
        cat << EOF
{
  "timestamp": "$(date -Iseconds)",
  "total": $total,
  "backends": {
EOF
        $TEST_GO && echo "    \"go\": {\"pass\": $go_pass, \"fail\": $go_fail, \"skip\": $go_skip},"
        $TEST_RUST && echo "    \"rust\": {\"pass\": $rust_pass, \"fail\": $rust_fail, \"skip\": $rust_skip},"
        $TEST_IUAL && echo "    \"iual\": {\"pass\": $iual_pass, \"fail\": $iual_fail, \"skip\": $iual_skip}"
        cat << EOF
  },
  "results": $json_results
}
EOF
    else
        # Summary
        echo ""
        echo "=== Summary ==="
        $TEST_GO && printf "Go:   %d/%d passed" $go_pass $((go_pass + go_fail))
        $TEST_GO && [ $go_skip -gt 0 ] && printf " (%d skipped)" $go_skip
        $TEST_GO && echo ""
        
        $TEST_RUST && printf "Rust: %d/%d passed" $rust_pass $((rust_pass + rust_fail))
        $TEST_RUST && [ $rust_skip -gt 0 ] && printf " (%d skipped)" $rust_skip
        $TEST_RUST && echo ""
        
        $TEST_IUAL && printf "iual: %d/%d passed" $iual_pass $((iual_pass + iual_fail))
        $TEST_IUAL && [ $iual_skip -gt 0 ] && printf " (%d skipped)" $iual_skip
        $TEST_IUAL && echo ""
        
        # List failures
        if [ ${#failed_tests[@]} -gt 0 ]; then
            echo ""
            echo -e "${RED}Failed tests:${NC}"
            for ft in "${failed_tests[@]}"; do
                echo "  - $ft"
            done
        fi
    fi
    
    # Save results if requested
    if $SAVE_RESULTS; then
        local timestamp=$(date +%Y%m%d_%H%M%S)
        mkdir -p "$RESULTS_DIR"
        
        if $JSON_OUTPUT; then
            # JSON already printed, save would need separate capture
            :
        else
            {
                echo "ual Test Results - $timestamp"
                echo ""
                $TEST_GO && echo "Go:   $go_pass/$((go_pass + go_fail)) passed ($go_skip skipped)"
                $TEST_RUST && echo "Rust: $rust_pass/$((rust_pass + rust_fail)) passed ($rust_skip skipped)"
                $TEST_IUAL && echo "iual: $iual_pass/$((iual_pass + iual_fail)) passed ($iual_skip skipped)"
                if [ ${#failed_tests[@]} -gt 0 ]; then
                    echo ""
                    echo "Failed:"
                    for ft in "${failed_tests[@]}"; do
                        echo "  - $ft"
                    done
                fi
            } > "$RESULTS_DIR/results_${timestamp}.txt"
            echo ""
            echo "Results saved to: $RESULTS_DIR/results_${timestamp}.txt"
        fi
    fi
    
    # Exit code: fail if any tests failed
    local has_failures=false
    $TEST_GO && [ $go_fail -gt 0 ] && has_failures=true
    $TEST_RUST && [ $rust_fail -gt 0 ] && has_failures=true
    $TEST_IUAL && [ $iual_fail -gt 0 ] && has_failures=true
    
    $has_failures && exit 1
    exit 0
}

# =============================================================================
# Main Entry Point
# =============================================================================

if $UPDATE_EXPECTED; then
    update_expected
else
    run_tests
fi