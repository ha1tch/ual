/*
 * ual Compute Benchmarks - C Reference Implementation
 * 
 * Compile: gcc -O2 -o c_bench c_bench.c -lm
 * Run: ./c_bench
 */

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <time.h>
#include <math.h>

/* High-resolution timing */
static inline double get_time_ns(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts.tv_sec * 1e9 + ts.tv_nsec;
}

/* ==========================================================================
 * MANDELBROT - Escape iteration for a single point
 * ========================================================================== */

double compute_mandelbrot(double cr, double ci) {
    double zr = 0, zi = 0;
    double zr2, zi2;
    const double max_iter = 1000.0;
    const double escape = 4.0;
    
    for (double iter = 0; iter < max_iter; iter++) {
        zr2 = zr * zr;
        zi2 = zi * zi;
        if (zr2 + zi2 > escape) {
            return iter;
        }
        zi = 2 * zr * zi + ci;
        zr = zr2 - zi2 + cr;
    }
    return max_iter;
}

/* ==========================================================================
 * INTEGRATE - Trapezoidal integration of x²
 * ========================================================================== */

double compute_integrate(double a, double b, double n) {
    double h = (b - a) / n;
    double sum = (a * a) / 2.0;
    
    for (double i = 1; i < n; i++) {
        double x = a + i * h;
        sum += x * x;
    }
    
    sum += (b * b) / 2.0;
    return h * sum;
}

/* ==========================================================================
 * LEIBNIZ - π calculation via Leibniz series
 * ========================================================================== */

double compute_leibniz(double terms) {
    double sum = 0.0;
    double sign = 1.0;
    double denom = 1.0;
    
    for (double i = 0; i < terms; i++) {
        sum += sign / denom;
        sign = -sign;
        denom += 2.0;
    }
    return 4.0 * sum;
}

/* ==========================================================================
 * NEWTON - Square root via Newton-Raphson iteration
 * ========================================================================== */

double compute_newton(double x) {
    double guess = x / 2;
    for (int i = 0; i < 20; i++) {
        guess = (guess + x / guess) / 2;
    }
    return guess;
}

/* ==========================================================================
 * ARRAY SUM - Sum with local buffer
 * ========================================================================== */

int64_t compute_array_sum(int64_t n) {
    int64_t buf[100];
    
    for (int64_t i = 0; i < n; i++) {
        buf[i] = i + 1;
    }
    
    int64_t sum = 0;
    for (int64_t i = 0; i < n; i++) {
        sum += buf[i];
    }
    return sum;
}

/* ==========================================================================
 * DP FIBONACCI - Dynamic programming with array
 * ========================================================================== */

int64_t compute_dp_fib(int64_t n) {
    int64_t dp[100];
    dp[0] = 0;
    dp[1] = 1;
    
    for (int64_t i = 2; i <= n; i++) {
        dp[i] = dp[i-1] + dp[i-2];
    }
    return dp[n];
}

/* ==========================================================================
 * MATH FUNCTIONS - Tests math library call overhead
 * ========================================================================== */

double compute_math_ops(double x) {
    return sqrt(x) + sin(x) + cos(x) + log(x + 1);
}

/* ==========================================================================
 * BENCHMARK RUNNER
 * ========================================================================== */

typedef double (*bench_func_f64)(double);
typedef double (*bench_func_f64_2)(double, double);
typedef double (*bench_func_f64_3)(double, double, double);
typedef int64_t (*bench_func_i64)(int64_t);

void run_benchmark_f64(const char* name, bench_func_f64 fn, double arg, int iterations) {
    /* Warmup */
    for (int i = 0; i < 1000; i++) {
        volatile double r = fn(arg);
        (void)r;
    }
    
    double start = get_time_ns();
    for (int i = 0; i < iterations; i++) {
        volatile double r = fn(arg);
        (void)r;
    }
    double end = get_time_ns();
    
    double ns_per_op = (end - start) / iterations;
    printf("%-30s %12.2f ns/op\n", name, ns_per_op);
}

void run_benchmark_f64_2(const char* name, bench_func_f64_2 fn, double a1, double a2, int iterations) {
    /* Warmup */
    for (int i = 0; i < 1000; i++) {
        volatile double r = fn(a1, a2);
        (void)r;
    }
    
    double start = get_time_ns();
    for (int i = 0; i < iterations; i++) {
        volatile double r = fn(a1, a2);
        (void)r;
    }
    double end = get_time_ns();
    
    double ns_per_op = (end - start) / iterations;
    printf("%-30s %12.2f ns/op\n", name, ns_per_op);
}

void run_benchmark_f64_3(const char* name, bench_func_f64_3 fn, double a1, double a2, double a3, int iterations) {
    /* Warmup */
    for (int i = 0; i < 1000; i++) {
        volatile double r = fn(a1, a2, a3);
        (void)r;
    }
    
    double start = get_time_ns();
    for (int i = 0; i < iterations; i++) {
        volatile double r = fn(a1, a2, a3);
        (void)r;
    }
    double end = get_time_ns();
    
    double ns_per_op = (end - start) / iterations;
    printf("%-30s %12.2f ns/op\n", name, ns_per_op);
}

void run_benchmark_i64(const char* name, bench_func_i64 fn, int64_t arg, int iterations) {
    /* Warmup */
    for (int i = 0; i < 1000; i++) {
        volatile int64_t r = fn(arg);
        (void)r;
    }
    
    double start = get_time_ns();
    for (int i = 0; i < iterations; i++) {
        volatile int64_t r = fn(arg);
        (void)r;
    }
    double end = get_time_ns();
    
    double ns_per_op = (end - start) / iterations;
    printf("%-30s %12.2f ns/op\n", name, ns_per_op);
}

int main(int argc, char** argv) {
    printf("==============================================================================\n");
    printf("C REFERENCE BENCHMARKS (gcc -O2)\n");
    printf("==============================================================================\n\n");
    
    run_benchmark_f64_2("Mandelbrot", compute_mandelbrot, 0.25, 0.5, 300000);
    run_benchmark_f64_3("Integrate", compute_integrate, 0.0, 1.0, 1000.0, 1000000);
    run_benchmark_f64("Leibniz", compute_leibniz, 100000.0, 10000);
    run_benchmark_f64("Newton", compute_newton, 2.0, 100000000);
    run_benchmark_i64("ArraySum", compute_array_sum, 50, 30000000);
    run_benchmark_i64("DPFib", compute_dp_fib, 40, 20000000);
    run_benchmark_f64("MathOps", compute_math_ops, 2.0, 40000000);
    
    printf("\n");
    return 0;
}
