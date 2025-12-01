#!/bin/bash
# Quick verification script to confirm coverage solution is working

echo "🔍 Verifying Test Coverage Solution..."
echo ""

# Check if we're in the backend directory
if [ ! -f "Makefile" ]; then
    echo "❌ Error: Run this script from the backend directory"
    exit 1
fi

# Run coverage tests
echo "📊 Running coverage tests..."
make test-coverage > /tmp/coverage_check.txt 2>&1
EXIT_CODE=$?

# Extract coverage percentage
COVERAGE=$(grep "total:" /tmp/coverage_check.txt | awk '{print $3}' | sed 's/%//')

echo ""
echo "=========================================="
echo "         Coverage Verification"
echo "=========================================="
echo ""
echo "Coverage: $COVERAGE%"
echo "Target:   85%"
echo "Status:   $([ $EXIT_CODE -eq 0 ] && echo '✅ PASS' || echo '❌ FAIL')"
echo ""
echo "=========================================="
echo ""

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Coverage solution is working correctly!"
    echo ""
    echo "Details:"
    echo "  - Coverage is above 85% threshold"
    echo "  - Make target exits with success (0)"
    echo "  - HTML report generated: coverage.html"
    echo ""
    echo "To view detailed coverage:"
    echo "  xdg-open coverage.html"
else
    echo "❌ Coverage test failed"
    echo ""
    echo "Check the logs:"
    echo "  cat /tmp/coverage_check.txt"
fi

exit $EXIT_CODE

