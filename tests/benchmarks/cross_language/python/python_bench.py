#!/usr/bin/env python3
"""
ual Cross-Language Benchmarks - Python Reference

Run: python3 python_bench.py [leibniz|mandelbrot|newton|all]

Workloads match the ual benchmark programs exactly.
"""

import sys


def compute_leibniz():
    """Leibniz series for π (1M terms) - matches ual benchmark"""
    total = 0.0
    sign = 1.0
    denom = 1.0
    terms = 1_000_000
    
    for _ in range(terms):
        total += sign / denom
        sign = -sign
        denom += 2.0
    
    return 4.0 * total


def compute_mandelbrot():
    """Mandelbrot 50x50 grid - matches ual benchmark"""
    width = 50
    height = 50
    max_iter = 100
    escape = 4.0
    total = 0.0
    
    x_min, x_max = -2.0, 1.0
    y_min, y_max = -1.5, 1.5
    x_step = (x_max - x_min) / width
    y_step = (y_max - y_min) / height
    
    for py in range(height):
        ci = y_min + py * y_step
        for px in range(width):
            cr = x_min + px * x_step
            zr, zi = 0.0, 0.0
            iteration = 0
            
            while iteration < max_iter:
                zr2 = zr * zr
                zi2 = zi * zi
                if zr2 + zi2 > escape:
                    break
                zi = 2.0 * zr * zi + ci
                zr = zr2 - zi2 + cr
                iteration += 1
            
            total += iteration
    
    return total


def compute_newton():
    """Newton-Raphson sqrt for 1000 values - matches ual benchmark"""
    total = 0.0
    limit = 1000
    
    for n in range(1, limit + 1):
        guess = n / 2.0
        for _ in range(20):
            guess = (guess + n / guess) / 2.0
        total += guess
    
    return total


def main():
    which = sys.argv[1] if len(sys.argv) > 1 else "all"
    
    if which == "leibniz":
        print(f"{compute_leibniz():.10f}")
    elif which == "mandelbrot":
        print(f"{compute_mandelbrot():.0f}")
    elif which == "newton":
        print(f"{compute_newton():.10f}")
    else:
        print(f"Leibniz: {compute_leibniz():.10f}")
        print(f"Mandelbrot: {compute_mandelbrot():.0f}")
        print(f"Newton: {compute_newton():.10f}")


if __name__ == "__main__":
    main()
