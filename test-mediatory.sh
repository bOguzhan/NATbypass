#!/bin/bash
set -e  # Exit on error

echo "=== Testing Mediatory Server Implementation ==="
echo ""

# Build the mediatory server
echo "Building mediatory server..."
go build -o bin/mediatory-server cmd/mediatory-server/main.go
echo "✓ Mediatory server built successfully"
echo ""

# Start the mediatory server in the background
echo "Starting mediatory server on port 8080..."
bin/mediatory-server &
MEDIATORY_PID=$!
echo "✓ Mediatory server started with PID $MEDIATORY_PID"
echo ""

# Sleep briefly to allow server to start
sleep 2

# Test the mediatory server's health endpoint
echo "Testing mediatory server health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
if [[ $HEALTH_RESPONSE == *"\"status\":\"ok\""* ]]; then
    echo "✓ Mediatory server health check passed"
else
    echo "✗ Mediatory server health check failed"
    echo "Response: $HEALTH_RESPONSE"
    # Kill the server before exiting
    kill $MEDIATORY_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test the version endpoint
echo "Testing version endpoint..."
VERSION_RESPONSE=$(curl -s http://localhost:8080/version)
if [[ $VERSION_RESPONSE == *"\"version\":\"0.1.0\""* ]]; then
    echo "✓ Version endpoint check passed"
else
    echo "✗ Version endpoint check failed"
    echo "Response: $VERSION_RESPONSE"
fi
echo ""

# Test client registration endpoint
echo "Testing client registration..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test-client", "properties":{"device":"laptop", "os":"linux"}}')

if [[ $REGISTER_RESPONSE == *"\"status\":\"registered\""* ]]; then
    echo "✓ Client registration passed"
    # Extract client_id from response
    CLIENT_ID=$(echo $REGISTER_RESPONSE | grep -o '"client_id":"[^"]*"' | cut -d'"' -f4)
    echo "  Client ID: $CLIENT_ID"
else
    echo "✗ Client registration failed"
    echo "Response: $REGISTER_RESPONSE"
fi
echo ""

# Test address endpoint
echo "Testing address endpoint..."
ADDRESS_RESPONSE=$(curl -s http://localhost:8080/api/v1/address)
if [[ $ADDRESS_RESPONSE == *"\"status\":\"success\""* ]]; then
    echo "✓ Address endpoint passed"
    IP=$(echo $ADDRESS_RESPONSE | grep -o '"ip":"[^"]*"' | cut -d'"' -f4)
    echo "  Detected IP: $IP"
else
    echo "✗ Address endpoint failed"
    echo "Response: $ADDRESS_RESPONSE"
fi
echo ""

# Test heartbeat endpoint (if we got a client_id)
if [ ! -z "$CLIENT_ID" ]; then
    echo "Testing heartbeat endpoint..."
    HEARTBEAT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/heartbeat \
      -H "Content-Type: application/json" \
      -d "{\"client_id\":\"$CLIENT_ID\"}")
    
    if [[ $HEARTBEAT_RESPONSE == *"\"status\":\"ok\""* ]]; then
        echo "✓ Heartbeat endpoint passed"
    else
        echo "✗ Heartbeat endpoint failed"
        echo "Response: $HEARTBEAT_RESPONSE"
    fi
    echo ""
fi

# Test stats endpoint
echo "Testing stats endpoint..."
STATS_RESPONSE=$(curl -s http://localhost:8080/stats)
if [[ $STATS_RESPONSE == *"\"status\":\"ok\""* ]]; then
    echo "✓ Stats endpoint passed"
    echo "  Active clients: $(echo $STATS_RESPONSE | grep -o '"active_clients":[0-9]*' | cut -d':' -f2)"
else
    echo "✗ Stats endpoint failed"
    echo "Response: $STATS_RESPONSE"
fi
echo ""

# Kill the mediatory server
echo "Stopping mediatory server..."
kill $MEDIATORY_PID
echo "✓ Mediatory server stopped"
echo ""

echo "=== All tests completed! ==="