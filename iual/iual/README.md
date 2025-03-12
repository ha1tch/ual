# iual v0.0.1

*iual is an exceedingly trivial interactive ual 0.0.1 interpreter*

## Overview

iual is a tiny interactive interpreter implementing a small subset of the ual language that combines Forth‑like stack manipulation with a container‑centric design. In this version, two default integer stacks are provided—the data stack (`@dstack`) and the return stack (`@rstack`)—each of which can be configured to operate in different perspectives (LIFO or FIFO). Additionally, iual supports a dedicated spawn stack (`@spawn`) for managing concurrent tasks. The interpreter accepts commands using two primary syntactic styles (colon‑based and function‑like), and supports advanced Forth operations for direct stack manipulation, as well as cross‑stack transfers via the `bring` operation.

## Key Features

- **Multiple Stacks:**  
  - Two default integer stacks are created:
    - **@dstack:** The main data stack for arithmetic and general operations.
    - **@rstack:** The return stack, typically used for subroutine returns or temporary storage.
  - Additional stacks can be created using the `new` command.

- **Stack Perspectives:**  
  Each stack can be set to operate in one of two modes:
  - **LIFO (Last-In‑First‑Out):**  
    - New values are appended at the top.
    - When printed, the top element appears within square brackets at the end.
  - **FIFO (First-In‑First‑Out):**  
    - New values are inserted at the front (index 0).
    - When printed, the front element is shown within square brackets at the beginning.
  - **Flip:**  
    - The `flip` command toggles the current perspective between LIFO and FIFO without physically reordering the underlying array.
    - The new perspective is reported in uppercase (e.g., "LIFO" or "FIFO").

- **Advanced Forth‑like Operations:**  
  In addition to basic push and pop, iual supports several advanced commands for stack manipulation:
  - **Tuck:**  
    - Duplicates the top element and inserts it below the top.
    - For example, given `a b`, executing `tuck` produces `b a b`.
  - **Pick:**  
    - Copies the nth element (counting from the top, starting at 0) to the top without removing it.
  - **Roll:**  
    - Removes the nth element and pushes it onto the top.
  - **Over:**  
    - Copies the second element to the top.
  - **Dup2, Drop2, Swap2, Over2:**  
    - These commands operate on the top two elements (or pairs), allowing you to duplicate, remove, swap, or copy them as a group.

- **Bring Operation:**  
  The `bring` command supports cross‑stack transfers:
  - **Usage:**  
    - `bring(int, <stackname>)`
    - It pops a value from the specified integer stack (the source) and pushes it onto the currently selected stack (the target).
  - The stack name may be given with or without the `@` prefix.
  - **Purpose:**  
    - This operation allows data to be moved directly between stacks without using intermediary variables.

- **@spawn Stack:**  
  The spawn stack is dedicated to managing concurrent tasks (goroutines):
  - Elements on the @spawn stack represent live execution contexts.
  - Supported operations include commands like `list`, `add`, `pause`, `resume`, and `stop` to control these tasks.

## Syntax

iual supports two primary syntactic styles, which are functionally equivalent but offer different levels of brevity and clarity.

### Colon‑Based Syntax

- **Format:**  
  `operation:parameter`
  
- **Description:**  
  The colon-based syntax is concise and Forth‑inspired. An operation name is immediately followed by a colon and its parameter.  
  - **Example:**  
    - `push:10` pushes the value `10` onto the current stack.
    - A compound command like `@dstack: push:1 push:2 push:3 print` applies all listed operations on the selected stack.
  
- **Advantages:**  
  - **Brevity:** Eliminates extra punctuation.
  - **Direct Binding:** The parameter is visibly attached to its operation.
  - **Ideal for Simple Commands:** Best used for straightforward operations or chaining multiple commands on one line.

### Function‑Like Syntax

- **Format:**  
  `operation(parameter)`  
  or for multiple parameters: `operation(param1, param2, ...)`
  
- **Description:**  
  The function‑like syntax uses parentheses to enclose parameters, similar to many high‑level languages. This style is particularly useful when you have nested operations or multiple parameters.  
  - **Example:**  
    - `push(10)` functions identically to `push:10`.
    - Nested calls, such as `div(10, push(2))`, allow you to perform an operation on the result of another.
    - When an operation takes multiple parameters, commas separate them. For instance:  
      `bring(str, str.split(@string_stack: pop, ","))`
  
- **Advantages:**  
  - **Familiarity:** More in line with standard high‑level language function calls.
  - **Clarity in Nesting:** Parentheses and commas clearly delineate the expression structure.
  - **Flexibility:** Facilitates complex and nested expressions.

### Mixing Syntaxes

Both syntaxes are equivalent at the semantic level, so you can use them interchangeably depending on the situation:
- **Simple, direct commands:** Use colon‑based syntax.
- **Complex or nested operations:** Use function‑like syntax.
- You can even mix both in the same compound command.

## Example Interactive Session

```plaintext
$ ./iual
iual v0.0.1
iual is an exceedingly trivial interactive ual 0.0.1 interpreter
Added spawn 'spawn'
> @dstack: push:1 push:2 push:3 push:4 push:5 push:6
> @dstack: print
@dstack: 1 2 3 4 5 [ 6 ]
> @dstack: fifo print
@dstack perspective set: FIFO
@dstack: [ 1 ] 2 3 4 5 6 
> @dstack: push:0 print
@dstack: [ 0 ] 1 2 3 4 5 6 
> @dstack: pop print
@dstack: [ 1 ] 2 3 4 5 6 
> @rstack: push(100) push(101) push:102 push(103) push:104
> @rstack: print
@rstack: 100 101 102 103 [ 104 ]
> @rstack: push:10 push:100 mul print
@rstack: 100 101 102 103 104 [ 1000 ]
> @rstack: flip print
@rstack perspective flipped to: LIFO
@rstack: 100 101 102 103 [ 104 ]
> @dstack: bring(int, rstack) print
Brought value from int stack 'rstack' to selected stack 'dstack'
@dstack: [ 104 ] 1 2 3 4 5 6 
```

## Advanced Forth‑Like Operations

- **Tuck:**  
  - **Syntax:** `a b  →  b a b`
  - **Description:** Duplicates the top element and inserts it just below the top. Useful for reusing a value in subsequent operations.
  
- **Pick:**  
  - **Syntax:** `... x_n ... x_0 n  →  ... x_n ... x_0 x_n`
  - **Description:** Copies the nth element (with 0 being the top) and pushes it onto the top without removing it.
  
- **Roll:**  
  - **Syntax:** `... x_n ... x_0 n  →  ... x_1 ... x_0 x_n`
  - **Description:** Removes the nth element and places it on the top of the stack.
  
- **Over:**  
  - **Syntax:** `a b  →  a b a`
  - **Description:** Copies the second element to the top, preserving the original order.
  
- **Dup2, Drop2, Swap2, Over2:**  
  - **Description:** These commands operate on pairs of elements. They allow you to duplicate, remove, swap, or copy the top two elements (or pairs) to facilitate more complex stack manipulations.

## Bring Operation

- **Syntax:**  
  `bring(int, <stackname>)`  
- **Description:**  
  The bring operation transfers a value from one integer stack (the source) to the currently selected stack (the target).  
  - It pops a value from the specified source stack and pushes it onto the target stack.
  - The source stack name can be given with or without the `@` prefix.
- **Use Case:**  
  Enables cross‑stack data movement without resorting to variables, maintaining the container-centric philosophy.

## FIFO, LIFO, and Flip Operations

- **FIFO:**  
  - **Command:** `fifo`
  - **Effect:** Sets the current stack's perspective to FIFO.  
    - In FIFO mode, push operations insert new elements at the front (index 0), so the printed stack shows the front element in square brackets at the beginning.
  - **Status Message:** Reports “FIFO” in uppercase.
  
- **LIFO:**  
  - **Command:** `lifo`
  - **Effect:** Sets the current stack's perspective to LIFO.
    - In LIFO mode, push operations append new elements at the end (top of the stack).  
    - When printed, the top element is shown in square brackets at the end.
  - **Status Message:** Reports “LIFO” in uppercase.
  
- **Flip:**  
  - **Command:** `flip`
  - **Effect:** Toggles the current perspective between FIFO and LIFO without altering the physical order of the stack.  
    - This command does not re-order the elements; it simply changes how future operations (like push) are interpreted.
  - **Status Message:** Reports the new perspective in uppercase (e.g., “FIFO” or “LIFO”).

## @spawn Stack

- **Description:**  
  The @spawn stack is a dedicated mechanism for managing concurrent tasks (goroutines).  
- **Features:**  
  - It registers and controls live execution contexts.
  - Supported spawn commands include:
    - `list`: Display the list of running tasks.
    - `add <name>`: Create a new spawn task.
    - `pause <name>`, `resume <name>`, `stop <name>`: Control individual spawn tasks.
  - When the spawn stack is selected, these commands enable you to monitor and control concurrent activities.

## Notes on Syntax

iual supports both colon-based syntax and function-like syntax:
- **Colon-Based Syntax:**  
  - Compact and ideal for chaining simple operations.  
  - Example: `@dstack: push:10 push:20 print`
- **Function-Like Syntax:**  
  - Uses parentheses to enclose parameters, which is useful for nesting operations and when multiple parameters are needed.  
  - Example: `@dstack: push(10) push(20) print`
- **Mixing Syntaxes:**  
  - You can mix both styles in a compound command, so choose the one that best suits the complexity of the expression.

## Future Improvements

- Integrate stack perspectives more deeply into a unified UalContext.
- Extend support for typed stacks for strings and floats.
- Enhance cross-stack operations and nesting capabilities.
- Refine error handling and reporting for more robust command parsing.

---

This README now covers advanced Forth‑like commands, the bring operation, the FIFO/LIFO/flip perspective controls, and the @spawn stack, along with detailed explanations of the two syntactic styles. Let me know if you need further modifications or additional details.
