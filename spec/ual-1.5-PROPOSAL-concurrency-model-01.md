# ual 1.6 PROPOSAL: Stack-Based Concurrency Model

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the version that the proposal is targeting to be integrated into the main ual spec in a forthcoming release.

---

## 1. Introduction

This document proposes a concurrency model for the ual programming language that extends its stack-based paradigm to distributed computing environments. The proposed approach leverages ual's existing stack primitives as the foundation for inter-process communication, synchronization, and resource sharing, creating a unified model that maintains language consistency while enabling efficient concurrent programming on resource-constrained systems.

### 1.1 Background

Embedded systems increasingly employ multi-core processors or multi-processor architectures to improve performance while maintaining power efficiency. The ZX Interface project, with its potential for multiple ESP32 microcontrollers, exemplifies this trend toward distributed processing in constrained environments. However, programming such systems traditionally requires complex concurrency primitives that often feel foreign to the base language's paradigm.

### 1.2 Design Philosophy

The proposed concurrency model adheres to the following principles:

1. **Stack-Centric Design**: Extends ual's stack-based paradigm to serve as the foundation for concurrency rather than introducing separate mechanisms.
2. **Minimal Conceptual Surface Area**: Relies on a small set of core concepts that compose to handle complex scenarios.
3. **Explicit Coordination**: Makes communication and synchronization patterns visible in the code.
4. **Zero Overhead When Unused**: Imposes no cost on single-threaded or single-node code.
5. **Resource-Appropriate**: Designed specifically for the memory and processing constraints of embedded systems.

### 1.3 Key Innovation: Stacks as Channels

The central concept of this proposal is the unification of stacks and communication channels. In this model:
- A stack operates as a communication medium between processes or nodes
- Stack operations (`push`, `pop`, etc.) serve as the fundamental message-passing primitives
- Stack types ensure type safety across process or node boundaries

This approach maintains ual's core paradigm while naturally extending it to handle concurrency without introducing foreign concepts.

## 2. Concurrency Models in Resource-Constrained Systems

### 2.1 Comparative Analysis of Existing Approaches

Several concurrency models have been employed in resource-constrained systems, each with distinct advantages and limitations:

#### 2.1.1 Thread-Based Concurrency (C/C++, Java)

**Characteristics:**
- Relies on shared-memory threads with locks, mutexes, and condition variables
- Requires significant runtime support for thread management
- Higher memory footprint due to per-thread stacks

**Limitations for Constrained Systems:**
- Memory overhead for thread stacks limits scalability
- Lock contention and context switching consume precious CPU cycles
- Debugging concurrency issues is notoriously difficult

#### 2.1.2 Actor Model (Erlang, Elixir)

**Characteristics:**
- Processes communicate through message passing with no shared state
- Each actor maintains private state and responds to messages
- Natural fit for distributed systems

**Limitations for Constrained Systems:**
- Message copying can be costly in memory-constrained environments
- Full actor implementations require sophisticated schedulers
- Often relies on garbage collection

#### 2.1.3 CSP Model (Go, occam)

**Characteristics:**
- Communication occurs through channels
- Goroutines or processes are lightweight compared to threads
- Structured concurrency with explicit synchronization

**Limitations for Constrained Systems:**
- Channel implementations still require runtime support
- Goroutines, while lightweight, still consume memory
- Not all embedded platforms can efficiently implement necessary primitives

#### 2.1.4 Event-Loop Model (JavaScript, embedded C)

**Characteristics:**
- Single-threaded execution with event callbacks
- Avoids complexities of thread synchronization
- Common in microcontroller programming

**Limitations for Constrained Systems:**
- Can lead to "callback hell" in complex systems
- Long-running operations block the entire system
- Difficult to take advantage of multiple cores

### 2.2 Why a New Approach is Needed

None of these models fully aligns with ual's stack-based paradigm or optimally addresses the needs of distributed embedded systems like the ZX Interface with multiple ESP32s. Therefore, a new approach is proposed that:

1. Builds naturally on ual's existing stack primitives
2. Enables efficient communication between nodes without excessive overhead
3. Provides synchronization mechanisms that respect resource constraints
4. Maintains conceptual integrity with the rest of the language

## 3. The Stack-as-Channel Concurrency Model

### 3.1 Fundamental Concept

In this model, stacks serve as the primary abstraction for both data storage and inter-process communication. This unification creates a remarkably consistent programming model: the same operations used to manipulate local data are used to communicate between processes or nodes.

### 3.2 Distributed Stack Types

Distributed stacks extend ual's typed stacks to span process or node boundaries:

```lua
-- Create a stack that spans nodes
@Stack.new(Integer, Distributed): alias:"calc_stack"

-- Assign stack endpoints to specific nodes
@calc_stack: bind(1, 2)  -- Connects nodes 1 and 2 via this stack
```

The `Distributed` attribute indicates that operations on this stack may involve communication with other processes or nodes.

### 3.3 Communication Patterns

The model supports various communication patterns through stack configuration:

#### 3.3.1 Point-to-Point (One-to-One)

```lua
@Stack.new(Integer, Distributed): alias:"calc"
@calc: bind(1, 2)  -- Direct connection between nodes 1 and 2
```

On Node 1:
```lua
@calc: push(42)  -- Sends 42 to Node 2
```

On Node 2:
```lua
local value = calc.pop()  -- Receives 42 from Node 1
```

#### 3.3.2 Broadcast (One-to-Many)

```lua
@Stack.new(String, Broadcast): alias:"broadcast"
@broadcast: bind_source(1)        -- Node 1 is sender
@broadcast: bind_targets({2,3,4}) -- Nodes 2,3,4 are receivers
```

On Node 1:
```lua
@broadcast: push("update")  -- Sends to all target nodes
```

On Nodes 2, 3, and 4:
```lua
local msg = broadcast.pop()  -- Each node receives "update"
```

#### 3.3.3 Collection (Many-to-One)

```lua
@Stack.new(Result, Collection): alias:"results"
@results: bind_targets(1)        -- Node 1 receives
@results: bind_sources({2,3,4})  -- Nodes 2,3,4 send
```

#### 3.3.4 Shared Bus (Many-to-Many)

```lua
@Stack.new(Message, Shared): alias:"bus"
@bus: bind_all({1,2,3,4,5})  -- All nodes can send and receive
```

### 3.4 Synchronization Primitives

Synchronization primitives are implemented as specialized stacks with specific behaviors:

#### 3.4.1 Mutex

```lua
@Stack.new(Mutex, Distributed): alias:"resource_mutex"
@resource_mutex: bind_all({1,2,3,4,5})  -- Available to all nodes

-- Acquire the mutex
@resource_mutex: acquire()  -- Blocks until mutex is acquired

-- Critical section
-- ... access shared resource ...

-- Release the mutex
@resource_mutex: release()
```

Under the hood, a mutex is implemented as a stack that contains either 0 or 1 tokens:
- `acquire()` attempts to pop the token (blocking if empty)
- `release()` pushes the token back

#### 3.4.2 Semaphore

```lua
@Stack.new(Semaphore, Distributed): alias:"resource_sem"
@resource_sem: initialize(3)  -- Initialize with 3 tokens

-- Acquire a semaphore slot
@resource_sem: acquire()

-- Use the resource
-- ...

-- Release the semaphore slot
@resource_sem: release()
```

#### 3.4.3 Barrier

```lua
@Stack.new(Barrier, Distributed): alias:"sync_point"
@sync_point: initialize(5)  -- Requires 5 arrivals

-- Reach barrier and wait for others
@sync_point: wait()  -- Blocks until all 5 nodes have called wait()
```

#### 3.4.4 Reader-Writer Lock

```lua
@Stack.new(RWLock, Distributed): alias:"data_lock"

-- Acquire for reading (multiple readers allowed)
@data_lock: read_acquire()
-- Read shared data
@data_lock: read_release()

-- Acquire for writing (exclusive access)
@data_lock: write_acquire()
-- Modify shared data
@data_lock: write_release()
```

### 3.5 Asynchronous Operations

The model includes non-blocking variants of stack operations:

```lua
-- Non-blocking operations
@compute_stack: try_push(value)  -- Returns success/failure instead of blocking
local value, success = compute_stack.try_pop()  -- Non-blocking pop
```

## 4. Implementation on Embedded Distributed Systems

### 4.1 Hardware Abstraction Layer

The concurrency model sits atop a hardware abstraction layer that handles the physical communication between processors:

```
┌─────────────────────────────────────────────┐
│             ual Application                  │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│         Stack-as-Channel Concurrency        │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│        Hardware Communication Layer          │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│          Physical Bus Protocol               │
└─────────────────────────────────────────────┘
```

### 4.2 Physical Communication Options

#### 4.2.1 Shared SPI Bus

A shared SPI bus can connect multiple ESP32 nodes:
- One master node (typically connected to the Z80 bus)
- Multiple slave nodes
- 4-wire connection with chip select lines
- Speeds up to 80MHz

#### 4.2.2 Dedicated UART Links

Direct UART connections between frequently communicating nodes:
- Point-to-point links
- Simple protocol with minimal overhead
- Multiple UART interfaces available on ESP32

#### 4.2.3 Custom Parallel Bus

A parallel data bus for higher throughput:
- 8-bit data lines with control signals
- Address/data multiplexing
- Hardware flow control

### 4.3 Optimized Bus Protocol

An efficient custom protocol inspired by TileLink but simplified for our needs:

```
[2 bytes: Header] [1 byte: Length] [0-255 bytes: Payload] [1 byte: Checksum]

Header format:
- 4 bits: Source Node ID
- 4 bits: Destination Node ID
- 4 bits: Message Type
- 4 bits: Priority/Channel
```

Message types include:
- **GET**: Read operations
- **PUT**: Write operations
- **ATOMIC**: Read-modify-write operations
- **SYNC**: Synchronization operations

### 4.4 Memory Considerations

The implementation carefully manages memory to operate efficiently on constrained systems:

1. **Zero-Copy Where Possible**: Data is transferred directly from source to destination without intermediate copies
2. **Static Allocation**: Communication buffers are pre-allocated at compile time
3. **Pool-Based Management**: Dynamic requirements use fixed memory pools
4. **Explicit Memory Regions**: ual's ownership system helps manage shared memory

## 5. Programming Patterns

### 5.1 Producer-Consumer Pattern

```lua
-- On producer node
function producer()
  while true do
    local data = generate_data()
    @data_queue: push(data)
  end
end

-- On consumer node
function consumer()
  while true do
    local data = data_queue.pop()
    process_data(data)
  end
end
```

### 5.2 Master-Worker Pattern

```lua
-- On master node
function master()
  -- Create distributed stacks for task distribution
  @Stack.new(Task, Distributed): alias:"tasks"
  @Stack.new(Result, Distributed): alias:"results"
  
  -- Distribute tasks
  for i = 1, 100 do
    @tasks: push(create_task(i))
  end
  
  -- Collect results
  local all_results = {}
  for i = 1, 100 do
    all_results[i] = results.pop()
  end
end

-- On worker nodes
function worker()
  @Stack.new(Task, Distributed): alias:"tasks"
  @Stack.new(Result, Distributed): alias:"results"
  
  while true do
    local task = tasks.pop()
    local result = process_task(task)
    @results: push(result)
  end
end
```

### 5.3 Pipeline Pattern

```lua
-- Distributed audio pipeline
@Stack.new(AudioBuffer, Distributed): alias:"raw"
@Stack.new(AudioBuffer, Distributed): alias:"filtered"
@Stack.new(AudioBuffer, Distributed): alias:"compressed"
@Stack.new(AudioBuffer, Distributed): alias:"output"

-- Node roles
@raw: bind(1, 2)              -- Capture → Filter
@filtered: bind(2, 3)         -- Filter → Compression
@compressed: bind(3, 4)       -- Compression → Output
@output: bind(4, 5)           -- Output → Playback

-- Node 1: Audio Capture
function audio_capture()
  while true do
    local buffer = capture_audio()
    @raw: push(buffer)
  end
end

-- Node 2: Audio Filtering
function audio_filter()
  while true do
    local buffer = raw.pop()
    local filtered_buffer = apply_filters(buffer)
    @filtered: push(filtered_buffer)
  end
end
```

### 5.4 Bulk Transfer Pattern

```lua
-- Transfer large amount of data efficiently
function transfer_data(large_data)
  @Stack.new(Chunk, Distributed): alias:"transfer"
  @Stack.new(Mutex, Distributed): alias:"transfer_mutex"
  
  -- Acquire transfer lock
  @transfer_mutex: acquire()
  
  -- Split data into manageable chunks
  local chunks = split_into_chunks(large_data, 1024)
  
  -- Send number of chunks first
  @transfer: push(#chunks)
  
  -- Send all chunks
  for i = 1, #chunks do
    @transfer: push(chunks[i])
  end
  
  -- Release lock
  @transfer_mutex: release()
end
```

## 6. Integration with ual Features

### 6.1 Typed Stacks and Concurrency

The concurrency model builds on ual's typed stack system to ensure type safety across node boundaries:

```lua
-- Type-safe distributed communication
@Stack.new(Integer, Distributed): alias:"ints"
@Stack.new(String, Distributed): alias:"strs"

-- The compiler ensures type safety across nodes
```

### 6.2 Error Handling Integration

The concurrency model integrates with ual's error stack mechanism:

```lua
@error > function try_remote_operation()
  local success = calc_stack.try_push(value)
  if not success then
    @error > push("Remote node unavailable")
    return nil
  end
  return true
end
```

### 6.3 Ownership System

The proposed concurrency model would work with ual's ownership system to manage distributed resources:

```lua
-- Create a resource with distributed ownership
@Stack.new(Resource, Distributed, Owned): alias:"shared_resource"

-- Transfer ownership between nodes
@shared_resource: transfer_ownership(node2)
```

## 7. Examples

### 7.1 Image Processing on Multi-ESP32 ZX Interface

This example demonstrates parallel image processing across multiple ESP32 nodes:

```lua
-- On coordinator node
function process_image(image_data)
  -- Split image into chunks
  local chunks = split_image(image_data, 4)
  
  -- Distribute chunks to worker nodes
  @Stack.new(ImageChunk, Distributed): alias:"chunks"
  @Stack.new(ImageChunk, Distributed): alias:"results"
  
  @chunks: bind_source(1)         -- Coordinator sends
  @chunks: bind_targets({2,3,4,5}) -- Workers receive
  
  @results: bind_targets(1)        -- Coordinator receives
  @results: bind_sources({2,3,4,5}) -- Workers send
  
  -- Send chunks to workers
  for i = 1, #chunks do
    @chunks: push(chunks[i])
  end
  
  -- Collect results
  local processed_chunks = {}
  for i = 1, #chunks do
    processed_chunks[i] = results.pop()
  end
  
  -- Reassemble image
  return combine_image(processed_chunks)
end

-- On worker nodes
function worker_process()
  @Stack.new(ImageChunk, Distributed): alias:"chunks"
  @Stack.new(ImageChunk, Distributed): alias:"results"
  
  while true do
    -- Get chunk to process
    local chunk = chunks.pop()
    
    -- Process it
    local processed = apply_filter(chunk)
    
    -- Send result back
    @results: push(processed)
  end
end
```

### 7.2 3D Rendering with Distributed Processing

This example shows 3D scene rendering distributed across nodes:

```lua
-- On main node
function render_scene(scene, view_params)
  -- Divide scene into sectors
  local sectors = divide_scene(scene, 4)
  
  -- Distribute rendering tasks
  @Stack.new(RenderTask, Distributed): alias:"render_tasks"
  @Stack.new(RenderedSector, Distributed): alias:"rendered_sectors"
  
  -- Configure stacks
  @render_tasks: bind_source(1)
  @render_tasks: bind_targets({2,3,4,5})
  
  @rendered_sectors: bind_targets(1)
  @rendered_sectors: bind_sources({2,3,4,5})
  
  -- Send rendering tasks with view parameters
  for i = 1, #sectors do
    @render_tasks: push({
      sector = sectors[i],
      view = view_params
    })
  end
  
  -- Collect rendered sectors
  local frame_parts = {}
  for i = 1, #sectors do
    frame_parts[i] = rendered_sectors.pop()
  end
  
  -- Combine rendered sectors
  return combine_frame(frame_parts)
end
```

## 8. Comparison with Other Languages

### 8.1 Versus Go's Channel Model

Go's channels provide a similar concept of communication primitive:

```go
// Go channel
c := make(chan int)
go func() { c <- 42 }()
value := <-c
```

**ual Stack-as-Channel:**
```lua
@Stack.new(Integer, Distributed): alias:"c"
-- In one task
@c: push(42)
-- In another task
value = c.pop()
```

**Key Differences:**
1. ual's approach builds on existing stack primitives rather than introducing a separate channel concept
2. Explicitly distributed nature of ual stacks vs. implicit sharing in Go
3. ual's approach has lower overhead for embedded systems
4. ual provides more control over the underlying transport mechanism

### 8.2 Versus Erlang's Message Passing

Erlang uses process mailboxes and selective receive:

```erlang
% Erlang
Pid ! {message, 42}.
receive
  {message, Value} -> handle(Value)
end.
```

**ual Stack-as-Channel:**
```lua
@message_stack: push(42)

-- Receiving side
value = message_stack.pop()
```

**Key Differences:**
1. ual uses typed stacks for message passing vs. Erlang's dynamic pattern matching
2. No selective receive in basic ual (simpler and more predictable)
3. Erlang's model requires a more substantial runtime
4. ual's approach is more memory-efficient for constrained systems

### 8.3 Versus Rust's Concurrency Model

Rust uses channels from its standard library:

```rust
// Rust
let (tx, rx) = mpsc::channel();
thread::spawn(move || {
    tx.send(42).unwrap();
});
let value = rx.recv().unwrap();
```

**ual Stack-as-Channel:**
```lua
@Stack.new(Integer, Distributed): alias:"channel"
-- In one task
@channel: push(42)
-- In another task
value = channel.pop()
```

**Key Differences:**
1. Rust separates sender and receiver ends vs. ual's unified stack abstraction
2. Rust's ownership system affects channel usage; ual's approach is simpler
3. ual's model aligns better with its existing paradigm
4. Lower overhead in ual for embedded applications

### 8.4 Versus C/C++ Threading Models

Traditional C/C++ threading relies on shared memory and synchronization primitives:

```cpp
// C++ with std::thread
std::mutex mtx;
std::condition_variable cv;
std::queue<int> queue;

// Producer
{
    std::lock_guard<std::mutex> lock(mtx);
    queue.push(42);
}
cv.notify_one();

// Consumer
{
    std::unique_lock<std::mutex> lock(mtx);
    cv.wait(lock, []{ return !queue.empty(); });
    int value = queue.front();
    queue.pop();
}
```

**ual Stack-as-Channel:**
```lua
@Stack.new(Integer, Distributed): alias:"queue"
-- Producer
@queue: push(42)
-- Consumer
value = queue.pop()  -- Implicitly waits if empty
```

**Key Differences:**
1. ual's approach is dramatically simpler and less error-prone
2. No explicit locking or signaling required in ual
3. C++ approach requires careful management of mutex lifetime
4. ual's unified abstraction handles both communication and synchronization

## 9. Addressing Challenges

### 9.1 Deadlock Prevention

The proposed model includes mechanisms to detect and prevent deadlocks:

1. **Timeout Support**: All blocking operations can specify a timeout
```lua
-- Try to pop with a 100ms timeout
local value, success = calc_stack.pop_timeout(100)
```

2. **Deadlock Detection**: The runtime can monitor for circular waiting patterns
```lua
-- Enable deadlock detection
@Stack.enableDeadlockDetection(true)
```

3. **Resource Ordering**: Utilities to help enforce consistent resource acquisition order
```lua
@Stack.new(ResourceGroup): alias:"resources"
@resources: add(resource1, resource2, resource3)
@resources: acquire_all()  -- Acquires in consistent order
```

### 9.2 Error Handling

Robust error handling for distributed operations:

1. **Communication Failures**: Detect and report network or node failures
```lua
local success = pcall(function()
  @remote_stack: push(value)
end)
if not success then
  -- Handle communication failure
end
```

2. **Node Health Monitoring**: Track node status
```lua
local status = node_status(3)
if status == NODE_OFFLINE then
  -- Reroute tasks from node 3
end
```

3. **Graceful Degradation**: Fallback mechanisms when nodes are unavailable
```lua
function try_distributed_then_local(value)
  if node_available(2) then
    @remote_calc: push(value)
    return remote_calc.pop()
  else
    -- Fallback to local calculation
    return local_calculate(value)
  end
end
```

### 9.3 Resource Efficiency

Strategies for efficient resource usage:

1. **Zero-Copy Transfers**: Avoid excessive data copying
2. **Memory Pooling**: Pre-allocate communication buffers
3. **Strategic Serialization**: Only serialize data when crossing physical boundaries
4. **Prioritization**: Critical messages receive processing priority

## 10. Implementation Considerations

### 10.1 Compiler Support

The compiler must be extended to understand distributed stacks:

1. **Stack Type Analysis**: Tracking distributed stack properties
2. **Node Awareness**: Understanding which code runs on which nodes
3. **Optimization**: Special optimizations for local vs. distributed operations

### 10.2 Runtime Requirements

Minimal runtime support is needed:

1. **Communication Layer**: Abstracts underlying transport mechanisms
2. **Buffer Management**: Efficient handling of message buffers
3. **Synchronization Primitives**: Implementation of mutex, semaphore, etc.

### 10.3 Physical Communication Layer Abstraction

The implementation includes a pluggable transport layer that can use different physical communication mechanisms:

```lua
-- Create stack with specific transport mechanism
@Stack.new(Integer, Distributed, {transport="SPI"}): alias:"data"
```

Supported transports could include:
- SPI bus
- UART links
- I2C connections
- Custom parallel buses
- Memory-mapped regions (for cores in the same chip)

## 11. Future Directions

### 11.1 Advanced Features

Future versions could add more sophisticated features:

1. **Priority-Based Scheduling**: Message priority handling
2. **Quality of Service**: Bandwidth and latency guarantees
3. **Fault Tolerance**: Automatic failover and recovery
4. **Dynamic Topology**: Runtime adjustment of node connections

### 11.2 Language Extensions

Potential language extensions to enhance the concurrency model:

1. **Pattern Matching**: Selective receive based on message patterns
2. **Supervision Trees**: Erlang-like fault tolerance
3. **Software Transactional Memory**: For more complex shared state

## 12. Conclusion

The Stack-as-Channel concurrency model proposed for ual represents a natural extension of its existing paradigm to distributed computing environments. By leveraging stacks as the fundamental abstraction for both data storage and inter-process communication, this model maintains language consistency while enabling powerful concurrent programming patterns.

This approach is particularly well-suited for resource-constrained embedded systems like the ZX Interface with multiple ESP32 processors, where efficiency, simplicity, and predictability are paramount. The model provides the necessary tools for building complex distributed applications without sacrificing the performance characteristics essential for embedded systems.

By building on ual's existing stack abstraction rather than introducing foreign concepts, this concurrency model creates a coherent programming experience that allows developers to apply their existing knowledge of the language to distributed computing problems. The result is a uniquely elegant approach to concurrency that aligns perfectly with ual's design philosophy while addressing the practical needs of modern multi-processor embedded systems.