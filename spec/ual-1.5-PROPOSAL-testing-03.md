# ual 1.5 PROPOSAL: Integrated Testing System (Part 3)

This is not part of the ual spec at this time. All documents marked as PROPOSAL are refinements and the version number indicates the proposal it's targeting to be integrated with into the main ual spec in a forthcoming release.

This document extends the ual 1.5 Testing System PROPOSAL, addressing practical concerns around test expressiveness, usability, and integration with the broader development ecosystem.

---

## 15. Improved Test Expression Syntax

### 15.1 Background and Motivation

The core stack-based testing approach introduced in Parts 1 and 2 provides a powerful foundation that aligns with ual's design philosophy. However, the verbosity required for common testing patterns may create friction in day-to-day testing activities. A set of carefully designed syntactic abstractions can maintain the underlying stack model while offering more concise expression of common test patterns.

### 15.2 Concise Assertion Expressions

We propose a set of macro-based assertion helpers that provide shorthand for common test patterns:

```lua
-- Current approach
@stack: push:2 push:3 add
@expect: push:5
compare_eq()
if_false fail("Addition failed")

-- New concise notation
assert_eq(
  { push:2 push:3 add },  -- Actual (operation to test)
  { push:5 }              -- Expected
)
```

This approach maintains the stack-based model under the hood while offering a more direct syntax for common cases.

### 15.3 Implementation Details

These concise expressions are implemented as macros that expand to the underlying stack operations:

```lua
macro_define assert_eq(actual, expected)
  -- Execute actual operation
  #{actual}
  
  -- Store actual result
  @test_result: push(@stack: pop())
  
  -- Execute expected operation
  #{expected}
  
  -- Compare results
  @test_result: pop() @stack: pop() ==
  
  if_false {
    fail(sprintf("Expected %v but got %v", 
      @stack: peek(), @test_result: peek()))
  }
end_macro
```

Additional assertion helpers provide a rich vocabulary for common test patterns:

```lua
-- Check that values are not equal
assert_ne({ push:3 push:2 sub }, { push:0 })

-- Approximate equality for floating point
assert_approx({ push:3.141592 }, { push:3.14 }, 0.01)

-- Check that a value is within a range
assert_range({ push:sensor.read() }, 3.2, 3.8)

-- Check that a value is greater than another
assert_gt({ push:counter() }, { push:0 })

-- Check that operation raises an error
assert_error({ push:"invalid" as_int() })

-- Check membership in a set
assert_contains({ push:get_status() }, {"ready", "running", "paused"})
```

## 16. Simplified Parameterized Testing

### 16.1 Background and Motivation

Parameterized testing (running the same test with different input parameters) is a cornerstone of thorough testing but can be verbose and awkward to express in a stack-based model. A more concise approach would make this common pattern more accessible while maintaining ual's explicit philosophy.

### 16.2 Table-Driven Test Pattern

We propose a simplified table-driven test pattern that combines stack operations with a more direct parameter specification:

```lua
-- Define test cases in a table
test_cases = {
  {a = 0, b = 0, expected = 0},
  {a = 1, b = 0, expected = 1},
  {a = 1, b = 1, expected = 2},
  {a = -1, b = 1, expected = 0},
  {a = 5, b = 3, expected = 8}
}

-- Run the same test with all cases
for_each_case(test_cases, function(case)
  assert_eq(
    { push:case.a push:case.b add },
    { push:case.expected }
  )
end)
```

This approach separates data from test logic while maintaining an explicit connection between them.

### 16.3 Implementation Details

The `for_each_case` function is a simple helper that iterates through test cases and runs the provided test function:

```lua
function for_each_case(cases, test_func)
  for i = 1, #cases do
    local case = cases[i]
    
    -- Create a test context that includes the case index
    local context = {
      case = case,
      index = i,
      total = #cases
    }
    
    -- Execute the test within a protected call
    local success, error_msg = pcall(function()
      test_func(context)
    end)
    
    -- Report any failures with case-specific information
    if not success then
      fail(sprintf("Case %d/%d failed: %s", 
        i, #cases, error_msg))
    end
  end
end
```

### 16.4 Named Test Cases

For clearer test reporting, cases can include descriptive names:

```lua
test_cases = {
  {name = "both zero", a = 0, b = 0, expected = 0},
  {name = "identity", a = 1, b = 0, expected = 1},
  {name = "basic addition", a = 1, b = 1, expected = 2},
  {name = "negative numbers", a = -1, b = 1, expected = 0}
}

-- Test failures report the case name for easier debugging
for_each_case(test_cases, function(ctx)
  assert_eq(
    { push:ctx.case.a push:ctx.case.b add },
    { push:ctx.case.expected }
  )
end)
```

## 17. Cognitive Load Reduction through Test Contexts

### 17.1 Background and Motivation

Managing multiple test stacks (`@stack`, `@expect`, `@assert`, etc.) can increase cognitive load during testing. A more structured approach can maintain stack semantics while providing a more intuitive testing environment.

### 17.2 Unified Test Context

We propose a unified test context that encapsulates the various stacks and operations needed for testing:

```lua
function test_addition() {
  -- Create a test context
  @test: context("addition test")
  
  -- Operate within this context
  @test: {
    value_a = 2
    value_b = 3
    result = value_a + value_b
    expect(result).to_equal(5)
  }
}
```

This approach maintains the stack-based model internally while providing a more focused testing API.

### 17.3 Implementation Details

The test context is implemented as a specialized stack type with methods that encapsulate common testing patterns:

```lua
function create_test_context(name)
  local context = Stack.new(Any)
  context.name = name
  
  -- Store test values
  context.values = {}
  
  -- Add expect method that returns an expectation object
  context.expect = function(value)
    return {
      -- Various assertion methods
      to_equal = function(expected)
        if value ~= expected then
          fail(sprintf("Expected %v to equal %v", value, expected))
        end
      end,
      
      to_be_greater_than = function(expected)
        if not (value > expected) then
          fail(sprintf("Expected %v to be greater than %v", 
                       value, expected))
        end
      end,
      
      -- Additional matchers...
    }
  end
  
  return context
end
```

### 17.4 Maintaining Stack Semantics

While the test context provides a more direct API, it maintains stack semantics internally:

```lua
-- Execute operation in test context
@test: {
  push:2 push:3 add
  result = pop()
  expect(result).to_equal(5)
}

-- This expands to the equivalent stack operations
@stack: push:2 push:3 add
@result: push(@stack: pop())
@test: expect(@result: pop()).to_equal(5)
```

This approach reduces cognitive load while preserving ual's stack-based foundation.

## 18. Enhanced Testing Matchers

### 18.1 Background and Motivation

Modern testing frameworks provide rich matching capabilities beyond simple equality, allowing developers to express complex assertions concisely. A set of stack-based matchers would enhance ual's testing capabilities while maintaining its explicit approach.

### 18.2 Comprehensive Matcher Library

The following matchers expand ual's testing capabilities to handle a wide range of assertion types:

```lua
-- String matchers
@test: {
  text = "Hello, World!"
  expect(text).to_contain("Hello")
  expect(text).to_match("^Hello")
  expect(text).to_have_length(13)
}

-- Numeric matchers
@test: {
  value = 3.14159
  expect(value).to_be_approximately(3.14, 0.01)
  expect(value).to_be_between(3.0, 4.0)
  expect(value).to_be_positive()
}

-- Collection matchers
@test: {
  list = {1, 2, 3, 4, 5}
  expect(list).to_contain(3)
  expect(list).to_have_size(5)
  expect(list).to_all_satisfy(function(x) return x > 0 end)
}

-- Error matchers
@test: {
  expect(function() 
    throw("Invalid operation")
  }).to_throw()
  
  expect(function() 
    invalid_operation()
  }).to_throw_matching("Invalid")
}

-- Type matchers
@test: {
  value = get_value()
  expect(value).to_be_of_type("string")
  expect(value).to_be_a_table()
}
```

### 18.3 Implementation Details

These matchers are implemented as methods on the expectation object returned by `context.expect()`:

```lua
context.expect = function(value)
  return {
    -- Equality matchers
    to_equal = function(expected) { /* ... */ },
    to_be = function(expected) { /* ... */ },
    
    -- String matchers
    to_contain = function(substring) {
      if type(value) ~= "string" then
        fail("Expected a string but got " .. type(value))
      end
      if not string.find(value, substring, 1, true) then
        fail(sprintf("Expected '%s' to contain '%s'", 
                     value, substring))
      end
    },
    
    to_match = function(pattern) { /* ... */ },
    to_have_length = function(length) { /* ... */ },
    
    -- Numeric matchers
    to_be_approximately = function(expected, tolerance) { /* ... */ },
    to_be_between = function(min, max) { /* ... */ },
    
    -- Collection matchers
    to_contain = function(item) { /* ... */ },
    to_have_size = function(size) { /* ... */ },
    
    -- Error matchers
    to_throw = function() { /* ... */ },
    to_throw_matching = function(pattern) { /* ... */ },
    
    -- Type matchers
    to_be_of_type = function(expected_type) { /* ... */ },
    
    -- Custom matchers
    to_satisfy = function(predicate, description) {
      if not predicate(value) then
        fail(description or "Failed custom matcher")
      end
    }
  }
end
```

## 19. Enhanced Test Output and Reporting

### 19.1 Background and Motivation

Clear and informative test output is crucial for quickly identifying and fixing issues. A well-designed test reporting system helps developers understand failures and maintain their tests effectively.

### 19.2 Structured Test Output

The testing system produces structured output with multiple detail levels:

```
-- Summary output (default)
Test results: 42 passed, 3 failed, 0 skipped (45 total)
Failures:
  - test_complex_calculation (math.test.ual:45): Expected 42 but got 41
  - test_network_connection (network.test.ual:78): Timeout connecting to server
  - test_user_validation (users.test.ual:123): User ID not validated correctly

-- Verbose output (with -v flag)
RUNNING  test_simple_addition... PASS (0.001s)
RUNNING  test_complex_calculation... FAIL (0.003s)
  Expected 42 but got 41
  at math.test.ual:45
  Stack trace:
    math.test.ual:46 - assert_eq
    math.test.ual:45 - test_complex_calculation
RUNNING  test_subtraction... PASS (0.001s)
...
```

### 19.3 Implementation Details

Test output is generated using a structured approach that collects results and formats them appropriately:

```lua
function run_tests(options)
  local results = {
    passed = {},
    failed = {},
    skipped = {},
    duration = 0
  }
  
  -- Run each test
  for _, test in ipairs(discovered_tests) do
    local start_time = os.clock()
    local success, error_msg = pcall(test.func)
    local duration = os.clock() - start_time
    
    local result = {
      name = test.name,
      file = test.file,
      line = test.line,
      duration = duration
    }
    
    if success then
      table.insert(results.passed, result)
      if options.verbose then
        print(string.format("RUNNING  %s... PASS (%.3fs)", 
                            test.name, duration))
      end
    else
      result.error = error_msg
      table.insert(results.failed, result)
      if options.verbose then
        print(string.format("RUNNING  %s... FAIL (%.3fs)", 
                            test.name, duration))
        print("  " .. error_msg)
        print(string.format("  at %s:%d", test.file, test.line))
      end
    end
    
    results.duration = results.duration + duration
  end
  
  -- Print summary
  print(string.format("Test results: %d passed, %d failed, %d skipped (%d total)",
                      #results.passed, #results.failed, 
                      #results.skipped, #discovered_tests))
  
  if #results.failed > 0 then
    print("Failures:")
    for _, failure in ipairs(results.failed) do
      print(string.format("  - %s (%s:%d): %s", 
                         failure.name, failure.file, 
                         failure.line, failure.error))
    end
  end
  
  -- Generate formatted output based on selected reporter
  if options.reporter then
    generate_report(results, options.reporter)
  end
  
  return results
end
```

### 19.4 Visual Failure Reports

For failures, the testing system provides visual context to help understand what went wrong:

```
FAIL: test_array_comparison
Expected arrays to be equal:
  Expected: [1, 2, 3, 4, 5]
  Actual:   [1, 2, 7, 4, 5]
                  ^
  Difference at index 2: Expected 3 but got 7
```

This detailed output makes it easier to diagnose test failures, especially for complex comparisons.

## 20. External Tool Integration

### 20.1 Background and Motivation

Modern development workflows involve various tools like CI systems, IDEs, and code coverage services. To be practical for real-world use, ual's testing system must integrate smoothly with this ecosystem.

### 20.2 Standard Output Formats

The testing system supports generating output in standard formats used by external tools:

```lua
-- Generate JUnit XML for CI systems
ual test Math --format=junit --output=results.xml

-- Generate JSON for custom processing
ual test Math --format=json --output=results.json
```

The implementation uses templates for different output formats:

```lua
function generate_report(results, format, output_file)
  if format == "junit" then
    generate_junit_xml(results, output_file)
  elseif format == "json" then
    generate_json_report(results, output_file)
  elseif format == "html" then
    generate_html_report(results, output_file)
  else
    print("Unknown report format: " .. format)
  end
end
```

### 20.3 IDE Integration

For IDE integration, the testing system provides machine-readable output and supports running specific tests:

```bash
# Run a specific test (for IDE "Run Test" functionality)
ual test Math::test_addition

# Run tests with location information for IDE navigation
ual test Math --location-info
```

### 20.4 CI Integration Examples

The testing system includes examples for integrating with common CI systems:

```yaml
# GitHub Actions example
name: ual CI
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install ual
        run: |
          # Install ual steps
      - name: Run tests
        run: |
          ual test --all --format=junit --output=test-results.xml
      - name: Upload test results
        uses: actions/upload-artifact@v2
        with:
          name: test-results
          path: test-results.xml
```

Similar examples are provided for other CI systems like Jenkins, GitLab CI, and CircleCI.

### 20.5 Coverage Service Integration

The testing system can generate coverage reports in formats compatible with popular coverage services:

```bash
# Generate coverage report in Cobertura format
ual test --all --coverage --coverage-format=cobertura --coverage-output=coverage.xml

# Generate coverage report in LCOV format
ual test --all --coverage --coverage-format=lcov --coverage-output=lcov.info
```

## 21. Implementation Priorities

While the features outlined in this document provide a comprehensive testing system, not all need to be implemented immediately. We propose the following implementation priorities:

### 21.1 Tier 1: Essential Features (Initial Implementation)

1. **Concise assertion expressions** - Essential for making tests readable
2. **Simplified parameterized testing** - Critical for thorough testing without repetition
3. **Enhanced test output formatting** - Necessary for effective debugging

### 21.2 Tier 2: Important Extensions (Near-term)

1. **Test contexts** - Reduces cognitive load in complex tests
2. **Basic matcher library** - Provides essential comparison capabilities
3. **JUnit/JSON output** - Enables CI integration

### 21.3 Tier 3: Advanced Features (Long-term)

1. **Comprehensive matcher library** - Completes the testing vocabulary
2. **Advanced CI integration** - Streamlines workflows
3. **Coverage service integration** - Provides insights into test quality

This prioritization ensures that the most critical features are available first, with a clear path toward a full-featured testing system.

## 22. Conclusion

This extension addresses the practical concerns raised about the ual testing system while maintaining its core philosophy. By providing more concise syntax, reducing cognitive load, enhancing output, and enabling integration with external tools, these refinements make ual's testing capabilities comparable to those of mature languages while preserving its unique stack-based approach.

The proposed enhancements maintain the balance between explicitness and usability that characterizes ual, ensuring that tests remain:

1. **Explicit** - The stack-based foundation remains visible
2. **Minimal** - Each feature serves a clear purpose without redundancy
3. **Zero-overhead** - Tests continue to incur no runtime cost in production builds
4. **Embedded-friendly** - Resources required for testing remain modest

With these refinements, ual's testing system provides a solid foundation for developing reliable software across various domains, from embedded systems to larger applications, while remaining true to the language's distinctive design philosophy.