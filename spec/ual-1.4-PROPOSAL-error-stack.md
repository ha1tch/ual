# ual 1.4 PROPOSAL: Error Stack Mechanism

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction

This document proposes an error handling mechanism for the ual programming language based on a dedicated error stack. Rather than introducing new syntax or borrowing error handling patterns from other languages, this proposal leverages ual's existing stack paradigm to create a native, cohesive solution for reliable error management. The proposed approach is particularly well-suited for embedded systems where resource constraints demand zero-overhead error handling with compile-time guarantees.

## 2. Background and Motivation

### 2.1 The Need for Robust Error Handling

Embedded systems frequently encounter conditions that require error handling:

1. **Hardware Interaction**: Sensors may fail, communication buses may time out, or peripherals may return invalid data.

2. **Resource Constraints**: Memory allocation failures, buffer overflows, or insufficient stack space.

3. **Environmental Factors**: Power fluctuations, temperature extremes, or electromagnetic interference.

4. **Input Validation**: Parsing configuration data, processing user inputs, or validating sensor readings.

Without reliable error handling, these conditions can lead to system crashes, undefined behavior, or silent data corruption.

### 2.2 Limitations of Traditional Approaches

Existing error handling mechanisms have significant drawbacks in embedded contexts:

1. **Exception Handling**: Traditional exceptions require runtime support, stack unwinding, and RTTI, all of which consume valuable resources.

2. **Return Code Checking**: Easy to ignore, leads to repetitive boilerplate, and pollutes control flow.

3. **Global Error States**: Creates hidden dependencies and complicates multi-threaded scenarios.

4. **Option Types/Result Types**: Often require language features (generics, sum types) that add complexity.

### 2.3 The ual Opportunity

ual's dual-paradigm approach (variable-based and stack-based) provides a unique opportunity to implement error handling that:

1. Fits naturally within the language's existing conceptual model
2. Introduces minimal cognitive overhead
3. Works seamlessly in both variable and stack contexts
4. Provides compile-time guarantees with zero runtime cost
5. Maintains ual's progressive disclosure principle

## 3. Design Principles

The proposed error stack mechanism adheres to these core principles:

1. **Use What's Already There**: Leverage ual's existing stack paradigm rather than inventing new concepts.

2. **Zero Runtime Overhead**: All error checking happens at compile time with no impact on binary size or execution speed.

3. **Mandatory Handling**: Errors cannot be silently ignored; they must be explicitly handled or propagated.

4. **Syntactic Consistency**: Error handling uses familiar ual syntax and conventions.

5. **Paradigm Unification**: The same error mechanism works in both variable-based and stack-based code.

6. **Progressive Discovery**: Simple errors are simple to handle, while advanced patterns build naturally on basic concepts.

## 4. Proposed Implementation

### 4.1 The @error Stack

The proposal introduces a predefined stack named `@error` that exists alongside the existing `dstack` and `rstack`:

```lua
-- The error stack is automatically available like dstack and rstack
@error > depth()  -- Check if there are any errors
@error > push("File not found")  -- Push an error
err = @error > pop()  -- Pop an error
```

### 4.2 Error-Capable Function Declaration

Functions that can produce errors are marked with the `@error >` prefix:

```lua
@error > function read_file(filename)
  if file_not_accessible then
    @error > push("Cannot access file: " .. filename)
    return nil
  end
  return file_contents
end
```

This annotation serves two purposes:
1. Documents that the function may push to the error stack
2. Enables compiler tracking of potential error states

### 4.3 Compiler Tracking and Enforcement

The compiler tracks the potential state of the `@error` stack throughout program execution:

1. After a call to an `@error >` function, the compiler considers the error stack to be potentially non-empty.

2. Before control flow exits a context where an `@error >` function was called, the code must either:
   - Check and handle the error stack, or
   - Propagate the error by being in a function also marked with `@error >`

3. At program termination points, the `@error` stack must be empty.

Violations of these rules result in compile-time errors, ensuring all errors are addressed.

### 4.4 Error Checking Patterns

The proposal supports multiple patterns for checking and handling errors:

#### 4.4.1 Direct Stack Inspection

```lua
@error > function example()
  read_file("config.txt")  -- Might push to @error stack
  
  if @error > depth() > 0 then
    err = @error > pop()
    fmt.Printf("Error: %s\n", err)
    return nil
  end
  
  -- Continue normal operation
end
```

#### 4.4.2 Error Propagation

```lua
@error > function process_file(filename)
  content = read_file(filename)  -- Might push to @error stack
  
  -- No explicit check means errors automatically propagate
  -- to caller (because this function is marked @error >)
  
  if content == nil then
    return nil  -- Early return if operation failed
  end
  
  return process(content)
end
```

#### 4.4.3 Helper Functions

The standard library would provide helper functions for common error handling patterns:

```lua
function try_or_default(function_call, default_value)
  result = function_call()
  if @error > depth() > 0 then
    @error > drop()  -- Discard the error
    return default_value
  end
  return result
end

-- Usage
value = try_or_default(function() return read_file("config.txt") end, "")
```

### 4.5 Stack Mode Integration

The error stack system works seamlessly with stack mode operations:

```lua
@error > function calculate_result()
  > read_sensor dup
  > @error depth if_true
  >   drop  -- Drop sensor reading if there was an error
  >   return nil  -- Error remains on @error stack
  > end_if_true
  
  > process_reading
  return pop()
end
```

## 5. Use Cases and Examples

### 5.1 Basic Error Handling

```lua
function main()
  @error > function setup_device()
    if not initialize_hardware() then
      @error > push("Hardware initialization failed")
      return false
    end
    return true
  end
  
  success = setup_device()
  if @error > depth() > 0 then
    err = @error > pop()
    fmt.Printf("Setup failed: %s\n", err)
    sys.Exit(1)
  end
  
  fmt.Printf("Device ready\n")
  -- Continue with normal operation
end
```

### 5.2 Error Chaining

```lua
@error > function process_config()
  content = read_file("config.txt")
  if @error > depth() > 0 then
    -- Enhance the error with context, but preserve the original
    original = @error > pop()
    @error > push("Config processing failed: " .. original)
    return nil
  end
  
  -- Process config
  return parsed_config
end
```

### 5.3 Multiple Error Sources

```lua
@error > function initialize_system()
  if not init_memory() then
    @error > push("Memory initialization failed")
    return false
  end
  
  if not init_peripherals() then
    @error > push("Peripheral initialization failed")
    return false
  end
  
  return true
end
```

### 5.4 Advanced Stack-Based Error Handling

```lua
@error > function process_sensor_data()
  > read_sensor_1 read_sensor_2 read_sensor_3
  > @error depth if_true
  >   -- Clear data stack if any sensor read failed
  >   drop drop drop
  >   return false
  > end_if_true
  
  > add div  -- Process readings
  return true
end
```

### 5.5 Result Pattern

The error stack can be combined with a result pattern for more complex returns:

```lua
@error > function compute_result(input)
  if input < 0 then
    @error > push("Input must be non-negative")
    return nil
  end
  
  -- Process and return complex result
  return {
    value = input * 2,
    processed = true,
    timestamp = sys.Millis()
  }
end
```

## 6. Implementation Considerations

### 6.1 Compiler Implementation

The compiler would track error stack states through a static analysis pass:

1. Mark functions annotated with `@error >`
2. Track calls to these functions
3. Verify all paths either check `@error > depth()` or are in `@error >` functions
4. Generate compile-time errors for unhandled error states

This approach requires no runtime overhead, as all checks are performed during compilation.

### 6.2 Error Stack Representation

The `@error` stack would be implemented just like any other stack in ual, with no special runtime handling required. The only difference is the compiler's static tracking of potential error states.

### 6.3 Integration with Existing Features

The error stack mechanism integrates seamlessly with ual's existing features:

1. **Stack Operations**: The `@error` stack supports all standard stack operations.
2. **Switch Statement**: Can be used to handle different error types.
3. **Macros**: Can generate error handling code.
4. **Conditional Compilation**: Can adapt error handling based on target platform.

### 6.4 Error Context and Information

While the basic mechanism uses string errors, extensions could support richer error information:

```lua
@error > function complex_operation()
  if failure_condition then
    @error > push({
      code = ERROR_IO_FAILURE,
      message = "Operation failed",
      subsystem = "file_io",
      timestamp = sys.Millis()
    })
    return nil
  end
  return result
end
```

## 7. Comparison with Other Languages

### 7.1 Compared to Zig

Zig uses explicit error handling with `try` and error return types:

**Zig:**
```zig
fn readFile(filename: []const u8) ![]u8 {
    // Implementation
}

// Usage
const content = try readFile("file.txt");
```

**ual (Proposed):**
```lua
@error > function read_file(filename)
  // Implementation
end

// Usage
content = read_file("file.txt")
if @error > depth() > 0 then
  // Handle error
end
```

Both approaches ensure errors are handled with zero runtime cost, but ual's approach maintains its stack-based paradigm.

### 7.2 Compared to Rust

Rust uses Result types with pattern matching:

**Rust:**
```rust
fn read_file(filename: &str) -> Result<String, io::Error> {
    // Implementation
}

// Usage
match read_file("file.txt") {
    Ok(content) => process(content),
    Err(e) => println!("Error: {}", e),
}
```

**ual (Proposed):**
```lua
@error > function read_file(filename)
  // Implementation
end

// Usage
content = read_file("file.txt")
if @error > depth() > 0 then
  err = @error > pop()
  fmt.Printf("Error: %s\n", err)
else
  process(content)
end
```

Rust's approach requires a more complex type system, while ual's approach builds on existing stack primitives.

### 7.3 Compared to Lua

Lua uses return values for errors:

**Lua:**
```lua
function readFile(filename)
  local file, err = io.open(filename, "r")
  if not file then return nil, err end
  local content = file:read("*all")
  file:close()
  return content
end

-- Usage
local content, err = readFile("file.txt")
if err then print("Error: " .. err) end
```

**ual (Proposed):**
```lua
@error > function read_file(filename)
  file = io.open(filename, "r")
  if file == nil then
    @error > push("Could not open file")
    return nil
  end
  content = file.read("*all")
  file.close()
  return content
end

-- Usage
content = read_file("file.txt")
if @error > depth() > 0 then
  err = @error > pop()
  fmt.Printf("Error: %s\n", err)
end
```

Lua's approach is convenient but doesn't enforce error checking, while ual's approach provides compile-time guarantees.

## 8. Future Directions

While the core proposal focuses on a clean, minimal implementation, future enhancements could include:

1. **Error Categories**: Standardized error types for different subsystems.
2. **Error Stack Inspection**: Functions to examine error stack without popping.
3. **Cleanup Actions**: Guaranteed resource cleanup during error propagation.
4. **Stack Trace Support**: Adding location information to errors.
5. **Error Recovery**: Mechanisms for recovering from specific error conditions.

## 9. Conclusion

The proposed error stack mechanism provides a cohesive, zero-overhead approach to error handling in ual that:

1. Leverages the language's existing stack paradigm rather than introducing foreign concepts.
2. Works seamlessly in both variable-based and stack-based code.
3. Provides compile-time guarantees of error handling.
4. Introduces minimal new syntax while maintaining readability.
5. Scales from simple to complex error handling scenarios.

By implementing error handling through the `@error` stack, ual maintains its identity as a hybrid stack/variable language while gaining safety guarantees comparable to languages like Zig and Rust. This approach aligns perfectly with ual's design philosophy and embedded systems focus, providing robust error handling without sacrificing performance or simplicity.