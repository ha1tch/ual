# ual 1.9 PROPOSAL: Enhanced Execution Mechanics - Pull and Value Handling

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements, and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction: Refining Execution Flow

This proposal introduces targeted refinements to ual's execution mechanics, designed to enhance code clarity and reduce verbosity while maintaining semantic consistency. These enhancements address specific friction points identified through code analysis without compromising ual's philosophical foundations of explicitness and container-centric design.

The key components of this proposal are:

1. **Pull Operation**: A unified operation that combines `peek` and `pop` to streamline the common pattern of accessing a stack element once and immediately discarding it, while preserving the correct separation of these operations when appropriate.

2. **Discard Pattern**: A mechanism inspired by Go's underscore placeholder that enables explicit value discarding without temporary variables.

3. **Consistent Value Binding**: A standardized approach to value handling across different contexts, including multi-return functions, `.consider{}` blocks, and stack operations.

These mechanisms work together to enhance code expressiveness while maintaining ual's commitment to explicit operations. By reducing boilerplate for common patterns and establishing clear rules for value handling, they make ual code more concise and maintainable.

## 2. Background and Motivation

### 2.1 Code Pattern Analysis

An examination of ual code from the standard library and example applications revealed several recurring patterns that can benefit from streamlining:

**Pattern 1: Peek followed by Pop**
```lua
value = stack.peek()  // Access top value
process(value)        // Use the value
stack.pop()           // Discard the value
```

This pattern appears particularly in algorithms that consume stack elements one at a time, such as parsers, evaluators, and transformations. The separation of peeking and popping creates both verbosity and the potential to forget the `pop` step.

**Pattern 2: Pop with Immediate Discard**
```lua
temp = stack.pop()  // Pop value into temporary variable
// temp is never used
```

This pattern occurs in situations where elements need to be removed from a stack but aren't needed for computation, creating unnecessary temporary variables. It is important to note that ual already has a `drop` operation for this purpose in stacked mode, but its usage in imperative mode is not always obvious to developers.

**Pattern 3: Inconsistent Value Binding**
```lua
// In consider blocks
result.consider {
  if_ok {
    process(_1)  // Using implied placeholder
  }
}

// In multi-return functions
x, y = multi_value_function()  // Using named variables
```

The inconsistent approaches to binding values across different contexts creates unnecessary cognitive load when switching between patterns.

### 2.2 Integration with ual's Design Philosophy

These proposed features build upon ual's existing design principles:

1. **Explicit Operations**: All operations remain explicit, with clear indication of what happens to each value.

2. **Container-Centric**: The enhancements maintain focus on operations applied to containers rather than individual values.

3. **Progressive Discovery**: Simple operations remain simple, while more sophisticated patterns become available when needed.

4. **Performance Consciousness**: The features are designed for efficient implementation with minimal runtime overhead.

## 3. Current Operation Semantics

Before introducing new operations, it's essential to clarify the current behavior of stack operations in both imperative and stacked modes.

### 3.1 Current Operation Semantics

| Operation | Imperative Mode Behavior | Stacked Mode Behavior | Value Access in Stacked Mode |
|-----------|--------------------------|------------------------|------------------------------|
| `push(value)` | Adds value to stack, returns nothing | Adds value to stack | N/A |
| `pop()` | Removes top value and returns it | Removes top value from stack | Value placed in `*_` placeholder |
| `peek()` | Returns top value without removing it | Returns top value without removing it | Value placed in `*` placeholder |
| `drop` | Not typically used in imperative mode | Removes top value without returning it (Forth behavior) | Value is discarded |
| `dup` | Not typically used in imperative mode | Duplicates top value | N/A |
| `swap` | Not typically used in imperative mode | Swaps top two values | N/A |

### 3.2 Value Flow in Stack Operations

In stacked mode, values flow through special placeholders:
- `*` contains the value from the most recent `peek()` operation
- `*_` contains the value from the most recent `pop()` operation

These placeholders are implicit and accessible only within the immediate context of the operation.

## 4. Proposed Feature: Pull Operation

### 4.1 Core Definition and Semantics

The `pull` operation combines `peek` and `pop` into a single atomic operation:

```lua
value = stack.pull()  // Equivalent to: value = stack.peek(); stack.pop();
```

This operation returns the top value from a stack and removes it in a single step. Unlike separate `peek` and `pop` operations, `pull` guarantees that the value is both accessed and removed atomically.

Key characteristics:
- Returns the stack element (like `peek`)
- Removes the element from the stack (like `pop`)
- Operates in a single atomic step
- Maintains consistent error handling with other stack operations

### 4.2 Operation Semantics Table

| Operation | Imperative Mode Behavior | Stacked Mode Behavior | Value Access in Stacked Mode |
|-----------|--------------------------|------------------------|------------------------------|
| `push(value)` | Adds value to stack, returns nothing | Adds value to stack | N/A |
| `pop()` | Removes top value and returns it | Removes top value from stack | Value placed in `*_` placeholder |
| `peek()` | Returns top value without removing it | Returns top value without removing it | Value placed in `*` placeholder |
| `pull()` | Removes top value and returns it (atomic) | Removes top value from stack (atomic) | Value placed in `_` placeholder |
| `drop` | Not typically used in imperative mode | Removes top value without returning it (Forth behavior) | Value is discarded |

### 4.3 Syntax and Integration

The `pull` operation follows ual's established syntax patterns for stack operations:

```lua
// Method syntax
value = stack.pull()     // Pull from top of stack
value = stack.pull(n)    // Pull from position n

// Stack selector syntax
@stack: pull()           // Pull from top of stack, value in _ placeholder
@stack: pull(n)          // Pull from position n, value in _ placeholder
```

The parentheses are required in all forms to maintain consistent method call syntax, distinguishing operations from values or properties.

### 4.4 Valid versus Invalid peek()+pop() Sequences

It's crucial to distinguish between legitimate uses of `peek()+pop()` and cases where `pull()` would be more appropriate:

#### 4.4.1 Valid peek()+pop() Sequences (Should NOT be replaced with pull)

1. **Different Index Access**:
   ```lua
   value = stack.peek(3)  // Access element at position 3
   stack.pop()  // Remove top element (not position 3)
   ```
   This pattern is valid because the peek and pop operate on different stack positions.

2. **Conditional Processing**:
   ```lua
   value = stack.peek()  // Look without removing
   if valid_condition(value) then
     process(value)
     stack.pop()  // Only remove if condition is met
   end
   ```
   This pattern is valid because removal is conditional.

3. **Borrowed Segment Access**:
   ```lua
   scope {
     @segment: borrow([1..3]@stack)
     value = segment.peek()  // Can peek but not pop from borrowed segment
   }
   @stack: pop()  // Remove after borrowing scope ends
   ```
   This pattern is valid because borrowing constrains the available operations.

4. **Multi-Stack Operations**:
   ```lua
   value1 = stack1.peek()
   value2 = stack2.peek()
   if value1 == value2 then
     stack1.pop()
     stack2.pop()
   end
   ```
   This pattern is valid because it coordinates across multiple stacks.

#### 4.4.2 Invalid peek()+pop() Sequences (Should be replaced with pull)

1. **Immediate Consumption**:
   ```lua
   // Before
   value = stack.peek()  // Get value
   process(value)        // Use it once
   stack.pop()           // Then immediately discard
   
   // After
   value = stack.pull()  // Get value and remove in one step
   process(value)        // Use it once
   ```

2. **Function Parameters**:
   ```lua
   // Before
   value = stack.peek()
   result = compute(value)
   stack.pop()
   
   // After
   result = compute(stack.pull())
   ```

3. **Return Values**:
   ```lua
   // Before
   value = stack.peek()
   stack.pop()
   return value
   
   // After
   return stack.pull()
   ```

4. **Sequential Processing**:
   ```lua
   // Before
   while stack.depth() > 0 do
     value = stack.peek()
     process(value)
     stack.pop()
   end
   
   // After
   while stack.depth() > 0 do
     value = stack.pull()
     process(value)
   end
   ```

### 4.5 Type Safety and Error Handling

Like other stack operations, `pull` maintains ual's type safety guarantees:

```lua
@Stack.new(Integer): alias:"i"
@Stack.new(String): alias:"s"

value = i.pull()    // Returns an Integer
text = s.pull()     // Returns a String
```

For empty stacks, `pull` follows the same error semantics as `pop`:

```lua
// Handling empty stack errors - standard approach
function process_stack(stack)
  if stack.depth() > 0 then
    value = stack.pull()
    return process(value)
  else
    return default_value
  end
end

// Handling empty stack errors - with error stack
@error > function safe_pull(stack)
  if stack.depth() == 0 then
    @error > push("Cannot pull from empty stack")
    return nil
  end
  return stack.pull()
end

// Handling empty stack errors - with consider pattern
safe_pull(stack).consider {
  if_ok(value) {
    process(value)
  }
  if_err(error) {
    handle_error(error)
  }
}

// Resource-constrained approach - defensive programming
function process_with_minimum_overhead(stack)
  // Pre-check avoids try-catch overhead in critical sections
  if stack.depth() > 0 then
    value = stack.pull()
    // Process with guaranteed valid value
  end
end
```

#### 4.5.1 Edge Case: Empty Stack Behavior

When attempting to `pull()` from an empty stack, the operation will throw the same error as `pop()` would:

```lua
// Both of these throw the same error: "Cannot pull/pop from empty stack"
value = empty_stack.pop()   // Error
value = empty_stack.pull()  // Error - identical behavior
```

This consistent behavior ensures that error handling patterns developed for `pop()` work identically for `pull()`.

The `pull()` operation must produce identical error messages to `pop()` for the same error conditions. This guarantees consistency in error handling across both operations.

#### 4.5.2 Edge Case: Index Out of Bounds

When using `pull(n)` with an index that exceeds the stack bounds, the operation throws an "Index out of bounds" error, identical to the behavior of `peek(n)` with an invalid index:

```lua
stack = Stack.new(Integer)
stack.push(1)
value = stack.pull(5)  // Error: "Index out of bounds"
```

### 4.6 Integration with Stack Perspectives

The `pull` operation adapts to different stack perspectives:

```lua
@stack: lifo       // Last-In-First-Out perspective (default)
value = stack.pull() // Pulls from the top

@stack: fifo       // First-In-First-Out perspective
value = stack.pull() // Pulls from the front

@stack: hashed     // Hash perspective
value = stack.pull("key") // Pulls the value associated with "key"
```

This maintains consistent semantics across ual's perspective system.

#### 4.6.1 Edge Case: Priority-Based Perspectives

For MAXFO (priority queue) and MINFO (minimum priority queue) perspectives, `pull()` follows the same priority semantics as `pop()`:

```lua
@stack: maxfo       // Maximum-First-Out perspective
value = stack.pull() // Pulls the highest priority element

@stack: minfo       // Minimum-First-Out perspective
value = stack.pull() // Pulls the lowest priority element
```

The `pull()` operation must maintain identical element selection behavior as `pop()` for all perspective types. This ensures consistency in the language semantics.

#### 4.6.2 Edge Case: Hashed Perspective Without Key

When using `pull()` in the hashed perspective without a key parameter, an error is raised:

```lua
@stack: hashed
value = stack.pull()  // Error: "Key parameter required for pull() in Hashed perspective"
```

This error message clearly indicates that a key is required when using the `pull()` operation with a hashed perspective.

### 4.7 Stacked Mode Usage

In stacked mode, the `pull` operation functions similarly to `pop` but places the value in the `_` placeholder for immediate access:

```lua
// Stacked mode usage
@stack: push(42)
@stack: pull()     // Removes 42 from stack, places it in _
@stack: push(_ * 2)  // Uses the pulled value (42) and pushes result (84)

// Equivalent imperative code
stack.push(42)
value = stack.pull()
stack.push(value * 2)
```

#### 4.7.1 Edge Case: Multiple Pull Operations

When multiple `pull` operations are performed in sequence, each operation overwrites the `_` placeholder with the most recently pulled value:

```lua
@stack: push(1) push(2) push(3)
@stack: pull()  // _ contains 3
@stack: pull()  // _ now contains 2, previous value is no longer accessible
@stack: push(_ + 10)  // Pushes 12 (2 + 10)
```

This behavior is consistent with how `*` and `*_` placeholders work for `peek` and `pop` operations, where only the most recent value is retained.

#### 4.7.2 Edge Case: Placeholder Lifetime

The `_` placeholder remains valid only within the current statement or until the next stack operation, whichever comes first:

```lua
@stack: pull()  // _ contains the pulled value
@stack: push(_ * 2)  // Valid: _ is used in the same statement

@stack: pull()  // _ contains a new pulled value
do_something_else()
@stack: push(_)  // Valid: _ still refers to the pulled value as no intervening stack operation has occurred

@stack: pull()  // _ contains a new pulled value
@stack: peek()  // This updates the * placeholder
@stack: push(_)  // Still valid: _ is unchanged by peek operation

@stack: pull()  // _ contains a new pulled value
@stack: pop()   // This updates the *_ placeholder
@stack: push(_)  // Still valid: _ is unchanged by pop operation
```

A statement boundary is precisely defined as code terminated by ';' or line break, or within a block between '{' and '}'. This clear definition ensures developers understand exactly when placeholders remain valid.

#### 4.7.3 Edge Case: Placeholder Conflicts

When multiple placeholder systems are active simultaneously, the following precedence rules apply:

1. Each placeholder (`_`, `*`, and `*_`) is independent and doesn't conflict with others
2. When multiple operations of the same type are performed in a single statement, placeholders refer only to the most recent operation of their type

```lua
@stack: pull() peek() pop()  // Sets _, *, and *_
@stack: push(_ + * + *_)     // Uses all three placeholders from their respective operations
```

While this is allowed, mixing multiple placeholder types in a single expression is not recommended for code clarity. The compiler will issue a warning: "Warning: Multiple placeholders in single expression may reduce readability".

### 4.8 Expanded Stacked Mode Examples

#### 4.8.1 Processing a Stack Using Pull in Stacked Mode

```lua
// Example: Processing a stack using pull in stacked mode
@data: push:1 push:2 push:3

// Using pull in a loop
while_true(@data: depth() > 0)
  @data: pull()        // Places value in _
  @results: push(_ * 2)  // Use the value via _ placeholder
end_while_true

// Contrast with traditional approach
while_true(@data: depth() > 0)
  @data: peek()        // Places value in *
  @results: push(* * 2)  // Use the value via * placeholder
  @data: pop()         // Remove the value (now in *_)
end_while_true
```

#### 4.8.2 Conditional Processing with Pull

```lua
// Example: Conditional processing with pull
@stack: pull()          // Pull value, available in _
if_true(_ > 10)
  @results: push(_ * 2)   // Use placeholder
end_if_true

// Contrast with traditional peek/pop for conditional processing
// Note: This is a case where separate peek/pop is preferred for conditional removal
@stack: peek()          // Peek value, available in *
if_true(* > 10)
  @results: push(* * 2)   // Use placeholder
  @stack: pop()         // Only remove if condition is met
end_if_true
```

#### 4.8.3 Combining Pull with Other Stack Operations

```lua
// Example: Using pull with other stack operations
@values: push:10 push:20 push:30
@values: {
  pull()              // Pull 30, available in _
  swap                // Swap remaining values (10, 20 -> 20, 10)
  push(_ + peek())    // Add pulled value to current top (30 + 20 = 50)
  pull()              // Pull the sum (50), now in _
  swap                // Swap with the last value
  push(_ * peek())    // Multiply pulled value by current top (50 * 10 = 500)
}
// Result: stack now contains [500]
```

#### 4.8.4 Comparing Placeholders in Complex Processing

```lua
// Example: Comparing placeholder usage
@stack: push:10 push:20

// Using peek and pop with their placeholders
@stack: peek()           // 20 is now in *
@stack: push(* + 5)      // Push 25 (20 + 5)
@stack: pop()            // Remove 20, now in *_
@stack: push(*_ * 2)     // Push 40 (20 * 2)

// Equivalent with pull
@stack: pull()           // 20 is now in _
@stack: push(_ + 5)      // Push 25 (20 + 5)
@stack: push(_ * 2)      // Push 40 (20 * 2)

// Resulting stack has [10, 25, 40] in both cases,
// but pull version is more concise
```

It's recommended to avoid mixing different placeholder types in a single expression. The compiler will issue a lint warning if mixed placeholder usage is detected: "Mixing placeholder types (_ and *) in expression may cause confusion".

### 4.9 Pull and Borrowed Segments

For borrowed stack segments, `pull()` is disallowed to maintain consistency with ual's borrowing semantics:

```lua
@data: push:1 push:2 push:3 push:4 push:5

scope {
  // Borrow elements 1-3 (second through fourth elements)
  @window: borrow([1..3]@data)
  
  value = window.peek()    // Valid: Can peek from borrowed segment
  // value = window.pull() // Error: Cannot pull/pop from borrowed segment
}
```

This restriction ensures that borrowed segments remain immutable through the borrowing scope, consistent with how `pop()` is currently disallowed on borrowed segments.

### 4.10 Pull Operation with Crosstacks

When using `pull()` on a crosstack level, the operation succeeds only if all constituent stacks have elements at that level. Otherwise, it raises an 'Incomplete level' error:

```lua
@matrix: Stack.new(Stack)
@row1: push:1 push:2
@row2: push:3        // Only one element
@matrix: push(row1) push(row2)

@matrix~0: pull()    // Works because level 0 exists in all stacks
// @matrix~1: pull() // Error: "Incomplete level" because row2 has no element at level 1
```

For partial level operations, use explicit loops with conditional checks instead of crosstack pull:

```lua
// Safer approach for uneven levels
for i = 0, matrix.depth() - 1 do
  if matrix.peek(i).depth() > level then
    value = matrix.peek(i).pull(level)
    process(value)
  end
end
```

## 5. Common Anti-Patterns to Avoid

The following examples illustrate common anti-patterns that developers should avoid when using the new features:

### 5.1 Using `pull` when conditional removal is needed

```lua
// Incorrect: Unconditional pull
value = stack.pull()
if condition(value) then
  process(value)
  // Cannot skip removal if condition fails
end

// Correct: Conditional removal
value = stack.peek()
if condition(value) then
  process(value)
  stack.pop()  // Only remove if condition passes
end
```

### 5.2 Accessing pulled values across different stacks

```lua
// Incorrect: Confusing placeholder usage
@stack1: pull()  // Sets _
@stack2: push(_ * 2)  // May be confusing about which _ this is

// Correct: Explicit binding for clarity
value = stack1.pull()
stack2.push(value * 2)
```

### 5.3 Mixing placeholders inconsistently

```lua
// Confusing: Mixed placeholder styles
@stack: peek() pull()  // Sets both * and _
@stack: push(* + _)    // Unclear intention and confusing to read

// Better: Consistent placeholder usage
@stack: pull() pull()  // Both values in _ (last one only)
// or
value1 = stack.pull()
value2 = stack.pull()
stack.push(value1 + value2)
```

## 6. Placeholder Usage Best Practices

### 6.1 Prefer Named Variables for Complex Logic

When operations span multiple lines or involve complex logic, prefer named variables over placeholders:

```lua
// Instead of
@stack: pull()
@stack: push(_ * _ + 42 / _)  // Confusing repeated placeholder

// Prefer
value = stack.pull()
stack.push(value * value + 42 / value)
```

### 6.2 Use Placeholders for Simple, Immediate Operations

Placeholders work best for simple operations immediately following the stack operation:

```lua
@stack: pull()
@stack: push(_ + 1)  // Simple, immediate use is clear
```

### 6.3 Maintain Consistency Within Code Blocks

Be consistent in how you access values within a block of code:

```lua
// Consistent placeholder usage
@stack: {
  pull()          // Use pull consistently
  push(_ * 2)
  pull()
  push(_ + 10)
}

// Consistent variable usage
value1 = stack.pull()
stack.push(value1 * 2)
value2 = stack.pull()
stack.push(value2 + 10)
```

### 6.4 Document Placeholder Meaning in Complex Code

When using placeholders in more complex scenarios, add comments to clarify:

```lua
@stack: pull()  // Pull price value
@tax: push(_ * 0.2)  // Calculate tax based on price
```

## 7. Proposed Feature: Discard Pattern

### 7.1 Core Definition and Semantics

The discard pattern enables explicit indication that a value should be discarded. It uses the underscore character `_` as a special identifier in certain contexts:

```lua
// In assignment contexts
_, value = multi_return_function()  // First return value is discarded

// In parameter lists
function process(_, b)  // First parameter is explicitly ignored
  return b * 2
end
```

Unlike simply omitting the variable, the discard pattern makes the intent explicit, improving code clarity and preventing accidental omissions.

### 7.2 Relationship to Existing `drop` Operation

It's important to distinguish the discard pattern from the existing `drop` operation:

- **`drop`**: A stack operation in stacked mode that removes the top value without returning it (following Forth semantics)
- **`_` discard**: A pattern for explicitly ignoring values in assignment and parameter contexts

These serve complementary purposes:

```lua
// Stacked mode - use drop to remove unwanted stack value
@stack: drop  // Removes top value, discarding it

// Assignment context - use _ to explicitly discard return value
_, result = function_with_multiple_returns()
```

### 7.3 Syntax Specification

The discard pattern has specific syntax rules to avoid ambiguity:

1. **Single-character Identifier**: Only the single underscore character `_` is recognized as a discard marker.
2. **Limited Contexts**: The discard is only recognized in:
   - Assignment left-hand side
   - Function parameter declarations
   - Function calls with multiple return values

The discard pattern does not introduce a general variable or value handle, distinguishing it from named variables or placeholders.

### 7.4 Semantic Guarantees

The discard pattern provides important guarantees:

1. **Explicit Intent**: The presence of `_` clearly communicates that the value is intentionally discarded.
2. **Eager Evaluation**: Operations that produce discarded values are still evaluated.
3. **No Binding**: The `_` does not bind to any scope or create a usable variable.
4. **Compile-Time Verification**: The compiler verifies that discarded values are not accidentally used.

### 7.5 Identifier Restrictions

To avoid confusion with the discard pattern, `_` is reserved and cannot be used as a user-defined variable name:

```lua
local _ = 42  // Error: Underscore is reserved and cannot be used as a variable name
```

This restriction prevents potential confusion between the discard pattern and regular variables.

### 7.6 Nested Contexts and Multiple Discards

The discard pattern is valid in all contexts where a variable binding would be valid, including nested destructuring:

```lua
// Multiple discards in simple context
_, _, result = complex_function()  // Discard first and second return values

// Nested destructuring with discards
a, _, {c, _} = complex_function()  // Discard second return value and nested field
```

There is no limit to the number of discard markers (`_`) that can appear in a binding context. Each one indicates an independent value to be discarded.

## 8. Proposed Feature: Consistent Value Binding

### 8.1 Core Definition and Semantics

This proposal formalizes a consistent approach to value binding across different contexts in ual:

1. **Named Variable Binding**: The standard approach using identifiers:
   ```lua
   value = stack.pull()  // Bind to named variable
   ```

2. **Pattern Binding**: Structured binding for multi-value contexts:
   ```lua
   quotient, remainder = divide(10, 3)  // Bind to multiple variables
   ```

3. **Placeholder Binding**: Temporary bindings in specific contexts:
   ```lua
   result.consider {
     if_ok(value) {  // Bind 'value' as placeholder
       process(value)
     }
   }
   ```

The key innovation is establishing clear, consistent rules for how values are bound and accessed across these different contexts.

### 8.2 Consider Block Binding

For `.consider{}` blocks, this proposal formalizes the binding mechanism:

```lua
// Explicit binding parameter
result.consider {
  if_ok(value) {     // Explicitly bind to 'value'
    process(value)
  }
  if_err(error) {    // Explicitly bind to 'error'
    handle_error(error)
  }
}
```

This makes the binding explicit rather than relying on implied placeholders like `_1`. If no explicit binding is provided, the legacy numbered placeholder is still supported for backward compatibility:

```lua
// Legacy binding with implicit placeholder
result.consider {
  if_ok {
    process(_1)      // Use first value via placeholder
  }
}
```

Binding parameter names must not conflict with reserved placeholders (`_`, `*`, `*_`). Attempting to use a reserved placeholder as a binding parameter will result in a compiler error:

```lua
result.consider {
  if_ok(_) {  // Error: Cannot use reserved placeholder '_' as binding parameter
    process(_)
  }
}
```

### 8.3 Scope and Visibility Rules

Value bindings follow clear scope and visibility rules:

1. **Block Scope**: Bindings are only valid within the block where they are defined.
2. **No Shadowing**: Explicit bindings cannot shadow variables from outer scopes.
3. **Read-Only**: Bindings cannot be reassigned within their scope.
4. **Independent Contexts**: Each binding context is independent (consider branches, function bodies).

Binding parameters temporarily shadow variables of the same name from outer scopes, but cannot be reassigned:

```lua
value = "outer"
result.consider {
  if_ok(value) {  // This 'value' shadows the outer 'value'
    process(value)  // Uses the binding value
    // value = 42  // Error: Cannot reassign binding parameter
  }
}
// Here, 'value' refers to the original "outer" value
```

These rules ensure that value handling remains predictable and safe across different contexts.

### 8.4 Multiple Pattern Matches

For consider blocks with multiple pattern clauses, binding parameters are only valid within their respective clause:

```lua
result.consider {
  if_ok(value) {
    // 'value' is valid here
    process(value)
  }
  if_err(error) {
    // 'error' is valid here, but 'value' is not accessible
    handle_error(error)
  }
  if_else {
    // Neither 'value' nor 'error' are accessible here
    handle_default_case()
  }
}
```

This ensures that developers can't accidentally access values from non-matching patterns, which would be logically inconsistent.

### 8.5 Placeholder Summary

The proposal standardizes the following placeholders:

| Placeholder | Source | Scope | Description |
|-------------|--------|-------|-------------|
| `*` | `peek()` operation in stacked mode | Operation | Contains the peeked value |
| `*_` | `pop()` operation in stacked mode | Operation | Contains the popped value |
| `_` | `pull()` operation in stacked mode or discard pattern | Operation or Pattern | Contains the pulled value or marks discard |
| `_1`, `_2`, etc. | Consider block legacy placeholders | Block | Contains values from pattern match (legacy) |
| Named parameters | Consider block binding parameters | Block | Contains values from pattern match (preferred) |

## 9. Code Examples and Comparison

### 9.1 Current vs. Enhanced Approach

#### Example 1: Value Processing

**Current Approach**:
```lua
function process_stack(stack)
  while stack.depth() > 0 do
    value = stack.peek()
    if is_valid(value) then
      result = transform(value)
      stack.pop()  // Remove after using
      @results: push(result)
    else
      stack.pop()  // Remove invalid value
    end
  end
end
```

**Enhanced Approach**:
```lua
function process_stack(stack)
  while stack.depth() > 0 do
    value = stack.pull()  // Get and remove in one step
    if is_valid(value) then
      result = transform(value)
      @results: push(result)
    end
    // Invalid value already removed by pull
  end
end
```

#### Example 2: Stacked Mode Processing

**Current Approach**:
```lua
@stack: {
  peek()      // Peek value, available in *
  dup         // Duplicate the top value
  push(* * 2) // Multiply by 2
  pop()       // Remove the original value, available in *_
}
```

**Enhanced Approach**:
```lua
@stack: {
  pull()      // Pull value, available in _
  push(_ * 2) // Multiply by 2
}
```

#### Example 3: Multi-Value Processing

**Current Approach**:
```lua
function process_coordinates(points)
  for i = 1, #points do
    x, y, z = get_coordinates(points[i])
    // Only using x and z
    process_point(x, z)
  end
end
```

**Enhanced Approach**:
```lua
function process_coordinates(points)
  for i = 1, #points do
    x, _, z = get_coordinates(points[i])  // Explicitly discard y
    process_point(x, z)
  end
end
```

#### Example 4: Result Handling

**Current Approach**:
```lua
function process_result(result)
  result.consider {
    if_ok {
      // Unclear where _1 comes from
      process(_1)
    }
    if_err {
      // Unclear where _1 comes from
      handle_error(_1)
    }
  }
end
```

**Enhanced Approach**:
```lua
function process_result(result)
  result.consider {
    if_ok(value) {  // Explicit binding
      process(value)
    }
    if_err(error) {  // Explicit binding
      handle_error(error)
    }
  }
end
```

### 9.2 Integrated Example: Data Processing Pipeline

The enhanced execution mechanics enable more concise, readable data processing pipelines:

```lua
function process_data_stream(input)
  @Stack.new(Data): alias:"processed"
  
  while input.has_next() do
    chunk = input.pull_chunk()
    
    // Parse the chunk, explicitly discarding metadata
    data, _, timestamp = parse_chunk(chunk)
    
    // Process only if valid timestamp
    timestamp.consider {
      if_match(function(t) return t > threshold end) (t) {
        result = transform(data)
        @processed: push(result)
      }
      if_else {
        // Log but continue processing
        log_invalid_timestamp(timestamp)
      }
    }
  end
  
  return processed
end
```

## 10. Migration Strategy and Legacy Support

### 10.1 Backward Compatibility

These features have been designed with backward compatibility in mind:

1. **Pull**: Doesn't affect existing code; provides an alternative to peek/pop sequences
2. **Discard**: Only applies in specific contexts and uses a reserved character; unlikely to conflict with existing code
3. **Binding**: Extends existing patterns without breaking them; legacy placeholder usage still works

### 10.2 Migration Approaches

**Pattern Migration Examples**:

1. **Peek/Pop Sequences**:
   ```lua
   // Before
   value = stack.peek()
   process(value)
   stack.pop()
   
   // After
   value = stack.pull()
   process(value)
   ```

2. **Stacked Mode Processing**:
   ```lua
   // Before
   @stack: peek() // Value in *
   @stack: push(* * 2)
   @stack: pop()
   
   // After
   @stack: pull() // Value in _
   @stack: push(_ * 2)
   ```

3. **Unused Pop Values**:
   ```lua
   // Before
   temp = stack.pop() // Unused value
   
   // After (imperative mode)
   stack.pop() // Just pop without assignment
   
   // After (stacked mode)
   @stack: drop // Use drop operation
   ```

4. **Unclear Consider Blocks**:
   ```lua
   // Before
   result.consider {
     if_ok {
       process(_1)  // Unclear what _1 represents
     }
   }
   
   // After
   result.consider {
     if_ok(value) {  // Clear what value represents
       process(value)
     }
   }
   ```

### 10.3 Legacy Placeholder Support

To support existing code, numbered placeholders (`_1`, `_2`, etc.) will continue to work in `.consider{}` blocks. Documentation will mark these as legacy features, encouraging migration to explicit bindings in new code.

The compiler will provide an optional warning when legacy placeholders are used, suggesting the explicit binding alternative:

```
Warning at line 42: Legacy placeholder '_1' used in consider block.
  Consider using explicit binding parameter: 'if_ok(value) { ... }'
```

This warning can be disabled for existing projects during migration.

## 11. Implementation Considerations

### 11.1 Pull Implementation

The `pull` operation can be efficiently implemented in TinyGo:

```go
// Implementation in TinyGo
func (s *Stack) Pull() interface{} {
    if s.depth == 0 {
        panic("Cannot pull from empty stack")
    }
    
    value := s.elements[s.depth-1]
    s.depth--
    s.elements[s.depth] = nil  // Aid garbage collection
    
    return value
}

// Implementation for indexed pull
func (s *Stack) PullAt(index int) interface{} {
    if index < 0 || index >= s.depth {
        panic("Index out of bounds")
    }
    
    value := s.elements[s.depth-1-index]
    
    // Remove the element (shift elements above it down)
    copy(s.elements[s.depth-1-index:], s.elements[s.depth-index:])
    s.depth--
    s.elements[s.depth] = nil  // Aid garbage collection
    
    return value
}
```

The implementation is straightforward and requires minimal changes to the existing stack implementation. Performance comparison with separate `peek`/`pop` operations would need benchmarking to verify any efficiency claims, but at minimum it reduces the API surface needed for common operations.

For perspective-specific implementations, the existing perspective mechanisms can be leveraged:

```go
// Pull implementation respecting the active perspective
func (s *Stack) Pull() interface{} {
    if s.depth == 0 {
        panic("Cannot pull from empty stack")
    }
    
    var value interface{}
    
    switch s.perspective {
    case LIFO:
        value = s.elements[s.depth-1]
        s.depth--
        s.elements[s.depth] = nil
    case FIFO:
        value = s.elements[0]
        copy(s.elements[0:], s.elements[1:s.depth])
        s.depth--
        s.elements[s.depth] = nil
    case MAXFO:
        // Find max element
        maxIdx := 0
        for i := 1; i < s.depth; i++ {
            if s.compareFunc(s.elements[i], s.elements[maxIdx]) > 0 {
                maxIdx = i
            }
        }
        value = s.elements[maxIdx]
        // Remove element at maxIdx
        copy(s.elements[maxIdx:], s.elements[maxIdx+1:s.depth])
        s.depth--
        s.elements[s.depth] = nil
    case MINFO:
        // Find min element
        minIdx := 0
        for i := 1; i < s.depth; i++ {
            if s.compareFunc(s.elements[i], s.elements[minIdx]) < 0 {
                minIdx = i
            }
        }
        value = s.elements[minIdx]
        // Remove element at minIdx
        copy(s.elements[minIdx:], s.elements[minIdx+1:s.depth])
        s.depth--
        s.elements[s.depth] = nil
    case HASHED:
        // Requires key parameter, handled separately
        panic("Key parameter required for pull() in Hashed perspective")
    }
    
    return value
}
```

The `pull()` operation must reuse the same element selection logic as `pop()` for all perspective types. This ensures that the behavior is consistent and predictable across all perspectives.

### 11.2 Discard Implementation

The discard pattern requires parser and compiler changes:

1. **Parser Changes**: Recognize `_` as a special token in assignment and parameter contexts
2. **Compiler Handling**: Generate appropriate code that evaluates expressions but doesn't create bindings for discarded values
3. **Static Analysis**: Verify discarded values aren't referenced later

The implementation complexity is moderate, as similar mechanics exist in the TinyGo toolchain from which ual derives its backend.

```go
// Pseudocode for compiler handling of discard pattern
func compileAssignment(lhs []Expr, rhs Expr) {
    values := compileExpr(rhs) // Compile the right-hand side expression
    
    for i, target := range lhs {
        if isDiscardPattern(target) {
            // Skip binding for discards, but ensure value is evaluated
            if i < len(values) {
                // Generate code to evaluate but discard the value
                generateDiscardCode(values[i])
            }
        } else {
            // Normal binding
            if i < len(values) {
                generateAssignmentCode(target, values[i])
            } else {
                generateAssignmentCode(target, nil) // Assign nil for missing values
            }
        }
    }
}
```

### 11.3 Consistent Binding Implementation

Implementing consistent value binding requires more substantial compiler changes:

1. **Parser Extensions**: Add support for binding parameters in `consider` blocks
2. **Scope Management**: Create and manage binding scopes with appropriate visibility rules
3. **Static Analysis**: Ensure bindings follow the read-only and no-shadowing rules
4. **Code Generation**: Generate appropriate access code for different binding contexts

This represents the most complex part of the proposal in terms of implementation effort.

```go
// Pseudocode for consider block with binding parameters
func compileConsiderBlock(expr Expr, patterns []Pattern) {
    // Compile the expression to be considered
    exprCode := compileExpr(expr)
    
    // Generate dispatch based on patterns
    for _, pattern := range patterns {
        // Compile pattern condition
        conditionCode := compilePatternCondition(pattern)
        
        // If pattern has binding parameters, create scope for them
        if hasBindingParams(pattern) {
            generateScopeStart()
            
            // Generate binding code for each parameter
            for _, param := range pattern.bindingParams {
                generateBindingCode(param, getMatchValue(pattern))
            }
            
            // Compile pattern body with bindings in scope
            compileBlock(pattern.body)
            
            generateScopeEnd()
        } else {
            // Legacy mode - compile without explicit bindings
            compileBlock(pattern.body)
        }
    }
}
```

### 11.4 Implementation Phases

Given the varying complexity of these features, a phased implementation approach is recommended:

1. **Phase 1**: Document existing behavior completely with comprehensive tables showing operations in both modes
2. **Phase 2**: Implement the `pull` operation in imperative mode with clear guidelines for appropriate usage
3. **Phase 3**: Extend `pull` to stacked mode with proper placeholder semantics
4. **Phase 4**: Implement the discard pattern
5. **Phase 5**: Implement consistent value binding
6. **Phase 6**: Add tooling to identify appropriate migration opportunities

This phased approach allows incremental improvement while managing implementation risk.

### 11.5 Performance Considerations

The `pull` operation is guaranteed to be at least as efficient as separate `peek`+`pop` operations, since it eliminates one function call and potential redundant bound checks. While no hard performance guarantee is made, implementations should strive to optimize the operation for common cases:

```go
// Optimized pull implementation for LIFO perspective (most common case)
func (s *Stack) Pull() interface{} {
    if s.depth == 0 {
        panic("Cannot pull from empty stack")
    }
    
    // Fast path for LIFO perspective
    if s.perspective == LIFO {
        value := s.elements[s.depth-1]
        s.depth--
        s.elements[s.depth] = nil
        return value
    }
    
    // Slower path for other perspectives
    // ...implementation for other perspectives...
}
```

In some implementations, the compiler may optimize sequences of `peek()`+`pop()` into `pull()` automatically when it can determine that the pattern matches the canonical use case for `pull`.

## 12. Integration with ual's Feature Set

### 12.1 Integration with Stack Perspectives

The proposed features integrate seamlessly with ual's perspective system:

```lua
@stack: fifo  // FIFO perspective
value = stack.pull()  // Pulls from front (oldest element)

@stack: lifo  // LIFO perspective
value = stack.pull()  // Pulls from top (newest element)

@stack: hashed  // Hashed perspective
value = stack.pull("key")  // Pulls value associated with key
```

This consistency across perspectives is especially valuable for code that needs to work with different stack configurations.

### 12.2 Integration with Ownership System

The features respect and integrate with ual's ownership semantics:

```lua
@Stack.new(Resource, Owned): alias:"resources"
resource = resources.pull()  // Transfers ownership of resource

@Stack.new(Data, Borrowed): alias:"view"
data = view.peek()  // Borrows without ownership transfer
_, _ = process(data)  // Explicitly discard multiple return values
```

For borrowed stack segments, `pull()` is disallowed just as `pop()` is, maintaining consistent ownership semantics.

### 12.3 Integration with Error Handling

The features work with ual's error handling mechanisms:

```lua
@error > function safely_pull(stack)
  if stack.depth() == 0 then
    @error > push("Empty stack")
    return nil
  end
  return stack.pull()
end

result = safely_pull(stack)
result.consider {
  if_ok(value) {
    // Use value
  }
  if_err(err) {
    // Handle error
  }
}
```

The consistent binding mechanism for consider blocks enhances error handling by making the flow of values more explicit and readable.

### 12.4 Integration with Crosstacks

The `pull` operation integrates naturally with ual's crosstack feature:

```lua
@matrix~0: pull()  // Pull from level 0 across all stacks
@matrix~1: push(_ * 2)  // Use pulled value
```

This allows for concise operations on cross-sections of multiple stacks.

### 12.5 Integration with Future Features

Future language features must respect the placeholder and binding semantics defined in this proposal. Extensions to ual's execution model must explicitly address interaction with `pull()`, placeholders, and binding mechanisms. A test suite for core behaviors will be established to ensure that future features maintain compatibility with these mechanisms.

Placeholders have lexical scope and do not propagate through macro expansion or generic instantiation. The detailed interactions with future macro or generic features will be addressed when those features are added to the language.

## 13. Comparison with Other Languages

### 13.1 Pull Operation Comparison

The `pull` operation can be compared to similar operations in other languages:

| Language   | Operation            | Semantics                               |
|------------|----------------------|------------------------------------------|
| Python     | `lst.pop()`          | Remove and return last element           |
| JavaScript | `array.pop()`        | Remove and return last element           |
| Forth      | `DUP PROCESS DROP`   | Duplicate, process, then drop (3 steps)  |
| ual        | `stack.pull()`       | Atomic peek-and-pop (1 step)             |

ual's approach most closely resembles Python's `pop()` but with consistent semantics across different perspectives.

### 13.2 Discard Pattern Comparison

The discard pattern has parallels in several languages:

| Language   | Discard Syntax     | Contexts                                |
|------------|--------------------|-----------------------------------------|
| Go         | `_`                | Assignments, Returns                     |
| Rust       | `_`                | Pattern matching, Variables              |
| Python     | `_` (convention)   | Variables (not enforced)                 |
| ual        | `_`                | Assignments, Parameters                  |

ual's approach follows Go's model closely, using `_` as a special token that explicitly indicates discarded values.

### 13.3 Value Binding Comparison

ual's value binding can be compared to approaches in other languages:

| Language   | Binding Mechanism                     | Contexts                         |
|------------|--------------------------------------|----------------------------------|
| Rust       | `match x { Value(v) => ... }`        | Pattern matching                  |
| JavaScript | `result.then(value => { ... })`      | Promises/callbacks                |
| Ruby       | `value.then { \|x\| ... }`             | Block parameters                  |
| ual        | `result.consider { if_ok(v) { ... }` | Consider blocks                   |

ual's approach balances explicitness with conciseness, avoiding both excessive verbosity and hidden magic.

## 14. Conclusion

The enhanced execution mechanics proposed in this document—pull operation, discard pattern, and consistent value binding—address common patterns in ual code to improve clarity and reduce verbosity. They build upon ual's existing foundations while maintaining its commitment to explicitness and container-centric design.

These features work together as a coherent set of improvements that make common programming patterns more concise and maintainable. The proposal carefully delineates between patterns where each approach is appropriate, ensuring that new features enhance rather than undermine the language's core paradigms.

Key benefits of these enhancements include:

1. **Reduced Boilerplate**: Common peek-then-pop patterns become more concise without sacrificing clarity.
2. **Clear Intent**: The discard pattern makes value discarding explicit rather than implicit.
3. **Consistent Binding**: Value access becomes more uniform across different language contexts.
4. **Edge Case Handling**: Clear semantics for all operations, even in edge cases like empty stacks or out-of-bounds accesses.
5. **Backward Compatibility**: All features maintain compatibility with existing code.

By addressing edge cases explicitly and providing comprehensive examples of both proper usage and anti-patterns, this proposal offers a robust foundation for implementation. The phased approach to implementation ensures that the most valuable features can be delivered quickly while more complex aspects receive appropriate attention.

These enhancements represent a valuable addition to ual's feature set that will benefit developers working on embedded systems and other resource-constrained environments, making the language more expressive and maintainable without compromising its distinctive design philosophy.

The explicit handling of all edge cases ensures that the proposal is complete and implementable:

1. **Perspective Interactions**: The proposal clearly defines that `pull()` must maintain identical element selection behavior as `pop()` for all perspective types, including MAXFO/MINFO with custom comparison functions, and establishes the error message for hashed perspective without a key parameter.

2. **Placeholder Conflicts**: The proposal establishes definitive precedence rules when multiple placeholder systems are active simultaneously, maintains the independence of different placeholder types, and adds compiler warnings for complex expressions with multiple placeholders.

3. **Discard Pattern Coverage**: The proposal specifies that the discard pattern works in all contexts where variable binding would be valid, including nested destructuring, with no limit to the number of discard markers.

4. **Consider Block Binding**: The proposal prevents conflicts between binding parameters and reserved names, clarifies scope rules for shadowing, and defines error handling for invalid binding parameters.

5. **Pull with Crosstacks**: The proposal defines behavior for `pull()` from crosstack levels, requiring completeness across all constituent stacks and suggesting alternatives for partial operations.

6. **Mixed Notation Clarity**: The proposal recommends against mixing placeholder types in expressions and includes lint warnings to promote clear, consistent code.

7. **Placeholder Lifecycle**: The proposal precisely defines statement boundaries for placeholder validity, with clear examples of valid and invalid usage patterns.

8. **Error Message Consistency**: The proposal mandates that `pull()` must produce identical error messages to `pop()` for the same error conditions.

9. **Macro and Generic Integration**: The proposal establishes that placeholders have lexical scope and do not propagate through potential future language extensions like macros or generics.

10. **Future Compatibility**: The proposal includes the principle that future language features must respect the placeholder and binding semantics defined here.

With these considerations in place, the enhanced execution mechanics provide a solid foundation for more expressive, maintainable ual code while preserving the language's core philosophy and ensuring compatibility with both existing and future features.