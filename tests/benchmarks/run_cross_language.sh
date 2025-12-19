#!/bin/bash
#
# ual Cross-Language Benchmark Runner
# Compares: C, Rust, Python, ual-Go, ual-Rust, iual
#
# All benchmarks measure TOTAL TIME for equivalent workloads
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UAL_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$UAL_ROOT"

export GOPATH=$HOME/go
export PATH=$PATH:/usr/lib/go-1.22/bin:$GOPATH/bin

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
echo "                              ual Cross-Language Benchmark Suite v0.7.4                                "
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
echo ""

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# ============================================================================
# Build C benchmark (inline for equivalent workload)
# ============================================================================
echo "Building C benchmarks..."

cat > "$TMPDIR/leibniz.c" << 'CEOF'
#include <stdio.h>
int main() {
    double sum = 0.0, sign = 1.0, denom = 1.0;
    for (int i = 0; i < 1000000; i++) {
        sum += sign / denom;
        sign = -sign;
        denom += 2.0;
    }
    printf("%.16g\n", 4.0 * sum);
    return 0;
}
CEOF

cat > "$TMPDIR/mandelbrot.c" << 'CEOF'
#include <stdio.h>
int main() {
    int total = 0;
    for (int py = 0; py < 50; py++) {
        for (int px = 0; px < 50; px++) {
            double cr = -2.0 + (px * 3.0 / 50.0);
            double ci = -1.5 + (py * 3.0 / 50.0);
            double zr = 0.0, zi = 0.0;
            int iter = 0;
            while (iter < 1000) {
                double zr2 = zr * zr, zi2 = zi * zi;
                if (zr2 + zi2 > 4.0) break;
                zi = 2.0 * zr * zi + ci;
                zr = zr2 - zi2 + cr;
                iter++;
            }
            total += iter;
        }
    }
    printf("%d\n", total);
    return 0;
}
CEOF

cat > "$TMPDIR/newton.c" << 'CEOF'
#include <stdio.h>
int main() {
    double total = 0.0;
    for (int n = 1; n <= 1000; n++) {
        double x = (double)n;
        double guess = x / 2.0;
        for (int i = 0; i < 20; i++) {
            guess = (guess + x / guess) / 2.0;
        }
        total += guess;
    }
    printf("%.16g\n", total);
    return 0;
}
CEOF

cat > "$TMPDIR/stackops.c" << 'CEOF'
#include <stdio.h>
int main() {
    long long stack[10000];
    int sp = 0;
    for (int i = 0; i < 10000; i++) stack[sp++] = i;
    long long sum = 0;
    while (sp > 0) sum += stack[--sp];
    printf("%lld\n", sum);
    return 0;
}
CEOF

gcc -O3 -o "$TMPDIR/c_leibniz" "$TMPDIR/leibniz.c" 2>/dev/null
gcc -O3 -o "$TMPDIR/c_mandelbrot" "$TMPDIR/mandelbrot.c" 2>/dev/null
gcc -O3 -o "$TMPDIR/c_newton" "$TMPDIR/newton.c" 2>/dev/null
gcc -O3 -o "$TMPDIR/c_stackops" "$TMPDIR/stackops.c" 2>/dev/null

# ============================================================================
# Build Rust benchmarks (inline for equivalent workload)
# ============================================================================
echo "Building Rust benchmarks..."

mkdir -p "$TMPDIR/rust_bench/src"
cat > "$TMPDIR/rust_bench/Cargo.toml" << 'REOF'
[package]
name = "bench"
version = "0.1.0"
edition = "2021"
[profile.release]
opt-level = 3
lto = true
REOF

# Leibniz
cat > "$TMPDIR/rust_bench/src/main.rs" << 'REOF'
fn main() {
    let mut sum = 0.0f64;
    let mut sign = 1.0f64;
    let mut denom = 1.0f64;
    for _ in 0..1_000_000 {
        sum += sign / denom;
        sign = -sign;
        denom += 2.0;
    }
    println!("{}", 4.0 * sum);
}
REOF
(cd "$TMPDIR/rust_bench" && cargo build --release 2>/dev/null)
cp "$TMPDIR/rust_bench/target/release/bench" "$TMPDIR/rust_leibniz"

# Mandelbrot
cat > "$TMPDIR/rust_bench/src/main.rs" << 'REOF'
fn main() {
    let mut total = 0i64;
    for py in 0..50 {
        for px in 0..50 {
            let cr = -2.0 + (px as f64 * 3.0 / 50.0);
            let ci = -1.5 + (py as f64 * 3.0 / 50.0);
            let mut zr = 0.0f64;
            let mut zi = 0.0f64;
            let mut iter = 0;
            while iter < 1000 {
                let zr2 = zr * zr;
                let zi2 = zi * zi;
                if zr2 + zi2 > 4.0 { break; }
                zi = 2.0 * zr * zi + ci;
                zr = zr2 - zi2 + cr;
                iter += 1;
            }
            total += iter;
        }
    }
    println!("{}", total);
}
REOF
(cd "$TMPDIR/rust_bench" && cargo build --release 2>/dev/null)
cp "$TMPDIR/rust_bench/target/release/bench" "$TMPDIR/rust_mandelbrot"

# Newton
cat > "$TMPDIR/rust_bench/src/main.rs" << 'REOF'
fn main() {
    let mut total = 0.0f64;
    for n in 1..=1000 {
        let x = n as f64;
        let mut guess = x / 2.0;
        for _ in 0..20 {
            guess = (guess + x / guess) / 2.0;
        }
        total += guess;
    }
    println!("{}", total);
}
REOF
(cd "$TMPDIR/rust_bench" && cargo build --release 2>/dev/null)
cp "$TMPDIR/rust_bench/target/release/bench" "$TMPDIR/rust_newton"

# Stack ops
cat > "$TMPDIR/rust_bench/src/main.rs" << 'REOF'
fn main() {
    let mut stack: Vec<i64> = Vec::with_capacity(10000);
    for i in 0..10000i64 { stack.push(i); }
    let mut sum = 0i64;
    while let Some(v) = stack.pop() { sum += v; }
    println!("{}", sum);
}
REOF
(cd "$TMPDIR/rust_bench" && cargo build --release 2>/dev/null)
cp "$TMPDIR/rust_bench/target/release/bench" "$TMPDIR/rust_stackops"

# ============================================================================
# Python benchmarks (inline)
# ============================================================================
echo "Preparing Python benchmarks..."

cat > "$TMPDIR/py_leibniz.py" << 'PYEOF'
total, sign, denom = 0.0, 1.0, 1.0
for _ in range(1000000):
    total += sign / denom
    sign = -sign
    denom += 2.0
print(4.0 * total)
PYEOF

cat > "$TMPDIR/py_mandelbrot.py" << 'PYEOF'
total = 0
for py in range(50):
    for px in range(50):
        cr = -2.0 + (px * 3.0 / 50.0)
        ci = -1.5 + (py * 3.0 / 50.0)
        zr, zi, it = 0.0, 0.0, 0
        while it < 1000:
            zr2, zi2 = zr*zr, zi*zi
            if zr2 + zi2 > 4.0: break
            zi = 2.0*zr*zi + ci
            zr = zr2 - zi2 + cr
            it += 1
        total += it
print(total)
PYEOF

cat > "$TMPDIR/py_newton.py" << 'PYEOF'
total = 0.0
for n in range(1, 1001):
    x = float(n)
    guess = x / 2.0
    for _ in range(20):
        guess = (guess + x / guess) / 2.0
    total += guess
print(total)
PYEOF

cat > "$TMPDIR/py_stackops.py" << 'PYEOF'
stack = []
for i in range(10000): stack.append(i)
s = 0
while stack: s += stack.pop()
print(s)
PYEOF

# ============================================================================
# Build ual-Rust project
# ============================================================================
echo "Building ual-Rust runtime..."

RUST_PROJECT="$TMPDIR/ual_rust"
mkdir -p "$RUST_PROJECT/src"
cat > "$RUST_PROJECT/Cargo.toml" << EOF
[package]
name = "ual_bench"
version = "0.1.0"
edition = "2021"

[dependencies]
rual = { path = "$UAL_ROOT/rual" }
lazy_static = "1.4"

[profile.release]
opt-level = 3
lto = true
EOF
echo "fn main(){}" > "$RUST_PROJECT/src/main.rs"
(cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)

# ============================================================================
# Run benchmarks
# ============================================================================

# Get time in nanoseconds (portable)
get_time_ns() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        perl -MTime::HiRes -e 'printf("%.0f\n", Time::HiRes::time()*1000000000)'
    else
        date +%s%N
    fi
}

time_cmd() {
    local start=$(get_time_ns)
    "$@" > /dev/null 2>&1
    local end=$(get_time_ns)
    echo $(( (end - start) / 1000000 ))
}

echo ""
echo "Running benchmarks (10 iterations each, showing average)..."
echo ""

run_averaged() {
    local name=$1
    shift
    local total=0
    for i in {1..10}; do
        local ms=$(time_cmd "$@")
        total=$((total + ms))
    done
    echo $((total / 10))
}

# Arrays for results
declare -a BENCH_NAMES=("Leibniz π (1M)" "Mandelbrot 50×50" "Newton sqrt ×1000" "Stack ops (10K)")

declare -a C_TIMES
declare -a RUST_TIMES
declare -a PYTHON_TIMES
declare -a UAL_GO_TIMES
declare -a UAL_RUST_TIMES
declare -a IUAL_TIMES

echo "  [1/6] C (gcc -O3)..."
C_TIMES[0]=$(run_averaged "c_leibniz" "$TMPDIR/c_leibniz")
C_TIMES[1]=$(run_averaged "c_mandelbrot" "$TMPDIR/c_mandelbrot")
C_TIMES[2]=$(run_averaged "c_newton" "$TMPDIR/c_newton")
C_TIMES[3]=$(run_averaged "c_stackops" "$TMPDIR/c_stackops")

echo "  [2/6] Rust (release+LTO)..."
RUST_TIMES[0]=$(run_averaged "rust_leibniz" "$TMPDIR/rust_leibniz")
RUST_TIMES[1]=$(run_averaged "rust_mandelbrot" "$TMPDIR/rust_mandelbrot")
RUST_TIMES[2]=$(run_averaged "rust_newton" "$TMPDIR/rust_newton")
RUST_TIMES[3]=$(run_averaged "rust_stackops" "$TMPDIR/rust_stackops")

echo "  [3/6] Python (CPython)..."
PYTHON_TIMES[0]=$(run_averaged "py_leibniz" python3 "$TMPDIR/py_leibniz.py")
PYTHON_TIMES[1]=$(run_averaged "py_mandelbrot" python3 "$TMPDIR/py_mandelbrot.py")
PYTHON_TIMES[2]=$(run_averaged "py_newton" python3 "$TMPDIR/py_newton.py")
PYTHON_TIMES[3]=$(run_averaged "py_stackops" python3 "$TMPDIR/py_stackops.py")

echo "  [4/6] ual-Go (compiled)..."
# Compile ual to Go
./ual compile "tests/benchmarks/programs/bench_compute_leibniz.ual" -o "$TMPDIR/ual_leibniz.go" 2>/dev/null
go build -o "$TMPDIR/ual_go_leibniz" "$TMPDIR/ual_leibniz.go" 2>/dev/null
./ual compile "tests/benchmarks/programs/bench_compute_mandelbrot.ual" -o "$TMPDIR/ual_mandelbrot.go" 2>/dev/null
go build -o "$TMPDIR/ual_go_mandelbrot" "$TMPDIR/ual_mandelbrot.go" 2>/dev/null
./ual compile "tests/benchmarks/programs/bench_compute_newton.ual" -o "$TMPDIR/ual_newton.go" 2>/dev/null
go build -o "$TMPDIR/ual_go_newton" "$TMPDIR/ual_newton.go" 2>/dev/null
./ual compile "tests/benchmarks/programs/bench_stack_push_pop.ual" -o "$TMPDIR/ual_stackops.go" 2>/dev/null
go build -o "$TMPDIR/ual_go_stackops" "$TMPDIR/ual_stackops.go" 2>/dev/null

UAL_GO_TIMES[0]=$(run_averaged "ual_go_leibniz" "$TMPDIR/ual_go_leibniz")
UAL_GO_TIMES[1]=$(run_averaged "ual_go_mandelbrot" "$TMPDIR/ual_go_mandelbrot")
UAL_GO_TIMES[2]=$(run_averaged "ual_go_newton" "$TMPDIR/ual_go_newton")
UAL_GO_TIMES[3]=$(run_averaged "ual_go_stackops" "$TMPDIR/ual_go_stackops")

echo "  [5/6] ual-Rust (compiled)..."
# Compile ual to Rust
./ual compile --target rust "tests/benchmarks/programs/bench_compute_leibniz.ual" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null
(cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)
cp "$RUST_PROJECT/target/release/ual_bench" "$TMPDIR/ual_rust_leibniz"

./ual compile --target rust "tests/benchmarks/programs/bench_compute_mandelbrot.ual" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null
(cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)
cp "$RUST_PROJECT/target/release/ual_bench" "$TMPDIR/ual_rust_mandelbrot"

./ual compile --target rust "tests/benchmarks/programs/bench_compute_newton.ual" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null
(cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)
cp "$RUST_PROJECT/target/release/ual_bench" "$TMPDIR/ual_rust_newton"

./ual compile --target rust "tests/benchmarks/programs/bench_stack_push_pop.ual" -o "$RUST_PROJECT/src/main.rs" 2>/dev/null
(cd "$RUST_PROJECT" && cargo build --release 2>/dev/null)
cp "$RUST_PROJECT/target/release/ual_bench" "$TMPDIR/ual_rust_stackops"

UAL_RUST_TIMES[0]=$(run_averaged "ual_rust_leibniz" "$TMPDIR/ual_rust_leibniz")
UAL_RUST_TIMES[1]=$(run_averaged "ual_rust_mandelbrot" "$TMPDIR/ual_rust_mandelbrot")
UAL_RUST_TIMES[2]=$(run_averaged "ual_rust_newton" "$TMPDIR/ual_rust_newton")
UAL_RUST_TIMES[3]=$(run_averaged "ual_rust_stackops" "$TMPDIR/ual_rust_stackops")

echo "  [6/6] iual (interpreted)..."
IUAL_TIMES[0]=$(run_averaged "iual_leibniz" ./iual -q "tests/benchmarks/programs/bench_compute_leibniz.ual")
IUAL_TIMES[1]=$(run_averaged "iual_mandelbrot" ./iual -q "tests/benchmarks/programs/bench_compute_mandelbrot.ual")
IUAL_TIMES[2]=$(run_averaged "iual_newton" ./iual -q "tests/benchmarks/programs/bench_compute_newton.ual")
IUAL_TIMES[3]=$(run_averaged "iual_stackops" ./iual -q "tests/benchmarks/programs/bench_stack_push_pop.ual")

# ============================================================================
# Display results
# ============================================================================

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
echo "                                         BENCHMARK RESULTS                                             "
echo "                                    (all times in milliseconds)                                        "
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
echo ""
printf "%-22s %8s %8s %8s %8s %10s %8s\n" "Benchmark" "C" "Rust" "Python" "ual-Go" "ual-Rust" "iual"
echo "───────────────────────────────────────────────────────────────────────────────────────────────────────"

for i in 0 1 2 3; do
    printf "%-22s %7dms %7dms %7dms %7dms %9dms %7dms\n" \
        "${BENCH_NAMES[$i]}" \
        "${C_TIMES[$i]}" \
        "${RUST_TIMES[$i]}" \
        "${PYTHON_TIMES[$i]}" \
        "${UAL_GO_TIMES[$i]}" \
        "${UAL_RUST_TIMES[$i]}" \
        "${IUAL_TIMES[$i]}"
done

echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"

# Calculate ratios vs C
echo ""
echo "Performance Ratios (vs C baseline = 1.0x):"
echo "───────────────────────────────────────────────────────────────────────────────────────────────────────"
printf "%-22s %8s %8s %8s %8s %10s %8s\n" "Benchmark" "C" "Rust" "Python" "ual-Go" "ual-Rust" "iual"
echo "───────────────────────────────────────────────────────────────────────────────────────────────────────"

for i in 0 1 2 3; do
    c_base=${C_TIMES[$i]}
    if [ "$c_base" -gt 0 ]; then
        rust_ratio=$(echo "scale=1; ${RUST_TIMES[$i]} / $c_base" | bc)
        python_ratio=$(echo "scale=1; ${PYTHON_TIMES[$i]} / $c_base" | bc)
        ual_go_ratio=$(echo "scale=1; ${UAL_GO_TIMES[$i]} / $c_base" | bc)
        ual_rust_ratio=$(echo "scale=1; ${UAL_RUST_TIMES[$i]} / $c_base" | bc)
        iual_ratio=$(echo "scale=1; ${IUAL_TIMES[$i]} / $c_base" | bc)
        
        printf "%-22s %8s %7sx %7sx %7sx %9sx %7sx\n" \
            "${BENCH_NAMES[$i]}" \
            "1.0x" \
            "$rust_ratio" \
            "$python_ratio" \
            "$ual_go_ratio" \
            "$ual_rust_ratio" \
            "$iual_ratio"
    fi
done

echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
echo ""
echo "Environment:"
echo "  C compiler:    $(gcc --version | head -1)"
echo "  Rust compiler: $(rustc --version)"
echo "  Python:        $(python3 --version)"
echo "  Go:            $(go version | cut -d' ' -f3)"
echo ""
echo "Notes:"
echo "  - All benchmarks perform equivalent computational work"
echo "  - Times include process startup overhead"
echo "  - ual-Go/ual-Rust: ual source compiled to native via Go/Rust"
echo "  - iual: ual source interpreted at runtime"
echo ""
