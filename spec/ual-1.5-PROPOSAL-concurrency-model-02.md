# ual 1.6 PROPOSAL: Concurrency Model (Part 2)

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal is targeting to be integrated into the main ual spec in a forthcoming release.

---

## 1. Introduction

This document extends the "Stack Perspectives for Concurrency and Algorithms" proposal by detailing ual's comprehensive concurrency model. While Part 1 focused on stack perspectives for FIFO/LIFO behavior, this proposal explores the complete concurrency mechanism centered around the @spawn stack and concurrent task management.

The proposed concurrency model builds upon ual's container-centric philosophy to create a unified approach to concurrent programming where tasks, messaging, and synchronization are handled through specialized stacks with explicit operations.

## 2. The @spawn Stack as Task Registry

### 2.1 Fundamental Concept: Persistent Task Registry

The @spawn stack serves as both an execution initiator and a persistent registry of running tasks. Unlike traditional stacks where values are typically popped and consumed, the @spawn stack maintains its elements (tasks) until their natural completion:

```lua
// Spawn a background task
@spawn: function() {
  // Long-running operation
  while_true(true)
    @results: push(collect_data())
    sleep(1000)
  end_while_true
}

// The task remains on the @spawn stack while executing
```

This persistence is fundamentally different from other stacks and represents a key innovation in ual's concurrency model.

### 2.2 Task Lifecycle on the @spawn Stack

When a function is pushed to the @spawn stack, it undergoes a specific lifecycle:

1. **Creation**: The function is pushed onto the @spawn stack
2. **Execution**: The function begins executing concurrently 
3. **Persistence**: The function remains on the stack during execution
4. **Completion**: When the function finishes, it is automatically removed from the stack

This lifecycle creates a natural, stack-based representation of system activity where the @spawn stack's contents directly reflect currently executing concurrent work.

### 2.3 Constrained Operations on @spawn Stack

Due to its special nature as a task registry, the @spawn stack has constrained operations compared to normal stacks:

| Operation | Supported? | Semantics |
|-----------|------------|-----------|
| push      | Yes        | Start a new concurrent task |
| pop       | No         | Not supported - tasks must complete naturally |
| peek      | Yes        | Get reference to most recently spawned task |
| depth     | Yes        | Get number of currently running tasks |
| wait      | Yes        | Wait for specific task or all tasks to complete |

The prohibition on popping tasks is intentional and critical:
- Tasks represent actual execution contexts with resources and state
- Forcibly removing them could lead to resource leaks and inconsistent state
- The system manages task lifecycle based on natural completion or explicit termination

## 3. Stack References as First-Class Values

### 3.1 Stack Reference Creation and Usage

To enable flexible communication between concurrent tasks, stack references can be stored in variables and passed as parameters:

```lua
// Create typed stacks
@Stack.new(Integer, Shared): alias:"data_stack"
@Stack.new(String, Shared): alias:"log_stack"

// Create references to these stacks
data_channel = @data_stack  // Variable holding stack reference
log_channel = @log_stack    // Another stack reference

// Pass references to concurrent task
@spawn: function(results, logs) {
  // Calculate result
  result = complex_calculation()
  
  // Send results through passed stack references
  @results: push(result)
  @logs: push("Calculation complete")
}(data_channel, log_channel)
```

### 3.2 Stack Reference Type Safety

Stack references maintain strict type compatibility with their source stacks:

```lua
@Stack.new(Integer): alias:"int_stack"
@Stack.new(String): alias:"str_stack"

int_ref = @int_stack
str_ref = @str_stack

// Type checking ensures references maintain types
@int_ref: push(42)       // Valid: Integer into Integer stack
@str_ref: push("hello")  // Valid: String into String stack
@int_ref: push("hello")  // Error: String cannot go into Integer stack
```

This type safety extends across concurrent boundaries, ensuring that even when stacks are shared between tasks, type correctness is maintained.

### 3.3 Stack Reference Ownership and Lifetime

Stack references follow these ownership rules:

1. **Non-Ownership**: References don't own the underlying stack
2. **Safe Sharing**: Multiple references to the same stack can exist
3. **Stack Lifetime**: References become invalid if the original stack goes out of scope
4. **Perspective Independence**: Each reference can have its own perspective on the stack

The compiler ensures that stack references don't outlive their source stacks, preventing dangling reference errors.

## 4. @spawn Function Constraints

### 4.1 Return Statement Prohibition

Functions executed through the @spawn stack cannot use return statements:

```lua
// Invalid: Cannot use return in spawned function
@spawn: function() {
  result = calculate()
  return result  // Compile error: Return not allowed in spawned function
}

// Valid: Use stack-based communication instead
@spawn: function(results) {
  result = calculate()
  @results: push(result)  // Communicate via stacks
}(result_stack)
```

This constraint is essential because:

1. **Execution Context**: The caller continues executing immediately after spawning, so there's no context to receive a return value
2. **Asynchronous Nature**: Results may be produced long after the spawning function has moved on
3. **Stack-Based Communication**: ual's paradigm encourages explicit stack-based data flow

### 4.2 Implicit @spawn Annotation

The @spawn stack selector before a function declaration serves as an explicit annotation that the function must be executed concurrently and cannot use returns:

```lua
// Explicit annotation that function runs concurrently
@spawn: function worker(data_channel, result_channel) {
  // Must use stack-based communication
  // Cannot use return statements
}
```

The compiler enforces these constraints at compile time, catching potential concurrency errors before runtime.

### 4.3 Error Handling Integration

Spawned functions integrate naturally with ual's error stack mechanism:

```lua
@spawn: function() {
  @error: function process() {
    // Operation that may generate errors
    file = open("data.txt")
    if file == nil then
      @error: push("Could not open file")
      return
    end
    // Continue processing
  }
  
  // Handle errors from this task
  if @error: depth() > 0 then
    err = @error: pop()
    @log: push("Task error: " .. err)
  end
}
```

This integration allows concurrent tasks to handle their own errors without affecting the main execution flow.

## 5. Task Management Operations

### 5.1 Task References

Unlike traditional values, tasks on the @spawn stack can be referenced while remaining on the stack:

```lua
// Spawn a task and get a reference to it
@spawn: function() { background_work() }
background_task = @spawn: peek()  // Reference to most recently spawned task
```

Task references enable management operations without removing the task from execution.

### 5.2 Task Status Inspection

The status of tasks can be inspected through the @spawn stack:

```lua
// Check if task is still running
if @spawn: is_active(task) then
  // Task is still executing
end

// Get information about a task
info = @spawn: info(task)
fmt.Printf("Task running for: %d ms\n", info.runtime)
```

These operations enable monitoring concurrent task state without interrupting execution.

### 5.3 Task Control

Tasks can be controlled through operations on the @spawn stack:

```lua
// Wait for a specific task to complete
@spawn: wait(task)

// Wait for all tasks to complete
@spawn: wait_all()

// Wait with timeout
success = @spawn: wait_timeout(task, 1000)  // 1 second timeout

// Terminate a task
@spawn: terminate(task)

// Pause/resume a task (if supported by runtime)
@spawn: pause(task)
@spawn: resume(task)
```

These control operations allow for coordinated concurrent execution without requiring complex synchronization primitives.

### 5.4 Task Prioritization

On platforms that support it, tasks can be prioritized:

```lua
// Set task priority
@spawn: priority(task, "high")
@spawn: priority(task, "low")

// Get current priority
current_priority = @spawn: get_priority(task)
```

This allows for resource allocation optimization based on task importance.

## 6. Communication Patterns

### 6.1 Channel-like Stacks

With the stack perspective operations from Part 1, stacks can function as FIFO channels for orderly communication:

```lua
// Create a channel-like stack
@Stack.new(Message, FIFO, Shared): alias:"channel"

// Producer
@spawn: function() {
  for i = 1, 10 do
    @channel: push(create_message(i))
    sleep(100)  // Some delay
  end
}

// Consumer
@spawn: function() {
  while_true(true)
    if channel.depth() > 0 then
      message = channel.pop()
      process_message(message)
    end
    sleep(50)  // Check periodically
  end_while_true
}
```

The FIFO perspective ensures messages are processed in sending order, creating a natural queue behavior.

### 6.2 Control Channels

Dedicated control channels enable task management without direct process manipulation:

```lua
// Create data and control channels
@Stack.new(Data, FIFO, Shared): alias:"data"
@Stack.new(Control, FIFO, Shared): alias:"control"

// Data collection task
@spawn: function(data_channel, control_channel) {
  running = true
  
  while_true(running)
    // Check for control messages
    if control_channel.depth() > 0 then
      command = control_channel.pop()
      if command == "STOP" then
        running = false
        break
      end
    end
    
    // Collect and send data
    @data_channel: push(collect_data())
    sleep(1000)
  end_while_true
  
  // Final notification
  @data_channel: push("COLLECTION_COMPLETE")
}(data, control)

// Later, signal task to stop
@control: push("STOP")
```

This pattern provides a clean separation between data flow and control flow, making task management more explicit.

### 6.3 Structured Message Passing

For more complex communication needs, structured messages can be used:

```lua
// Message format
message = {
  type = "COMMAND",
  action = "PROCESS",
  params = {item_id = 42, priority = "high"},
  callback = task_id
}

// Send complex message
@channel: push(message)

// Receive and process based on message type
msg = channel.pop()
if msg.type == "COMMAND" then
  handle_command(msg)
elseif msg.type == "NOTIFICATION" then
  handle_notification(msg)
end
```

This approach enables rich communication protocols while maintaining stack-based data flow.

## 7. Advanced Concurrency Patterns

### 7.1 Worker Pool Pattern

Multiple worker tasks can process items from a shared queue:

```lua
function worker_pool(items, worker_count)
  // Create communication channels
  @Stack.new(Task, FIFO, Shared): alias:"tasks"
  @Stack.new(Result, FIFO, Shared): alias:"results"
  
  // Get references for passing to workers
  task_queue = @tasks
  result_queue = @results
  
  // Push all work to the task queue
  for i = 1, #items do
    @tasks: push(items[i])
  end
  
  // Create termination signals (one per worker)
  for i = 1, worker_count do
    @tasks: push("TERMINATE")
  end
  
  // Spawn worker tasks
  for i = 1, worker_count do
    @spawn: function(id, task_queue, result_queue) {
      while_true(true)
        // Get next task
        task = task_queue.pop()
        
        // Check for termination
        if task == "TERMINATE" then
          break
        end
        
        // Process task
        result = process_task(task)
        
        // Send result
        @result_queue: push({
          task_id = task,
          worker_id = id,
          result = result
        })
      end_while_true
    }(i, task_queue, result_queue)
  end
  
  // Return the result queue for the caller to consume
  return results
end
```

This pattern distributes work across multiple concurrent workers while maintaining orderly task processing.

### 7.2 Pipeline Processing

Multi-stage processing can be implemented through a pipeline of tasks:

```lua
function setup_pipeline(input_data)
  // Create pipeline stages with different perspectives
  @Stack.new(Data, FIFO, Shared): alias:"raw_data"     // FIFO for ordered processing
  @Stack.new(Data, FIFO, Shared): alias:"parsed_data"
  @Stack.new(Data, FIFO, Shared): alias:"processed_data"
  
  // Get references for passing to pipeline stages
  stage1_out = @raw_data
  stage2_in = @raw_data
  stage2_out = @parsed_data
  stage3_in = @parsed_data
  stage3_out = @processed_data
  
  // Initial data loading
  for i = 1, #input_data do
    @raw_data: push(input_data[i])
  end
  
  // Add end marker
  @raw_data: push("END_OF_DATA")
  
  // Pipeline stage 1: Data preparation
  @spawn: function(output) {
    // Stage 1 processing
    while_true(true)
      item = output.pop()
      if item == "END_OF_DATA" then
        @output: push(item)  // Forward end marker
        break
      end
      @output: push(prepare_data(item))
    end_while_true
  }(stage1_out)
  
  // Pipeline stage 2: Parsing
  @spawn: function(input, output) {
    // Stage 2 processing
    while_true(true)
      item = input.pop()
      if item == "END_OF_DATA" then
        @output: push(item)  // Forward end marker
        break
      end
      @output: push(parse_data(item))
    end_while_true
  }(stage2_in, stage2_out)
  
  // Pipeline stage 3: Final processing
  @spawn: function(input, output) {
    // Stage 3 processing
    while_true(true)
      item = input.pop()
      if item == "END_OF_DATA" then
        break
      end
      @output: push(process_data(item))
    end_while_true
  }(stage3_in, stage3_out)
  
  // Return final output stack
  return processed_data
end
```

This pattern enables parallel processing while maintaining data ordering and dependencies between stages.

### 7.3 Event Notification System

Events can be broadcast to multiple listeners:

```lua
function setup_event_system()
  // Create event channel
  @Stack.new(Event, FIFO, Shared): alias:"events"
  
  // Create listener registry
  listeners = {}
  
  // Create event dispatcher
  @spawn: function(event_channel, listeners) {
    while_true(true)
      // Wait for event
      event = event_channel.pop()
      
      // Notify all listeners
      for i = 1, #listeners do
        @listeners[i]: push(event)
      end
    end_while_true
  }(events, listeners)
  
  // Functions to manage event system
  return {
    publish = function(event) {
      @events: push(event)
    },
    
    subscribe = function() {
      // Create listener queue
      @Stack.new(Event, FIFO, Shared): alias:"listener"
      table.insert(listeners, @listener)
      return listener
    }
  }
end

// Usage
event_system = setup_event_system()

// Publisher
@spawn: function() {
  event_system.publish({type = "STATUS_CHANGE", status = "active"})
}

// Subscriber
@spawn: function() {
  my_events = event_system.subscribe()
  while_true(true)
    event = my_events.pop()
    handle_event(event)
  end_while_true
}
```

This pattern enables decoupled communication between components through a shared event distribution system.

## 8. Implementation Considerations

### 8.1 Task Lifecycle Management

The lifecycle of tasks on the @spawn stack must be carefully managed:

```
// Pseudocode for @spawn stack task management
func spawn_task(func, args...) {
    task = create_task(func, args)
    register_task(task)  // Add to internal registry
    start_task(task)     // Begin execution
    push_to_spawn_stack(task)  // Add reference to @spawn stack
    
    // Set up automatic completion handling
    on_task_completion(task, func() {
        remove_from_spawn_stack(task)  // Auto-remove on completion
        cleanup_task_resources(task)
    })
}
```

This ensures tasks are properly tracked, executed, and cleaned up without manual management.

### 8.2 Task Preemption and Scheduling

The implementation must handle preemption and scheduling based on target platform capabilities:

1. **Preemptive Systems**: True parallel execution with OS-level threads
2. **Cooperative Systems**: Task switching at yield points for single-threaded environments
3. **Hybrid Systems**: Mix of cooperative and preemptive scheduling

The @spawn stack implementation adapts to these different execution models while maintaining consistent semantics.

### 8.3 Initialization and Stack Size

When a stack is passed between components, its initialization state must be considered:

```lua
function worker(task_queue, result_queue) {
  // Operations on passed stacks
}
```

Rules for stack reference passing:
1. **No Implicit Creation**: Stack references must point to already created stacks
2. **Parameter Validation**: Compiler verifies stack types match expected types
3. **Size Constraints**: Stacks have implementation-defined depth limits
4. **Overflow Handling**: Implementations must define behavior for stack overflow

These considerations ensure reliable cross-component communication.

## 9. Comparison with Other Languages

### 9.1 vs. Goroutines and Channels

Go's approach to concurrency:
```go
// Go - Explicit goroutines and channels
go func() {
    // Concurrent code
}()

ch := make(chan int)
ch <- 42       // Send
value := <-ch  // Receive
```

ual's approach:
```lua
// ual - @spawn and FIFO stacks
@spawn: function() {
  // Concurrent code
}

@Stack.new(Integer, FIFO, Shared): alias:"channel"
@channel: push(42)     // Send
value = channel.pop()  // Receive
```

Key differences:
1. **Unified Model**: ual's stack-based approach unifies concurrent execution and communication
2. **Explicit Task Registry**: The @spawn stack provides visibility into running tasks
3. **Constrained Functions**: ual enforces communication patterns through compiler constraints
4. **Perspective-Based Channels**: ual uses stack perspectives rather than separate channel types

### 9.2 vs. Erlang's Processes and Mailboxes

Erlang's approach to concurrency:
```erlang
% Erlang - Process spawning and message passing
Pid = spawn(fun() -> loop() end).
Pid ! {message, 42}.  % Send
receive
  {message, Value} -> handle(Value)
end.
```

ual's approach:
```lua
// ual - @spawn and stack-based messaging
@spawn: function() {
  // Concurrent code
}
task = @spawn: peek()  // Get reference

@Stack.new(Message, FIFO, Shared): alias:"mailbox"
@mailbox: push({type = "message", value = 42})  // Send
message = mailbox.pop()  // Receive
```

Key differences:
1. **Process Model**: Erlang uses a pure actor model with independent processes
2. **Message Patterns**: Erlang uses pattern-matching for message handling
3. **Selective Receive**: Erlang allows selecting specific messages
4. **Stack Persistence**: ual's message stacks persist messages until explicitly consumed

### 9.3 vs. C#/Java Thread Pools

C#/Java thread pool approach:
```csharp
// C# - Thread pool and task system
ThreadPool.QueueUserWorkItem(state => {
    // Background work
});

var task = Task.Run(() => {
    // Async work
    return result;
});
var result = await task;
```

ual's approach:
```lua
// ual - @spawn and result stacks
@Stack.new(Result, FIFO, Shared): alias:"results"
result_channel = @results

@spawn: function(results) {
  // Background work
  @results: push(computed_result)
}(result_channel)

// Later
result = results.pop()  // Waits for result
```

Key differences:
1. **Return Values**: C#/Java use futures/promises for return values; ual uses explicit stacks
2. **Task Management**: ual provides direct access to the task registry
3. **Awaiting**: ual uses stack operations rather than special await syntax
4. **Explicit Communication**: ual makes communication channels explicit

## 10. Conclusion

The concurrency model proposed for ual represents a natural extension of its stack-based paradigm to concurrent programming. By using the @spawn stack as both an execution initiator and task registry, and leveraging typed stacks with perspective operations for communication, ual creates a unified approach to concurrency that maintains conceptual integrity while providing powerful capabilities.

The key innovations in this design include:
1. **@spawn Stack as Task Registry**: Maintaining task presence during execution
2. **Stack References as Channels**: Enabling flexible communication patterns
3. **Function Constraints**: Ensuring appropriate concurrent communication patterns
4. **Task Management Operations**: Providing visibility and control over concurrent execution
5. **Rich Communication Patterns**: Supporting sophisticated messaging while maintaining the stack paradigm

This approach demonstrates how ual's consistent container-centric philosophy can extend to concurrent programming without introducing disconnected concepts. The result is a concurrency model that feels like a natural part of the language rather than a bolted-on feature, making concurrent programming in ual both powerful and approachable.