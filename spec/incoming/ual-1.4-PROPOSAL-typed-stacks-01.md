# ual 1.4 PROPOSAL: Typed Stacks
This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---
# Proposed Typed Stack System for ual

## 1. Introduction

This document outlines a proposed type system for stacks in the ual programming language, designed to enhance type safety while maintaining ual's focus on simplicity and efficiency for embedded systems. The typed stack system enables compile-time and runtime type checking for stack operations without introducing the complexity of a full static type system.

## 2. Design Philosophy

The ual typed stack system adheres to the following principles:

1. **Type Safety**: Provide compile-time and runtime guarantees about stack content types
2. **Explicitness**: Make type conversions clear and intentional
3. **Zero Overhead**: Implement in a way that allows efficient compilation for embedded targets
4. **Compatibility**: Maintain backward compatibility with existing ual code
5. **Simplicity**: Avoid complex type theory constructs while providing practical benefits

## 3. Stack Type System

### 3.1 Typed Stack Creation

Stacks can be created with an explicit type parameter that constrains what values may be pushed onto them:

```lua
istack = Stack.new(Integer)  -- Stack that only accepts integers
fstack = Stack.new(Float)    -- Stack that only accepts floating-point values
sstack = Stack.new(String)   -- Stack that only accepts strings
astack = Stack.new(Any)      -- Stack that accepts any type (default if no type specified)
```

When no type is specified, the stack defaults to accepting any type, consistent with previous ual versions:

```lua
stack = Stack.new()  -- Equivalent to Stack.new(Any)
```

### 3.2 Built-in Stack Types

The following built-in types are supported:

- `Integer`: Whole number values
- `Float`: Floating-point number values
- `String`: Text string values
- `Boolean`: True/false values
- `Reference`: Memory address/reference values
- `Any`: Any type of value (no type constraints)
- `Table`: Table/dictionary values
- `Array`: Sequential collection values
- `Stack`: Stack objects (enabling meta-stack operations)

### 3.3 Default System Stacks

The system automatically provides two default stacks with specific types:

```lua
-- These are automatically initialized at program startup:
dstack = Stack.new(Integer)  -- The main data stack is typed as Integer
rstack = Stack.new(Integer)  -- The return stack is typed as Integer
```

This typing reflects the historical use of these stacks primarily for numeric operations and return addresses. The predefined `@dstack` and `@rstack` in stacked mode refer to these typed stacks.

## 4. Type Checking and Enforcement

### 4.1 Push Operations

When a value is pushed to a typed stack, the system checks that the value's type is compatible with the stack's defined type:

```lua
istack = Stack.new(Integer)
istack.push(42)       -- OK: Integer matches the stack type
istack.push("hello")  -- Error: String cannot be pushed to Integer stack
```

Type checking occurs:
- At compile time when possible (for literal values)
- At runtime otherwise (for expressions and variables)

### 4.2 Cross-Stack Operations with Type Conversion

#### 4.2.1 The bring_<type> Operation

The `bring_<type>` operation is a fundamental aspect of ual's typed stack system. It performs three distinct actions as a single atomic operation:

1. **Pop**: Removes the top value from the source stack
2. **Convert**: Transforms the value to the target stack's required type
3. **Push**: Places the converted value onto the current stack

```lua
fstack = Stack.new(Float)
sstack = Stack.new(String)

-- This single atomic operation:
fstack.bring_string(sstack.pop())

-- Is conceptually equivalent to, but more efficient and safer than:
value = sstack.pop()         -- 1. Pop from source stack
converted = to_float(value)  -- 2. Convert to required type
fstack.push(converted)       -- 3. Push to destination stack
```

#### 4.2.2 Benefits of Atomicity

The atomic nature of `bring_<type>` provides several important advantages:

1. **Safety**: Eliminates the possibility of errors or interruptions between the three steps
2. **Efficiency**: Enables optimized implementations without intermediate variables
3. **Clarity**: Communicates the intent clearly - "bring this value here, converting as needed"
4. **Conciseness**: Reduces code verbosity, especially important in stacked mode

#### 4.2.3 Stack-Specific Bring Operations

Each stack type provides the appropriate `bring_<type>` operations based on what conversions make sense for that type:

```lua
istack.bring_float(value)    -- Convert float to integer (truncation/rounding)
istack.bring_string(value)   -- Parse string as integer
istack.bring_boolean(value)  -- Convert boolean to 0/1

fstack.bring_integer(value)  -- Convert integer to float
fstack.bring_string(value)   -- Parse string as float
fstack.bring_boolean(value)  -- Convert boolean to 0.0/1.0

sstack.bring_integer(value)  -- Convert integer to string representation
sstack.bring_float(value)    -- Convert float to string representation
sstack.bring_boolean(value)  -- Convert boolean to "true"/"false"
```

#### 4.2.4 Error Handling

When a conversion error occurs during the `bring_<type>` operation, the entire operation fails cleanly:

```lua
sstack.push("hello")  -- A non-numeric string
fstack.bring_string(sstack.pop())  -- Conversion fails, entire operation fails
```

In this scenario:
- The value is still removed from the source stack
- No value is added to the destination stack
- An appropriate error is raised

This behavior ensures consistency in error cases, as a partially completed operation could leave stacks in an unexpected state.

### 4.3 Type Testing

Stacks provide methods to test the type of the top value without consuming it:

```lua
if sstack.is_numeric() then
  -- Top value can be treated as a number
end

if istack.is_positive() then
  -- Top value is greater than zero
end
```

## 5. Stacked Mode Integration

### 5.1 Stacked Mode Syntax Evolution

#### 5.1.1 Legacy Syntax

In ual 1.3, stacked mode uses the angle bracket syntax:

```lua
@dstack > push:42 dup add
@rstack > push(dstack.pop())
```

#### 5.1.2 New Colon Syntax

ual 1.4 introduces an alternative, cleaner syntax using colons:

```lua
@dstack: push:42 dup add
@rstack: push(dstack.pop())
```

#### 5.1.3 Rationale for Syntax Evolution

The new colon syntax offers several advantages:

1. **Visual Clarity**: Reduces visual noise and improves readability
2. **Consistency**: Uses the same delimiter (colon) that's used in other parts of the language
3. **Extensibility**: Allows for cleaner integration with new features like stack aliases
4. **Ergonomics**: Is easier to type and less error-prone
5. **Mathematical Integration**: Provides better visual separation when incorporating mathematical expressions

Both syntaxes remain fully supported for backward compatibility, but the colon syntax is recommended for new code and is used in all new documentation and examples.

### 5.2 Stack Aliases

Stack aliases provide short names for improved readability in stacked mode:

```lua
@Stack.new(Float): alias:"f"  -- Create and alias in one line
@Stack.new(String): alias:"s"
@Stack.new(Integer): alias:"i"

-- Using the aliases
@f: push:3.14 dup mul
@s: push:"Hello" push:"World" concat
```

Aliases are particularly valuable when working with multiple specialized stacks and can be defined at any point in the code.

```lua
-- Creating multiple stacks with aliases for a specific algorithm
@Stack.new(Float): alias:"in"    -- Input values
@Stack.new(Float): alias:"tmp"   -- Temporary calculations
@Stack.new(Float): alias:"out"   -- Output results
```

### 5.3 Cross-Stack Operations with Shorthand Syntax

A special shorthand syntax enables concise cross-stack operations in stacked mode:

```lua
@sstack: alias:"s"; @fstack: alias:"f"; @istack: alias:"i"

@s: push:"25.5"     -- Push string to string stack
@f: <s              -- Pull from string stack with conversion (shorthand for bring_string(sstack.pop()))
@i: <f              -- Pull from float stack with conversion (shorthand for bring_float(fstack.pop()))
```

A double angle bracket performs a peek operation without consuming the source value:

```lua
@f: <<s           -- Peek from string stack with conversion (doesn't consume the source value)
```

This is equivalent to:

```lua
fstack.bring_string(sstack.peek())
```

### 5.4 Mathematical Expression Integration

In addition to traditional stack operations, stacked mode can incorporate direct mathematical expressions for improved readability:

```lua
@fstack: alias:"f"; @sstack: alias:"s"

@s: push:"25.5"
@f: <s dup (9/5)*32 sum  -- Convert C to F using direct mathematical notation
@s: <f                   -- Convert result back to string
```

This is equivalent to the more traditional RPN approach:

```lua
@s: push:"25.5"
@f: <s dup push:9 push:5 div mul push:32 add
@s: <f
```

The ability to use direct mathematical expressions within stacked mode combines the efficiency of stack operations with the readability of familiar mathematical notation, particularly beneficial for complex calculations.

### 5.5 Combining Features for Expressive Code

These features can be combined for highly expressive and readable code:

```lua
@Stack.new(Float): alias:"f"
@Stack.new(String): alias:"s"

function celsius_to_fahrenheit(celsius_str)
  @s: push(celsius_str)
  @f: <s dup (9/5)*32 sum  -- Direct mathematical expression
  return f.pop()
end
```

The combination of stack aliases, shorthand cross-stack operations, and integrated mathematical expressions creates a powerful yet concise syntax for stack-oriented programming.

## 6. Stack of Stacks

### 6.1 Stack as First-Class Type

Stacks themselves can be stored in other stacks, enabling meta-operations:

```lua
sostack = Stack.new(Stack)  -- A stack that holds other stacks
temp_stack = Stack.new(Integer)
temp_stack.push(42)

sostack.push(temp_stack)    -- Push an entire stack onto the stack-of-stacks
retrieved = sostack.pop()   -- Pop a stack off the stack-of-stacks
value = retrieved.pop()     -- Use the retrieved stack (value = 42)
```

### 6.2 Stack Manipulation at Meta Level

Stack-of-stacks enables powerful patterns for managing computational state:

```lua
-- Save current state of all stacks
function save_state(sostack)
  sostack.push(dstack.clone())  -- Push a copy of dstack
  sostack.push(rstack.clone())  -- Push a copy of rstack
end

-- Restore previous state
function restore_state(sostack)
  rstack = sostack.pop()  -- Restore rstack
  dstack = sostack.pop()  -- Restore dstack
end

-- Usage
sostack = Stack.new(Stack)
save_state(sostack)
-- ... do operations ...
restore_state(sostack)  -- Return to previous state
```

### 6.3 Nested Computation Environments

Stack-of-stacks enables creating isolated environments for subroutines:

```lua
function with_new_env(sostack, func)
  -- Save current stacks
  old_dstack = dstack
  old_rstack = rstack
  
  -- Create fresh stacks
  dstack = Stack.new(Integer)
  rstack = Stack.new(Integer)
  
  -- Execute function in new environment
  result = func()
  
  -- Restore original stacks
  dstack = old_dstack
  rstack = old_rstack
  
  return result
end
```

## 7. Type-Specific Stack Operations

### 7.1 Integer Stack Operations

Integer stacks provide methods specific to integer manipulation:

```lua
istack.and()      -- Bitwise AND of top two values
istack.or()       -- Bitwise OR of top two values
istack.xor()      -- Bitwise XOR of top two values
istack.shl()      -- Shift left top value by second value
istack.shr()      -- Shift right top value by second value
istack.mod()      -- Modulo operation
istack.clz()      -- Count leading zeros
```

### 7.2 Float Stack Operations

Float stacks provide methods for floating-point mathematics:

```lua
fstack.sin()      -- Sine of top value
fstack.cos()      -- Cosine of top value
fstack.tan()      -- Tangent of top value
fstack.sqrt()     -- Square root of top value
fstack.pow()      -- Power operation (x^y)
fstack.round()    -- Round to nearest integer
fstack.floor()    -- Round down to integer
fstack.ceil()     -- Round up to integer
```

### 7.3 String Stack Operations

String stacks provide methods for string manipulation:

```lua
sstack.concat()        -- Concatenate top two strings
sstack.substring()     -- Extract substring
sstack.length()        -- Push string length
sstack.find()          -- Find substring
sstack.replace()       -- Replace substring
sstack.split(delim)    -- Split string into multiple values on stack
```

### 7.4 Boolean Stack Operations

Boolean stacks provide logical operations:

```lua
bstack.and()      -- Logical AND
bstack.or()       -- Logical OR
bstack.not()      -- Logical NOT
bstack.xor()      -- Logical XOR
```

### 7.5 Common Operations

All stack types provide the standard stack operations:

```lua
stack.push(value)    -- Push value onto stack
stack.pop()          -- Remove and return top value
stack.peek()         -- Return top value without removing
stack.dup()          -- Duplicate top value
stack.swap()         -- Swap top two values
stack.over()         -- Copy second value to top
stack.depth()        -- Return current stack depth
stack.clone()        -- Create a copy of the entire stack
```

## 8. Implementation Details

### 8.1 Hardware vs. Software Floating-Point

For platforms with different floating-point capabilities, ual provides a consistent interface with optimized implementations:

```lua
-- Create a float stack
fstack = Stack.new(Float)  -- Uses hardware float when available
```

The compile-time flag `USE_SOFTWARE_FLOAT` can override hardware floating-point:

```
ualc program.ual -DUSE_SOFTWARE_FLOAT  -- Forces software floating-point
```

The implementation automatically selects the appropriate floating-point approach:
- Hardware floating-point when supported by the platform (default)
- Software floating-point for platforms without FPU
- Software floating-point when explicitly requested

### 8.2 Type Checking Implementation

Type checking for typed stacks is implemented as follows:

1. **Compile-time checking** for literals and constant expressions
2. **Runtime validation** for values that can't be type-checked at compile time
3. **No implicit type conversion** - all conversions must be explicit through `bring_<type>` methods

### 8.3 Integration with Existing Code

For backward compatibility, code that uses untyped stacks continues to work:

```lua
-- Legacy code (still valid)
stack = Stack.new()
stack.push(42)
stack.push("string")  -- Works because stack is untyped (Any)
```

### 8.4 Performance Considerations

The typed stack system is designed for minimal overhead:

1. **Type information**: Stored once per stack, not per value
2. **Type checking**: Performed at compile time when possible
3. **Inlining**: Common type operations are inlined
4. **No boxing**: Primitive types are not wrapped or boxed

For most operations, typed stacks have identical runtime performance to untyped stacks, with the added benefit of earlier error detection.

## 9. Use Cases and Examples

### 9.1 Simple Calculator

```lua
package calculator

import "fmt"

function calculate(expression)
  @Stack.new(Integer): alias:"i"  -- Stack for integer operations
  @Stack.new(Float): alias:"f"    -- Stack for floating-point operations
  @Stack.new(String): alias:"s"   -- Stack for parsing

  -- Parse the expression
  @s: push(expression)
  parse_expression(s, f)
  
  return f.pop()
end

function parse_expression(sstack, fstack)
  -- Split the expression into tokens
  @sstack: split:" "
  
  -- Process each token
  while_true(sstack.depth() > 0)
    token = sstack.pop()
    
    -- Try to convert token to number
    if is_numeric(token) then
      @fstack: bring_string(token)
    else
      -- Handle operators
      if token == "+" then
        @fstack: add
      elseif token == "-" then
        @fstack: sub
      elseif token == "*" then
        @fstack: mul
      elseif token == "/" then
        @fstack: div
      elseif token == "sin" then
        @fstack: sin
      elseif token == "cos" then
        @fstack: cos
      end
    end
  end_while_true
end

function is_numeric(str)
  -- Check if string represents a valid number
  -- Implementation omitted for brevity
end
```

### 9.2 Temperature Converter with New Syntax

```lua
package temperature

import "fmt"

function celsius_to_fahrenheit(celsius_str)
  -- Create and alias stacks for better readability
  @Stack.new(String): alias:"s"
  @Stack.new(Float): alias:"f"
  
  -- Convert from string to float
  @s: push(celsius_str)
  @f: <s  -- Pull from string stack with conversion
  
  -- Calculate F = C * 9/5 + 32 using direct mathematical expression
  @f: dup (9/5)*32 sum
  
  -- Convert result back to string
  @s: <f  -- Pull from float stack with conversion
  
  return s.pop()
end

function main()
  temp_c = "25.5"
  temp_f = celsius_to_fahrenheit(temp_c)
  fmt.Printf("%s°C = %s°F\n", temp_c, temp_f)
  return 0
end
```

### 9.3 Stack-of-Stacks Example

```lua
package state_machine

function process_transaction(transaction_data)
  -- Create a stack to hold computational state stacks
  @Stack.new(Stack): alias:"states"
  
  -- Initialize processing stacks
  @Stack.new(Any): alias:"data"
  @data: push(transaction_data)
  
  -- Process transaction through multiple states
  process_state("validate", data, states)
  process_state("authorize", data, states)
  process_state("execute", data, states)
  
  -- Check if we need to rollback
  if data.peek().status == "error" then
    -- Rollback states in reverse order
    while_true(states.depth() > 0)
      state = states.pop()
      rollback_state(state, data)
    end_while_true
    
    return false
  end
  
  return true
end

function process_state(state_name, data, states)
  -- Save current state before processing
  @Stack.new(Any): alias:"state_stack"
  @state_stack: push(state_name)
  @state_stack: push(data.clone())  -- Save copy of data stack
  
  -- Add to states history
  @states: push(state_stack)
  
  -- Process this state
  -- (implementation varies by state_name)
end

function rollback_state(state_stack, data)
  -- Retrieve state information
  saved_data = state_stack.pop()
  state_name = state_stack.pop()
  
  -- Execute rollback logic for this state
  -- (implementation varies by state_name)
  
  fmt.Printf("Rolling back state: %s\n", state_name)
end
```

### 9.4 Cross-Stack Data Processing

This example demonstrates the atomic `bring_<type>` operation and shorthand syntax for complex data processing:

```lua
package data_processor

import "fmt"

function process_sensor_data(raw_data)
  -- Create typed stacks for different parts of processing
  @Stack.new(String): alias:"s"   -- For raw input and text processing
  @Stack.new(Float): alias:"f"    -- For numerical calculations
  @Stack.new(Integer): alias:"i"  -- For digital signal processing
  @Stack.new(Boolean): alias:"b"  -- For threshold detection

  -- Parse the raw data
  @s: push(raw_data)
  @s: split:","  -- Split CSV format data

  -- Process temperature reading
  @f: <s                       -- Bring from string to float (atomic operation)
  @f: push:273.15 add         -- Convert to Kelvin
  temp_kelvin = f.peek()
  
  -- Process pressure reading
  @f: <s                       -- Bring string to float
  @f: push:1000 div           -- Convert to kilopascals
  pressure_kpa = f.peek()
  
  -- Check threshold values
  @i: <f                       -- Convert float to integer
  @i: push:300 gt             -- Compare with threshold
  @b: <i                       -- Convert comparison result to boolean
  threshold_exceeded = b.pop()
  
  -- Format results
  @s: push("Temperature: ")
  @s: <f push:" K, Pressure: " <f push:" kPa")
  @s: concat concat concat concat
  
  return s.pop(), threshold_exceeded
end
```

### 9.5 Error Handling with .consider

```lua
package validation

function parse_numeric_input(input_str)
  @Stack.new(String): alias:"s"
  @Stack.new(Float): alias:"f"
  
  result = {}
  
  @s: push(input_str)
  
  -- Try to convert to float with error handling
  success, err = pcall(function()
    @f: <s  -- Shorthand for bring_string(s.pop())
    result.Ok = f.pop()
  end)
  
  if not success then
    result.Err = "Invalid number format: " .. input_str
  end
  
  return result
end

function process_user_input(input_str)
  -- Use the .consider construct for elegant error handling
  parse_numeric_input(input_str).consider {
    if_ok  process_valid_number(_1)
    if_err request_new_input(_1)
  }
end
```

## 10. Comparison with Other Languages

### 10.1 Forth

Traditional Forth has no type system, treating all values as untyped cells. ual's typed stacks provide stronger guarantees while maintaining stack-based operations.

### 10.2 Factor

Factor has a sophisticated static type system with inference. ual's approach is simpler but provides practical type safety for embedded applications.

### 10.3 Typed Assembly Languages

ual's approach resembles Typed Assembly Languages (TALs) that add type annotations to low-level code, but focuses specifically on stack operations.

## 11. Limitations and Future Directions

### 11.1 Current Limitations

1. No user-defined types or interfaces
2. Limited parametric polymorphism
3. No type inference beyond literal values
4. No union or intersection types

### 11.2 Future Directions

1. Stack type parameters (e.g., `Stack.new(Array(Integer))`)
2. User-defined stack types with validation
3. Enhanced type inference for complex expressions
4. Integration with the ual macro system for compile-time type checking

## 12. Backward Compatibility Considerations

### 12.1 Syntax Evolution Strategy

The introduction of the colon syntax alongside the existing angle bracket syntax presents both opportunities and challenges. To ensure a smooth transition:

1. **Both syntaxes remain valid**: All ual 1.3 code continues to work without modification
2. **Gradual migration**: Developers can migrate code to the new syntax at their own pace
3. **New features prefer new syntax**: Features like aliases work with both syntaxes but are designed for the colon syntax
4. **Documentation evolution**: New examples use colon syntax exclusively

### 12.2 Mixing Syntax Styles

Developers can mix syntax styles within a single codebase or even within a single function:

```lua
-- Legacy style
@dstack > push:42 dup add

-- New style
@fstack: push:3.14 dup mul

-- Both are valid and can be used together
```

While mixing styles is supported, maintaining consistent style within a module is recommended for readability.

### 12.3 Automated Migration

To assist developers in transitioning to the new syntax, a source code transformation tool will be provided:

```
ual-syntax-updater input.ual --output=updated.ual
```

This tool will transform angle bracket syntax to colon syntax while preserving all semantics.

## 13. Conclusion

The proposed typed stack system for ual provides a pragmatic approach to type safety that aligns with the language's focus on embedded systems. By adding type constraints to stacks rather than implementing a comprehensive static type system, ual maintains its simplicity and efficiency while addressing many common sources of errors.

The introduction of the colon syntax, stack aliases, and integrated mathematical expressions significantly improves the readability and expressiveness of stack-oriented code. Meanwhile, the atomic `bring_<type>` operation provides safe and efficient cross-stack operations with clear semantics.

Typed stacks offer several key benefits:
- Catching type errors earlier in the development process
- Enabling specialized operations for different data types
- Providing explicit, clear conversion between types
- Supporting platform-specific optimizations
- Enabling meta-programming through stacks-of-stacks

This approach represents a middle ground between untyped stack languages and statically typed languages, creating a unique niche that's well-suited for embedded systems programming where both resource constraints and reliability are important.