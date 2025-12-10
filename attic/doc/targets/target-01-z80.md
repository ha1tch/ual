# Assessing the Z80 on ZX Spectrum as a ual Target

## Z80 vs 6502: A More Promising Architecture

The Z80 presents a significantly better target than the 6502 for implementing ual, especially on the ZX Spectrum family:

### Z80 Advantages Over 6502
- **Richer register set**: 8 main registers (A, B, C, D, E, F, H, L), plus alternate set, IX/IY index registers
- **Register pairs**: BC, DE, HL can be used as 16-bit values or address pointers
- **Flexible stack**: Hardware stack with push/pop for registers and register pairs
- **Movable stack pointer**: SP can be positioned anywhere in RAM (not fixed like 6502)
- **Block instructions**: LDIR, LDDR, etc. for efficient memory operations
- **Bit manipulation**: Dedicated bit set/reset/test instructions
- **Larger instruction set**: ~700 opcodes vs ~150 for 6502

These characteristics make implementing ual's stack operations more straightforward and efficient.

## ZX Spectrum Hardware Considerations

### Memory Configuration
- **48K Spectrum**: 16KB ROM + 48KB RAM
- **128K Spectrum**: 16KB ROM + 128KB RAM with bank switching
- **Memory layout**: ROM at 0x0000, display file at 0x4000, attributes at 0x5800
- **User RAM**: Starts around 0x5CCB in 48K model

### Hardware Limitations
- **Contended memory**: CPU must wait for ULA during screen refresh
- **No hardware FPU**: Floating-point operations must be software-implemented
- **Limited I/O**: Basic ULA for display, sound, and keyboard
- **Banking (128K only)**: 16KB banks switched in/out at 0xC000

## Implementation Viability

### 1. Stack-Based Operations

The Z80's register structure enables efficient implementation of ual's stack paradigm:
- Register pairs (BC, DE, HL) could serve as stack pointers
- Index registers (IX, IY) provide indexed addressing useful for stack management
- Hardware stack operations (PUSH/POP) can optimize certain operations

```assembly
; Example of implementing stack operation in Z80
; HL = pointer to data stack
; BC, DE = temporary register pairs

; ual: push(42)
LD BC, 42         ; Value to push
LD (HL), C        ; Store low byte
INC HL
LD (HL), B        ; Store high byte
INC HL

; ual: dup()
DEC HL
DEC HL            ; Point to top item
LD C, (HL)        ; Load low byte
INC HL
LD B, (HL)        ; Load high byte
INC HL            ; Restore pointer
LD (HL), C        ; Store low byte again
INC HL
LD (HL), B        ; Store high byte again
INC HL
```

### 2. Memory Management on Spectrum

**48K Spectrum**:
- ual runtime would consume ~8-12KB
- Each stack might require 1-4KB
- Leaving ~20-30KB for user program
- **Conclusion**: Tight but workable for smaller programs

**128K Spectrum**:
- Runtime can reside in a fixed bank
- Stacks can use dedicated banks
- User program gets more space
- **Conclusion**: Much more viable, supporting larger programs

### 3. Type System Implementation

Typed stacks are implementable with reasonable overhead:
- Each value would need a type tag (1 byte)
- Type checking mostly at compile time
- Conversions between types would be software-implemented

The `bring_<type>` operation could be implemented as specialized routines, though string-to-number conversions would be costly.

### 4. Performance Realities

Performance would be adequate for many use cases:
- Basic stack operations: 20-100 cycles
- Type conversions: 100-500 cycles
- String operations: 500-5000 cycles
- Floating point: 1000-10000 cycles

This would result in usable performance for many applications, though complex numerical code would be quite slow.

### 5. Precedents on ZX Spectrum

Several sophisticated languages run on the Spectrum:
- **HiSoft Pascal**: Compiled Pascal with decent performance
- **Z88DK C Compiler**: Full C implementation for Z80
- **White Lightning**: FORTH implementation
- **Beta BASIC**: Extended BASIC with structured programming

These demonstrate that implementing high-level language constructs is viable.

## Feature-by-Feature Viability

| Feature | Viability | Notes |
|---------|-----------|-------|
| Basic stack operations | High | Direct mapping to Z80 code |
| Multiple typed stacks | Medium | Memory overhead but feasible |
| Integer operations | High | Native support |
| Floating point | Medium | Slow but possible using ROM routines |
| String handling | Medium | Memory-intensive but feasible |
| Error stack | Medium | Implementation possible |
| Bitwise operations | High | Direct Z80 bit instructions |
| Direct memory access | High | Straightforward implementation |
| Stack as first-class | Medium | Possible with overhead |
| Ownership system | Low | Complex to implement efficiently |

## Implementation Strategy for ZX Spectrum

1. **Target 128K Spectrum primarily**:
   - Use banking for runtime library
   - Support 48K with reduced feature set

2. **Optimize common operations**:
   - Hand-coded Z80 routines for stack operations
   - Use ROM routines where beneficial (math, etc.)

3. **Memory management**:
   - Virtual memory techniques with bank switching
   - Paging for larger programs

4. **User interface**:
   - Custom display routines using Spectrum's character set
   - Leverage ROM routines for input handling

## Conclusion

The Z80 on ZX Spectrum represents a **viable but challenging target** for ual:

- **Better than 6502**: More registers, better stack support, more memory
- **Memory constraints**: Workable on 48K, comfortable on 128K
- **Performance challenges**: Complex features would be slow
- **Feature subset**: Likely need to implement a reduced feature set

A ual implementation for ZX Spectrum would be an excellent case of retrofuturism - bringing modern language safety concepts to vintage hardware while respecting its constraints. The 128K models in particular offer enough resources to make a meaningful implementation feasible while maintaining the authentic retro computing experience.
