# ual Primer for Rust Developers

## Introduction

As a Rust developer, you're already familiar with a language that prioritizes safety, performance, and memory efficiency. **ual** will feel both familiar and refreshing, offering an alternative approach to these same priorities. Where Rust enforces safety through ownership, borrowing, and lifetimes, ual achieves similar guarantees through explicit stack operations and a container-centric type system.

## Familiar Territory

Let's start with aspects of ual that will feel familiar coming from Rust:

### Memory Safety Without Runtime Overhead

Like Rust, ual provides strong safety guarantees without garbage collection or runtime checks:

```lua
-- In Rust:
-- fn process(data: &mut Vec<i32>) {
--    data.push(42);
-- }

-- In ual (proposed ownership system):
function process(data)
  @Stack.new(Array, Mutable): alias:"arr"
  @arr: <:mut data               -- Mutable borrow
  @arr: push(arr.peek().append(42))
end
```

The ual approach makes borrowing explicit through stack operations, with the compiler enforcing similar safety rules to Rust's borrow checker.

### Type Safety

Rust's strong type system has a counterpart in ual's container-centric types:

```lua
-- In Rust:
-- let numbers: Vec<i32> = vec![1, 2, 3];
-- let text: String = "Hello".to_string();
-- numbers.push(text); // Error: expected i32, found String

-- In ual:
@Stack.new(Integer): alias:"numbers"
@Stack.new(String): alias:"text"

@numbers: push:1 push:2 push:3
@text: push("Hello")

@numbers: <text  -- Error: string value cannot be brought into integer stack
```

The key difference: in Rust, types are properties of values and variables; in ual, types are properties of containers (stacks).

### Pattern Matching

If you appreciate Rust's pattern matching, you'll find ual's approach conceptually similar but with its own style:

```lua
-- In Rust:
-- match result {
--    Ok(value) => println!("Success: {}", value),
--    Err(error) => println!("Error: {}", error),
-- }

-- In ual:
result.consider {
  if_ok fmt.Printf("Success: %v\n", _1)
  if_err fmt.Printf("Error: %v\n", _1)
}
```

The `.consider{}` construct provides similar expressiveness to Rust's `match` for result handling.

### Zero-Cost Abstractions

Like Rust, ual prioritizes zero-cost abstractions:

```lua
-- In Rust:
-- let doubled: Vec<i32> = numbers.iter().map(|x| x * 2).collect();

-- In ual (using stack operations):
@Stack.new(Integer): alias:"numbers"
@Stack.new(Integer): alias:"doubled"

@numbers: push:1 push:2 push:3

-- Map operation
@numbers: depth() while_true(_1 > 0)
  @doubled: numbers.peek() push:2 mul
  @numbers: pop() drop
@numbers: end_while_true
```

All stack operations compile to efficient code with no runtime overhead.

## The Stack-Based Approach

Now let's explore how ual's stack-based paradigm differs from Rust's approach:

### Explicit Data Flow

In Rust, data flows through variables and function parameters. In ual, data can also flow through stacks:

```lua
-- Traditional variable approach (similar to Rust)
function calculate_area(width, height)
  return width * height
end

-- Stack-based approach
function calculate_area_stack()
  mul()  -- Multiply the top two stack values
  return pop()
end

-- Using the stack functions
push(5)
push(10)
area = calculate_area_stack()  -- area = 50
```

The stack-based approach can be particularly elegant for certain algorithms.

### Stacked Mode Syntax

For more complex stack operations, ual offers a concise "stacked mode" syntax:

```lua
function factorial(n)
  @dstack: push(n) push:1 eq if_true
    @dstack: drop push:1
    return dstack.pop()
  @dstack: end_if_true
  
  @dstack: push(n) push:1 sub
  @dstack: factorial mul
  
  return dstack.pop()
end
```

This syntax emphasizes data transformations similar to how Rust's iterators create transformation pipelines.

### Container-Centric vs. Value-Centric

This is perhaps the biggest conceptual shift from Rust to ual:

```lua
-- In Rust (value-centric):
-- let x: i32 = 42;
-- let y: String = "hello".to_string();

-- In ual (container-centric):
@Stack.new(Integer): alias:"i"  -- Container for integers
@Stack.new(String): alias:"s"   -- Container for strings

@i: push(42)       -- Put value in container
@s: push("hello")  -- Put value in container
```

In Rust, a value has an intrinsic type. In ual, a value's acceptability depends on the container it's entering.

## ual's Stack-Based Ownership

### The Proposed Ownership System

ual's proposed ownership system (1.5) will feel conceptually familiar to Rust developers, but with a stack-based twist:

```lua
-- In Rust:
-- fn process(owned: String) { /* Takes ownership */ }
-- fn borrow(borrowed: &String) { /* Borrows immutably */ }
-- fn borrow_mut(borrowed: &mut String) { /* Borrows mutably */ }

-- In ual:
@Stack.new(String, Owned): alias:"owned"
@Stack.new(String, Borrowed): alias:"borrowed"
@Stack.new(String, Mutable): alias:"mutable"

@owned: push("Hello")          -- Owned value

@borrowed: <<owned             -- Immutable borrow (shorthand)
@borrowed: borrow(owned.peek()) -- Immutable borrow (explicit)

@mutable: <:mut owned          -- Mutable borrow (shorthand)
@mutable: borrow_mut(owned.peek()) -- Mutable borrow (explicit)
```

The same concepts apply - ownership, immutable borrowing, and mutable borrowing - but the mechanics differ.

### Ownership Transfer

Transferring ownership in ual is explicit:

```lua
-- In Rust:
-- let s1 = String::from("hello");
-- let s2 = s1; // Ownership moved from s1 to s2
-- // s1 is no longer valid

-- In ual:
@Stack.new(String, Owned): alias:"s1"
@Stack.new(String, Owned): alias:"s2"

@s1: push("hello")
@s2: <:own s1        -- Take ownership from s1
-- s1.pop() would error, as the value is now owned by s2
```

### Resource Management and RAII

ual's approach to resource management parallels Rust's RAII principle:

```lua
-- In Rust:
-- {
--    let file = File::open("example.txt")?; // Opened here
--    // Use file...
-- } // File automatically closed here when it goes out of scope

-- In ual:
function process_file(filename)
  @Stack.new(File, Owned): alias:"f"
  @f: push(open_file(filename, "r"))
  
  -- Use the file...
  
  -- File automatically closed when 'f' goes out of scope
end
```

The stack-based approach makes resource ownership and cleanup explicit while maintaining the same safety guarantees.

## Error Handling

### Result Pattern

Rust's `Result<T, E>` pattern has a parallel in ual:

```lua
-- In Rust:
-- let result = operation_that_might_fail();
-- match result {
--     Ok(value) => println!("Success: {}", value),
--     Err(e) => println!("Error: {}", e),
-- }

-- In ual:
result = operation_that_might_fail()

result.consider {
  if_ok fmt.Printf("Success: %v\n", _1)
  if_err fmt.Printf("Error: %v\n", _1)
}
```

### Propagating Errors

Like Rust's `?` operator, ual provides ways to propagate errors:

```lua
-- In Rust:
-- fn process() -> Result<String, Error> {
--     let file = File::open("data.txt")?;
--     let content = file.read_to_string()?;
--     Ok(content)
-- }

-- In ual:
@error > function process()
  file_result = io.open("data.txt", "r")
  file_result.consider {
    if_err {
      @error > push(_1)
      return nil
    }
  }
  
  file = file_result.Ok
  content_result = file.read_all()
  file.close()
  
  content_result.consider {
    if_err {
      @error > push(_1)
      return nil
    }
  }
  
  return content_result.Ok
end
```

The `@error >` stack provides a mechanism similar to Rust's error propagation.

## Type Conversion and Generics

### Type Conversion

ual handles type conversions with explicit operations:

```lua
-- In Rust:
-- let num_str = "42";
-- let num: i32 = num_str.parse()?;

-- In ual:
@Stack.new(String): alias:"s"
@Stack.new(Integer): alias:"i"

@s: push("42")
@i: <s  -- Convert and transfer from string stack to integer stack
```

### Generics through Stack-Based Abstraction

While ual doesn't have traditional generics like Rust, it achieves similar flexibility through stack operations:

```lua
-- In Rust:
-- fn reverse<T>(slice: &mut [T]) {
--     let len = slice.len();
--     for i in 0..len/2 {
--         slice.swap(i, len - 1 - i);
--     }
-- }

-- In ual (proposed in 1.5):
function reverse(stack Stack)
  @Stack.new(Any): alias:"temp"
  
  -- Move all elements to temporary stack (reverses order)
  while_true(stack.depth() > 0)
    @temp: bring(stack.pop())
  end_while_true
  
  -- Move all elements back to original stack
  while_true(temp.depth() > 0)
    stack.push(temp.pop())
  end_while_true
end
```

This approach achieves similar results to generics but with an emphasis on operations rather than types.

## A Practical Example: Safe Resource Handling

Let's explore a complete example showing ual's safety mechanisms in a context familiar to Rust developers:

```lua
-- Safe file processing with error handling and resource management

@error > function process_file(filename)
  -- Create resource stack with ownership semantics
  @Stack.new(File, Owned): alias:"file"
  
  -- Try to open file
  result = io.open(filename, "r").consider {
    if_ok file.push(_1)
    if_err {
      @error > push("Failed to open file: " .. _1)
      return nil
    }
  }

  -- Stack for collecting lines
  @Stack.new(String): alias:"lines"
  
  -- Process file line by line
  line_result = file.peek().read_line().consider {
    if_ok {
      while_true(_1 != nil)
        @lines: push(_1)
        
        -- Read next line
        line_result = file.peek().read_line()
        _1 = line_result.Ok
      end_while_true
    }
    if_err {
      @error > push("Error reading file: " .. _1)
      return nil
    }
  }
  
  -- File is automatically closed when 'file' stack goes out of scope
  
  -- Process collected lines
  @Stack.new(Integer): alias:"counts"
  
  while_true(lines.depth() > 0)
    line = lines.pop()
    @counts: push(#line)  -- Push line length
  end_while_true
  
  return counts
end

-- Usage
function main()
  filename = "example.txt"
  
  -- Process file and handle errors
  counts = process_file(filename)
  
  if counts then
    fmt.Printf("Line counts: %v\n", counts)
  else if @error > depth() > 0 then
    fmt.Printf("Error: %s\n", @error > pop())
  end
  
  return 0
end
```

This example demonstrates how ual provides similar safety guarantees to Rust but with a stack-based approach.

## Key Differences from Rust

### 1. Explicit vs. Implicit

Rust's borrow checker works implicitly; ual's stack operations make data flow explicit:

```lua
-- Rust's implicit borrowing:
-- fn process(data: &mut Vec<i32>) {
--     data.push(42); // Implicit mutable borrow
-- }

-- ual's explicit stack operations:
function process(data)
  @Stack.new(Array, Mutable): alias:"arr"
  @arr: <:mut data  -- Explicit mutable borrow
  arr.peek().append(42)
end
```

### 2. Container-Centric vs. Value-Centric

Rust's type system focuses on values; ual's focuses on containers:

```lua
-- Rust (value has a type):
-- let x: i32 = 42;

-- ual (container accepts types):
@Stack.new(Integer): alias:"i"
@i: push(42)
```

### 3. Stack-Based vs. Expression-Based

Rust is primarily expression-based; ual offers both approaches with an emphasis on stack operations:

```lua
-- Rust's expression-based style:
-- let area = width * height;

-- ual's variable style:
area = width * height

-- ual's stack style:
@dstack: push(width) push(height) mul
area = dstack.pop()
```

### 4. Error Handling Approaches

Rust uses `Result` and `?`; ual uses `.consider{}` and `@error >`:

```lua
-- Rust:
-- fn process() -> Result<String, Error> {
--     let file = File::open("data.txt")?;
--     let content = file.read_to_string()?;
--     Ok(content)
-- }

-- ual:
@error > function process()
  file = io.open("data.txt", "r").consider {
    if_err {
      @error > push(_1)
      return nil
    }
    if_ok return _1
  }
  
  content = file.read_all().consider {
    if_err {
      @error > push(_1)
      return nil
    }
    if_ok return _1
  }
end
```

## Why a Rust Developer Might Love ual

1. **Similar Safety Focus**: Both languages prioritize memory safety without runtime overhead
2. **Explicit Operations**: If you like Rust's explicitness, you'll appreciate ual's stack operations
3. **Resource Efficiency**: ual shares Rust's emphasis on efficiency and low overhead
4. **Ownership Model**: ual's stack-based ownership provides similar guarantees with a different mental model
5. **Progressive Learning**: Start with familiar imperative code and gradually adopt stack-based patterns

## Compelling Use Cases for Rust Developers

### 1. Embedded Systems with Clear Data Flow

Embedded systems often involve transforming data through multiple processing stages. While Rust handles this well with iterators and ownership, ual's stack-based approach can make data flow even more explicit:

```lua
function process_sensor_data(raw_readings)
  @Stack.new(Integer): alias:"raw"
  @Stack.new(Float): alias:"normalized"
  @Stack.new(Float): alias:"filtered"
  
  -- Load raw readings
  for i = 1, #raw_readings do
    @raw: push(raw_readings[i])
  end
  
  -- Normalize readings (0-1023 ADC values to 0.0-3.3V)
  @raw: depth() while_true(_1 > 0)
    @normalized: <raw push:1023.0 div push:3.3 mul
  @raw: end_while_true
  
  -- Apply moving average filter
  prev = 0
  @normalized: depth() while_true(_1 > 0)
    current = normalized.pop()
    @filtered: push((current + prev) / 2)
    prev = current
  @normalized: end_while_true
  
  return filtered
end
```

The explicit data movement through different stacks creates a visual pipeline that's immediately apparent in the code structure.

### 2. State Machines Without Borrow Checker Battles

Rust's borrow checker can sometimes make state machine implementations challenging, particularly with complex state transitions. Ual's stack approach offers a refreshingly straightforward alternative:

```lua
function create_state_machine()
  @Stack.new(String): alias:"state"
  @state: push("IDLE")  -- Initial state
  
  @Stack.new(Event): alias:"events"
  @Stack.new(Function): alias:"handlers"
  
  -- Define state transitions
  function transition(new_state)
    @state: drop push(new_state)
  end
  
  -- Register event handlers
  function on(event_type, handler)
    @events: push(event_type)
    @handlers: push(handler)
  end
  
  -- Handle incoming events
  function process(event)
    @Stack.new(Integer): alias:"i"
    found = false
    
    for i = 0, events.depth() - 1 do
      if events.peek(i) == event.type and 
         (events.peek(i) == "*" or state.peek() == event.state) then
        handlers.peek(i)(event, state.peek(), transition)
        found = true
        break
      end
    end
    
    if not found then
      fmt.Printf("Unhandled event: %s in state %s\n", 
                event.type, state.peek())
    end
  end
  
  return {
    on = on,
    process = process,
    get_state = function() return state.peek() end
  }
end
```

This approach avoids ownership issues entirely while maintaining type safety.

### 3. Hardware Register Manipulation Without Unsafe

Rust often requires `unsafe` blocks for direct hardware access. Ual provides safe abstractions for low-level operations that feel natural:

```lua
-- Set specific bits in a hardware register
function configure_peripheral(base_addr)
  @Stack.new(Integer): alias:"reg"
  
  -- Read current register value
  @reg: push(memory_read_32(base_addr + 0x04))
  
  -- Clear configuration bits (bits 8-11)
  @reg: push(~(0x0F << 8)) and
  
  -- Set new configuration (value 0x5 in bits 8-11)
  @reg: push(0x05 << 8) or
  
  -- Write back
  memory_write_32(base_addr + 0x04, reg.pop())
  
  -- Enable peripheral (bit 0)
  @reg: push(memory_read_32(base_addr + 0x00))
  @reg: push(1) or
  memory_write_32(base_addr + 0x00, reg.pop())
end
```

The stack operations map naturally to register manipulation without requiring special safety annotations.

### 4. Parallel Data Transformations Without Fighting Lifetimes

When implementing parallel data transformations in Rust, lifetimes and ownership can become complex. Ual's approach simplifies this:

```lua
function parallel_transform(data)
  -- Create stacks for processing stages
  @Stack.new(Integer): alias:"input"
  @Stack.new(Integer): alias:"stage1"
  @Stack.new(Integer): alias:"stage2"
  @Stack.new(Integer): alias:"output"
  
  -- Load input data
  for i = 1, #data do
    @input: push(data[i])
  end
  
  -- Spawn workers for each stage
  @spawn: function(input, output) {
    -- Stage 1: Double each value
    while_true(input.depth() > 0)
      @output: input.pop() push:2 mul
    end_while_true
  }(input, stage1)
  
  @spawn: function(input, output) {
    -- Stage 2: Add 10 to each value
    while_true(input.depth() > 0)
      @output: input.pop() push:10 add
    end_while_true
  }(stage1, stage2)
  
  @spawn: function(input, output) {
    -- Stage 3: Calculate square root
    while_true(input.depth() > 0)
      @output: input.pop() sqrt
    end_while_true
  }(stage2, output)
  
  -- Wait for all stages to complete
  @spawn: wait_all()
  
  -- Collect results
  results = {}
  while_true(output.depth() > 0)
    table.insert(results, 1, output.pop())
  end_while_true
  
  return results
end
```

This pipeline-based approach offers clear data flow without complex borrowing annotations.

### 5. Resource Management Without Drop Traits

Rust's RAII is powerful but requires implementing Drop traits. Ual achieves similar safety with stack lifetime:

```lua
function process_multiple_resources()
  -- All resources automatically cleaned up when stacks go out of scope
  @Stack.new(File, Owned): alias:"log_file"
  @Stack.new(Network, Owned): alias:"connection"
  @Stack.new(Database, Owned): alias:"db"
  
  @log_file: push(open_file("logs.txt", "w"))
  
  -- If any operation fails, all previously opened resources are properly closed
  result = connect_to_server("example.com", 8080).consider {
    if_ok connection.push(_1)
    if_err {
      log_file.peek().write("Connection failed: " .. _1)
      return { Err = _1 }
    }
  }
  
  result = open_database("users.db").consider {
    if_ok db.push(_1)
    if_err {
      log_file.peek().write("Database open failed: " .. _1)
      return { Err = _1 }
    }
  }
  
  -- Actual processing using all resources
  process_data(log_file.peek(), connection.peek(), db.peek())
  
  return { Ok = true }
}
```

The stack-based ownership model ensures proper cleanup without explicit destructors.

## ual's Relationship to Rust

It's worth reflecting on how ual might relate to Rust in the broader programming language ecosystem. Just as languages like Scala, Kotlin, and Clojure emerged not to replace Java but to explore alternative approaches within the JVM ecosystem, ual aspires to be a complementary language that provides a fresh perspective on systems programming.

Ual doesn't aim to compete with or replace Rust, but rather to offer an alternative way of thinking about many of the same problems Rust solves brilliantly. While Rust has revolutionized systems programming with its ownership model and zero-cost abstractions, ual explores how these same principles might be expressed through a stack-based paradigm.

If ual earns its place in the programming language landscape, it might become to Rust what these alternative JVM languages became to Java: a thoughtful contribution that expands the range of tools available to developers and enriches the broader conversation about how we design safe, efficient systems.

This relationship could benefit both communities:
- Rust developers might find in ual a source of fresh ideas and alternative approaches
- The explicit stack-based model in ual might influence future Rust libraries or patterns
- Concepts from both languages could cross-pollinate, leading to innovations in systems programming

Time will tell if ual can earn such a position, but the aspiration is not to compete with Rust's excellence but to contribute to the ongoing evolution of systems programming languages in a way that respects and builds upon Rust's groundbreaking work.

## Next Steps

To dive deeper into ual from a Rust developer's perspective:
- Explore the [stack-based ownership system](https://github.com/ha1tch/ual/blob/main/spec/ual-1.5-PROPOSAL-ownership-system.md) to see parallels with Rust's ownership
- Learn about [typed stacks](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-typed-stacks-01.md) for understanding ual's approach to type safety
- Check out the [error handling mechanisms](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-error-stack.md) that parallel Rust's Result pattern

ual offers a refreshing perspective on the same problems Rust solves, with its own unique approach that maintains safety while exploring new programming paradigms.