//! Benchmarks for rual stack operations

use bencher::{benchmark_group, benchmark_main, Bencher};
use rual::{Stack, Perspective};

fn bench_lifo_push_pop(b: &mut Bencher) {
    let stack: Stack<i64> = Stack::new(Perspective::LIFO);
    b.iter(|| {
        for i in 0..1000 {
            stack.push(i).unwrap();
        }
        for _ in 0..1000 {
            stack.pop().unwrap();
        }
    });
}

fn bench_fifo_push_pop(b: &mut Bencher) {
    let stack: Stack<i64> = Stack::new(Perspective::FIFO);
    b.iter(|| {
        for i in 0..1000 {
            stack.push(i).unwrap();
        }
        for _ in 0..1000 {
            stack.pop().unwrap();
        }
    });
}

benchmark_group!(benches, bench_lifo_push_pop, bench_fifo_push_pop);
benchmark_main!(benches);
