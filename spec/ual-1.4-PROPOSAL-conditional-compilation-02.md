# ual 1.4 PROPOSAL: Conditional Compilation

## 1. Introduction

This document proposes an integrated approach to conditional compilation for the UAL programming language. Conditional compilation is essential for embedded systems development where code must adapt to different hardware platforms, memory constraints, and optional features. Instead of introducing a separate preprocessor with its own syntax (as in C/C++), this proposal recommends leveraging UAL's macro system to provide a more cohesive and maintainable approach to platform-specific code generation.

## 2. Background and Motivation

### 2.1 The Need for Conditional Compilation in Embedded Systems

Embedded systems development presents unique challenges that make conditional compilation particularly valuable:

1. **Hardware Diversity**: Code often needs to run on multiple microcontrollers with different register layouts, peripherals, and capabilities.

2. **Resource Constraints**: Different platforms have varying amounts of memory, processing power, and peripheral features, requiring tailored implementations.

3. **Feature Toggling**: The ability to include or exclude features based on project requirements helps optimize resource usage.

4. **Debug vs. Release**: Different code paths are often needed during development versus production deployment.

### 2.2 Limitations of Traditional Approaches

Existing approaches to conditional compilation have significant drawbacks:

1. **C Preprocessor Model**: Traditional `#ifdef`/`#endif` directives create a separate language within the code that complicates readability, debugging, and maintenance. Deeply nested conditionals become particularly problematic.

2. **Go's File-Level Approach**: While clean in simple cases, Go's file-based build constraints can lead to duplication when conditionals need to be applied at a finer granularity.

3. **Build System Complexity**: Relying on build systems for conditional inclusion often leads to complex build scripts that are difficult to maintain and understand.

### 2.3 The UAL Opportunity

UAL's design philosophy emphasizes simplicity, predictability, and zero runtime overhead—ideally suited for embedded systems. By integrating conditional compilation directly into UAL's macro system, we can provide powerful capabilities while maintaining language cohesion and avoiding the pitfalls of separate preprocessor languages.

## 3. Design Principles

The proposed conditional compilation system for UAL adheres to these principles:

1. **Zero Runtime Overhead**: All conditional compilation occurs at compile time with no impact on runtime performance.

2. **Syntactic Consistency**: Conditional compilation uses standard UAL syntax rather than introducing a separate preprocessor dialect.

3. **Readability**: Conditional code should be easy to read and understand, with clear boundaries and minimal nesting complexity.

4. **Flexibility**: The system should support both coarse-grained (file-level) and fine-grained (function or block level) conditional compilation.

5. **Deterministic Behavior**: Given the same inputs and configuration, conditional compilation should produce identical results.

6. **Clear Error Reporting**: Compilation errors in conditional blocks should provide clear source locations and helpful error messages.

## 4. Proposed Implementation

### 4.1 Compile-Time Environment Variables

A set of predefined compile-time variables would be available during macro expansion:

```lua
-- Predefined compile-time variables
TARGET          -- String identifying the target platform (e.g., "AVR", "ESP32", "Z80")
FEATURES        -- Table of enabled features (e.g., FEATURES.hardware_float)
DEBUG           -- Boolean indicating if this is a debug build
USE_SOFTWARE_FLOAT -- Boolean that can override hardware floating-point
CPU_BITS        -- Integer indicating the target CPU bit width (8, 16, 32, etc.)
COMPILER_VERSION -- String indicating the UAL compiler version
```

These variables would be set through compiler flags or build configuration files:

```
ualc program.ual -DTARGET=ESP32 -DDEBUG=true -DFEATURES.hardware_float=true
```

### 4.2 Basic Conditional Macros

The foundation of the conditional compilation system consists of simple macros that evaluate conditions at compile time:

```lua
macro_define when(condition, code)
  if condition then
    return code
  else
    return ""
  end
end_macro

macro_define unless(condition, code)
  if not condition then
    return code
  else
    return ""
  end
end_macro
```

These macros take two parameters:
1. A condition to evaluate at compile time
2. A code block to include or exclude based on the condition

The macros return either the code block (as a string) or an empty string, which is then inserted into the compilation unit.

### 4.3 Usage Patterns

#### 4.3.1 Platform-Specific Code

```lua
macro_expand when(TARGET == "AVR", [[
  function init_timers()
    -- AVR-specific timer initialization
    TCCR0A = 0x83
    TCCR0B = 0x04
    TIMSK0 = 0x01
  end
]])

macro_expand when(TARGET == "ESP32", [[
  function init_timers()
    -- ESP32-specific timer initialization
    timer_config_t config = {
      alarm_en = true,
      counter_en = false,
      intr_type = TIMER_INTR_LEVEL,
      counter_dir = TIMER_COUNT_UP,
      auto_reload = true,
      divider = 80  -- 1 MHz
    }
    timer_init(TIMER_GROUP_0, TIMER_0, &config)
  end
]])
```

#### 4.3.2 Feature Toggling

```lua
macro_expand when(FEATURES.hardware_float and not USE_SOFTWARE_FLOAT, [[
  -- Hardware floating-point implementation
  function calculate_trajectory(angle, velocity)
    local radians = angle * (3.14159 / 180.0)
    local sin_val = math.sin(radians)
    local cos_val = math.cos(radians)
    return {
      x_velocity = velocity * cos_val,
      y_velocity = velocity * sin_val,
      max_height = (velocity * sin_val)^2 / (2 * 9.81)
    }
  end
]])

macro_expand when(not FEATURES.hardware_float or USE_SOFTWARE_FLOAT, [[
  -- Software floating-point implementation using lookup tables
  function calculate_trajectory(angle, velocity)
    local idx = angle % 360
    local sin_val = SIN_TABLE[idx]
    local cos_val = COS_TABLE[idx]
    return {
      x_velocity = (velocity * cos_val) >> 8, -- Fixed-point math
      y_velocity = (velocity * sin_val) >> 8,
      max_height = ((velocity * sin_val) * (velocity * sin_val)) / (2 * 981)
    }
  end
]])
```

#### 4.3.3 Debug vs. Release Code

```lua
macro_expand when(DEBUG, [[
  function assert(condition, message)
    if not condition then
      fmt.Printf("[ASSERT] %s\n", message)
      -- Store assertion failure information in EEPROM for post-crash analysis
      store_crash_info(message)
      sys.Exit(1)
    end
  end
  
  function debug_print(format, ...)
    fmt.Printf("[DEBUG] " .. format .. "\n", ...)
  end
  
  function trace_execution(func_name)
    debug_print("Entering function: %s", func_name)
    -- Increment function call counter in debug memory region
    increment_call_counter(func_name)
    return function()
      debug_print("Exiting function: %s", func_name)
    end
  end
]])

macro_expand unless(DEBUG, [[
  function assert(condition, message)
    -- Empty function in release builds
  end
  
  function debug_print(format, ...)
    -- Empty function in release builds
  end
  
  function trace_execution(func_name)
    return function() end -- No-op function in release builds
  end
]])
```

### 4.4 Advanced Conditional Compilation

#### 4.4.1 Multi-Platform Selection

For code targeting multiple platforms, a selection macro simplifies organization:

```lua
macro_define platform_select(options)
  for platform, code in pairs(options) do
    if TARGET == platform then
      return code
    end
  end
  return options.default or ""
end_macro
```

This enables clean multi-platform implementations:

```lua
-- Define pin control functions for different platforms
macro_expand platform_select({
  AVR = [[
    function set_pin(pin, value)
      local port = pin / 8
      local bit = pin % 8
      if value == 0 then
        PORT[port] = PORT[port] & ~(1 << bit)
      else
        PORT[port] = PORT[port] | (1 << bit)
      end
    end
    
    function configure_pin(pin, mode)
      local port = pin / 8
      local bit = pin % 8
      if mode == PIN_OUTPUT then
        DDR[port] = DDR[port] | (1 << bit)
      else
        DDR[port] = DDR[port] & ~(1 << bit)
        -- Enable pull-up for input pins
        PORT[port] = PORT[port] | (1 << bit)
      end
    end
  ]],
  
  ESP32 = [[
    function set_pin(pin, value)
      gpio_set_level(pin, value)
    end
    
    function configure_pin(pin, mode)
      gpio_config_t io_conf = {}
      io_conf.pin_bit_mask = 1 << pin
      if mode == PIN_OUTPUT then
        io_conf.mode = GPIO_MODE_OUTPUT
      else
        io_conf.mode = GPIO_MODE_INPUT
        io_conf.pull_up_en = 1
      end
      io_conf.pull_down_en = 0
      io_conf.intr_type = GPIO_INTR_DISABLE
      gpio_config(&io_conf)
    end
  ]],
  
  Z80 = [[
    function set_pin(pin, value)
      -- Z80 port-based I/O
      if value == 0 then
        out_port(IO_PORT, in_port(IO_PORT) & ~(1 << pin))
      else
        out_port(IO_PORT, in_port(IO_PORT) | (1 << pin))
      end
    end
    
    function configure_pin(pin, mode)
      if mode == PIN_OUTPUT then
        out_port(IO_DIR, in_port(IO_DIR) | (1 << pin))
      else
        out_port(IO_DIR, in_port(IO_DIR) & ~(1 << pin))
      end
    end
  ]],
  
  default = [[
    function set_pin(pin, value)
      -- Generic implementation for unknown platforms
      io.digitalWrite(pin, value)
    end
    
    function configure_pin(pin, mode)
      if mode == PIN_OUTPUT then
        io.pinMode(pin, io.OUTPUT)
      else
        io.pinMode(pin, io.INPUT)
      end
    end
  ]]
})
```

#### 4.4.2 Module Configuration Attributes

Module-level attributes can be defined to control module-wide compilation parameters:

```lua
macro_define attribute(name, value)
  if not MODULE_ATTRS then
    MODULE_ATTRS = {}
  end
  MODULE_ATTRS[name] = value
  return ""  -- Returns nothing in the output code
end_macro

macro_define if_attribute(name, value, code)
  if MODULE_ATTRS and MODULE_ATTRS[name] == value then
    return code
  else
    return ""
  end
end_macro

macro_define get_attribute(name, default)
  if MODULE_ATTRS and MODULE_ATTRS[name] ~= nil then
    return MODULE_ATTRS[name]
  else
    return default
  end
end_macro
```

These macros create a module-level attribute system for configuration:

```lua
-- Set memory model for this module
macro_expand attribute("memory_model", "tiny")
macro_expand attribute("max_buffer_size", 64)
macro_expand attribute("use_static_allocation", true)

-- Use attributes to control implementation details
macro_expand if_attribute("memory_model", "tiny", [[
  -- Constants optimized for tiny memory model
  local MAX_BUFFER_SIZE = macro_expand get_attribute("max_buffer_size", 32)
  local MAX_ELEMENTS = 16
  
  function allocate(size)
    if size > MAX_BUFFER_SIZE then
      return nil  -- Too large for tiny memory model
    end
    return tiny_heap_alloc(size)
  end
]])

macro_expand if_attribute("use_static_allocation", true, [[
  -- Static buffer allocation
  local buffers = {}
  for i = 1, 4 do
    buffers[i] = {
      data = {},
      size = 0,
      in_use = false
    }
  end
  
  function allocate_buffer(size)
    for i = 1, #buffers do
      if not buffers[i].in_use and size <= macro_expand get_attribute("max_buffer_size", 32) then
        buffers[i].in_use = true
        buffers[i].size = size
        return buffers[i].data
      end
    end
    return nil  -- No buffers available
  end
  
  function free_buffer(buf)
    for i = 1, #buffers do
      if buffers[i].data == buf then
        buffers[i].in_use = false
        return true
      end
    end
    return false  -- Buffer not found
  end
]])
```

#### 4.4.3 Feature-Based Implementation Selection

For memory-sensitive environments, we can conditionally include only the features needed:

```lua
macro_define feature_implementation(feature_name, implementations)
  if FEATURES[feature_name] then
    return implementations.enabled or ""
  else
    return implementations.disabled or ""
  end
end_macro
```

Usage example:

```lua
-- Conditionally include JSON parsing capabilities
macro_expand feature_implementation("json_support", {
  enabled = [[
    function parse_json(json_string)
      -- Full JSON parser implementation
      local result = {}
      local pos = 1
      local len = #json_string
      
      -- Parser state machine
      while pos <= len do
        -- Complex parsing logic for complete JSON support
        -- ...
      end
      
      return result
    end
    
    function stringify_json(data)
      -- Full JSON serialization implementation
      -- ...
    end
  ]],
  
  disabled = [[
    function parse_json(json_string)
      -- Simple key-value parser for minimal JSON support
      -- Only handles flat objects with string values
      local result = {}
      
      -- Extract key-value pairs with simple pattern matching
      for key, value in json_string:gmatch('"([^"]+)":"([^"]+)"') do
        result[key] = value
      end
      
      return result
    end
    
    function stringify_json(data)
      -- Minimal serialization for flat objects
      local parts = {}
      for k, v in pairs(data) do
        if type(v) == "string" then
          table.insert(parts, string.format('"%s":"%s"', k, v))
        end
      end
      return "{" .. table.concat(parts, ",") .. "}"
    end
  ]]
})
```

### 4.5 Combining Code Generation with Conditional Compilation

One of the most powerful aspects of this approach is the ability to combine conditional compilation with code generation:

```lua
macro_define generate_pin_functions(pins)
  local result = ""
  
  -- Generate functions for each pin
  for pin_name, pin_info in pairs(pins) do
    local pin_num = pin_info.pin
    local pin_type = pin_info.type or "digital"
    
    -- Generate platform-specific digital pin functions
    if pin_type == "digital" then
      result = result .. macro_expand(when(TARGET == "AVR", string.format([[
        function set_%s(value)
          if value == 0 then
            PORTA &= ~(1 << %d)
          else
            PORTA |= (1 << %d)
          end
        end
        
        function read_%s()
          return (PINA & (1 << %d)) != 0
        end
      ]], pin_name, pin_num, pin_num, pin_name, pin_num)))
      
      result = result .. macro_expand(when(TARGET == "ESP32", string.format([[
        function set_%s(value)
          gpio_set_level(%d, value)
        end
        
        function read_%s()
          return gpio_get_level(%d)
        end
      ]], pin_name, pin_num, pin_name, pin_num)))
    end
    
    -- Generate platform-specific analog pin functions
    if pin_type == "analog" and pin_info.adc_channel then
      result = result .. macro_expand(when(TARGET == "AVR", string.format([[
        function read_analog_%s()
          ADMUX = (ADMUX & 0xF0) | %d  -- Set ADC channel
          ADCSRA |= (1 << ADSC)        -- Start conversion
          while (ADCSRA & (1 << ADSC)) -- Wait for completion
          return ADC                    -- Return result
        end
      ]], pin_name, pin_info.adc_channel)))
      
      result = result .. macro_expand(when(TARGET == "ESP32", string.format([[
        function read_analog_%s()
          adc_reading_t reading
          adc_oneshot_read_channel(adc_handle, %d, &reading)
          return reading
        end
      ]], pin_name, pin_info.adc_channel)))
    end
  end
  
  return result
end_macro

-- Usage
macro_expand generate_pin_functions({
  led = { pin = 13, type = "digital" },
  button = { pin = 7, type = "digital" },
  temperature = { pin = 5, type = "analog", adc_channel = 2 }
})
```

This powerful combination enables generating customized code for specific platforms from a single pin definition table.

## 5. Comparison with Other Languages

UAL's approach to conditional compilation draws inspiration from several languages while maintaining its own simplicity and consistency:

### 5.1 Elixir-Like Integration

Like Elixir, UAL uses the language's own constructs instead of a separate preprocessor dialect. Conditionals feel like regular code, and the same language semantics apply.

**Elixir:**
```elixir
if target() == :avr do
  def init_timers do
    # AVR-specific implementation
  end
end
```

**UAL (Proposed):**
```lua
macro_expand when(TARGET == "AVR", [[
  function init_timers()
    -- AVR-specific implementation
  end
]])
```

Both approaches avoid introducing a separate preprocessor language, leading to better readability and maintainability.

### 5.2 Rust-Like Feature Flags

Similar to Rust's feature system, UAL supports granular feature toggling for including or excluding functionality.

**Rust:**
```rust
#[cfg(feature = "hardware_float")]
fn calculate(x: f32) -> f32 {
  x * x.sin()
}

#[cfg(not(feature = "hardware_float"))]
fn calculate(x: f32) -> f32 {
  x * SIN_TABLE[(x as usize) % 360]
}
```

**UAL (Proposed):**
```lua
macro_expand when(FEATURES.hardware_float, [[
  function calculate(x)
    return x * math.sin(x)
  end
]])

macro_expand unless(FEATURES.hardware_float, [[
  function calculate(x)
    return x * SIN_TABLE[x % 360]
  end
]])
```

Both approaches allow for fine-grained control over which features are included at compile time.

### 5.3 Unlike C Preprocessor

UAL avoids the pitfalls of C's `#ifdef`/`#endif` approach, which often leads to hard-to-read nested conditions and challenging maintenance.

**C:**
```c
#ifdef AVR
void init_timers(void) {
  TCCR0A = 0x83;
  TCCR0B = 0x04;
  #ifdef ENABLE_PWM
    TCCR0A |= (1 << COM0A1);
    #if F_CPU == 16000000
      OCR0A = 128;
    #else
      OCR0A = 64;
    #endif
  #endif
}
#endif
```

**UAL (Proposed):**
```lua
macro_expand when(TARGET == "AVR", [[
  function init_timers()
    TCCR0A = 0x83
    TCCR0B = 0x04
    
    macro_expand when(FEATURES.enable_pwm, [[
      TCCR0A = TCCR0A | (1 << COM0A1)
      
      local ocr_value = 64
      if CPU_FREQ == 16000000 then
        ocr_value = 128
      end
      OCR0A = ocr_value
    ]])
  end
]])
```

UAL's approach is more readable and maintainable, especially for complex conditional compilation scenarios.

### 5.4 Unlike Go's File-Level Approach

While Go uses file-level build tags, UAL allows for more fine-grained control within files, reducing code duplication.

**Go (file-level approach):**
```go
// uart_avr.go
//go:build avr
// +build avr

package hardware

func InitUART() {
  // AVR-specific UART initialization
}

// uart_esp32.go
//go:build esp32
// +build esp32

package hardware

func InitUART() {
  // ESP32-specific UART initialization
}
```

**UAL (Proposed, within a single file):**
```lua
macro_expand when(TARGET == "AVR", [[
  function init_uart()
    -- AVR-specific UART initialization
  end
]])

macro_expand when(TARGET == "ESP32", [[
  function init_uart()
    -- ESP32-specific UART initialization
  end
]])
```

UAL's approach reduces the need for multiple files when the conditional parts are small or numerous.

## 6. Implementation Considerations

### 6.1 Compilation Process

The conditional compilation would be integrated into the existing macro processing phase:

1. Parse the source file and identify macro definitions and expansions
2. Initialize the compile-time environment with platform and feature flags
3. Process each macro expansion, with special handling for conditional macros
4. Output the final processed source code with conditional sections included or excluded
5. Compile the processed source code normally

### 6.2 Error Handling

Special attention must be paid to error handling within conditional blocks:

1. **Syntax Errors**: Syntax in excluded blocks should not cause compilation failures
2. **Type Checking**: Conditional blocks should be type-checked even if not included
3. **Error Messages**: Error messages should include information about the conditional context

### 6.3 Source Mapping

To aid debugging, the compiler should maintain source mapping information:

1. Track which lines in the output correspond to which conditional blocks
2. Include source location information in error messages
3. Optionally generate annotated output showing which conditional blocks were included

### 6.4 Scope Considerations

Local variables and functions defined in conditional blocks have some special considerations:

1. Variables defined in excluded blocks don't exist at runtime
2. Variables with the same name in mutually exclusive blocks can coexist
3. Functions defined in conditional blocks must handle all their references conditionally

## 7. Limitations and Future Directions

### 7.1 Current Limitations

1. **No Runtime Configuration**: All conditions must be resolvable at compile time
2. **No Dynamic Feature Detection**: Features must be explicitly declared, not detected at runtime
3. **Limited AST Manipulation**: The macro system doesn't currently support complex AST transformations

### 7.2 Future Directions

1. **IDE Integration**: Tools to visualize and navigate conditional code blocks
2. **Conditional Type Checking**: More sophisticated type checking across conditional boundaries
3. **Feature Dependencies**: Declaration of feature dependencies to avoid incompatible feature sets
4. **Partial Evaluation**: Allow certain complex expressions to be partially evaluated at compile time

## 8. Conclusion

The proposed conditional compilation system for UAL leverages the language's existing macro capabilities to provide a clean, consistent, and powerful approach to platform-specific code and feature toggling. This approach maintains UAL's focus on simplicity and embedded systems suitability while avoiding the pitfalls of traditional preprocessor directives.

By keeping conditional compilation within the language rather than adding a separate preprocessor layer, UAL code remains more readable, maintainable, and integrated. The combination of conditional compilation with code generation provides particularly powerful capabilities for cross-platform embedded development.

This proposal recommended adoption of this approach for UAL 1.4, along with standardization of the compile-time variable names and conditional macro interfaces.