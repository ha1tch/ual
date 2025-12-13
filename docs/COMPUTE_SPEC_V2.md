# COMPUTE_SPEC_V2.md — The Algebraic Optimization Kernel

## 1. Executive Summary

The `.compute()` construct introduces **Context-Bound Optimization Islands** to `ual`. It creates a semantic boundary where the language switches from **Logistics** (stack-based byte movement) to **Logic** (register-based native arithmetic).

**V2 Capabilities:**
1.  **Native Math:** Zero-copy arithmetic on `int64` or `float64` types.
2.  **Type Rigidity:** The attached stack's type dictates the kernel's native type.
3.  **Local Buffers:** Stack-allocated temporary arrays (`var buf[1024]`).
4.  **Container Views:** Zero-copy read/write access to container data as arrays (`self.pixels[i]`).

## 2. Syntax Definition

The construct attaches to a `StackBlock` (the setup phase).

### 2.1 Grammar
```ebnf
ComputeStmt   ::= StackBlock ".compute" "(" ComputeKernel ")"
ComputeKernel ::= "{" Bindings? ComputeBody "}"
Bindings      ::= "|" Ident ("," Ident)* "|" | "||"

ComputeBody   ::= (Statement)*

Statement     ::= VarDecl 
                | ArrayDecl 
                | Assignment 
                | ReturnStmt 
                | IfStmt 
                | WhileStmt 
                | BreakStmt 
                | ContinueStmt
                | ExprStmt

VarDecl       ::= "var" Ident "=" Expr
ArrayDecl     ::= "var" Ident "[" IntLit "]"  -- Local fixed-size array
Assignment    ::= Target "=" Expr
ReturnStmt    ::= "return" (Expr ("," Expr)*)?

Target        ::= Ident 
                | Ident "[" Expr "]"          -- Local array write
                | MemberExpr "[" Expr "]"     -- Container array write (self.prop[i])

Expr          ::= InfixExpr
MemberExpr    ::= "self" "." Ident
```

### 2.2 Example
```ual
@graph {
    -- [Logistics Phase]
    push:start_node
}.compute(
    -- [Logic Phase]
    {|start|
        -- Local Array (Tier 2)
        var queue[1024]
        var head = 0
        var tail = 0
        
        queue[tail] = start
        tail = tail + 1
        
        -- Container Array Write (Tier 3)
        -- Direct zero-copy write to @graph storage
        self.visited[start] = 1
        
        while head < tail {
            var u = queue[head]
            head = head + 1
            
            -- ... logic ...
        }
        return
    }
)
```

## 3. Memory Model & Types

The compute kernel operates in a **Locked Scope**. The stack mutex is held for the duration of the block.

### 3.1 Type Homogeneity
The kernel is **Type Rigid**. There is no runtime type dispatch.
*   If `@stack` is `TypeInt64`: All variables, literals, and array views are `int64`.
*   If `@stack` is `TypeFloat64`: All variables, literals, and array views are `float64`.

### 3.2 Data Tiers

| Tier | Name | Source | Access | Mutability | Implementation |
|---|---|---|---|---|---|
| **1** | **Scalar** | Stack Binding / `self.prop` | Copy (Register) | Local only | Go `int64`/`float64` variable |
| **2** | **Local Array** | `var x[N]` | Direct Indexing | Read/Write | Go Array `[N]T` |
| **3** | **Container Array** | `self.prop[i]` | Mapped View | Read/Write | `unsafe.Slice` over raw bytes |

## 4. Semantics & Rules

### 4.1 Binding (`|a,b|`)
*   **Order:** LIFO (Top-down). `{|a, b|}` means `a = pop()`, then `b = pop()`.
*   **Consumption:** Arguments are removed from the stack upon entry.
*   **Constraint:** Bindings are **Forbidden** on Hash-perspective stacks (ambiguous pop). Use `{||}` and access data via `self`.

### 4.2 Returns
*   **Sequential Push:** `return a, b` pushes `a`, then `b` (`b` is new top).
*   **Hash Stacks:** Returns are written to reserved keys `__result_0__`, `__result_1__`, etc.
*   **Void:** Omitting `return` is valid (Consumer pattern).

### 4.3 `self` Access Rules
*   **Scalar Read (`self.x`):** Reads raw bytes, converts to scalar. **Read-Only** relative to the container (modifying the local var does not update the stack).
*   **Array Access (`self.x[i]`):** Casts the underlying byte slice to a typed slice (`[]T`). **Read-Write**. Mutating this view updates the stack storage immediately.
*   **Constraint:** A property cannot be accessed as both Scalar and Array in the same block.

### 4.4 Scope Parsimony
*   **No External Access:** Cannot touch other stacks.
*   **No Globals:** Cannot touch global variables.
*   **Panic:** Errors (div by zero, bounds check) cause a hard panic (catchable via `try`).

## 5. Compiler Architecture

### 5.1 Lexer (`lexer.go`)
*   Ensure `[` and `]` are tokenized.
*   Ensure `compute`, `self` are keywords.

### 5.2 Parser (`parser.go`)
*   **`parseComputeStmt`**: Isolate the infix parser.
*   **`parseInfixPrimary`**: Handle `Ident` (var), `MemberExpr` (`self.x`), `IndexExpr` (`x[i]`).
*   **`parseComputeVarDecl`**: Handle array syntax `var x[100]`.
*   **`parseComputeAssign`**: Handle indexed assignment `x[i] = y`.

### 5.3 Codegen (`codegen.go`)

The generation must occur in three passes within the function body.

#### Pass 1: Analysis & Header Generation
Scan the AST of the compute body to identify all `self.prop[i]` usages.
For every unique property accessed as an array, emit the mapping boilerplate using `unsafe`.

```go
// Generated Header
// 1. Get raw bytes (Thread-safe because lock is held)
_raw_pixels, _ok := stack.GetRaw("pixels")
if !_ok { panic("compute: pixels missing") }

// 2. Map bytes to typed slice (Zero-Copy)
// Target: *int64 or *float64 based on Stack Type
_ptr_pixels := (*int64)(unsafe.Pointer(&_raw_pixels[0]))
_view_pixels := unsafe.Slice(_ptr_pixels, len(_raw_pixels)/8)
```

#### Pass 2: Variable Declaration
*   **Bindings:** `_val, _ := stack.PopRaw(); var a = bytesToT(_val)`
*   **Local Arrays:** `var queue [1024]int64`

#### Pass 3: Body Translation
Translate expressions to Go, substituting `self` accesses.

*   `self.mass` (Scalar) → `bytesToT(stack.GetRaw("mass"))`
*   `self.pixels[i]` (Array) → `_view_pixels[i]`

### 5.4 Math Library
Inside `generateComputeExpr`, detect function calls.
*   Map `sqrt(x)` → `math.Sqrt(x)`
*   Map `abs(x)` → `math.Abs(x)`
*   etc.

## 6. Runtime Support (`stack.go`)

The runtime must export specific "Raw" methods for the compiler to target. These methods **assume the lock is already held**.

```go
// Puts a value without locking. Used for LIFO/FIFO returns.
func (s *Stack) PushRaw(data []byte)

// Gets a value without locking. Used for Bindings.
func (s *Stack) PopRaw() ([]byte, error)

// Gets a reference to the underlying byte slice.
// Used for 'self.prop' and 'self.prop[i]'.
func (s *Stack) GetRaw(key string) ([]byte, bool)

// Sets a value. Used for Hash returns.
func (s *Stack) SetRaw(key string, value []byte)

// Helper for 'self[i]' on Indexed stacks
func (s *Stack) GetAtRaw(index int) ([]byte, bool)
```

## 7. V2 Implementation Checklist

1.  **Update Parser:** Add array declaration `var x[N]` and index access `x[i]` logic.
2.  **Update Codegen (Preamble):** Implement the `unsafe.Slice` view generation for container arrays.
3.  **Update Codegen (Body):** Update variable resolution to use the `_view_` variables for array access.
4.  **Update Codegen (Math):** Add the math function allowlist/mapping. ✓ (Done in V1)
5.  **Test:** Create examples using local arrays and `self` buffers to verify zero-copy performance.

---

## V1 Status (Current Implementation)

| Feature | Status | Notes |
|---------|--------|-------|
| Type rigidity | ✓ | Stack type dictates kernel type |
| Bindings `{|a,b|}` | ✓ | LIFO order |
| Empty bindings `{||}` | ✓ | For Hash stacks |
| `self.property` | ✓ | Hash perspective read |
| `self[i]` | ✓ | Indexed perspective read |
| `set("key", value)` | ✓ | Hash perspective write |
| `get("key")` | ✓ | Hash perspective read outside compute |
| Multiple returns | ✓ | `return a, b` |
| Void return | ✓ | Consumer pattern |
| Control flow | ✓ | `if`, `while`, `break`, `continue` |
| Negative literals | ✓ | `-5.0`, `-x` |
| Math functions | ✓ | `sqrt`, `abs`, `sin`, `cos`, `pow`, etc. |
| Local arrays `var x[N]` | ✓ | V2 |
| Local array read `x[i]` | ✓ | V2 |
| Local array write `x[i] = v` | ✓ | V2 |
| Container array read `self.prop[i]` | ✓ | V2 (requires pre-allocated storage) |
| Container array write `self.prop[i] = v` | ✓ | V2 (requires pre-allocated storage) |
| `unsafe.Slice` view generation | ✓ | V2 (auto-generated preamble) |

### V2 Notes

**Container Array Views** (`self.prop[i]`): The compiler generates `unsafe.Slice` mappings for zero-copy access to Hash stack properties. However, the property must contain pre-allocated byte storage of appropriate size before use. This is typically done by:
1. External initialization (loading from file, network, etc.)
2. A dedicated allocation primitive (future work)
3. Pushing the correct number of bytes during setup

---
**End of Specification.**
