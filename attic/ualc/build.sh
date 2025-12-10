#!/bin/bash
set -e

# Directory structure
ROOT_DIR=$(pwd)
BUILD_DIR="$ROOT_DIR/build"
SRC_DIR="$ROOT_DIR/examples"

# Create build directory if it doesn't exist
mkdir -p "$BUILD_DIR"

# Compile function
compile_ual() {
    local input_file="$1"
    local base_name=$(basename "$input_file" .ual)
    local output_file="$BUILD_DIR/${base_name}.go"
    
    echo "Compiling $input_file to $output_file..."
    go run "$ROOT_DIR/main.go" "$input_file" "$output_file"
    
    echo "Compiling $output_file with TinyGo..."
    # Customize the TinyGo target based on your hardware
    # Available targets: arduino, arduino-nano33, pico, etc.
    tinygo build -o "$BUILD_DIR/${base_name}.hex" -target arduino "$output_file"
    
    echo "Build complete: $BUILD_DIR/${base_name}.hex"
}

# If a specific file is provided, compile just that file
if [ "$1" != "" ]; then
    if [ -f "$1" ]; then
        compile_ual "$1"
    else
        echo "File not found: $1"
        exit 1
    fi
else
    # Otherwise, compile all .ual files in the examples directory
    for file in "$SRC_DIR"/*.ual; do
        if [ -f "$file" ]; then
            compile_ual "$file"
        fi
    done
fi

echo "All builds completed successfully!"