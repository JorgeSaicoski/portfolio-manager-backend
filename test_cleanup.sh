#!/bin/bash

# Test script to verify user cleanup functionality
echo "Testing User Cleanup Functionality..."
echo "======================================"

cd /home/bardockgaucho/GolandProjects/portfolio-manager/backend

# Run only the user cleanup tests
go test ./cmd/test -run TestUser_CleanupUserData -v -count=1 2>&1 | tee /tmp/cleanup_test.log

# Check for failures
if grep -q "FAIL.*TestUser_CleanupUserData" /tmp/cleanup_test.log; then
    echo ""
    echo "❌ TESTS FAILED"
    echo "==============="
    grep "Error:" /tmp/cleanup_test.log | head -20
    exit 1
else
    echo ""
    echo "✅ ALL TESTS PASSED"
    echo "=================="
    exit 0
fi

