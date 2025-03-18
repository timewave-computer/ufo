#!/bin/bash
# Script to run the IBC test and capture the output to a file

# Ensure we're in the right directory
cd "$(dirname "$0")/.." || exit 1

echo "====================================================================="
echo "🔄 Running IBC Light Client test with detailed logging..."
echo "This test verifies:"
echo "  ✓ IBC connection and channel creation"
echo "  ✓ Validator set changes"
echo "  ✓ Epoch boundary crossing"
echo "  ✓ IBC packet sending and acknowledgment"
echo "  ✓ Light client updates"
echo "====================================================================="

# Clean up any previous output
rm -f test_output.log

# Run the test with a timeout and redirect all output to the log file
echo "Starting test at $(date '+%Y-%m-%d %H:%M:%S')"
timeout 10m nix develop .#ibc --command go test ./tests/ibc -run TestIBCLightClientUpdates -v > test_output.log 2>&1
EXIT_CODE=$?

# Check the exit code to determine the outcome
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ SUCCESS! Test passed successfully (exit code: 0)"
    echo "Showing summary of key events:"
    
    # Extract and display key events with timestamps
    echo "====================================================================="
    echo "📋 KEY EVENT TIMELINE"
    echo "====================================================================="
    
    # Chain startup events
    grep -E "CHAIN STARTED" test_output.log
    
    # IBC connection and channel creation events
    grep -E "IBC CLIENTS CREATED|IBC CONNECTION CREATED|IBC CHANNEL CREATED" test_output.log
    
    # Epoch boundary crossing
    grep -E "EPOCH BOUNDARY CROSSED" test_output.log
    
    # Validator set changes
    grep -E "VALIDATOR SET CHANGE" test_output.log
    
    # IBC operations
    grep -E "ICS-20 PACKET SENT|PACKETS RELAYED|ICS-20 ACKNOWLEDGED" test_output.log
    
    # Light client updates
    grep -E "LIGHT CLIENT UPDATED|CLIENT UPDATE DETECTED" test_output.log
    
    # Test summary
    echo ""
    echo "====================================================================="
    echo "📊 TEST SUMMARY"
    echo "====================================================================="
    grep -A 7 "==== TEST SUMMARY ====" test_output.log
    
elif [ $EXIT_CODE -eq 124 ]; then
    echo "❌ TIMEOUT! Test exceeded the 10-minute timeout."
    echo "Showing the last key events before timeout:"
    grep -E "CHAIN STARTED|IBC CLIENTS CREATED|IBC CHANNEL CREATED|EPOCH BOUNDARY CROSSED|VALIDATOR SET CHANGE|ICS-20 PACKET SENT|PACKETS RELAYED|LIGHT CLIENT UPDATED" test_output.log | tail -10
    
    echo "Showing the last 20 lines of log for debugging:"
    tail -n 20 test_output.log
else
    echo "❌ FAILURE! Test failed with exit code $EXIT_CODE."
    echo "Showing key failure points:"
    grep -E "FAIL:|Error:|fatal|panic" test_output.log | tail -10
    
    echo "Showing the last successful steps before failure:"
    grep -E "CHAIN STARTED|IBC CLIENTS CREATED|IBC CHANNEL CREATED|EPOCH BOUNDARY CROSSED|VALIDATOR SET CHANGE|ICS-20 PACKET SENT|PACKETS RELAYED|LIGHT CLIENT UPDATED" test_output.log | tail -5
    
    echo "Showing the last 20 lines of log for debugging:"
    tail -n 20 test_output.log
fi

# Check for key validation points in the test output
echo ""
echo "====================================================================="
echo "🔍 DETAILED VALIDATION REPORT"
echo "====================================================================="

# Check for test failure
if grep -q "FAIL.*github.com/timewave/ufo/tests/ibc" test_output.log; then
    TEST_FAILED=true
    echo "❌ TEST FAILED"
else
    TEST_FAILED=false
    echo "✅ TEST PASSED"
fi

# Check for IBC connection and channel creation
if grep -q "IBC CHANNEL CREATED" test_output.log; then
    echo "✅ IBC channel was successfully created"
else
    echo "❌ No IBC channel creation detected"
fi

# Check for validator set changes
if grep -q "VALIDATOR SET CHANGE" test_output.log; then
    echo "✅ Validator set changes detected: $(grep "VALIDATOR SET CHANGE" test_output.log | wc -l) rotations"
else
    echo "❌ No validator set changes detected"
fi

# Check for epoch boundary crossing
if grep -q "EPOCH BOUNDARY CROSSED" test_output.log; then
    echo "✅ Epoch boundary crossing confirmed: $(grep "EPOCH BOUNDARY CROSSED" test_output.log | wc -l) boundaries"
else 
    echo "❌ No epoch boundary crossing detected"
fi

# Check for IBC packet sending after validator change
if grep -q "ICS-20 PACKET SENT" test_output.log; then
    echo "✅ IBC packet sent: $(grep "ICS-20 PACKET SENT" test_output.log | wc -l) packets"
else
    echo "❌ No IBC packet sending detected"
fi

# Check for IBC packet acknowledgment
if grep -q "ICS-20 ACKNOWLEDGED" test_output.log; then
    echo "✅ IBC packet acknowledgment confirmed: $(grep "ICS-20 ACKNOWLEDGED" test_output.log | wc -l) acknowledgments"
else
    echo "❌ No IBC packet acknowledgment detected"
fi

# Check for client update verification
if grep -q "LIGHT CLIENT UPDATED\|CLIENT UPDATE DETECTED" test_output.log; then
    echo "✅ IBC light client updates confirmed: $(grep -E "LIGHT CLIENT UPDATED|CLIENT UPDATE DETECTED" test_output.log | wc -l) updates"
else
    echo "❌ No client updates detected"
fi

# Performance metrics
echo ""
echo "====================================================================="
echo "⏱️ PERFORMANCE METRICS"
echo "====================================================================="
# Get test duration
START_TIME=$(grep -m 1 "\[" test_output.log | sed 's/\[//;s/\].*//')
END_TIME=$(grep "\[" test_output.log | tail -1 | sed 's/\[//;s/\].*//')
if [[ -n "$START_TIME" && -n "$END_TIME" ]]; then
    START_SEC=$(date -j -f "%Y-%m-%d %H:%M:%S.%N" "$START_TIME" +%s 2>/dev/null || date -d "$START_TIME" +%s 2>/dev/null)
    END_SEC=$(date -j -f "%Y-%m-%d %H:%M:%S.%N" "$END_TIME" +%s 2>/dev/null || date -d "$END_TIME" +%s 2>/dev/null)
    if [[ -n "$START_SEC" && -n "$END_SEC" ]]; then
        DURATION=$((END_SEC - START_SEC))
        echo "Total test duration: ${DURATION} seconds"
    else
        echo "Could not calculate test duration"
    fi
fi

# Operation timing
echo "IBC channel creation time: $(grep -E "Creating IBC channel" -A 2 test_output.log | grep "IBC CHANNEL CREATED" | head -1 | sed 's/.*(\([^)]*\)).*/\1/' || echo "N/A")"
echo "Packet relay time: $(grep "PACKETS RELAYED" test_output.log | grep -o "took [^)]*" | head -1 || echo "N/A")"
echo "IBC transfer time: $(grep "ICS-20 PACKET SENT" test_output.log | grep -o "took [^)]*" | head -1 || echo "N/A")"

echo ""
echo "✨ Full log available in test_output.log"

# Overall test result
if [ "$TEST_FAILED" = false ] && [ $EXIT_CODE -eq 0 ]; then
    echo "====================================================================="
    echo "🎉 Test PASSED! All essential operations completed successfully."
    exit 0
else
    echo "====================================================================="
    echo "❌ Test FAILED! Check error details in test_output.log."
    exit 1
fi 