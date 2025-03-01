# TODO 1.3: Remaining Items After ual 1.3 Improvements

Looking at the previous TODO list in light of the substantial ual 1.3 improvements, here's my assessment of which items are still relevant:

### Addressed in ual 1.3

1. **Multiple stacks** - ✓ FULLY ADDRESSED. The 1.3 spec introduces stacks as first-class objects with `Stack.new()` and predefined `dstack`/`rstack`, allowing for unlimited custom stacks with consistent interfaces.
    
2. **Switch statement** - ✓ ADDRESSED. The 1.3 spec adds a comprehensive `switch_case` construct with multi-value cases and fall-through behavior.
    
3. **Stacked mode syntax** - ✓ ADDRESSED. The new stacked mode with `>` prefix and `@stack >` selection syntax provides a concise way to express stack operations.
    

### Still Relevant and Important

1. **Debugging facilities** - Still needed. Tools for inspecting stack state and tracing execution remain valuable, especially with multiple stack objects.
    
2. **Fixed-point or floating-point arithmetic** - Still needed for applications requiring decimal calculations.
    
3. **String manipulation operations** - Would still benefit from expansion beyond the current `str` package.
    
4. **Structured data serialization** - Still relevant for data storage and communication.
    
5. **Inline assembly** - Still relevant for performance-critical code on specific hardware.
    
6. **Conditional compilation** - Still needed for multi-platform support. (Note: The macro system proposal for 1.4 would address this.)
    
7. **Standard library expansion** - Still valuable for common embedded tasks.
    
8. **Binary inclusion facilities** - Still relevant for embedded graphics, fonts, etc.
    
9. **Interrupt handling** - Critical for many embedded applications.
    

### Moderate Priority

1. **Documentation annotations** - Would improve tooling and documentation, especially for stack effects.
    
2. **Optional type annotations** - Could provide better error checking while maintaining flexibility.
    
3. **Compiler optimization hints** - Valuable for performance-critical applications.
    
4. **Testing framework** - Would improve development experience.
    
5. **Memory usage annotations** - Would help with resource management.
    
6. **Configurable compilation targets** - Would enhance cross-platform development.
    
7. **Concurrency primitives** - Relevant but depends on target platform capabilities.
    

### Lower Priority Given ual's Focus

1. **Hot reload capabilities** - Less common in constrained environments and may not align with ual's core focus.

### New Priorities to Consider for ual 1.4

1. **Macro system** - As proposed in the 1.4 draft, would address conditional compilation, code generation, and platform-specific adaptations.
    
2. **Standard stack patterns library** - Common stack manipulation patterns that could be imported and used with the new stack objects.
    
3. **Improved stack effect documentation** - Given the new stacked mode and first-class stacks, better tools for documenting stack effects would be valuable.
    
4. **Stack verification** - Runtime or compile-time verification of stack depth and types.
    
5. **Integration with hardware description languages** - For FPGA or custom chip development.
    

The ual 1.3 updates have made substantial progress by fully addressing multiple stacks, adding the switch statement, and introducing the concise stacked mode syntax. The first-class stack objects in particular represent a significant evolution that makes the language more flexible and powerful.

The most pressing remaining needs appear to be around debugging facilities, arithmetic beyond integers, and features to support hardware-specific development like interrupt handling and inline assembly. The proposed macro system for 1.4 would address many of these needs, particularly for conditional compilation and platform-specific code.