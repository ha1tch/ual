# rual — Rust Runtime for ual

Rust runtime library for ual programs compiled with `--target rust`.

## Status

**Production Ready** — The Rust backend has achieved 100% output equivalence with Go:

- 79/79 examples generate valid Rust code
- 79/79 examples compile with rustc 1.75+
- 79/79 examples produce output identical to Go backend (100% parity)

Binary size comparison:

| Build Profile | Go | Rust |
|---------------|-----|------|
| Default | 1.9M | 13M |
| Stripped | 1.3M | 403K |
| Small (`--small`) | 1.3M | 343K |

Rust with `--small` produces binaries ~4x smaller than Go.

See `test_rust_backend.sh` in the main ual directory for the test suite.

## Overview

This crate provides the runtime primitives that ual programs need when compiled to Rust:

- **`Stack<T>`** — Type-safe stacks with perspectives (LIFO, FIFO, Indexed, Hash)
- **`Value`** — Dynamic typing for heterogeneous stacks
- **`View`** — Borrowed perspectives on stacks
- **`BlockingStack<T>`** — Stacks with blocking `take()` and timeout support
- **`WSDeque` / `WSStack`** — Work-stealing primitives

## Design Philosophy

ual treats coordination as primary and computation as subordinate. Stacks are boundaries where processes meet. Perspectives determine how access parameters are interpreted, not how data is stored.

## Usage

```rust
use rual::{Stack, Perspective, BlockingStack};

// Typed stack with LIFO perspective
let stack: Stack<i64> = Stack::new(Perspective::LIFO);
stack.push(42).unwrap();
stack.push(17).unwrap();

assert_eq!(stack.pop().unwrap(), 17);  // LIFO: last in, first out

// Same data, different perspective via View
use std::sync::Arc;
use rual::View;

let shared = Arc::new(Stack::<i64>::new(Perspective::LIFO));
shared.push(1).unwrap();
shared.push(2).unwrap();
shared.push(3).unwrap();

let lifo_view = View::lifo(Arc::clone(&shared));
let fifo_view = View::fifo(Arc::clone(&shared));

assert_eq!(lifo_view.peek().unwrap(), 3);  // Top
assert_eq!(fifo_view.peek().unwrap(), 1);  // Bottom

// Blocking take with timeout
let blocking = BlockingStack::<i64>::new(Perspective::FIFO);
blocking.push(100).unwrap();

let value = blocking.take_timeout(Some(1000)).unwrap();  // Wait up to 1 second
assert_eq!(value, 100);
```

## Perspectives

| Perspective | Push | Pop | Use Case |
|-------------|------|-----|----------|
| LIFO | Append | Take newest | Traditional stack, recursion |
| FIFO | Append | Take oldest | Queue, message passing |
| Indexed | Append | By index | Array-like access |
| Hash | By key | By key | Key-value store |

## Compute Blocks

For compute blocks (tight numerical loops), use `stack.lock()` to get a `StackGuard` with raw access:

```rust
let stack: Stack<i64> = Stack::new(Perspective::Indexed);
// ... push data ...

{
    let mut guard = stack.lock();
    let slice = guard.as_mut_slice();
    
    // Direct slice manipulation — no locking overhead per access
    for i in 0..slice.len() {
        slice[i] *= 2;
    }
}
```

## Requirements

- Rust 1.75 or later (1.80+ recommended for running benchmarks)

## License

MIT — see the main ual repository for details.
