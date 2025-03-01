# ual

## Overview

ual is a high-level, stack-based language designed for use on minimal or retro platforms with small virtual machines or embedded hardware. It bridges the gap between low-level hardware control and high-level programming abstractions, making it ideal for resource-constrained environments.

## Key Features

- **Stack-based operations** for arithmetic and memory access
- **Lua-style** control flow, scoping, and data structures
- **Go-like package** conventions (uppercase = exported, lowercase = private)
- **Binary/hexadecimal literals** for hardware-oriented programming
- **Bitwise operators** for direct register and mask manipulation
- **Multiple return values** and flexible iteration constructs
- **Cross-platform compilation** targeting various architectures

## Language Inspirations

ual draws inspiration from several established languages:

- **From Lua**: The clean syntax, function definitions, local variables, tables as the primary data structure, and multiple return values.

- **From Forth**: The stack-based computational model, direct hardware access, and emphasis on small implementation footprint.

- **From Go**: The package system with uppercase/lowercase visibility rules, clear import declarations, and pragmatic approach to language design.

## Example

```lua
package main

import "con"
import "fmt"

function main()
  -- Basic calculation using stack operations
  push(10)
  push(20)
  add()  -- Adds top two stack values
  
  -- Display the result
  fmt.Printf("Result: %d\n", pop())
  
  -- Bitwise operations for hardware access
  local port_value = 0x55 & 0x0F  -- Mask off high bits
  local shifted = port_value << 2  -- Shift left by 2 bits
  
  return 0
end
```

## Stacked Mode Example

ual's stacked mode provides concise, Forth-like expressiveness while maintaining readability:

```lua
package main

import "fmt"

-- Implement the Fibonacci sequence using stack operations
function fibonacci(n)
  -- Using stacked mode with the '>' prefix
  > push:1 push:1                -- Initialize with first two Fibonacci numbers
  
  > push(n) push:2 sub           -- Calculate how many more numbers to generate
  while_true(dstack.peek() > 0)
    > over over add              -- Add the top two numbers
    > rot drop                   -- Remove the oldest number
    > push(dstack.peek(1)) push:1 sub  -- Decrement counter
  end_while_true
  
  > drop                         -- Remove the counter
  return dstack.pop()            -- Return the nth Fibonacci number
end

function main()
  -- Calculate some Fibonacci numbers
  for i = 1, 10 do
    fmt.Printf("Fibonacci %d: %d\n", i, fibonacci(i))
  end
  
  -- Demonstrate multi-stack operations
  @dstack > push:10 push:20 mul  -- Use data stack
  @rstack > push:5 push:5 add    -- Use return stack
  
  -- Combine results from both stacks
  > push(rstack.pop()) mul
  
  fmt.Printf("Result of stack operations: %d\n", dstack.pop())
  
  return 0
end
```

This example demonstrates how ual's stacked mode combines the expressiveness of Forth-style stack manipulation with the readability and structure of a modern language. The `>` prefix denotes stacked mode lines, and the `@stack >` syntax allows operations on specific stacks.

## Use Cases

ual is particularly well-suited for:

- **Embedded systems programming** on microcontrollers
- **Retro computing** on vintage hardware
- **Resource-constrained IoT devices**
- **Educational environments** for teaching programming concepts
- **Cross-platform development** spanning modern and classic architectures

## The UALSYSTEM Architecture

ual serves as the foundation for the UALSYSTEM cross-compilation architecture, which enables code written in different paradigms (register-based, stack-based) to target multiple hardware platforms:

- Classic Z80 hardware
- Uxn virtual machines
- Modern RISC-V platforms (including ESP32)
- And more

This approach allows developers to work in their familiar programming model while deploying to a wide range of platforms.

## Standard Library

ual includes a minimal yet practical standard library:

- **con** - Console operations
- **fmt** - String formatting
- **sys** - System operations
- **io** - Low-level input/output
- **str** - String manipulation
- **math** - Basic numeric functions

## Compiler Implementation

The ual compiler is designed for flexibility and performance:

- **Multi-tiered compilation** to various target platforms
- **Optimization passes** for efficient code generation
- **Small runtime footprint** suitable for constrained environments

ual embraces the following design principles:

- **Minimalism**: A concise core with carefully selected features
- **Practicality**: Real-world utility for embedded and retro computing
- **Accessibility**: Familiar syntax with modern programming constructs
- **Efficiency**: Optimized for resource-constrained environments
- **Flexibility**: Multiple programming paradigms in one language
- **Reuse**: Don't reinvent the wheel, use existing toolchains

## License
ual is available under the Apache 2.0 license.

## Author

haitch  
haitch@duck.com  
Social: https://oldbytes.space/@haitchfive
