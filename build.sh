#!/bin/bash
#
# ual Build Script
# Usage: ./build.sh [options]
#
# Options:
#   --clean     Remove build artifacts before building
#   --install   Install ual to $GOPATH/bin
#   --test      Run tests after building
#   --all       Build, test, and install
#

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Parse arguments
DO_CLEAN=false
DO_INSTALL=false
DO_TEST=false

for arg in "$@"; do
    case $arg in
        --clean)
            DO_CLEAN=true
            ;;
        --install)
            DO_INSTALL=true
            ;;
        --test)
            DO_TEST=true
            ;;
        --all)
            DO_CLEAN=true
            DO_TEST=true
            DO_INSTALL=true
            ;;
        --help|-h)
            echo "ual Build Script"
            echo ""
            echo "Usage: ./build.sh [options]"
            echo ""
            echo "Options:"
            echo "  --clean     Remove build artifacts before building"
            echo "  --install   Install ual to \$GOPATH/bin"
            echo "  --test      Run tests after building"
            echo "  --all       Build, test, and install"
            echo "  --help      Show this help message"
            exit 0
            ;;
        *)
            warn "Unknown option: $arg"
            ;;
    esac
done

# Check Go installation
if ! command -v go &> /dev/null; then
    error "Go is not installed. Please install Go 1.22 or later."
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
info "Using Go version: $GO_VERSION"

# Clean if requested
if [ "$DO_CLEAN" = true ]; then
    info "Cleaning build artifacts..."
    rm -f ual
    rm -f cmd/ual/ual
    rm -f examples/*.go
    info "Clean complete"
fi

# Build the compiler
info "Building ual..."

cd cmd/ual
go build -o ual .
cd "$SCRIPT_DIR"

# Copy to project root for convenience
cp cmd/ual/ual .

info "Build complete: ./ual"

# Show version
VERSION=$(cat VERSION 2>/dev/null || echo "unknown")
info "ual version: $VERSION"

# Run tests if requested
if [ "$DO_TEST" = true ]; then
    info "Running tests..."
    
    # Test the runtime library
    info "Testing runtime library..."
    go test ./... 2>&1 | grep -E "^(ok|FAIL|---)" | tail -10 || true
    
    # Test example compilation
    info "Testing example compilation..."
    PASS=0
    FAIL=0
    
    for f in examples/*.ual; do
        if [ -f "$f" ]; then
            basename=$(basename "$f" .ual)
            if ./ual compile "$f" > /dev/null 2>&1; then
                ((PASS++)) || true
            else
                warn "Failed to compile: $f"
                ((FAIL++)) || true
            fi
        fi
    done
    
    info "Examples: $PASS passed, $FAIL failed"
    
    if [ $FAIL -gt 0 ]; then
        warn "Some examples failed to compile"
    fi
fi

# Install if requested
if [ "$DO_INSTALL" = true ]; then
    info "Installing ual..."
    
    # Determine install location
    if [ -n "$GOPATH" ]; then
        INSTALL_DIR="$GOPATH/bin"
    elif [ -n "$GOBIN" ]; then
        INSTALL_DIR="$GOBIN"
    else
        INSTALL_DIR="$HOME/go/bin"
    fi
    
    mkdir -p "$INSTALL_DIR"
    cp ual "$INSTALL_DIR/"
    
    info "Installed to: $INSTALL_DIR/ual"
    
    # Check if in PATH
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        warn "$INSTALL_DIR is not in your PATH"
        echo "  Add this to your shell profile:"
        echo "    export PATH=\$PATH:$INSTALL_DIR"
    fi
fi

# Summary
echo ""
echo "=============================================="
echo "ual v$VERSION"
echo "=============================================="
echo ""
echo "Usage:"
echo "  ./ual compile <file.ual>    Compile to Go source"
echo "  ./ual build <file.ual>      Compile to executable"
echo "  ./ual run <file.ual>        Compile and run"
echo "  ./ual tokens <file.ual>     Show tokens"
echo "  ./ual ast <file.ual>        Show AST"
echo ""
echo "Options:"
echo "  -o <path>                   Output file"
echo "  -v, --verbose               Verbose output"
echo ""
echo "Examples:"
echo "  ./ual compile examples/01_fibonacci.ual"
echo "  ./ual build examples/01_fibonacci.ual"
echo "  ./ual run examples/01_fibonacci.ual"
echo ""
