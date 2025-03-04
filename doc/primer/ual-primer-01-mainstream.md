# ual Primer for Mainstream Programmers

## Introduction

If you're coming from languages like Python, JavaScript, or Java, **ual** offers a familiar yet refreshing programming experience. It combines the straightforward syntax you're used to with a unique stack-based approach that enables powerful patterns, particularly for resource-constrained environments.

## The Familiar Territory

Let's start with what will feel immediately comfortable:

### Package System and Imports

```lua
package myapp

import "fmt"
import "io"
```

Just like in Go or Java, ual organizes code into packages and has a clear import system.

### Variables and Functions

```lua
function calculate_area(width, height)
  local area = width * height
  return area
end

function main()
  local w = 10
  local h = 5
  local result = calculate_area(w, h)
  fmt.Printf("Area: %d\n", result)
  return 0
end
```

This looks quite similar to Lua or JavaScript - local variables, function definitions, and return values work as you'd expect.

### Control Flow

ual provides familiar control structures:

```lua
-- If statement
if_true(x > 10)
  fmt.Printf("x is greater than 10\n")
end_if_true

-- While loop
local i = 0
while_true(i < 10)
  fmt.Printf("i = %d\n", i)
  i = i + 1
end_while_true

-- For loop
for i = 1, 10 do
  fmt.Printf("i = %d\n", i)
end
```

### Data Structures

ual provides tables (similar to objects or dictionaries) and arrays:

```lua
-- Table (like objects in JavaScript or dictionaries in Python)
local person = {
  name = "Alice",
  age = 30,
  ["favorite color"] = "blue"
}

-- Array (zero-indexed like most languages)
local numbers = [1, 2, 3, 4, 5]
```

## The Unique Stack Approach

Now let's explore what makes ual unique - its stack-based operations.

### What is a Stack?

Think of a stack like a pile of plates - you can add (push) to the top or remove (pop) from the top. In programming, stacks are useful for managing temporary data.

### Basic Stack Operations

```lua
-- Push values onto the stack
push(10)
push(20)

-- Stack now contains: [10, 20] (20 is on top)

-- Add the top two values
add()  -- pops 10 and 20, pushes their sum 30

-- Stack now contains: [30]

-- Get the result
local result = pop()  -- result = 30, stack is now empty
```

### The Dual Paradigm Advantage

What makes ual powerful is that you can mix traditional variable-based programming with stack operations:

```lua
function calculate_hypotenuse(a, b)
  -- Push values to stack
  push(a)
  push(a)
  mul()     -- a²
  
  push(b)
  push(b)
  mul()     -- b²
  
  add()     -- a² + b²
  sqrt()    -- √(a² + b²)
  
  -- Pop result into a variable
  local result = pop()
  return result
end
```

### The Stacked Mode Syntax

For more complex stack operations, ual offers a concise "stacked mode" syntax:

```lua
function factorial(n)
  > push(n) push:1 eq if_true
    > drop push:1
    return pop()
  > end_if_true
  
  > push(n) dup push:1 sub
  > factorial mul
  
  return pop()
end
```

The `>` prefix indicates stacked mode, where operations act on the default data stack. This can be more readable for stack-heavy algorithms.

ual 1.4 introduces an alternative, cleaner syntax using colons:

```lua
function factorial(n)
  @dstack: push(n) push:1 eq if_true
    @dstack: drop push:1
    return pop()
  @dstack: end_if_true
  
  @dstack: push(n) dup push:1 sub
  @dstack: factorial mul
  
  return pop()
end
```

Both syntaxes are supported, but the colon version is recommended for new code as it's cleaner and more consistent with other language elements.

## Type Safety Without the Hassle

One of ual's innovations is its container-centric type system:

```lua
@Stack.new(Integer): alias:"i"  -- Stack that accepts integers
@Stack.new(String): alias:"s"   -- Stack that accepts strings

@i: push(42)       -- Valid: integer into integer stack
@s: push("hello")  -- Valid: string into string stack
@i: push("hello")  -- Error: string cannot go into integer stack
```

This provides type safety while avoiding the complexity of traditional static typing.

## Why You Might Love ual

1. **Clean, familiar syntax** for everyday programming
2. **Stack operations** for elegant solutions to certain problems
3. **Low-level control** when you need it, high-level abstractions when you don't
4. **Type safety** without verbose type annotations
5. **Resource efficiency** for programs that need to run in constrained environments

## A Practical Example

Let's build a simple temperature converter using both paradigms:

```lua
package temperature

import "fmt"

function celsius_to_fahrenheit_variables(celsius)
  -- Traditional imperative approach
  local fahrenheit = (celsius * 9/5) + 32
  return fahrenheit
end

function celsius_to_fahrenheit_stack(celsius)
  -- Stack-based approach
  push(celsius)     -- Push input value
  push(9)           -- Push 9
  push(5)           -- Push 5
  div()             -- 9/5
  mul()             -- celsius * (9/5)
  push(32)          -- Push 32
  add()             -- (celsius * 9/5) + 32
  return pop()      -- Return result
end

-- Using stacked mode (with colon syntax) for a more concise version
function celsius_to_fahrenheit_stacked(celsius_str)
  @Stack.new(String): alias:"s"
  @Stack.new(Float): alias:"f"
  
  @s: push(celsius_str)
  @f: <s dup (9/5)*32 sum  -- Convert string to float, calculate F
  
  return f.pop()
end

function main()
  local celsius = 25
  
  fmt.Printf("%d°C = %d°F (variables)\n", 
             celsius, 
             celsius_to_fahrenheit_variables(celsius))
  
  fmt.Printf("%d°C = %d°F (stack)\n", 
             celsius, 
             celsius_to_fahrenheit_stack(celsius))
             
  fmt.Printf("%s°C = %f°F (stacked mode)\n", 
             "25.5", 
             celsius_to_fahrenheit_stacked("25.5"))
  
  return 0
end
```

This example shows how you can solve the same problem using different approaches in ual, choosing the style that best fits each situation.

## Next Steps

To learn more about ual:
- Try implementing simple algorithms with both variable and stack approaches
- Explore the [typed stack system](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-typed-stacks-01.md) for type safety
- Experiment with the [result pattern](https://github.com/ha1tch/ual/blob/main/spec/ual-1.3-spec-P1.md#482-semantics) for elegant error handling

ual gives you the freedom to write code in a familiar style while exploring the power of stack-based programming at your own pace.