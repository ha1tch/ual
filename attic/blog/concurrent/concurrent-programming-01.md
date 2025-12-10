# Concurrency in ual: A Stack-Based Approach to Parallel Programming

## Introduction

If you've worked with concurrency in languages like Go or Erlang, you're familiar with the challenge of coordinating multiple execution flows. Go introduces goroutines and channels; Erlang offers processes and message passing; Rust provides threads and carefully managed shared state. Each presents its own mental model for concurrent programming.

ual takes a different approach, leveraging its stack-based paradigm to create a unified concurrency model that feels natural within the language rather than being a separate subsystem. This document explores how ual's approach differs from traditional concurrency models and how it can simplify concurrent programming.

## A Conceptual Shift: Containers as Communicators

### Traditional Concurrency Models

Most mainstream concurrency models separate execution units from communication mechanisms:

- **Go**: Goroutines (execution) communicate through channels (separate data structures)
- **Erlang**: Processes (execution) communicate through mailboxes (separate message queues)
- **Java/C#**: Threads (execution) communicate through shared memory or queue objects

This separation creates a conceptual split between "things that run" and "things that communicate."

### ual's Unified Container Approach

In ual, both execution and communication occur through the same fundamental abstraction: stacks.

- **@spawn stack**: Manages concurrent tasks (execution)
- **Regular stacks with perspectives**: Enable communication between tasks

This unification eliminates the conceptual divide between execution and communication, creating a more cohesive mental model.

## Stack Perspectives: One Container, Multiple Behaviors

### The Dual Nature of Communication

In concurrent programming, we often need two patterns of data exchange:
- **LIFO (stack)**: Last-in, first-out access for depth-first algorithms and nested operations
- **FIFO (queue)**: First-in, first-out access for ordered message processing and breadth-first algorithms

Most languages force you to choose different data structures for these patterns. ual instead introduces the concept of "stack perspectives" – viewpoints that change how you interact with a stack without changing its underlying structure.

### Perspective Operations

Three simple operations change how you view a stack:

```lua
@stack: lifo  // Traditional stack behavior (default)
@stack: fifo  // Queue-like behavior
@stack: flip  // Toggle between perspectives
```

These operations affect only the selector's view of the stack, not the stack itself. This means:

1. Different parts of a program can have different perspectives on the same stack
2. Changing perspective is an O(1) operation regardless of stack size
3. Perspective changes are thread-safe since they're properties of the selector, not the stack

## The @spawn Stack: Making Concurrency Visible

### Concurrency as an Explicit Registry

Unlike traditional concurrency models where running tasks are largely invisible, ual makes concurrent execution explicit through the @spawn stack:

```lua
// Start a background task
@spawn: function() {
  // Long-running operation
}
```

What makes this unique is that the task remains on the @spawn stack throughout its execution. This creates a visible registry of all running tasks, making the concurrent state of the program directly observable.

### Task Lifecycle and Management

The @spawn stack serves as both an initiator and registry:

1. **Creation**: Task is pushed onto @spawn stack
2. **Execution**: Task runs concurrently
3. **Persistence**: Task remains on stack during execution
4. **Completion**: Task is automatically removed upon completion

This approach makes concurrency management more transparent. You can inspect the @spawn stack to see all running tasks, wait for specific tasks to complete, or manage task priorities.

## Communication Patterns in ual

### Stack-Based Message Passing

In ual, communication between concurrent tasks happens through shared stacks with perspectives:

```lua
// Create a communication channel
@Stack.new(Message, Shared): alias:"channel"

// Producer (using FIFO perspective for ordered sending)
@channel: fifo
@channel: push(create_message())

// Consumer 
message = channel.pop()  // Receive in sending order
```

The FIFO perspective ensures messages are processed in the order they were sent, creating a natural queue behavior without introducing a separate queue type.

### The Philosophical Shift

This approach embodies a key philosophical shift: instead of building separate abstractions for different communication patterns, ual recognizes that these patterns are viewpoints on the same underlying concept.

In essence, ual treats a "channel" not as a fundamentally different thing from a "stack," but as a different perspective on the same container. This insight simplifies the language while expanding its expressive power.

## Comparing with Familiar Concurrency Models

### Go's Goroutines and Channels

Go separated execution and communication:

```go
// Go
go func() {
    // Concurrent code
}()

ch := make(chan int)
ch <- 42       // Send
value := <-ch  // Receive
```

In ual, both use the stack abstraction:

```lua
// ual
@spawn: function() {
  // Concurrent code
}

@Stack.new(Integer, Shared): alias:"channel"
@channel: fifo
@channel: push(42)     // Send
value = channel.pop()  // Receive
```

The unified abstraction reduces the conceptual load, as you're working with variations of the same mechanism rather than entirely different concepts.

### Erlang's Actor Model

Erlang's actor model centers around processes with mailboxes:

```erlang
% Erlang
Pid = spawn(fun() -> loop() end).
Pid ! {message, 42}.  % Send

receive
  {message, Value} -> handle(Value)
end.
```

In ual, the stack itself becomes the mailbox:

```lua
// ual
@spawn: function(inbox) {
  while_true(true)
    message = inbox.pop()
    handle_message(message)
  end_while_true
}(message_stack)

// Send message
@message_stack: fifo
@message_stack: push({type = "message", value = 42})
```

While conceptually similar, ual's approach integrates messaging with its core stack abstraction rather than introducing a separate mailbox concept.

## Advanced Concurrency Patterns Made Simple

### Worker Pool Pattern

Worker pools distribute tasks among multiple concurrent workers:

```lua
function worker_pool(items, worker_count)
  // Create communication channels
  @Stack.new(Task, FIFO, Shared): alias:"tasks"
  @Stack.new(Result, FIFO, Shared): alias:"results"
  
  // Push all work to the task queue
  for i = 1, #items do
    @tasks: push(items[i])
  end
  
  // Spawn worker tasks
  for i = 1, worker_count do
    @spawn: function(task_queue, result_queue) {
      while_true(task_queue.depth() > 0)
        // Get next task
        task = task_queue.pop()
        
        // Process task
        result = process_task(task)
        
        // Send result
        @result_queue: push(result)
      end_while_true
    }(tasks, results)
  end
  
  // Return the result queue for the caller to consume
  return results
end
```

This pattern would require separate thread/process creation and queue/channel management in most languages. In ual, it's all expressed through the same stack abstraction.

### Pipeline Processing

Multi-stage processing pipelines are another common pattern:

```lua
// Create pipeline stages
@Stack.new(Data, FIFO, Shared): alias:"raw_data"
@Stack.new(Data, FIFO, Shared): alias:"parsed_data"
@Stack.new(Data, FIFO, Shared): alias:"processed_data"

// Stage 1: Data preparation
@spawn: function(output) {
  // Prepare data and push to output
}(raw_data)

// Stage 2: Parsing
@spawn: function(input, output) {
  // Parse data from input and push to output
}(raw_data, parsed_data)

// Stage 3: Final processing
@spawn: function(input, output) {
  // Process data from input and push to output
}(parsed_data, processed_data)
```

The FIFO perspective ensures data flows through the pipeline in the correct order, creating a natural flow between stages.

## Conceptual Simplifications in ual

### Unifying Stacks and Channels

ual's approach eliminates the artificial boundary between stacks and channels. In essence, it recognizes that these are the same data structure viewed from different perspectives:

- **Stack**: A container where we add and remove from the same end
- **Channel/Queue**: A container where we add and remove from opposite ends

By making this perspective explicit rather than requiring different data structures, ual reduces the conceptual complexity of concurrent programming.

### Making Concurrency Visible

Traditional concurrency models often treat running tasks as hidden implementation details. ual's @spawn stack makes them visible first-class entities, creating a more transparent model of concurrent execution.

### Localizing Perspective Changes

By making perspective a property of the selector rather than the stack itself, ual creates a more modular approach to concurrent patterns. Different tasks can have different perspectives on the same stack without interfering with each other.

## Philosophical Underpinnings

### The Relational View of Concurrency

ual's concurrency model embodies a relational philosophy – the idea that computation is fundamentally about relationships between entities rather than the entities themselves.

In this view, concurrent tasks are not isolated actors but participants in a shared computational space, connected through explicit relationships (stacks). This aligns with how we naturally think about concurrent systems as interacting components rather than isolated processes.

### The Contextual Nature of Data Flow

The stack perspective concept emphasizes that "first" and "last" are contextual viewpoints rather than absolute properties. This philosophical insight – that order is relative to perspective – simplifies concurrent programming by eliminating the need for different data structures to represent different access patterns.

## When to Use ual's Concurrency Model

ual's approach to concurrency is particularly well-suited for:

1. **Systems with clear data flows**: Where information moves through defined stages or pipelines
2. **Resource-constrained environments**: The unified model is more efficient than maintaining separate concurrency primitives
3. **Mixed algorithmic patterns**: When you need both LIFO and FIFO patterns in the same application
4. **Transparent concurrency**: When you want concurrent activity to be visible and traceable

## Conclusion

ual's approach to concurrency represents a conceptual simplification rather than a technical complication. By extending its stack-based paradigm to encompass concurrent execution and communication, ual creates a unified model that feels natural and cohesive.

This approach offers several advantages:

1. **Reduced cognitive load**: One fundamental abstraction (stacks) rather than multiple concurrency primitives
2. **Explicit concurrency**: Running tasks are visible first-class entities
3. **Flexible communication**: The same data structure can serve multiple access patterns
4. **Transparent relationships**: The connections between concurrent components are explicit and observable

By recognizing that various seemingly distinct concurrency concepts are actually different perspectives on the same underlying idea, ual simplifies concurrent programming without sacrificing expressive power. This philosophical insight – that viewpoint matters more than intrinsic structure – offers a fresh approach to one of programming's most challenging domains.
