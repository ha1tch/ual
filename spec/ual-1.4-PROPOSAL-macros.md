# ual 1.4 PROPOSAL: macro system
This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the version that the proposal is targeting to be integrated into the main ual spec.

---
# Proposed Macro System for ual
## 1. Introduction

This document outlines a proposed macro system for the ual programming language, designed specifically for the needs of embedded and resource-constrained systems. The proposed macro system emphasizes compile-time code generation, simplicity, and zero runtime overhead, making it ideal for ual's target use cases.

## 2. Design Philosophy

The ual macro system adheres to the following principles:

1. **Compile-time only**: All macro processing happens during compilation with zero runtime overhead.
2. **Resource-efficient**: The implementation uses minimal memory and processing power during compilation.
3. **Predictable**: Macro expansion produces consistent, deterministic results.
4. **Syntactically consistent**: The macro syntax follows ual's existing conventions.
5. **Developer-friendly**: Macros produce clear error messages and traceable code for debugging.

## 3. Macro Syntax and Semantics

### 3.1 Macro Definition

Macros are defined using the `macro_define` keyword followed by the macro name, parameter list, body, and `end_macro`:

```lua
macro_define macro_name(param1, param2, ...)
  -- Macro body (standard ual code that may use parameters)
end_macro
```

### 3.2 Macro Expansion

Macros are expanded using the `macro_expand` keyword:

```lua
macro_expand macro_name(arg1, arg2, ...)
```

### 3.3 Conditional Compilation

The macro system supports conditional compilation:

```lua
macro_if CONDITION
  -- Code when condition is true
macro_elseif OTHER_CONDITION
  -- Code when other condition is true
macro_else
  -- Code when no conditions are true
macro_endif
```

### 3.4 File Inclusion

Files can be included at the macro level:

```lua
macro_include "filename.ual"
```

### 3.5 Compile-time Calculations

Simple calculations can be performed at compile time:

```lua
macro_define calculate_value
  local result = 0
  for i = 1, 10 do
    result = result + i * i
  end
  return result
end_macro

local value = macro_expand calculate_value  -- Replaced with the computed value
```

## 4. Use Cases and Examples

### 4.1 Hardware Abstraction Layers

Macros excel at creating hardware abstraction layers for different microcontrollers:

```lua
macro_define setup_gpio(pin, mode)
  macro_if TARGET == "AVR"
    -- AVR-specific GPIO setup
    local port = pin / 8
    local bit = pin % 8
    if mode == "OUTPUT" then
      DDRA + port = DDRA + port | (1 << bit)
    else
      DDRA + port = DDRA + port & ~(1 << bit)
    end
  macro_elseif TARGET == "ESP32"
    -- ESP32-specific GPIO setup
    gpio_config_t io_conf = {}
    io_conf.pin_bit_mask = 1 << pin
    io_conf.mode = mode == "OUTPUT" ? GPIO_MODE_OUTPUT : GPIO_MODE_INPUT
    gpio_config(io_conf)
  macro_else
    -- Generic implementation using ual's io package
    io.PinMode(pin, mode == "OUTPUT" ? io.OUTPUT : io.INPUT)
  macro_endif
end_macro

-- Usage
macro_expand setup_gpio(13, "OUTPUT")
```

### 4.2 Lookup Tables

Generating lookup tables at compile time saves memory and computation:

```lua
macro_define generate_sin_table(size)
  local result = "{\n"
  for i = 0, size-1 do
    local angle = (i * 360 / size) * (3.14159 / 180)
    local sin_val = math.sin(angle)
    -- Format as fixed-point for embedded systems
    local fixed_point = math.floor(sin_val * 32767)
    result = result .. string.format("  %d,\n", fixed_point)
  end
  result = result .. "}"
  return result
end_macro

-- Generate a sine lookup table with 256 entries
local sin_table = macro_expand generate_sin_table(256)
```

### 4.3 Register Manipulation

Macros can abstract away complex register manipulation patterns:

```lua
macro_define set_register_bits(register, start_bit, bit_count, value)
  local mask = ((1 << bit_count) - 1) << start_bit
  local shifted_value = (value & ((1 << bit_count) - 1)) << start_bit
  register = (register & ~mask) | shifted_value
end_macro

-- Usage - sets bits 4-7 of PORTA to value 0x5
macro_expand set_register_bits(PORTA, 4, 4, 0x5)
```

### 4.4 State Machines

Generating state machine code from declarations:

```lua
macro_define state_machine(name, states, events, transitions)
  -- Generate state enum
  local state_enum = "-- State definitions\n"
  for i, state in ipairs(states) do
    state_enum = state_enum .. string.format("local %s_STATE_%s = %d\n", name, state, i-1)
  end
  
  -- Generate event handlers
  local handlers = string.format([[
function %s_process_event(current_state, event)
  switch_case(current_state)
]], name)

  for i, state in ipairs(states) do
    handlers = handlers .. string.format([[
  case %s_STATE_%s:
    switch_case(event)
]], name, state)
    
    for event, target in pairs(transitions[state] or {}) do
      handlers = handlers .. string.format([[
    case %s_EVENT_%s:
      -- Transition to %s
      return %s_STATE_%s
]], name, event, target, name, target)
    end
    
    handlers = handlers .. [[
    default:
      -- No transition
      return current_state
    end_switch
]]
  end
  
  handlers = handlers .. [[
  default:
    -- Invalid state
    return current_state
  end_switch
end
]]

  return state_enum .. "\n" .. handlers
end_macro

-- Usage
local states = {"IDLE", "ACTIVE", "ERROR"}
local events = {"START", "STOP", "ERROR"}
local transitions = {
  IDLE = {START = "ACTIVE"},
  ACTIVE = {STOP = "IDLE", ERROR = "ERROR"},
  ERROR = {STOP = "IDLE"}
}

macro_expand state_machine("MOTOR", states, events, transitions)
```

### 4.5 Stack Effect Documentation and Verification

Generating stack documentation and verification code:

```lua
macro_define stack_effect(func_name, in_count, out_count, description)
  local result = string.format([[
-- Stack effect: %s ( %s -- %s )
function %s()
  -- Verify stack has enough items
  if dstack.depth() < %d then
    error("Stack underflow in %s: expected %d items")
  end
]], description, in_count, out_count, func_name, in_count, func_name, in_count)
  
  -- Add the function body where the macro is expanded
  return result
end_macro

-- Usage
macro_expand stack_effect("swap_top_pair", 4, 4, "Swap the top two pairs")
  local a = dstack.pop()
  local b = dstack.pop()
  local c = dstack.pop()
  local d = dstack.pop()
  
  dstack.push(b)
  dstack.push(a)
  dstack.push(d)
  dstack.push(c)
end
```

### 4.6 Interface Implementation

Generating implementations of common interfaces:

```lua
macro_define implement_storage(name, methods)
  local result = string.format([[
-- Storage implementation for %s
local %s = {}
]], name, name)

  -- Add standard methods
  if methods.read then
    result = result .. string.format([[
function %s.read(address)
  %s
end
]], name, methods.read)
  end
  
  if methods.write then
    result = result .. string.format([[
function %s.write(address, value)
  %s
end
]], name, methods.write)
  end
  
  if methods.erase then
    result = result .. string.format([[
function %s.erase(address, size)
  %s
end
]], name, methods.erase)
  end
  
  return result
end_macro

-- Usage
macro_expand implement_storage("EEPROM", {
  read = "return eeprom_read_byte(address)",
  write = "eeprom_write_byte(address, value)",
  erase = "for i = 0, size-1 do eeprom_write_byte(address + i, 0xFF) end"
})
```

### 4.7 Platform-Specific Code

Conditionally including platform-specific code:

```lua
macro_if TARGET == "Z80"
  -- Include Z80-specific libraries
  import "z80_hardware"
  
  function init_system()
    -- Z80-specific initialization
    z80_hardware.init_memory_map()
    z80_hardware.set_interrupt_mode(2)
  end
macro_elseif TARGET == "AVR"
  -- Include AVR-specific libraries
  import "avr_hardware"
  
  function init_system()
    -- AVR-specific initialization
    avr_hardware.set_clock_prescaler(0)
    avr_hardware.init_timers()
  end
macro_else
  -- Generic initialization
  function init_system()
    -- Basic initialization for unknown platform
    sys.init()
  end
macro_endif
```

### 4.8 Stacked Mode Integration

Macros that generate stacked mode code:

```lua
macro_define stack_swap_over_drop
  > swap over drop
end_macro

function process_values(a, b)
  dstack.push(a)
  dstack.push(b)
  
  -- This expands to the three stack operations
  macro_expand stack_swap_over_drop
  
  return dstack.pop()
end
```

### 4.9 Memory-Efficient Data Structures

Generating specialized data structures optimized for memory constraints:

```lua
macro_define packed_struct(name, fields)
  local result = string.format("-- Packed structure: %s\n", name)
  
  -- Calculate offsets and generate accessors
  local offset = 0
  for field_name, field_size in pairs(fields) do
    -- Generate getter
    result = result .. string.format([[
function %s_get_%s(ptr)
  return 
]], name, field_name)

    -- Generate bit extraction code based on field size
    if field_size == 1 then
      result = result .. string.format(
        "  (memory_read_byte(ptr + %d) >> %d) & 0x01\n", 
        math.floor(offset / 8), offset % 8
      )
    elseif field_size <= 8 then
      -- Field fits in a single byte
      result = result .. string.format(
        "  (memory_read_byte(ptr + %d) >> %d) & 0x%02X\n", 
        math.floor(offset / 8), offset % 8, (1 << field_size) - 1
      )
    else
      -- Multi-byte field - more complex extraction
      -- (Simplified implementation for example)
      result = result .. "  -- Multi-byte field extraction\n"
    end
    
    result = result .. "end\n\n"
    
    -- Generate setter
    result = result .. string.format([[
function %s_set_%s(ptr, value)
]], name, field_name)

    -- Generate bit setting code
    if field_size == 1 then
      result = result .. string.format([[
  local byte = memory_read_byte(ptr + %d)
  if value == 0 then
    byte = byte & ~(1 << %d)
  else
    byte = byte | (1 << %d)
  end
  memory_write_byte(ptr + %d, byte)
]], math.floor(offset / 8), offset % 8, offset % 8, math.floor(offset / 8))
    elseif field_size <= 8 then
      result = result .. string.format([[
  local byte = memory_read_byte(ptr + %d)
  byte = byte & ~(0x%02X << %d)
  byte = byte | ((value & 0x%02X) << %d)
  memory_write_byte(ptr + %d, byte)
]], math.floor(offset / 8), (1 << field_size) - 1, offset % 8, 
   (1 << field_size) - 1, offset % 8, math.floor(offset / 8))
    else
      -- Multi-byte field (simplified)
      result = result .. "  -- Multi-byte field setting\n"
    end
    
    result = result .. "end\n\n"
    
    offset = offset + field_size
  end
  
  -- Calculate total size in bytes (rounding up)
  local size_bytes = math.ceil(offset / 8)
  result = result .. string.format([[
function %s_size()
  return %d
end
]], name, size_bytes)
  
  return result
end_macro

-- Usage - create a packed structure with bit fields
macro_expand packed_struct("SENSOR_DATA", {
  active = 1,        -- 1 bit
  error = 1,         -- 1 bit
  sensor_type = 3,   -- 3 bits
  value = 11         -- 11 bits
})
```

### 4.10 Debug and Trace Facilities

Conditional debugging code that can be removed in production builds:

```lua
macro_define debug_mode(enable)
  macro_if enable
    function debug_print(msg, ...)
      fmt.Printf("[DEBUG] " .. msg .. "\n", ...)
    end
    
    function assert(condition, message)
      if not condition then
        fmt.Printf("[ASSERT] %s\n", message)
        sys.Exit(1)
      end
    end
  macro_else
    function debug_print(msg, ...)
      -- Empty function, optimized away in release builds
    end
    
    function assert(condition, message)
      -- Empty function, optimized away in release builds
    end
  macro_endif
end_macro

-- Enable debugging in development builds
macro_expand debug_mode(DEBUG_BUILD)
```

## 5. Implementation Considerations

### 5.1 Parsing and Processing

The macro processor would be implemented as a separate preprocessing step before the main ual compilation:

1. Parse the input file, recognizing macro definitions and expansions
2. Process macros in a top-down manner
3. Output processed ual code with macros expanded
4. Feed the expanded code to the main ual compiler

### 5.2 Error Reporting

The macro preprocessor should provide clear error messages:

1. Line and column information from the original source
2. Descriptive error messages
3. Context showing the problematic code

### 5.3 Debugging Support

To help with debugging macro-generated code:

1. Option to output intermediate files showing expanded macros
2. Comments in generated code indicating the macro source
3. Source mapping information for debuggers

### 5.4 Scope and Hygiene

To prevent variable name collisions:

1. Macros operate in the lexical scope where they are expanded
2. Local variables declared in macros should be prefixed to avoid conflicts
3. The preprocessor should warn about potential variable capture issues

## 6. Limitations

The proposed macro system intentionally has certain limitations to maintain simplicity and efficiency:

1. No recursive macro expansion (for predictability)
2. Limited compile-time computation capabilities
3. No complex AST manipulation
4. No runtime macro facilities

These limitations help ensure the macro system remains appropriate for embedded and resource-constrained environments.

## 7. Conclusion

The proposed macro system for ual provides powerful compile-time code generation capabilities while maintaining the language's focus on simplicity and efficiency. By enabling developers to abstract away complex, repetitive, or platform-specific code patterns, macros enhance ual's utility for embedded systems programming without compromising its resource efficiency or predictability.

This macro system would be particularly valuable for:

1. Hardware abstraction layers
2. Cross-platform development
3. Memory-efficient data structures
4. Code generation for specialized algorithms
5. Conditional compilation for different target platforms

By focusing on compile-time processing with zero runtime overhead, ual macros would enable more flexible and expressive code while preserving the language's suitability for resource-constrained environments.