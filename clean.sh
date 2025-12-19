#!/bin/bash
# Clean generated files from the ual distribution

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

# Remove generated .rs files in examples
count=0
for f in examples/*.rs; do
    if [ -f "$f" ]; then
        rm "$f"
        count=$((count + 1))
    fi
done
echo "  Removed $count generated .rs files from examples/"

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

# Remove Rust target directories
if [ -d "target" ]; then
    rm -rf "target"
    echo "  Removed ./target/ (Rust build cache)"
fi

# Remove cross-language benchmark Rust build artifacts
if [ -d "tests/benchmarks/cross_language/rust/target" ]; then
    rm -rf "tests/benchmarks/cross_language/rust/target"
    echo "  Removed tests/benchmarks/cross_language/rust/target/"
fi

# Remove cross-language C benchmark binary
if [ -f "tests/benchmarks/cross_language/c/bench" ]; then
    rm "tests/benchmarks/cross_language/c/bench"
    echo "  Removed tests/benchmarks/cross_language/c/bench"
fi

# Remove any Rust binaries at project root (numbered examples)
count=0
for f in [0-9][0-9][0-9]_*; do
    if [ -f "$f" ] && [ -x "$f" ]; then
        rm "$f"
        count=$((count + 1))
    fi
done
if [ $count -gt 0 ]; then
    echo "  Removed $count Rust binaries from project root"
fi

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

# Remove the iual binary if it exists at project root
if [ -f "./iual" ]; then
    rm "./iual"
    echo "  Removed ./iual binary"
fi

# Remove cmd/iual/iual binary if present
if [ -f "cmd/iual/iual" ]; then
    rm "cmd/iual/iual"
    echo "  Removed cmd/iual/iual binary"
fi

# Remove benchmark data and reports
rm -f benchmarks/*.json benchmarks/*.csv benchmarks/*.html 2>/dev/null
rm -rf benchmarks/reports 2>/dev/null
echo "  Cleaned benchmark data and reports"

# Remove old test benchmark results (keep latest)
count=0
if [ -d "tests/benchmarks/results" ]; then
    # Find what latest.json points to
    latest_json=""
    if [ -L "tests/benchmarks/results/latest.json" ]; then
        latest_json=$(readlink "tests/benchmarks/results/latest.json")
    fi
    
    for f in tests/benchmarks/results/benchmark*.json; do
        if [ -f "$f" ]; then
            fname=$(basename "$f")
            if [ "$fname" != "$latest_json" ]; then
                rm "$f"
                count=$((count + 1))
            fi
        fi
    done
fi
if [ $count -gt 0 ]; then
    echo "  Removed $count old benchmark results (kept latest)"
fi

# Remove old test benchmark reports (keep latest)
count=0
if [ -d "tests/benchmarks/reports" ]; then
    # Find what latest.html points to
    latest_html=""
    if [ -L "tests/benchmarks/reports/latest.html" ]; then
        latest_html=$(readlink "tests/benchmarks/reports/latest.html")
    fi
    
    for f in tests/benchmarks/reports/report*.html; do
        if [ -f "$f" ]; then
            fname=$(basename "$f")
            if [ "$fname" != "$latest_html" ]; then
                rm "$f"
                count=$((count + 1))
            fi
        fi
    done
fi
if [ $count -gt 0 ]; then
    echo "  Removed $count old benchmark reports (kept latest)"
fi

# Remove any temp files
find . -name "*.tmp" -delete 2>/dev/null || true
find . -name ".DS_Store" -delete 2>/dev/null || true
find . -name "*.bak" -delete 2>/dev/null || true

# Remove __MACOSX directory if present
if [ -d "__MACOSX" ]; then
    rm -rf "__MACOSX"
    echo "  Removed __MACOSX/"
fi

# Clean Go build cache for this project
go clean 2>/dev/null || true

echo "Done."