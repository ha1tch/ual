# ual 1.9 PROPOSAL: Consolidated Grammar Specification

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction: Toward Syntactic Coherence

This document consolidates ual's grammar from its foundational 1.3 specification through subsequent evolutions up to version 1.8. As ual has matured, its syntax has evolved systematically while maintaining its core philosophy of explicit stack operations, container-centric design, and progressive discovery. This consolidation serves as both a reference document and a reflection on how ual's grammar has evolved to better support its design principles.

Throughout ual's evolution, we see a pattern of extending rather than replacing - new syntactic forms build upon established metaphors, creating a coherent language where relationships between data and their contexts remain explicit and visible. This approach stands in contrast to traditional programming paradigms that focus primarily on manipulating individual values with less emphasis on the contexts containing them.

By documenting ual's complete grammar in one place, we provide a foundation for coherent future evolution that respects the language's philosophical principles and maintains its distinctive character as a bridge between stack-based and traditional programming models.

## 2. Overall Structure and Lexical Elements

### 2.1 File Structure

ual programs are organized into packages, with each file belonging to a single package. A typical source file follows this structure:

```
<program>          ::= <package-decl> { <import-decl> } { <top-level-decl> }

<package-decl>     ::= "package" <identifier>

<import-decl>      ::= "import" <string-literal>

<top-level-decl>   ::= <function-def> | <global-var-decl> | <enum-decl>
```

### 2.2 Lexical Structure

#### 2.2.1 Identifiers

```
<identifier>       ::= <letter> { <letter> | <digit> | "_" }

<letter>           ::= "A" | ... | "Z" | "a" | ... | "z"

<digit>            ::= "0" | ... | "9"
```

Identifiers beginning with an uppercase letter are exported from their package, while lowercase identifiers are package-private.

#### 2.2.2 Keywords

```
<keyword>          ::= "function" | "end" | "if_true" | "if_false" | "while_true" 
                     | "return" | "local" | "do" | "for" | "in" | "push" | "pop" 
                     | "dup" | "swap" | "add" | "sub" | "mul" | "div" | "store" 
                     | "load" | "import" | "package" | "drop" | "over" | "rot" 
                     | "nrot" | "nip" | "tuck" | "roll" | "pick" | "dup2" | "drop2" 
                     | "swap2" | "over2" | "depth" | "len" | "pushr" | "popr" 
                     | "peekr" | "switch_case" | "case" | "default" | "end_switch" 
                     | "Stack" | "enum" | "scope" | "defer_op"
```

End markers for control structures (e.g., `end_if_true`, `end_while_true`) are also recognized.

#### 2.2.3 Literals

```
<literal>          ::= <number-literal> | <string-literal> | <bool-literal> | <nil-literal>

<number-literal>   ::= <decimal-literal> | <binary-literal> | <hex-literal>

<decimal-literal>  ::= <digit> { <digit> }

<binary-literal>   ::= ("0b" | "0B") { "0" | "1" }

<hex-literal>      ::= ("0x" | "0X") { <digit> | "A"..."F" | "a"..."f" }

<string-literal>   ::= '"' { <any-char-except-quote> | <escape-char> } '"'
                     | "'" { <any-char-except-quote> | <escape-char> } "'"

<bool-literal>     ::= "true" | "false"

<nil-literal>      ::= "nil"
```

Hash literals (introduced in 1.7) use a tilde (`~`) to separate keys and values:

```
<hash-literal>     ::= "{" [ <key-value-pair> { "," <key-value-pair> } ] "}"

<key-value-pair>   ::= <expression> "~" <expression>
```

#### 2.2.4 Comments

```
<comment>          ::= <single-line-comment> | <multi-line-comment>

<single-line-comment> ::= "--" <any-chars-until-eol>
                       | "//" <any-chars-until-eol>

<multi-line-comment>  ::= "/*" <any-chars-until-end-marker> "*/"
```

## 3. Declarations and Definitions

### 3.1 Function Definitions

```
<function-def>     ::= [ "@error" ">" ] "function" <identifier> "(" [ <param-list> ] ")" 
                     ( <block> "end" | "{" <block> "}" )

<param-list>       ::= <identifier> [ ":" <type> ] { "," <identifier> [ ":" <type> ] }

<type>             ::= <identifier> [ "(" <type> ")" ]
```

Functions marked with `@error >` can push values to the error stack and have special error handling semantics.

### 3.2 Variable Declarations

```
<global-var-decl>  ::= <identifier> "=" <expression>

<local-decl>       ::= "local" <identifier> [ "=" <expression> ]
```

### 3.3 Enum Declarations

```
<enum-decl>        ::= "enum" <identifier> "{" <enum-variants> "}"

<enum-variants>    ::= <identifier> [ "=" <expression> ] { "," <identifier> [ "=" <expression> ] }
```

## 4. Stack Operations and Perspectives

### 4.1 Stack Creation and Selection

```
<stack-creation>   ::= "@" "Stack" "." "new" "(" <type> [ "," <key-type> ] [ "," <ownership-mode> ] 
                     [ "," "PrimaryPerspective" ":" <perspective-type> ] ")" 
                     [ ":" "alias" ":" <string-literal> ]

<key-type>         ::= "KeyType" ":" <type>

<ownership-mode>   ::= "Owned" | "Borrowed" | "Mutable"

<perspective-type> ::= "LIFO" | "FIFO" | "MAXFO" | "MINFO" | "Hashed"
```

### 4.2 Stack Selection

```
<stack-selector>   ::= "@" <identifier> ":" 
                     | "@" <identifier> ">" 
                     | ":"
                     | ">"

<stack-context-block> ::= <stack-selector> "{" <block> "}"
```

The angle bracket syntax (`>`) is deprecated but supported for backward compatibility. The colon syntax (`:`) is preferred.

### 4.3 Stacked Mode Statements

```
<stacked-stmt>     ::= <stack-selector> <stacked-op-list>

<stacked-op-list>  ::= <stacked-op> { <stacked-op> }

<stacked-op>       ::= <identifier> [ <stacked-param> ]
                     | <identifier> "(" <expr-list> ")"

<stacked-param>    ::= ":" <literal>
                     | "(" <expression> ")"
```

Multiple stack operations can be chained with semicolons:

```
<multi-stack-stmt> ::= <stacked-stmt> { ";" <stacked-stmt> }
```

### 4.4 Stack Perspectives

```
<perspective-operation> ::= <stack-selector> <perspective>

<perspective>      ::= "lifo" | "fifo" | "maxfo" | "minfo" | "hashed" | "flip"
```

### 4.5 Ownership and Borrowing Operations

```
<borrow-operation> ::= <stack-selector> "borrow" "(" <range-expr> "@" <identifier> ")"
                     | <stack-selector> "borrow_mut" "(" <range-expr> "@" <identifier> ")"

<take-operation>   ::= <stack-selector> "take" "(" <expression> ")"

<borrow-shorthand> ::= <stack-selector> "<<" <range-expr> <identifier>
                     | <stack-selector> "<:mut" <range-expr> <identifier>
                     | <stack-selector> "<:own" <identifier>

<range-expr>       ::= "[" <expression> ".." <expression> "]"
                     | "[" <expression> { "," <expression> } "]"
```

### 4.6 Crosstacks

```
<crosstack-selector> ::= <expression> "~" <expression>
                       | "[" <expression> ".." <expression> "]" "~" <expression>
                       | "[" <expression> { "," <expression> } "]" "~" <expression>
                       | <expression> "~"
```

## 5. Statements and Control Flow

### 5.1 Basic Statements

```
<statement>        ::= <assignment-stmt>
                     | <stack-stmt>
                     | <if-true-stmt>
                     | <if-false-stmt>
                     | <while-true-stmt>
                     | <for-num-stmt>
                     | <for-gen-stmt>
                     | <return-stmt>
                     | <function-call-stmt>
                     | <local-decl>
                     | <do-block>
                     | <switch-stmt>
                     | <stacked-stmt>
                     | <defer-stmt>
                     | <scope-block>
                     | <consider-block>
                     | <empty>

<assignment-stmt>  ::= <var-list> "=" <expr-list>

<var-list>         ::= <variable> { "," <variable> }

<variable>         ::= <identifier>
                     | <index-access>

<index-access>     ::= <expression> "[" <expression> "]"
                     | <expression> "." <identifier>

<stack-stmt>       ::= <stack-operation> "(" [ <expr-list> ] ")"
                     | <stack-operation-noargs>

<function-call-stmt> ::= <identifier> "(" [ <expr-list> ] ")"

<return-stmt>      ::= "return" [ <expr-list> ]
```

### 5.2 Control Structures

```
<if-true-stmt>     ::= "if_true" "(" <expression> ")" 
                     ( <block> ["end_if_true"] | "{" <block> "}" )

<if-false-stmt>    ::= "if_false" "(" <expression> ")" 
                     ( <block> ["end_if_false"] | "{" <block> "}" )

<while-true-stmt>  ::= "while_true" "(" <expression> ")" 
                     ( <block> ["end_while_true"] | "{" <block> "}" )

<do-block>         ::= "do" <block> "end"

<scope-block>      ::= "scope" "{" <block> "}"
```

### 5.3 For Loops

```
<for-num-stmt>     ::= "for" <identifier> "=" <expression> "," <expression> [ "," <expression> ] 
                     "do" <block> "end"

<for-gen-stmt>     ::= "for" <identifier> "in" <expression> "do" <block> "end"
```

### 5.4 Switch Statement

```
<switch-stmt>      ::= "switch_case" "(" <expression> ")" <case-list> 
                     [ "default" ":" <block> ] "end_switch"

<case-list>        ::= { <case-stmt> }

<case-stmt>        ::= "case" <case-expr> ":" <block>

<case-expr>        ::= <expression>
                     | "[" <expr-list> "]"
```

The enhanced bitmap-based switch supports multi-value matching using the array syntax within `case`.

### 5.5 Defer Operations

```
<defer-stmt>       ::= "defer_op" "{" <block> "}"
                     | "@defer" ":" "push" "{" <block> "}"
```

### 5.6 Consider Blocks (Pattern Matching)

```
<consider-block>   ::= <expression> "." "consider" "{" <pattern-list> "}"

<pattern-list>     ::= { <pattern-clause> }

<pattern-clause>   ::= <pattern-type> [ "(" <expr-list> ")" ] "{" <block> "}"

<pattern-type>     ::= "if_ok" | "if_err" | "if_equal" | "if_match" | "if_type" | "if_else"
```

## 6. Expressions

```
<expression>       ::= <literal>
                     | <variable>
                     | <function-call-expr>
                     | "(" <expression> ")"
                     | <binary-op-expr>
                     | <table-constructor>
                     | <array-constructor>
                     | <stack-creation-expr>
                     | <hash-literal>

<expr-list>        ::= <expression> { "," <expression> }

<binary-op-expr>   ::= <expression> <binary-op> <expression>

<binary-op>        ::= "+" | "-" | "*" | "/" | "==" | "!=" | "<" | ">" | "<=" | ">="
                     | "&" | "|" | "^" | "<<" | ">>" | "and" | "or"

<function-call-expr> ::= <identifier> "(" [ <expr-list> ] ")"

<stack-creation-expr> ::= "Stack" "." "new" "(" [ <arguments> ] ")"
```

### 6.1 Table and Array Constructors

```
<table-constructor> ::= "{" [ <table-field-list> ] "}"

<table-field-list>  ::= <table-field> { "," <table-field> }

<table-field>       ::= <key-def> <expression>

<key-def>           ::= <identifier> "="
                     | "[" <expression> "]" "="
                     | (empty)

<array-constructor> ::= "[" [ <expr-list> ] "]"
```

## 7. Illustrative Examples of Grammar Evolution

### 7.1 Simple Function: 1.3 to 1.8 Evolution

#### 7.1.1 Original Style (ual 1.3)

```lua
function factorial(n)
  if_true(n <= 1) 
    return 1
  end_if_true
  
  return n * factorial(n - 1)
end
```

#### 7.1.2 Compact Blocks (ual 1.4)

```lua
function factorial(n)
  if_true(n <= 1) { return 1 }
  
  return n * factorial(n - 1)
end
```

#### 7.1.3 Stack-Oriented Style (ual 1.4+)

```lua
function factorial(n)
  @Stack.new(Integer): alias:"s"
  
  @s: push(n)
  @s: push:1 leq if_true {
    return 1
  }
  
  @s: push(n) push:1 sub factorial mul
  return s.pop()
end
```

#### 7.1.4 Full Stack-Based with Consider (ual 1.8)

```lua
function factorial(n)
  n.consider {
    if_equal(0) {
      return 1
    }
    if_equal(1) {
      return 1
    }
    if_match(function(x) return x > 1 end) {
      return n * factorial(n - 1)
    }
    if_else {
      error("Factorial undefined for negative numbers")
    }
  }
end
```

### 7.2 Stack Perspective Evolution

#### 7.2.1 Original Perspective (ual 1.5)

```lua
@stack: lifo  // Default: Last-In-First-Out
@stack: push:1 push:2 push:3
// stack is now [3, 2, 1]
value = stack.pop()  // value = 3
```

#### 7.2.2 FIFO Perspective (ual 1.5)

```lua
@stack: fifo  // First-In-First-Out
@stack: push:1 push:2 push:3
// stack is now [1, 2, 3] from a FIFO perspective
value = stack.pop()  // value = 1
```

#### 7.2.3 Hashed Perspective (ual 1.7)

```lua
@map: hashed   // Key-value perspective
@map: push("answer", 42)
@map: push("greeting", "hello")
value = map.pop("answer")  // value = 42
```

#### 7.2.4 Crosstack Perspective (ual 1.8)

```lua
@matrix: Stack.new(Stack)
@matrix: push(row1) push(row2) push(row3)

// Horizontal access across all rows at level 0
@matrix~0: sum  // Sum first element of each row
```

### 7.3 Borrowed Stack Segments (ual 1.6)

```lua
@data: push:1 push:2 push:3 push:4 push:5

scope {
  // Borrow elements 1-3 (second through fourth elements)
  @window: borrow([1..3]@data)
  
  // Process borrowed segment without copying
  sum = process_segment(window)
  
  // No modification to original data during borrow scope
}

// Shorthand notation
@view: <<[0..2]data  // Borrow first three elements
```

### 7.4 Error Handling Evolution

#### 7.4.1 Original Error Stack Approach (ual 1.5)

```lua
@error > function read_file(filename)
  file = io.open(filename, "r")
  if file == nil then
    @error > push("File not found: " .. filename)
    return nil
  end
  
  return file
end
```

#### 7.4.2 Using Consider for Error Handling (ual 1.8)

```lua
function read_file(filename)
  result = io.open(filename, "r")
  
  result.consider {
    if_ok {
      return _1  // Return file handle on success
    }
    if_err {
      fmt.Printf("Error opening file: %s\n", _1)
      return nil
    }
  }
end
```

## 8. The Philosophy Behind the Grammar

ual's grammar has evolved consistently with its philosophical foundation of making relationships between data explicit. Several principles are evident in this evolution:

### 8.1 Explicit Context

The grammar consistently emphasizes the context in which operations occur. Stack selectors (`@stack:`) make clear which stack is being operated on, and the evolution from angle bracket to colon syntax refined this clarity without abandoning the explicit context.

The introduction of stack context blocks in 1.4 (`@stack: { operations }`) provided a way to maintain explicit context while reducing repetition, demonstrating ual's balance between explicitness and verbosity.

### 8.2 Progressive Complexity

Throughout its evolution, ual has maintained a "progressive complexity" approach. Basic patterns remain simple, while more sophisticated capabilities build on familiar foundations. For example:

- Simple stack operations: `@stack: push:42`
- Stack context blocks: `@stack: { push:1 push:2 add }`
- Multi-stack operations: `@s: push:"42"; @i: <s`
- Crosstacks: `@matrix~0: sum`

Each complexity layer builds naturally on previous patterns, allowing developers to progressively discover advanced features as needed.

### 8.3 Relationships Over Individuals

Traditional programming often focuses on individual values, with containers as implementation details. ual's grammar fundamentally inverts this, emphasizing the containers and relationships between values.

The evolution of stack perspectives (LIFO, FIFO, MAXFO, MINFO, Hashed) demonstrates this philosophy - rather than creating separate data structures, ual extends the relationship model through different perspectives on the same container type.

### 8.4 Visual Distinctiveness for Conceptual Differences

ual's syntax uses visually distinctive markers to signal conceptual differences:

- `@` for stack selectors
- `:` for parameter separators and stack selection
- `~` for key-value separation and crosstacks
- `.consider{}` for pattern matching
- `[x..y]` for ranges

These visual cues help reinforce the conceptual model while keeping the syntax clear and readable.

## 9. Conclusion

The consolidated grammar presented in this document reflects ual's evolution from its 1.3 foundations through version 1.8, capturing the syntactic refinements that have progressively enhanced the language's expressiveness while maintaining its core philosophy.

What emerges is not just a language syntax but a coherent system for thinking about programming that centers on explicit relationships between data. This approach offers a compelling alternative to traditional programming models, suggesting that building software from "molecules, DNA, and proteins" rather than individual atoms can lead to more expressive, maintainable code.

As ual continues to evolve, this consolidated grammar provides a foundation for ensuring that new features remain consistent with the language's philosophical principles and syntactic patterns, preserving its unique identity as a bridge between stack-based and traditional programming models.