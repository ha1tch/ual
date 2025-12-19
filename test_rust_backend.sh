#!/bin/bash
# ual Rust Backend Test Suite
# Tests that Rust-compiled programs produce identical output to Go-compiled programs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colours for output (disabled if not tty)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Counters
PASS=0
FAIL=0
SKIP=0
FAILED_TESTS=""

# Temp directory for test artifacts
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

log_pass() { echo -e "${GREEN}✓${NC} $1"; }
log_fail() { echo -e "${RED}✗${NC} $1"; }
log_skip() { echo -e "${YELLOW}○${NC} $1"; }
log_info() { echo -e "  $1"; }

# Portable sed -i (works on both macOS and Linux)
sed_inplace() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "$@"
    else
        sed -i "$@"
    fi
}

# Check for required tools
check_prerequisites() {
    echo "=== Checking prerequisites ==="
    
    # Check Go
    if ! command -v go &> /dev/null; then
        echo "ERROR: Go is not installed"
        exit 1
    fi
    GO_VERSION=$(go version | sed -E 's/.*go([0-9]+\.[0-9]+).*/\1/')
    echo "Go: $GO_VERSION"
    
    # Check Rust
    if ! command -v rustc &> /dev/null; then
        echo "WARNING: Rust is not installed - skipping Rust backend tests"
        return 1
    fi
    
    RUST_VERSION=$(rustc --version | sed -E 's/.*([0-9]+\.[0-9]+\.[0-9]+).*/\1/')
    RUST_MAJOR=$(echo "$RUST_VERSION" | cut -d. -f1)
    RUST_MINOR=$(echo "$RUST_VERSION" | cut -d. -f2)
    
    echo "Rust: $RUST_VERSION"
    
    # Check minimum version (1.75)
    if [ "$RUST_MAJOR" -lt 1 ] || ([ "$RUST_MAJOR" -eq 1 ] && [ "$RUST_MINOR" -lt 75 ]); then
        echo "WARNING: Rust version 1.75+ required (found $RUST_VERSION) - skipping Rust backend tests"
        return 1
    fi
    
    # Check cargo
    if ! command -v cargo &> /dev/null; then
        echo "WARNING: Cargo is not installed - skipping Rust backend tests"
        return 1
    fi
    
    CARGO_VERSION=$(cargo --version | sed -E 's/.*([0-9]+\.[0-9]+\.[0-9]+).*/\1/')
    echo "Cargo: $CARGO_VERSION"
    
    # Check ual compiler
    if [ ! -x "./ual" ]; then
        echo "Building ual compiler..."
        go build -o ual ./cmd/ual/
    fi
    
    echo ""
    return 0
}

# Set up Rust test project
setup_rust_project() {
    mkdir -p "$TEST_DIR/rust_proj/src"
    
    cat > "$TEST_DIR/rust_proj/Cargo.toml" << 'CARGO'
[package]
name = "ual_test"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "RUAL_PATH" }
lazy_static = "1.4"

[profile.dev]
opt-level = 0
debug = false
CARGO
    
    # Replace RUAL_PATH with actual path
    sed_inplace "s|RUAL_PATH|$SCRIPT_DIR/rual|g" "$TEST_DIR/rust_proj/Cargo.toml"
    
    # Pre-build dependencies
    echo "fn main() {}" > "$TEST_DIR/rust_proj/src/main.rs"
    (cd "$TEST_DIR/rust_proj" && cargo build 2>/dev/null) || {
        echo "ERROR: Failed to set up Rust test project"
        return 1
    }
    
    return 0
}

# Test a single example
test_example() {
    local ual_file="$1"
    local name=$(basename "$ual_file" .ual)
    
    # Generate and compile Go version
    if ! ./ual compile --target go "$ual_file" -o "$TEST_DIR/${name}.go" 2>/dev/null; then
        log_skip "$name (Go codegen failed)"
        SKIP=$((SKIP + 1))
        return
    fi
    
    # Build Go binary
    if ! go build -o "$TEST_DIR/${name}_go" "$TEST_DIR/${name}.go" 2>/dev/null; then
        log_skip "$name (Go build failed)"
        SKIP=$((SKIP + 1))
        return
    fi
    
    # Run Go binary and capture output
    GO_OUTPUT=$("$TEST_DIR/${name}_go" 2>&1) || true
    
    # Generate Rust version
    if ! ./ual compile --target rust "$ual_file" -o "$TEST_DIR/rust_proj/src/main.rs" 2>/dev/null; then
        log_fail "$name (Rust codegen failed)"
        FAIL=$((FAIL + 1))
        FAILED_TESTS="$FAILED_TESTS $name(codegen)"
        return
    fi
    
    # Build Rust binary
    if ! (cd "$TEST_DIR/rust_proj" && cargo build 2>/dev/null); then
        log_fail "$name (Rust build failed)"
        FAIL=$((FAIL + 1))
        FAILED_TESTS="$FAILED_TESTS $name(build)"
        return
    fi
    
    # Run Rust binary and capture output
    RUST_OUTPUT=$("$TEST_DIR/rust_proj/target/debug/ual_test" 2>&1) || true
    
    # Compare outputs
    if [ "$GO_OUTPUT" = "$RUST_OUTPUT" ]; then
        log_pass "$name"
        PASS=$((PASS + 1))
    else
        log_fail "$name (output mismatch)"
        FAIL=$((FAIL + 1))
        FAILED_TESTS="$FAILED_TESTS $name(output)"
        
        # Show diff for debugging
        echo "  Go output:"
        echo "$GO_OUTPUT" | head -5 | sed 's/^/    /'
        echo "  Rust output:"
        echo "$RUST_OUTPUT" | head -5 | sed 's/^/    /'
    fi
}

# Main
main() {
    echo "ual Rust Backend Test Suite"
    echo "============================"
    echo ""
    
    if ! check_prerequisites; then
        echo ""
        echo "Rust backend tests skipped (prerequisites not met)"
        exit 0
    fi
    
    echo "=== Setting up Rust test project ==="
    if ! setup_rust_project; then
        echo "Failed to set up Rust test project"
        exit 1
    fi
    echo "Done"
    echo ""
    
    echo "=== Running tests ==="
    for ual_file in examples/*.ual; do
        test_example "$ual_file"
    done
    
    echo ""
    echo "=== Summary ==="
    TOTAL=$((PASS + FAIL + SKIP))
    echo "Passed: $PASS/$TOTAL"
    echo "Failed: $FAIL/$TOTAL"
    echo "Skipped: $SKIP/$TOTAL"
    
    if [ -n "$FAILED_TESTS" ]; then
        echo ""
        echo "Failed tests:$FAILED_TESTS"
    fi
    
    if [ $FAIL -gt 0 ]; then
        exit 1
    fi
    exit 0
}

main "$@"
