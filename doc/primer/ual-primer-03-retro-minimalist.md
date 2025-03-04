# ual Primer for Retro and Minimalist Programmers

## Introduction

If you're drawn to the elegance of minimalist systems, the tactile nature of retro computing, or the aesthetics of retrofuturism, ual will feel like a breath of fresh air. This language bridges the gap between vintage stack-based systems like Forth and modern safety features, without sacrificing the minimalist principles that make retro programming so appealing.

## Minimalist Design Philosophy

ual embodies several design principles that align with minimalist computing:

1. **Essential Complexity Only**: No bloat, no unnecessary abstractions
2. **Visible Mechanics**: The stack paradigm makes program flow tangible
3. **Direct Hardware Control**: Bitwise operations and memory access when needed
4. **Resource Efficiency**: Optimized for constrained environments
5. **Cross-Platform**: From Z80 to RISC-V, write once and target many architectures

## The Stack as a Conceptual Model

At its core, ual embraces the elegance of stack-based computation:

```lua
-- Basic stack operations
> push:10 push:20 add  -- Result: 30
> dup mul              -- Result: 900
```

If you've worked with Forth or other stack languages, this fundamental concept will be familiar. However, ual modernizes this approach with explicit typing and improved readability.

## Typed Stacks: Minimalism Meets Safety

```lua
@Stack.new(Integer): alias:"i"  -- Stack for integers
@Stack.new(String): alias:"s"   -- Stack for strings

@i: push(42)          -- Valid
@s: push("hello")     -- Valid
@i: push("hello")     -- Error: wrong type
```

This combines the beauty of stack-based programming with type safety guarantees, while remaining simple and explicit.

## Cross-Stack Operations

One of ual's most elegant features is the atomic `bring_<type>` operation for cross-stack transfers:

```lua
@s: push("42")        -- Push string to string stack
@i: <s                -- Transfer to integer stack with conversion

-- Can also be written as:
@i: bring_string(s.pop())
```

## Retrofuturistic Systems Programming

### Direct Hardware Access

For retro and minimal hardware, ual provides direct access with bitwise operations:

```lua
-- Z80-style port I/O
function out_port(port, value)
  port_write_byte(port, value)
end

function in_port(port)
  return port_read_byte(port)
end

-- Toggle bit 3 on port 0x42
function toggle_bit_3()
  val = in_port(0x42)
  val = val ^ 0x08  -- XOR with bit 3 mask
  out_port(0x42, val)
end
```

### Memory-Mapped Hardware

```lua
-- Memory-mapped video buffer for retro display
VIDEO_MEM = 0xA000
VIDEO_WIDTH = 40
VIDEO_HEIGHT = 25

-- Put character at x,y position
function put_char(x, y, char)
  if x >= 0 and x < VIDEO_WIDTH and y >= 0 and y < VIDEO_HEIGHT then
    addr = VIDEO_MEM + (y * VIDEO_WIDTH) + x
    memory_write_byte(addr, char)
  end
end

-- Draw a box on screen
function draw_box(x, y, width, height, char)
  -- Draw top and bottom
  for i = 0, width-1 do
    put_char(x+i, y, char)
    put_char(x+i, y+height-1, char)
  end
  
  -- Draw sides
  for i = 1, height-2 do
    put_char(x, y+i, char)
    put_char(x+width-1, y+i, char)
  end
end
```

## Stack-Based Text Processing

Text processing in ual can be elegantly expressed with stacks:

```lua
function tokenize(text, delimiter)
  @Stack.new(String): alias:"tokens"
  @Stack.new(String): alias:"work"
  
  @work: push(text)
  
  -- Split the string and push tokens in reverse order
  while_true(work.peek():find(delimiter) > 0)
    pos = work.peek():find(delimiter)
    token = work.peek():sub(1, pos-1)
    @tokens: push(token)
    @work: push(work.pop():sub(pos+1))
  end_while_true
  
  -- Push the last token
  if work.peek():len() > 0 then
    @tokens: push(work.pop())
  else
    work.drop()
  end
  
  -- Create result array with tokens in original order
  local result = []
  while_true(tokens.depth() > 0)
    table.insert(result, 1, tokens.pop())
  end_while_true
  
  return result
end
```

## Retrofuturistic Computing Example: ASCII Art Generator

Here's a more complete example showing how ual can be used for a retrofuturistic ASCII art generator:

```lua
package asciiart

import "con"
import "fmt"

-- Character brightness values (from darkest to brightest)
CHARS = " .:-=+*#%@"

-- Convert grayscale value (0-255) to ASCII character
function gray_to_char(value)
  @Stack.new(Integer): alias:"i"
  @i: push(value) push:25 div  -- Scale to 0-9
  
  -- Clamp to valid range
  @i: dup push:0 lt if_true
    @i: drop push:0
  @i: end_if_true
  
  @i: dup push:9 gt if_true
    @i: drop push:9
  @i: end_if_true
  
  -- Convert to character index and return char
  index = i.pop() + 1
  return CHARS:sub(index, index)
end

-- Generate a sine wave pattern
function generate_sine_pattern(width, height, scale, offset)
  @Stack.new(Integer): alias:"i"
  @Stack.new(Float): alias:"f"
  
  result = {}
  
  for y = 0, height-1 do
    line = ""
    for x = 0, width-1 do
      @f: push(x) push:scale mul push:offset add sin
      @f: push:1.0 push:0.5 mul add  -- Range 0.5 to 1.5
      @i: <f push:255 mul            -- Scale to 0-255
      
      line = line .. gray_to_char(i.pop())
    end
    table.insert(result, line)
  end
  
  return result
end

-- Generate a circular pattern
function generate_circle_pattern(width, height, radius)
  @Stack.new(Integer): alias:"i"
  @Stack.new(Float): alias:"f"
  
  result = {}
  center_x = width / 2
  center_y = height / 2
  
  for y = 0, height-1 do
    line = ""
    for x = 0, width-1 do
      @f: push(x) push:center_x sub dup mul
      @f: push(y) push:center_y sub dup mul
      @f: add sqrt                     -- Distance from center
      
      @f: push:radius div              -- Normalize by radius
      @f: push:1.0 swap sub abs        -- Proximity to circle edge
      @f: push:4.0 mul                 -- Enhance contrast
      @f: push:1.0 min                 -- Clamp to 1.0
      @i: <f push:255 mul              -- Scale to 0-255
      
      line = line .. gray_to_char(i.pop())
    end
    table.insert(result, line)
  end
  
  return result
end

-- Display ASCII art pattern
function display_pattern(pattern)
  con.Cls()
  for i = 1, #pattern do
    fmt.Printf("%s\n", pattern[i])
  end
end

-- Main animation loop
function animate_sine_wave()
  @Stack.new(Float): alias:"f"
  @f: push:0.0  -- Initial offset
  
  width = 80
  height = 24
  
  while_true(true)
    pattern = generate_sine_pattern(width, height, 0.1, f.pop())
    display_pattern(pattern)
    
    -- Update offset for next frame
    @f: push(sys.Millis()) push:1000.0 div
    
    -- Delay
    sys.Sleep(50)
  end_while_true
end

function main()
  animate_sine_wave()
  return 0
end
```

## Stack as Memory: A Retro Mental Model

One of the appeals of stack-based programming for retro computing enthusiasts is how it mirrors the way hardware works. ual preserves this connection while adding safety:

```lua
function emulate_6502_stack()
  -- Create a stack that simulates the 6502's 256-byte stack
  @Stack.new(Integer): alias:"s6502"
  
  -- Initialize stack pointer (SP) to 0xFF (empty stack)
  sp = 0xFF
  
  -- Push operation (decrement SP, then store)
  function stack_push(value)
    sp = (sp - 1) & 0xFF  -- 8-bit wrap
    memory_write_byte(0x0100 + sp, value & 0xFF)
  end
  
  -- Pop operation (load, then increment SP)
  function stack_pop()
    value = memory_read_byte(0x0100 + sp)
    sp = (sp + 1) & 0xFF  -- 8-bit wrap
    return value
  end
  
  -- Demonstrate usage
  stack_push(0x42)
  stack_push(0x7F)
  fmt.Printf("Popped: 0x%02X\n", stack_pop())  -- 0x7F
  fmt.Printf("Popped: 0x%02X\n", stack_pop())  -- 0x42
end
```

## Minimalist Resource Management

ual's proposed ownership system (1.5) provides elegant resource management:

```lua
function process_file(filename)
  @Stack.new(File, Owned): alias:"f"
  @f: push(open_file(filename, "r"))
  
  -- File automatically closed when 'f' goes out of scope
  return process_content(f.peek().read_all())
end
```

This minimalist approach ensures resources are properly managed without complex RAII patterns or garbage collection.

## Why Choose ual for Retro and Minimalist Programming?

1. **Beautiful Minimalism**: Express complex ideas with minimal syntax
2. **Tangible Computation**: Stack operations provide a concrete mental model
3. **Cross-Platform**: Write code that works on classic Z80 up to modern RISC-V
4. **Resource Efficiency**: Optimized for constrained environments
5. **Direct Control**: Bitwise operations and memory access
6. **Modern Safety**: Type checking and memory safety without bloat

## Next Steps

To explore more about ual for retro and minimalist programming:
- Dive into the [stack as first-class objects](https://github.com/ha1tch/ual/blob/main/spec/ual-1.3-spec-P2.md) for more advanced stack operations
- Explore [typed stacks](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-typed-stacks-01.md) for adding safety to stack operations
- Check out the [macro system](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-macros.md) for code generation and optimization

ual offers a perfect blend of retro programming principles and modern safety features, allowing you to create efficient, elegant code for everything from vintage computers to modern minimal systems.