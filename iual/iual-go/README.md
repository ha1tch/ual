
# iual v0.0.1
*iual is an exceedingly trivial interactive ual 0.0.1 interpreter*

## Overview

iual is a toy interpreter that combines several advanced concepts in stack programming with interactive command execution. It is designed as a playground for testing ideas in dynamic stack manipulation, compound commands, and concurrent (spawn) goroutine management. The interpreter supports multiple types of stacks (integer, string, and float) and a special spawn stack for managing goroutines.

## Key Concepts

### 1. **Dynamic Stacks**
- **Multiple Types:**  
  iual allows you to create and work with different types of stacks:  
  - **Integer stacks (int)**
  - **String stacks (str)**
  - **Float stacks (float)**
- **Default Stacks:**  
  By default, two stacks are available:  
  - **dstack (data stack)**
  - **rstack (return stack)**
- **Stack Operations:**  
  Each stack supports common Forth-like operations (e.g., push, pop, dup, swap, drop) along with arithmetic operations for numeric stacks and concatenation or string manipulation for string stacks.

### 2. **Compound Command Syntax & Function-Like Operations**
- **Compound Commands:**  
  iual allows you to execute multiple operations on a selected stack in a single line using a compound command syntax.  
  For example:
  ```
  @dstack: push:1 pop mul
  ```
  This command selects the stack `dstack` and then sequentially executes:
  - `push:1` – Push the value 1.
  - `pop` – Pop the top value.
  - `mul` – Multiply the top two elements.
- **Function-Like Syntax:**  
  Operations that take multiple parameters can be expressed in a function-like syntax:
  ```
  @dstack: div(10,2)
  ```
  This is equivalent to:
  ```
  @dstack: push:10 push:2 div
  ```
  This extensible syntax makes it easy to add new multi-parameter operations.

### 3. **Spawn (Goroutine) Management**
- **Spawn Stack:**  
  The spawn stack is a special stack available as `@spawn`. It is used for managing goroutines (or spawns) that run concurrently.
- **Spawn Operations:**  
  When you select `@spawn`, you can perform operations such as:
  - `list` – List all running spawns.
  - `add <name>` – Create and start a new spawn goroutine.
  - `pause <name>` / `resume <name>` / `stop <name>` – Control the execution of a spawn.
- **Script Execution in Spawns:**  
  iual supports sending scripts to spawns. You can store commands in a string stack and then use the `bring` operation (e.g. `bring(str,@sstack)`) to load a script into a spawn’s internal container. A subsequent `run` command then executes the script in that spawn's goroutine. This feature allows each spawn to execute its own set of instructions independently.

### 4. **The Bring Operation**
- **Purpose:**  
  The `bring` operation moves data from one stack to another after performing type conversion if necessary.
- **Usage Examples:**  
  - For numeric stacks:  
    ```
    @dstack: bring(int,@rstack)
    ```
    Pops a value from the `rstack` (an integer stack) and pushes it onto `dstack`.
  - For spawn scripts:  
    ```
    @spawn: bring(str,@sstack) run
    ```
    Retrieves a multi-line script from the string stack `sstack`, stores it in the spawn, and then executes it.

## Why Test These Ideas?

- **Exploration of Stack-Based Languages:**  
  iual is inspired by traditional Forth-like languages. By working with explicit stacks, you get insight into low-level data management and control flow, which can be valuable for understanding computer architecture and interpreter design.

- **Concurrency and Scripting:**  
  Integrating goroutine management (spawns) with an interactive stack-based language provides a unique way to experiment with concurrent execution and dynamic scripting. It’s an interesting exploration of how scripts can be transferred, stored, and executed within independent threads of execution.

- **Extensibility:**  
  The function-like syntax for multi-parameter operations makes it easy to extend the language. Researchers and hobbyists can add new operations, adjust conversion rules, or experiment with alternative stack behaviors (e.g., FIFO vs. LIFO, or even more exotic behaviors).

- **Interactive Prototyping:**  
  iual is an interactive interpreter, making it an excellent tool for rapid prototyping of new ideas in language design. Its simple command syntax, compound commands, and the ability to send scripts to concurrently running spawns encourage creative experimentation.

- **Educational Value:**  
  This toy interpreter can be used as an educational tool to demonstrate the principles of stack-based programming, concurrent programming in Go, and the design of simple domain-specific languages (DSLs).

