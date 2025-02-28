#!/bin/bash
set -e  # Exit on error

echo "=== Testing UDP Server Implementation ==="
echo ""

# Build the application server
echo "Building application server..."
make build-application
echo "✓ Application server built successfully"
echo ""

# Start the application server in the background
echo "Starting application server on port 9000..."
bin/application-server &
APPLICATION_PID=$!
echo "✓ Application server started with PID $APPLICATION_PID"
echo ""

# Sleep briefly to allow server to start
sleep 2

# Test the application server's health endpoint
echo "Testing application server health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:9000/health)
if [[ $HEALTH_RESPONSE == *"\"status\":\"ok\""* ]]; then
    echo "✓ Application server health check passed"
else
    echo "✗ Application server health check failed"
    echo "Response: $HEALTH_RESPONSE"
    # Kill the server before exiting
    kill $APPLICATION_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Run the UDP server tests
echo "Running UDP server unit tests..."
go test -v ./internal/nat -run "TestUDP.*"
TEST_EXIT_CODE=$?
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "✓ UDP server tests passed"
else
    echo "✗ UDP server tests failed with exit code $TEST_EXIT_CODE"
fi
echo ""

# Kill the application server
echo "Stopping application server..."
kill $APPLICATION_PID 2>/dev/null || true
echo "✓ Application server stopped"
echo ""

echo "=== UDP Server Tests Completed! ==="
exit $TEST_EXIT_CODE