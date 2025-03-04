# ual Primer for Embedded Systems Programmers

## Introduction

As an embedded systems programmer, you know the challenges: constrained resources, diverse hardware targets, and the eternal struggle between safety and efficiency. **ual** was designed with these constraints in mind, offering memory safety guarantees without runtime overhead while maintaining the direct hardware access you need.

## Core Features for Embedded Development

### Zero Runtime Overhead Safety

ual provides compile-time safety guarantees with no runtime cost, similar to Rust's approach but with a distinct stack-based model:

```lua
-- Type safety through typed stacks
@Stack.new(Integer): alias:"i"
@Stack.new(Float): alias:"f"

-- Memory safety through ownership (proposed in 1.5)
@Stack.new(Resource, Owned): alias:"dev" 
@dev: push(open_device(0x40))

-- Error propagation without exceptions
@error > function read_register(addr)
  if addr > MAX_ADDR then
    @error > push("Invalid address")
    return nil
  end
  return memory_read(addr)
end
```

### Direct Hardware Access

ual excels at hardware-oriented programming with built-in binary/hex literals and bitwise operations:

```lua
-- Define register addresses as constants
PORT_A_DATA = 0x1A
PORT_A_DIR = 0x1B

-- Set pin 5 as output
function configure_pin5_output()
  -- Read current direction register
  local dir = memory_read_byte(PORT_A_DIR)
  
  -- Set bit 5 (make pin 5 output)
  dir = dir | (1 << 5)  -- Bitwise OR with shifted bit
  
  -- Write back to direction register
  memory_write_byte(PORT_A_DIR, dir)
end

-- Toggle pin 5
function toggle_pin5()
  -- Read current data register
  local data = memory_read_byte(PORT_A_DATA)
  
  -- Toggle bit 5
  data = data ^ (1 << 5)  -- Bitwise XOR with shifted bit
  
  -- Write back to data register
  memory_write_byte(PORT_A_DATA, data)
end
```

### Conditional Compilation

ual's macro system (proposed in 1.4) enables clean cross-platform code:

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

## Stack-Based Programming for Embedded Systems

While stack-based languages aren't new to embedded programming (Forth has been used for decades), ual offers a modern approach with type safety and improved readability.

### Efficient Register Manipulation

Stack operations can be particularly elegant for register manipulation:

```lua
-- Using stacked mode with colon syntax
function set_register_bits(reg_addr, start_bit, num_bits, value)
  @Stack.new(Integer): alias:"i"
  
  -- Calculate mask and shifted value
  @i: push(1) push(num_bits) shl push:1 sub  -- Create mask like (1 << num_bits) - 1
  @i: dup push(value) and push(start_bit) shl  -- (value & mask) << start_bit
  local shifted_value = i.pop()
  
  @i: push(start_bit) shl  -- Shift mask to final position
  local positioned_mask = i.pop()
  
  -- Read current register value
  local reg_value = memory_read_byte(reg_addr)
  
  -- Clear bits in mask position, then set new bits
  reg_value = (reg_value & ~positioned_mask) | shifted_value
  
  -- Write back
  memory_write_byte(reg_addr, reg_value)
end
```

### Memory Efficient Programming

Stack-based operations often result in more compact code - important for constrained environments:

```lua
-- Using stack operations for signal processing
function process_adc_reading()
  @Stack.new(Integer): alias:"i"
  @Stack.new(Float): alias:"f"
  
  -- Read ADC value and convert to voltage
  @i: push(read_adc())           -- Raw ADC value (0-1023)
  @f: <i                         -- Convert to float
  @f: push:1023.0 div push:3.3 mul  -- Convert to voltage
  
  -- Apply offset and scaling corrections
  @f: push:0.05 sub push:1.02 mul   -- Adjust for sensor characteristics
  
  return f.pop()
end
```

## Resource Management

ual's proposed ownership system (1.5) makes resource management both safe and efficient:

```lua
function sample_sensor()
  -- Create an owned resource stack for the sensor
  @Stack.new(Sensor, Owned): alias:"s"
  @s: push(open_sensor(TEMP_SENSOR_ID))
  
  -- Take a reading with mutable access
  @Stack.new(Sensor, Mutable): alias:"sm"
  @sm: <:mut s                      -- Borrow mutably
  @sm: push(sm.pop().start_sample())  -- Start sampling
  
  sys.Sleep(10)  -- Wait for sample
  
  -- Read the result with borrowed access
  @Stack.new(Sensor, Borrowed): alias:"sb"
  @sb: <<s                         -- Borrow immutably
  @sb: push(sb.pop().read_value())  -- Read the value
  
  local reading = sb.pop()
  
  -- Sensor automatically closed when 's' goes out of scope
  return reading
end
```

## Error Handling Without Exceptions

The error stack mechanism (proposed in 1.4) is particularly valuable for embedded systems where exceptions are often unavailable or undesirable:

```lua
@error > function init_periph()
  if not init_gpio() then
    @error > push("GPIO initialization failed")
    return false
  end
  
  if not init_spi() then
    @error > push("SPI initialization failed")
    return false
  end
  
  return true
end

function setup()
  if init_periph() then
    led_blink(3)  -- Success indicator
  else if @error > depth() > 0 then
    err = @error > pop()
    log_error(err)
    led_error_code(1)  -- Error indicator
  end
end
```

## Compile-Time Computation

ual's macro system (proposed in 1.4) can perform calculations at compile time:

```lua
-- Generate sine lookup table at compile time
macro_define generate_sin_table(size)
  local result = "local SIN_TABLE = {\n"
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

-- Create a 256-entry sine table
macro_expand generate_sin_table(256)

-- Later in the code, use the table instead of calculating sin values
function fast_sin(angle_degrees)
  local index = angle_degrees % 360
  local table_index = math.floor(index * 256 / 360)
  return SIN_TABLE[table_index] / 32767.0
end
```

## A Complete Example: I2C Sensor Interface

Here's a more comprehensive example showing how ual can be used to create a clean, safe I2C sensor interface:

```lua
package sensors

import "i2c"
import "sys"

-- Constants
BME280_ADDR = 0x76
BME280_REG_ID = 0xD0
BME280_ID = 0x60
BME280_REG_CTRL_MEAS = 0xF4
BME280_REG_CONFIG = 0xF5
BME280_REG_TEMP = 0xFA

-- Initialize sensor
@error > function bme280_init(i2c_bus)
  @Stack.new(I2C, Owned): alias:"bus"
  @bus: push(i2c_bus)
  
  -- Check device ID
  @Stack.new(Integer, Owned): alias:"i"
  @i: push(BME280_ADDR) push(BME280_REG_ID) push(1)
  
  success, id = bus.peek().read_reg_byte(i.pop(), i.pop())
  if not success or id != BME280_ID then
    @error > push("BME280 not found or wrong ID")
    return false
  end
  
  -- Configure sensor (normal mode, 16x oversampling, 500ms sampling)
  @i: push(BME280_ADDR) push(BME280_REG_CTRL_MEAS) push(0xB7)
  success = bus.peek().write_reg_byte(i.pop(), i.pop(), i.pop())
  if not success then
    @error > push("Failed to configure BME280")
    return false
  end
  
  @i: push(BME280_ADDR) push(BME280_REG_CONFIG) push(0xA0)
  success = bus.peek().write_reg_byte(i.pop(), i.pop(), i.pop())
  if not success then
    @error > push("Failed to configure BME280")
    return false
  end
  
  return true
end

-- Read temperature
@error > function bme280_read_temp(i2c_bus)
  @Stack.new(I2C, Borrowed): alias:"bus"
  @bus: push(i2c_bus)
  
  @Stack.new(Integer, Owned): alias:"i"
  @i: push(BME280_ADDR) push(BME280_REG_TEMP) push(3)
  
  -- Read raw temperature (3 bytes)
  success, msb, lsb, xlsb = bus.peek().read_reg_bytes(i.pop(), i.pop(), i.pop())
  
  if not success then
    @error > push("Failed to read BME280 temperature")
    return 0
  end
  
  -- Convert to 20-bit value
  @i: push(msb) push(8) shl push(lsb) or push(8) shl push(xlsb) push(4) shr or
  raw_temp = i.pop()
  
  -- Calculate actual temperature (simplified conversion)
  @Stack.new(Float, Owned): alias:"f"
  @f: <i push:100.0 div
  
  return f.pop()
end

-- Usage example
function main()
  -- Initialize I2C bus
  @Stack.new(I2C, Owned): alias:"i2c"
  @i2c: push(i2c.init(0, 100000))  -- Bus 0, 100kHz
  
  -- Initialize sensor
  if not bme280_init(i2c.peek()) then
    if @error > depth() > 0 then
      fmt.Printf("Init error: %s\n", @error > pop())
    end
    return 1
  end
  
  -- Main loop
  while_true(true)
    -- Read temperature
    temp = bme280_read_temp(i2c.peek())
    if @error > depth() > 0 then
      fmt.Printf("Read error: %s\n", @error > pop())
    else
      fmt.Printf("Temperature: %.2f°C\n", temp)
    end
    
    sys.Sleep(1000)
  end_while_true
  
  return 0
end
```

## Why Choose ual for Embedded Systems?

1. **Safety without Cost**: Compile-time guarantees with zero runtime overhead
2. **Direct Hardware Access**: Binary/hex literals and bitwise operations
3. **Resource Management**: Stack-based ownership for deterministic cleanup
4. **Cross-Platform**: Conditional compilation for targeting diverse hardware
5. **Compile-Time Computation**: Generate lookup tables and configurations at compile time
6. **Error Handling**: No exceptions, no unwinding, just predictable error propagation

## Next Steps

To explore more about ual for embedded systems:
- Learn about the [error stack system](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-error-stack.md) for robust error handling
- Dive into [conditional compilation](https://github.com/ha1tch/ual/blob/main/spec/ual-1.4-PROPOSAL-conditional-compilation-02.md) for cross-platform code
- Explore the [stack-based ownership model](https://github.com/ha1tch/ual/blob/main/spec/ual-1.5-PROPOSAL-ownership-system.md) for safe resource management

ual combines the efficiency of low-level programming with the safety of modern languages, making it an excellent choice for your next embedded project.