
# ual 1.1 Standard Library Specification

## 1. Packages

The standard library includes the following packages:

1. **con** — Console or terminal operations  
2. **fmt** — String formatting  
3. **sys** — System-level operations  
4. **io** — Low-level input/output for digital pins  
5. **str** — String manipulation  
6. **math** — Basic numeric functions  

---

## 2. Package: con

### 2.1 Purpose

Provides basic operations for screen or console I/O. The underlying implementation determines how text is displayed (memory-mapped screen, serial terminal, etc.).

### 2.2 Specification

```lua
package con

-- Clears the console or screen and resets the cursor to (0,0).
function Cls()

-- Writes the string `s` at the current cursor location. The cursor
-- advances according to platform rules.
function Print(s)

-- Moves the cursor to coordinate (x, y), then writes the string `s`.
function Printat(x, y, s)
  At(x, y)
  Print(s)

-- Moves the console cursor to coordinate (x, y).
function At(x, y)
```

---

## 3. Package: fmt

### 3.1 Purpose

Offers minimal string formatting, allowing placeholders in a format string. Calls into `con` for printing.

### 3.2 Specification

```lua
package fmt

import "con"

-- Produces a string based on `format` and extra arguments, then immediately prints it.
-- Supported placeholders: 
--   %d -> decimal integer
--   %x -> hexadecimal integer
--   %s -> string
function Printf(format, ...)

-- Produces a string based on `format` and extra arguments, returning the result without printing.
-- The same placeholders apply as in Printf.
function Sprintf(format, ...)
```

---

## 4. Package: sys

### 4.1 Purpose

Houses essential system-level or OS-like functionality.

### 4.2 Specification

```lua
package sys

-- Terminates program execution, with an integer exit code.
function Exit(code)

-- Returns a monotonically increasing time value (e.g. milliseconds since startup).
function Millis()

-- Resets or reboots the system.
function Reboot()
```

---

## 5. Package: io

### 5.1 Purpose

Offers functions for configuring and controlling digital I/O pins, similar to GPIO. The platform determines how pins map to addresses or ports.

### 5.2 Specification

```lua
package io

-- Public constants for pin modes.
OUTPUT = 1
INPUT  = 0

-- Sets the mode of the pin (e.g. INPUT or OUTPUT).
function PinMode(pin, mode)

-- Writes a digital value (0 or 1) to an output pin.
function WritePin(pin, value)

-- Reads the current digital value (0 or 1) from the pin.
function ReadPin(pin)
```

---

## 6. Package: str

### 6.1 Purpose

Supplies basic string operations beyond simple concatenation.

### 6.2 Specification

```lua
package str

-- Returns the index of the first occurrence of `needle` in `haystack`,
-- or -1 if not found.
function Index(haystack, needle)

-- Splits `s` into an array of substrings, using `sep` as the delimiter.
-- For example, Split("A,B", ",") returns ["A","B"].
function Split(s, sep)

-- Joins the elements of `arr` (which should be an array of strings) into
-- one string, with `sep` placed between them.
function Join(arr, sep)
```

---

## 7. Package: math

### 7.1 Purpose

Provides arithmetic helpers beyond the built-in operators.

### 7.2 Specification

```lua
package math

-- Returns the absolute value of `n`.
function Abs(n)

-- Returns the smaller of `a` and `b`.
function Min(a, b)

-- Returns the larger of `a` and `b`.
function Max(a, b)

-- Performs integer exponentiation. For example, Pow(2,3) yields 8.
function Pow(base, exponent)
```

---

## 8. Usage Example

Below is a short **main.ual** sample showing how these packages might be used:

```lua
package main

import "con"
import "fmt"
import "sys"
import "io"
import "str"
import "math"

function main()
  con.Cls()
  fmt.Printf("Welcome!\n")

  local phrase = "Hello,World"
  local pos = str.Index(phrase, ",")        -- 5
  local parts = str.Split(phrase, ",")      -- ["Hello","World"]
  local joined = str.Join(parts, "-")       -- "Hello-World"
  fmt.Printf("Index: %d, joined: %s\n", pos, joined)

  io.PinMode(13, io.OUTPUT)
  io.WritePin(13, 1)
  
  local start = sys.Millis()
  while_true(sys.Millis() - start < 1000)
  end_while_true
  io.WritePin(13, 0)

  fmt.Printf("Abs(-5)=%d, Max(10,2)=%d\n", math.Abs(-5), math.Max(10,2))
  sys.Exit(0)
end
```

---

## 9. Implementation Requirements

1. **Exported Identifiers**: Must begin with an uppercase letter (e.g. `Cls`, `Printf`, `Exit`).  
2. **Behavior**: The functions in each package must operate as specified.  
3. **Platform Details**: The underlying implementation of console writes, pin I/O, system calls, etc. depends on the hardware or virtual machine.  
4. **Error Handling**: 
   - `str.Index` returns `-1` if not found.  
   - `math.Pow` can overflow, which may wrap or clamp to a maximum depending on the platform.  
   - `sys.Exit` terminates execution with the provided code.  

---

## 10. Conclusion

This standard library specification ensures that **ual** programs can:

- Perform **console** I/O (`con`)  
- Use **format** strings (`fmt`)  
- Execute **system** functions (`sys`)  
- Control **digital pins** (`io`)  
- Manipulate **strings** (`str`)  
- Use **basic math** helpers (`math`)

By adopting this minimal yet coherent set of packages, developers can write ual code that remains succinct, clear, and portable to multiple embedded or retro platforms.
