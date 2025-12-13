# Changelog

All notable changes to ual will be documented in this file.

## [0.7.3] - 2025-12-12

### Fixed

**len() Operation**

The `len()` operation now correctly returns `int64` for use in expressions:

```ual
@data = stack.new(i64)
@data push:1 push:2 push:3

var count i64 = @data: len()    -- works: count = 3
if (@data: len() > 0) {         -- works in conditionals
    print("not empty")
}
```

Previously, `len()` returned Go's `int` type, causing type mismatches in UAL expressions.

**clear() Operation**

The `clear()` operation now works on all stacks, not just `@error` and `@spawn`:

```ual
@buffer = stack.new(i64)
@buffer push:1 push:2 push:3
@buffer clear                   -- now works: stack is empty
```

**ual run Output**

The `ual run` command no longer prints the version banner, providing clean program output:

```bash
$ ual run examples/001_fibonacci.ual
4181
```

The version banner still appears for `ual build` and other commands.

### Added

**Version Package**

Version is now managed centrally in `version/version.go`:

```go
import "github.com/ha1tch/ual/pkg/version"
fmt.Println(version.Version)  // "0.7.3"
```

**Version Synchronisation Script**

New `sync_version.sh` script keeps VERSION file and version package in sync:

```bash
./sync_version.sh              # Sync VERSION → version.go
./sync_version.sh 0.7.4        # Set new version in both files
./sync_version.sh --check      # Verify they match (for CI)
```

**Clean Script**

New `clean.sh` script removes generated files:

```bash
./clean.sh
# Removes:
#   examples/*.go (generated Go sources)
#   examples/<name> (compiled binaries)
#   ./ual, cmd/ual/ual (compiler binaries)
#   .DS_Store, *.tmp, __MACOSX/
```

**Design Document**

Added `DESIGN_v0.8.md` documenting planned features for v0.8:
- Files and sockets as stack sources/sinks
- Error stack architecture with forced handling
- `expect(n)` primitive for quorum/barrier patterns
- `@{a, b, c}` stack set syntax
- Hash literal syntax for parameters

### Changed

**Examples Renumbered**

All 71 examples now use consistent 3-digit prefixes (001-071), organised by feature:

- 001-023: Core language features
- 024-030: Consider construct
- 031-033: Select construct
- 034-051: Compute blocks
- 052-058: Stack operations (bring, freeze, hash, view, etc.)
- 059-071: Algorithms, benchmarks, and utilities

### Disabled

**walk() and filter() Operations**

These operations have been disabled pending design review. The implementations created temporary stacks but discarded results, making them ineffective. Use `reduce()` or explicit loops instead:

```ual
-- Instead of walk(), use explicit iteration:
var i i64 = 0
while (i < @data: len()) {
    var item i64 = @data: get(i)
    process(item)
    i = i + 1
}
```

### Removed

**attic/ Directory**

Historical specifications and deprecated code moved to separate archive. The main distribution now contains only current, production code.

---

### Runtime Unification (2025-12-13)

This update completes Phase 1 and Phase 2 of runtime unification, bringing the interpreter to full parity with the compiler.

#### Added

**Interpreter (iual)**

The UAL interpreter is now a first-class tool with full feature parity:

```bash
iual examples/001_fibonacci.ual       # Run with interpreter
iual --trace examples/022_pipeline.ual # Trace concurrent execution
iual -q examples/067_simple.ual       # Quiet mode
```

Features:
- Tree-walking interpreter at `cmd/iual/`
- True goroutine-based concurrency (matches compiler semantics)
- Shared runtime types with compiler via `pkg/runtime/`
- All 71 example programs pass in both compiler and interpreter
- `--trace` flag for execution tracing
- Quiet mode (`-q`) for scripting

**Runtime Package (pkg/runtime/)**

Unified runtime types shared between compiler and interpreter:

- `Value` — Type-safe value wrapper (int64, float64, string, bytes, bool)
- `ValueStack` — Stack of Values with type checking
- `ScopeStack` — Variable scope management for interpreter
- Existing types retained: `Stack`, `View`, `Bring`, `Walk`

**Package Restructure**

Shared packages extracted to `pkg/`:

```
pkg/
├── ast/         # Abstract syntax tree definitions
├── lexer/       # Lexical analysis
├── parser/      # Parser
├── runtime/     # Stack, Value, Views, Walk, Bring
└── version/     # Version management
```

#### Changed

**Interpreter Concurrency Model**

The interpreter now uses true goroutines for `@spawn pop play`:

```ual
@spawn < { producer() }
@spawn < { consumer() }
@spawn pop play    -- Now launches real goroutine
@spawn pop play    -- Now launches real goroutine
```

Previously, spawn tasks ran synchronously. Now they run concurrently, matching the compiler's `go _task()` semantics. This enables proper producer-consumer patterns and pipeline parallelism.

**Blocking Take**

The `take` operation now blocks correctly in the interpreter, waiting for concurrent producers:

```ual
@channel = stack.new(i64)
@channel perspective(FIFO)

@spawn < {
    @channel < 42    -- Producer pushes value
}
@spawn pop play

var value i64 = 0
@channel take:value  -- Blocks until value available
print(value)         -- 42
```

#### Fixed

- Interpreter `take` no longer uses artificial timeouts
- Variable scoping in interpreter matches compiler behaviour
- Stack perspective changes work correctly in interpreter
- Spawn task execution order matches compiler (LIFO)

#### Tests

- 71/71 example programs pass with compiler (`ual run`)
- 71/71 example programs pass with interpreter (`iual`)
- Runtime package has comprehensive unit tests
- Concurrency tests (pipeline, take_sync) work identically in both tools

## [0.7.2] - 2025-12-11

### Added

**Negative Literal Support**

Push negative values directly to stacks:

```ual
@nums = stack.new(i64)
@nums push:-42              -- negative integer

@floats = stack.new(f64)
@floats push:-3.14          -- negative float
```

Works in all contexts: stack operations, function arguments, and compute blocks.

**Compile-Time Type Checking for Push**

The compiler now validates type compatibility at compile time:

```ual
@integers = stack.new(i64)
@integers push:42           -- OK: integer to integer stack
@integers push:3.14         -- ERROR: cannot push float literal to i64 stack

@floats = stack.new(f64)
@floats push:3.14           -- OK: float to float stack
@floats push:42             -- OK: integer widened to float
```

**CLI Verbosity Controls**

New command-line options for controlling output verbosity:

```bash
ual -q run program.ual      # Quiet: only program output
ual run program.ual         # Normal: version header + completion messages
ual -v build program.ual    # Verbose: detailed progress
ual -vv run program.ual     # Debug: temp dirs, runtime paths, etc.
```

**Version Flag Variants**

All standard version flag styles now work:

```bash
ual version                 # Subcommand style
ual --version               # GNU style
ual -version                # Go style
```

**Makefile Build System**

New Makefile for streamlined builds:

```bash
make              # Build compiler
make test         # Run all tests (runtime + examples)
make test-runtime # Run Go unit tests only
make test-examples # Verify all .ual files compile
make bench        # Run benchmarks
make install      # Install to $GOPATH/bin
make clean        # Remove build artifacts
make check        # fmt + vet + test (CI-friendly)
```

### Changed

**CLI Output**

- Version header (`ual 0.7.2`) now displayed on stderr for all commands
- Quiet mode (`-q`) suppresses all non-error output
- Verbose mode (`-v`) shows compilation progress
- Debug mode (`-vv`) shows temp directories and runtime paths

**Build System**

- `build.sh --test` now correctly excludes `examples/` directory
- Examples are tested separately via compilation check

### Fixed

**TestIndexedStack**

Updated test to match current Indexed perspective semantics where parameterless `Pop()` removes the last element (array-like behaviour).

### Files Added

- `Makefile` — Comprehensive build system

### Files Modified

- `cmd/ual/main.go` — Version bump, verbosity controls, version flags
- `cmd/ual/parser.go` — Unary minus support in `parsePrimary()`
- `cmd/ual/codegen.go` — Type checking for push, `UnaryExpr` codegen
- `stack_test.go` — Updated TestIndexedStack
- `build.sh` — Fixed test command to exclude examples/
- `VERSION` — Updated to 0.7.2

## [0.7.1] - 2025-12-10

### Added

**Cross-Language Benchmark Suite**

Comprehensive benchmarks comparing ual against C, Go, and Python for compute-intensive algorithms:

| Algorithm | C (-O2) | Go | ual | Python |
|-----------|---------|-----|-----|--------|
| Mandelbrot (1000 iter) | 4,078 ns | 4,161 ns | 4,170 ns | 116,461 ns |
| Integrate (1000 steps) | 1,565 ns | 1,206 ns | 1,598 ns | 59,590 ns |
| Leibniz (100k terms) | 127 μs | 120 μs | 120 μs | 6,773 μs |
| Newton (20 iter) | 53 ns | 7.6 ns | 10.2 ns | 938 ns |
| Array Sum (50 elem) | 36 ns | 34.7 ns | 34.7 ns | 2,618 ns |
| DP Fibonacci (n=40) | 17 ns | 57.7 ns | 61.2 ns | 2,121 ns |
| Math Functions | 17 ns | 29.4 ns | 29.0 ns | 195 ns |

**Key findings:**
- ual matches Go performance (compiles to Go)
- Go matches C for most numeric workloads
- Python is 30-100x slower than ual/Go/C

**Benchmark categories:**
- `compute_bench_test.go` — Pure computation (Go vs ual patterns)
- `pipeline_bench_test.go` — Full ual pipeline with stack overhead
- `c/c_bench.c` — C reference implementation (gcc -O2)
- `python/python_bench.py` — Python reference implementation (CPython)

**Runner script:** `./bench.sh [compute|pipeline|overhead|c|python|all]`

### Changed

**Benchmark reorganisation:**
- Moved C benchmarks to `benchmarks/c/` subdirectory
- Moved Python benchmarks to `benchmarks/python/` subdirectory
- Updated `RESULTS.md` with cross-language analysis and performance positioning

### Files Added

- `benchmarks/c/c_bench.c` — C reference benchmarks
- `benchmarks/python/python_bench.py` — Python reference benchmarks
- `benchmarks/RESULTS.md` — Comprehensive analysis with guidance

## [0.7.0] - 2025-12-10

### Added

**Compute Construct for Zero-Copy Native Math**

The `.compute()` construct creates an "optimization island" where arithmetic operations run on native CPU types, avoiding the serialization overhead ("Byte Tax") of ual's `[]byte` storage:

```ual
-- LIFO stack example (pop bindings)
@physics = stack.new(f64)
@physics push(10.0)
@physics push(20.0)

@physics {
}.compute(
    {|a, b|
        var result = a * b
        return result
    }
)
-- Stack now contains 200.0
```

**Hash Stack with `self.property` Access:**

```ual
-- Hash perspective for named state
@body = stack.new(f64, Hash)
@body set("mass", 10.0)
@body set("velocity", 5.0)

@body {
}.compute(
    {||
        var m = self.mass
        var v = self.velocity
        var ke = 0.5 * m * v * v
        return ke
    }
)
-- Result stored at __result_0__, access via self.__result_0__
```

**Indexed Stack with `self[i]` Access:**

```ual
-- Indexed perspective for array-style access
@vec = stack.new(f64, Indexed)
@vec push(1.0)
@vec push(2.0)
@vec push(3.0)

@vec {
}.compute(
    {||
        var sum = 0.0
        var i = 0
        while i < 3 {
            var val = self[i]
            sum = sum + (val * val)
            i = i + 1
        }
        return sum
    }
)
-- Result: 1 + 4 + 9 = 14
```

**Features:**

1. **Type Rigidity** — The compute block assumes the type of the attached stack (i64, f64, etc.). All variables inside are native Go types.

2. **Bindings** — `{|a, b|}` pops values from the stack into native variables (LIFO order: a = top, b = second). For Hash stacks, use `{||}` (empty bindings) and access via `self.property`.

3. **`set` Operation** — `@stack set("key", value)` stores named values in Hash perspective stacks.

4. **`get` Operation** — `@stack get("key")` reads values from Hash perspective stacks and pushes to dstack.

5. **`self.property` Access** — Inside compute blocks, `self.property` reads named values from Hash perspective stacks without additional serialization.

6. **`self[i]` Access** — Inside compute blocks, `self[i]` reads values by index from Indexed perspective stacks.

7. **Zero-Copy Interior** — Operations within the block use native CPU instructions. Only entry (pop) and exit (push) pay the serialization cost.

8. **Infix Syntax** — Standard mathematical notation inside compute: `var force = mass * acceleration`

9. **Negative Literals** — Unary minus support: `var a = -5.0` and `var b = -x`

10. **Multiple Returns** — `return a, b` pushes values in left-to-right order (b ends up on top). For Hash stacks, returns use `__result_N__` keys.

11. **Void Compute** — Omitting `return` consumes the bindings without pushing results (consumer pattern)

12. **Control Flow** — `if`, `while`, `break`, `continue`, and variable assignments supported inside compute blocks

13. **Empty Bindings** — `{||}` syntax for compute/closures that take no parameters

14. **Math Functions** — Common math functions auto-prefix with `math.`: `sqrt(x)`, `abs(x)`, `sin(x)`, `cos(x)`, `pow(x, y)`, `log(x)`, `exp(x)`, `floor(x)`, `ceil(x)`, `round(x)`, `min(x, y)`, `max(x, y)`

15. **Local Arrays** — Stack-allocated fixed-size arrays for algorithms: `var buf[1024]` declares a local array, `buf[i]` reads, `buf[i] = expr` writes. Enables BFS queues, DP tables, and other medium-sized algorithms.

```ual
@result {}.compute({|n|
    var dp[100]
    dp[0] = 0
    dp[1] = 1
    var i = 2
    while i <= n {
        dp[i] = dp[i - 1] + dp[i - 2]
        i = i + 1
    }
    return dp[n]
})
```

16. **Container Array Views** — Zero-copy read/write access to container properties as arrays: `self.prop[i]` reads and `self.prop[i] = expr` writes. The compiler generates `unsafe.Slice` mappings for direct byte-level access. Requires the property to contain pre-allocated byte storage of appropriate size.

**Perspective-Specific Semantics:**

| Perspective | Bindings | Returns | `self` Access |
|-------------|----------|---------|---------------|
| LIFO/FIFO   | Pop from stack | Push to stack | Not available |
| Hash        | Not allowed (use `self`) | Uses `__result_N__` keys | `self.property` |
| Indexed     | Pop from stack | Push to stack | `self[i]` |

**Design Decisions (Hash Perspective):**

The Hash perspective treats containers as **records/objects** rather than sequences:

1. **No pop-bindings** — Hash stacks have no "top" to pop from. All input comes via `self.property`. Using `{|a, b|}` bindings on a Hash stack is a compile-time error.

2. **Return writes to reserved keys** — Since Hash stacks are key-value stores, `return expr` writes to `__result_0__`, `return a, b` writes to `__result_0__` and `__result_1__`, etc. Access results via `self.__result_0__` in subsequent compute blocks or via `get("__result_0__")`.

3. **Record semantics** — This design keeps Hash containers conceptually clean: they are always addressed by key, never by position. The alternative (mixing LIFO push/pop with key-value access) was rejected as confusing.

**Implementation:**
- New tokens: `TokCompute`, `TokSelf`, `TokSet`, `TokGet`
- New AST nodes: `ComputeStmt`, `MemberExpr`, `IndexExpr`, `UnaryExpr`
- Runtime: `Stack.PopRaw()`, `Stack.PushRaw()`, `Stack.GetRaw()`, `Stack.SetRaw()`, `Stack.GetAtRaw()`, `Stack.Lock()`, `Stack.Unlock()`
- Parser: Empty bindings support (`{||}` as `TokBarBar`), unary minus, `break`/`continue` in compute
- Codegen: Perspective-aware return handling, type-aware `print`/`dot` operations

## [0.6.0] - 2025-12-10

### Added

**Select Construct for Concurrent Stack Waiting**

The `.select()` construct provides Go-like select semantics for waiting on multiple stacks:

```ual
@inbox {
    -- setup block
}.select(
    @inbox {|msg| handle(msg)}
    @commands {|cmd| run(cmd)}
    _: { default_handler() }
)
```

**Features:**

1. **Multi-stack waiting** — Wait on multiple stacks concurrently, first with data wins

2. **Per-case timeouts** — Each case can have its own timeout with handler:
   ```ual
   @inbox {|msg|
       process(msg)
       timeout(100, {||
           log("timeout, retrying")
           retry()
       })
   }
   ```

3. **Retry and restart control flow**:
   - `retry()` — Restart waiting on the current case
   - `restart()` — Restart the entire select

4. **Default case** — `_: { handler }` for non-blocking select (like Go)

5. **Setup block** — Code block before `.select()` runs before waiting begins

6. **Default stack scoping** — Cases without explicit `@stack` use the setup block's stack

**Semantics:**
- With default case: Non-blocking, checks each stack sequentially
- Without default case: Blocking, waits until any case has data
- Timeout handlers run when a case's wait times out
- `retry()` in timeout handler resets that case's timer
- Select completes after any handler runs (unless retry/restart)

**New tokens:** `TokSelect`, `TokTimeout`, `TokRetry`, `TokRestart`

**New AST nodes:**
- `SelectStmt` (Block, DefaultStack, Cases)
- `SelectCase` (Stack, Bindings, Handler, TimeoutMs, TimeoutFn)

**Runtime additions:**
- `TakeWithContext(ctx, timeoutMs)` — Context-aware blocking take

**Example files:**
- `40_select_basic.ual` — Basic select with default
- `41_select_blocking.ual` — Blocking select without default
- `42_select_timeout.ual` — Per-case timeout with retry

## [0.5.0] - 2025-12-10

### Added

**Consider Construct for Structured Error Handling**

The `.consider()` construct provides pattern matching on block execution outcomes:

```ual
@dstack {
    push:42
    status:ok
}.consider(
    ok: dot,
    error |e|: handle_error(e),
    _: default_handler()
)
```

**Features:**

1. **Status matching** — Cases match on status labels (ok, error, custom labels)

2. **Explicit status setting** — `status:label` or `status:label(value)` sets status anywhere:
   ```ual
   func divide(a i64, b i64) i64 {
       if (b == 0) {
           status:error
           return 0
       }
       status:ok
       return a / b
   }
   ```

3. **Value bindings** — Extract status data into variables:
   ```ual
   .consider(
       result |val|: push:val,
       error |code|: log_error(code)
   )
   ```

4. **Default case** — `_:` matches any unhandled status

5. **Global status propagation** — Functions can set status for calling consider blocks

6. **Nested considers** — Status is properly saved/restored

7. **Implicit error detection** — If @error stack has content and status is still "ok", automatically switches to "error"

**New tokens:** `TokConsider`, `TokStatus`

**New AST nodes:**
- `ConsiderStmt` (Block, Cases)
- `ConsiderCase` (Label, Bindings, Handler)
- `StatusStmt` (Label, Value)

**Parser additions:**
- `parseConsider()` - Parse consider block after stack block
- `parseConsiderCase()` - Parse individual case with optional bindings
- `parseStatusStmt()` - Parse status:label statements
- Extended `parseStackBlock()` to handle status: and other statements

**Codegen additions:**
- Global `_consider_status` and `_consider_value` variables
- `generateConsiderStmt()` - Generate switch with save/restore semantics
- `generateStatusStmt()` - Set global status
- Type assertion for value bindings

**Example files:**
- `31_consider_simple.ual` - Basic consider usage
- `32_consider_status.ual` - Explicit status setting
- `33_consider_comprehensive.ual` - Multiple case patterns
- `34_consider_bindings.ual` - Value binding extraction
- `35_consider_functions.ual` - Functions setting status
- `36_consider_docs.ual` - Documentation example

### Design Philosophy

The consider construct implements "Safety by Structure" rather than "Safety by Discipline":

- If you don't handle a status, the program panics (mandatory handling)
- Error handling is co-located with the code that might fail
- Clear visual separation between happy path and error handling
- No forgotten error checks possible

This addresses the "Go Marshalling Horror" where error checking on every line obscures business logic.

## [0.0.9] - 2025-12-09

### Added

**Phase 5: Functions**

Function declarations:
```
func greet() {
    push:42 dot
}

func showSum(a i64, b i64) {
    push:a push:b add dot
}

func sum(a i64, b i64) i64 {
    push:a push:b add let:result
    var result i64 = 0
    return result
}
```

Function calls as statements:
```
greet()
showSum(10, 20)
```

Function calls in expressions:
```
var s i64 = sum(100, 200)
```

Colon shorthand for single-argument functions:
```
var x i64 = double:5       -- same as double(5)
var y i64 = double:double:3 -- chaining works
show:42                    -- as statement
```

Error-capable function syntax (parsed, codegen pending):
```
@error < func risky() i64 {
    -- can push to @error
    return 0
}
```

New tokens: TokFunc, TokReturn

New AST nodes:
- FuncDecl (Name, Params, ReturnType, CanFail, Body)
- FuncParam (Name, Type)
- FuncCall (Name, Args)
- ReturnStmt (Value)

Parser additions:
- parseFuncDecl, parseReturnStmt
- Function call parsing in parseIdentStmt
- Function call parsing in expression primary
- Colon shorthand: `name:arg` for single-argument functions
- Lookahead to distinguish view ops from function shorthand

Codegen:
- Global stack declarations (moved out of main)
- generateFuncDecl, generateFuncCall, generateReturnStmt
- generateExprValue for expressions in function contexts
- goTypeFor type mapping

**Note:** Function names cannot be reserved operation keywords (add, sub, mul, etc.)

## [0.0.8] - 2025-12-09

### Added

**Phase 4: For Iteration**

Basic for loop over stacks:
```
@numbers for{ dot }              -- each element to @dstack, print
```

Named parameter:
```
@numbers for{|v|
    push:v push:v mul dot        -- squares
}
```

Index and value (with perspective):
```
@numbers.fifo for{|i, v|
    push:i dot                   -- index
    push:v dot                   -- value
}
```

Perspective modifiers:
- `.lifo` - iterate top to bottom (default)
- `.fifo` - iterate bottom to top
- `.indexed` - iterate with index access

New tokens: TokFor, TokPipe

New AST node: ForStmt (Stack, Perspective, Params, Body)

Parser additions: parseForStmt with |params| handling

Codegen: generateForStmt with snapshot semantics

**Algorithm Examples (algorithms.ual)**

Demonstrates 12 fundamental algorithms:
- Fibonacci sequence generation
- Factorial computation
- Sum, Min, Max reductions
- Linear search with position
- Stack reversal
- Count with condition
- Filter elements
- Map/transform elements
- GCD (Euclidean algorithm)
- Prime number check
- Power/exponentiation

## [0.0.7] - 2025-12-09

### Added

**Phase 3: Control Flow**

Conditional statements:
```
if (x > 5) {
    push:100 dot
}

if (x > 10) {
    push:1 dot
} elseif (x > 5) {
    push:2 dot
} else {
    push:3 dot
}
```

While loops:
```
var count i64 = 5
while (count > 0) {
    push:count dot
    push:count dec let:count
}
```

Loop control:
```
while (x > 0) {
    if (x == 5) { break }
    if (x == 3) { continue }
    push:x dot
    push:x dec let:x
}
```

Comparison operators in conditions:
- `>` greater than
- `<` less than
- `>=` greater or equal
- `<=` less or equal
- `==` equal
- `!=` not equal

New tokens:
- TokIf, TokElseIf, TokElse
- TokWhile, TokBreak, TokContinue  
- TokSymGt, TokSymLt, TokSymGe, TokSymLe, TokSymEq, TokSymNe

New AST nodes:
- IfStmt (with ElseIf branches)
- WhileStmt
- BreakStmt, ContinueStmt
- BinaryExpr (for conditions)

Parser additions:
- parseIfStmt, parseWhileStmt
- parseCondition, parseBlock

Codegen additions:
- generateIfStmt, generateWhileStmt
- generateCondition, generateCondExpr

## [0.0.6] - 2025-12-09

### Added

**Phase 2: Type Stacks & Variables**

Symbol table for variable tracking:
- Scoped symbol lookup
- Type-indexed variable storage
- Compile-time index optimization

Variable declarations (Go-style):
```
var x i64 = 42              -- @i64[0] = 42
var y f64 = 3.14            -- @f64[0] = 3.14
var a, b i64 = 10, 20       -- multiple declarations
var n = 99                  -- type inference (i64)
```

Dynamic assignment from @dstack:
```
push:5 push:7 mul let:z     -- result -> @i64["z"]
```

Variable borrowing to @dstack:
```
push:x                      -- copies @i64[x] to @dstack
push:x push:y add           -- arithmetic with variables
```

Type stacks (Hash perspective for named slots):
- @i64, @u64 (integer)
- @f64 (float)
- @string
- @bytes

### Fixed

- Element struct casing bug in stack.go (Element vs element)
- Added TypeUint64 constant
- Unused stack variable warnings in generated code

## [0.0.5] - 2025-12-09

### Added

**Phase 1: Extended Operators**

Unary arithmetic:
```
neg         -- negate
abs         -- absolute value
inc         -- increment by 1
dec         -- decrement by 1
```

Min/Max:
```
min         -- smaller of two
max         -- larger of two
```

Bitwise operations:
```
band        -- bitwise and
bor         -- bitwise or
bxor        -- bitwise xor
bnot        -- bitwise not (unary)
shl         -- shift left
shr         -- shift right
```

Comparison operations (push bool to @bool):
```
eq          -- equal
ne          -- not equal
lt          -- less than
gt          -- greater than
le          -- less or equal
ge          -- greater or equal
```

**@bool stack:**

New default stack for boolean/comparison results:
```
push:5 push:3 gt    -- pops from @dstack, pushes true to @bool
push:5 push:5 eq    -- pops from @dstack, pushes true to @bool
```

### Example

```
-- Unary
push:10 neg dot           -- -10
push:5 neg abs dot        -- 5

-- Min/Max
push:3 push:7 min dot     -- 3
push:3 push:7 max dot     -- 7

-- Bitwise
push:5 push:3 band dot    -- 1
push:1 push:3 shl dot     -- 8

-- Comparison
push:5 push:3 gt          -- true → @bool
```

---

## [0.0.4] - 2025-12-09

### Added

**Default stacks (Forth-style):**

- `@dstack` - data stack (i64), implicit target for operations
- `@rstack` - return stack (i64), for temporary storage
- `@error` - error stack (bytes), for error handling
- `--no-forth` flag to disable default stacks

**Implicit stack operations:**

Operations without explicit stack reference use @dstack:
```
push:10 push:20 add print    -- uses @dstack implicitly
```

**I/O operations:**

- `print` - peek and print top of stack (non-destructive)
- `dot` - pop and print top of stack (destructive, Forth-style)

**Return stack operations:**

- `tor` - move from data stack to return stack (>r in Forth)
- `fromr` - move from return stack to data stack (r> in Forth)

### Example

```
-- Implicit @dstack usage
push:10
push:20
add
print       -- prints 30 (peek, non-destructive)

-- Return stack
push:42
tor         -- move to @rstack
push:99
fromr       -- move back from @rstack
dot         -- prints 42 (pop)
dot         -- prints 99 (pop)
```

---

## [0.0.3] - 2025-12-09

### Added

**Syntax improvements for less arid code:**

- **Lua-style comments**: `-- this is a comment`
- **Block comments**: `/* multi-line comment */`
- **Colon sugar for single arguments**: `push:42` instead of `push(42)`
- **Optional colon after stack selector**: `@stack push:10` works same as `@stack: push(10)`
- **Multiple operations per line** (Forth-style): `@calc push:3 push:4 add`
- **Block syntax for stack operations**:
  ```
  @numbers {
      push:10
      push:20
      add
  }
  ```

**Forth-like stack operations:**

- `add` - pop two, push sum
- `sub` - pop two, push difference
- `mul` - pop two, push product
- `div` - pop two, push quotient
- `mod` - pop two, push remainder
- `dup` - duplicate top element
- `drop` - discard top element
- `swap` - swap top two elements
- `over` - copy second element to top
- `rot` - rotate top three elements

### Changed

- Parser refactored to support multiple syntax forms
- Codegen handles StackBlock nodes for grouped operations
- All changes are backwards compatible with v0.0.2 syntax

### Examples

New `forth.ual` example demonstrating Forth-style programming:
```
@calc = stack.new(i64)
@calc push:3 push:4 add push:10 push:2 sub mul
result = @calc: pop()  -- result = 56
```

---

## [0.0.2] - 2025-12-09

### Added

**Decoupled perspectives (Views):**

- `View` type with independent cursor and perspective
- Multiple views can attach to same stack
- Each view maintains its own position
- Operations: `Attach`, `Detach`, `Peek`, `Pop`, `Walk`, `Advance`, `SetCursor`

**Work-stealing implementation:**

- Traditional Chase-Lev deque for comparison
- ual-based work-stealing (unlimited and capped variants)
- Comprehensive benchmarks showing ~10-20% overhead under contention

**Capped stacks:**

- `NewCappedStack(perspective, type, capacity)` pre-allocates memory
- Zero allocations in hot path
- `IsFull()` check, push fails when full

**Phase 1 compiler (ual):**

- Lexer (440 lines)
- Parser with AST (698 lines)
- Go code generator (420 lines)
- CLI: `ual compile`, `ual tokens`, `ual ast`

### Performance

Work-stealing comparison (1 owner + 3 thieves):
- Traditional: 374 ns
- ual Capped: 425 ns (1.14x)

### Files Added

- `view.go` - View implementation
- `view_test.go` - View tests
- `worksteal.go` - Work-stealing implementations
- `worksteal_test.go` - Work-stealing tests and benchmarks
- `cmd/ual/` - Compiler source

---

## [0.0.1] - 2025-12-09

### Added

**Core stack implementation:**

- Stack with four perspectives: LIFO, FIFO, Indexed, Hash
- Element types: Int64, Float64, String, Bytes, Bool
- O(1) operations for all perspectives
- Thread-safe with RWMutex

**Bring operation:**

- Atomic transfer between stacks
- Type conversion during transfer
- Fails atomically (source unchanged on error)

**Walk operations:**

- `Walk` - traverse and transform
- `Filter` - select elements matching predicate
- `Reduce` - fold to single value
- `Map` - transform to new stack

**Frozen stacks:**

- `Freeze()` makes stack immutable
- Push/Pop error, Peek/Walk still work

### Performance

| Operation | Time |
|-----------|------|
| Pop LIFO | 35 ns |
| Pop FIFO | 34 ns |
| Pop Hash | 269 ns |
| Peek LIFO | 20 ns |
| Bring | 300 ns |

### Files

- `stack.go` - Core stack implementation
- `bring.go` - Bring operation and type conversion
- `walk.go` - Walk, Filter, Reduce, Map
- `stack_test.go` - Tests
- `bench_test.go` - Benchmarks

---

## Version History Summary

| Version | Date | Highlights |
|---------|------|------------|
| 0.7.2 | 2025-12-11 | Negative literals, compile-time type checking, CLI verbosity controls, Makefile |
| 0.7.1 | 2025-12-10 | Cross-language benchmarks (C, Go, ual, Python), benchmark suite reorganisation |
| 0.7.0 | 2025-12-10 | Compute construct (.compute), self.property, self[i], set/get, local arrays, container array views |
| 0.6.0 | 2025-12-10 | Select construct (.select), multi-stack waiting, timeouts |
| 0.5.0 | 2025-12-10 | Consider construct (.consider), status: statements, structured error handling |
| 0.0.9 | 2025-12-09 | Functions (func, return), colon shorthand |
| 0.0.8 | 2025-12-09 | For iteration with perspectives, algorithm examples |
| 0.0.7 | 2025-12-09 | Control flow (if/elseif/else, while, break, continue) |
| 0.0.6 | 2025-12-09 | Type stacks, variables (var, let), symbol table |
| 0.0.5 | 2025-12-09 | Extended operators (unary, bitwise, comparison), @bool stack |
| 0.0.4 | 2025-12-09 | Default stacks (@dstack, @rstack, @error), print/dot, tor/fromr |
| 0.0.3 | 2025-12-09 | Forth-style syntax, block operations, Lua comments |
| 0.0.2 | 2025-12-09 | Decoupled views, work-stealing, compiler |
| 0.0.1 | 2025-12-09 | Core stack, perspectives, bring, walk |
