# ual 1.6 PROPOSAL: Borrowed Stack Segments

This document proposes an extension to ual’s container-centric paradigm by introducing *borrowed stack segments*. Borrowed segments allow parts of a stack to be temporarily viewed and manipulated without transferring ownership, much like Rust’s slice borrowing—but with a focus on explicitness and readability. This proposal builds on ual’s pragmatic genericity and ownership model to offer zero-copy, transient access to stack data while preserving strong safety assurances that are highly visible in the code.

---

## 1. Introduction

ual’s design philosophy centers on explicit data transfers and container-centric operations. Every push, pop, or transformation is visible, making safety assurances clear. Borrowed stack segments extend this philosophy by allowing a programmer to create a “window” into an existing stack. This window borrows a contiguous range of elements, enabling read (or limited write) operations without copying the underlying data. As with Rust’s borrowing, this approach ensures that no changes occur to the original stack in ways that could invalidate the borrowed segment.

---

## 2. Motivation and Background

### 2.1 The Case for Borrowed Segments

In many systems—especially those involving large datasets or performance-critical operations—it is desirable to work with a subsection of a data structure without incurring the cost of copying. Rust achieves this through slices with zero-cost borrowing, backed by its advanced lifetime and borrow-checking mechanisms.

ual intentionally avoids implicit references to local variables. Instead, it relies on explicit ownership transfers between stacks. While this model enhances transparency and safety, it can lead to unnecessary data copies when only temporary access is needed. Borrowed stack segments solve this by:
- Allowing non-owning, read-only (or carefully controlled mutable) views into a stack.
- Eliminating extra copying overhead.
- Making the safety guarantees explicit in the code—nothing leaves or enters a stack without ual’s notice.

### 2.2 Comparing with Rust

Rust’s borrowing is automatic and enforced by the compiler’s lifetime inference. For example, a Rust slice:

```rust
fn get_slice(vec: &Vec<i32>) -> &[i32] {
    &vec[1..3]
}
```

ensures that the returned slice remains valid for as long as `vec` is alive, and any attempt to mutate `vec` while the slice is in use is rejected at compile time.

In ual, the absence of references traditionally meant every operation was an ownership move. Borrowed segments introduce a similar mechanism explicitly:
- Borrowing a segment does not move the data.
- The borrowed segment is only valid within an explicit scope.
- The operation is visible in the code, reinforcing secure and predictable data access.

---

## 3. Design Goals

The primary goals of the Borrowed Stack Segments proposal are:

1. **Zero-Copy Data Access:** Enable operations on a portion of a stack without duplicating data.
2. **Explicit Ownership & Lifetime:** Make the borrowing operation explicit, with a clearly defined scope and lifetime.
3. **Enhanced Readability:** Improve code clarity by showing precisely what data is being borrowed and for how long.
4. **Safety Through Visibility:** The explicit borrowing reinforces safety, making it easy to verify that no invalid mutations occur.
5. **Integration with Go/TinyGo:** Leverage Go/TinyGo’s underlying type safety and atomic operations while extending ual’s container-centric model.

---

## 4. Proposed Syntax and Semantics

### 4.1 Borrowing a Stack Segment

We propose a new syntax for borrowing a contiguous range from a stack. For example:

```lua
@mystack: push:1 push:2 push:3 push:4
@window: borrow([1..2] @mystack)
@someotherstack: push(window.pick())
```

- **borrow([1..2] @mystack):** This operation creates a borrowed segment (window) from elements at positions 1 to 2 of `mystack` without transferring ownership.
- **window.pick():** Accesses an element from the borrowed segment.

### 4.2 Lifetime and Scope

Borrowed segments must have an explicitly defined lifetime. For example, using a scope block:

```lua
@mystack: push:10 push:20 push:30 push:40

-- Begin borrow scope
scope {
  @window: borrow([2..3] @mystack)
  -- Work with the borrowed segment
  local x = window.pick()  -- Reads element from window
  @result: push(x)
}
-- End borrow scope; the window is now invalid, ensuring mystack can be safely modified.
```

The compiler must ensure that:
- The borrowed segment does not outlive its parent stack.
- No modifications that would affect the borrowed range occur while the borrow is active.
- The borrow is only allowed for non-owning operations (read-only, or explicitly controlled mutable access).

---

## 5. Code Examples

### 5.1 Basic Borrowing Example

```lua
@mystack: push:1 push:2 push:3 push:4
@window: borrow([1..2] @mystack)
-- Borrowed segment now represents elements [1, 2]
local first_elem = window.pick()  -- Assume pick() retrieves the first element of the window
@result: push(first_elem)         -- Transfers the value from the borrowed segment
```

### 5.2 Borrowing with Scope Enforcement

```lua
@mystack: push:100 push:200 push:300

scope {
  @window: borrow([2..3] @mystack)
  local second_elem = window.pick()
  @result: push(second_elem)
}
-- Outside the scope, modifications to mystack are allowed without affecting the now-invalid window.
```

### 5.3 Mutable Borrowing (Proposed)

For scenarios requiring temporary mutable access, we might extend the syntax:
  
```lua
scope {
  @window: borrow_mut([1..2] @mystack)
  window.modify(function(x) return x * 2 end)
}
```

Here, `borrow_mut` allows controlled, exclusive modification of the segment, and the compiler ensures no concurrent mutations occur.

---

## 6. Comparison with Rust

### 6.1 Rust’s Zero-Cost Borrowing

Rust automatically ties the lifetime of a slice to its parent vector:

```rust
fn get_slice(vec: &Vec<i32>) -> &[i32] {
    &vec[1..3]
}
```

- **Safety:** Rust prevents mutation of the vector while the slice is alive.
- **Zero-Copy:** The slice is a view into the original vector, with no duplication.
- **Implicit Lifetimes:** Lifetimes are inferred automatically.

### 6.2 ual’s Explicit Borrowed Segments

In contrast, ual’s approach:
- **Explicit Scope:** Borrowing is visible in the code with explicit `borrow` and scope blocks.
- **Clarity:** The borrowed range and its lifetime are declared, making the safety guarantees visible.
- **Simplicity:** While it may require more explicit notation, the operations are straightforward and aligned with ual’s container-centric model.
- **Flexibility:** ual can support both read-only and mutable borrows through distinct operations (`borrow` vs. `borrow_mut`).

The explicit nature of ual’s borrow system means that safety assurances are not hidden—developers see exactly what’s borrowed and for how long. This transparency can lead to more secure code, as it’s easier to verify correctness than in systems where the rules are implicit.

---

## 7. Implementation Considerations in Go/TinyGo

Since our first ual compiler implementation generates Go/TinyGo code, the following internal behaviors are proposed:

### 7.1 Atomic Guarantees

- **Atomic Operations:** Underlying operations for `push`, `pop`, and `borrow` must use Go’s atomic primitives to ensure that the data transfers and state changes are thread-safe.
- **Typed Stacks:** The generated code will create typed stack structures in Go that enforce type safety, similar to Go’s built-in type system but extended with explicit ownership markers.

### 7.2 Compiler Behavior

The ual compiler must enforce:
- **Scope Boundaries:** Borrowed segments are tied to explicit scope blocks. Once the scope ends, any attempt to access the borrowed segment must be flagged as an error.
- **Non-Mutability Guarantee:** For read-only borrows, the compiler should insert checks (or assume via static analysis) that the parent stack is not modified during the borrow’s lifetime.
- **Lifetime Visibility:** Although not as automated as Rust’s lifetime inference, the compiler should track borrow scopes and generate warnings if there’s any risk of the borrowed segment becoming invalid due to modifications on the parent stack.
- **Conversion to Go:** The ual-to-Go code generator will map these constructs to equivalent Go code, using channels, atomic variables, or custom structs as needed to simulate the borrowed segment behavior.

### 7.3 Internal API Example in Go

An internal representation in Go might look like this:

```go
type Stack[T any] struct {
    data []T
    mu   sync.RWMutex
}

func (s *Stack[T]) Push(item T) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data = append(s.data, item)
}

func (s *Stack[T]) Borrow(start, end int) []T {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // Return a slice that represents the borrowed segment.
    // The caller must not modify the underlying data.
    return s.data[start:end]
}
```

The generated code from ual would wrap these operations, ensuring that the borrow is only valid within a defined scope.

---

## 8. Conclusion

Borrowed stack segments extend ual’s container-centric philosophy by providing a mechanism for zero-copy, temporary views into a stack. This approach—while reminiscent of Rust’s slice borrowing—remains explicit and visible, resulting in code that is easier to read and verify. 

By making safety assurances part of the visible code (every data transfer is explicit, and borrowed segments have clear lifetimes), ual offers a secure, predictable, and efficient model for embedded systems. The design complements Go/TinyGo’s inherent type safety with additional atomic guarantees and controlled scope operations.

Ultimately, while Rust’s advanced borrow-checking is powerful, ual’s explicit borrowed stack segments provide many of the same operational benefits—enhancing safety and clarity without the extra complexity. The proposed internal implementations and compiler behaviors ensure that these borrowed segments integrate smoothly into the overall ual ecosystem, delivering robust, secure code in resource-constrained environments.