# ual 1.5 PROPOSAL: Pragmatic Genericity for Stack-Based Programming
This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the version that the proposal is targeting to be integrated into the main ual spec in a forthcoming release.

---

## 1. Introduction

This proposal introduces pragmatic genericity features for ual that allow algorithms to operate on stacks with different element types while maintaining the language's explicit, container-centric philosophy. Unlike traditional genericity approaches that rely heavily on complex type systems, ual's approach leverages its stack-based nature to provide generic capabilities with minimal additional complexity.

The proposed features include an extended `bring` operation, type-based switching, and explicit type annotations, all designed to work harmoniously with ual's existing design principles and container-centric model. These features enable generic programming without requiring the overhead of traditional template-based or interface-based genericity systems.

## 2. Background and Motivation

### 2.1 The Need for Generic Programming

Generic programming allows developers to write algorithms that work across different data types without sacrificing type safety or requiring code duplication. In a stack-based language like ual, genericity is particularly valuable for:

1. **Standard Algorithms**: Operations like reversing, sorting, or filtering stacks should work regardless of element type
2. **Container Utilities**: Functions that manipulate stack structures without depending on element types
3. **Generic Data Structures**: Building higher-level data structures on top of stacks that can work with any element type
4. **Code Reuse**: Avoiding duplication of essentially identical algorithms for different stack types

Without genericity, developers must either write separate implementations for each type or resort to non-type-safe approaches that lose ual's safety guarantees.

### 2.2 Challenges in Stack-Based Genericity

Adding genericity to a stack-based language presents unique challenges:

1. **Type Safety**: How to maintain type safety when working with unknown stack types
2. **Value Movement**: How to move values between stacks of different or unknown types
3. **Type-Specific Operations**: How to perform operations that depend on the specific element type
4. **Container-Centric Model**: How to maintain ual's container-centric philosophy while adding genericity

### 2.3 Lessons from Other Languages

Traditional approaches to genericity in other languages offer important lessons:

#### 2.3.1 C++ Templates

C++ uses compile-time templates that generate specialized code for each type:

```cpp
template <typename T>
void reverse(std::vector<T>& vec) {
    // Implementation that works for any type T
}
```

While powerful, this approach leads to code bloat, complex error messages, and a steep learning curve.

#### 2.3.2 Java/C# Generics

Java and C# use erasure-based or runtime-type-based generics:

```java
public <T> void reverse(List<T> list) {
    // Implementation with type parameters but limited type-specific operations
}
```

This approach provides cleaner syntax but often restricts operations on generic types.

#### 2.3.3 Go Interfaces

Go uses interfaces for a more lightweight approach to genericity:

```go
func Process(data interface{}) {
    switch v := data.(type) {
    case int:
        // Handle integer
    case string:
        // Handle string
    }
}
```

This type switch approach provides flexibility without complex template machinery.

#### 2.3.4 Rust Traits

Rust uses traits to constrain generic types:

```rust
fn sort<T: Ord>(slice: &mut [T]) {
    // Implementation for any type that implements Ord
}
```

This provides powerful constraints while maintaining safety.

### 2.4 The ual Opportunity

ual's stack-based nature presents a unique opportunity for a different approach to genericity:

1. **Container-Centric**: Focus on operations on stacks rather than individual values
2. **Explicit Type Handling**: Make type conversions and tests explicit
3. **Pragmatic Approach**: Provide just enough genericity for common use cases
4. **Zero Runtime Overhead**: Implement genericity features with compile-time checks

## 3. Proposed Generic Features

### 3.1 Extended `bring` Operation

We propose extending ual's existing `bring_<type>` operation to include a type-preserving variant:

```lua
// Current type-converting operations:
@i: bring_string(s.pop())  // Convert string to integer during transfer
@f: bring_integer(i.pop()) // Convert integer to float during transfer

// New type-preserving operation:
@target: bring(source.pop()) // Transfer with original type preserved
```

The new `bring` operation works with any compatible source and target stacks, preserving the original type of the value. This enables generic functions to move values between stacks without knowing their specific types.

At the compiler level, this is implemented as a type-checked operation that ensures the target stack can accept values of the source stack's element type.

### 3.2 Type-Based Switch Statement

We propose adding a type-based switch statement similar to Go's type switch:

```lua
switch_type(stack)
  case Stack(Integer):
    // Handle integer stack
    fmt.Println("Integer stack with", stack.depth(), "elements")
  
  case Stack(String):
    // Handle string stack
    fmt.Println("String stack with", stack.depth(), "elements")
  
  default:
    // Handle other stack types
    fmt.Println("Unknown stack type")
end_switch
```

This enables generic functions to provide type-specific behavior when needed while maintaining type safety. The `switch_type` statement performs a compile-time or runtime type check on the stack and executes the matching case.

### 3.3 Explicit Type Annotations

We propose adding explicit type annotations for function parameters and return values:

```lua
function reverse(stack Stack) Stack {
  // Implementation for any stack type
}

function map(input Stack, output Stack, mapper function) {
  // Implementation for mapping between stacks
}
```

These annotations provide type information for generic functions without requiring complex template syntax. The compiler uses these annotations to ensure type safety in generic functions.

For more specific type constraints, we can use qualified types:

```lua
function add_elements(stack Stack(Numeric)) {
  // Only works with stacks of numeric types
}
```

Where `Numeric` is a type qualifier that includes Integer, Float, etc.

### 3.4 Stack Type Introspection

To support generic programming, stacks need basic type introspection capabilities:

```lua
element_type = stack.type()  // Get the element type of a stack
is_compatible = stack1.compatible_with(stack2)  // Check if stacks have compatible types
```

These operations provide the necessary information for generic functions to work with stacks of unknown types.

## 4. Examples

### 4.1 Generic Stack Reverse

```lua
function reverse(stack Stack) {
  // Create a temporary stack
  @Stack.new(Any): alias:"temp"
  
  // Move all elements to temporary stack (reverses order)
  while_true(stack.depth() > 0)
    @temp: bring(stack.pop())
  end_while_true
  
  // Move all elements back to original stack
  while_true(@temp: depth() > 0)
    stack.push(@temp: pop())
  end_while_true
}

// Usage with different stack types
function demo_reverse() {
  @Stack.new(Integer): alias:"i"
  @Stack.new(String): alias:"s"
  
  @i: push:1 push:2 push:3
  @s: push:"alpha" push:"beta" push:"gamma"
  
  reverse(i)  // Now contains 3, 2, 1
  reverse(s)  // Now contains "gamma", "beta", "alpha"
}
```

### 4.2 Generic Map Function

```lua
function map(source Stack, target Stack, mapper function) {
  // Create temporary stack for preserving original order
  @Stack.new(Any): alias:"temp"
  
  // Move all elements to temporary stack (reverses order)
  while_true(source.depth() > 0)
    @temp: bring(source.pop())
  end_while_true
  
  // Apply mapper and move elements back
  while_true(@temp: depth() > 0)
    value = @temp: pop()
    result = mapper(value)
    target.push(result)
    source.push(value)  // Restore original stack
  end_while_true
}

// Usage
function demo_map() {
  @Stack.new(Integer): alias:"i"
  @Stack.new(String): alias:"s"
  
  @i: push:1 push:2 push:3
  
  // Map integers to strings
  map(i, s, function(n) return n.to_string() end)
  
  // s now contains "1", "2", "3"
}
```

### 4.3 Type-Specific Processing

```lua
function process_stack(stack Stack) {
  switch_type(stack)
    case Stack(Integer):
      // Integer-specific processing
      sum = 0
      @Stack.new(Integer): alias:"temp"
      
      // Copy stack to preserve original
      while_true(stack.depth() > 0)
        value = stack.pop()
        sum = sum + value
        @temp: push(value)
      end_while_true
      
      // Restore original stack
      while_true(@temp: depth() > 0)
        stack.push(@temp: pop())
      end_while_true
      
      fmt.Println("Sum of integers:", sum)
    
    case Stack(String):
      // String-specific processing
      combined = ""
      @Stack.new(String): alias:"temp"
      
      while_true(stack.depth() > 0)
        value = stack.pop()
        combined = combined + value
        @temp: push(value)
      end_while_true
      
      while_true(@temp: depth() > 0)
        stack.push(@temp: pop())
      end_while_true
      
      fmt.Println("Combined string:", combined)
    
    default:
      fmt.Println("Cannot process stack of this type")
  end_switch
}
```

### 4.4 Generic Quicksort

```lua
function quicksort(stack Stack, compare function) {
  // Base case: stack with 0 or 1 elements is already sorted
  if_true(stack.depth() <= 1)
    return
  end_if_true
  
  @Stack.new(Any): alias:"less"
  @Stack.new(Any): alias:"equal"
  @Stack.new(Any): alias:"greater"
  
  // Pick the first element as pivot
  pivot = stack.pop()
  @equal: push(pivot)
  
  // Partition remaining elements
  while_true(stack.depth() > 0)
    elem = stack.pop()
    result = compare(elem, pivot)
    
    if_true(result < 0)
      @less: push(elem)
    end_if_true
    
    if_true(result == 0)
      @equal: push(elem)
    end_if_true
    
    if_true(result > 0)
      @greater: push(elem)
    end_if_true
  end_while_true
  
  // Recursively sort partitions
  quicksort(less, compare)
  quicksort(greater, compare)
  
  // Recombine partitions
  while_true(@less: depth() > 0)
    stack.push(@less: pop())
  end_while_true
  
  while_true(@equal: depth() > 0)
    stack.push(@equal: pop())
  end_while_true
  
  while_true(@greater: depth() > 0)
    stack.push(@greater: pop())
  end_while_true
}

function demo_quicksort() {
  @Stack.new(Integer): alias:"i"
  @Stack.new(String): alias:"s"
  
  @i: push:3 push:1 push:4 push:2
  @s: push:"delta" push:"alpha" push:"gamma" push:"beta"
  
  // Sort integers
  quicksort(i, function(a, b) return a - b end)
  
  // Sort strings
  quicksort(s, function(a, b) 
    if_true(a < b) return -1 end_if_true
    if_true(a > b) return 1 end_if_true
    return 0
  end)
  
  // i now contains 1, 2, 3, 4
  // s now contains "alpha", "beta", "delta", "gamma"
}
```

## 5. Design Rationale

### 5.1 Why Pragmatic Genericity?

ual's approach to genericity differs significantly from traditional approaches in other languages:

1. **Container-Centric vs. Value-Centric**: Rather than focusing on generic types for individual values, ual focuses on operations on stacks as containers.

2. **Explicit vs. Implicit**: Rather than using implicit template instantiation or type erasure, ual makes type handling explicit through operations like `bring` and `switch_type`.

3. **Minimal Additions vs. Complex Type Systems**: Rather than adding a complex type system with concepts like higher-kinded types or variance, ual adds just enough to support common generic programming patterns.

This pragmatic approach aligns well with ual's overall design philosophy:

1. **Explicitness**: Making operations and type handling visible rather than hidden
2. **Progressive Discovery**: Simple patterns remain simple, complex patterns build naturally
3. **Zero Runtime Overhead**: Genericity implemented through compile-time checks
4. **Embedded Systems Focus**: Lightweight approach suitable for resource-constrained environments

### 5.2 The Role of the `bring` Operation

The extended `bring` operation is central to ual's genericity model:

1. It builds on the existing `bring_<type>` pattern, maintaining consistency
2. It makes value transfers between stacks explicit, even in generic contexts
3. It preserves type safety while enabling generic operations
4. It avoids introducing new conceptual models for value movement

By extending an existing operation rather than adding entirely new mechanisms, we maintain the conceptual integrity of the language while adding generic capabilities.

### 5.3 Type-Based Switch vs. Interface System

For handling type-specific behaviors, we chose a type-based switch rather than a traditional interface system:

1. **Explicitness**: The type switch makes type testing explicit rather than hidden behind interface dispatch
2. **Simplicity**: It avoids the complexity of a full interface system with implementation checking
3. **Go Influence**: It follows the pattern of Go, which is already an influence on ual's design
4. **Flexibility**: It allows handling types that weren't designed to work together

This approach provides the necessary type-specific behavior while maintaining ual's explicit, straightforward style.

### 5.4 Why Not Traditional Generic Type Parameters?

We deliberately avoided C++-style generic type parameters (`function<T>`) for several reasons:

1. **Syntactic Complexity**: They add significant syntactic overhead
2. **Conceptual Mismatch**: They focus on parameterizing individual values rather than operations on containers
3. **Implementation Complexity**: They require complex template instantiation machinery
4. **User Experience**: They often lead to cryptic error messages and steep learning curves

ual's approach achieves similar goals with a simpler, more explicit model that better fits its container-centric philosophy.

### 5.5 Genericity Without the Weight

Traditional genericity systems often come with significant complexity and overhead:

1. **C++ Templates**: Code bloat from template instantiation, complex SFINAE patterns, challenging error messages
2. **Java Generics**: Type erasure limitations, wildcards complexity, runtime overhead
3. **Rust Traits**: Complex higher-ranked trait bounds, implicit implementation selection

ual's approach achieves genericity with much less complexity:

1. **No Template Instantiation**: No need to generate separate code for each type
2. **No Complex Constraints**: Simple type qualifiers rather than complex constraint systems
3. **Explicit Type Handling**: Clear operations for type conversion and testing
4. **Container-Centric Model**: Focus on stack operations rather than individual value types

This lightweight approach is particularly valuable for embedded systems programming, where code size and compilation complexity matter.

## 6. Implementation Considerations

### 6.1 Extended `bring` Operation

The implementation of the extended `bring` operation involves:

1. **Type Checking**: Verify that the target stack can accept values of the source element type
2. **Value Transfer**: Move the value between stacks while preserving its type
3. **Error Handling**: Generate appropriate errors for incompatible stack types

At the compiler level, this requires tracking stack element types and ensuring type compatibility at compile time when possible.

### 6.2 Type-Based Switch Statement

The `switch_type` statement can be implemented as:

1. **Type Testing**: Check the runtime type of the specified stack or value
2. **Case Matching**: Execute the code block for the matching type
3. **Default Handling**: Execute the default case if no type matches

For stack type switching, this involves checking the element type of the stack rather than the type of an individual value.

### 6.3 Type Annotations

Type annotations for function parameters and return values require:

1. **Parser Updates**: Parse the type annotations in function declarations
2. **Type Checking**: Verify that arguments match parameter types
3. **Type Propagation**: Use annotation information to check type safety within functions

### 6.4 Backward Compatibility

These features are designed to maintain backward compatibility:

1. **Optional Features**: Existing code continues to work without using generic features
2. **Consistent Patterns**: New features follow established patterns in the language
3. **Progressive Adoption**: Developers can gradually adopt generic features

## 7. Comparison with Other Languages

### 7.1 vs. C++ Templates

```cpp
// C++ template approach
template <typename T>
void reverse(std::vector<T>& vec) {
    std::reverse(vec.begin(), vec.end());
}

// ual approach
function reverse(stack Stack) {
    // Implementation using stack operations
}
```

ual's approach is more explicit and avoids template instantiation complexity.

### 7.2 vs. Go Type Switch

```go
// Go type switch
func process(value interface{}) {
    switch v := value.(type) {
    case int:
        fmt.Println("Integer:", v)
    case string:
        fmt.Println("String:", v)
    }
}

// ual type switch
function process(stack Stack) {
    switch_type(stack)
        case Stack(Integer):
            fmt.Println("Integer stack")
        case Stack(String):
            fmt.Println("String stack")
    end_switch
}
```

ual's approach is similar to Go's but focuses on stack types rather than individual values.

### 7.3 vs. Rust Traits

```rust
// Rust trait-based approach
fn process<T: Display>(value: T) {
    println!("{}", value);
}

// ual approach
function process(value Any) {
    switch_type(value)
        case String:
            fmt.Println(value)
        case Integer:
            fmt.Println(value.to_string())
        // Other cases
    end_switch
}
```

ual's approach is more explicit about type-specific behavior rather than relying on trait implementations.

## 8. Future Directions

While this proposal focuses on pragmatic genericity features, several potential extensions could be considered in the future:

1. **Type Qualifiers**: More sophisticated type qualifiers like `Numeric`, `Comparable`, etc.
2. **Stack Composition**: Ways to compose stacks with different element types
3. **Generic Type Inference**: Better type inference for generic functions
4. **Interface System**: A lightweight interface system for more formalized type behavior

## 9. Conclusion

This proposal introduces pragmatic genericity features for ual that enable generic programming while maintaining the language's explicit, container-centric philosophy. By extending the `bring` operation, adding type-based switching, and supporting explicit type annotations, we provide the necessary tools for writing generic algorithms without the complexity of traditional genericity systems.

These features maintain ual's focus on explicitness, progressive discovery, and zero runtime overhead, making them well-suited for the language's embedded systems target. They enable developers to write reusable, type-safe code that works across different stack types, enhancing ual's utility while preserving its unique character.

We recommend adopting these pragmatic genericity features for ual 1.5 to enhance the language's expressiveness and reduce code duplication while maintaining its distinctive stack-based, container-centric approach.