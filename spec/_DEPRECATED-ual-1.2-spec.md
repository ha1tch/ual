Below is the complete **ual 1.2 Specification** document in a new canvas, updated to include:

- Support for both Lua and C-style comments.  
- The new result handling syntactic sugar (with concise and full function literal forms).  
- A revised set of Forth-inspired stack manipulation keywords with names that start with alphabetic characters to avoid parsing issues.

---

# ual 1.2 Specification

## 1. Introduction

**ual** is a high-level, Lua-like language designed for use on minimal or retro platforms with small virtual machines or embedded hardware. The language emphasizes:

1. **Stack-based operations** for arithmetic and memory access  
2. **Lua-style** control flow, scoping, and data structures (arrays, tables, strings)  
3. **Go-like package** conventions (uppercase = exported, lowercase = private)  
4. **Multiple returns** and flexible for loops (numeric and iterator-based)  
5. **Binary / Hexadecimal** numeric literals for easy hardware-oriented programming  
6. **Bitwise operators** for direct manipulation of registers and masks

This document outlines the **lexical structure**, **grammar**, **scoping rules**, **packages system**, **operational semantics**, and (new in 1.2) **result handling syntactic sugar** of ual, providing a foundation for compilers or interpreters targeting small virtual machines and embedded platforms.

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
import, package, drop, over, rot, nrot, nip, tuck, roll, pick, dup2,
drop2, swap2, over2, depth, pushr, popr, peekr
```

(Other keywords for structure termination like `end_if_true`, `end_while_true` may also be recognized. Implementations can choose how best to parse these.)

### 2.3 Numeric Literals

**ual 1.2** supports **decimal, binary, and hex** integer literals with **case-insensitive** prefixes for binary and hex:

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

### 2.6 Comments

ual now supports multiple comment styles:

- **Lua-style comments:**  
  Use a double-dash `--` for single-line comments.
  
- **C++-style comments:**  
  Use `//` for single-line comments.

- **C-style block comments:**  
  Enclosed between `/*` and `*/` for multi-line comments.

Examples:

```lua
-- This is a Lua-style single-line comment
// This is a C++-style single-line comment
/* This is a C-style
   block comment */
```

### 2.7 Operators and Delimiters

- **Arithmetic:** `+`, `-`, `*`, `/`  
- **Comparison:** `==`, `!=`, `<`, `>`  
- **Bitwise:** `&`, `|`, `^`, `<<`, `>>`  
- **Assignment:** `=`  
- **Other Delimiters:** Parentheses `( )`, braces `{ }`, brackets `[ ]`, commas `,`, etc.

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

The following Forth-inspired stack manipulation keywords are included:

- **drop:** Remove the top stack element.
- **over:** Copy the second element to the top.
- **rot:** Rotate the top three items.
- **nrot:** Rotate the top three items in the opposite direction.
- **nip:** Remove the second item.
- **tuck:** Duplicate the top element under the second element.
- **roll:** For deeper stack reordering.
- **pick:** For selecting an item from deeper in the stack.
- **dup2:** Duplicate the top two items.
- **drop2:** Remove the top two items.
- **swap2:** Swap the top two pairs of items.
- **over2:** Copy the second pair of items to the top.
- **depth:** Return the current number of items on the stack.

Additionally, these words manage the return stack:

- **pushr:** Push an item from the parameter stack to the return stack.
- **popr:** Pop an item from the return stack onto the parameter stack.
- **peekr:** Retrieve (without removing) the top item on the return stack.

The grammar production for stack statements is extended to include these keywords:

```
<stack-stmt>       ::= "push(" <expr> ")"
                     | "pop()"
                     | "dup()"
                     | "swap()"
                     | "add()" | "sub()" | "mul()" | "div()"
                     | "store()" | "load()"
                     | "drop" | "over" | "rot" | "nrot" | "nip" | "tuck"
                     | "roll" | "pick"
                     | "dup2" | "drop2" | "swap2" | "over2" | "depth"
                     | "pushr(" <expr> ")" | "popr()" | "peekr()"
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

### 4.8 Result Handling Syntactic Sugar (New in ual 1.2)

#### 4.8.1 Syntax

A result value (a table with an `Ok` or `Err` field) can be followed by a chained method call:

```
<result-expression> "." "consider" <result-handler-block>
```

Where:

```
<result-handler-block> ::= "{" { <result-handler> } "}"
```

and

```
<result-handler> ::= "if_ok" <result-handler-body>
                   | "if_err" <result-handler-body>
```

The `<result-handler-body>` can be either a raw expression (which the compiler expands into an anonymous function taking the implicit parameter, denoted as `_1` in sugar syntax) or a full function literal.

#### 4.8.2 Semantics

When chaining `.consider { ... }` on a result:

- If the result contains an `Err` field (non-nil and non-empty), the handler associated with `if_err` is executed.
- Otherwise, if the result contains an `Ok` field, the handler associated with `if_ok` is executed.
- Developers can use the sugar syntax:

  ```
  divide(10,2).consider {
    if_ok  fmt.Printf("ok: %d\n", _1)
    if_err fmt.Printf("fuggedaboutit: %s\n", _1)
  }
  ```

  or, if more control is needed, provide full function literals:

  ```lua
  divide(10,2).consider {
    if_ok = function(val)
      fmt.Printf("ok: %d\n", val)
    end,
    if_err = function(err)
      fmt.Printf("fuggedaboutit: %s\n", err)
    end,
  }
  ```

This feature provides a more expressive, Rust-like result handling mechanism while retaining the ability to specify full function literals when needed.

---

## 5. Scoping and Export Rules

### 5.1 Lexical Scope

- **Local:** Declared via `local x = ...`; visible until block/function ends.  
- **Global (package-level):** Declared at top level.  
  - **Uppercase** first letter → exported.  
  - **Lowercase** first letter → private.

### 5.2 Packages and Visibility

- One file = one package.  
- `import "pkg"` → gain access to **uppercase** symbols from that package.

---

## 6. Operational Semantics

### 6.1 Stack Statements

- `push(expr)`: Evaluate expr, put result on stack.  
- `pop()`: Remove the top item from the stack.  
- `dup()`, `swap()`: Duplicate or swap the top items of the stack.  
- Arithmetic operations like `add()`, `sub()`, `mul()`, `div()` pop the top two items, compute the result, and push it back onto the stack.  
- `store()` and `load()` operate on memory addresses or the framebuffer via the stack.
- Forth-inspired operations include:
  - **drop:** Remove the top stack element.
  - **over:** Copy the second element to the top.
  - **rot:** Rotate the top three items.
  - **nrot:** Rotate the top three items in the opposite direction.
  - **nip:** Remove the second item.
  - **tuck:** Duplicate the top element under the second element.
  - **roll:** For deeper stack reordering.
  - **pick:** For selecting an item from deeper in the stack.
  - **dup2:** Duplicate the top two items.
  - **drop2:** Remove the top two items.
  - **swap2:** Swap the top two pairs of items.
  - **over2:** Copy the second pair of items to the top.
  - **depth:** Return the current number of items on the stack.
- Return stack operations:
  - **pushr(expr):** Push an item from the parameter stack to the return stack.
  - **popr():** Pop an item from the return stack onto the parameter stack.
  - **peekr():** Retrieve (without removing) the top item on the return stack.

### 6.2 Control Flow

- `if_true(expr) ... end_if_true`: Execute the block if `expr` ≠ 0.
- `if_false(expr) ... end_if_false`: Execute the block if `expr` = 0.
- `while_true(expr) ... end_while_true`: Loop as long as `expr` ≠ 0.
- `for i = start, end, step do ... end`: Typical numeric loop.
- `for var in iterator do ... end`: Generic iterator-based loop.

### 6.3 Multiple Return

- `return a, b, c`: Return multiple values from a function.  
- The caller can do `x, y = someFunc(...)`. Extra or missing values are truncated or set to 0/nil.

---

## 7. Packages System Semantics

1. **Single Package Declaration:** Each file starts with `package <name>`.
2. **Imports:** Use `import "otherPkg"` to gain access to uppercase symbols from other packages.
3. **Export vs. Private:** Identifiers starting with an uppercase letter are exported; lowercase ones are private to the package.

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
  -- Parse placeholders (%d, %s, etc.) and return formatted string
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

1. **Parser:**  
   - Implement the EBNF rules, supporting binary/hex numeric literals (case-insensitive) and multiple comment styles (Lua, C++, and C block comments).
2. **Symbol Resolution:**  
   - Maintain a map of packages to uppercase symbols (e.g., `pkg.FuncName`).
3. **Code Generation:**  
   - Convert expressions (including bitwise operators) into opcodes.  
   - Handle numeric literal parsing, local variable allocation (using offsets or a local stack area), and memory structures for arrays/tables.
4. **Operator Precedence:**  
   - Either treat all operators uniformly or define a precedence hierarchy similar to C.
5. **Optimization:**  
   - Inline short functions, eliminate redundant stack operations, etc.

---

## 10. Conclusion

**ual 1.1** provided a comprehensive, minimal-yet-powerful language design featuring:

- **Go-like** package and import systems for modular code.
- **Lua-like** syntax for functions, loops, data structures, and scoping.
- **Binary and hex** numeric literals for hardware-oriented tasks.
- **Bitwise operators** for register manipulation and masking.

**ual 1.2** extends this design by adding:

- Built-in syntactic sugar for result handling (inspired by Rust’s `Result` type) via the `.consider` method.
- Support for multiple comment styles: Lua-style (`--`), C++-style (`//`), and C-style block comments (`/* ... */`).
- A revised set of Forth-inspired stack manipulation keywords (drop, over, rot, nrot, nip, tuck, roll, pick, dup2, drop2, swap2, over2, depth) and new return stack keywords (pushr, popr, peekr) with names that avoid parsing issues.

These enhancements ensure that ual remains both expressive and concise, making it well-suited for small virtual machines and embedded platforms while offering modern conveniences in error handling and code clarity.

---

This completes the integrated **ual 1.2 Specification** document.