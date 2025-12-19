#!/bin/bash
# run_negative_tests.sh - Test that invalid programs produce errors
#
# Usage: ./run_negative_tests.sh [--verbose]

cd "$(dirname "$0")/../.."

VERBOSE=false
[[ "$1" == "--verbose" ]] && VERBOSE=true

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

IUAL="./iual"
[[ ! -x "$IUAL" ]] && IUAL="./cmd/iual/iual"

PASS=0
FAIL=0

echo "=== Negative Tests: Parser Errors ==="
echo ""

for f in tests/negative/parser/*.ual; do
    name=$(basename "$f" .ual)
    
    # Parser errors should be caught - use timeout to prevent hangs
    output=$(timeout 2 $IUAL "$f" 2>&1) || true
    
    if echo "$output" | grep -qi "error\|panic\|invalid\|unexpected\|expected"; then
        echo -e "  ${GREEN}✓${NC} $name"
        ((PASS++))
    else
        echo -e "  ${RED}✗${NC} $name - expected error, got success"
        if $VERBOSE; then
            echo "    Output: $output"
        fi
        ((FAIL++))
    fi
done

echo ""
echo "=== Negative Tests: Runtime Errors ==="
echo ""

for f in tests/negative/runtime/*.ual; do
    name=$(basename "$f" .ual)
    
    # Runtime errors should produce non-zero exit or error output
    output=$(timeout 2 $IUAL "$f" 2>&1) || true
    
    if echo "$output" | grep -qi "error\|panic\|undefined\|underflow\|bounds\|invalid"; then
        echo -e "  ${GREEN}✓${NC} $name"
        ((PASS++))
    else
        echo -e "  ${RED}✗${NC} $name - expected error, got:"
        if $VERBOSE; then
            echo "    Output: $output"
        fi
        ((FAIL++))
    fi
done

echo ""
echo "=== Summary ==="
echo -e "Passed: ${GREEN}$PASS${NC}"
echo -e "Failed: ${RED}$FAIL${NC}"

if [[ $FAIL -gt 0 ]]; then
    exit 1
fi
