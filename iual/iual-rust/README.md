# iual - Interactive Universal Assembly Language (Rust Version)

This is a Rust implementation of the iual interpreter, a stack-based language inspired by Forth with additional features like different stack types, asynchronous task management, and more.

## Project Structure

```
iual/
├── Cargo.toml
├── src/
│   ├── main.rs           # Entry point, CLI handling
│   ├── lib.rs            # Library exports
│   ├── memory.rs         # Global memory for STORE/LOAD
│   ├── conversion.rs     # Type conversion helpers
│   ├── selector.rs       # Stack selector implementation
│   ├── cli.rs            # CLI command processor
│   ├── stacks/
│   │   ├── mod.rs        # Stack module exports
│   │   ├── int_stack.rs  # Integer stack implementation
│   │   ├── str_stack.rs  # String stack implementation
│   │   └── float_stack.rs # Float stack implementation
│   └── spawn/
│       ├── mod.rs        # Spawn module exports
│       ├── task.rs       # Managed task implementation
│       └── manager.rs    # Task manager
```

## Features

- **Multiple Stack Types**: Support for integer, string, and float stacks
- **Stack Modes**: LIFO (Last In, First Out) and FIFO (First In, First Out) modes
- **Asynchronous Task Management**: Spawn, pause, resume, and stop tasks
- **Command Scripting**: Execute scripts stored in string stacks
- **Global Memory**: Store and load values from a global memory space
- **Return Stack Operations**: For more complex control flow
- **Bitwise Operations**: AND, OR, XOR, shift left, shift right
- **Rich Command Set**: Forth-like stack manipulation commands

## Commands

### Stack Selection

- `@stackname`: Select a stack to operate on
- `@spawn`: Select the spawn stack for task management

### Stack Creation

- `new <name> <int|str|float>`: Create a new stack with the given name and type

### Task Management

- `spawn <name>`: Create a new task
- `pause <name>`: Pause a task
- `resume <name>`: Resume a paused task
- `stop <name>`: Stop a task
- `list`: List all tasks

### Stack Operations

For int stacks:
- `push <value>`: Push a value onto the stack
- `pop`: Remove and return the top value
- `dup`: Duplicate the top value
- `swap`: Swap the top two values
- `drop`: Remove the top value
- `add`, `sub`, `mul`, `div`: Arithmetic operations
- `and`, `or`, `xor`, `shl`, `shr`: Bitwise operations
- `store`, `load`: Memory operations
- More: `tuck`, `pick`, `roll`, `over2`, `drop2`, `swap2`, `depth`

For string stacks:
- `push <value>`: Push a string onto the stack
- `pop`, `dup`, `swap`, `drop`: Basic stack operations
- `add`: Concatenate strings
- `sub <char>`: Remove trailing occurrences of a character
- `mul <n>`: Repeat a string n times
- `div <delim>`: Split by delimiter and join with spaces

For float stacks:
- Similar to int stacks but with floating-point values

### Additional Features

- `lifo`, `fifo`: Change stack mode
- `flip`: Reverse the stack
- `send <type> <stack> <task>`: Send a value from a stack to a task
- `pushr`, `popr`, `peekr`: Return stack operations

### Compound Commands

Execute multiple operations in sequence:
```
@dstack: push 10 push 5 add print
```

## Building and Running

```bash
# Clone the repository
git clone https://github.com/yourusername/iual-rs.git
cd iual-rs

# Build
cargo build --release

# Run
cargo run --release
```

## Example Usage

```
> new myints int
Created new int stack 'myints'

> @myints
Stack selector set to 'myints' of type int

> push 10
Pushed 10 to stack

> push 20
Pushed 20 to stack

> add
> print
IntStack (lifo mode): [30]

> @spawn
Stack selector set to 'spawn' of type spawn

> add mytask
Added task 'mytask'

> send int myints mytask
Sent message to 'mytask'

> quit
Exiting...
```

## License

MIT