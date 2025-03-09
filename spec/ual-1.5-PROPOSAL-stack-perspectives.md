# ual 1.5 PROPOSAL: Stack Perspectives for Concurrency and Algorithms

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the version that the proposal is targeting to be integrated into the main ual spec in a forthcoming release.

---

## 1. Introduction

This proposal refines ual's container-centric approach by introducing stack perspectives, a mechanism for contextually altering how operations interact with stacks without changing the underlying data structure. The refinements focus on three key areas:

1. **Selector-Based Perspectives**: Enabling different views of the same stack through selector operations.
2. **FIFO/LIFO Access Patterns**: Allowing stacks to function as both traditional stacks and channel-like queues.
3. **Pattern-Based Toggling**: Supporting algorithmic patterns requiring alternating access directions.

These refinements maintain ual's design philosophy of explicit operations and minimalist syntax while enabling sophisticated concurrent programming patterns and elegant algorithmic solutions.

## 2. Background and Motivation

### 2.1 Stack Access Patterns

Both LIFO (stack) and FIFO (queue) access patterns are fundamental to computing:

- **LIFO**: Native to call stacks, expression evaluation, and depth-first traversals
- **FIFO**: Critical for message passing, breadth-first algorithms, and orderly processing

Traditional approaches require separate data structures for these patterns, increasing language complexity and forcing developers to choose a structure based on access pattern rather than conceptual fit.

### 2.2 Concurrent Communication Needs

Concurrent systems typically require channel-like FIFO semantics for message passing, but traditional stack-based languages struggle to provide this without introducing distinct channel types disconnected from their core paradigm.

### 2.3 Algorithmic Pattern Challenges

Many algorithms require alternating between different access patterns or operating on both ends of a data structure. Traditional implementations involve complex logic with multiple pointers or separate data structures.

### 2.4 Design Goals

This proposal adheres to the following principles:

1. **Unified Container Model**: Maintain the stack as ual's fundamental container.
2. **Perspective, Not Structure**: Separate access pattern from underlying data structure.
3. **Minimal Additions**: Achieve maximum capability with minimal new language elements.
4. **Explicit Operations**: Make access patterns visibly clear in the code.
5. **Zero Physical Reorganization**: Never physically reorder stack elements.

## 3. Proposed Stack Perspective Operations

### 3.1 Selector-Based Perspective Operations

We propose three core perspective operations for stack selectors:

```lua
@stack: lifo  // Set perspective to Last-In-First-Out (traditional stack)
@stack: fifo  // Set perspective to First-In-First-Out (queue-like)
@stack: flip  // Toggle between current perspective and its opposite
```

These operations affect only how the selector interacts with the stack, not the stack itself. This means:

1. **Localized Change**: Only the current selector's behavior changes
2. **Multiple Perspectives**: Different selectors can have different perspectives on the same stack
3. **Zero Structural Impact**: The stack's physical organization remains unchanged

The default perspective for all selectors is LIFO, matching traditional stack behavior.

### 3.2 Perspective Semantics

The perspective controls how `push` operations interact with the stack:

#### In LIFO perspective (default):
- `push`: Add to top of stack
- `pop`: Remove from top of stack

#### In FIFO perspective:
- `push`: Add to bottom of stack 
- `pop`: Remove from top of stack

The critical insight is that changing only where items are pushed, while always popping from the same end, creates different access patterns without physically reorganizing stack elements.

### 3.3 Operation Properties

The perspective operations have distinct properties:

- `lifo`: **Idempotent** - Setting LIFO perspective multiple times has no additional effect
- `fifo`: **Idempotent** - Setting FIFO perspective multiple times has no additional effect
- `flip`: **Non-idempotent** - Each call toggles the current perspective to its opposite

This combination provides both explicit control (through `lifo`/`fifo`) and efficient toggling (through `flip`).

## 4. Examples and Use Cases

### 4.1 Concurrent Message Passing

```lua
function producer_consumer()
  // Create shared stack for message passing
  @Stack.new(Message, Shared): alias:"channel"
  
  // Producer task
  @spawn: function() {
    @channel: fifo  // Set FIFO perspective for message ordering
    
    for i = 1, 10 do
      @channel: push(create_message(i))
    end
  }
  
  // Consumer task 
  @spawn: function() {
    while_true(channel.depth() > 0)
      message = channel.pop()  // Messages arrive in sending order
      process_message(message)
    end_while_true
  }
end
```

### 4.2 Palindrome Checking Algorithm

```lua
function is_palindrome(word)
  @Stack.new(Char): alias:"chars"
  
  // Push all characters
  for i = 1, #word do
    @chars: push(word:sub(i, i))
  end
  
  // Check from both ends simultaneously
  @chars: fifo
  while chars.depth() > 1 do
    first = chars.pop()  // From beginning (due to FIFO)
    @chars: flip
    last = chars.pop()   // From end (due to LIFO after flip)
    @chars: flip         // Back to FIFO for next iteration
    
    if first != last then return false end
  end
  
  return true
end
```

### 4.3 Adaptive Search Algorithm

```lua
function adaptive_search(graph, start)
  @Stack.new(Node, Shared): alias:"frontier"
  visited = {}
  
  @frontier: push(start)
  visited[start] = true
  
  // Start with breadth-first search
  @frontier: fifo
  
  while frontier.depth() > 0 do
    node = frontier.pop()
    process(node)
    
    // Toggle between BFS and DFS based on conditions
    if should_switch_strategy(node) then
      @frontier: flip
    end
    
    // Add neighbors
    for neighbor in graph.neighbors(node) do
      if not visited[neighbor] then
        @frontier: push(neighbor)
        visited[neighbor] = true
      end
    end
  end
end
```

## 5. Design Rationale

### 5.1 The Power of Perspective vs. Physical Reordering

This proposal recognizes that access pattern is a property of the viewer (selector), not an inherent property of the data structure (stack). This insight leads to several key advantages:

1. **Zero Structural Change**: Perspectives never physically reorder stack elements, eliminating potential race conditions in concurrent contexts.

2. **Multiple Simultaneous Views**: Different parts of a program can have different perspectives on the same stack.

3. **O(1) Operations**: All perspective changes are constant-time operations regardless of stack size.

4. **Thread Safety**: Since perspectives are selector properties, they don't affect shared state.

After careful analysis, we found no legitimate use case where physically reordering a stack provides benefits that can't be achieved more efficiently and safely through perspectives.

### 5.2 Three Operations vs. Fewer

We propose including all three operations (`lifo`, `fifo`, and `flip`) for several reasons:

1. **Complementary Functions**: Each serves a different programming pattern:
   - `lifo`/`fifo`: Provide explicit, self-documenting perspective setting
   - `flip`: Enables concise expression of alternating access patterns

2. **Cognitive Alignment**: `lifo`/`fifo` clearly communicate intent, while `flip` naturally expresses toggling patterns.

3. **Implementation Simplicity**: All three share the same simple implementation (a boolean flag per selector).

4. **Algorithm Elegance**: Certain algorithms (like palindrome checking) express most naturally with `flip`.

5. **Safety with Flexibility**: The idempotent operators (`fifo`/`lifo`) provide safety and predictability, while `flip` offers concise toggling for algorithms that naturally alternate perspectives.

### 5.3 Stack Selectors as the Locus of Perspective

By making perspective a property of the selector rather than the stack, we achieve:

1. **Localized Reasoning**: Perspective changes affect only the current context.

2. **Multi-Threaded Safety**: Different threads can safely use different perspectives on shared stacks.

3. **Minimal State**: Perspectives require just one bit of state per selector rather than per stack.

4. **Stack Purity**: The underlying stack remains conceptually pure - a simple ordered sequence of elements.

This design recognizes that "first" and "last" are relative concepts - matters of viewpoint rather than absolute properties of data.

## 6. Implementation Considerations

### 6.1 Selector Implementation

At the implementation level, perspectives require minimal changes:

```
// Pseudocode for the selector
type StackSelector {
    stack       *Stack
    perspective bool  // false = LIFO, true = FIFO
}

// Operations
func (sel *StackSelector) fifo() {
    sel.perspective = true
}

func (sel *StackSelector) lifo() {
    sel.perspective = false
}

func (sel *StackSelector) flip() {
    sel.perspective = !sel.perspective
}

// Push operation accounts for perspective
func (sel *StackSelector) push(value) {
    if sel.perspective == true {
        // FIFO: push to bottom of stack
        sel.stack.pushToBottom(value)
    } else {
        // LIFO: push to top of stack
        sel.stack.pushToTop(value)
    }
}
```

### 6.2 Efficient Stack Implementation

For efficient implementation of both perspectives, a double-ended queue (deque) structure is optimal, allowing O(1) operations at both ends.

### 6.3 Concurrency Considerations

Since perspectives are properties of selectors, not stacks, they introduce no new concurrency concerns. Shared stacks still require appropriate synchronization, but perspective changes are local to each selector.

## 7. Comparison with Other Languages

### 7.1 Go's Channels vs. ual's Perspective Stacks

Go provides separate constructs for stacks and channels:

```go
// Go - Separate concepts
stack := []int{}
stack = append(stack, 42)         // Push
value := stack[len(stack)-1]      // Pop

ch := make(chan int)
ch <- 42                         // Send
value := <-ch                    // Receive
```

ual unifies these with perspective:

```lua
// ual - Unified concept with perspectives
@stack: push(42)
value = stack.pop()

@stack: fifo  // Now becomes channel-like
@stack: push(42)  // Send
value = stack.pop()  // Receive
```

### 7.2 Rust's Collections vs. ual's Perspective Stacks

Rust requires different collection types:

```rust
// Rust - Different types
let mut stack = Vec::new();
stack.push(42);
let value = stack.pop().unwrap();

let mut queue = VecDeque::new();
queue.push_back(42);
let value = queue.pop_front().unwrap();
```

ual uses one container with perspectives:

```lua
// ual - One container, different perspectives
@stack: lifo  // Explicit stack behavior
@stack: push(42)
value = stack.pop()

@stack: fifo  // Queue behavior
@stack: push(42)
value = stack.pop()
```

## 8. Future Directions

While maintaining our commitment to minimalism, future proposals might explore:

1. **Additional Perspective Operations**: Beyond FIFO/LIFO for specialized algorithms
2. **Composite Perspectives**: For multi-phase algorithm patterns 
3. **Stack References with Perspective**: Supporting perspectives in passed stack references

These would be considered only if they demonstrate clear value while maintaining conceptual integrity.

## 9. Conclusion

The proposed stack perspective operations extend ual's stack-based paradigm to elegantly encompass both traditional LIFO stack patterns and FIFO queue/channel patterns without introducing separate container types. By recognizing that access pattern is a property of the viewer rather than the data, we achieve maximum flexibility with minimal language additions.

This design maintains ual's commitment to explicitness and minimalism while enabling sophisticated concurrent programming patterns and elegant algorithmic solutions. The unification of stacks and channels through the perspective concept demonstrates how careful language design can reduce complexity while increasing expressive power.
