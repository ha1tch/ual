# Negative Tests

This directory contains tests that verify ual correctly rejects invalid programs.

## Running

```bash
./run_negative_tests.sh           # Run all negative tests
./run_negative_tests.sh --verbose # Show output on failure
```

## Test Categories

### Parser Errors (`parser/`)

Tests that verify the parser rejects syntactically invalid programs:

| File | Tests |
|------|-------|
| `err_unclosed_brace.ual` | Missing closing brace |
| `err_invalid_token.ual` | Invalid token (###) |
| `err_missing_paren.ual` | Unclosed parenthesis |
| `err_compute_no_bindings.ual` | Compute block without `\|..\|` |
| `err_function_no_body.ual` | Function declaration without body |
| `err_while_no_condition.ual` | While without condition |

### Runtime Errors (`runtime/`)

Tests that verify runtime errors are caught and reported:

| File | Tests |
|------|-------|
| `err_type_mismatch.ual` | Pop f64 to i64 dstack |
| `err_undefined_var.ual` | Use of undeclared variable |
| `err_undefined_func.ual` | Call to undefined function |
| `err_undefined_stack.ual` | Reference to undefined stack |
| `err_array_bounds.ual` | Array index out of bounds |

## Adding Tests

1. Create a `.ual` file in `parser/` or `runtime/`
2. Name it `err_<description>.ual`
3. Include a comment explaining what error is expected
4. Run `./run_negative_tests.sh` to verify

The test passes if the output contains "error", "panic", "invalid", "unexpected", or similar error indicators.

## Exit Codes

- `0` — All tests passed (errors were correctly detected)
- `1` — One or more tests failed (invalid program was accepted)
