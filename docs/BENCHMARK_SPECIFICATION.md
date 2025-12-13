# UAL Benchmark Suite Specification v1.0

## Table of Contents

1. [Project Context](#project-context)
2. [Goals](#goals)
3. [Algorithm Selection](#algorithm-selection)
4. [Directory Structure](#directory-structure)
5. [UAL Implementation Requirements](#ual-implementation-requirements)
6. [Go Implementation Requirements](#go-implementation-requirements)
7. [C Implementation Requirements](#c-implementation-requirements)
8. [Python Implementation Requirements](#python-implementation-requirements)
9. [Runtime Micro-Benchmarks](#runtime-micro-benchmarks)
10. [Harness Requirements](#harness-requirements)
11. [Report Formats](#report-formats)
12. [Interactive Dashboard](#interactive-dashboard)
13. [Execution Instructions](#execution-instructions)
14. [Continuity with Legacy Benchmarks](#continuity-with-legacy-benchmarks)
15. [Implementation Priority](#implementation-priority)
16. [Notes for Implementer](#notes-for-implementer)

---

## Project Context

UAL (Universal Assembly Language) v0.7.3 is a stack-based programming language that:

- **Compiles to Go** via `cmd/ual` (the compiler)
- **Interprets directly** via `cmd/iual` (the interpreter)
- Uses **pkg/runtime** for shared stack/value types
- Supports **concurrency** via `@spawn`, blocking `take`, and goroutines
- Has **71 working example programs** in `examples/`

The codebase lives at `https://github.com/ha1tch/ual` and follows Go module conventions with `go.mod` declaring `module github.com/ha1tch/ual`.

### Current State

The interpreter was recently brought to parity with the compiler:

- Both use `pkg/runtime` types (Value, ValueStack, ScopeStack)
- `@spawn pop play` launches true goroutines in both
- `take` blocks correctly in both
- All 71 example tests pass in both compiler and interpreter

### Existing Benchmarks (Legacy)

The existing `benchmarks/` directory contains partial work:

- `compute_bench_test.go` — Go vs "ual-style" Go (simulated, not actual UAL)
- `pipeline_bench_test.go` — Stack overhead measurements
- `c/c_bench.c` — C reference for 7 algorithms
- `python/` — Empty
- `RESULTS.md` — Historical data for Mandelbrot, Integrate, Leibniz, Newton, ArraySum, DPFib, MathOps

This should be moved to `benchmarks_legacy/` and replaced with a proper suite.

### Industry Context

The Computer Language Benchmarks Game (the canonical benchmark suite) uses 10 algorithms. The kostya/benchmarks project uses 6 categories. Academic benchmark suites (plb2) use 4 algorithms. Our suite uses **12 algorithms** to provide comprehensive coverage while remaining manageable.

---

## Goals

1. **Compare ual (compiled) vs iual (interpreted)** — Quantify interpretation overhead
2. **Compare ual vs Go vs C vs Python** — Position UAL in the language performance spectrum
3. **Benchmark pkg/runtime primitives** — Stack ops, blocking Take, concurrency
4. **Track UAL version-to-version** — Preserve continuity with legacy results where algorithms overlap
5. **Demonstrate UAL's strengths** — Stack-native patterns, concurrency
6. **Provide reproducible methodology** — Automated, documented, consistent
7. **Generate interactive visualisations** — Filterable charts for exploring results

---

## Algorithm Selection

### Overview (12 Algorithms)

| # | Algorithm | Category | Input | Expected Output |
|---|-----------|----------|-------|-----------------|
| 1 | Fibonacci (recursive) | Classic Recursive | n=35 | 9227465 |
| 2 | Factorial | Classic Recursive | n=20 | 2432902008176640000 |
| 3 | Fibonacci (DP) | Classic Iterative | n=40 | 102334155 |
| 4 | Primes (Sieve) | Classic Iterative | limit=1000000 | 78498 |
| 5 | Mandelbrot | Numerical | cr=0.25, ci=0.5, max=1000 | iteration count |
| 6 | Integration | Numerical | x² from 0 to 1, n=10000 | ≈0.333333 |
| 7 | Newton-Raphson | Numerical | x=2.0, iter=20 | ≈1.41421356 |
| 8 | Quicksort | Array/Sorting | 10000 integers, seed=42 | sorted array |
| 9 | Binary Search | Array/Sorting | 1M array, 500 searches | 250 found |
| 10 | Pipeline | Concurrency | 1..1000 | 1001000 |
| 11 | Fan-out/Fan-in | Concurrency | 100 items, 4 workers | 338350 |
| 12 | RPN Calculator | Stack-Native | "3 4 + 2 * 7 /" | 2 |

### Detailed Algorithm Specifications

#### 1. Fibonacci (Naive Recursive)

- **Category**: Classic Recursive
- **Input**: n = 35
- **Output**: fib(35) = 9227465
- **Purpose**: Measures function call overhead, recursion depth
- **Implementation notes**: No memoisation, pure recursion. The naive algorithm has exponential time complexity O(2^n), which is intentional for measuring call overhead.

#### 2. Factorial

- **Category**: Classic Recursive
- **Input**: n = 20
- **Output**: 20! = 2432902008176640000
- **Purpose**: Simple recursion, integer overflow boundary
- **Implementation notes**: Use int64. Value fits within int64 range (max ~9.2e18). Verify no overflow.

#### 3. Fibonacci (Dynamic Programming)

- **Category**: Classic Iterative
- **Input**: n = 40
- **Output**: fib(40) = 102334155
- **Purpose**: Array access patterns, loop performance
- **Implementation notes**: Use fixed-size array [100]int64, bottom-up DP. This is O(n) time, O(n) space.
- **Legacy continuity**: Matches existing DPFib benchmark

#### 4. Primes (Sieve of Eratosthenes)

- **Category**: Classic Iterative
- **Input**: limit = 1,000,000
- **Output**: Count of primes ≤ 1,000,000 = 78,498
- **Purpose**: Bit/byte array manipulation, nested loops
- **Implementation notes**: Use byte array for sieve (not bit-packed for simplicity). Return count only, not the list. Standard Eratosthenes: mark multiples of each prime starting from 2.

#### 5. Mandelbrot (Single Point Escape)

- **Category**: Numerical
- **Input**: cr = 0.25, ci = 0.5, max_iter = 1000
- **Output**: Iteration count at escape (or max_iter if doesn't escape)
- **Purpose**: Floating-point computation, tight loops
- **Legacy continuity**: Matches existing Mandelbrot benchmark
- **Implementation notes**: Standard escape algorithm:
  ```
  zr, zi = 0, 0
  for iter in 0..max_iter:
      zr2 = zr * zr
      zi2 = zi * zi
      if zr2 + zi2 > 4.0: return iter
      zi = 2 * zr * zi + ci
      zr = zr2 - zi2 + cr
  return max_iter
  ```

#### 6. Numerical Integration (Trapezoidal)

- **Category**: Numerical
- **Input**: Integrate x² from a=0 to b=1, n=10,000 steps
- **Output**: ≈ 0.333333... (1/3)
- **Purpose**: Float accumulation, loop with arithmetic
- **Legacy continuity**: Matches existing Integrate benchmark (scaled from n=1000 to n=10000)
- **Implementation notes**: Trapezoidal rule:
  ```
  h = (b - a) / n
  sum = f(a)/2 + f(b)/2
  for i in 1..n-1:
      sum += f(a + i*h)
  result = h * sum
  ```
  Where f(x) = x²

#### 7. Newton-Raphson Square Root

- **Category**: Numerical
- **Input**: x = 2.0, iterations = 20
- **Output**: ≈ 1.41421356... (√2)
- **Purpose**: Iterative convergence, division
- **Legacy continuity**: Matches existing Newton benchmark
- **Implementation notes**: Fixed iteration count (not convergence-based):
  ```
  guess = x / 2
  for i in 0..20:
      guess = (guess + x/guess) / 2
  return guess
  ```

#### 8. Quicksort

- **Category**: Array/Sorting
- **Input**: Array of 10,000 pseudo-random integers (seeded for reproducibility)
- **Output**: Sorted array (verify first, middle, last elements)
- **Purpose**: Recursion with array manipulation, partition logic
- **Implementation notes**:
  - Use seed = 42 for reproducibility across all implementations
  - Simple LCG for random generation: `next = (1103515245 * current + 12345) mod 2^31`
  - Generate values in range 0..999999
  - Standard Lomuto or Hoare partition scheme
  - Verify: arr[0], arr[4999], arr[9999] after sort
  - Expected verification values will be determined by reference implementation

#### 9. Binary Search

- **Category**: Array/Sorting
- **Input**: Sorted array of 1,000,000 integers, search for 500 values
- **Output**: Count of found values = 250
- **Purpose**: Array access patterns, comparison operations
- **Implementation notes**:
  - Array contains even numbers: [0, 2, 4, 6, ..., 1999998]
  - Search targets: 250 odd numbers (won't be found) + 250 even numbers (will be found)
  - Targets: [1, 3, 5, ...499] (odds) and [0, 4000, 8000, ..., 996000] (evens, stride 4000)
  - Standard binary search returning boolean found/not-found
  - Count and return total found

#### 10. Pipeline (Producer → Transformer → Consumer)

- **Category**: Concurrency
- **Input**: Produce integers 1..1000
- **Output**: Sum of doubled values = 1001000
- **Purpose**: Channel/stack coordination, blocking Take, goroutine overhead
- **Implementation notes**:
  - Three concurrent stages connected by FIFO stacks
  - Producer: pushes 1, 2, 3, ..., 1000, then sentinel 0
  - Transformer: takes value, doubles it, pushes to next stack; stops on sentinel
  - Consumer: takes values, accumulates sum; stops on sentinel
  - Verification: sum = 2*(1+2+...+1000) = 2 * 500500 = 1001000

#### 11. Fan-Out/Fan-In

- **Category**: Concurrency
- **Input**: 100 work items, 4 workers
- **Output**: Sum of squares = Σ(i²) for i=1..100 = 338350
- **Purpose**: Spawn scaling, work distribution
- **Implementation notes**:
  - One distributor pushes work items 1..100 to work queue
  - Four workers each: take item, compute square, push to results
  - One collector accumulates all results
  - Work queue is FIFO; results queue can be LIFO
  - Use sentinel values or counter to signal completion
  - Verification: 1² + 2² + ... + 100² = 338350

#### 12. RPN Calculator

- **Category**: Stack-Native
- **Input**: Expression "3 4 + 2 * 7 /" as space-separated tokens
- **Output**: ((3+4)*2)/7 = 2 (integer division)
- **Purpose**: Native stack paradigm, Forth-style operations
- **Implementation notes**:
  - Parse space-separated tokens
  - Numbers: push to operand stack
  - Operators (+, -, *, /): pop two operands, compute, push result
  - Supported operators: + - * / (integer arithmetic)
  - Final result: pop and return single value from stack
  - This showcases UAL's natural stack-based paradigm

---

## Directory Structure

```
benchmarks/
├── README.md                      # Overview and quick-start
├── SPECIFICATION.md               # This document
├── run_all.sh                     # Master runner script
├── Makefile                       # Build targets
│
├── algorithms/                    # Algorithm implementations
│   ├── ual/                       # UAL source files
│   │   ├── 01_fib_recursive.ual
│   │   ├── 02_factorial.ual
│   │   ├── 03_fib_dp.ual
│   │   ├── 04_primes_sieve.ual
│   │   ├── 05_mandelbrot.ual
│   │   ├── 06_integrate.ual
│   │   ├── 07_newton.ual
│   │   ├── 08_quicksort.ual
│   │   ├── 09_binary_search.ual
│   │   ├── 10_pipeline.ual
│   │   ├── 11_fanout.ual
│   │   └── 12_rpn.ual
│   │
│   ├── go/                        # Pure Go implementations
│   │   ├── algorithms.go          # All algorithm functions
│   │   ├── algorithms_test.go     # Correctness tests
│   │   └── main.go                # CLI dispatcher
│   │
│   ├── c/                         # C implementations
│   │   ├── benchmarks.c           # All algorithms
│   │   ├── benchmarks.h           # Header declarations
│   │   └── Makefile
│   │
│   └── python/                    # Python implementations
│       ├── benchmarks.py          # All algorithms as functions
│       └── run.py                 # CLI runner
│
├── runtime/                       # pkg/runtime micro-benchmarks
│   ├── go.mod                     # Module for benchmarks
│   ├── stack_bench_test.go        # Push/Pop/Peek performance
│   ├── value_bench_test.go        # Value creation, conversion
│   ├── take_bench_test.go         # Blocking Take latency
│   └── concurrency_bench_test.go  # Spawn overhead, contention
│
├── harness/                       # Benchmark execution harness
│   ├── go.mod
│   ├── main.go                    # Entry point
│   ├── runner.go                  # Orchestration logic
│   ├── compiler.go                # Compile UAL → run binary
│   ├── interpreter.go             # Run via iual
│   ├── native.go                  # Run Go/C/Python
│   ├── measure.go                 # Timing, memory measurement
│   ├── verify.go                  # Output verification
│   └── report.go                  # Generate results
│
├── results/                       # Generated output
│   ├── latest.json                # Machine-readable results
│   ├── latest.md                  # Human-readable report
│   ├── dashboard.jsx              # Interactive React dashboard
│   └── history/                   # Historical runs
│       └── .gitkeep
│
└── legacy/                        # Old benchmarks (preserved)
    ├── compute_bench_test.go
    ├── pipeline_bench_test.go
    ├── c/
    ├── python/
    └── RESULTS.md
```

---

## UAL Implementation Requirements

Each UAL algorithm file must:

1. **Be self-contained** — No imports or dependencies beyond UAL builtins
2. **Print only the result** — Single line output for verification
3. **Use standard UAL constructs** — Stacks, variables, functions, spawn where appropriate
4. **Match the algorithm spec exactly** — Same input parameters, same expected output
5. **Include header comment** — Algorithm name, input, expected output

### Example: 01_fib_recursive.ual

```ual
-- Fibonacci (Naive Recursive)
-- Input: n = 35
-- Output: 9227465

func fib(n i64) i64 {
    if (n <= 1) {
        return n
    }
    var a i64 = call fib(n - 1)
    var b i64 = call fib(n - 2)
    return (a + b)
}

var result i64 = call fib(35)
print(result)
```

### Example: 03_fib_dp.ual

```ual
-- Fibonacci (Dynamic Programming)
-- Input: n = 40
-- Output: 102334155

@dp = stack.new(i64, 100)

-- Initialize: fib(0) = 0, fib(1) = 1
@dp set:0 0
@dp set:1 1

var i i64 = 2
var n i64 = 40

while (i <= n) {
    var a i64 = 0
    var b i64 = 0
    @dp get:(i - 1) let:a
    @dp get:(i - 2) let:b
    @dp set:i (a + b)
    push:i inc let:i
}

var result i64 = 0
@dp get:n let:result
print(result)
```

### Example: 05_mandelbrot.ual

```ual
-- Mandelbrot (Single Point Escape)
-- Input: cr = 0.25, ci = 0.5, max_iter = 1000
-- Output: iteration count at escape

var cr f64 = 0.25
var ci f64 = 0.5
var max_iter f64 = 1000.0
var escape f64 = 4.0

var zr f64 = 0.0
var zi f64 = 0.0
var zr2 f64 = 0.0
var zi2 f64 = 0.0
var iter f64 = 0.0
var result f64 = 0.0

while (iter < max_iter) {
    push:zr push:zr mul let:zr2
    push:zi push:zi mul let:zi2
    
    if ((zr2 + zi2) > escape) {
        push:iter let:result
        break
    }
    
    -- zi = 2 * zr * zi + ci
    push:2.0 push:zr mul push:zi mul push:ci add let:zi
    -- zr = zr2 - zi2 + cr
    push:zr2 push:zi2 sub push:cr add let:zr
    
    push:iter inc let:iter
}

if (result == 0.0) {
    push:max_iter let:result
}

print(result)
```

### Example: 10_pipeline.ual

```ual
-- Pipeline: Producer -> Transformer -> Consumer
-- Input: 1..1000
-- Output: Sum of doubled values = 1001000

@raw = stack.new(i64)
@raw perspective(FIFO)

@doubled = stack.new(i64)
@doubled perspective(FIFO)

@result = stack.new(i64)

-- Producer: push 1..1000, then 0 sentinel
@spawn < {
    var i i64 = 1
    while (i <= 1000) {
        @raw < i
        push:i inc let:i
    }
    @raw < 0
}

-- Transformer: take from raw, double, push to doubled
@spawn < {
    var n i64 = 1
    @raw take:n
    while (n != 0) {
        @doubled < (n * 2)
        @raw take:n
    }
    @doubled < 0
}

-- Consumer: take from doubled, accumulate sum
@spawn < {
    var sum i64 = 0
    var v i64 = 1
    @doubled take:v
    while (v != 0) {
        push:sum push:v add let:sum
        @doubled take:v
    }
    @result < sum
}

-- Launch all three workers
@spawn pop play
@spawn pop play
@spawn pop play

-- Wait for and print result
var total i64 = 0
@result take:total
print(total)
```

### Example: 12_rpn.ual

```ual
-- RPN Calculator
-- Input: "3 4 + 2 * 7 /"
-- Output: 2

@stack = stack.new(i64)

-- Hardcoded tokens for benchmark (no string parsing in UAL)
-- Expression: 3 4 + 2 * 7 /
-- Evaluation: 3, 4 -> + -> 7, 2 -> * -> 14, 7 -> / -> 2

@stack < 3
@stack < 4

-- Add: pop 4, pop 3, push 7
var a i64 = 0
var b i64 = 0
@stack > let:b
@stack > let:a
@stack < (a + b)

@stack < 2

-- Multiply: pop 2, pop 7, push 14
@stack > let:b
@stack > let:a
@stack < (a * b)

@stack < 7

-- Divide: pop 7, pop 14, push 2
@stack > let:b
@stack > let:a
@stack < (a / b)

-- Result
var result i64 = 0
@stack > let:result
print(result)
```

---

## Go Implementation Requirements

Each Go algorithm must:

1. Be a function taking appropriate parameters
2. Return the result (not print it)
3. Be callable from a main dispatcher
4. Use idiomatic Go (not simulate UAL patterns)
5. Include correctness test

### algorithms/go/algorithms.go

```go
package main

// FibRecursive computes fibonacci naively
func FibRecursive(n int) int64 {
    if n <= 1 {
        return int64(n)
    }
    return FibRecursive(n-1) + FibRecursive(n-2)
}

// Factorial computes n!
func Factorial(n int) int64 {
    if n <= 1 {
        return 1
    }
    return int64(n) * Factorial(n-1)
}

// FibDP computes fibonacci with dynamic programming
func FibDP(n int) int64 {
    if n <= 1 {
        return int64(n)
    }
    dp := make([]int64, n+1)
    dp[0], dp[1] = 0, 1
    for i := 2; i <= n; i++ {
        dp[i] = dp[i-1] + dp[i-2]
    }
    return dp[n]
}

// PrimesSieve counts primes up to limit using Sieve of Eratosthenes
func PrimesSieve(limit int) int {
    if limit < 2 {
        return 0
    }
    sieve := make([]bool, limit+1)
    for i := range sieve {
        sieve[i] = true
    }
    sieve[0], sieve[1] = false, false
    
    for i := 2; i*i <= limit; i++ {
        if sieve[i] {
            for j := i * i; j <= limit; j += i {
                sieve[j] = false
            }
        }
    }
    
    count := 0
    for _, isPrime := range sieve {
        if isPrime {
            count++
        }
    }
    return count
}

// Mandelbrot computes escape iteration for a single point
func Mandelbrot(cr, ci float64, maxIter int) int {
    zr, zi := 0.0, 0.0
    for iter := 0; iter < maxIter; iter++ {
        zr2 := zr * zr
        zi2 := zi * zi
        if zr2+zi2 > 4.0 {
            return iter
        }
        zi = 2*zr*zi + ci
        zr = zr2 - zi2 + cr
    }
    return maxIter
}

// Integrate computes trapezoidal integration of x^2 from a to b
func Integrate(a, b float64, n int) float64 {
    h := (b - a) / float64(n)
    sum := (a*a + b*b) / 2.0
    for i := 1; i < n; i++ {
        x := a + float64(i)*h
        sum += x * x
    }
    return h * sum
}

// Newton computes square root via Newton-Raphson
func Newton(x float64, iterations int) float64 {
    guess := x / 2
    for i := 0; i < iterations; i++ {
        guess = (guess + x/guess) / 2
    }
    return guess
}

// Quicksort sorts array in place
func Quicksort(arr []int) {
    quicksortHelper(arr, 0, len(arr)-1)
}

func quicksortHelper(arr []int, low, high int) {
    if low < high {
        p := partition(arr, low, high)
        quicksortHelper(arr, low, p-1)
        quicksortHelper(arr, p+1, high)
    }
}

func partition(arr []int, low, high int) int {
    pivot := arr[high]
    i := low - 1
    for j := low; j < high; j++ {
        if arr[j] <= pivot {
            i++
            arr[i], arr[j] = arr[j], arr[i]
        }
    }
    arr[i+1], arr[high] = arr[high], arr[i+1]
    return i + 1
}

// BinarySearch counts how many targets are found in sorted array
func BinarySearch(arr []int, targets []int) int {
    found := 0
    for _, target := range targets {
        if bsearch(arr, target) {
            found++
        }
    }
    return found
}

func bsearch(arr []int, target int) bool {
    low, high := 0, len(arr)-1
    for low <= high {
        mid := (low + high) / 2
        if arr[mid] == target {
            return true
        } else if arr[mid] < target {
            low = mid + 1
        } else {
            high = mid - 1
        }
    }
    return false
}

// Pipeline runs producer->transformer->consumer concurrency pattern
func Pipeline() int64 {
    raw := make(chan int64, 100)
    doubled := make(chan int64, 100)
    done := make(chan int64)
    
    // Producer
    go func() {
        for i := int64(1); i <= 1000; i++ {
            raw <- i
        }
        close(raw)
    }()
    
    // Transformer
    go func() {
        for n := range raw {
            doubled <- n * 2
        }
        close(doubled)
    }()
    
    // Consumer
    go func() {
        var sum int64
        for v := range doubled {
            sum += v
        }
        done <- sum
    }()
    
    return <-done
}

// FanOutFanIn distributes work across workers
func FanOutFanIn(items, workers int) int64 {
    work := make(chan int64, items)
    results := make(chan int64, items)
    
    // Start workers
    var wg sync.WaitGroup
    for w := 0; w < workers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for n := range work {
                results <- n * n
            }
        }()
    }
    
    // Distribute work
    for i := 1; i <= items; i++ {
        work <- int64(i)
    }
    close(work)
    
    // Wait for workers and close results
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var sum int64
    for r := range results {
        sum += r
    }
    return sum
}

// RPNCalculate evaluates RPN expression
func RPNCalculate(tokens []string) int64 {
    stack := make([]int64, 0, 10)
    
    for _, tok := range tokens {
        switch tok {
        case "+":
            b, a := stack[len(stack)-1], stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            stack = append(stack, a+b)
        case "-":
            b, a := stack[len(stack)-1], stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            stack = append(stack, a-b)
        case "*":
            b, a := stack[len(stack)-1], stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            stack = append(stack, a*b)
        case "/":
            b, a := stack[len(stack)-1], stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            stack = append(stack, a/b)
        default:
            n, _ := strconv.ParseInt(tok, 10, 64)
            stack = append(stack, n)
        }
    }
    return stack[0]
}
```

### algorithms/go/main.go

```go
package main

import (
    "fmt"
    "os"
    "strconv"
    "sync"
    "time"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run . <algorithm>")
        fmt.Println("Algorithms: fib_recursive, factorial, fib_dp, primes, mandelbrot,")
        fmt.Println("            integrate, newton, quicksort, binary_search, pipeline,")
        fmt.Println("            fanout, rpn")
        os.Exit(1)
    }
    
    algo := os.Args[1]
    
    switch algo {
    case "fib_recursive":
        fmt.Println(FibRecursive(35))
    case "factorial":
        fmt.Println(Factorial(20))
    case "fib_dp":
        fmt.Println(FibDP(40))
    case "primes":
        fmt.Println(PrimesSieve(1000000))
    case "mandelbrot":
        fmt.Println(Mandelbrot(0.25, 0.5, 1000))
    case "integrate":
        fmt.Printf("%.6f\n", Integrate(0.0, 1.0, 10000))
    case "newton":
        fmt.Printf("%.8f\n", Newton(2.0, 20))
    case "quicksort":
        arr := generateArray(10000, 42)
        Quicksort(arr)
        fmt.Printf("%d,%d,%d\n", arr[0], arr[4999], arr[9999])
    case "binary_search":
        arr, targets := generateSearchData()
        fmt.Println(BinarySearch(arr, targets))
    case "pipeline":
        fmt.Println(Pipeline())
    case "fanout":
        fmt.Println(FanOutFanIn(100, 4))
    case "rpn":
        tokens := []string{"3", "4", "+", "2", "*", "7", "/"}
        fmt.Println(RPNCalculate(tokens))
    default:
        fmt.Fprintf(os.Stderr, "Unknown algorithm: %s\n", algo)
        os.Exit(1)
    }
}

func generateArray(size int, seed int) []int {
    arr := make([]int, size)
    current := seed
    for i := range arr {
        current = (1103515245*current + 12345) & 0x7fffffff
        arr[i] = current % 1000000
    }
    return arr
}

func generateSearchData() ([]int, []int) {
    // Array of evens: 0, 2, 4, ..., 1999998
    arr := make([]int, 1000000)
    for i := range arr {
        arr[i] = i * 2
    }
    
    // Targets: 250 odds (not found) + 250 evens (found)
    targets := make([]int, 500)
    for i := 0; i < 250; i++ {
        targets[i] = i*2 + 1           // odd, won't be found
        targets[250+i] = i * 4000      // even, will be found
    }
    return arr, targets
}
```

---

## C Implementation Requirements

All algorithms in a single `benchmarks.c` with function declarations in `benchmarks.h`.

Compile with: `gcc -O2 -o bench benchmarks.c -lm`

Run with: `./bench <algorithm_name>`

### algorithms/c/benchmarks.h

```c
#ifndef BENCHMARKS_H
#define BENCHMARKS_H

#include <stdint.h>

int64_t fib_recursive(int n);
int64_t factorial(int n);
int64_t fib_dp(int n);
int primes_sieve(int limit);
int mandelbrot(double cr, double ci, int max_iter);
double integrate(double a, double b, int n);
double newton(double x, int iterations);
void quicksort(int *arr, int size);
int binary_search(int *arr, int arr_size, int *targets, int target_count);
int64_t pipeline(void);
int64_t fanout(int items, int workers);
int64_t rpn_calculate(void);

#endif
```

### algorithms/c/benchmarks.c

```c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <pthread.h>
#include "benchmarks.h"

int64_t fib_recursive(int n) {
    if (n <= 1) return n;
    return fib_recursive(n-1) + fib_recursive(n-2);
}

int64_t factorial(int n) {
    if (n <= 1) return 1;
    return (int64_t)n * factorial(n-1);
}

int64_t fib_dp(int n) {
    if (n <= 1) return n;
    int64_t dp[101];
    dp[0] = 0; dp[1] = 1;
    for (int i = 2; i <= n; i++) {
        dp[i] = dp[i-1] + dp[i-2];
    }
    return dp[n];
}

int primes_sieve(int limit) {
    if (limit < 2) return 0;
    char *sieve = calloc(limit + 1, 1);
    memset(sieve, 1, limit + 1);
    sieve[0] = sieve[1] = 0;
    
    for (int i = 2; i * i <= limit; i++) {
        if (sieve[i]) {
            for (int j = i * i; j <= limit; j += i) {
                sieve[j] = 0;
            }
        }
    }
    
    int count = 0;
    for (int i = 0; i <= limit; i++) {
        if (sieve[i]) count++;
    }
    free(sieve);
    return count;
}

int mandelbrot(double cr, double ci, int max_iter) {
    double zr = 0, zi = 0;
    for (int iter = 0; iter < max_iter; iter++) {
        double zr2 = zr * zr;
        double zi2 = zi * zi;
        if (zr2 + zi2 > 4.0) return iter;
        zi = 2 * zr * zi + ci;
        zr = zr2 - zi2 + cr;
    }
    return max_iter;
}

double integrate(double a, double b, int n) {
    double h = (b - a) / n;
    double sum = (a*a + b*b) / 2.0;
    for (int i = 1; i < n; i++) {
        double x = a + i * h;
        sum += x * x;
    }
    return h * sum;
}

double newton(double x, int iterations) {
    double guess = x / 2;
    for (int i = 0; i < iterations; i++) {
        guess = (guess + x / guess) / 2;
    }
    return guess;
}

static void swap(int *a, int *b) { int t = *a; *a = *b; *b = t; }

static int partition_arr(int *arr, int low, int high) {
    int pivot = arr[high];
    int i = low - 1;
    for (int j = low; j < high; j++) {
        if (arr[j] <= pivot) {
            i++;
            swap(&arr[i], &arr[j]);
        }
    }
    swap(&arr[i+1], &arr[high]);
    return i + 1;
}

static void quicksort_helper(int *arr, int low, int high) {
    if (low < high) {
        int p = partition_arr(arr, low, high);
        quicksort_helper(arr, low, p - 1);
        quicksort_helper(arr, p + 1, high);
    }
}

void quicksort(int *arr, int size) {
    quicksort_helper(arr, 0, size - 1);
}

static int bsearch_arr(int *arr, int size, int target) {
    int low = 0, high = size - 1;
    while (low <= high) {
        int mid = (low + high) / 2;
        if (arr[mid] == target) return 1;
        if (arr[mid] < target) low = mid + 1;
        else high = mid - 1;
    }
    return 0;
}

int binary_search(int *arr, int arr_size, int *targets, int target_count) {
    int found = 0;
    for (int i = 0; i < target_count; i++) {
        if (bsearch_arr(arr, arr_size, targets[i])) found++;
    }
    return found;
}

// Simplified pipeline and fanout - would need pthreads for full impl
int64_t pipeline(void) {
    int64_t sum = 0;
    for (int64_t i = 1; i <= 1000; i++) {
        sum += i * 2;
    }
    return sum;
}

int64_t fanout(int items, int workers) {
    int64_t sum = 0;
    for (int i = 1; i <= items; i++) {
        sum += (int64_t)i * i;
    }
    return sum;
}

int64_t rpn_calculate(void) {
    // "3 4 + 2 * 7 /"
    int64_t stack[10];
    int sp = 0;
    
    stack[sp++] = 3;
    stack[sp++] = 4;
    // +
    int64_t b = stack[--sp], a = stack[--sp];
    stack[sp++] = a + b;
    
    stack[sp++] = 2;
    // *
    b = stack[--sp]; a = stack[--sp];
    stack[sp++] = a * b;
    
    stack[sp++] = 7;
    // /
    b = stack[--sp]; a = stack[--sp];
    stack[sp++] = a / b;
    
    return stack[0];
}

int main(int argc, char **argv) {
    if (argc < 2) {
        printf("Usage: %s <algorithm>\n", argv[0]);
        return 1;
    }
    
    if (strcmp(argv[1], "fib_recursive") == 0) {
        printf("%ld\n", fib_recursive(35));
    } else if (strcmp(argv[1], "factorial") == 0) {
        printf("%ld\n", factorial(20));
    } else if (strcmp(argv[1], "fib_dp") == 0) {
        printf("%ld\n", fib_dp(40));
    } else if (strcmp(argv[1], "primes") == 0) {
        printf("%d\n", primes_sieve(1000000));
    } else if (strcmp(argv[1], "mandelbrot") == 0) {
        printf("%d\n", mandelbrot(0.25, 0.5, 1000));
    } else if (strcmp(argv[1], "integrate") == 0) {
        printf("%.6f\n", integrate(0.0, 1.0, 10000));
    } else if (strcmp(argv[1], "newton") == 0) {
        printf("%.8f\n", newton(2.0, 20));
    } else if (strcmp(argv[1], "quicksort") == 0) {
        int arr[10000];
        int current = 42;
        for (int i = 0; i < 10000; i++) {
            current = (1103515245 * current + 12345) & 0x7fffffff;
            arr[i] = current % 1000000;
        }
        quicksort(arr, 10000);
        printf("%d,%d,%d\n", arr[0], arr[4999], arr[9999]);
    } else if (strcmp(argv[1], "binary_search") == 0) {
        int *arr = malloc(1000000 * sizeof(int));
        for (int i = 0; i < 1000000; i++) arr[i] = i * 2;
        int targets[500];
        for (int i = 0; i < 250; i++) {
            targets[i] = i * 2 + 1;
            targets[250 + i] = i * 4000;
        }
        printf("%d\n", binary_search(arr, 1000000, targets, 500));
        free(arr);
    } else if (strcmp(argv[1], "pipeline") == 0) {
        printf("%ld\n", pipeline());
    } else if (strcmp(argv[1], "fanout") == 0) {
        printf("%ld\n", fanout(100, 4));
    } else if (strcmp(argv[1], "rpn") == 0) {
        printf("%ld\n", rpn_calculate());
    } else {
        fprintf(stderr, "Unknown algorithm: %s\n", argv[1]);
        return 1;
    }
    return 0;
}
```

---

## Python Implementation Requirements

All algorithms in `benchmarks.py` as functions. Runner in `run.py`.

### algorithms/python/benchmarks.py

```python
"""UAL Benchmark Algorithms - Python Implementation"""

def fib_recursive(n):
    """Naive recursive fibonacci"""
    if n <= 1:
        return n
    return fib_recursive(n - 1) + fib_recursive(n - 2)

def factorial(n):
    """Recursive factorial"""
    if n <= 1:
        return 1
    return n * factorial(n - 1)

def fib_dp(n):
    """Dynamic programming fibonacci"""
    if n <= 1:
        return n
    dp = [0] * (n + 1)
    dp[0], dp[1] = 0, 1
    for i in range(2, n + 1):
        dp[i] = dp[i-1] + dp[i-2]
    return dp[n]

def primes_sieve(limit):
    """Sieve of Eratosthenes"""
    if limit < 2:
        return 0
    sieve = [True] * (limit + 1)
    sieve[0] = sieve[1] = False
    for i in range(2, int(limit**0.5) + 1):
        if sieve[i]:
            for j in range(i*i, limit + 1, i):
                sieve[j] = False
    return sum(sieve)

def mandelbrot(cr, ci, max_iter):
    """Mandelbrot escape iteration"""
    zr, zi = 0.0, 0.0
    for iter in range(max_iter):
        zr2 = zr * zr
        zi2 = zi * zi
        if zr2 + zi2 > 4.0:
            return iter
        zi = 2 * zr * zi + ci
        zr = zr2 - zi2 + cr
    return max_iter

def integrate(a, b, n):
    """Trapezoidal integration of x^2"""
    h = (b - a) / n
    total = (a*a + b*b) / 2.0
    for i in range(1, n):
        x = a + i * h
        total += x * x
    return h * total

def newton(x, iterations):
    """Newton-Raphson square root"""
    guess = x / 2
    for _ in range(iterations):
        guess = (guess + x / guess) / 2
    return guess

def quicksort(arr):
    """Quicksort in place"""
    def helper(low, high):
        if low < high:
            pivot = arr[high]
            i = low - 1
            for j in range(low, high):
                if arr[j] <= pivot:
                    i += 1
                    arr[i], arr[j] = arr[j], arr[i]
            arr[i+1], arr[high] = arr[high], arr[i+1]
            p = i + 1
            helper(low, p - 1)
            helper(p + 1, high)
    helper(0, len(arr) - 1)

def binary_search(arr, targets):
    """Count found targets in sorted array"""
    def bsearch(target):
        low, high = 0, len(arr) - 1
        while low <= high:
            mid = (low + high) // 2
            if arr[mid] == target:
                return True
            elif arr[mid] < target:
                low = mid + 1
            else:
                high = mid - 1
        return False
    
    return sum(1 for t in targets if bsearch(t))

def pipeline():
    """Producer->Transformer->Consumer using generators"""
    def producer():
        for i in range(1, 1001):
            yield i
    
    def transformer(source):
        for n in source:
            yield n * 2
    
    return sum(transformer(producer()))

def fanout(items, workers):
    """Sum of squares (simplified, no real parallelism)"""
    return sum(i * i for i in range(1, items + 1))

def rpn_calculate(tokens):
    """RPN calculator"""
    stack = []
    for tok in tokens:
        if tok == '+':
            b, a = stack.pop(), stack.pop()
            stack.append(a + b)
        elif tok == '-':
            b, a = stack.pop(), stack.pop()
            stack.append(a - b)
        elif tok == '*':
            b, a = stack.pop(), stack.pop()
            stack.append(a * b)
        elif tok == '/':
            b, a = stack.pop(), stack.pop()
            stack.append(a // b)
        else:
            stack.append(int(tok))
    return stack[0]


def generate_array(size, seed):
    """Generate pseudo-random array"""
    arr = []
    current = seed
    for _ in range(size):
        current = (1103515245 * current + 12345) & 0x7fffffff
        arr.append(current % 1000000)
    return arr

def generate_search_data():
    """Generate array and targets for binary search"""
    arr = [i * 2 for i in range(1000000)]
    targets = [i * 2 + 1 for i in range(250)] + [i * 4000 for i in range(250)]
    return arr, targets
```

### algorithms/python/run.py

```python
#!/usr/bin/env python3
"""Runner for Python benchmarks"""

import sys
import benchmarks as b

def main():
    if len(sys.argv) < 2:
        print("Usage: python run.py <algorithm>")
        sys.exit(1)
    
    algo = sys.argv[1]
    
    if algo == "fib_recursive":
        print(b.fib_recursive(35))
    elif algo == "factorial":
        print(b.factorial(20))
    elif algo == "fib_dp":
        print(b.fib_dp(40))
    elif algo == "primes":
        print(b.primes_sieve(1000000))
    elif algo == "mandelbrot":
        print(b.mandelbrot(0.25, 0.5, 1000))
    elif algo == "integrate":
        print(f"{b.integrate(0.0, 1.0, 10000):.6f}")
    elif algo == "newton":
        print(f"{b.newton(2.0, 20):.8f}")
    elif algo == "quicksort":
        arr = b.generate_array(10000, 42)
        b.quicksort(arr)
        print(f"{arr[0]},{arr[4999]},{arr[9999]}")
    elif algo == "binary_search":
        arr, targets = b.generate_search_data()
        print(b.binary_search(arr, targets))
    elif algo == "pipeline":
        print(b.pipeline())
    elif algo == "fanout":
        print(b.fanout(100, 4))
    elif algo == "rpn":
        print(b.rpn_calculate(["3", "4", "+", "2", "*", "7", "/"]))
    else:
        print(f"Unknown algorithm: {algo}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
```

---

## Runtime Micro-Benchmarks

These test `pkg/runtime` primitives directly using Go's testing framework.

### runtime/go.mod

```
module github.com/ha1tch/ual/benchmarks/runtime

go 1.22

require github.com/ha1tch/ual v0.7.3

replace github.com/ha1tch/ual => ../..
```

### runtime/stack_bench_test.go

```go
package runtime_test

import (
    "testing"
    "github.com/ha1tch/ual/pkg/runtime"
)

func BenchmarkStack_Push(b *testing.B) {
    stack := runtime.NewStack(runtime.LIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Push(data)
    }
}

func BenchmarkStack_Pop(b *testing.B) {
    stack := runtime.NewStack(runtime.LIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    for i := 0; i < b.N; i++ {
        stack.Push(data)
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Pop()
    }
}

func BenchmarkStack_PushPop(b *testing.B) {
    stack := runtime.NewStack(runtime.LIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Push(data)
        stack.Pop()
    }
}

func BenchmarkStack_FIFO_PushPop(b *testing.B) {
    stack := runtime.NewStack(runtime.FIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Push(data)
        stack.Pop()
    }
}

func BenchmarkValueStack_Push(b *testing.B) {
    stack := runtime.NewValueStack()
    val := runtime.NewInt(42)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Push(val)
    }
}

func BenchmarkValueStack_Pop(b *testing.B) {
    stack := runtime.NewValueStack()
    val := runtime.NewInt(42)
    for i := 0; i < b.N; i++ {
        stack.Push(val)
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Pop()
    }
}
```

### runtime/take_bench_test.go

```go
package runtime_test

import (
    "testing"
    "time"
    "github.com/ha1tch/ual/pkg/runtime"
)

func BenchmarkTake_Immediate(b *testing.B) {
    stack := runtime.NewStack(runtime.LIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Push(data)
        stack.Take(runtime.TakeOptions{})
    }
}

func BenchmarkTake_Blocking(b *testing.B) {
    stack := runtime.NewStack(runtime.LIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        go func() {
            time.Sleep(time.Microsecond)
            stack.Push(data)
        }()
        stack.Take(runtime.TakeOptions{})
    }
}

func BenchmarkTake_WithTimeout(b *testing.B) {
    stack := runtime.NewStack(runtime.LIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stack.Push(data)
        stack.Take(runtime.TakeOptions{TimeoutMs: 100})
    }
}
```

### runtime/concurrency_bench_test.go

```go
package runtime_test

import (
    "sync"
    "testing"
    "github.com/ha1tch/ual/pkg/runtime"
)

func BenchmarkSpawn_Overhead(b *testing.B) {
    var wg sync.WaitGroup
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        wg.Add(1)
        go func() {
            wg.Done()
        }()
        wg.Wait()
    }
}

func BenchmarkPipeline_TwoStage(b *testing.B) {
    in := runtime.NewStack(runtime.FIFO, 0, 0)
    out := runtime.NewStack(runtime.FIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        wg.Add(1)
        go func() {
            defer wg.Done()
            val, _ := in.Take(runtime.TakeOptions{})
            out.Push(val)
        }()
        in.Push(data)
        out.Take(runtime.TakeOptions{})
        wg.Wait()
    }
}

func BenchmarkContention_FourWorkers(b *testing.B) {
    stack := runtime.NewStack(runtime.FIFO, 0, 0)
    data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        for w := 0; w < 4; w++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                stack.Push(data)
                stack.Pop()
            }()
        }
        wg.Wait()
    }
}
```

---

## Harness Requirements

The harness (`harness/`) orchestrates benchmark execution.

### Core Responsibilities

1. **Compile UAL files** using `cmd/ual`:
   ```bash
   ./ual compile algorithms/ual/03_fib_dp.ual -o /tmp/fib_dp
   /tmp/fib_dp
   ```

2. **Interpret UAL files** using `cmd/iual`:
   ```bash
   ./iual -q run algorithms/ual/03_fib_dp.ual
   ```

3. **Run Go implementations**:
   ```bash
   cd algorithms/go && go build -o bench . && ./bench fib_dp
   ```

4. **Run C implementations**:
   ```bash
   cd algorithms/c && make && ./bench fib_dp
   ```

5. **Run Python implementations**:
   ```bash
   cd algorithms/python && python3 run.py fib_dp
   ```

### Measurement Protocol

For each (algorithm, implementation) pair:

1. **Warm-up**: Run 2 times, discard results
2. **Measure**: Run N times (default 10, configurable via `--iterations`)
3. **Record**:
   - Wall clock time (nanoseconds)
   - Peak memory RSS (kilobytes)
   - Exit code
   - Output (for verification)
4. **Compute**:
   - Mean time
   - Standard deviation
   - Min/Max

### Output Verification

Each algorithm has an expected output. The harness:

1. Captures stdout from each run
2. Compares to expected value (exact match or within tolerance for floats)
3. Flags mismatches in the report
4. Continues execution (doesn't abort on mismatch)

### Expected Values

| Algorithm | Expected Output | Tolerance |
|-----------|-----------------|-----------|
| fib_recursive | 9227465 | exact |
| factorial | 2432902008176640000 | exact |
| fib_dp | 102334155 | exact |
| primes | 78498 | exact |
| mandelbrot | (varies by point) | exact |
| integrate | 0.333333 | ±0.000001 |
| newton | 1.41421356 | ±0.00000001 |
| quicksort | (verify sorted) | exact |
| binary_search | 250 | exact |
| pipeline | 1001000 | exact |
| fanout | 338350 | exact |
| rpn | 2 | exact |

### CLI Interface

```bash
# Run all benchmarks
./harness

# Run specific algorithm
./harness --algorithm fib_dp

# Run specific implementations
./harness --impl go,ual_compiled

# Set iteration count
./harness --iterations 20

# Output formats
./harness --format json
./harness --format markdown
./harness --format all

# Skip verification
./harness --no-verify

# Verbose output
./harness -v
```

---

## Report Formats

### JSON Schema (results/latest.json)

```json
{
  "metadata": {
    "timestamp": "2025-12-13T17:30:00Z",
    "ual_version": "0.7.3",
    "harness_version": "1.0.0",
    "iterations": 10,
    "environment": {
      "os": "linux",
      "arch": "amd64",
      "cpu": "Intel Xeon E-2324G",
      "go_version": "1.22.2",
      "gcc_version": "11.4.0",
      "python_version": "3.10.12"
    }
  },
  "algorithms": {
    "fib_recursive": {
      "expected_output": "9227465",
      "implementations": {
        "c": {
          "mean_ns": 850000000,
          "stddev_ns": 5000000,
          "min_ns": 842000000,
          "max_ns": 860000000,
          "memory_kb": 128,
          "output": "9227465",
          "correct": true
        },
        "go": {
          "mean_ns": 920000000,
          "stddev_ns": 8000000,
          "min_ns": 905000000,
          "max_ns": 935000000,
          "memory_kb": 1024,
          "output": "9227465",
          "correct": true
        },
        "ual_compiled": {
          "mean_ns": 935000000,
          "stddev_ns": 9000000,
          "min_ns": 920000000,
          "max_ns": 950000000,
          "memory_kb": 1100,
          "output": "9227465",
          "correct": true
        },
        "ual_interpreted": {
          "mean_ns": 12500000000,
          "stddev_ns": 200000000,
          "min_ns": 12100000000,
          "max_ns": 12900000000,
          "memory_kb": 2048,
          "output": "9227465",
          "correct": true
        },
        "python": {
          "mean_ns": 45000000000,
          "stddev_ns": 1000000000,
          "min_ns": 43500000000,
          "max_ns": 47000000000,
          "memory_kb": 15000,
          "output": "9227465",
          "correct": true
        }
      }
    }
    // ... other algorithms
  },
  "runtime": {
    "stack_push": {
      "mean_ns": 15,
      "stddev_ns": 1,
      "iterations": 10000000
    },
    "stack_pop": {
      "mean_ns": 12,
      "stddev_ns": 1,
      "iterations": 10000000
    },
    "take_immediate": {
      "mean_ns": 18,
      "stddev_ns": 2,
      "iterations": 1000000
    },
    "take_blocking": {
      "mean_ns": 850,
      "stddev_ns": 50,
      "iterations": 100000
    }
  },
  "summary": {
    "ual_compiled_vs_go": {
      "mean_ratio": 1.05,
      "min_ratio": 0.98,
      "max_ratio": 1.15
    },
    "ual_interpreted_vs_compiled": {
      "mean_ratio": 75,
      "min_ratio": 50,
      "max_ratio": 120
    }
  }
}
```

### Markdown Report (results/latest.md)

```markdown
# UAL Benchmark Results

**Date**: 2025-12-13
**UAL Version**: 0.7.3
**Iterations**: 10

## Environment

| Property | Value |
|----------|-------|
| OS | linux |
| Architecture | amd64 |
| CPU | Intel Xeon E-2324G |
| Go | 1.22.2 |
| GCC | 11.4.0 |
| Python | 3.10.12 |

## Executive Summary

| Comparison | Mean Ratio | Range |
|------------|------------|-------|
| ual (compiled) vs Go | 1.05x | 0.98x - 1.15x |
| ual (interpreted) vs compiled | 75x | 50x - 120x |
| Python vs Go | 85x | 40x - 150x |

## Algorithm Comparison

### Time (nanoseconds)

| Algorithm | C | Go | ual (compiled) | ual (interpreted) | Python |
|-----------|--:|---:|---------------:|------------------:|-------:|
| fib_recursive | 850M | 920M | 935M | 12.5B | 45B |
| factorial | 45 | 58 | 61 | 4.5K | 125K |
| fib_dp | 450 | 580 | 610 | 45K | 1.2M |
| primes | 12M | 15M | 15.5M | 890M | 2.1B |
| mandelbrot | 4.1K | 4.2K | 4.2K | 320K | 116K |
| integrate | 15K | 12K | 16K | 980K | 5.9M |
| newton | 53 | 7.6 | 10.2 | 850 | 938 |
| quicksort | 1.2M | 1.5M | 1.6M | 95M | 450M |
| binary_search | 45K | 52K | 54K | 3.2M | 28M |
| pipeline | N/A | 850K | 920K | 12M | 2.1M |
| fanout | N/A | 125K | 145K | 8.5M | 350K |
| rpn | 25 | 35 | 38 | 2.8K | 15K |

### Ratio vs Go Baseline

| Algorithm | C | ual (compiled) | ual (interpreted) | Python |
|-----------|--:|---------------:|------------------:|-------:|
| fib_recursive | 0.92x | 1.02x | 13.6x | 48.9x |
| factorial | 0.78x | 1.05x | 77.6x | 2155x |
| fib_dp | 0.78x | 1.05x | 77.6x | 2069x |
| ... | ... | ... | ... | ... |

## Runtime Micro-benchmarks

| Operation | Mean (ns) | Std Dev |
|-----------|----------:|--------:|
| Stack Push | 15 | ±1 |
| Stack Pop | 12 | ±1 |
| Stack Push+Pop | 25 | ±2 |
| Take (immediate) | 18 | ±2 |
| Take (blocking) | 850 | ±50 |
| Spawn overhead | 1200 | ±100 |

## Key Findings

1. **Compiled UAL ≈ Go**: Generated code performs within 5% of idiomatic Go
   for most algorithms.

2. **Interpreter overhead**: 50-120x slower than compiled, depending on
   algorithm complexity. Tight loops suffer most.

3. **Concurrency competitive**: Pipeline and fan-out patterns perform
   within 10% of native Go goroutines.

4. **When to compile vs interpret**:
   - Tasks < 1ms: Interpretation overhead dominates
   - Tasks > 100ms: Interpretation is acceptable for prototyping
   - Production: Always compile

## Verification

All outputs matched expected values. ✓
```

---

## Interactive Dashboard

### Overview

A standalone React artifact that loads `latest.json` and provides interactive visualisation of benchmark results.

### Features

| Feature | Description |
|---------|-------------|
| Implementation filter | Checkboxes to show/hide C, Go, ual_compiled, ual_interpreted, Python |
| Algorithm filter | Multi-select dropdown to focus on specific algorithms |
| Metric selector | Switch between Time (ns), Memory (KB), or Ratio view |
| Scale toggle | Linear or Logarithmic Y-axis |
| Baseline selector | "Show ratios relative to: [Go / C / ual_compiled]" |
| Sort controls | By name, by time, by ratio |

### Chart Types

1. **Grouped Bar Chart** (default)
   - X-axis: Algorithms
   - Y-axis: Time (ns) with log scale
   - Bars grouped by implementation
   - Hover tooltip with exact values

2. **Ratio Chart**
   - X-axis: Algorithms
   - Y-axis: Multiplier (1.0x baseline)
   - Horizontal line at 1.0
   - Easy comparison to baseline

3. **Interpreter Overhead Chart**
   - Dedicated view showing interpreted/compiled ratio
   - Identifies algorithms where interpretation hurts most

4. **Runtime Panel**
   - Separate section for micro-benchmarks
   - Simple horizontal bar chart

### Implementation

Single React JSX file using:
- **Recharts** for charts (available in Claude artifacts)
- **Tailwind CSS** for styling
- **useState/useMemo** for state management
- **Embedded JSON** or fetch from file

### File: results/dashboard.jsx

```jsx
import React, { useState, useMemo } from 'react';
import { 
  BarChart, Bar, XAxis, YAxis, CartesianGrid, 
  Tooltip, Legend, ResponsiveContainer, ReferenceLine 
} from 'recharts';

// Benchmark data will be embedded here by the harness
const BENCHMARK_DATA = null; // Replace with actual JSON

const IMPLEMENTATIONS = [
  { id: 'c', name: 'C', color: '#555555' },
  { id: 'go', name: 'Go', color: '#00ADD8' },
  { id: 'ual_compiled', name: 'UAL (compiled)', color: '#8B5CF6' },
  { id: 'ual_interpreted', name: 'UAL (interpreted)', color: '#A78BFA' },
  { id: 'python', name: 'Python', color: '#3776AB' },
];

const ALGORITHMS = [
  'fib_recursive', 'factorial', 'fib_dp', 'primes',
  'mandelbrot', 'integrate', 'newton', 'quicksort',
  'binary_search', 'pipeline', 'fanout', 'rpn'
];

export default function BenchmarkDashboard() {
  const [data, setData] = useState(BENCHMARK_DATA);
  const [selectedImpls, setSelectedImpls] = useState(['go', 'ual_compiled', 'ual_interpreted']);
  const [selectedAlgos, setSelectedAlgos] = useState(ALGORITHMS);
  const [logScale, setLogScale] = useState(true);
  const [viewMode, setViewMode] = useState('absolute'); // 'absolute' | 'ratio'
  const [baseline, setBaseline] = useState('go');

  // Transform data for chart
  const chartData = useMemo(() => {
    if (!data) return [];
    
    return selectedAlgos
      .filter(algo => data.algorithms[algo])
      .map(algo => {
        const row = { name: algo };
        const algoData = data.algorithms[algo].implementations;
        
        selectedImpls.forEach(impl => {
          if (algoData[impl]) {
            const value = algoData[impl].mean_ns;
            if (viewMode === 'ratio' && algoData[baseline]) {
              row[impl] = value / algoData[baseline].mean_ns;
            } else {
              row[impl] = value;
            }
          }
        });
        
        return row;
      });
  }, [data, selectedImpls, selectedAlgos, viewMode, baseline]);

  // Toggle implementation
  const toggleImpl = (impl) => {
    setSelectedImpls(prev => 
      prev.includes(impl) 
        ? prev.filter(i => i !== impl)
        : [...prev, impl]
    );
  };

  // Format large numbers
  const formatValue = (value) => {
    if (viewMode === 'ratio') return `${value.toFixed(2)}x`;
    if (value >= 1e9) return `${(value / 1e9).toFixed(1)}B`;
    if (value >= 1e6) return `${(value / 1e6).toFixed(1)}M`;
    if (value >= 1e3) return `${(value / 1e3).toFixed(1)}K`;
    return value.toFixed(0);
  };

  if (!data) {
    return (
      <div className="p-8 text-center">
        <p className="text-gray-500">No benchmark data loaded.</p>
        <p className="text-sm text-gray-400 mt-2">
          Run the benchmark harness to generate results.
        </p>
      </div>
    );
  }

  return (
    <div className="p-4 space-y-6 bg-white min-h-screen">
      {/* Header */}
      <div className="border-b pb-4">
        <h1 className="text-2xl font-bold">UAL Benchmark Dashboard</h1>
        <p className="text-gray-500">
          Version {data.metadata.ual_version} — {data.metadata.timestamp}
        </p>
      </div>

      {/* Controls */}
      <div className="flex flex-wrap gap-6">
        {/* Implementation toggles */}
        <div>
          <label className="block text-sm font-medium mb-2">Implementations</label>
          <div className="flex flex-wrap gap-2">
            {IMPLEMENTATIONS.map(impl => (
              <label 
                key={impl.id} 
                className="flex items-center gap-1 px-2 py-1 rounded border cursor-pointer hover:bg-gray-50"
                style={{ borderColor: selectedImpls.includes(impl.id) ? impl.color : '#ddd' }}
              >
                <input
                  type="checkbox"
                  checked={selectedImpls.includes(impl.id)}
                  onChange={() => toggleImpl(impl.id)}
                  className="sr-only"
                />
                <span 
                  className="w-3 h-3 rounded-sm" 
                  style={{ backgroundColor: selectedImpls.includes(impl.id) ? impl.color : '#ddd' }}
                />
                <span className="text-sm">{impl.name}</span>
              </label>
            ))}
          </div>
        </div>

        {/* View mode */}
        <div>
          <label className="block text-sm font-medium mb-2">View</label>
          <select 
            value={viewMode}
            onChange={e => setViewMode(e.target.value)}
            className="border rounded px-2 py-1"
          >
            <option value="absolute">Absolute Time</option>
            <option value="ratio">Ratio to Baseline</option>
          </select>
        </div>

        {/* Baseline (for ratio view) */}
        {viewMode === 'ratio' && (
          <div>
            <label className="block text-sm font-medium mb-2">Baseline</label>
            <select 
              value={baseline}
              onChange={e => setBaseline(e.target.value)}
              className="border rounded px-2 py-1"
            >
              {IMPLEMENTATIONS.map(impl => (
                <option key={impl.id} value={impl.id}>{impl.name}</option>
              ))}
            </select>
          </div>
        )}

        {/* Scale toggle */}
        <div>
          <label className="block text-sm font-medium mb-2">Scale</label>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={logScale}
              onChange={e => setLogScale(e.target.checked)}
              className="rounded"
            />
            <span className="text-sm">Logarithmic</span>
          </label>
        </div>
      </div>

      {/* Main Chart */}
      <div className="border rounded-lg p-4">
        <h2 className="text-lg font-semibold mb-4">
          {viewMode === 'absolute' ? 'Execution Time' : `Ratio vs ${baseline}`}
        </h2>
        <ResponsiveContainer width="100%" height={400}>
          <BarChart data={chartData} margin={{ top: 20, right: 30, left: 20, bottom: 60 }}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis 
              dataKey="name" 
              angle={-45} 
              textAnchor="end" 
              height={80}
              tick={{ fontSize: 12 }}
            />
            <YAxis 
              scale={logScale ? 'log' : 'linear'} 
              domain={logScale ? ['auto', 'auto'] : [0, 'auto']}
              tickFormatter={formatValue}
            />
            <Tooltip 
              formatter={(value, name) => [formatValue(value), name]}
              labelStyle={{ fontWeight: 'bold' }}
            />
            <Legend />
            {viewMode === 'ratio' && (
              <ReferenceLine y={1} stroke="#666" strokeDasharray="5 5" />
            )}
            {IMPLEMENTATIONS.map(impl => (
              selectedImpls.includes(impl.id) && (
                <Bar 
                  key={impl.id} 
                  dataKey={impl.id} 
                  name={impl.name}
                  fill={impl.color} 
                />
              )
            ))}
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Summary Statistics */}
      {data.summary && (
        <div className="border rounded-lg p-4">
          <h2 className="text-lg font-semibold mb-4">Summary</h2>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-purple-50 rounded p-3">
              <div className="text-sm text-purple-600">UAL (compiled) vs Go</div>
              <div className="text-2xl font-bold text-purple-700">
                {data.summary.ual_compiled_vs_go.mean_ratio.toFixed(2)}x
              </div>
              <div className="text-xs text-purple-500">
                Range: {data.summary.ual_compiled_vs_go.min_ratio.toFixed(2)}x - 
                {data.summary.ual_compiled_vs_go.max_ratio.toFixed(2)}x
              </div>
            </div>
            <div className="bg-indigo-50 rounded p-3">
              <div className="text-sm text-indigo-600">Interpreted vs Compiled</div>
              <div className="text-2xl font-bold text-indigo-700">
                {data.summary.ual_interpreted_vs_compiled.mean_ratio.toFixed(0)}x
              </div>
              <div className="text-xs text-indigo-500">
                Range: {data.summary.ual_interpreted_vs_compiled.min_ratio.toFixed(0)}x - 
                {data.summary.ual_interpreted_vs_compiled.max_ratio.toFixed(0)}x
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Runtime Micro-benchmarks */}
      {data.runtime && (
        <div className="border rounded-lg p-4">
          <h2 className="text-lg font-semibold mb-4">Runtime Micro-benchmarks</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Object.entries(data.runtime).map(([name, stats]) => (
              <div key={name} className="bg-gray-50 rounded p-3">
                <div className="text-sm text-gray-600">{name.replace(/_/g, ' ')}</div>
                <div className="text-xl font-bold">{stats.mean_ns} ns</div>
                <div className="text-xs text-gray-500">±{stats.stddev_ns}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Footer */}
      <div className="text-center text-sm text-gray-400 pt-4 border-t">
        Generated by UAL Benchmark Harness v1.0
      </div>
    </div>
  );
}
```

### Standalone HTML Export

For viewing outside Claude, the harness can generate a self-contained HTML file:

```html
<!DOCTYPE html>
<html>
<head>
  <title>UAL Benchmark Dashboard</title>
  <script src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
  <script src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
  <script src="https://unpkg.com/recharts@2/umd/Recharts.min.js"></script>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
  <div id="root"></div>
  <script>
    const BENCHMARK_DATA = /* EMBEDDED JSON */;
    // Dashboard component code here
  </script>
</body>
</html>
```

---

## Execution Instructions

### Prerequisites

```bash
# Go 1.22+
apt-get update && apt-get install -y golang-go

# GCC for C benchmarks
apt-get install -y gcc

# Python 3.10+
apt-get install -y python3

# Set up Go environment
export GOPATH=$HOME/go
export PATH=$PATH:/usr/lib/go-1.22/bin:$GOPATH/bin
export GOPROXY=https://proxy.golang.org,direct
export GONOSUMDB=*
```

### Building UAL

```bash
cd /path/to/ual

# Build compiler
go build -o ual ./cmd/ual

# Build interpreter
go build -o iual ./cmd/iual

# Verify both work
./ual --version
./iual --version
```

### Running Benchmarks

```bash
cd benchmarks

# Full suite
./run_all.sh

# Specific algorithm
./run_all.sh --algorithm fib_dp

# Specific implementations
./run_all.sh --impl go,ual_compiled

# More iterations for precision
./run_all.sh --iterations 20

# Skip slow implementations
./run_all.sh --skip python

# Generate only specific output
./run_all.sh --format json
./run_all.sh --format markdown
./run_all.sh --format dashboard
```

### Running Runtime Benchmarks

```bash
cd benchmarks/runtime

# All runtime benchmarks
go test -bench=. -benchmem

# Specific benchmark
go test -bench=BenchmarkStack_Push -benchmem

# More iterations
go test -bench=. -benchtime=5s
```

---

## Continuity with Legacy Benchmarks

### Preserved Algorithms

These algorithms maintain parameter compatibility with `benchmarks_legacy/RESULTS.md`:

| New | Legacy | Parameters | Compatible |
|-----|--------|------------|------------|
| fib_dp | DPFib | n=40 | ✓ |
| mandelbrot | Mandelbrot | cr=0.25, ci=0.5, max=1000 | ✓ |
| integrate | Integrate | a=0, b=1, n=10000 | Scaled from n=1000 |
| newton | Newton | x=2.0, iter=20 | ✓ |

### Migration Steps

1. Move `benchmarks/` to `benchmarks_legacy/`
2. Create new `benchmarks/` with this specification
3. After first run, compare results for overlapping algorithms
4. Document any regressions or improvements

---

## Implementation Priority

### Phase 1: Infrastructure (Batch 1-2)
- Create directory structure
- Set up harness skeleton
- Implement measurement utilities
- Create Makefile and run scripts

### Phase 2: UAL Implementations (Batch 3-5)
- Implement all 12 algorithms in UAL
- Test each with iual interpreter
- Verify outputs match expected values

### Phase 3: Reference Implementations (Batch 6-8)
- Go implementations with tests
- C implementations with Makefile
- Python implementations

### Phase 4: Runtime Benchmarks (Batch 9)
- Stack micro-benchmarks
- Take/blocking benchmarks
- Concurrency benchmarks

### Phase 5: Integration (Batch 10-11)
- Complete harness implementation
- JSON report generation
- Markdown report generation

### Phase 6: Dashboard (Batch 12)
- React dashboard component
- Data integration
- Standalone HTML export

### Phase 7: Validation (Batch 13-14)
- Run full benchmark suite
- Compare with legacy results
- Document findings
- Final cleanup and documentation

---

## Notes for Implementer

### Environment Setup

The UAL codebase is at `/home/claude` after extracting from zip. Use:

```bash
export GOPATH=$HOME/go
export PATH=$PATH:/usr/lib/go-1.22/bin:$GOPATH/bin
export GOPROXY=https://proxy.golang.org,direct
export GONOSUMDB=*
```

### Container Stability

Container may crash during long operations. Create backups after each batch:

```bash
zip -r backups/benchmark_batch_N.zip benchmarks/
```

### Testing UAL Programs

Before benchmarking, verify each UAL program works:

```bash
./iual -q run benchmarks/algorithms/ual/03_fib_dp.ual
# Should output: 102334155
```

### Interpreter Behaviour

- The interpreter now has true goroutines for `@spawn pop play`
- `take` blocks indefinitely if stack is empty (use with producers)
- Both compiler and interpreter use `pkg/runtime` types

### Known UAL Syntax Notes

- Comments: `-- comment` or `// comment`
- Variables: `var name type = value`
- Functions: `func name(params) return_type { body }`
- Stacks: `@name = stack.new(type)` or `@name = stack.new(type, size)`
- Push: `@stack < value`
- Pop: `@stack > let:variable` or `@stack >`
- Take (blocking): `@stack take:variable`
- Spawn: `@spawn < { block }` then `@spawn pop play`

### Verification Values

Run reference implementations first to establish expected outputs, especially for:
- quicksort (depends on LCG sequence)
- mandelbrot (verify escape iteration)
- integrate (verify precision)

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-12-13 | Initial specification |

---

*End of specification*
