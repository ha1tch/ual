# ual 1.1 Specification

## 1. Introduction

**ual** is a high-level, Lua-like language designed for use on minimal or retro platforms with small virtual machines or small embedded hardware. The language emphasizes:

1. **Stack-based operations** for arithmetic and memory access 
2. **Lua-style** control flow, scoping, and data structures (arrays, tables, strings).  
3. **Go-like package** conventions (uppercase = exported, lowercase = private).  
4. **Multiple returns** and flexible for loops (numeric and iterator-based).  
5. **Binary / Hexadecimal** numeric literals for easy hardware-oriented programming.  
6. **Bitwise operators** for direct manipulation of registers and masks.

This document outlines the **lexical structure**, **grammar**, **scoping rules**, **packages system**, and **operational semantics** of ual 1.1, providing a foundation for compilers or interpreters targeting small virtual machines and embedded platforms.

---

## 2. Lexical Structure

### 2.1 Identifiers

- Must start with a letter (`A–Z` or `a–z`) or underscore (`_`), and can contain letters, digits (`0–9`), and underscores.  
- Examples: `foo`, `myVar_2`, `_helper`.  
- **Case** matters for **export** conventions:
  - **Uppercase first letter** (e.g., `Print`) means an **exported** symbol (accessible outside its package).
  - **Lowercase first letter** (e.g., `parseFormat`) means **private** to the package.

### 2.2 Keywords

```
function, end, if_true, if_false, while_true, return, local, do,
for, in, push, pop, dup, swap, add, sub, mul, div, store, load,
import, package
```

(Other keywords for structure termination like `end_if_true`, `end_while_true` may also be recognized. Implementations can choose how best to parse these.)

### 2.3 Numeric Literals

**ual 1.1** supports **decimal, binary, and hex** integer literals with **case-insensitive** prefixes for binary and hex:

1. **Decimal** (unprefixed): `123`  
2. **Binary**: `0b1010` or `0B1010` (the part after `0b`/`0B` must be `0` or `1`).  
3. **Hex**: `0x1f`, `0X1F`, etc. (the part after `0x`/`0X` can be `0–9`, `A–F`, `a–f`).

Examples:

```lua
local decVal = 123
local binVal = 0b1011        -- decimal 11
local hexVal = 0xAbCd        -- decimal 43981, ignoring overflow rules
```

The implementation may store these in **8-bit, 16-bit, or 32-bit** integers, with overflow wrapping or error conditions depending on the platform.

### 2.4 String Literals

- Single- or double-quoted, e.g. `"Hello"` or `'World'`.  
- Escapes can be defined as needed (e.g., `"\n"`).

### 2.5 Boolean and Nil (Optional)

- `true` or `false` for booleans (optional).
- `nil` for a null value (optional).

### 2.6 Operators and Delimiters

- **Arithmetic**: `+`, `-`, `*`, `/`.  
- **Comparison**: `==`, `!=`, `<`, `>`.  
- **Bitwise**: `&`, `|`, `^`, `<<`, `>>`.  
- **Assignment**: `=`.  
- **Other Delimiters**: parentheses `( )`, braces `{ }`, brackets `[ ]`, commas `,`, etc.

---

## 3. Package Declarations and Imports

### 3.1 Package Declaration

Each **ual** file **must** begin with:

```
package <identifier>
```

- This identifier names the package.  
- Symbols in this file that begin with an **uppercase** letter are **exported** (public).  
- Symbols that begin with a **lowercase** letter are **private** to this package.

### 3.2 Imports

```
import "package_name"
```

- Brings all **uppercase** symbols from the named package into scope as `package_name.Symbol`.
- Packages exist in a flat namespace. If multiple packages share the same name, that’s an error.

Example:

```lua
import "con"
import "fmt"

function main()
  con.Cls()
  fmt.Printf("Hello from ual!\n")
end
```

---

## 4. Grammar

Below is an **EBNF** grammar that integrates packages, imports, numeric literals (decimal/binary/hex), bitwise operators, and standard control structures.

### 4.1 Overall File Structure

```
<program>          ::= <package-decl> { <import-decl> } { <top-level-decl> }

<package-decl>     ::= "package" <identifier>

<import-decl>      ::= "import" <string-literal>

<top-level-decl>   ::= <function-def> | <global-var-decl>
```

A typical ual file:

1. One `package` statement.  
2. Zero or more `import "somePkg"` statements.  
3. Zero or more top-level definitions (functions, global variables).

### 4.2 Global Variables

```
<global-var-decl>  ::= <identifier> "=" <expr>
```

- If `<identifier>` begins with uppercase, it’s **exported**.  
- If it’s lowercase, it’s private to the package.

### 4.3 Functions

```
<function-def>     ::= "function" <identifier> "(" [ <param-list> ] ")" <block> "end"

<param-list>       ::= <identifier> { "," <identifier> }
```

- Similarly, uppercase function name → exported.

### 4.4 Blocks and Statements

```
<block>            ::= { <statement> }

<statement>        ::= <assignment-stmt>
                     | <stack-stmt>
                     | <if-true-stmt>
                     | <if-false-stmt>
                     | <while-true-stmt>
                     | <for-num-stmt>
                     | <for-gen-stmt>
                     | <return-stmt>
                     | <function-call-stmt>
                     | <local-decl-stmt>
                     | <do-block>
                     | <empty>
```

#### 4.4.1 Assignment

```
<assignment-stmt>  ::= <var-list> "=" <expr-list>

<var-list>         ::= <variable> { "," <variable> }
<expr-list>        ::= <expr> { "," <expr> }

<variable>         ::= <identifier>
                     | <index-access>

<index-access>     ::= <expr> "[" <expr> "]"
                     | <expr> "." <identifier> (optional dot syntax)
```

Multiple assignment is supported: `a, b = x, y`.

#### 4.4.2 Stack Operations

```
<stack-stmt>       ::= "push(" <expr> ")"
                     | "pop()"
                     | "dup()"
                     | "swap()"
                     | "add()" | "sub()" | "mul()" | "div()"
                     | "store()" | "load()"
                     | "len()"
```

#### 4.4.3 If-True / If-False

```
<if-true-stmt>     ::= "if_true(" <expr> ")" <block> ("end_if_true")?

<if-false-stmt>    ::= "if_false(" <expr> ")" <block> ("end_if_false")?
```

- Execute `<block>` if `<expr>` is nonzero (if_true) or zero (if_false).

#### 4.4.4 While Loops

```
<while-true-stmt>  ::= "while_true(" <expr> ")" <block> ("end_while_true")?
```

- Repeats as long as `<expr>` is nonzero.

#### 4.4.5 For Loops

```
<for-num-stmt>     ::= "for" <identifier> "=" <expr> "," <expr> [ "," <expr> ]
                       "do" <block> "end"
  -- e.g. for i = 1, 10, 2 do ... end

<for-gen-stmt>     ::= "for" <identifier> "in" <expr> "do" <block> "end"
  -- e.g. for item in someIterator() do ... end
```

#### 4.4.6 Return

```
<return-stmt>      ::= "return" [ <expr-list> ]
```

Allows multiple return values, e.g. `return a, b`.

#### 4.4.7 Function Calls

```
<function-call-stmt> ::= <identifier> "(" [ <arg-list> ] ")"

<arg-list>         ::= <expr> { "," <expr> }
```

#### 4.4.8 Local Declarations

```
<local-decl-stmt>  ::= "local" <identifier> [ "=" <expr> ]
```

#### 4.4.9 Do-Block

```
<do-block>         ::= "do" <block> "end"
```

### 4.5 Expressions

```
<expr> ::= <literal>
         | <variable>
         | <function-call-expr>
         | "(" <expr> ")"
         | <binary-op-expr>
         | <table-constructor>
         | <array-constructor>
```

#### 4.5.1 Binary Operator Expressions

```
<binary-op-expr> ::= <expr> <binary-op> <expr>

<binary-op>      ::= "+" | "-" | "*" | "/" 
                   | "==" | "!=" | "<" | ">"
                   | "&" | "|" | "^"
                   | "<<" | ">>"
```

- This includes arithmetic (`+`, `-`, etc.), comparisons (`==`, `<`), and **bitwise** ops (`&`, `|`, `^`, `<<`, `>>`).
- Some implementations give different **precedences** to these operators; others might treat them uniformly. Parentheses clarify order.

### 4.6 Literals

```
<literal> ::= <number-literal>
            | <string-literal>
            | <bool-literal>
            | <nil-literal>

<number-literal> ::= <decimal-literal> | <binary-literal> | <hex-literal>

<decimal-literal> ::= DIGIT { DIGIT }+          -- e.g. 123
<binary-literal>  ::= ("0b" | "0B") { '0'|'1' }+
<hex-literal>     ::= ("0x" | "0X") { DIGIT | [A-Fa-f] }+
```

- **Case-insensitive** binary/hex prefixes.  
- Hex digits `A–F` or `a–f`.  
- Unprefixed numeric sequences are decimal.  

### 4.7 Data Constructors

#### 4.7.1 Table Constructor

```
<table-constructor> ::= "{" [ <table-field-list> ] "}"

<table-field-list>  ::= <table-field> { "," <table-field> }

<table-field>       ::= <keydef> <expr>

<keydef>           ::= <identifier> "="
                     | "[" <expr> "]" "="
                     | (empty)
```

#### 4.7.2 Array Constructor

```
<array-constructor> ::= "[" <expr-list> "]"
```

---

## 5. Scoping and Export Rules

### 5.1 Lexical Scope

- **Local**: Declared via `local x = ...`; visible until block/function ends.  
- **Global** (package-level): Declared at top level.  
  - **Uppercase** first letter → exported.  
  - **Lowercase** first letter → private.

### 5.2 Packages and Visibility

- One file = one package.  
- `import "pkg"` → gain access to **uppercase** symbols from that package.

---

## 6. Operational Semantics

### 6.1 Stack Statements

- `push(expr)`: Evaluate expr, put result on stack.  
- `pop()`: Remove the top item.  
- `dup()`, `swap()` → manipulate top stack elements.  
- `add()`, `sub()`, etc. → pop two items, combine, push result.  
- `store()`, `load()` → typically `[value, addr]` or `[addr]` on stack.

### 6.2 Control Flow

- `if_true(expr) ... end_if_true`: run block if expr ≠ 0.  
- `if_false(expr)`: run block if expr = 0.  
- `while_true(expr)`: loop until expr = 0.  
- `for i = start, end, step do ... end`: typical numeric loop.  
- `for var in iterator do ... end`: generic loop.  

### 6.3 Multiple Return

- `return a, b, c`.  
- Caller can do `x, y = someFunc(...)`. Extra or missing values are truncated or set to 0/nil.

---

## 7. Packages System Semantics

1. **Single Package Declaration**: `package <name>`.  
2. **Imports**: `import "otherPkg"`.  
   - Gains access to uppercase symbols in `otherPkg`.  
3. **Export vs. Private**:  
   - If an identifier starts uppercase, it’s exposed to importers.  
   - Otherwise, it’s hidden (internal).

---

## 8. Example Program

**File: con.ual** (console I/O)

```lua
package con

-- Exported
function Cls()
  -- Clear screen logic
end

function Print(s)
  -- Print string 's' at current cursor
end
```

**File: fmt.ual** (formatting)

```lua
package fmt

import "con"

function Printf(format, ...)
  local s = Sprintf(format, ...)
  con.Print(s)
end

function Sprintf(format, ...)
  -- parse placeholders (%d, %s, etc.) 
  return "..."
end
```

**File: main.ual**

```lua
package main

import "con"
import "fmt"

function main()
  con.Cls()
  fmt.Printf("Binary: 0B1010 => decimal %d\n", 0b1010)
  fmt.Printf("Hex: 0xFf => %d\n", 0xFf)
end
```

---

## 9. Implementation Guidelines

1. **Parser**: Implement EBNF rules, supporting binary/hex numeric literals (case-insensitive).  
2. **Symbol Resolution**: Maintain a map of packages → uppercase symbols for references like `pkg.FuncName`.  
3. **Code Generation**:  
   - Convert expressions (including bitwise ops) into opcodes, handle numeric literal parsing.  
   - For local variables, track memory offsets or a local stack area.  
   - For arrays/tables, allocate memory structures.  
4. **Operator Precedence**: Either keep everything at one level or define a typical hierarchy (like C).  
5. **Optimization**: Inline short functions, remove redundant stack ops, etc.

---

## 10. Conclusion

**ual 1.1** provides a comprehensive, minimal-yet-powerful language design:

- **Go-like** package and import system for modular code.  
- **Lua-like** syntax for functions, loops, data structures, and scoping.  
- **Binary and hex** numeric literals for hardware-oriented tasks.  
- **Bitwise operators** for register manipulation and masking.  

This specification can be adapted to multiple embedded or retro platforms to produce compact, stack-based programs that are still approachable for developers familiar with higher-level languages.