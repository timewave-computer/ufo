#!/bin/bash
# Script to run the IBC test and capture the output to a file

echo "Running IBC test with a 5-minute timeout and capturing output to test_output.log"

# Run test and tee output to both terminal and log file
timeout 5m nix develop .#ibc --command ./scripts/run_ibc_tests.sh patched -run TestIBCLightClientUpdates -v 2>&1 | tee test_output.log
EXIT_CODE=$?

if [ $EXIT_CODE -eq 124 ]; then
    echo "Test timed out after 5 minutes!"
    echo "Last 50 lines of output:"
    tail -n 50 test_output.log
    echo "Full log available in test_output.log"
    exit 1
fi

echo "Test completed with exit code: $EXIT_CODE"

# Look for key validation points in the test output
echo "Checking test results..."

# Check for test failure
if grep -q "FAIL.*github.com/timewave/ufo/tests/ibc" test_output.log; then
    TEST_FAILED=true
    echo "❌ TEST FAILED"
else
    TEST_FAILED=false
    echo "✅ TEST PASSED"
fi

# Check for validator set changes
if grep -q "Chain1 validator set rotated" test_output.log; then
    echo "✅ Validator set change detected!"
else
    echo "❌ No validator set change detected"
fi

# Check for epoch boundary crossing
if grep -q "completed epochs" test_output.log; then
    echo "✅ Epoch boundary crossing confirmed!"
else 
    echo "❌ No epoch boundary crossing detected"
fi

# Check for IBC packet sending after validator change
if grep -q "IBC transfer hash after validator changes" test_output.log; then
    echo "✅ IBC packet sent after validator changes!"
else
    echo "❌ No IBC packet sent after validator changes"
fi

# Check for client update verification
if grep -q "CLIENT UPDATE DETECTED" test_output.log; then
    echo "✅ IBC client update confirmed!"
else
    echo "❌ No client update detected"
fi

# Overall test result
if [ "$TEST_FAILED" = false ] && [ $EXIT_CODE -eq 0 ]; then
    echo "🎉 Test PASSED! All validations completed successfully."
    grep -A 5 "Test completed successfully" test_output.log || echo "Success message not found in log"
else
    echo "❌ Test FAILED! See error details below:"
    grep -A 10 "FAIL.*github.com/timewave/ufo/tests/ibc" test_output.log || tail -n 50 test_output.log
    exit 1
fi

echo "Full log available in test_output.log" 