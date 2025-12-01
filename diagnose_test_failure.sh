#!/bin/bash

# Diagnostic script to identify why tests fail despite passing

echo "=========================================="
echo "Test Failure Diagnostic"
echo "=========================================="
echo ""

echo "Running tests with verbose output..."
echo ""

# Run tests and capture full output
go test ./cmd/test/... -v 2>&1 | tee /tmp/full_test_output.txt

EXIT_CODE=$?

echo ""
echo "=========================================="
echo "Analysis"
echo "=========================================="
echo "Test exit code: $EXIT_CODE"
echo ""

# Count test results
PASS_COUNT=$(grep -c "^--- PASS:" /tmp/full_test_output.txt || echo "0")
FAIL_COUNT=$(grep -c "^--- FAIL:" /tmp/full_test_output.txt || echo "0")
SKIP_COUNT=$(grep -c "^--- SKIP:" /tmp/full_test_output.txt || echo "0")

echo "Test Results:"
echo "  PASS: $PASS_COUNT"
echo "  FAIL: $FAIL_COUNT"
echo "  SKIP: $SKIP_COUNT"
echo ""

if [ $FAIL_COUNT -gt 0 ]; then
    echo "Failed tests:"
    grep "^--- FAIL:" /tmp/full_test_output.txt
    echo ""
fi

# Check for panics
PANIC_COUNT=$(grep -c "panic:" /tmp/full_test_output.txt || echo "0")
if [ $PANIC_COUNT -gt 0 ]; then
    echo "⚠️  Panics detected: $PANIC_COUNT"
    grep -A 5 "panic:" /tmp/full_test_output.txt
    echo ""
fi

# Check for fatal errors
FATAL_COUNT=$(grep -c "FATAL:" /tmp/full_test_output.txt || echo "0")
if [ $FATAL_COUNT -gt 0 ]; then
    echo "⚠️  Fatal errors detected: $FATAL_COUNT"
    grep "FATAL:" /tmp/full_test_output.txt
    echo ""
fi

# Check for race conditions
RACE_COUNT=$(grep -c "DATA RACE" /tmp/full_test_output.txt || echo "0")
if [ $RACE_COUNT -gt 0 ]; then
    echo "⚠️  Race conditions detected: $RACE_COUNT"
    echo ""
fi

echo "=========================================="
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Tests passed successfully!"
else
    echo "❌ Tests failed with exit code: $EXIT_CODE"
    echo ""
    echo "This could be caused by:"
    echo "  - A test calling t.Fail() or t.FailNow()"
    echo "  - An assertion failure"
    echo "  - A panic in a test or TestMain"
    echo "  - TestMain calling os.Exit() with non-zero code"
    echo ""
    echo "Check the full output in: /tmp/full_test_output.txt"
fi
echo "=========================================="

exit $EXIT_CODE

