# ual-to-TinyGo Compiler

This project implements a compiler that translates ual 1.1 code into TinyGo, targeting small embedded platforms.

## What is ual?

ual (micro assembly language) is a high-level, Lua-like language designed for use on minimal or retro platforms with small virtual machines or embedded hardware. It features:

- Stack-based operations for arithmetic and memory access
- Lua-style control flow, scoping, and data structures
- Go-like package conventions (uppercase = exported, lowercase = private)
- Multiple returns and flexible for loops
- Binary/hexadecimal numeric literals
- Bitwise operators for hardware programming

## Project Components

- **Lexer**: Tokenizes ual source code
- **Parser**: Builds an abstract syntax tree (AST) from tokens
- **Code Generator**: Transforms the AST into TinyGo code
- **Standard Library**: Implementations of the ual standard packages
  - `con`: Console operations
  - `fmt`: String formatting
  - `sys`: System-level operations
  - `io`: Digital I/O for hardware pins
  - `str`: String manipulation
  - `math`: Basic math functions

## Requirements

- Go (1.18 or later)
- TinyGo (0.27.0 or later)

## Installation

```bash
# Clone this repository
git clone https://github.com/yourusername/ual-compiler
cd ual-compiler

# Install dependencies
go mod tidy

# Make the build script executable
chmod +x build.sh
```

## Usage

### Compiling a ual Program

```bash
# Compile a specific ual file
./build.sh examples/blinky.ual

# Compile all examples
./build.sh
```

The compiled hex files will be placed in the `build` directory.

### Flashing to a Microcontroller

Once you have the hex file, you can flash it to your microcontroller using the appropriate method for your hardware:

```bash
# Example for Arduino
tinygo flash -target arduino build/blinky.hex
```

## Example Programs

- **blinky.ual**: Basic LED blinking example
- Add your own examples in the `examples` directory!

## Extending the Compiler

### Adding New Syntax Features

1. Update the lexer in `lexer/lexer.go` to recognize new tokens
2. Modify the parser in `parser/parser.go` to handle the new syntax
3. Enhance the code generator in `codegen/codegen.go` to output the correct TinyGo code

### Adding Standard Library Functions

Modify or extend the implementation files in each package directory.

## License

MIT License - See LICENSE file for details.