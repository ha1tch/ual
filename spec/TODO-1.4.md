# TODO 1.4: Prioritized Roadmap After ual 1.3 and Proposed Enhancements

This document outlines the current status of ual development priorities, taking into account both the completed features in ual 1.3 and the proposed enhancements for ual 1.4.

## Recently Completed

1. **Multiple stacks** - ✓ FULLY ADDRESSED. First-class stack objects with `Stack.new()` and predefined `dstack`/`rstack`.
    
2. **Switch statement** - ✓ FULLY ADDRESSED. The 1.3 spec adds a comprehensive `switch_case` construct.
    
3. **Stacked mode syntax** - ✓ FULLY ADDRESSED. The new stacked mode with `>` prefix and `@stack >` selection syntax.

4. **Error handling** - ✓ FULLY ADDRESSED. The `.consider` pattern in 1.3 and significantly enhanced with the proposed `@error` stack system for 1.4, providing compile-time verification.

## High Priority - Core Language Features

1. **Debugging facilities** - PARTIALLY ADDRESSED by the `@error` stack proposal, but would benefit from additional tools for stack state inspection and execution tracing.
    
2. **Fixed-point or floating-point arithmetic** - NEEDED. Support for non-integer calculations for applications requiring decimal calculations.
    
3. **Conditional compilation** - SUBSTANTIALLY ADDRESSED in the proposed macro system for 1.4, which uses ual's native syntax rather than introducing a separate preprocessor language.
    
4. **Standard library expansion** - PARTIALLY ADDRESSED but needs further development for common embedded tasks.
    
5. **String manipulation operations** - PARTIALLY ADDRESSED in the `str` package but would benefit from expansion.
    
## High Priority - Embedded Systems Support

1. **Interrupt handling** - NEEDED. Critical for responsive embedded applications.
    
2. **Inline assembly** - NEEDED. Essential for performance-critical code sections and direct hardware access.
    
3. **Binary inclusion facilities** - NEEDED. Important for embedding resources like graphics, fonts, and lookup tables.
    
4. **Configurable compilation targets** - PARTIALLY ADDRESSED in the 1.4 conditional compilation proposal through TARGET variables and platform selection, but still requires complete toolchain configuration.

## Medium Priority

1. **Optional type annotations** - NEEDED. Would improve error checking while maintaining flexibility.
    
2. **Compiler optimization hints** - NEEDED. Valuable for performance-critical applications.
    
3. **Memory usage annotations** - NEEDED. Would help with resource management in constrained environments.
    
4. **Testing framework** - NEEDED. Would improve development experience, potentially leveraging the `@error` stack.

## New Priorities for Future Versions

1. **Macro system** - PROPOSED for 1.4. Would address conditional compilation, code generation, and platform-specific adaptations.
    
2. **Standard stack patterns library** - PROPOSED. Common stack manipulation patterns that could be imported and used with the stack objects.
    
3. **Stack verification** - PARTIALLY ADDRESSED through `@error` stack compile-time checking, but could be extended to general stack operations.
    
## Implementation Priority for ual 1.4

Based on the overall roadmap, the following items appear most critical for immediate implementation in ual 1.4:

1. **Error handling with `@error` stack** - Provides unified error management with compile-time guarantees and no runtime overhead.

2. **Macro system** - Enables conditional compilation, code generation, and platform-specific adaptations.

3. **Fixed-point arithmetic** - Adds essential numeric capabilities for embedded applications.

4. **Basic interrupt handling** - Enables responsive embedded applications.

These four features would make ual 1.4 a significant advancement for embedded development, balancing language improvements with practical capabilities for real-world use cases.

## Longer-Term Vision

Future versions of ual could focus on:

1. Expanding platform-specific optimizations and hardware abstractions
2. Enhancing development tooling for debugging and analysis
3. Growing the standard library for common embedded tasks
4. Adding optional type checking for larger projects
5. Improving resource usage analysis for constrained environments

The overall goal remains to maintain ual's focus on embedded systems while steadily improving its expressiveness, safety, and developer experience.