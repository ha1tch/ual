# ual 1.5 PROPOSAL: Defer Stack Mechanism 

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the version that the proposal is targeting to be integrated into the main ual spec in a forthcoming release.

---

## 1. Introduction

This document proposes a `defer` mechanism for the ual programming language, implemented through a dedicated `@defer` stack and offering a convenient `defer_op` syntactic sugar. This feature aims to simplify resource management and enhance error handling robustness, while staying true to ual's stack-based paradigm and zero-runtime-overhead philosophy, making it particularly valuable for embedded systems programming.

## 2. Background and Motivation

### 2.1 The Need for Guaranteed Cleanup

In robust programming, especially in resource-constrained environments like embedded systems, ensuring proper resource cleanup is critical. Resources like file handles, memory allocations, hardware peripherals, and network connections must be reliably released to prevent leaks and maintain system stability. Common scenarios requiring guaranteed cleanup include:

1. **File Operations:** Closing files after use, even if errors occur during processing.
2. **Memory Management:** Deallocating dynamically allocated memory to avoid leaks.
3. **Hardware Access:** Releasing control of peripherals or disabling clocks.
4. **Synchronization Primitives:** Releasing locks or mutexes.

### 2.2 Challenges with Manual Cleanup

Manual cleanup, while common, is error-prone:

1. **Forgetting to Cleanup:** Developers might simply forget to add cleanup code, especially in complex functions with multiple exit points.
2. **Error Handling Complexity:** Ensuring cleanup in all error paths can lead to verbose and repetitive code.
3. **Code Maintenance:** Changes to code flow might inadvertently break cleanup logic.

### 2.3 The `defer` Solution

Many modern languages offer a `defer` mechanism to address these challenges. `defer` allows developers to schedule a function call or code block to be executed automatically when the current function or scope exits, guaranteeing cleanup regardless of how the scope is exited (normal completion or early return due to errors).

For ual, a `defer` mechanism should:

1. Integrate seamlessly with ual's stack-based nature.
2. Introduce minimal new syntax and cognitive overhead.
3. Have zero runtime performance cost.
4. Enhance code readability and maintainability, especially in error handling scenarios.

## 3. Design Principles

The proposed `defer` mechanism adheres to these core principles:

1. **Stack-Based Implementation**: Leverage ual's stack paradigm using a dedicated `@defer` stack.
2. **Zero Runtime Overhead**: Deferred action scheduling and execution happen at compile time.
3. **Guaranteed Execution**: Deferred actions are always executed upon scope exit.
4. **Syntactic Sugar for Common Use Cases**: Provide `defer_op` for concise syntax in typical scenarios.
5. **Explicit and Predictable**: Deferral behavior should be clear and easy to understand.
6. **Embedded Systems Focus**: Designed for efficiency and minimal resource usage.

## 4. Proposed Implementation: `@defer` Stack

### 4.1 The `@defer` Stack

This proposal introduces a predefined stack named `@defer`, automatically available alongside `@dstack`, `@rstack`, and `@error`. The `@defer` stack is used to store *deferred action blocks*.

### 4.2 Deferring Actions: `@defer: push { ... }`

To schedule a code block for deferred execution, use the syntax:

```lua
@defer: push {
  -- Code block to be executed on scope exit
  -- (Typically cleanup operations)
}
```

The code block pushed onto the `@defer` stack is standard ual code, allowing full access to ual's features and stack operations.  Note the use of the **colon `:` as the stack selector** for `@defer` in stacked mode.

### 4.3 Scope Exit Processing

The ual compiler will automatically insert code at the end of each scope (currently functions and `do` blocks are considered scopes) to process the `@defer` stack:

1. **Scope Exit Point:**  At the point where a scope (function, `do` block) is about to exit (either by reaching the `end` keyword or a `return` statement).
2. **Process `@defer` Stack:** The compiler-generated code will execute a loop that continues as long as the `@defer` stack is not empty:
   - **Pop Action Block:** Pop a code block from the top of the `@defer` stack.
   - **Execute Action Block:** Execute the popped code block.

3. **LIFO Execution Order:** Due to the stack nature, deferred actions will be executed in LIFO (Last-In, First-Out) order, which is the standard and generally desired behavior for resource cleanup.

### 4.4 Example with `@defer` Stack:

```lua
function process_file_defer_stack(filename: String)
  @Stack.new(File, Owned): alias:"file_stack"

  @file_stack: push(io.open(filename, "r")) -- Acquire resource

  -- Defer file closing using @defer stack
  @defer: push {
    @file_stack: depth() if_true {
      @file_stack: pop() dup if_true {
        pop().close() -- Close the file
      } drop
    }
  }

  -- ... Function body: process file content ...

  -- No explicit file close needed here.
  -- @defer stack will automatically close the file when this function exits.
end
```

## 5. Proposed Implementation: `defer_op` Sugar

### 5.1 `defer_op { ... }` Syntax

For the common use case of simply deferring a code block for scope exit, this proposal introduces `defer_op` as syntactic sugar:

```lua
defer_op {
  -- Code block to be deferred
}
```

### 5.2 Compiler Transformation

The compiler will automatically transform `defer_op { ... }` into the more explicit `@defer: push { ... }` form during parsing.  Specifically:

```
defer_op {
  -- Code block
}
```

is transformed into:

```lua
@defer: push {
  -- Code block (exactly the same block)
}
```

### 5.3 Example with `defer_op` Sugar:

```lua
function process_file_defer_sugar(filename: String)
  @Stack.new(File, Owned): alias:"file_stack"

  @file_stack: push(io.open(filename, "r")) -- Acquire resource

  -- Defer file closing using defer_op sugar
  defer_op {
    @file_stack: depth() if_true {
      @file_stack: pop() dup if_true {
        pop().close() -- Close the file
      } drop
    }
  }

  -- ... Function body: process file content ...

  -- No explicit file close needed here.
  -- defer_op block will automatically close the file.
end
```

### 5.4 Choosing Between `@defer: push` and `defer_op`

- **`@defer: push { ... }`**: Use for more explicit control or when you need to directly manipulate the `@defer` stack (although direct manipulation is generally discouraged for typical `defer` usage). Favored when explicitness about stack operations is desired.
- **`defer_op { ... }`**: Use for the common, simple case of deferring a block of code for scope exit. Favored for readability and conciseness in typical resource management scenarios, providing syntactic sugar for the underlying `@defer` stack mechanism.

## 6. Use Cases and Examples

### 6.1 Simple File Handling with `defer_op`

```lua
function read_config_file(filename: String): String {
  @Stack.new(File, Owned): alias:"config_file"
  @config_file: push(io.open(filename, "r"))

  defer_op { -- Ensure file is closed on exit
    @config_file: depth() if_true {
      @config_file: pop() dup if_true { pop().close() } drop
    }
  }

  if @error: depth() > 0 then return "" end -- Handle file open errors elsewhere, note colon notation

  local content = config_file.peek().read("*all")
  return content
}
```

### 6.2 Multiple Deferred Actions

```lua
function manage_multiple_resources()
  @Stack.new(Resource, Owned): alias:"res1_stack"
  @Stack.new(Resource, Owned): alias:"res2_stack"

  @res1_stack: push(acquire_resource_1())
  defer_op { release_resource_1(res1_stack.pop()) } -- Defer release of resource 1

  @res2_stack: push(acquire_resource_2())
  defer_op { release_resource_2(res2_stack.pop()) } -- Defer release of resource 2

  -- ... Function body: use res1_stack and res2_stack ...

  -- Resources 2 and then 1 will be released in LIFO order when function exits.
end
```

### 6.3 Defer in `do` Blocks

```lua
function process_data_block(data)
  do -- Start a do block for scoped resource management
    @Stack.new(Buffer, Owned): alias:"buffer_stack"
    @buffer_stack: push(allocate_buffer(data.size))

    defer_op { release_buffer(buffer_stack.pop()) } -- Defer buffer release

    -- ... Process data using buffer_stack ...
    process_internal_data(buffer_stack.peek(), data)

  end -- do block ends, buffer_stack and deferred action go out of scope.
  -- Buffer is automatically released here.

  -- ... Continue processing with main function scope ...
  finalize_data_processing(data)
end
```

## 7. Implementation Considerations

### 7.1 Compiler Implementation

The compiler implementation would involve:

1. **Parsing `defer_op`:** Recognize `defer_op { ... }` syntax and parse the code block within it.
2. **Transformation to `@defer: push`:** Transform `defer_op` into `@defer: push` representation in the intermediate representation (AST or similar).
3. **Scope Exit Code Injection:** At each function and `do` block exit point, inject code to process the `@defer` stack (loop to pop and execute actions).
4. **Zero Runtime Overhead:** Ensure that the defer mechanism itself introduces no runtime performance cost. The generated code for deferred actions will be executed, but the scheduling and management of deferral should be compile-time.

### 7.2 `@defer` Stack Representation

The `@defer` stack itself is a compile-time construct.  At runtime, it does not exist as a separate stack data structure. The compiler directly generates the code to execute the deferred action blocks in LIFO order at scope exit.

### 7.3 Interaction with Error Handling

Deferred actions should generally execute even if errors occur in the main function body. If an error occurs *within* a deferred action itself, the behavior needs to be defined:

- **Option 1 (Ignore Errors in Defer):** Errors within deferred actions are silently ignored (simplest, but might hide issues).
- **Option 2 (Push to `@error` Stack):** Errors within deferred actions are pushed onto the `@error` stack, allowing them to be potentially handled by an enclosing error handler (more robust, but might require careful consideration of error contexts).

Option 2 (push to `@error` stack) is likely the more robust and ual-idiomatic approach, but requires further consideration of error context and potential for cascading errors.

## 8. Comparison with Other Languages

### 8.1 Compared to Go's `defer`

Go's `defer` is a first-class language feature and is very widely used for resource management. ual's `defer_op` sugar is directly inspired by Go's `defer`, aiming to provide similar convenience within ual's stack-based context.  Go's `defer` is runtime-based, while ual's is designed for compile-time guarantees and zero runtime overhead.

### 8.2 Compared to Rust's RAII (Resource Acquisition Is Initialization)

Rust heavily relies on RAII and destructors for automatic resource management. While very effective and zero-cost, Rust's approach is tied to its ownership and borrowing system and doesn't have a direct equivalent of `defer`. ual's `@defer` stack and `defer_op` offer a more explicit and potentially more flexible mechanism for deferred actions, while still aiming for zero runtime cost through compile-time implementation.

## 9. Backward Compatibility

This proposal introduces a new feature (`@defer` stack and `defer_op` sugar). It does not directly impact backward compatibility with existing ual code. Existing code will simply not use the `defer` mechanism and will continue to function as before.  Adoption of `defer` is entirely opt-in and progressive.

## 10. Future Directions

### 10.1 Error Handling in Deferred Actions

Further define the behavior when errors occur *within* deferred action blocks (e.g., push to `@error` stack, or other mechanisms).

### 10.2 Integration with Ownership System

Explore potential deeper integration with ual's ownership system, perhaps allowing deferred actions to interact with ownership transfer or borrowing rules.

### 10.3 Debugging Support for Deferred Actions

Enhance debugging tools to provide visibility into deferred actions and their execution order.

## 11. Conclusion

The proposed `@defer` stack mechanism with `defer_op` syntactic sugar provides a powerful and ual-idiomatic solution for guaranteed resource cleanup. By leveraging the stack paradigm and offering both explicit and concise syntax options, ual gains a valuable feature for writing more robust, maintainable, and resource-safe code, particularly beneficial for embedded systems programming where reliability and efficiency are paramount. The zero-runtime-overhead design ensures that ual retains its performance characteristics while significantly enhancing its capabilities for resource management and error handling. We recommend the adoption of this `@defer` mechanism for ual to enhance its practicality and developer experience.