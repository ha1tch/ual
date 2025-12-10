# ual 1.9 PROPOSAL: Sidestacks (Part 2) - Advanced Semantics

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Junction Lifecycle Management

While the core sidestack proposal established the foundation for creating junctions with `tag` and connecting stacks with `bind`, this section formalizes the complete lifecycle of junctions including removal, rebinding, and invalidation scenarios.

### 1.1 Junction Removal

The `untag` operation removes a junction from a stack element:

```lua
// Remove a specific junction named "j1" from element at index 1
@main: untag(1, "j1")

// Stack mode syntax
@main: untag:1 name:"j1"
```

When a junction is removed:

1. The association between the element and the junction name is eliminated
2. Any bound stack is automatically unbound from this junction
3. The element itself remains unchanged in the stack
4. Operations on the corresponding junction reference (`main^j1`) become invalid

If the specified junction does not exist, the operation has no effect.

For convenience, an operation to remove all junctions from an element is provided:

```lua
// Remove all junctions from element at index 1
@main: untag_all(1)

// Stack mode syntax
@main: untag_all:1
```

### 1.2 Junction Binding and Rebinding

The binding between a junction and a stack is managed through explicit operations:

```lua
// Bind a stack to a junction
@main^j1: bind(@details)

// Unbind a stack from a junction
@main^j1: unbind()

// Check if a junction is bound
is_bound = main^j1.is_bound()
```

When rebinding a junction to a different stack:

1. The previous binding is automatically removed
2. The new stack is bound to the junction
3. No operations are performed on either the old or new stack
4. Type compatibility is checked at compile-time (or runtime for dynamic typing)

Example of rebinding:

```lua
// Initial binding
@main^j1: bind(@details1)

// Later rebinding to a different stack
@main^j1: bind(@details2)  // details1 is automatically unbound
```

Attempting to bind an incompatible stack type results in a compile-time error:

```lua
@Stack.new(Integer): alias:"main"
@Stack.new(String): alias:"details"

@main: tag(0, "j1")
@main^j1: bind(@details)  // Error: Cannot bind Stack<String> to junction in Stack<Integer>
```

### 1.3 Junction Invalidation

Junction invalidation occurs in several scenarios, with specific semantics for each:

#### 1.3.1 Element Removal

When an element with junctions is removed from a stack:

```lua
// Element at index 1 has a junction "j1"
@main: drop(1)  // Remove the element
```

The following occurs:

1. All junctions associated with that element are automatically removed
2. Any stacks bound to those junctions remain intact but are unbound
3. Subsequent operations on those junction references (`main^j1`) become invalid

#### 1.3.2 Stack Destruction

When a stack with junctions is destroyed:

```lua
// main has elements with junctions
main = nil  // Destroy the stack reference
```

1. All junctions in the stack are invalidated
2. Bound stacks remain intact but are unbound
3. Operations on junction references become invalid

#### 1.3.3 Explicit Stack Clearing

When a stack is explicitly cleared:

```lua
// Clear all elements from the stack
@main: clear()
```

1. All junctions in the stack are invalidated
2. Bound stacks remain intact but are unbound
3. Operations on junction references become invalid

### 1.4 Junction Movement

When elements with junctions move within a stack (due to operations like `swap`, `rot`, etc.), the junctions move with the elements:

```lua
// Element at index 0 has junction "j1"
@main: swap(0, 1)  // Swap elements at indices 0 and 1
```

After this operation:
1. The junction "j1" is now associated with the element at index 1
2. The binding remains intact
3. References to the junction through its name continue to work (`main^j1`)

This ensures junctions remain attached to their elements regardless of element position within the stack.

## 2. Junction Traversal Guarantees

This section establishes formal guarantees about junction traversal, access patterns, and error semantics.

### 2.1 Traversal Order

When accessing elements through junctions, specific traversal order guarantees apply:

#### 2.1.1 Single Junction Traversal

For direct junction access:

```lua
// Access the stack bound to junction "j1"
@main^j1: operation
```

1. The junction is resolved at the time of access
2. The bound stack's current perspective is used for the operation
3. The operation executes on the bound stack in its current state

#### 2.1.2 Multi-Junction Traversal

For traversals across multiple junctions:

```lua
// Traverse multiple junctions
@main^j1^j2^j3: operation
```

1. Junctions are resolved from left to right
2. Each junction must exist and be bound for the traversal to succeed
3. The perspective of the final stack in the chain applies to the operation
4. The operation executes on the final stack in the chain

### 2.2 Error Semantics

Junction operations have well-defined error semantics:

#### 2.2.1 Non-existent Junction

When accessing a non-existent junction:

```lua
// Junction "j1" does not exist
@main^j1: operation
```

1. A runtime error is raised: "Junction 'j1' does not exist on stack"
2. The operation is not executed
3. Program execution continues according to error handling rules

#### 2.2.2 Unbound Junction

When accessing an unbound junction:

```lua
// Junction "j1" exists but is not bound to any stack
@main^j1: operation
```

1. A runtime error is raised: "Junction 'j1' is not bound to any stack"
2. The operation is not executed
3. Program execution continues according to error handling rules

#### 2.2.3 Invalid Junction Access

When a junction reference becomes invalid:

```lua
// Store reference to a junction
junction = main^j1

// Later, after the junction is removed
@junction: operation
```

1. A runtime error is raised: "Invalid junction reference"
2. The operation is not executed
3. Program execution continues according to error handling rules

### 2.3 Type Safety Boundaries

Junctions maintain type safety across stack boundaries:

#### 2.3.1 Cross-Junction Type Safety

Operations across junctions respect the type of each stack:

```lua
@Stack.new(Integer): alias:"main"
@Stack.new(String): alias:"details"

@main: tag(0, "j1")
@main^j1: bind(@details)

@main: push(42)        // Valid: Integer into Integer stack
@main^j1: push("hello")  // Valid: String into String stack
@main^j1: push(42)       // Error: Integer cannot go into String stack
```

#### 2.3.2 Junction Type Constraints

Junction binding enforces type compatibility:

```lua
// Junction binding with type constraints
@main^j1: bind(@details, {compatible_with: String})
```

This ensures that only stacks with compatible element types can be bound.

#### 2.3.3 Type Conversion Across Junctions

Explicit type conversion is required when transferring values across junctions with incompatible types:

```lua
// Element from main (Integer) to details (String)
@main^j1: bring_integer(main.pop())  // Explicit conversion required
```

## 3. Junction-Specific Stack Operations

This section defines how stack operations interact with junctions, including cross-junction operations, propagation rules, and borrowing semantics.

### 3.1 Cross-Junction Operations

Operations can span across junctions with specific semantics:

#### 3.1.1 Value Movement Across Junctions

Moving values between stacks connected by junctions:

```lua
// Transfer top element from main to its junction "j1"
@main^j1: <main

// Transfer top element from junction "j1" to main
@main: <main^j1
```

This is equivalent to:

```lua
// Same as @main^j1: <main
value = main.pop()
main^j1.push(value)

// Same as @main: <main^j1
value = main^j1.pop()
main.push(value)
```

#### 3.1.2 Stack Operations Across Junctions

Stack operations can span junctions:

```lua
// Apply an operation across junctions
@main: zip(@main^j1)  // Interleave elements from main and main^j1
```

The semantics depend on the specific operation, but generally:

1. The operation receives elements from both stacks
2. Results go to the target stack (the one before the colon)
3. Type safety is enforced at operation boundaries

### 3.2 Operation Propagation

Operations on junctions affect only the target stack, not connected stacks:

```lua
// Clear the details stack bound to junction "j1"
@main^j1: clear()

// This does not affect the main stack or the junction itself
```

Exceptions to this rule must be explicitly documented for specific operations.

#### 3.2.1 Junction-Aware Operations

Certain operations are junction-aware and have specific behavior with junctions:

```lua
// Deep copy including all junction-connected stacks
deep_copy = main.deep_clone(true)  // true indicates include junctions

// Deep equals comparing junction structure
are_equal = main.deep_equals(other, true)  // true indicates compare junctions
```

#### 3.2.2 Propagation Control

Operations can control propagation explicitly:

```lua
// Operation with propagation control
@main: transform(transform_func, {propagate: true})
```

When `propagate` is true, the operation applies to the target stack and recursively to all bound stacks.

### 3.3 Junction-Aware Borrowing

Borrowed segments interact with junctions according to specific rules:

#### 3.3.1 Borrowing Across Junctions

Borrowing can span junction boundaries:

```lua
// Borrow a segment from a junction-connected stack
@segment: borrow([0..2]@main^j1)
```

This creates a borrowed segment from the stack bound to junction "j1".

#### 3.3.2 Junction Preservation in Borrowing

When borrowing elements with junctions:

```lua
// Borrow elements that have junctions
@segment: borrow([0..2]@main)
```

1. Junctions associated with the borrowed elements remain accessible
2. Junction operations through the borrowed segment are valid
3. Junction bindings remain intact

Example:

```lua
// Element 1 in main has junction "j1"
@segment: borrow([0..2]@main)

// Access junction through borrowed segment
@segment^j1: operation  // Valid, accesses the junction on element 1
```

#### 3.3.3 Borrowing with Junction Tree

For complex borrowing across multiple junctions:

```lua
// Borrow with junction tree specification
@segment: borrow([0..2]@main, {include_junctions: ["j1", "j2"]})
```

This borrows not only the main segment but also specified junctions and their bound stacks.

### 3.4 Junction State and Persistence

Junction state can be inspected and preserved:

#### 3.4.1 Junction Inspection

Operations to inspect junction state:

```lua
// Get all junction names on an element
junction_names = main.junctions_at(1)

// Check if an element has a specific junction
has_junction = main.has_junction(1, "j1")

// Get information about a junction
junction_info = main.junction_info(1, "j1")
```

#### 3.4.2 Junction State Preservation

Junction state can be saved and restored:

```lua
// Capture junction state
state = main.capture_junction_state()

// Restore junction state
main.restore_junction_state(state)
```

This allows for preserving complex junction relationships across program states.

## 4. Implementation Notes

This section provides guidance for implementing the junction lifecycle management, traversal guarantees, and junction-specific operations.

### 4.1 Junction Reference Implementation

Junctions are implemented as metadata attached to stack elements:

```
Junction {
  name: string
  target_stack: Stack | null
  properties: Map<string, any>
}

Element {
  value: any
  junctions: Map<string, Junction>
}
```

This structure enables efficient junction lookup while maintaining the stack's core functionality.

### 4.2 Performance Considerations

Implementations should optimize for:

1. Fast junction lookup by name
2. Efficient traversal across multiple junctions
3. Minimal overhead for stacks without junctions
4. Compact representation of junction metadata

### 4.3 Memory Management

Junction reference cycles require careful handling:

1. Weak references should be used where appropriate to prevent memory leaks
2. Junction binding should not create strong reference cycles
3. Stack destruction should properly clean up all junction references

### 4.4 Compatibility with Existing Features

Junction operations must maintain compatibility with:

1. Stack perspectives (LIFO, FIFO, etc.)
2. Type safety mechanisms
3. Borrowed segments
4. Serialization and debugging tools

## 5. Conclusion

This expanded specification formalizes the complete lifecycle of junctions, establishes clear traversal guarantees, and defines how junction-specific operations behave. These additions provide a comprehensive foundation for implementing and using sidestacks in a variety of contexts, from simple hierarchical data to complex graph structures.

By addressing junction removal, rebinding, invalidation, traversal semantics, error handling, and cross-junction operations, this specification ensures that sidestacks can be implemented consistently and used effectively across different implementations of the ual language.