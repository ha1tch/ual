# ual 1.5 PROPOSAL: Stack-Based Ownership System

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction

This document proposes a stack-based ownership system for the ual programming language that builds upon the typed stacks introduced in ual 1.4. The proposed system provides memory safety guarantees comparable to Rust's borrow checker but through a container-centric model that aligns with ual's stack-based paradigm. This approach maintains ual's focus on embedded systems while adding powerful safety guarantees with zero runtime overhead.

## 2. Background and Motivation

### 2.1 The Memory Safety Challenge

Memory safety bugs remain one of the most significant sources of vulnerabilities and crashes in software systems. These bugs typically manifest as:

1. **Use-after-free**: Accessing memory that has already been deallocated
2. **Double-free**: Attempting to deallocate the same memory multiple times
3. **Dangling pointers**: References to memory that no longer contains valid data
4. **Data races**: Concurrent access to shared memory without proper synchronization

Traditional approaches to this problem include garbage collection (with runtime overhead), manual memory management (error-prone), and reference counting (performance impact). Rust introduced a new approach through its ownership system and borrow checker, which provides memory safety guarantees at compile time with zero runtime overhead.

### 2.2 An Overview of the Rust Ownership Model

For developers unfamiliar with Rust, its ownership system is based on three core principles:

1. **Single Ownership**: Each value has exactly one owner variable at any time
2. **Borrowing**: References to values can be "borrowed" either mutably (one exclusive reference) or immutably (multiple shared references)
3. **Lifetimes**: The compiler tracks how long references are valid

For example, in Rust:

```rust
fn process(data: &mut Vec<i32>) {
    // Mutable (exclusive) borrow of data
    data.push(42);
}

fn main() {
    let mut numbers = vec![1, 2, 3];
    
    process(&mut numbers);  // Pass a mutable reference
    
    // This would cause a compile error if process kept a reference to numbers
    println!("{:?}", numbers);
}
```

The Rust compiler's borrow checker analyzes the flow of ownership throughout the program, ensuring that references never outlive the data they point to and that mutable references never coexist with other references to the same data.

While powerful, Rust's approach comes with a steep learning curve. The ownership rules are enforced implicitly at variable assignments and function boundaries, making it sometimes difficult to visualize and understand the flow of ownership.

### 2.3 The ual Opportunity

ual's stack-based paradigm offers a natural alternative model for representing ownership. Instead of associating ownership with variables, we can associate it with stacks—explicit containers that hold values and transfer them according to well-defined rules.

This approach aligns with ual's design philosophy:

1. **Explicitness**: Make ownership transfers visible and intuitive
2. **Zero Runtime Overhead**: All checks performed at compile time
3. **Progressive Discovery**: Simple patterns are simple, complex patterns build naturally
4. **Embedded Systems Focus**: Designed for resource-constrained environments
5. **Dual Paradigm Support**: Works in both stack-based and variable-based code

## 3. ual's Type System: A Different Approach

Before introducing the ownership system, it's important to understand how ual's type system differs from traditional approaches.

### 3.1 Container-Centric vs. Value-Centric Typing

Most programming languages associate types with values or variables:

```python
# Python (value has a type)
x = 42  # x has type int
y = "hello"  # y has type str
```

```rust
// Rust (variable has a type)
let x: i32 = 42;
let y: String = String::from("hello");
```

In contrast, ual associates types with containers (stacks) rather than individual values:

```lua
@Stack.new(Integer): alias:"i"  -- Stack that accepts integers
@Stack.new(String): alias:"s"   -- Stack that accepts strings

@i: push(42)       -- Valid: integer into integer stack
@s: push("hello")  -- Valid: string into string stack
@i: push("hello")  -- Error: string cannot go into integer stack
```

This container-centric approach creates a fundamentally different model for thinking about types:

1. **Boundary Checking**: Type checking happens at container boundaries (when values enter or leave)
2. **Contextual Validity**: Values are valid or invalid based on their context, not their intrinsic nature
3. **Flow-Based Reasoning**: Type safety follows the flow of data between containers

### 3.2 The bring_&lt;type&gt; Operation

A key innovation in ual's type system is the atomic `bring_<type>` operation, which combines popping, type conversion, and pushing:

```lua
@s: push("42")     -- Push string to string stack
@i: bring_string(s.pop())  -- Convert from string to integer during transfer
```

With shorthand notation:

```lua
@s: push("42")
@i: <s            -- Shorthand for bring_string(s.pop())
```

This operation provides critical guarantees:

1. **Atomicity**: The operation either fully succeeds or fully fails
2. **Explicitness**: The type conversion is clearly visible
3. **Efficiency**: No intermediate variables needed

This model creates a natural foundation for thinking about ownership as another property of containers alongside types.

## 4. Proposed Stack-Based Ownership System

### 4.1 Ownership as a Stack Property

The proposal extends ual's typed stacks to include ownership semantics:

```lua
@Stack.new(Integer, Owned): alias:"io"    -- Stack of owned integers
@Stack.new(Float, Borrowed): alias:"fb"   -- Stack of borrowed floats
@Stack.new(String, Mutable): alias:"sm"   -- Stack of mutable string references
```

Each stack enforces both type constraints and ownership rules. Values moving between stacks must comply with both.

### 4.2 Ownership Modes

The system supports three primary ownership modes:

1. **Owned**: The stack owns the values it contains and is responsible for their lifetime
2. **Borrowed**: The stack contains non-mutable references to values owned elsewhere
3. **Mutable**: The stack contains exclusive mutable references to values owned elsewhere

### 4.3 Ownership Transfer Operations

Similar to `bring_<type>` for type conversion, the system introduces operations for ownership transfers:

```lua
-- Take ownership (consumes the source value)
@owned: take(borrowed.pop())

-- Borrow immutably (doesn't consume the source value)
@borrowed: borrow(owned.peek())

-- Borrow mutably (exclusive access, doesn't consume)
@mutable: borrow_mut(owned.peek())
```

With shorthand notation:

```lua
@io: push(42)          -- Push owned integer
@ib: <<io              -- Borrow immutably (shorthand for borrow(io.peek()))
@im: <:mut io          -- Borrow mutably (shorthand for borrow_mut(io.peek()))
```

### 4.4 Combining Type and Ownership Transfers

Operations can combine type and ownership transfers:

```lua
@so: push("42")        -- Push owned string
@ib: <:b so            -- Borrow and convert to integer
@fm: <:mut ib          -- Mutable borrow and convert to float
```

This translates to:

```lua
@ib: bring_string:borrow(so.peek())
@fm: bring_integer:mutable(ib.peek())
```

### 4.5 Lifetime Tracking

The compiler tracks the lifetime of values and references through stack operations:

```lua
function process()
  @Stack.new(Integer, Owned): alias:"io"    -- Owned stack with function scope
  @io: push(42)
  
  @Stack.new(Integer, Borrowed): alias:"ib" -- Borrowed stack with function scope
  @ib: <<io                                 -- Borrow from owned stack
  
  compute(ib.pop())                         -- Use borrowed reference
  
  -- Borrow expires at end of function
}
```

The compiler ensures that borrowed references never outlive their source values by tracking stack lifetimes.

## 5. Detailed Examples

### 5.1 Basic Ownership Flow

```lua
function transfer_example()
  @Stack.new(Integer, Owned): alias:"src"
  @Stack.new(Integer, Owned): alias:"dst"
  
  @src: push(42)       -- Create owned value
  @dst: <:own src      -- Transfer ownership (src loses it)
  
  -- src.pop()         -- Error: value no longer owned by src
  return dst.pop()     -- OK: dst owns the value now
end
```

### 5.2 Borrowing and Mutations

```lua
function borrowing_example()
  @Stack.new(Integer, Owned): alias:"io"
  @io: push(10)
  
  -- Immutable borrowing
  @Stack.new(Integer, Borrowed): alias:"ib"
  @ib: <<io                           -- Borrow immutably
  @Stack.new(Integer, Borrowed): alias:"ib2"
  @ib2: <<io                          -- Multiple immutable borrows allowed
  
  -- At this point, can't mutate through io because active borrows exist
  -- @io: push(io.pop() + 1)          -- Error: can't mutate while borrowed
  
  print(ib.pop(), ib2.pop())          -- Use borrowed values
  
  -- Mutable borrowing
  @Stack.new(Integer, Mutable): alias:"im"
  @im: <:mut io                       -- Mutable borrow
  
  -- Other borrows not allowed during mutable borrow
  -- @ib: <<io                        -- Error: can't immutably borrow during mutable borrow
  
  @im: push(im.pop() + 1)             -- Modify through mutable reference
  
  print(io.peek())                    -- Will print 11 (modification visible)
end
```

### 5.3 Resources and Cleanup

```lua
function handle_resource()
  @Stack.new(Resource, Owned): alias:"res"
  @res: push(open_file("config.txt"))  -- Acquire resource
  
  -- Process with borrowed access
  @Stack.new(Resource, Borrowed): alias:"rb"
  @rb: <<res
  read_config(rb.pop())
  
  -- Modify with mutable access
  @Stack.new(Resource, Mutable): alias:"rm"
  @rm: <:mut res
  write_config(rm.pop())
  
  -- Resource automatically closed when owned stack goes out of scope
}
```

### 5.4 Error Handling With Ownership

```lua
function process_with_errors()
  @Stack.new(Resource, Owned): alias:"ro"
  @ro: push(acquire_resource())
  
  result = {}
  
  -- Try to process
  success, err = pcall(function()
    @Stack.new(Resource, Mutable): alias:"rm"
    @rm: <:mut ro
    process_resource(rm.pop())
    result.Ok = true
  end)
  
  if not success then
    result.Err = err
    -- Resource still owned by ro, will be properly cleaned up
  end
  
  return result
}
```

### 5.5 Stacked Mode Integration

```lua
function temperature_conversion(celsius_str)
  @Stack.new(String, Owned): alias:"s"
  @Stack.new(Float, Owned): alias:"f"
  
  @s: push(celsius_str)
  @f: <:own s                   -- Take ownership and convert to float
  
  -- Calculate with stacked mode
  @f: dup (9/5)*32 sum          -- Direct mathematical notation
  
  return f.pop()
end
```

## 6. Stack-Based Ownership vs. Rust's Borrow Checker

### 6.1 Conceptual Differences

The fundamental distinction between ual's stack-based ownership and Rust's borrow checker is the mental model:

**Rust**: Ownership follows variables and is transferred through assignments and function calls:

```rust
let a = vec![1, 2, 3];    // a owns the vector
let b = a;                // ownership moved to b, a is no longer valid
```

**ual**: Ownership is tied to containers (stacks) and transfers are explicit stack operations:

```lua
@a: push(create_array(1, 2, 3))   -- Value owned by stack a
@b: <:own a                       -- Explicitly transfer from a to b
```

### 6.2 Explicitness

**Rust**: Ownership transfers are often implicit in normal code flow:

```rust
fn process(data: Vec<i32>) {  // Takes ownership of data
    // ...
}

fn main() {
    let numbers = vec![1, 2, 3];
    process(numbers);        // Ownership implicitly transferred
    // numbers is no longer valid here
}
```

**ual**: Ownership transfers are always visually explicit:

```lua
function process()
  @Stack.new(Array, Owned): alias:"data"
  -- ...
end

function main()
  @Stack.new(Array, Owned): alias:"numbers"
  @numbers: push(create_array(1, 2, 3))
  
  @Stack.new(Array, Owned): alias:"process_data"
  @process_data: <:own numbers   -- Explicitly transfer ownership
  process()
  
  -- numbers.pop() would error here
end
```

### 6.3 Error Messages

**Rust**: Borrow checker errors often reference complex lifetimes and variable relationships:

```
error[E0505]: cannot move out of `numbers` because it is borrowed
  --> src/main.rs:8:13
   |
7  |     let reference = &numbers;
   |                     -------- borrow of `numbers` occurs here
8  |     process(numbers);
   |             ^^^^^^^ move out of `numbers` occurs here
9  |     println!("{:?}", reference);
   |                       --------- borrow later used here
```

**ual**: Stack-based ownership errors would reference specific stack operations:

```
Error at line 8: Cannot transfer ownership from 'numbers' to 'process_data'
Reason: Active borrow exists at stack 'num_ref'
```

### 6.4 Learning Curve

**Rust**: Requires understanding concepts like lifetimes, borrowing rules, and ownership semantics that are enforced implicitly.

**ual**: Makes the ownership model more concrete and visible through explicit stack operations. The mental model of "containers with rules" may be more intuitive for many developers.

## 7. Implementation Considerations

### 7.1 Compiler Tracking

The compiler would track several aspects of each stack:

1. **Type**: What type of values the stack accepts
2. **Ownership Mode**: Whether the stack owns values or borrows them
3. **Borrow State**: Active borrows from this stack
4. **Lifetime**: When the stack goes out of scope

### 7.2 Compile-Time Checks

At each stack operation, the compiler verifies:

1. **Type Compatibility**: Value type matches stack type or can be converted
2. **Ownership Rules**: Transfer operation is valid given current ownership
3. **Borrow Validity**: No active borrows that would prevent the operation
4. **Lifetime Constraints**: References don't outlive their source data

### 7.3 Integration with ual Features

The ownership system would integrate with existing ual features:

1. **Stacked Mode**: Ownership transfer notation works in stacked mode
2. **Error Handling**: `.consider` pattern works with ownership errors
3. **Conditional Compilation**: Different ownership strategies for different platforms
4. **Macro System**: Generate ownership-aware code at compile time

### 7.4 Zero Runtime Overhead

Like Rust's borrow checker, all ownership checks would happen at compile time with zero runtime overhead. The generated code would be identical to manually managed code but with guaranteed safety.

## 8. Use Cases and Examples

### 8.1 Hardware Abstraction Layers

```lua
function configure_gpio(pin, mode)
  @Stack.new(HardwareRegister, Owned): alias:"reg"
  @reg: push(get_gpio_register(pin))
  
  @Stack.new(HardwareRegister, Mutable): alias:"mreg"
  @mreg: <:mut reg
  
  if mode == PIN_OUTPUT then
    @mreg: push(mreg.pop() | (1 << pin))
  else
    @mreg: push(mreg.pop() & ~(1 << pin))
  end
  
  -- Register access automatically completed when mreg goes out of scope
end
```

### 8.2 Resource Management

```lua
function process_file(filename)
  @Stack.new(File, Owned): alias:"fo"
  result = {}
  
  -- Try to open file
  success, err = pcall(function()
    @fo: push(open_file(filename))
  end)
  
  if not success then
    result.Err = "Failed to open file: " .. err
    return result
  end
  
  -- Process with borrowed access
  @Stack.new(File, Borrowed): alias:"fb"
  @fb: <<fo
  result.Ok = read_content(fb.pop())
  
  -- File automatically closed when fo goes out of scope
  return result
end
```

### 8.3 Data Processing Pipeline

```lua
function process_sensor_data(raw_data)
  @Stack.new(String, Owned): alias:"raw"
  @Stack.new(Array, Owned): alias:"parsed"
  @Stack.new(Float, Owned): alias:"results"
  
  @raw: push(raw_data)
  @parsed: <:own raw                     -- Take ownership while parsing
  parse_csv(@parsed: peek())
  
  @Stack.new(Array, Borrowed): alias:"analysis"
  @analysis: <<parsed                    -- Borrow for analysis
  
  -- Process each reading
  for i = 0, array_length(analysis.peek()) - 1 do
    @Stack.new(Float, Borrowed): alias:"reading"
    @reading: push(array_get(analysis.peek(), i))
    
    @results: push(process_reading(reading.pop()))
  end
  
  return results.pop()
end
```

## 9. Backward Compatibility

### 9.1 Gradual Adoption

The ownership system would be designed for gradual adoption:

1. **Untyped Stacks**: Continue to work without ownership constraints
2. **Default Ownership**: `Stack.new(Integer)` defaults to `Stack.new(Integer, Owned)`
3. **Mixed Code**: Ownership-aware and regular code can coexist
4. **Safety Zones**: Apply ownership rules to critical sections first

### 9.2 Migration Path

Existing ual 1.4 code could be migrated incrementally:

1. Add ownership annotations to stack declarations
2. Replace direct operations with ownership-aware versions
3. Refactor any code that violates ownership rules
4. Use ownership-aware shorthand notation for new code

## 10. Comparison with Other Approaches

### 10.1 vs. Garbage Collection

**GC Languages** (Python, Java, JavaScript, etc.):
- Runtime overhead for tracking and collecting objects
- Unpredictable pause times
- No compile-time safety guarantees

**ual Stack-Based Ownership**:
- Zero runtime overhead
- Deterministic resource cleanup
- Compile-time safety guarantees

### 10.2 vs. Reference Counting

**Reference Counting** (Swift, Objective-C ARC):
- Runtime overhead for increment/decrement operations
- Potential for reference cycles
- No compile-time safety guarantees for all cases

**ual Stack-Based Ownership**:
- Zero runtime overhead
- No reference cycle problems
- Compile-time safety guarantees

### 10.3 vs. Manual Memory Management

**Manual Management** (C, older C++):
- Full control but error-prone
- Requires discipline and careful coding
- No safety guarantees

**ual Stack-Based Ownership**:
- Maintains control over resource lifetime
- Enforces correctness through compiler
- Compile-time safety guarantees

## 11. Limitations and Future Directions

### 11.1 Current Limitations

1. **No Thread Safety**: The initial proposal doesn't address concurrency
2. **No Higher-Order Ownership**: Cannot express complex sharing patterns
3. **No Ownership Polymorphism**: Functions can't be generic over ownership modes

### 11.2 Future Directions

1. **Thread-Safe Ownership**: Extend model to support concurrent access patterns
2. **Ownership Polymorphism**: Functions that accept different ownership modes
3. **Complex Sharing**: Support for more complex sharing patterns like readers-writer locks
4. **IDE Integration**: Visual tooling for ownership flow

## 12. Conclusion

The proposed stack-based ownership system for ual offers a unique approach to memory safety that builds naturally on ual's container-centric type system. By making ownership an explicit property of stacks rather than an implicit property of variables, the system creates a clear, visible model for reasoning about resource lifetime and access.

This approach provides memory safety guarantees comparable to Rust's borrow checker but with a potentially more intuitive mental model: "values live in containers with rules." For embedded systems developers in particular, this explicit, stack-oriented approach may offer a more natural fit with how they already think about hardware resources.

The stack-based ownership model continues ual's philosophy of progressive disclosure—simple patterns are simple, and complexity only emerges when needed. By making ownership transfers explicit stack operations, the system creates a visual representation of resource flow through the program, potentially reducing the steep learning curve associated with Rust's ownership model.

Most importantly, like Rust's borrow checker, the stack-based ownership system achieves safety guarantees with zero runtime overhead—all checks happen at compile time, resulting in efficient code for even the most resource-constrained environments. This makes it ideal for ual's target domain of embedded systems programming.

We recommend the adoption of this stack-based ownership system for ual 1.5, complementing the typed stacks introduced in ual 1.4 and furthering ual's mission to be a safe, efficient language for embedded systems development.