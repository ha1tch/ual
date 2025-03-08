# ual

## Overview

ual is a high-level, stack-based programming language designed for resource-constrained environments like embedded systems and retro computing platforms. What distinguishes ual is its unified approach to program safety through a consistent stack-based paradigm, bridging the gap between low-level hardware control and high-level programming abstractions without sacrificing safety or performance.

### Quick Links

- [Current Specification (v1.3)](https://github.com/ha1tch/ual/blob/main/spec/ual-1.3-spec-P1.md)
- [Unified Stack-Based Safety Approach](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-DESIGN-unified-stack-approach.md)
- [Explore the Language Evolution](#language-evolution)
- [Getting Started with ual](#getting-started)

## Getting Started

Choose a primer based on your background:

- [For Mainstream Programmers](https://github.com/ha1tch/ual/blob/main/doc/primer/ual-primer-01-mainstream.md) - Coming from Python, JavaScript, Java, etc.
- [For Embedded Systems Programmers](https://github.com/ha1tch/ual/blob/main/doc/primer/ual-primer-02-embedded.md) - Focused on hardware control and resource constraints
- [For Retro & Minimalist Programmers](https://github.com/ha1tch/ual/blob/main/doc/primer/ual-primer-03-retro-minimalist.md) - Interested in stack-based computing and minimalism

## Key Features

- **Stack-based operations** for arithmetic and memory access
- **Container-centric type system** where types are properties of stacks rather than values
- **Stack-based memory safety** through ownership and borrowing (proposed in 1.5)
- **Error propagation** via dedicated error stack with compile-time tracking (proposed in 1.4)
- **Zero runtime overhead** for safety features through compile-time checking
- **Lua-style** control flow, scoping, and data structures
- **Go-like package** conventions (uppercase = exported, lowercase = private)
- **Binary/hexadecimal literals** and **bitwise operators** for hardware-oriented programming
- **Multiple return values** and flexible iteration constructs

## Distinctive Approach to Safety

ual takes a unique approach to program safety by unifying three traditionally separate concerns:

1. **Type Safety**: Types are attributes of stacks (containers), not values. The `bring_<type>` operation atomically transfers values between stacks with appropriate type conversion.

2. **Memory Safety**: Ownership is tied to stacks, with explicit transfer operations. This provides Rust-like memory safety guarantees with more visible ownership flow (proposed in 1.5).

3. **Error Control**: Errors propagate through a dedicated error stack that is tracked by the compiler, ensuring errors cannot be silently ignored (proposed in 1.4).

This unified stack-based approach provides strong safety guarantees with zero runtime overhead—critical for embedded systems where both reliability and efficiency are essential.

## Language Inspirations

ual synthesizes concepts from several established languages while creating its own path:

- **From Lua**: Clean syntax, function definitions, tables, and multiple returns
- **From Forth**: Stack-based computation model and resource efficiency
- **From Go**: Package system and pragmatic design philosophy
- **From Rust**: Compile-time safety guarantees without runtime cost
- **From Factor**: Advanced stack-based programming techniques

## Basic Example

```lua
package main

import "fmt"

function calculate(a, b)
  -- Push values onto the Integer stack
  @Stack.new(Integer): alias:"i"
  @i: push(a) push(b) mul
  
  return i.pop()
end

function main()
  result = calculate(10, 20)
  fmt.Printf("Result: %d\n", result)
  
  -- Bitwise operations for hardware access
  local port_value = 0x55 & 0x0F  -- Mask off high bits
  local shifted = port_value << 2  -- Shift left by 2 bits
  
  return 0
end
```

## Stacked Mode and Type Conversion Example

ual's stacked mode provides concise notation while the type system ensures safety:

```lua
package main

import "fmt"

function process_data(raw_input)
  -- Create typed stacks with aliases for clarity
  @Stack.new(String): alias:"s"
  @Stack.new(Float): alias:"f"
  @Stack.new(Integer): alias:"i"
  
  -- Parse string input and convert between types
  @s: push(raw_input)
  @s: split:","              -- Split CSV format data
  
  -- Convert string to float (shorthand for bring_string)
  @f: <s
  
  -- Perform calculation with direct mathematical expression
  @f: dup (9/5)*32 sum       -- Convert Celsius to Fahrenheit
  
  -- Convert float to integer (truncating decimal)
  @i: <f
  
  -- Format results
  @s: push("Result: ") push(i.pop()) concat
  
  return s.pop()
end

function main()
  fmt.Printf("%s\n", process_data("25.5"))
  return 0
end
```

## Memory Safety Example (Proposed in 1.5)

```lua
function handle_resource(filename)
  -- Create an owned resource stack
  @Stack.new(Resource, Owned): alias:"ro"
  @ro: push(open_file(filename))
  
  -- Borrow immutably for reading
  @Stack.new(Resource, Borrowed): alias:"rb"
  @rb: <<ro                       -- Borrow without consuming
  read_config(rb.pop())
  
  -- Borrow mutably for writing
  @Stack.new(Resource, Mutable): alias:"rm"
  @rm: <:mut ro                   -- Mutable borrow
  write_config(rm.pop())
  
  -- Resource automatically closed when owned stack goes out of scope
  return true
end
```

## Error Handling Example (Proposed in 1.4)

```lua
@error > function read_file(filename)
  if file_not_accessible then
    @error > push("Cannot access file: " .. filename)
    return nil
  end
  return file_contents
end

function process()
  content = read_file("config.txt")
  if @error > depth() > 0 then
    err = @error > pop()
    fmt.Printf("Error: %s\n", err)
    return false
  end
  
  -- Process content
  return true
end
```

## Use Cases

ual is particularly well-suited for:

- **Embedded systems programming** where safety and efficiency are both critical
- **Resource-constrained IoT devices** that can't afford runtime overhead
- **Retro computing** and vintage hardware
- **Cross-platform development** spanning modern and classic architectures
- **Educational environments** for teaching both stack-based programming and safety concepts

## The UALSYSTEM Architecture

ual serves as the foundation for the UALSYSTEM cross-compilation architecture, which enables code written in different paradigms to target multiple hardware platforms:

- Classic Z80 hardware
- Uxn virtual machines
- Modern RISC-V platforms (including ESP32)
- AVR microcontrollers
- And more

This approach allows developers to work in their familiar programming model while deploying to a wide range of platforms.

## Standard Library

ual includes a practical standard library:

- **con** - Console operations
- **fmt** - String formatting
- **sys** - System operations
- **io** - Low-level input/output
- **str** - String manipulation
- **math** - Basic numeric functions

## Design Philosophy

ual embraces the following design principles:

- **Unified Safety Model**: Type safety, memory safety, and error control through a consistent paradigm
- **Zero Runtime Overhead**: Safety guarantees without performance penalties
- **Explicitness**: Making operations like type conversions and ownership transfers visible
- **Progressive Complexity**: Simple operations remain simple, complexity available when needed
- **Dual Paradigm**: Combining stack-based and variable-based programming styles
- **Resource Efficiency**: Optimized for constrained environments
- **Hardware Accessibility**: Direct access to hardware when needed

## Language Evolution

ual is an evolving language with:

- **ual 1.3**: Current stable version with stack operations, typed stacks, and switch statements
  - [Language Basics (Part 1)](https://github.com/ha1tch/ual/blob/main/spec/ual-1.3-spec-P1.md)
  - [Stacks as First-Class Objects (Part 2)](https://github.com/ha1tch/ual/blob/main/spec/ual-1.3-spec-P2.md)

- **ual 1.4**: Proposed extensions
  - [Typed Stacks](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-typed-stacks-01.md) and [bring_&lt;type&gt; Operations](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-typed-stacks-02.md)
  - [Error Stack Mechanism](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-error-stack.md)
  - [Macro System](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-macros.md)
  - [Conditional Compilation](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-conditional-compilation-02.md)
  - [Type System Design](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-DESIGN-type-system.md)

- **ual 1.5**: Future proposals
  - [Stack-Based Ownership System](https://github.com/ha1tch/ual/blob/main/spec/ual-1.5-PROPOSAL-ownership-system.md)

## License
    ual is available under the Apache 2.0 license.
    https://github.com/ha1tch/ual/blob/main/LICENSE


## Author
    haitch  
    email:  h (at) ual.fi  
    Social: https://oldbytes.space/@haitchfive
