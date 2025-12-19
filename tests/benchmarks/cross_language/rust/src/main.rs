//! ual Cross-Language Benchmarks - Rust Reference
//!
//! Run: cargo run --release [leibniz|mandelbrot|newton|all]

use std::env;

/// Leibniz series for π (1M terms) - matches ual benchmark
fn compute_leibniz() -> f64 {
    let mut sum = 0.0;
    let mut sign = 1.0;
    let mut denom = 1.0;
    let terms = 1_000_000;
    
    for _ in 0..terms {
        sum += sign / denom;
        sign = -sign;
        denom += 2.0;
    }
    
    4.0 * sum
}

/// Mandelbrot 50x50 grid - matches ual benchmark
fn compute_mandelbrot() -> f64 {
    let width = 50;
    let height = 50;
    let max_iter = 100;
    let escape = 4.0;
    let mut total = 0.0;
    
    let x_min = -2.0_f64;
    let x_max = 1.0_f64;
    let y_min = -1.5_f64;
    let y_max = 1.5_f64;
    let x_step = (x_max - x_min) / (width as f64);
    let y_step = (y_max - y_min) / (height as f64);
    
    for py in 0..height {
        let ci = y_min + (py as f64) * y_step;
        for px in 0..width {
            let cr = x_min + (px as f64) * x_step;
            let mut zr = 0.0;
            let mut zi = 0.0;
            let mut iter = 0;
            
            while iter < max_iter {
                let zr2 = zr * zr;
                let zi2 = zi * zi;
                if zr2 + zi2 > escape { break; }
                zi = 2.0 * zr * zi + ci;
                zr = zr2 - zi2 + cr;
                iter += 1;
            }
            total += iter as f64;
        }
    }
    
    total
}

/// Newton-Raphson sqrt for 1000 values - matches ual benchmark
fn compute_newton() -> f64 {
    let mut sum = 0.0;
    let limit = 1000;
    
    for n in 1..=limit {
        let nf = n as f64;
        let mut guess = nf / 2.0;
        for _ in 0..20 {
            guess = (guess + nf / guess) / 2.0;
        }
        sum += guess;
    }
    
    sum
}

fn main() {
    let args: Vec<String> = env::args().collect();
    let which: &str = if args.len() > 1 { &args[1] } else { "all" };
    
    match which {
        "leibniz" => println!("{:.10}", compute_leibniz()),
        "mandelbrot" => println!("{:.0}", compute_mandelbrot()),
        "newton" => println!("{:.10}", compute_newton()),
        _ => {
            println!("Leibniz: {:.10}", compute_leibniz());
            println!("Mandelbrot: {:.0}", compute_mandelbrot());
            println!("Newton: {:.10}", compute_newton());
        }
    }
}
