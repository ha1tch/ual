# ual 1.5 PROPOSAL: Integrated Testing System

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

---

## 1. Introduction

This document proposes an integrated testing system for the ual programming language, designed to allow developers to write, organize, and execute tests efficiently while maintaining the core principles of explicitness, minimalism, and suitability for embedded systems.

The proposed system leverages XTXT stream multiplexing to allow inline tests (via `:test` streams) while also supporting separate `.test.ual` files for cases where inline testing is not desired. This dual approach provides flexibility while ensuring that tests remain easy to discover, maintain, and execute without introducing runtime overhead in production builds.

## 2. Background and Motivation

### 2.1 The Need for a Unified Testing System

Testing is essential for ensuring correctness, maintainability, and system reliability. However, traditional testing approaches often fall short in the context of embedded or resource-constrained systems:

- They require separate files, making tests less discoverable and harder to maintain alongside the code they test.
- They introduce runtime overhead, which is particularly problematic in embedded contexts where resources are limited.
- They rely on complex external frameworks that do not integrate naturally into a language's existing paradigms.

The goal of this proposal is to create a testing model that fits ual's philosophy:
- Explicit yet minimal in its approach
- Incurring zero runtime overhead in production code
- Stack-based, structured, and deterministic in its execution

### 2.2 Lessons from Other Languages

#### 2.2.1 Go-Style Testing

The Go programming language provides a testing approach that has influenced this proposal in several ways:

```go
// In main.go
package main

func Add(a, b int) int {
    return a + b
}

// In main_test.go
package main

import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("Add(2, 3) = %d; want %d", got, want)
    }
}
```

**Beneficial aspects of Go's approach:**
- Tests reside in the same package as the implementation, allowing access to unexported identifiers.
- The `go test` command automatically discovers and runs tests.
- Tests are excluded from production builds without additional configuration.
- Test functions follow a simple convention (prefixed with `Test`) rather than requiring complex annotations.

**Limitations in Go's approach:**
- Tests require separate `_test.go` files, which can lead to file proliferation in large projects.
- The testing framework, while minimal, is still a separate entity rather than being integrated into the language.

#### 2.2.2 Rust's Testing Model

Rust offers an integrated testing approach with both inline and separate tests:

```rust
// Regular code
fn add(a: i32, b: i32) -> i32 {
    a + b
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_add() {
        assert_eq!(add(2, 3), 5);
    }
}
```

**Beneficial aspects of Rust's approach:**
- Tests can be included directly in the source file, improving discoverability.
- The `cargo test` command provides automated discovery and execution.
- Tests are conditionally compiled, ensuring zero overhead in production.

**Limitations in Rust's approach:**
- Heavily dependent on Rust's attribute system and macros.
- Tests in separate files have limited access to non-public items.

### 2.3 The ual Opportunity

ual has a unique opportunity to create a testing system that leverages its distinctive features:

- **XTXT Integration**: Using XTXT to embed tests directly in the source file as `:test` streams.
- **Dual Test Placement**: Falling back to `.test.ual` files if inline tests are not appropriate or not found.
- **Compiler-Level Integration**: Automatically including or excluding tests based on build mode.
- **Stack-Based Testing**: Leveraging ual's stack paradigm to create expressive, concise test assertions.

## 3. Proposed Testing System

### 3.1 Test Placement Options

The proposal offers two placement options for tests, each with its own strengths and use cases.

#### 3.1.1 Inline Tests with `:test` Streams (Preferred)

XTXT stream multiplexing allows tests to be embedded directly within the source file while maintaining logical separation:

```
Frame 1:
:code
package Math

function add(a, b) {
    return a + b
}

:test
function test_add() {
    // Test code here
}
```

In this approach:
- The `:code` stream contains the normal program code.
- The `:test` stream contains test functions and supporting code.
- The compiler ignores the `:test` stream during normal builds.
- The test runner extracts and executes the `:test` stream during test mode.

This approach is preferred when:
- Tests benefit from proximity to the code they test.
- Tests need access to package-private functionality.
- The codebase is small to medium in size.

#### 3.1.2 Separate Test Files (Fallback)

As an alternative, tests can be placed in separate `.test.ual` files that mirror the structure of the main source files:

```
// In math.ual
package Math

function add(a, b) {
    return a + b
}

// In math.test.ual
package Math

function test_add() {
    // Test code here
}
```

This approach is useful when:
- The main source file is already large or complex.
- Tests require significant setup or auxiliary functions.
- The team prefers strict separation of implementation and tests.

### 3.2 Test Definitions

Test functions follow a simple naming convention to support automatic discovery:

```lua
function test_add() {
    // Test code here
}
```

Each function whose name begins with `test_` is automatically executed when the test runner is invoked. This approach is inspired by Go's testing conventions, offering a balance of explicitness and simplicity.

### 3.3 Assertion Framework

The testing system leverages ual's stack-based nature to create a natural, expressive assertion framework.

#### 3.3.1 Stack-Based Assertions

Unlike traditional assertion libraries that rely on method calls or function invocations, ual's testing system treats assertions as stack operations:

```lua
function test_add() {
    @stack: push:2 push:3 add  -- Operation under test
    @assert: push:5            -- Expected result
    compare_eq()               -- Assertion
}
```

This approach matches ual's fundamental paradigm, making assertions feel natural rather than like bolt-on additions.

#### 3.3.2 Stack Manipulation for Complex Assertions

For more complex assertions, the testing system provides stack manipulation operations specifically designed for testing:

```lua
function test_complex_operation() {
    @data: push:input1 push:input2 complex_operation
    @expect: dup2
    @results: compare_deep_eq
    if_false fail("Complex operation produced incorrect result")
}
```

#### 3.3.3 Test Case Organization

Similar to how stacks are used for data flow in normal ual code, they can be used to organize test cases:

```lua
function test_add_cases() {
    @cases: push:{2, 3, 5}  -- input1, input2, expected
            push:{-1, 1, 0}
            push:{0, 0, 0}
    @cases: for_each(run_add_test)
}

function run_add_test(case) {
    case.input1 case.input2 add
    case.expected compare_eq
    if_false fail(sprintf("add(%d, %d) != %d", 
               case.input1, case.input2, case.expected))
}
```

This stack-based approach to test case organization enables a high degree of reuse and composability.

### 3.4 Testing Macros

The ual macro system is leveraged to provide syntactic sugar for common testing patterns while maintaining the underlying stack-based model.

#### 3.4.1 Test Definition Macros

```lua
macro_define test(name, body)
    function test_#{name}()
        #{body}
    end
end_macro

// Usage
test("addition", {
    @stack: push:2 push:3 add
    @assert: push:5
    compare_eq()
})
```

#### 3.4.2 Assertion Macros

```lua
macro_define assert_eq(actual, expected)
    #{actual}
    #{expected}
    compare_eq()
    if_false fail(sprintf("Expected %v, got %v", 
                       @expect: peek, @stack: peek))
end_macro

// Usage
test("addition", {
    assert_eq({push:2 push:3 add}, {push:5})
})
```

These macros transform into the underlying stack operations, allowing developers to choose between the more concise macro syntax or the more explicit stack operations depending on their preference.

## 4. Compiler and Test Runner Behavior

### 4.1 Compiler Test Mode

The ual compiler provides a special test mode that is activated when running tests:

```bash
ual test Math
```

In test mode, the compiler:
1. Looks for a `:test` stream inside each `.ual` file.
2. If no `:test` stream exists, checks for a `.test.ual` file with the same base name.
3. If both exist, prioritizes the `:test` stream and logs a warning.
4. If neither exists, reports "No tests found" for that package.

### 4.2 Test Selection Logic

The following table describes the compiler's behavior in different scenarios:

| Condition | Compiler Action | Log Output |
|-----------|-----------------|------------|
| `:test` stream found, no `.test.ual` | Runs `:test` | (No warning) |
| No `:test`, `.test.ual` exists | Runs `.test.ual` | (No warning) |
| `:test` and `.test.ual` exist | Runs `:test` | "Warning: Ignoring math.test.ual: :test stream takes precedence" |
| No `:test` and no `.test.ual` | Reports "No tests found" | "Error: No test cases found for package Math." |

### 4.3 Test Discovery and Execution

The test runner follows these steps:
1. Identify all functions with names beginning with `test_`
2. For each function:
   - Execute the function
   - Capture any failures or errors
   - Report results
3. Provide a summary of test execution

### 4.4 Build Mode Exclusion

When compiling for production with:

```bash
ual build Math
```

The compiler completely ignores:
- All `:test` streams in source files
- All `.test.ual` files

This ensures zero runtime overhead in production builds.

## 5. Implementation Details

### 5.1 XTXT Stream Parsing for Test Mode

The compiler's handling of XTXT streams for testing is implemented as follows:

```lua
function parse_for_testing(file_path)
    local content = read_file(file_path)
    local frames = parse_xtxt(content)
    
    local code_stream = extract_stream(frames, "code")
    local test_stream = extract_stream(frames, "test")
    
    if test_stream then
        return {
            code = code_stream,
            test = test_stream,
            has_inline_tests = true
        }
    else
        local test_file_path = file_path:gsub(".ual$", ".test.ual")
        if file_exists(test_file_path) then
            local test_content = read_file(test_file_path)
            return {
                code = code_stream,
                test = test_content,
                has_inline_tests = false
            }
        else
            return {
                code = code_stream,
                test = nil,
                has_inline_tests = false
            }
        }
    end
end
```

### 5.2 Test Runner Implementation

The test runner is implemented as a library that integrates with the ual compiler:

```lua
function run_tests(package_name)
    local package = load_package(package_name)
    local test_functions = discover_test_functions(package)
    
    local results = {
        passed = 0,
        failed = 0,
        errors = {}
    }
    
    for _, func in ipairs(test_functions) do
        local success, error_msg = pcall(func)
        if success then
            results.passed = results.passed + 1
        else
            results.failed = results.failed + 1
            table.insert(results.errors, {
                function_name = func_name,
                error = error_msg
            })
        end
    end
    
    report_results(results)
    
    return results.failed == 0
end
```

### 5.3 Stack Manipulation for Testing

The testing framework provides specialized stack operations for testing purposes:

```lua
function compare_eq()
    local expected = expect.pop()
    local actual = stack.pop()
    
    if actual == expected then
        push(true)
    else
        push(false)
    end
end

function compare_deep_eq()
    local expected = expect.pop()
    local actual = results.pop()
    
    if type(expected) ~= type(actual) then
        push(false)
        return
    end
    
    if type(expected) == "table" then
        for k, v in pairs(expected) do
            if actual[k] ~= v then
                push(false)
                return
            end
        end
        
        for k, v in pairs(actual) do
            if expected[k] == nil then
                push(false)
                return
            end
        end
        
        push(true)
    else
        push(actual == expected)
    end
end
```

## 6. Examples

### 6.1 Basic Test Example

```
Frame 1:
:code
package Math

function add(a, b) {
    return a + b
}

:test
function test_add() {
    @stack: push:2 push:3 add
    @expect: push:5
    compare_eq()
    if_false fail("Addition did not work correctly")
}
```

### 6.2 Test Cases Example

```
Frame 1:
:code
package Math

function factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

:test
function test_factorial() {
    @cases: push:{0, 1}
            push:{1, 1}
            push:{2, 2}
            push:{3, 6}
            push:{4, 24}
            push:{5, 120}
    
    @cases: for_each(run_factorial_test)
}

function run_factorial_test(case) {
    @stack: push(case[1]) factorial
    @expect: push(case[2])
    compare_eq()
    if_false fail(sprintf("factorial(%d) != %d", 
               case[1], case[2]))
}
```

### 6.3 Stack-Based Test Utilities

```
Frame 1:
:test
-- Stack-based test helpers
function setup_test_stack() {
    @test_stack: clear
    @test_stack: push:10 push:20 push:30
    return test_stack
}

function test_stack_operations() {
    @stack: setup_test_stack()
    @stack: dup
    
    @expect: push:30 push:30
    compare_eq_stack_top(2)
    
    @stack: swap
    @expect: push:30 push:20
    compare_eq_stack_top(2)
}
```

## 7. Future Directions

The current proposal establishes a foundation for a deterministic, structured testing approach in ual. Future enhancements may include:

### 7.1 Mock Support

For embedded development, the ability to mock hardware interfaces and external dependencies is crucial. Future versions could add a `@mock` mechanism that leverages ual's stack-based design for elegant mocking:

```lua
@mock: gpio.read(12) returns:1
@mock: spi.transfer(0x42) returns:{0x01, 0x02, 0x03}
```

### 7.2 Code Coverage

Integrating code coverage analysis would provide insights into test effectiveness. This could be implemented at the compiler level to track execution paths through the stack operations.

### 7.3 Parameterized Testing

While the current proposal supports basic test case iteration, a more formalized parameterized testing framework could be developed:

```lua
@test.params: {
    {input: 1, expected: 1},
    {input: 2, expected: 2},
    {input: 3, expected: 6}
}

@test.run: function(param) {
    @stack: push(param.input) factorial
    @expect: push(param.expected)
    compare_eq()
}
```

### 7.4 Test Organization Beyond Package Scope

Future enhancements could support test suites that span multiple packages, enabling higher-level testing scenarios:

```lua
@test.suite: "System Integration"

function test_end_to_end_flow() {
    // Test that crosses package boundaries
}
```

## 8. Conclusion

This proposal introduces a comprehensive, lightweight, and structured testing system for ual that:

- Allows inline tests with `:test` streams (leveraging XTXT)
- Supports separate `.test.ual` files when needed
- Ensures the compiler automatically detects, executes, and isolates tests from production builds
- Maintains ual's philosophy of explicitness, minimalism, and zero runtime overhead
- Leverages ual's stack-based paradigm for natural, expressive test assertions

The proposed testing system provides a balance between integration with the language and separation of concerns, enabling developers to write tests that are both discoverable and maintainable while ensuring no testing overhead in production builds.

By building on ual's existing strengths—particularly its stack-based computation model and XTXT stream multiplexing capabilities—this testing system provides a natural extension to the language that feels consistent with its overall design philosophy.