# ual 1.3 Specification (Part 2)

## 5. Stacks as First-Class Objects (New in 1.3)

### 5.1 Stack Creation and Management

ual 1.3 introduces stacks as first-class objects that can be created, manipulated, and passed around like any other value:

```lua
-- Create a new stack
myStack = Stack.new()

-- Standard stack operations as methods
myStack.push(42)
value = myStack.pop()
myStack.dup()
myStack.swap()
```

### 5.2 Predefined Stacks

The system automatically initializes two standard stacks:

- `dstack` - The main data stack for general calculations
- `rstack` - Return stack for function calls and temporary storage

These stacks are immediately available without explicit creation, similar to how `stdout` and `stderr` are available in many languages.

### 5.3 Stack Methods

All stack objects provide the same comprehensive set of operations as methods:

```lua
-- Basic Operations
stack.push(value)    -- Push value onto stack
stack.pop()          -- Remove and return top value
stack.dup()          -- Duplicate top value
stack.swap()         -- Swap top two values
stack.over()         -- Copy second value to top
stack.rot()          -- Rotate top three values (a b c -> b c a)
stack.nrot()         -- Reverse rotate top three values (a b c -> c a b)
stack.nip()          -- Remove second value (a b -> b)
stack.tuck()         -- Duplicate top into third position (a b -> b a b)
stack.pick()         -- Copy nth item to top (0-based)
stack.roll()         -- Move nth item to top (0-based)

-- Pair Operations
stack.dup2()         -- Duplicate top pair (a b -> a b a b)
stack.drop2()        -- Remove top pair (a b c d -> a b)
stack.swap2()        -- Swap top two pairs (a b c d -> c d a b)
stack.over2()        -- Copy second pair to top (a b c d -> a b c d a b)

-- Inspection
stack.depth()        -- Return current stack depth
stack.len()          -- Alias for depth() for backward compatibility
stack.peek()         -- Return top value without removing
stack.peek(n)        -- Return nth value without removing (0-based)
```

### 5.4 Stack Transfer Operations

Values can be easily moved between stacks:

```lua
-- Transfer value from one stack to another
dstack.push(rstack.pop())

-- Store value temporarily on a different stack
rstack.push(dstack.pop())
-- Do some operations
dstack.push(rstack.pop())
```

### 5.5 Legacy Stack Operations and Compilation

For compatibility with Forth-style programming and previous versions of ual, the legacy stack operations are maintained as syntactic sugar that compile to stack object method calls:

```
push(x)    → dstack.push(x)
pop()      → dstack.pop()
dup()      → dstack.dup()
swap()     → dstack.swap()
add()      → dstack.add()
sub()      → dstack.sub()
mul()      → dstack.mul()
div()      → dstack.div()
store()    → dstack.store()
load()     → dstack.load()
drop       → dstack.drop()
over       → dstack.over()
rot        → dstack.rot()
nrot       → dstack.nrot()
nip        → dstack.nip()
tuck       → dstack.tuck()
roll       → dstack.roll()
pick       → dstack.pick()
dup2       → dstack.dup2()
drop2      → dstack.drop2()
swap2      → dstack.swap2()
over2      → dstack.over2()
depth      → dstack.depth()
len        → dstack.depth()
```

For return stack operations, the following equivalences apply:

```
pushr(x)   → rstack.push(x)
popr()     → dstack.push(rstack.pop())
peekr()    → rstack.peek()
```

At the compiler level, these operations are transformed directly into method calls on the appropriate stack object. This transformation occurs during the parsing phase, before code generation, ensuring no runtime overhead for using the more convenient syntax.

### 5.6 Memory Operations

The `store()` and `load()` operations, now implemented as stack methods, maintain their semantic meaning:

```lua
-- store() pops address and value from stack and stores value at address
stack.push(0x1000)  -- Address to store at
stack.push(42)      -- Value to store
stack.store()       -- Stores 42 at address 0x1000

-- load() pops address from stack and pushes value at that address
stack.push(0x1000)  -- Address to load from
stack.load()        -- Pushes value at address 0x1000 onto stack
```

At the compiler level, these operations map to appropriate memory access instructions in the target architecture, with TinyGo handling the necessary memory safety checks and platform-specific implementations.

---

## 6. Stacked Mode Syntax (New in 1.3)

### 6.1 Basic Stacked Mode

Stacked mode provides a concise syntax for stack operations on a per-line basis. Lines beginning with `>` use implicit stack operations:

```lua
> push:10 dup add push:5 swap sub
```

This is equivalent to:

```lua
dstack.push(10)
dstack.dup()
dstack.add()
dstack.push(5)
dstack.swap()
dstack.sub()
```

### 6.2 Stack Selection

By default, stacked mode operates on `dstack`. To operate on a different stack, use the `@` prefix:

```lua
@rstack > push:42 dup mul
```

This is equivalent to:

```lua
rstack.push(42)
rstack.dup()
rstack.mul()
```

Custom stacks can be used the same way:

```lua
myStack = Stack.new()
@myStack > push:10 push:20 add
```

### 6.3 Parameter Syntax

Stacked mode supports three parameter styles:

1. **No parameters**: Function name alone for operations taking no parameters
    
    ```lua
    > dup swap rot
    ```
    
2. **Literal values**: Using the `:` syntax for simple literals
    
    ```lua
    > push:10 push:0xFF push:true
    ```
    
3. **Expressions**: Using parentheses for expressions or variables
    
    ```lua
    > push(x) push(a+b) factorial(n)
    ```

### 6.4 Mixed Operations

Stacked mode can mix stack operations and regular function calls:

```lua
> push:10 dup add factorial(3) mul
```

This flexibility allows for clear expression of algorithms that combine stack manipulations with traditional function calls.

### 6.5 Compilation of Stacked Mode

Stacked mode is implemented as a syntactic transformation during the parsing phase. The compiler processes lines marked with `>` by:

1. Identifying the target stack (default `dstack` or explicitly named with `@stack`)
2. Converting each operation into a method call on that stack
3. Transforming parameter syntax appropriately:
   - `:literal` becomes direct values
   - `(expression)` is parsed as standard expressions
   - Parameterless operations get empty parentheses

For example, the line:
```lua
@myStack > push:10 dup factorial(n) add
```

Is transformed into:
```lua
myStack.push(10)
myStack.dup()
myStack.push(factorial(n))
myStack.add()
```

This transformation ensures that stacked mode incurs no runtime overhead while providing a more concise and readable syntax for stack-oriented programming.

---

## 7. Switch Statement (New in 1.3)

### 7.1 Syntax

The switch statement provides multi-way branching based on a value:

```lua
switch_case(expression)
  case value1:
    -- code for first case
  case value2, value3:  -- Multiple values
    -- code for second and third cases
  default:
    -- default case
end_switch
```

### 7.2 Semantics

- Cases are tested in order, and execution jumps to the first matching case.
- Multiple values can be specified in a single case, separated by commas.
- The `default` case executes when no other cases match.
- Like in Go, execution falls through from one case to the next unless explicitly terminated.
- Fall-through can be prevented by ending case blocks with return, break, or other control flow statements.

### 7.3 Compilation and Implementation Details

The switch statement is compiled to structured code that maps efficiently to TinyGo's switch statement, which in turn is optimized by LLVM. The implementation varies based on the switch expression type and case patterns:

#### 7.3.1 Integer Switch Statements

For switches on integer values, the compiler can generate highly optimized code:

1. **Dense Integer Ranges**: When case values form a dense range (e.g., consecutive numbers), TinyGo/LLVM will typically generate a jump table, providing O(1) case selection regardless of the number of cases.

2. **Sparse Integer Cases**: When case values are integers but spread apart, the compiler may use a combination of range checks and direct comparisons, optimizing for the specific distribution of values.

```lua
-- Dense integer range - compiled to jump table
switch_case(value)
  case 1: -- Action 1
  case 2: -- Action 2
  case 3: -- Action 3
  default: -- Default action
end_switch
```

#### 7.3.2 String and Complex Switch Statements

For string values or other complex types, the switch statement compiles to a series of comparisons:

```lua
-- String switch - compiled to sequential equality checks
switch_case(command)
  case "help", "?": -- Show help
  case "exit", "quit": -- Exit program
  default: -- Unknown command
end_switch
```

For string comparisons, the compiler generates appropriate string equality checks, taking advantage of any optimizations available in the TinyGo runtime.

#### 7.3.3 Optimizations

The following optimizations are applied to switch statements:

1. **Constant Folding**: Case expressions are evaluated at compile time when possible.
2. **Case Ordering**: When appropriate, cases may be reordered to improve branch prediction.
3. **Type-Specific Optimizations**: Different comparison strategies are used based on the type of the switch expression.
4. **Duplicate Elimination**: Multiple case values that execute the same code are efficiently combined.

#### 7.3.4 Implementation Examples

At the compiler level, a switch statement like:

```lua
switch_case(value)
  case 1, 2:
    doSomething()
  case 3:
    doSomethingElse()
  default:
    handleDefault()
end_switch
```

Is transformed into equivalent Go code:

```go
switch value {
case 1, 2:
    doSomething()
case 3:
    doSomethingElse()
default:
    handleDefault()
}
```

TinyGo and LLVM then optimize this code based on the specific pattern of cases and target architecture.

---

## 8. Scoping and Export Rules

### 8.1 Lexical Scope

- **Local:** Declared via `local x = ...`; visible until block/function ends.
- **Global (package-level):** Declared at top level.
    - **Uppercase** first letter → exported.
    - **Lowercase** first letter → private.

### 8.2 Packages and Visibility

- One file = one package.
- `import "pkg"` → gain access to **uppercase** symbols from that package.

---

## 9. Operational Semantics

### 9.1 Stack Statements

Standard stack operations work with stack objects (either implicit or explicit):

- `stack.push(expr)`: Evaluate expr, put result on stack.
- `stack.pop()`: Remove the top item from the stack.
- `stack.dup()`, `stack.swap()`: Manipulate stack elements.
- Arithmetic operations like `stack.add()`, `stack.sub()`, etc., operate on the stack.

Forth-inspired operations include:
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
- **depth/len:** Return the current number of items on the stack.

The convenience operations for return stack maintain their semantics but are implemented as syntax sugar:
- **pushr(expr):** Push a value from data stack to return stack
- **popr():** Pop a value from return stack to data stack
- **peekr():** Retrieve the top value from return stack without removing it

### 9.2 Control Flow

- `if_true(expr) ... end_if_true`: Execute the block if `expr` ≠ 0.
- `if_false(expr) ... end_if_false`: Execute the block if `expr` = 0.
- `while_true(expr) ... end_while_true`: Loop as long as `expr` ≠ 0.
- `for i = start, end, step do ... end`: Typical numeric loop.
- `for var in iterator do ... end`: Generic iterator-based loop.
- `switch_case(expr) ... end_switch`: Multi-way branching based on value.

### 9.3 Multiple Return

- `return a, b, c`: Return multiple values from a function.
- The caller can do `x, y = someFunc(...)`. Extra or missing values are truncated or set to 0/nil.

---

## 10. Packages System Semantics

1. **Single Package Declaration:** Each file starts with `package <name>`.
2. **Imports:** Use `import "otherPkg"` to gain access to uppercase symbols from other packages.
3. **Export vs. Private:** Identifiers starting with an uppercase letter are exported; lowercase ones are private to the package.

---

## 11. Example Programs

### 11.1 Basic Stack Object Usage

```lua
package main

import "fmt"

function main()
  -- Create custom stacks
  calcStack = Stack.new()
  tempStack = Stack.new()
  
  -- Use standard stacks
  dstack.push(10)
  dstack.push(20)
  dstack.add()
  
  -- Use a custom stack
  calcStack.push(5)
  calcStack.push(7)
  calcStack.mul()
  
  -- Transfer between stacks
  tempStack.push(dstack.pop())
  tempStack.push(calcStack.pop())
  
  -- Print results
  fmt.Printf("Values: %d, %d\n", tempStack.pop(), tempStack.pop())
  
  return 0
end
```

### 11.2 Stacked Mode Example

```lua
package main

import "fmt"

function factorial(n)
  > push(n) push:1 eq if_true
    > drop push:1
    return dstack.pop()
  > end_if_true
  
  > push(n) dup push:1 sub
  > factorial mul
  
  return dstack.pop()
end

function main()
  -- Calculate 5! using stacked mode
  > push:5 factorial
  fmt.Printf("5! = %d\n", dstack.pop())
  
  -- Calculate with mixed stacks
  @dstack > push:10 dup
  @rstack > dstack.pop() push:5 mul
  @dstack > push(rstack.pop()) add
  
  fmt.Printf("Result: %d\n", dstack.pop())
  
  return 0
end
```

### 11.3 Switch Statement Example

```lua
package main

import "fmt"

function processCommand(cmd)
  switch_case(cmd)
    case "help", "?":
      fmt.Printf("Available commands: help, status, exit\n")
    
    case "status":
      fmt.Printf("System status: OK\n")
    
    case "exit", "quit":
      fmt.Printf("Exiting...\n")
      return true
      
    default:
      fmt.Printf("Unknown command: %s\n", cmd)
  end_switch
  
  return false
end

function main()
  processCommand("help")
  processCommand("status")
  processCommand("exit")
  
  return 0
end
```

### 11.4 Combined Features Example

```lua
package main

import "fmt"
import "io"

function readFile(filename)
  result = {}
  
  file = io.open(filename, "r").consider {
    if_ok  result.file = _1
    if_err return { Err = "Could not open file: " .. _1 }
  }
  
  content = result.file.read("*all").consider {
    if_ok  result.Ok = _1
    if_err return { Err = "Read error: " .. _1 }
  }
  
  result.file.close()
  return result
end

function processData(data)
  lineStack = Stack.new()
  resultStack = Stack.new()
  
  for line in data:gmatch("[^\r\n]+") do
    @lineStack > push(line)
  end
  
  while_true(lineStack.depth() > 0)
    line = lineStack.pop()
    
    -- Check line type with switch
    switch_case(line:sub(1,1))
      case "#":
        -- Comment line, ignore
      case "+", "-", "*", "/":
        -- Process operator
        @resultStack > push(line:sub(1,1)) processOperator
      default:
        -- Assume number
        @resultStack > push(tonumber(line)) 
    end_switch
  end_while_true
  
  return resultStack.pop()
end

function processOperator(op)
  @resultStack > swap
  local b = resultStack.pop()
  local a = resultStack.pop()
  
  switch_case(op)
    case "+":
      @resultStack > push(a+b)
    case "-":
      @resultStack > push(a-b)
    case "*":
      @resultStack > push(a*b)
    case "/":
      @resultStack > push(a/b)
  end_switch
end

function main()
  result = readFile("input.txt").consider {
    if_ok  return processData(_1)
    if_err fmt.Printf("Error: %s\n", _1)
  }
  
  return 0
end
```

### 11.5 Using Forth-style Return Stack Operations

```lua
package main

import "fmt"

function demonstrateReturnStack()
  push(10)       -- Push 10 onto data stack
  push(20)       -- Push 20 onto data stack
  pushr(pop())   -- Move 20 to return stack 
  
  -- Do some operations on data stack
  push(5)
  mul()         -- 10 * 5 = 50
  
  -- Retrieve value from return stack
  push(peekr()) -- Copy return stack value (20) to data stack
  add()         -- 50 + 20 = 70
  
  -- This achieves the same result using object notation
  dstack.push(10)
  dstack.push(20)
  rstack.push(dstack.pop())
  dstack.push(5)
  dstack.mul()
  dstack.push(rstack.peek())
  dstack.add()
  
  -- The most concise form using stacked mode
  > push:10 push:20
  @rstack > dstack.pop()
  > push:5 mul push(rstack.peek()) add
  
  return pop()  -- Return 70
end

function main()
  result = demonstrateReturnStack()
  fmt.Printf("Result: %d\n", result) -- Prints "Result: 70"
  return 0
end
```

---

## 12. Implementation Guidelines

### 12.1 Parser

- Implement the EBNF rules, supporting binary/hex numeric literals (case-insensitive) and multiple comment styles.
- Add support for stacked mode syntax with stack selection.
- Implement switch statement parsing.
- Handle the transformation of legacy stack operations to method calls.

### 12.2 Symbol Resolution

- Maintain a map of packages to uppercase symbols (e.g., `pkg.FuncName`).
- Track stack objects and their methods.
- Handle the predefined stacks (`dstack`, `rstack`).

### 12.3 Code Generation

#### 12.3.1 Stack Operations

- Transform all stack operations (both direct and via stacked mode) into method calls on appropriate stack objects.
- Generate efficient code for stack method implementations, minimizing overhead.
- Implement memory operations (`store()`, `load()`) with appropriate TinyGo memory access patterns.

#### 12.3.2 Switch Statements

- Map switch statements to TinyGo switch constructs.
- Optimize case handling based on value type and distribution.
- Implement proper fallthrough semantics.

#### 12.3.3 Stack Objects

- Generate efficient implementations of stack objects.
- Optimize stack operations for the target platform.
- Initialize predefined stacks at program startup.

### 12.4 Operator Precedence

- Either treat all operators uniformly or define a precedence hierarchy similar to C.
- Ensure consistent behavior of binary operators across all expression contexts.

### 12.5 Optimization

- Inline short functions, eliminate redundant stack operations, etc.
- Optimize stack transfers and operations when the target platform allows.
- Apply platform-specific optimizations for various target architectures.

### 12.6 TinyGo Integration

- Ensure generated code aligns with TinyGo's expectations and optimizations.
- Leverage TinyGo's compilation pipeline for efficient target-specific code generation.
- Apply appropriate build tags and compiler directives for different target platforms.

---

## 13. Conclusion

ual 1.3 builds upon the solid foundation of ual 1.2 while introducing several powerful new features:

- **Stacks as First-Class Objects**: Treating stacks as regular objects with methods provides greater flexibility and cleaner semantics than specialized keywords. This allows for unlimited custom stacks, consistent interfaces, and more straightforward stack manipulation while maintaining backward compatibility through syntactic sugar for legacy operations like `pushr`, `popr`, and `peekr`.
    
- **Stacked Mode Syntax**: The concise, line-based syntax for stack operations makes stack-oriented code more readable while preserving ual's clean, structured nature. The `>` prefix and `@stack >` stack selection offer both convenience and clarity.
    
- **Switch Statement**: The addition of a switch_case construct provides a more elegant approach to multi-way conditionals, particularly valuable for state machines and command processing common in embedded systems.
    
- **Flexible Parameter Notation**: The colon syntax for literals in stacked mode (`push:10`) offers a clean way to distinguish literal values from expressions requiring evaluation (`push(x+y)`).

These innovations maintain ual's primary design goals:

1. **Embedded Focus**: All features are designed to compile efficiently to code suitable for resource-constrained environments, with careful implementation mappings to TinyGo's capabilities.
    
2. **Paradigm Balance**: ual 1.3 enhances its hybrid nature, allowing developers to blend stack-based and variable-based programming styles according to what best fits each task.
    
3. **Progressive Adoption**: Developers can gradually adopt stack-oriented features as they become comfortable, without being forced into an unfamiliar paradigm.
    
4. **Clear Semantics**: New features maintain ual's emphasis on readability and predictability, with explicit markers like `>` and `@` to clearly indicate special syntax.
    
5. **Compatibility**: Legacy operations remain supported through syntactic sugar, making the transition to the new object-oriented approach seamless.

The result is a programming language uniquely positioned between traditional imperative languages and stack-based concatenative languages, offering a bridge between these worlds that preserves the strengths of both approaches. ual 1.3 continues the language's evolution toward becoming an expressive yet efficient tool for both modern embedded systems and retro computing platforms.