# ual 1.3 Specification (Part 1)

## 1. Introduction

**ual** is a high-level, Lua-like language designed for use on minimal or retro platforms with small virtual machines or embedded hardware. The language emphasizes:

1. **Stack-based operations** for arithmetic and memory access
2. **Lua-style** control flow, scoping, and data structures (arrays, tables, strings)
3. **Go-like package** conventions (uppercase = exported, lowercase = private)
4. **Multiple returns** and flexible for loops (numeric and iterator-based)
5. **Binary / Hexadecimal** numeric literals for easy hardware-oriented programming
6. **Bitwise operators** for direct manipulation of registers and masks

This document outlines the **lexical structure**, **grammar**, **scoping rules**, **packages system**, **operational semantics**, **result handling syntactic sugar**, and new features in 1.3 including **stack as first-class objects**, **stacked mode syntax**, **switch statement** and other enhancements to the ual programming language.

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
drop2, swap2, over2, depth, len, pushr, popr, peekr, switch_case, 
case, default, end_switch, Stack
```

(Other keywords for structure termination like `end_if_true`, `end_while_true` may also be recognized. Implementations can choose how best to parse these.)

### 2.3 Numeric Literals

**ual 1.3** supports **decimal, binary, and hex** integer literals with **case-insensitive** prefixes for binary and hex:

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

ual supports multiple comment styles:

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

### 2.8 Special Syntax Markers (New in 1.3)

- **Stacked Mode Prefix:** `>` for implicit data stack operations
- **Stack Selection:** `@stackname >` to specify which stack to use in stacked mode
- **Literal Value Specifier:** `:` to indicate literal values in stacked mode (e.g., `push:10`)

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
- Packages exist in a flat namespace. If multiple packages share the same name, that's an error.

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

Below is an **EBNF** grammar that integrates packages, imports, numeric literals (decimal/binary/hex), bitwise operators, standard control structures, and ual 1.3 additions.

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

- If `<identifier>` begins with uppercase, it's **exported**.
- If it's lowercase, it's private to the package.

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
                     | <switch-stmt>
                     | <stacked-mode-stmt>
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
- **depth/len:** Return the current number of items on the stack (both keywords perform the same function).

In addition, convenience operations for the return stack are maintained from previous versions:

- **pushr(expr):** Push a value from data stack to return stack (syntax sugar for `rstack.push(expr)`)
- **popr():** Pop a value from return stack to data stack (syntax sugar for `dstack.push(rstack.pop())`)
- **peekr():** Retrieve the top value from return stack without removing it (syntax sugar for `rstack.peek()`)

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
                     | "dup2" | "drop2" | "swap2" | "over2" | "depth" | "len"
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

#### 4.4.10 Switch Statement (New in 1.3)

```
<switch-stmt>      ::= "switch_case(" <expr> ")" <case-list> ["default:" <block>] "end_switch"

<case-list>        ::= { <case-stmt> }

<case-stmt>        ::= "case" <expr-list> ":" <block>
```

Example:

```lua
switch_case(value)
  case 1:
    -- code for case 1
  case 2, 3:
    -- code for cases 2 and 3
  default:
    -- default code
end_switch
```

The switch statement is implemented as follows:

1. The switch expression is evaluated once.
2. The resulting value is compared to each case value in order.
3. When a match is found, execution begins at the first statement of the corresponding case block.
4. Execution continues through subsequent case blocks (fall-through behavior) unless explicitly terminated.
5. The default case executes when no other case matches.

At the compiler level, the switch statement maps directly to TinyGo's switch mechanism:
- For dense integer ranges, TinyGo can generate efficient jump tables
- For sparse or non-integer cases, it generates a series of comparisons
- String comparison and other complex cases are properly handled through TinyGo's runtime

#### 4.4.11 Stacked Mode Statement (New in 1.3)

```
<stacked-mode-stmt> ::= [<stack-selector>] ">" <stacked-op-list>

<stack-selector>    ::= "@" <identifier>

<stacked-op-list>   ::= <stacked-op> { <stacked-op> }

<stacked-op>        ::= <stack-func-name> [<stacked-param>]
                      | <identifier> "(" <arg-list> ")"

<stacked-param>     ::= ":" <literal>
                      | "(" <expr> ")"
```

Examples:

```lua
> push:10 dup add              -- Implicit data stack
@rstack > push:42 swap         -- Explicit stack selection
@myStack > push(x+y) dup mul   -- Expression with parentheses
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
         | <stack-creation-expr>
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

#### 4.5.2 Stack Creation (New in 1.3)

```
<stack-creation-expr> ::= "Stack.new()"
```

Creates a new stack object that can be used for stack operations.

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

### 4.8 Result Handling Syntactic Sugar

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
    
- **Edge Cases**:
    
    - If the result contains both an `Err` field (non-nil and non-empty) and an `Ok` field, the `Err` field takes precedence and only the `if_err` handler is executed.
    - If the result contains neither an `Err` nor an `Ok` field, no handler is executed and the operation completes silently.
    - If the `Err` field exists but is nil or an empty value, it is treated as not present, and the system falls back to checking for the `Ok` field.
    - If the appropriate handler (`if_ok` or `if_err`) is not provided for the evaluated condition, no action is taken for that condition.
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
    

This feature provides a more expressive, Rust-like result handling mechanism while retaining the ability to specify full function literals when needed. Unlike Rust's `Result` type, which is a strict enum allowing only one variant at a time, ual's implementation uses table fields, requiring these explicit precedence rules to ensure consistent behavior across all implementations.