#!/bin/bash
# Clean generated files from the UAL distribution

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "Cleaning generated files..."

# Remove generated .go files in examples
count=0
for f in examples/*.go; do
    if [ -f "$f" ]; then
        rm "$f"
        count=$((count + 1))
    fi
done
echo "  Removed $count generated .go files from examples/"

# Remove compiled binaries in examples (files matching *.ual basename)
count=0
for f in examples/*.ual; do
    base="${f%.ual}"
    if [ -f "$base" ]; then
        rm "$base"
        count=$((count + 1))
    fi
done
echo "  Removed $count compiled binaries from examples/"



# Remove the ual binary if it exists at project root
if [ -f "./ual" ]; then
    rm "./ual"
    echo "  Removed ./ual binary"
fi

# Remove cmd/ual/ual binary if present
if [ -f "cmd/ual/ual" ]; then
    rm "cmd/ual/ual"
    echo "  Removed cmd/ual/ual binary"
fi


 # Remove the ual binary if it exists at project root
 if [ -f "./iual" ]; then
    rm "./iual"
    echo "  Removed ./iual binary"
 fi  
  
 # Remove cmd/iual/ual binary if present
if [ -f "cmd/iual/iual" ]; then
   rm "cmd/iual/iual"
   echo "  Removed cmd/iual/iual binary"
fi 

# Remove any temp files
find . -name "*.tmp" -delete 2>/dev/null || true
find . -name ".DS_Store" -delete 2>/dev/null || true

# Remove __MACOSX directory if present
if [ -d "__MACOSX" ]; then
    rm -rf "__MACOSX"
    echo "  Removed __MACOSX/"
fi

echo "Done."
