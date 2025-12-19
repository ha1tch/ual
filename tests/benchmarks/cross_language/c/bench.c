/*
 * ual Cross-Language Benchmarks - C Reference
 * 
 * Compile: gcc -O2 -o bench bench.c -lm
 * Run: ./bench [leibniz|mandelbrot|newton|all]
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* Benchmark 1: Leibniz series for π (1M terms) - matches ual benchmark */
double compute_leibniz(void) {
    double sum = 0.0;
    double sign = 1.0;
    double denom = 1.0;
    int terms = 1000000;
    
    for (int i = 0; i < terms; i++) {
        sum += sign / denom;
        sign = -sign;
        denom += 2.0;
    }
    
    return 4.0 * sum;
}

/* Benchmark 2: Mandelbrot 50x50 grid - matches ual benchmark */
double compute_mandelbrot(void) {
    int width = 50, height = 50;
    int max_iter = 100;
    double escape = 4.0;
    double total = 0.0;
    
    double x_min = -2.0, x_max = 1.0;
    double y_min = -1.5, y_max = 1.5;
    double x_step = (x_max - x_min) / width;
    double y_step = (y_max - y_min) / height;
    
    for (int py = 0; py < height; py++) {
        double ci = y_min + py * y_step;
        for (int px = 0; px < width; px++) {
            double cr = x_min + px * x_step;
            double zr = 0.0, zi = 0.0;
            int iter = 0;
            
            while (iter < max_iter) {
                double zr2 = zr * zr;
                double zi2 = zi * zi;
                if (zr2 + zi2 > escape) break;
                zi = 2.0 * zr * zi + ci;
                zr = zr2 - zi2 + cr;
                iter++;
            }
            total += iter;
        }
    }
    
    return total;
}

/* Benchmark 3: Newton-Raphson sqrt for 1000 values - matches ual benchmark */
double compute_newton(void) {
    double sum = 0.0;
    int limit = 1000;
    
    for (int n = 1; n <= limit; n++) {
        double guess = (double)n / 2.0;
        for (int i = 0; i < 20; i++) {
            guess = (guess + (double)n / guess) / 2.0;
        }
        sum += guess;
    }
    
    return sum;
}

int main(int argc, char** argv) {
    const char* which = (argc > 1) ? argv[1] : "all";
    volatile double result;
    
    if (strcmp(which, "leibniz") == 0) {
        result = compute_leibniz();
        printf("%.10f\n", result);
    } else if (strcmp(which, "mandelbrot") == 0) {
        result = compute_mandelbrot();
        printf("%.0f\n", result);
    } else if (strcmp(which, "newton") == 0) {
        result = compute_newton();
        printf("%.10f\n", result);
    } else {
        /* Run all */
        printf("Leibniz: %.10f\n", compute_leibniz());
        printf("Mandelbrot: %.0f\n", compute_mandelbrot());
        printf("Newton: %.10f\n", compute_newton());
    }
    
    return 0;
}
