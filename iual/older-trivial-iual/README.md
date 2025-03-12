# iual v0.0.1
*iual is an exceedingly trivial interactive ual 0.0.1 interpreter*

## Overview

iual is a toy interpreter that combines Forth-like stack manipulation with concurrent spawn management. It supports:

- **Dynamic Stacks:**  
  Int, string, and (potentially) float stacks. Two default int stacks, `dstack` and `rstack`, are created at startup.

- **Forth-like Operations:**  
  Standard operations (push, pop, dup, swap, drop) and advanced ones (tuck, pick, roll, over2, drop2, swap2, depth).  
  Memory and bitwise operations (store, load, and, or, xor, shl, shr) are also provided for int stacks.

- **Compound Commands:**  
  You can select a stack using a selector (e.g. `@dstack:`) and then run a series of commands in one line.  
  Multi-parameter operations can be written in a function-like syntax (e.g. `div(10,2)`).

- **Spawn Scripting:**  
  Spawns (accessed via `@spawn`) are managed as separate goroutines (using pthreads).  
  You can load a script from a string stack using the `bring` operation and execute it with `run`.

## Why Explore These Concepts?

- **Stack Programming:**  
  Explore low-level control using stacks as the primary data structure.

- **Interactive DSL Design:**  
  Combine Forth-like commands with an extensible compound syntax.

- **Concurrency:**  
  Manage independent spawn goroutines, each capable of executing its own script.

- **Extensibility:**  
  The design lets you easily add new operations and experiment with different stack paradigms (LIFO vs. FIFO).

This project is a playground for testing ideas in interactive, stack-based programming.

---

Version: 0.0.1
