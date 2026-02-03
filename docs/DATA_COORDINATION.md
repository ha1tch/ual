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

Volume III explores this pattern in depth — from simple projection through batch ingestion, known-schema mapping, and handling of unknown keys and optional fields.

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

---
---

# Volume II: Iteration Patterns

*This section documents iteration patterns that build on the four core concepts. These are ergonomic additions — they introduce no new primitives, only named patterns over existing ones.*

---

## Nested Traversal

### The Problem

Collections can contain constitutions that themselves contain collections. The HTML builder example from Volume I demonstrates this:

```
constitution div {
    class: string
    children: collection of block_element
}

union block_element { text_node, paragraph, div }
```

A `div` contains a collection of `block_element`, which can itself include more `div` nodes. Traversing this requires nested `for` + `match` blocks:

```
for @doc {
    match text_node {|content|
        process(content)
    }
    match paragraph {|class, content|
        process(class, content)
    }
    match div {|class, children|
        process(class)
        for children {
            match text_node {|content|
                process(content)
            }
            match paragraph {|class, content|
                process(class, content)
            }
            match div {|class, children|
                -- and again, and again...
            }
        }
    }
}
```

For deeply nested structures, this is verbose and easy to make inconsistent — forgetting a recursive case, traversing the wrong containment field, or handling a variant differently at different depths.

### `.nested()` — Declared Containment Traversal

`.nested()` is a compile-time-checked traversal that iterates a collection and any collections reachable through explicitly declared containment fields, in pre-order depth-first order.

```
for @doc.nested() {|node|
    match text_node {|content|
        process(content)
    }
    match paragraph {|class, content|
        process(class, content)
    }
    match div {|class|
        process(class)
    }
}
```

The bound name `node` is a confined reification of the current union element — it exists only to drive `match` within this block. It is not a value, cannot be stored, returned, or pushed to a stack.

### The `nested` Field Annotation

`.nested()` does **not** discover structure at runtime. It follows fields explicitly marked as containment points:

```
constitution div {
    class: string
    children: nested collection of block_element
}
```

The `nested` annotation tells the compiler: "this field is a containment boundary — `.nested()` should follow it."

Only fields annotated with `nested` are traversed. A constitution can have multiple collection fields, but only those marked `nested` participate in `.nested()` iteration.

```
constitution section {
    title: string
    body: nested collection of block_element   -- traversed by .nested()
    footnotes: collection of footnote           -- NOT traversed by .nested()
}
```

This is field-directed, not type-directed. The compiler knows exactly which paths to follow.

### What `.nested()` Is

`.nested()` is **desugaring, not magic**. It is defined as equivalent to a specific expansion of nested `for` + `match` blocks. The compiler generates the expansion; the programmer sees the shorthand.

Properties:

- **Traversal order**: Pre-order depth-first (visit node, then nested children)
- **Traversal targets**: Statically known from `nested` field annotations
- **Bindings**: Still confined to the iteration block
- **No new values created**: Same reification rules as `for`
- **Mechanical expansion**: If it can't be desugared, it's rejected

### What `.nested()` Is Not

`.nested()` does **not**:

- Inspect fields dynamically
- Traverse unknown schemas
- Follow arbitrary pointers or indices
- Create new relations at runtime
- Weaken confinement or let heterogeneity escape
- Provide runtime depth information
- Support wrapping patterns (open/close)

It is a higher-level iterator, like zip, but for declared containment.

### What `.nested()` Does Not Handle: Wrapping

Many tree operations need to execute code both **before and after** visiting children:

```
-- Desired but NOT supported by .nested():
emit("<div>")
    -- visit children
emit("</div>")
```

This is a wrapping pattern. It requires enter/exit semantics — the ability to run code at two points for the same node. A flat pre-order traversal visits each node once; it cannot express "do X, then visit children, then do Y."

**For wrapping, use explicit nested iteration:**

```
for @doc {
    match text_node {|content|
        emit(content)
    }
    match paragraph {|class, content|
        emit("<p class='" + class + "'>" + content + "</p>")
    }
    match div {|class, children|
        emit("<div class='" + class + "'>")
        for children {
            -- recurse manually
        }
        emit("</div>")
    }
}
```

`.nested()` is for **flat operations** over tree structures: searching, counting, validating, extracting. When you need structural awareness of depth or wrapping, write the `for` + `match` explicitly — as shown in the HTML builder example in Volume I, Part II (Concept 4: Collection).

This creates a natural division of labour:
- `.nested()` for **analysis and extraction** — flat passes over the whole tree
- Explicit `for` + `match` for **rendering and wrapping** — structural emission with open/close semantics
- `.nested_events()` (see Open Questions) may bridge this gap in the future

### Recursion Safety

Because collections can contain constitutions that reference the same union, recursive containment is possible:

```
union block_element { text_node, paragraph, div }

constitution div {
    class: string
    children: nested collection of block_element   -- div is in block_element
}
```

This is a valid recursive structure — not a graph cycle (collections own their contents; there are no reference edges back to ancestors). The compiler must still handle unbounded depth:

- **Static detection**: The compiler identifies recursive `nested` paths at compile time
- **Depth limit rule**: If the compiler can prove the containment graph is acyclic, `max_depth` is optional. If recursion is possible, `max_depth` is required — without it, `.nested()` is rejected at compile time.

```
for @doc.nested(max_depth: 64) {|node|
    ...
}
```

If the depth limit is reached, traversal stops for that branch. No error, no panic — the branch is simply not followed further.

### Use Cases

| Use Case | Why `.nested()` fits |
|----------|---------------------|
| Count all nodes in a document | Flat traversal, no wrapping needed |
| Search for a node by content | Visit each node once, check condition |
| Validate structure | Check every node against rules |
| Extract all text content | Collect from all text_node matches |
| Flatten a tree to a list | Push each node's data to typed stacks |
| Calculate total size | Accumulate across all nodes |

| Use Case | Why `.nested()` doesn't fit |
|----------|---------------------------|
| HTML rendering (open/close tags) | Requires wrapping |
| Pretty-printing with indentation | Requires depth awareness |
| Tree transformation (map) | Requires building new structure |
| Event bubbling in GUI | Requires bottom-up traversal |

### Example: Extract All Text

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
    children: nested collection of block_element
}

union block_element { text_node, paragraph, div }

collection @doc of block_element

@all_text = stack.new(string)

for @doc.nested() {|node|
    match text_node {|content|
        @all_text < content
    }
    match paragraph {|class, content|
        @all_text < content
    }
}
```

### Example: Count Nodes by Type

```
.compute(i64) {
    var text_count i64 = 0
    var para_count i64 = 0
    var div_count i64 = 0

    for @doc.nested() {|node|
        match text_node {|content|
            text_count = text_count + 1
        }
        match paragraph {|class, content|
            para_count = para_count + 1
        }
        match div {|class|
            div_count = div_count + 1
        }
    }

    push:text_count
    push:para_count
    push:div_count
}
```

### Implementation

`.nested()` expands to a recursive iteration function at compile time. The compiler lowers `.nested()` to a generated traversal over the order vector and per-variant stores. The user never sees discriminants or indices — only confined bindings via `match`. Internally, the generated code is equivalent to:

```go
func traverseBlockElement(items *BlockElementCollection, depth, maxDepth int, fn func(kind uint8, index int)) {
    if depth > maxDepth {
        return
    }
    for _, entry := range items.order {
        fn(entry.kind, entry.index)
        if entry.kind == KIND_DIV {
            children := items.divs.children[entry.index]
            traverseBlockElement(children, depth+1, maxDepth, fn)
        }
    }
}
```

The generated function:
- Follows only `nested`-annotated fields
- Respects the depth limit
- Visits in pre-order (node before children)
- Has no runtime schema inspection

### Design Rationale

`.nested()` exists because:

1. **Nested structures are common** — documents, GUI trees, ASTs
2. **Flat traversal is the most common operation** — search, count, validate, extract
3. **Explicit nesting is verbose** — repeating `for` + `match` at every level is easy to make inconsistent
4. **It desugars completely** — no new runtime capability, just iteration sugar
5. **It respects confinement** — bindings inside `.nested()` follow the same rules as `for` and `.tuple()` (see Volume I, Part III: Type Confinement)

It does not exist for:
1. **Wrapping** — use explicit `for` + `match`
2. **Transformation** — building new trees requires explicit construction
3. **Dynamic structures** — unknown schemas can't be traversed statically

---

## Open Questions for Future Iteration Patterns

The following patterns may be useful but are not yet specified:

**`.nested_events()` / `.walk(pre, post)`**: Enter/exit event traversal. Yields each node twice — once on entry, once on exit — enabling wrapping patterns (open/close tags, indentation) without depth tracking or runtime tree objects. The user supplies two blocks: one for enter, one for exit. Indentation can be computed inside `.compute()` by incrementing/decrementing a local counter on enter/exit. This is the most likely candidate for early adoption, as it would make the HTML rendering example ergonomic without compromising confinement.

**`.breadth_first()`**: Level-order traversal of nested structures. Useful for GUI layout calculations where siblings matter more than depth.

**`.leaves()`**: Visit only nodes with no nested children. Useful for extracting terminal content.

**`.ancestors()`**: Given a position, iterate upward through containment. Requires index-based access, which interacts with the mutation/deletion design.

**Filtered traversal**: `.nested(only: text_node | paragraph)` — skip non-matching nodes during traversal rather than matching and ignoring. Optimisation opportunity, not semantic change.

These will be designed based on real use cases as the language evolves.

---
---

# Volume III: JSON Patterns

*This section describes idiomatic patterns for working with JSON data in ual. JSON is one of the most common external data formats, and how a language handles it reveals a lot about its data philosophy.*

---

## The Principle

**JSON is a wire format, not a runtime model.**

Most languages parse JSON into an in-memory object graph — nested maps, arrays, boxed values. The program then navigates this graph, casting values as it goes. Types are recovered at each access point, or not at all.

ual takes a different approach: **projection**. JSON is parsed into an opaque handle. You extract typed values out of it into typed stacks. After projection, everything is typed columns — the JSON is gone.

This is the same boundary described in Volume I, Part VII ("Arbitrary JSON/XML"): documents with unknown schemas remain opaque, and the extracted values are typed. The patterns in this volume show how that principle scales from simple extraction to complex ingestion scenarios.

> **Projection is the boundary: after projection, everything is typed stacks.**

---

## The Projection API

ual provides a query interface for extracting typed values from opaque JSON handles:

```
@doc = json.parse(input)

-- Typed extraction by path
@doc.string("name")          -- extract string at key "name"
@doc.i64("age")              -- extract i64 at key "age"
@doc.bool("active")          -- extract bool at key "active"
@doc.f64("score")            -- extract f64 at key "score"

-- Structure navigation
@doc.array()                  -- iterate array elements as opaque handles
@doc.object_keys()            -- iterate object keys as strings
@doc.at("address")            -- navigate to nested object, returns opaque handle

-- Presence checks
@doc.has("field_name")        -- does this key exist?
@doc.is_null("field_name")   -- is this key explicitly null?
```

Every extraction is typed at the call site. There is no generic "get value" that returns a dynamic type. If the JSON field doesn't match the expected type, the extraction fails — it doesn't silently coerce.

---

## Pattern 1: Simple Projection

The most basic pattern: parse, project, coordinate.

```
@doc = json.parse(input)

@names  = stack.new(string)
@ages   = stack.new(i64)
@active = stack.new(bool)

@names  < @doc.string("name")
@ages   < @doc.i64("age")
@active < @doc.bool("active")

relation person(@names, @ages, @active)

for person {|name, age, active|
    -- fully typed, fully coordinated
}
```

This follows the relation pattern from Volume I, Part II: the stacks exist independently, the relation coordinates them. The JSON document is just the source — once projected, it plays no further role.

---

## Pattern 2: Batch Ingestion

When JSON contains an array of objects, you columnise it into typed stacks with pre-allocated capacity.

```
@docs = json.parse(input).array()

@id      = stack.new(i64,    capacity: @docs.len())
@email   = stack.new(string, capacity: @docs.len())
@country = stack.new(string, capacity: @docs.len())

for @docs {|d|
    @id      < d.i64("id")
    @email   < d.string("email")
    @country < d.string("country", default: "unknown")
}

relation users(@id, @email, @country)
```

Pre-allocation with `capacity: @docs.len()` eliminates reallocation during ingestion — the same technique described in Volume I, Part V (Performance Characteristics). Each append becomes a simple write and length increment.

The `default:` parameter handles missing fields without separate null-checking logic. For more complex optionality, see Pattern 5.

---

## Pattern 3: Known Schema — Constitution Mapping

When you control the JSON schema (or know it in advance), you can map directly to constitutions. This is the "JSON/XML with known schema" case from the Volume I use case table.

### Direct Constitution Mapping

```
constitution user {
    id: i64
    email: string
    active: bool
}

@users = stack.new(user)

for json.parse(input).array() {|d|
    @users <- user(
        d.i64("id"),
        d.string("email"),
        d.bool("active")
    )
}

for @users {|id, email, active|
    -- process each user
}
```

The constitution's atomic append guarantee (Volume I, Part II) applies here: either all fields are extracted and pushed, or none are. A type mismatch on any field rejects the entire row.

### Tagged JSON — Union Mapping

When JSON uses a discriminant field (like `"type"`) to indicate variant, this maps naturally to ual unions:

```
constitution click  { x: i64, y: i64 }
constitution scroll { dy: i64 }
constitution hover  { x: i64, y: i64, target: string }

union event { click, scroll, hover }

@events = stack.new(event)

for json.parse(input).array() {|d|
    kind := d.string("type")
    match kind {
        "click"  => @events <- click(d.i64("x"), d.i64("y"))
        "scroll" => @events <- scroll(d.i64("dy"))
        "hover"  => @events <- hover(d.i64("x"), d.i64("y"), d.string("target"))
    }
}
```

The JSON `"type"` field provides the discriminant that ual stores separately in the union's discriminant array (Volume I, Part V). The mapping is explicit — no reflection, no automatic schema inference.

Unknown `kind` values fall through the match silently. To catch them, add a default arm that pushes to an error stack.

---

## Pattern 4: Unknown Keys — Dictionary Encoding

Not all JSON has a known schema. Arbitrary key-value objects — metadata, configuration, user-defined fields — have keys discovered at runtime.

This is the "Runtime-Determined Structure" boundary from Volume I, Part VII. Constitutions can't model it because the fields aren't known at compile time. But you can still store it in typed columns using dictionary encoding.

### The Approach

Intern keys into an integer namespace. Store typed values in separate lanes. Use a discriminant to indicate which lane holds the value for each entry.

```
@key_id  = stack.new(i64)
@kind    = stack.new(u8)       -- 0=string, 1=i64, 2=bool, 3=f64
@v_str   = stack.new(string)
@v_i64   = stack.new(i64)
@v_bool  = stack.new(bool)
@v_f64   = stack.new(f64)

relation entry(@key_id, @kind, @v_str, @v_i64, @v_bool, @v_f64)
```

For each JSON property:
- Intern the key string to a `key_id`
- Write to the appropriate lane
- Fill other lanes with defaults (empty string, 0, false, 0.0)
- Push atomically via the relation

```
-- key interning
keys := map.new(string, i64)
next_id := 0

for @doc.object_keys() {|key|
    if not keys.has(key) {
        keys[key] = next_id
        next_id = next_id + 1
    }

    kid := keys[key]

    -- dispatch by JSON value type
    if @doc.at(key).is_string() {
        entry <- kid, 0, @doc.string(key), 0, false, 0.0
    } else if @doc.at(key).is_i64() {
        entry <- kid, 1, "", @doc.i64(key), false, 0.0
    } else if @doc.at(key).is_bool() {
        entry <- kid, 2, "", 0, @doc.bool(key), 0.0
    } else if @doc.at(key).is_f64() {
        entry <- kid, 3, "", 0, false, @doc.f64(key)
    }
}
```

### Ergonomic Cost

This pattern is verbose. The default-filled lanes waste space. The dispatch logic is manual.

This is honest: unknown schemas are harder to work with in ual than in a dynamic language. The trade-off is that once ingested, the data is in dense typed storage and can be processed with all of ual's coordination tools — relations for coordination, hash perspectives or explicit maps for lookup (see Volume I, Part VII).

---

## Pattern 5: Optional Fields — Validity Columns

JSON is full of missing fields and `null` values. ual stacks don't have `nil` or `null` — every position holds a typed value (see Volume I, Part III on why confinement forbids sentinel values).

The idiomatic solution is a parallel validity column:

```
@age         = stack.new(i64)
@age_present = stack.new(bool)

for @docs {|d|
    if d.has("age") and not d.is_null("age") {
        @age         < d.i64("age")
        @age_present < true
    } else {
        @age         < 0          -- placeholder
        @age_present < false
    }
}

relation age_col(@age, @age_present)
```

This is columnar nullability — the same technique used in Apache Arrow and Parquet. The validity column is a dense boolean array with minimal overhead.

For constitutions with multiple optional fields, the pattern scales:

```
constitution patient {
    id: i64
    name: string
    age: i64
    blood_type: string
}

-- parallel validity
@age_present        = stack.new(bool)
@blood_type_present = stack.new(bool)
```

The validity stacks are coordinated with the constitution's field stacks via a relation, following the coordination pattern from Volume I, Part II.

### Future Direction

This pattern is functional but verbose. A future language feature — optional field annotations or nullable perspectives — could reduce the boilerplate while preserving the underlying dense storage model.

---

## Pattern 6: Validation as Flat Passes

Once JSON is projected into typed stacks, validation is ordinary ual iteration. You validate typed columns, not JSON trees.

```
constitution user {
    id: i64
    email: string
    country: string
}

@users = stack.new(user)
-- ... ingestion from Pattern 3 ...

@bad_ids     = stack.new(i64)
@bad_reasons = stack.new(string)

for @users {|id, email, country|
    if not is_valid_email(email) {
        @bad_ids     < id
        @bad_reasons < "invalid email"
    }
    if country == "" {
        @bad_ids     < id
        @bad_reasons < "missing country"
    }
}

relation errors(@bad_ids, @bad_reasons)
```

This is where the projection model pays off. Validation logic operates on typed values with full compiler support. Cross-field constraints work naturally through relations — the same coordination mechanism used throughout the language.

For nested JSON that has been ingested into constitutions with collections, Volume II's `.nested()` traversal applies: validate the entire tree with a single flat pass rather than manually recursing.

---

## Pattern 7: Change Tracking — Patch Events

Rather than mutating an in-memory JSON object graph, treat changes as typed events in a union. This fits the append-only direction discussed in Volume I, Part VII (Mutation and Deletion).

```
constitution set_str  { path: string, value: string }
constitution set_i64  { path: string, value: i64 }
constitution set_bool { path: string, value: bool }
constitution del      { path: string }

union patch_op { set_str, set_i64, set_bool, del }

collection @changelog of patch_op

-- record changes
@changelog <- set_str("/user/name", "Alice")
@changelog <- set_i64("/user/age", 31)
@changelog <- del("/user/temp_token")
```

The changelog is a compact, typed, columnar event log. You can:

- Replay it to reconstruct state
- Filter by path prefix
- Count changes by type
- Validate before applying

```
for @changelog {
    match set_str {|path, value|
        apply_string_change(path, value)
    }
    match set_i64 {|path, value|
        apply_int_change(path, value)
    }
    match del {|path|
        apply_deletion(path)
    }
}
```

This avoids the mutation problem entirely. The "current state" is the result of replaying the log — a functional approach that uses ual's `collection` and `union` machinery directly.

---

## What These Patterns Do Not Cover

### Deeply Nested Unknown JSON

If the JSON is arbitrarily nested with unknown structure at every level, projection becomes impractical — you'd need recursive dictionary encoding with parent references. At that point, you're building a document database inside ual.

**Recommendation:** Use an external query tool (jq, JSONPath) to flatten the data before projecting into ual. Let specialised tools handle structural navigation; let ual handle typed processing.

### JSON Schema Validation

Validating that JSON conforms to a JSON Schema (with `$ref`, `allOf`, `oneOf`, etc.) is a specialised task that doesn't map well to ual's columnar model. Use a schema validator as a preprocessing step; ingest only validated data.

### Round-Trip Fidelity

If you need to read JSON, modify one field, and write it back preserving all other fields (including unknown ones), the projection model loses information. You projected only the fields you knew about; the rest was discarded.

**Recommendation:** For round-trip scenarios, keep the opaque handle alongside your projected stacks. Modify through the handle's mutation API (if provided), not through ual stacks.

---

## Summary

| Scenario | Pattern | Key Technique |
|----------|---------|---------------|
| Single object, known fields | Simple Projection (1) | `relation` over extracted stacks |
| Array of objects | Batch Ingestion (2) | Pre-allocated stacks, `for` over array |
| Known schema | Constitution Mapping (3) | `constitution` + atomic append |
| Tagged variants | Union Mapping (3) | `union` + match on discriminant field |
| Unknown keys | Dictionary Encoding (4) | Key interning + typed value lanes |
| Optional fields | Validity Columns (5) | Parallel boolean stacks |
| Post-ingestion checks | Validation Passes (6) | `for` over typed stacks/relations |
| Tracking modifications | Patch Events (7) | `union` + `collection` as event log |

### The Principle, Restated

JSON arrives as a tree of dynamically typed values. ual doesn't try to represent that tree. Instead, it projects the parts it needs into typed columns and works with those.

The JSON tree is the source. The typed stacks are the truth.

This is the same insight that drives Volume I's core design: **heterogeneity is a boundary phenomenon**. At the boundary (parsing), you accept dynamic structure. Inside the boundary (processing), everything is typed, dense, and coordinated.

The boundary is projection. After projection, ual's four keywords — `relation`, `constitution`, `union`, `collection` — handle the rest.