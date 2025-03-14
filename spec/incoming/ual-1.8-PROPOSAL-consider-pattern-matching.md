# ual 1.8 PROPOSAL: Generalized Pattern Matching with Consider Blocks

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction

This proposal extends ual's container-centric philosophy by generalizing the `.consider{}` construct beyond its initial error handling scope. Building on insights from the `@error` stack proposal, stack borrowing, and typed stacks, it creates a unified pattern matching mechanism that provides elegant, expressive control flow while maintaining ual's commitment to explicitness and efficiency. This extension completes ual's repertoire of control structures, making pattern matching and conditional logic as native to the language as stack operations.

### 1.1 Core Concept: Pattern Matching as First-Class Control Flow

The central insight of this proposal is recognizing pattern matching as a fundamental programming paradigm worthy of first-class status in the language. While initially conceived for result handling with `if_ok`/`if_err` branches, the `.consider{}` construct naturally extends to general pattern matching through a set of related branches:

```lua
value.consider {
  if_match(function(v) return v > 10 end) {
    -- Handle values greater than 10
  }
  if_equal(5) {
    -- Handle value exactly equal to 5
  }
  if_type(String) {
    -- Handle string values
  }
  if_else {
    -- Handle all other cases
  }
}
```

This pattern-centric approach to control flow creates a more declarative, intention-revealing code style while enabling sophisticated compiler optimizations similar to those for `switch_case` statements.

## 2. Historical Context and Motivation

### 2.1 Pattern Matching Across Programming Paradigms

Pattern matching has evolved through multiple programming paradigms:

- **Logic Programming**: Prolog (1972) pioneered pattern matching as its primary execution model through unification.
- **Functional Programming**: ML (1973) and its descendants (OCaml, Haskell) elevated pattern matching to a core language feature.
- **Object-Oriented Programming**: Languages like Scala and Kotlin integrated pattern matching with object orientation.
- **Multi-Paradigm Languages**: Rust's match expressions combine pattern matching with ownership and borrowing semantics.

Traditional imperative languages have generally underserved pattern matching, relying on chains of conditionals or switch statements with limited matching capabilities.

### 2.2 From Error Handling to General Matching

The evolution of `.consider{}` from error handling to general pattern matching parallels the history of pattern matching in programming languages:

1. **Initial Specialized Purpose**: First introduced for structured error handling, focusing on `Result` types with `Ok`/`Err` variants.
2. **Recognition of Broader Utility**: Observation that the pattern generalizes to many other conditional scenarios.
3. **Language Integration**: Extending to integrate with ual's type system, stack perspectives, and borrowing.
4. **Optimization Opportunities**: Recognizing compiler optimization potential similar to `switch_case`.

### 2.3 Code Blocks in Programming Language History

Code blocks as language constructs have a rich history:

- **Algol 60** (1960) introduced the `begin`/`end` block structure that influenced many languages.
- **Smalltalk** (1980) pioneered blocks as first-class objects that could be passed and invoked.
- **Ruby** (1995) made blocks ubiquitous throughout its standard library, using them for iteration, resource management, and DSLs.
- **Rust** (2010s) used blocks extensively with ownership semantics, where code block boundaries become significant for borrowing.

The evolution from syntactic grouping to semantically meaningful constructs mirrors ual's approach to making control flow explicit and meaningful.

### 2.4 Ruby's Block Pattern and Its Influence

Ruby's approach to blocks offers particular insights for ual's `.consider{}` design:

```ruby
# Ruby block examples
# Iteration
[1, 2, 3].each do |item|
  puts item
end

# Resource management
File.open("file.txt") do |file|
  # file is automatically closed after block execution
  contents = file.read
end

# Custom control structures
retryable(times: 3) do
  # code that might fail
end
```

Ruby's blocks excel at:
1. **Deferred Execution**: Code to be executed later or conditionally
2. **Context Capture**: Operating within a specific context or environment
3. **Resource Management**: Ensuring proper setup and cleanup
4. **Custom Control Structures**: Enabling domain-specific control flow

These patterns inform ual's approach to code blocks and the `.consider{}` construct, though ual's design emphasizes explicit stack operations and compile-time optimizations.

### 2.5 Rust's Match Expression as Exhaustive Pattern Matching

Rust's `match` expression represents a different approach to pattern matching:

```rust
// Rust match example
match value {
    0 => println!("Zero"),
    1..=5 => println!("Small number"),
    n if n > 100 => println!("Large number"),
    _ => println!("Something else"),
}
```

Key aspects of Rust's approach include:
1. **Exhaustiveness Checking**: Compiler verifies all possible cases are handled
2. **Expression Context**: Can be used anywhere an expression is expected
3. **Binding Patterns**: Can bind variables to parts of the matched value
4. **Pattern Guards**: Expressions that further refine patterns

While powerful, Rust's pattern matching can be complex to learn and reason about. The ual `.consider{}` construct aims for a middle ground—more powerful than simple conditionals but more approachable than Rust's comprehensive pattern matching.

## 3. Proposed Syntax and Semantics

### 3.1 Consider Block Basic Syntax

The consider block has the following general form:

```lua
expression.consider {
  if_pattern1(pattern_args) {
    -- Code to execute when pattern1 matches
  }
  
  if_pattern2(pattern_args) {
    -- Code to execute when pattern2 matches
  }
  
  -- Additional patterns as needed
  
  if_else {
    -- Default case when no patterns match
  }
}
```

Where:
- `expression` is any valid ual expression that produces a value
- `if_pattern` clauses represent different matching conditions
- `pattern_args` are the arguments for the specific pattern matcher
- Code blocks contain the code to execute when a pattern matches

### 3.2 Standard Pattern Matching Constructs

The proposal defines several standard pattern matchers:

#### 3.2.1 `if_equal` - Equality Matching

```lua
value.consider {
  if_equal(5) {
    fmt.Printf("Value is 5\n")
  }
  if_equal("hello") {
    fmt.Printf("Value is 'hello'\n")
  }
}
```

Matches when the expression equals the provided value using the `==` operator.

#### 3.2.2 `if_match` - Predicate Function Matching

```lua
value.consider {
  if_match(function(v) return v > 10 and v < 20 end) {
    fmt.Printf("Value between 10 and 20\n")
  }
  if_match(function(v) return v % 2 == 0 end) {
    fmt.Printf("Value is even\n")
  }
}
```

Matches when the provided function returns a true value when called with the expression.

#### 3.2.3 `if_type` - Type Matching

```lua
value.consider {
  if_type(Integer) {
    fmt.Printf("Value is an integer\n")
  }
  if_type(String) {
    fmt.Printf("Value is a string\n")
  }
}
```

Matches when the expression's type matches the specified type.

#### 3.2.4 `if_ok` and `if_err` - Result Pattern Matching

```lua
result.consider {
  if_ok {
    fmt.Printf("Success: %v\n", _1)
  }
  if_err {
    fmt.Printf("Error: %v\n", _1)
  }
}
```

Specialized patterns for result objects with `Ok` or `Err` fields, as defined in the original error stack proposal.

#### 3.2.5 `if_else` - Default Case

```lua
value.consider {
  if_match(function(v) return v > 0 end) {
    fmt.Printf("Positive\n")
  }
  if_equal(0) {
    fmt.Printf("Zero\n")
  }
  if_else {
    fmt.Printf("Negative\n")
  }
}
```

Executes when no other patterns match. Must be the last pattern in the consider block.

### 3.3 Fallthrough Behavior

Unlike `switch_case`, the `.consider{}` construct does not fall through by default. Only the first matching pattern's code block is executed:

```lua
value.consider {
  if_match(function(v) return v > 0 end) {
    fmt.Printf("Positive\n")
    -- No fallthrough to other patterns
  }
  if_equal(42) {
    -- This will NOT execute even if value is 42,
    -- because the previous pattern already matched
    fmt.Printf("Answer to everything\n")
  }
}
```

This design choice emphasizes clarity and predictability, avoiding the common bugs associated with unintentional fallthrough in switch statements.

### 3.4 Pattern Evaluation Order

Patterns are evaluated in the order they appear in the code:

```lua
value.consider {
  if_equal(0) { ... }         -- Checked first
  if_match(is_small) { ... }  -- Checked second if first didn't match
  if_type(Integer) { ... }    -- Checked third if previous didn't match
  if_else { ... }             -- Only if no other patterns matched
}
```

This guarantee allows programmers to build patterns from specific to general, similar to how function overloading and pattern matching work in functional languages.

### 3.5 Integration with Stack Perspectives

The `.consider{}` construct naturally integrates with ual's stack perspectives:

```lua
@stack: fifo
@stack: push(10) push(20) push(30)

@stack: peek().consider {
  if_equal(10) {
    fmt.Printf("First element is 10 (FIFO perspective)\n")
  }
}

@stack: lifo
@stack: peek().consider {
  if_equal(30) {
    fmt.Printf("First element is 30 (LIFO perspective)\n")
  }
}
```

This integration allows pattern matching to adapt to different stack perspectives, creating powerful combinations for algorithms that benefit from both perspectives and pattern matching.

### 3.6 Integration with Stack Segment Borrowing

Consider blocks work seamlessly with borrowed stack segments:

```lua
scope {
  @segment: borrow([5..10]@stack)
  
  segment.peek(0).consider {
    if_match(function(v) return v > threshold end) {
      process_above_threshold(segment)
    }
    if_else {
      process_normal(segment)
    }
  }
}
```

The integration with borrowed segments allows pattern matching on elements within safety-bounded regions of stacks, combining two of ual's key safety mechanisms.

## 4. Code Block Semantics

### 4.1 Scope and Variable Visibility

Code blocks within `.consider{}` create a new scope for local variables:

```lua
x = 5
value.consider {
  if_equal(10) {
    local x = 20  -- Shadows outer x
    -- x is 20 here
  }
}
-- x is still 5 here
```

Variables declared in one pattern's code block are not visible in other pattern blocks, even if the flow of execution would never reach both blocks.

### 4.2 Return and Break Behavior

A `return` statement within a consider block returns from the enclosing function, not just from the block:

```lua
function process(value)
  value.consider {
    if_equal(0) {
      return "zero"  -- Returns from process() function
    }
  }
  return "non-zero"
}
```

This is consistent with ual's function-level return semantics in other contexts.

### 4.3 Block Result Values

Consider blocks as a whole do not produce a value, but individual pattern blocks can contain expressions that produce values used within the block:

```lua
value.consider {
  if_match(is_valid) {
    result = compute_result(value)
    store_result(result)
  }
}
```

This is different from Rust's match expressions which can return values directly.

## 5. Implementation and Optimization

### 5.1 Compiler Implementation

The `.consider{}` construct would be implemented as a syntactic transformation during the parsing phase:

```lua
// Simplified representation of compiler transformation
value.consider {
  if_equal(5) { block1 }
  if_match(pred) { block2 }
  if_else { block3 }
}

// Transforms to something like:
{
  let temp = value
  if temp == 5 then
    block1
  elseif pred(temp) then
    block2
  else
    block3
  end
}
```

The transformation ensures the expression is evaluated exactly once and stored in a temporary variable to avoid side effects from multiple evaluations.

### 5.2 Optimization Opportunities

The compiler can apply several optimizations to consider blocks:

1. **Decision Tree Optimization**: Converting patterns into an optimized decision tree
2. **Jump Table Generation**: Using jump tables for `if_equal` with consecutive integer values
3. **Type Specialization**: Generating specialized code for `if_type` cases
4. **Constant Folding**: Pre-computing results for constant expressions in patterns
5. **Dead Code Elimination**: Removing unreachable pattern branches

These optimizations parallel those applied to `switch_case` statements as outlined in the ual specification.

### 5.3 Exhaustiveness Checking

Unlike Rust's match, the `.consider{}` construct does not perform exhaustiveness checking by default. However, a compiler flag could enable warnings when no `if_else` clause is provided and the patterns might not cover all possible values.

## 6. Examples and Use Cases

### 6.1 Enhanced Error Handling

```lua
function process_file(filename)
  file_result = io.open(filename, "r")
  
  file_result.consider {
    if_ok {
      file = _1
      
      content_result = file.read("*all")
      file.close()
      
      content_result.consider {
        if_ok {
          return { Ok = _1 }
        }
        if_err {
          return { Err = "Read error: " .. _1 }
        }
      }
    }
    if_err {
      return { Err = "Cannot open file: " .. _1 }
    }
  }
end
```

This layered error handling demonstrates how `.consider{}` creates clear, nested error handling paths.

### 6.2 State Machine Implementation

```lua
function process_event(state, event)
  state.consider {
    if_equal("idle") {
      event.consider {
        if_equal("start") { return "running" }
        if_else { return "idle" }
      }
    }
    if_equal("running") {
      event.consider {
        if_equal("pause") { return "paused" }
        if_equal("stop") { return "idle" }
        if_else { return "running" }
      }
    }
    if_equal("paused") {
      event.consider {
        if_equal("resume") { return "running" }
        if_equal("stop") { return "idle" }
        if_else { return "paused" }
      }
    }
    if_else {
      return "idle"  // Unknown state
    }
  }
end
```

This demonstrates how nested consider blocks create elegant state machine implementations.

### 6.3 Mathematical Function Definition by Cases

```lua
function factorial(n)
  n.consider {
    if_equal(0) {
      return 1
    }
    if_match(function(x) return x > 0 end) {
      return n * factorial(n - 1)
    }
    if_else {
      error("Factorial undefined for negative numbers")
    }
  }
end
```

This pattern mimics mathematical function definition by cases, creating concise, readable algorithm implementations.

### 6.4 Type-Based Processing

```lua
function process_value(value)
  value.consider {
    if_type(Integer) {
      return value * 2
    }
    if_type(String) {
      return value .. value
    }
    if_type(Boolean) {
      return not value
    }
    if_else {
      return nil
    }
  }
end
```

This demonstrates how `.consider{}` with `if_type` creates elegant polymorphic functions.

### 6.5 Advanced Pattern Matching with Predicates

```lua
function classify_number(n)
  n.consider {
    if_match(function(x) return x < 0 end) {
      return "negative"
    }
    if_equal(0) {
      return "zero"
    }
    if_match(function(x) return x > 0 and x % 2 == 0 end) {
      return "positive even"
    }
    if_else {
      return "positive odd"
    }
  }
end
```

This shows how predicate functions enable sophisticated condition testing beyond simple equality.

## 7. Comparison with Other Languages

### 7.1 Ruby's Code Blocks vs. ual's Consider Blocks

Ruby uses anonymous blocks for iteration, resource management, and custom control structures:

```ruby
# Ruby
[1, 2, 3].each do |item|
  puts item
end

File.open("file.txt") do |file|
  contents = file.read
end
```

While ual's consider blocks:

```lua
-- ual
@array: each(function(item) {
  fmt.Printf("%d\n", item)
})

file_result = io.open("file.txt", "r")
file_result.consider {
  if_ok {
    file = _1
    contents = file.read("*all")
    file.close()
  }
}
```

Key differences:
1. Ruby blocks are primarily for deferred execution, while ual's consider blocks focus on pattern matching
2. Ruby uses blocks to create mini-DSLs, while ual emphasizes explicit stack operations
3. Ruby's blocks can be converted to `Proc` objects, while ual maintains a clearer distinction between blocks and functions

### 7.2 Rust's Match Expression vs. ual's Consider Blocks

Rust's match expression provides comprehensive pattern matching:

```rust
// Rust
match value {
    0 => println!("Zero"),
    1..=5 => println!("Small"),
    n if n % 2 == 0 => println!("Even"),
    _ => println!("Other"),
}
```

Compared to ual's consider blocks:

```lua
-- ual
value.consider {
  if_equal(0) {
    fmt.Printf("Zero\n")
  }
  if_match(function(n) return n >= 1 and n <= 5 end) {
    fmt.Printf("Small\n")
  }
  if_match(function(n) return n % 2 == 0 end) {
    fmt.Printf("Even\n")
  }
  if_else {
    fmt.Printf("Other\n")
  }
}
```

Key differences:
1. Rust's match is an expression that returns a value; ual's consider is a statement
2. Rust offers more pattern types (ranges, destructuring, guards); ual focuses on a smaller set of patterns
3. Rust enforces exhaustiveness; ual makes it optional
4. ual's approach emphasizes readability over complex matching capabilities

### 7.3 Haskell's Pattern Matching vs. ual's Consider Blocks

Haskell uses pattern matching at the function definition level:

```haskell
-- Haskell
factorial :: Integer -> Integer
factorial 0 = 1
factorial n = n * factorial (n - 1)
```

While ual uses it within function bodies:

```lua
-- ual
function factorial(n)
  n.consider {
    if_equal(0) {
      return 1
    }
    if_else {
      return n * factorial(n - 1)
    }
  }
end
```

Key differences:
1. Haskell patterns are at the function declaration level; ual's are explicit consider blocks
2. Haskell patterns are tied to type system and algebraic data types; ual's are more imperative
3. Haskell offers destructuring patterns; ual focuses on simpler predicates and equality

### 7.4 Switch Statements vs. ual's Consider Blocks

Traditional switch statements in languages like C, JavaScript, or Go:

```javascript
// JavaScript
switch (value) {
  case 0:
    console.log("Zero");
    break;
  case 1:
  case 2:
    console.log("Small");
    break;
  default:
    console.log("Other");
}
```

Compared to ual's approach:

```lua
-- ual
value.consider {
  if_equal(0) {
    fmt.Printf("Zero\n")
  }
  if_match(function(v) return v == 1 or v == 2 end) {
    fmt.Printf("Small\n")
  }
  if_else {
    fmt.Printf("Other\n")
  }
}
```

Key differences:
1. ual's consider has no fallthrough by default; switches typically do
2. ual's consider works with any expression and pattern; switches typically only use equality
3. ual's consider has a more expression-oriented syntax; switches feel more statement-oriented
4. ual enables predicate functions for complex conditions; switches typically require simpler conditions

## 8. Design Decisions and Rationale

### 8.1 Why Method Chaining Instead of Standalone Construct?

The decision to use method chaining syntax (`value.consider{}`) rather than a standalone construct (`consider value {}`) offers several benefits:

1. **Consistency with Result Handling**: Maintains the established pattern from the error stack proposal
2. **Method-Like Semantics**: Reinforces that the operation is conceptually applied to a value
3. **Extensibility**: Allows future extension to other receiver types like stacks or tables
4. **Expression Integration**: Works naturally with expression results and method chains

This approach reinforces ual's philosophy of making operations explicit and visible in the code.

### 8.2 No-Fallthrough Design Decision

Unlike switch statements in C-like languages, the consider construct does not fall through to subsequent patterns by default. This decision was made for several reasons:

1. **Preventing Bugs**: Unintended fallthrough is a common source of bugs in switch statements
2. **Clarity of Intent**: Each pattern clearly corresponds to a specific action
3. **Optimization Opportunities**: Non-fallthrough enables more aggressive compiler optimizations
4. **Consistency**: Aligns with modern language designs like Swift, Rust, and Kotlin

This makes the code more predictable and reduces the need for explicit `break` statements.

### 8.3 Codeblocks vs. Anonymous Functions

The proposal establishes clear criteria for when to use code blocks versus anonymous functions:

#### Use Code Blocks When:
1. **Control Flow Matters**: The operation involves conditional execution or early returns
2. **Pattern Matching**: The logic depends on matching patterns or types
3. **Readability**: The intent is more clearly expressed through structural code blocks
4. **Compile-Time Optimization**: The pattern can benefit from compiler optimizations

#### Use Anonymous Functions When:
1. **Callback Passing**: The code needs to be passed as an argument to be executed later
2. **Closures**: The code needs to capture and retain its lexical environment
3. **First-Class Treatment**: The code needs to be stored, passed, or returned as a value
4. **Multiple Invocation**: The code will be executed multiple times

This distinction helps developers choose the most appropriate construct for their needs.

### 8.4 Patterns as First-Class Constructs

Treating patterns (`if_equal`, `if_match`, etc.) as first-class constructs rather than just syntax sugar has significant benefits:

1. **Extensibility**: New patterns can be added to the language or libraries
2. **Composition**: Patterns can potentially be composed in the future
3. **Documentation**: Patterns have clear semantics that can be documented
4. **Optimization**: Each pattern type can have specialized optimization strategies

This approach allows the pattern matching system to evolve while maintaining a consistent interface.

## 9. Future Extensions

### 9.1 User-Defined Patterns

A natural extension would allow defining custom patterns:

```lua
-- Hypothetical future syntax for custom patterns
if_range = function(min, max)
  return function(value)
    return value >= min and value <= max
  end
end

value.consider {
  if_range(1, 10) {
    fmt.Printf("Value between 1 and 10\n")
  }
}
```

This would enable domain-specific patterns that make code more expressive and maintainable.

### 9.2 Destructuring Patterns

Another potential extension is destructuring patterns for compound types:

```lua
-- Hypothetical future syntax for destructuring
point.consider {
  if_struct{x = 0, y = _} {
    fmt.Printf("Point on y-axis, y = %d\n", y)
  }
  if_struct{x = _, y = 0} {
    fmt.Printf("Point on x-axis, x = %d\n", x)
  }
}
```

This would bring some of the power of Rust's and Haskell's pattern matching to ual.

### 9.3 Consider as Expression

Consider blocks could be extended to become expressions that return values:

```lua
-- Hypothetical future syntax for consider expressions
message = value.consider {
  if_equal(0) { return "Zero" }
  if_match(is_positive) { return "Positive" }
  if_else { return "Negative" }
}
```

This would align more closely with match expressions in functional languages, enabling more concise code in some scenarios.

### 9.4 Exhaustiveness Checking

A more sophisticated type system could enable exhaustiveness checking:

```lua
-- Hypothetical future syntax with exhaustiveness warnings
-- @exhaustive
value.consider {
  if_equal(0) { ... }
  if_match(is_positive) { ... }
  -- Warning: pattern is non-exhaustive, missing case for negative numbers
}
```

This would provide stronger safety guarantees similar to Rust's match expressions.

## 10. Implementation Path and Migration Strategy

### 10.1 Implementation Phases

The implementation could be phased:

1. **Core Patterns**: Implement `if_equal`, `if_match`, `if_type`, and `if_else`
2. **Existing Integration**: Maintain backward compatibility with `if_ok`/`if_err`
3. **Optimization Pass**: Add compiler optimizations for common patterns
4. **Extensions**: Add higher-level patterns as the language evolves

This phased approach allows for incremental adoption and testing.

### 10.2 Backward Compatibility

The generalized consider construct maintains backward compatibility with the original result handling pattern:

```lua
-- Existing code continues to work
result.consider {
  if_ok { process_result(_1) }
  if_err { handle_error(_1) }
}
```

This ensures that existing code will work unchanged with the new implementation.

### 10.3 Migration Guidance

For users migrating to the new consider patterns, guidance would include:

1. **Replace Nested If Statements**: Complex conditionals are often clearer as consider blocks
2. **Replace Type Checking Chains**: Series of type checks can be unified into a single consider block
3. **Simplify State Machines**: State/event handling becomes more concise with consider
4. **Maintain Performance**: Consider blocks compile to efficient code, similar to switch statements

## 11. Conclusion

The generalized `.consider{}` construct extends ual's pattern matching capabilities beyond error handling, creating a powerful, expressive mechanism for conditional logic that aligns with the language's philosophy of explicit, safe operations. By drawing inspiration from functional languages' pattern matching, Ruby's code blocks, and Rust's match expressions, while maintaining ual's unique stack-centric perspective, this proposal enriches the language with a construct that is both powerful and approachable.

The `.consider{}` pattern complements ual's existing features like stack perspectives, segment borrowing, and error handling, creating a cohesive whole that enables more declarative, intention-revealing code. The explicit, non-fallthrough design makes code clearer and less error-prone, while the extensible pattern system allows for future growth and specialization.

This proposal represents a significant step in ual's evolution, adding sophisticated pattern matching capabilities while maintaining the language's commitment to explicitness, safety, and efficiency in embedded systems programming.