#!/usr/bin/env python3
"""
UAL Compute Benchmarks - Python Reference Implementation

Run: python3 python_bench.py
"""

import math
import time
from typing import Callable, Any

# =============================================================================
# MANDELBROT - Escape iteration for a single point
# =============================================================================

def compute_mandelbrot(cr: float, ci: float) -> float:
    zr, zi = 0.0, 0.0
    max_iter = 1000.0
    escape = 4.0
    
    iter = 0.0
    while iter < max_iter:
        zr2 = zr * zr
        zi2 = zi * zi
        if zr2 + zi2 > escape:
            return iter
        zi = 2 * zr * zi + ci
        zr = zr2 - zi2 + cr
        iter += 1
    return max_iter

# =============================================================================
# INTEGRATE - Trapezoidal integration of x²
# =============================================================================

def compute_integrate(a: float, b: float, n: float) -> float:
    h = (b - a) / n
    sum_val = (a * a) / 2.0
    
    i = 1.0
    while i < n:
        x = a + i * h
        sum_val += x * x
        i += 1
    
    sum_val += (b * b) / 2.0
    return h * sum_val

# =============================================================================
# LEIBNIZ - π calculation via Leibniz series
# =============================================================================

def compute_leibniz(terms: float) -> float:
    sum_val = 0.0
    sign = 1.0
    denom = 1.0
    
    i = 0.0
    while i < terms:
        sum_val += sign / denom
        sign = -sign
        denom += 2.0
        i += 1
    return 4.0 * sum_val

# =============================================================================
# NEWTON - Square root via Newton-Raphson iteration
# =============================================================================

def compute_newton(x: float) -> float:
    guess = x / 2
    for _ in range(20):
        guess = (guess + x / guess) / 2
    return guess

# =============================================================================
# ARRAY SUM - Sum with local buffer
# =============================================================================

def compute_array_sum(n: int) -> int:
    buf = [0] * 100
    
    for i in range(n):
        buf[i] = i + 1
    
    sum_val = 0
    for i in range(n):
        sum_val += buf[i]
    return sum_val

# =============================================================================
# DP FIBONACCI - Dynamic programming with array
# =============================================================================

def compute_dp_fib(n: int) -> int:
    dp = [0] * 100
    dp[0] = 0
    dp[1] = 1
    
    for i in range(2, n + 1):
        dp[i] = dp[i-1] + dp[i-2]
    return dp[n]

# =============================================================================
# MATH FUNCTIONS - Tests math library call overhead
# =============================================================================

def compute_math_ops(x: float) -> float:
    return math.sqrt(x) + math.sin(x) + math.cos(x) + math.log(x + 1)

# =============================================================================
# BENCHMARK RUNNER
# =============================================================================

def run_benchmark(name: str, fn: Callable, args: tuple, iterations: int) -> float:
    """Run benchmark and return ns/op"""
    # Warmup
    for _ in range(min(1000, iterations // 10)):
        fn(*args)
    
    start = time.perf_counter_ns()
    for _ in range(iterations):
        fn(*args)
    end = time.perf_counter_ns()
    
    ns_per_op = (end - start) / iterations
    print(f"{name:<30} {ns_per_op:>12.2f} ns/op")
    return ns_per_op

def main():
    print("=" * 78)
    print("PYTHON REFERENCE BENCHMARKS (CPython)")
    print("=" * 78)
    print()
    
    # Adjust iterations based on expected runtime
    # Python is ~50-100x slower than C/Go, so reduce iterations
    
    run_benchmark("Mandelbrot", compute_mandelbrot, (0.25, 0.5), 5000)
    run_benchmark("Integrate", compute_integrate, (0.0, 1.0, 1000.0), 10000)
    run_benchmark("Leibniz", compute_leibniz, (100000.0,), 100)
    run_benchmark("Newton", compute_newton, (2.0,), 1000000)
    run_benchmark("ArraySum", compute_array_sum, (50,), 500000)
    run_benchmark("DPFib", compute_dp_fib, (40,), 500000)
    run_benchmark("MathOps", compute_math_ops, (2.0,), 5000000)
    
    print()

if __name__ == "__main__":
    main()
