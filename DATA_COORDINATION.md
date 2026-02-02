# ual Data Coordination: A Complete Guide

## Introduction

This document describes ual's approach to heterogeneous data — how to work with values of different types together while preserving type safety, avoiding runtime tags, and maintaining dense storage.

The approach rests on a philosophical foundation: **relations are real, but objects are not**. You can name and work with coordinations of typed data without reifying them as runtime values. This enables heterogeneous operations without the costs that traditional languages pay.

---

## Part I: The Problem

### What Many Languages Struggle With

Consider a simple task: pass several values of different types to a function, process them together, and preserve type safety throughout.

Many mainstream procedural languages — particularly those in the C lineage like Go, Java, and C itself — handle this poorly. (Languages with algebraic data types, like Rust or OCaml, have better answers but pay different costs. The analysis here focuses on the procedural mainstream.)

In Go, you might try:

```go
func processConfig[T any](args ...T) []T {
    return args
}

processConfig("HOST", 8080, true)  // Does this work?
```

It doesn't. Go's generics require all variadic arguments to share the same type `T`. Mixed types force the compiler to widen `T` to `any`, losing all type information.

Java fares no better:

```java
static <T> List<T> process(T... args) {
    return Arrays.asList(args);
}

process("HOST", 8080, true);  // T becomes Object
```

The compiler infers `T = Object`. Your types are gone.

C++ solves this with variadic templates:

```cpp
template<typename... Ts>
auto processConfig(Ts&&... args) {
    return std::make_tuple(std::forward<Ts>(args)...);
}

auto config = processConfig("HOST", 8080, true);
// type: tuple<const char*, int, bool>
```

This preserves types, but at the cost of template metaprogramming complexity and notoriously cryptic error messages.

Dynamic languages like Python or Elixir sidestep the issue by abandoning static typing altogether.

### Why Is This So Hard?

The difficulty stems from a fundamental assumption in procedural languages: **values can escape their lexical scope**. You can return them, store them in data structures, pass them by reference. Because values travel, they must carry their types with them — either explicitly (tagged unions, boxed values) or implicitly (type erasure to a common supertype).

When you want heterogeneous collections, you face an unpleasant choice:

1. **Tag every value** — runtime overhead, loss of static guarantees
2. **Erase types to a common ancestor** — loss of type information
3. **Complex type machinery** — templates, HLists, type-level programming

Most languages pick one of these compromises and live with the consequences.

### A Different Question

What if we asked a different question?

Instead of "how do I pass mixed types to a function?", what if we asked "why are mixed types going to the same place?"

When you call `processConfig("HOST", 8080, true)`, you're passing three values that happen to be syntactically adjacent. But semantically, they're different things: a name (string), a port (integer), a flag (boolean). They have different meanings, different operations, different lifetimes.

Cramming them into one parameter list — and then struggling to recover their types — is solving the wrong problem.

---

## Part II: The Four Core Concepts

ual handles heterogeneous data through four interconnected concepts, each with its own keyword:

| Keyword | What it is | What it does |
|---------|------------|--------------|
| **`relation`** | A named coordination | Connects existing typed stacks |
| **`constitution`** | A named structure | Creates typed stacks, defines a shape |
| **`union`** | A set of constitutions | Allows one of several shapes |
| **`collection`** | An ordered aggregation | Stores multiple instances, preserves order |

### The Four Keywords Together

These four keywords form a coherent family:

```
-- Coordinates existing stacks
relation person(@names, @ages, @scores)

-- Defines and creates structure
constitution button {
    label: string
    onclick: action
}

constitution radio {
    label: string
    group: string
    selected: bool
}

-- Combines constitutions into one type
union widget { button, radio }

-- Ordered heterogeneous storage
collection @panel of widget
```

Each keyword announces what kind of thing you're defining:
- `relation` — coordination of existing stacks
- `constitution` — structure that creates its own stacks
- `union` — alternation between constitutions
- `collection` — aggregation of instances

### Concept 1: Relation

#### What It Is

A relation coordinates existing typed stacks. It names their association without creating new storage.

#### Syntax

```
@names = stack.new(string)
@ages = stack.new(i64)
@scores = stack.new(f64)

relation person(@names, @ages, @scores)
```

Or inline:

```
person := @names, @ages, @scores
```

#### What It Does

- Names a coordination of stacks
- Defines their order for operations
- Enables multi-push and zip iteration
- Type-checks each position

#### What It Is Not

A relation is **not a type**. You cannot:

```
@people = stack.new(person)      -- ERROR: relation is not a type
send(p: person)                   -- ERROR: cannot pass a relation
widget := person | employee       -- ERROR: cannot union relations
```

A relation exists only to coordinate. It has no instances, no identity, no runtime representation.

#### Operations

**Multi-push (atomic):**
```
person <- "Alice", 30, 95.5
person <- "Bob", 25, 87.3
```

Multi-push is atomic: all values are appended together, or none are. This guarantees that coordinated stacks remain equal length even if an error occurs. The operation desugars to:
```
@names < "Alice"
@ages < 30
@scores < 95.5
```
But the implementation ensures all three appends succeed or the relation remains unchanged.

**Iteration (zip):**
```
for person {|name, age, score|
    -- name: string, age: i64, score: f64
    -- bindings exist only in this block
}
```

Iterates in lockstep across all coordinated stacks. Bindings are confined to the block.

#### Equal Length Invariant

Coordinated stacks maintained through atomic multi-push will have equal lengths — this is the expected case.

Zip stopping at shortest is the defined behaviour when a relation coordinates arbitrary existing stacks that may differ in length — iteration stops when any coordinated stack is exhausted. It's a fail-safe, not the normal case. If you're using atomic multi-push consistently, your stacks will be aligned and zip will process all elements.

#### Problems Solved

| Problem | Example |
|---------|---------|
| Heterogeneous variadics | `config <- "HOST", 8080, true` |
| Parallel data structures | Names, ages, scores for same entities |
| Coordinated iteration | Process related values together |
| Type-safe multi-push | Compiler verifies each position |

#### Example: Configuration

```
@keys = stack.new(string)
@values = stack.new(string)
@defaults = stack.new(string)

relation env_config(@keys, @values, @defaults)

env_config <- "HOST", "localhost", "127.0.0.1"
env_config <- "PORT", "8080", "3000"
env_config <- "DEBUG", "true", "false"

for env_config {|key, value, default|
    @output < key + "=" + value
}
```

---

### Concept 2: Constitution

#### What It Is

A constitution defines a structure and creates the stacks to hold it. The relation is primary; the stacks exist because of it.

#### Syntax

```
constitution tcp_packet {
    source_port: u16
    dest_port: u16
    sequence: u32
    ack_number: u32
    flags: u8
    window: u16
    checksum: u16
    payload: bytes
}
```

#### What It Does

- Defines a named structure
- Creates typed stacks for each field
- Acts as a type for stack creation
- Can be instantiated, unioned, collected

#### What It Is

A constitution **is a type**. You can:

```
@packets = stack.new(tcp_packet)          -- stack of tcp_packet
@packets <- tcp_packet(443, 8080, ...)    -- push instance
union widget { button, radio }            -- union of constitutions
collection @panel of widget               -- collection of union
```

#### What It Is Not

A constitution instance is not an object. Instances don't have identity. You can't pass a constitution instance as a standalone value outside of reification.

#### Length Invariant

All field stacks of a constitution have identical length at all times. Constitution instances are appended atomically.

```
-- VALID: atomic append
@packets <- tcp_packet(443, 8080, 12345, 0, 0x02, 65535, 0, data)

-- INVALID: per-field writes risk ragged columns
tcp_packet.source_port < 443   -- ERROR: use atomic append
```

Per-field access is for reading; writing is always full-row.

#### Problems Solved

| Problem | Example |
|---------|---------|
| Fixed-structure records | Phone records, log entries |
| Network packets | TCP, UDP, custom protocols |
| Database rows | Query results with known schema |
| Nested structures | Constitution containing constitution |

#### Example: Phone Records

```
constitution phone_record {
    caller: string
    callee: string
    start_time: i64
    duration: i64
    cell_tower: string
}

@calls = stack.new(phone_record)

@calls <- phone_record("Alice", "Bob", 1706541234, 342, "Tower-7A")
@calls <- phone_record("Bob", "Carol", 1706541300, 120, "Tower-3B")

for @calls {|caller, callee, start, dur, tower|
    -- process each record
}
```

#### Example: Nested Structure

```
constitution address {
    street: string
    city: string
    postcode: string
}

constitution person {
    name: string
    age: i64
    home: address
}

@people = stack.new(person)

for @people {|name, age, home|
    for home {|street, city, postcode|
        -- nested reification
    }
}
```

---

### Concept 3: Union

#### What It Is

A union defines a set of constitutions. A value of the union type is one of those constitutions, but which one varies per instance.

#### Syntax

```
constitution button {
    label: string
    onclick: action
}

constitution radio {
    label: string
    group: string
    selected: bool
}

constitution dropdown {
    items: [string]
    selected: i64
}

union widget { button, radio, dropdown }
```

The `union` keyword declares that `widget` can be any of the listed constitutions.

#### What It Does

- Combines multiple constitutions into one type
- Allows heterogeneous collections of known shapes
- Enables matching to determine shape
- Preserves type safety (no type erasure)

#### Matching: By Identity, Not Structure

Matches on the **active constitution variant**. The discriminant is stored separately from field data, preserving typed dense storage. Match arms are checked against constitution identity, not inferred from field shape.

This means:
- Two constitutions with identical fields are still distinct
- No ambiguity possible
- Evolution is by constitution identity

```
for @widgets {
    match button {|label, onclick|
        -- this is a button
    }
    match radio {|label, group, selected|
        -- this is a radio
    }
    match dropdown {|items, selected|
        -- this is a dropdown
    }
}
```

#### Problems Solved

| Problem | Example |
|---------|---------|
| GUI containers with mixed children | Panel with buttons, radios, dropdowns |
| Document elements | Text nodes, paragraphs, divs |
| AST nodes | Expressions, statements, declarations |
| Protocol messages | Different message types in one stream |

#### Example: GUI Panel

```
constitution button {
    label: string
    onclick: action
}

constitution radio {
    label: string
    group: string
    selected: bool
}

constitution checkbox {
    label: string
    checked: bool
}

widget := button | radio | checkbox

@panel = stack.new(widget)

@panel <- button("OK", on_ok)
@panel <- radio("Option A", "group1", true)
@panel <- checkbox("Remember me", false)
@panel <- button("Cancel", on_cancel)

for @panel {
    match button {|label, onclick|
        render_button(label, onclick)
    }
    match radio {|label, group, selected|
        render_radio(label, group, selected)
    }
    match checkbox {|label, checked|
        render_checkbox(label, checked)
    }
}
```

---

### Concept 4: Collection

#### What It Is

A collection is an ordered aggregation of constitution instances (or union instances). It explicitly preserves insertion order and supports heterogeneous elements via unions.

#### Syntax

```
collection @children of widget
```

Or for homogeneous:
```
collection @items of button
```

#### What It Does

- Stores instances in order
- Supports unions (heterogeneous collections)
- Enables ordered iteration
- Preserves insertion sequence

#### Difference from Stack

In ual, "stack" refers to a typed, appendable collection with configurable access patterns called **perspectives**. A perspective controls how you access the underlying storage — LIFO (stack), FIFO (queue), indexed (array), or hash (map) — without changing the underlying dense column layout. Some perspectives (like hash) use auxiliary structures such as indices to provide efficient lookup. The default perspective is LIFO, but others are available.

"Collection" specifically means an ordered heterogeneous aggregation via union, where insertion order is always preserved.

| Aspect | Stack | Collection |
|--------|-------|----------|
| Primary purpose | Typed storage with perspectives | Ordered heterogeneous aggregation |
| Heterogeneous | No (single type) | Yes (via union) |
| Order guarantee | Depends on perspective | Always insertion order |

#### Problems Solved

| Problem | Example |
|---------|---------|
| Ordered mixed children | GUI layouts, document structure |
| HTML building | Mixed text, elements, components |
| Event streams | Different event types in order |
| Command lists | Mixed command types, execution order matters |

#### Example: HTML Builder

```
constitution text_node {
    content: string
}

constitution paragraph {
    class: string
    content: string
}

constitution div {
    class: string
    children: collection of block_element
}

union block_element { text_node, paragraph, div }

collection @doc of block_element

@doc <- text_node("Hello, world!")
@doc <- paragraph("intro", "Welcome to the site.")
@doc <- div("container", nested_content)
@doc <- text_node("Goodbye!")

for @doc {
    match text_node {|content|
        emit(content)
    }
    match paragraph {|class, content|
        emit("<p class='" + class + "'>" + content + "</p>")
    }
    match div {|class, children|
        emit("<div class='" + class + "'>")
        for children { ... }
        emit("</div>")
    }
}
```

---

## Part III: Type Confinement

### The Principle

ual confines certain type complexities to lexical blocks rather than allowing them to escape as values. This is the key to enabling heterogeneous operations without the costs of tagged values or type erasure.

### `.compute()` — Arithmetic Confinement

ual's default operations are stack-based — you push, pop, and manipulate values through perspectives. Native arithmetic (direct operations on f64, i64, etc.) isn't the default mode. When you need native numeric computation, you enter a `.compute()` block:

```
.compute(f64) {
    var x f64 = 3.14159
    var y f64 = x * x
    push:y
}
```

Inside: native typed operations. Outside: stack values. The native types don't escape — only the computed result, pushed to a typed stack.

### `.tuple()` — Heterogeneous Confinement

`.tuple()` confines heterogeneous bindings to a block:

```
.tuple(@names, @ages, @scores) {|name, age, score|
    -- name: string, age: i64, score: f64
    -- can work with all three together
    -- can call helpers with individual values
    -- can push results to typed stacks
    process_person(name, age)
    @output < name + ": " + itoa(age)
}
```

Inside: typed bindings from multiple sources. Outside: typed stacks. The tuple doesn't escape.

### Reification and Confinement

When you write:

```
for config {|name, port, flag|
    -- body
}
```

You're doing two things:

1. **Reifying** the relation — making it active, drawing values from the stacks
2. **Confining** the bindings — `name`, `port`, `flag` exist only inside this block

The bindings are typed variables. Inside the block, you have full typed access. But you cannot extract the "tuple" as a value:

```
for config {|name, port, flag|
    @stuff < (name, port, flag)   -- ERROR: no tuple type exists
    return (name, port, flag)     -- ERROR: cannot return tuple
}
```

There's no syntax for reifying the heterogeneous group as a value. The only way to preserve something is to push it to an appropriately typed stack.

### Why Confinement Matters

Traditional languages let values escape scope. Because values travel, types must travel with them — leading to tagged unions, boxing, and type erasure.

ual inverts this: **heterogeneity is a lexical phenomenon, not a runtime phenomenon**. You can work with multiple types together in a confined scope. But you cannot store heterogeneous bundles that escape and require runtime typing.

The complexity stays local. The rest of the system remains simple typed stacks.

### Working Within Confinement

Confinement doesn't mean you can't do useful work. Inside a `.tuple()` or `for` block, you can:

- Perform any typed operations on the bindings
- Call functions with individual typed values
- Push results to typed stacks
- Build new data structures

```
.tuple(@widgets) {|w|
    match button {|label, onclick|
        -- call helper with typed values
        process_button(label, onclick)
        -- push to typed stacks
        @labels < label
        @handlers < onclick
    }
}
```

The helper receives typed values:

```
fn process_button(label: string, onclick: action) {
    -- receives typed values, not a "button record"
}
```

You pass the **individual typed values**, not the bundle. This is not a limitation — it's what enables the safety guarantees.

### Summary: Type Confinement

| Construct | What's confined | What escapes |
|-----------|-----------------|--------------|
| `.compute()` | Native arithmetic types | Result pushed to stack |
| `.tuple()` | Heterogeneous bindings | Individual values pushed to typed stacks |
| `for relation` | Zipped bindings | Individual values pushed to typed stacks |
| `match` | Union variant bindings | Individual values pushed to typed stacks |

---

## Part IV: Relations vs Objects

### The Philosophical Foundation

Most languages treat "real" as synonymous with "reified." If something matters, it must be an object you can manipulate.

ual separates these concerns:

- **Stacks are real** — they hold typed values, exist at runtime
- **Relations are real** — they constrain operations, determine type flow
- **Relations are not objects** — they cannot be stored, passed, or escape

A relation is:
- **Real** — it affects program behaviour
- **Named** — you can refer to it
- **Not an object** — it has no memory address, no identity

### Two Kinds of Relations

#### Coordination

A coordination relation connects existing stacks:

```
@names = stack.new(string)
@ages = stack.new(i64)

relation person(@names, @ages)
```

The stacks exist independently. The relation names their coordination. You could use the same stacks in different relations:

```
relation person(@names, @ages)
relation employee(@names, @salaries)
```

`@names` participates in both. The stacks are primary; the relations are views.

#### Constitution

A constitution relation creates its members:

```
constitution tcp_packet {
    source_port: u16
    dest_port: u16
    ...
}
```

The stacks don't exist independently — they exist *because* the constitution declares them. The relation is primary; the stacks are derived.

### The Distinction

| Aspect | Relation (Coordination) | Constitution |
|--------|------------------------|--------------|
| Creates stacks | No | Yes |
| Is a type | No | Yes |
| Can have instances | No | Yes |
| Can be in a union | No | Yes |
| Can be in a collection | No | Yes |
| Coordinates existing stacks | Yes | No |
| Reified with `.tuple()`/`for` | Yes | Yes |

### Why This Matters

The keywords enforce the semantics:

- **Relation**: "I relate these things" — the things exist independently
- **Constitution**: "I constitute these parts" — the parts exist because of me

You can't misuse a relation as a type because the language doesn't let you. A reader seeing `constitution` knows: this defines a structure. A reader seeing `relation` knows: this coordinates existing stacks.

---

## Part V: Implementation

### Relation

A relation compiles to a struct holding pointers to the coordinated slices:

```go
type PersonRelation struct {
    names  *[]string
    ages   *[]int64
    scores *[]float64
}

func (r *PersonRelation) Push(name string, age int64, score float64) {
    *r.names = append(*r.names, name)
    *r.ages = append(*r.ages, age)
    *r.scores = append(*r.scores, score)
}

func (r *PersonRelation) Each(fn func(string, int64, float64)) {
    n := min(len(*r.names), len(*r.ages), len(*r.scores))
    for i := 0; i < n; i++ {
        fn((*r.names)[i], (*r.ages)[i], (*r.scores)[i])
    }
}
```

No runtime overhead beyond pointer indirection.

### Constitution

A constitution compiles to multiple typed slices, one per field:

```go
type TcpPacketStore struct {
    source_port []uint16
    dest_port   []uint16
    sequence    []uint32
    // ...
}

func (s *TcpPacketStore) Push(src, dst uint16, seq uint32, ...) {
    s.source_port = append(s.source_port, src)
    s.dest_port = append(s.dest_port, dst)
    s.sequence = append(s.sequence, seq)
    // ...
}
```

Instances are indices into these parallel slices. All slices always have equal length.

### Union

A union adds a discriminant per element:

```go
type WidgetStore struct {
    kinds     []uint8  // 0=button, 1=radio, 2=dropdown
    indices   []int    // index into the appropriate store
    buttons   ButtonStore
    radios    RadioStore
    dropdowns DropdownStore
}
```

The discriminant is stored separately from the field data in a compact array. This is **not** the same as tagged values:

- Traditional tagging: each value carries its type inline (`[tag|data][tag|data][tag|data]`)
- ual approach: discriminants in one array, data in typed arrays (`[tag,tag,tag] + [data,data,data]`)

Values remain untagged in their typed slices. The discriminant array tells you which constitution each slot belongs to, but the values themselves don't self-describe.

### Collection

A collection is a union store with guaranteed insertion order:

```go
type WidgetCollection struct {
    order []struct {
        kind  uint8
        index int
    }
    buttons   ButtonStore
    radios    RadioStore
    dropdowns DropdownStore
}
```

Iteration follows the order slice, dispatching to the appropriate store.

### Performance Characteristics

#### Memory Layout

Traditional approach (tagged values):

```go
type Value struct {
    Tag  uint8
    Data any  // boxed
}

values := []Value{
    {Tag: STRING, Data: "HOST"},  // allocation
    {Tag: INT, Data: 8080},       // allocation
    {Tag: BOOL, Data: true},      // allocation
}
```

Each value is boxed. Each box is a separate allocation. Memory is scattered.

ual approach (typed slices):

```go
keys := []string{"HOST", "DB", "CACHE"}
ports := []int64{8080, 5432, 6379}
flags := []bool{true, false, true}
```

Three dense arrays. Contiguous memory. Cache-friendly.

#### Append Cost

ual requires multiple appends — one per field:

```
config <- "HOST", 8080, true   // three appends
```

But each append is smaller and cheaper:

| Approach | Per-record cost |
|----------|-----------------|
| Tagged values | 1 struct + N boxes (heap allocations) |
| Typed slices (reserved) | 0 allocations |
| Typed slices (not reserved) | Amortised O(1) |

With pre-allocation:

```
@names = stack.new(string, capacity: 1000)
@ports = stack.new(i64, capacity: 1000)
@flags = stack.new(bool, capacity: 1000)
```

Each append becomes a simple write and length increment. Nanoseconds.

#### Iteration Cost

Typed slices iterate faster than maps or tagged collections:

```go
// Typed slices: sequential memory access
for i := 0; i < n; i++ {
    process(keys[i], ports[i], flags[i])
}
```

The CPU prefetcher loves sequential access. Cache misses are minimised. No pointer chasing. No unboxing.

---

## Part VI: What These Concepts Solve

| Use Case | Solution |
|----------|----------|
| Heterogeneous variadics (config) | `relation` coordinating typed stacks |
| Phone records, log entries | `constitution` + stack |
| TCP/UDP packets | `constitution` (possibly nested) |
| HTTP headers (variable-length, homogeneous) | `constitution header` + stack |
| GUI panel with mixed widgets | `union` + `collection` |
| HTML builder (text, paragraphs, divs) | `union` + `collection` |
| Nested documents | `union` + `collection` |
| JSON/XML with known schema | Nested `constitution` |
| Database pagination | `constitution` + pre-allocated stacks |

---

## Part VII: What These Concepts Do Not Solve

### Runtime-Determined Structure

If the shape of your data depends on runtime values — fields discovered at execution time — these concepts don't help. They require structure known at compile time.

```
-- NOT POSSIBLE:
for field in runtime_fields {
    constitution.add_field(field.name, field.type)
}
```

### Arbitrary JSON/XML

Documents with unknown schemas cannot be modelled as constitutions. You don't know the fields until you parse the document.

**Solution:** Query interfaces. Treat unknown documents as external data sources, extract values by path into typed stacks.

```
@doc = json.parse(input)

@names = stack.new(string)
@ages = stack.new(i64)

@names < @doc.string("name")
@ages < @doc.i64("age")
```

The document remains opaque. The extracted values are typed.

### Key-Based O(1) Lookup

Relations and collections are optimised for sequential access and iteration. O(1) lookup by key requires an index.

**Solution:** Use a hash perspective (which uses an index), or build an explicit map from key to position. Data stays in dense storage; the index provides random access.

```
@names = stack.new(string)
@values = stack.new(i64)

relation config(@names, @values)

-- build index separately
index := map.new(string, i64)
index["HOST"] = 0
index["PORT"] = 1

-- O(1) lookup
pos := index["HOST"]
-- then access @names[pos], @values[pos]
```

### Escaping Heterogeneous Values

You cannot extract a tuple or union instance and pass it around freely. Heterogeneity is confined to reification blocks.

```
for @widgets {
    match button {|label, onclick|
        return (label, onclick)    -- ERROR: cannot escape
    }
}
```

**Solution:** Push individual values to appropriately typed stacks, or pass individual typed values to functions. The confinement is intentional — it's what makes the system safe.

### Circular References

Constitutions cannot directly reference themselves by identity. There's no pointer, no object identity.

```
constitution node {
    value: i64
    parent: node        -- PROBLEMATIC: which instance?
}
```

**Solution:** Use indices for relationships.

```
constitution node {
    value: i64
    parent_index: i64   -- index into @nodes stack, or -1
}

@nodes = stack.new(node)
```

### Mutation and Deletion (Future Work)

The current model defines creation and iteration but leaves mutation and deletion unspecified. These operations raise questions that need careful answers:

**Indexed update:**
- Can you write `@packets[i].flags = 0x10`?
- If yes, what are the borrowing/aliasing rules?
- If no, is copy-on-write or rebuild the idiom?

**Deletion from constitutions:**
- If you delete row `i`, do indices shift?
- Are stable indices required? (Implies tombstones or indirection)
- Is compaction explicit or automatic?

**Deletion from collections:**
- Removing from a union collection affects the order array
- Do per-variant stores compact independently?
- What happens to indices held elsewhere?

**Likely directions:**

1. **Append-only by default.** Many use cases (event logs, packet streams, batch processing) don't need mutation. Append-only is simple and fast.

2. **Explicit rebuild for mutation.** Filter and reconstruct rather than mutate in place. Functional idiom, no aliasing issues.

3. **Opt-in mutable stores.** A separate `mutable constitution` or `buffer` type with defined borrowing rules for cases that genuinely need update.

4. **Index stability via indirection.** If stable indices matter, store an indirection layer. Cost is one extra lookup; benefit is indices remain valid after deletion.

These decisions will be made based on real use cases. The current model is complete for creation, coordination, and iteration — the most common operations in ual's target domain of orchestration and data processing.

---

## Part VIII: Quick Reference

### Syntax Summary

```
-- Typed stack
@names = stack.new(string)
@names = stack.new(string, capacity: 100)

-- Relation (coordination)
relation person(@names, @ages, @scores)
person := @names, @ages, @scores

-- Constitution (structure)
constitution tcp_packet {
    source_port: u16
    dest_port: u16
    payload: bytes
}

-- Union
union widget { button, radio, dropdown }

-- Collection
collection @panel of widget

-- Multi-push (relation)
person <- "Alice", 30, 95.5

-- Instance push (constitution)
@packets <- tcp_packet(443, 8080, data)

-- Iteration (relation)
for person {|name, age, score|
    -- confined bindings
}

-- Iteration (constitution)
for @packets {|src, dst, payload|
    -- confined bindings
}

-- Iteration with match (union/collection)
for @panel {
    match button {|label, onclick| ... }
    match radio {|label, group, selected| ... }
}

-- Type confinement
.tuple(@a, @b, @c) {|x, y, z|
    -- heterogeneous bindings confined here
}

.compute(f64) {
    -- native arithmetic confined here
}
```

### Decision Guide

| I need to... | Use |
|--------------|-----|
| Coordinate existing stacks | `relation` |
| Define a reusable structure | `constitution` |
| Allow one of several shapes | `union` |
| Store ordered mixed elements | `collection` |
| Store multiple instances of one shape | `stack.new(constitution)` |
| Work with heterogeneous data temporarily | `.tuple()` or `for` with relation |
| Handle unknown schemas | Query interface (external) |
| Look up by key | Add an index (map) |

### What Exists at Runtime

| ual Concept | Runtime Representation |
|-------------|------------------------|
| Typed stack | `[]T` (dense slice) |
| Relation | Struct with pointers to slices |
| Constitution | Multiple parallel slices |
| Union | Discriminant slice + per-variant stores |
| Collection | Order slice + union stores |
| Tuple bindings | Local variables (block-scoped) |

### What Doesn't Exist at Runtime

- Tuple values
- Per-value type tags (values don't self-describe)
- Boxed primitives
- Relation objects
- Escaped heterogeneous bundles

Note: Union discriminants **do** exist at runtime, but they're stored separately from field data in compact arrays — not inline with each value. This preserves dense typed storage while enabling variant dispatch.

---

## Conclusion

ual's approach to heterogeneous data rests on a simple insight: **the problems with mixed types stem from letting them escape**.

Traditional languages struggle because values travel freely, so types must travel with them — leading to tagging, boxing, and erasure.

ual confines heterogeneity to lexical blocks. Inside, you have full typed access. Outside, only typed stacks exist. Relations coordinate without reifying. Constitutions define structure without creating escapable objects. Unions allow variation without type erasure. Collections preserve order without sacrificing type safety.

The result: dense storage, static types, predictable iteration, and no per-value tags. The sophistication is in the language semantics. The generated code is simple.

Relations without objects. Heterogeneity without escape. Types without boxing.