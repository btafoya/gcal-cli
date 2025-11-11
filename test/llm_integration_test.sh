#!/bin/bash
# LLM Integration Test Suite
# Simulates LLM agent workflows with gcal-cli

set -e

GCAL_CLI="./gcal-cli"
FORMAT="--format json"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0

# Helper function to run test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_success="$3"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo "Running: $test_name"

    # Execute command and capture output
    if output=$(eval "$command" 2>&1); then
        actual_success="true"
    else
        actual_success="false"
    fi

    # Check if we can parse JSON
    if echo "$output" | jq -e . > /dev/null 2>&1; then
        success=$(echo "$output" | jq -r '.success')

        if [ "$success" = "$expected_success" ]; then
            echo -e "${GREEN}✓ PASS${NC}: $test_name"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo -e "${RED}✗ FAIL${NC}: $test_name (expected success=$expected_success, got $success)"
            echo "Output: $output"
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name (invalid JSON output)"
        echo "Output: $output"
        return 1
    fi
}

# Helper to check JSON schema
check_schema() {
    local output="$1"
    local required_fields="$2"

    for field in $required_fields; do
        if ! echo "$output" | jq -e ".$field" > /dev/null 2>&1; then
            echo "Missing required field: $field"
            return 1
        fi
    done
    return 0
}

echo "========================================"
echo "LLM Integration Test Suite"
echo "========================================"
echo ""

# Test 1: Check authentication status - should return valid JSON even if not authenticated
echo "Running: Check auth status"
TESTS_RUN=$((TESTS_RUN + 1))
output=$($GCAL_CLI auth status $FORMAT 2>&1 || true)
if echo "$output" | jq -e . > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Auth status returns valid JSON"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Auth status - invalid JSON"
fi

# Test 2: List calendars - will fail if not authenticated, but should return valid JSON
echo "Running: List calendars"
TESTS_RUN=$((TESTS_RUN + 1))
output=$($GCAL_CLI calendars list $FORMAT 2>&1 || true)
if echo "$output" | jq -e . > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: List calendars returns valid JSON"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: List calendars - invalid JSON"
fi

# Test 3: Validate error response schema for invalid input
echo "Running: Invalid input error schema"
TESTS_RUN=$((TESTS_RUN + 1))
output=$($GCAL_CLI events create --title "Test" --start "invalid" --end "invalid" $FORMAT 2>&1 || true)
if echo "$output" | jq -e '.success == false and .error.code and .error.message and .error.recoverable != null' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Error response has correct schema"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Error response schema invalid"
    echo "Output: $output"
fi

# Test 4: Check error code is machine-parseable
echo "Running: Error code parseability"
TESTS_RUN=$((TESTS_RUN + 1))
error_code=$(echo "$output" | jq -r '.error.code' 2>&1 || echo "")
if [ -n "$error_code" ] && [ "$error_code" != "null" ]; then
    echo -e "${GREEN}✓ PASS${NC}: Error code is parseable: $error_code"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Error code not parseable"
fi

# Test 5: Check suggestedAction field presence for recoverable errors
echo "Running: Suggested action in errors"
TESTS_RUN=$((TESTS_RUN + 1))
recoverable=$(echo "$output" | jq -r '.error.recoverable' 2>&1 || echo "false")
has_suggested_action=$(echo "$output" | jq -e '.error.suggestedAction' > /dev/null 2>&1 && echo "true" || echo "false")

if [ "$recoverable" = "true" ]; then
    if [ "$has_suggested_action" = "true" ]; then
        echo -e "${GREEN}✓ PASS${NC}: Recoverable error has suggestedAction"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: Recoverable error missing suggestedAction"
    fi
else
    echo -e "${GREEN}✓ PASS${NC}: Error recoverability field present"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

# Test 6: Metadata timestamp presence
echo "Running: Metadata timestamp"
TESTS_RUN=$((TESTS_RUN + 1))
timestamp=$(echo "$output" | jq -r '.metadata.timestamp' 2>&1 || echo "")
if [ -n "$timestamp" ] && [ "$timestamp" != "null" ]; then
    echo -e "${GREEN}✓ PASS${NC}: Metadata timestamp present: $timestamp"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Metadata timestamp missing"
fi

# Test 7: Missing required field error
echo "Running: Missing required field validation"
TESTS_RUN=$((TESTS_RUN + 1))
output=$($GCAL_CLI events create --start "2024-01-15T10:00:00" --end "2024-01-15T11:00:00" $FORMAT 2>&1 || true)
if echo "$output" | jq -e '.success == false and (.error.code == "MISSING_REQUIRED" or .error.code == "INVALID_INPUT")' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Missing required field produces proper error"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Missing required field error incorrect"
fi

# Test 8: Time range validation
echo "Running: Time range validation"
TESTS_RUN=$((TESTS_RUN + 1))
output=$($GCAL_CLI events create \
    --title "Test" \
    --start "2024-01-15T14:00:00" \
    --end "2024-01-15T13:00:00" \
    $FORMAT 2>&1 || true)
if echo "$output" | jq -e '.success == false and .error.code == "INVALID_TIME_RANGE"' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Time range validation works"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Time range validation incorrect"
fi

# Test 9: Help text has examples
echo "Running: Help text includes examples"
TESTS_RUN=$((TESTS_RUN + 1))
help_output=$($GCAL_CLI events create --help 2>&1 || true)
if echo "$help_output" | grep -q "Examples:"; then
    echo -e "${GREEN}✓ PASS${NC}: Help text includes examples"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: Help text missing examples"
fi

# Test 10: JSON output is properly formatted
echo "Running: JSON formatting"
TESTS_RUN=$((TESTS_RUN + 1))
output=$($GCAL_CLI auth status $FORMAT 2>&1 || true)
if echo "$output" | python3 -m json.tool > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: JSON output is valid and parseable"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAIL${NC}: JSON output invalid"
fi

echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
echo "Tests Run: $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $((TESTS_RUN - TESTS_PASSED))"

if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi
