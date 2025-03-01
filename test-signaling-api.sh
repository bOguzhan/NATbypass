#!/bin/bash
set -e  # Exit on error

echo "=== Testing Signaling API Endpoints ==="
echo ""

# Build the mediatory server if needed
echo "Building mediatory server..."
make build-mediatory
echo "✓ Mediatory server built successfully"
echo ""

# Start the mediatory server in the background
echo "Starting mediatory server on port 8080..."
bin/mediatory-server &
MEDIATORY_PID=$!
echo "✓ Mediatory server started with PID $MEDIATORY_PID"
echo ""

# Sleep to allow server to start
sleep 2

# Test the health endpoint to make sure the server is running
echo "Verifying the server is running..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
if [[ $HEALTH_RESPONSE != *"\"status\":\"ok\""* ]]; then
    echo "✗ Server is not responding correctly"
    echo "Response: $HEALTH_RESPONSE"
    kill $MEDIATORY_PID 2>/dev/null || true
    exit 1
fi
echo "✓ Server is running"
echo ""

# Register two test clients for connection testing
echo "Registering test client 1..."
CLIENT1_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name":"client-1", "properties":{"device":"laptop", "os":"linux"}}')

if [[ $CLIENT1_RESPONSE == *"\"status\":\"registered\""* ]]; then
    CLIENT1_ID=$(echo $CLIENT1_RESPONSE | grep -o '"client_id":"[^"]*"' | cut -d'"' -f4)
    echo "✓ Client 1 registered with ID: $CLIENT1_ID"
else
    echo "✗ Client 1 registration failed"
    echo "Response: $CLIENT1_RESPONSE"
    kill $MEDIATORY_PID 2>/dev/null || true
    exit 1
fi

echo "Registering test client 2..."
CLIENT2_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name":"client-2", "properties":{"device":"desktop", "os":"windows"}}')

if [[ $CLIENT2_RESPONSE == *"\"status\":\"registered\""* ]]; then
    CLIENT2_ID=$(echo $CLIENT2_RESPONSE | grep -o '"client_id":"[^"]*"' | cut -d'"' -f4)
    echo "✓ Client 2 registered with ID: $CLIENT2_ID"
else
    echo "✗ Client 2 registration failed"
    echo "Response: $CLIENT2_RESPONSE"
    kill $MEDIATORY_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test connection registration endpoint
echo "Creating connection request from client 1 to client 2..."
CONNECT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/connect \
  -H "Content-Type: application/json" \
  -d "{\"source_id\":\"$CLIENT1_ID\",\"target_id\":\"$CLIENT2_ID\",\"source_ip\":\"192.168.1.2\",\"source_port\":12345}")

if [[ $CONNECT_RESPONSE == *"\"status\":\"connection_registered\""* ]]; then
    CONN_ID=$(echo $CONNECT_RESPONSE | grep -o '"connection_id":"[^"]*"' | cut -d'"' -f4)
    echo "✓ Connection registered with ID: $CONN_ID"
else
    echo "✗ Connection registration failed"
    echo "Response: $CONNECT_RESPONSE"
    kill $MEDIATORY_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test getting active connections for a client
echo "Retrieving active connections for client 1..."
CONNECTIONS_RESPONSE=$(curl -s http://localhost:8080/api/v1/connections/$CLIENT1_ID)

if [[ $CONNECTIONS_RESPONSE == *"\"status\":\"success\""* ]]; then
    echo "✓ Retrieved connections successfully"
else
    echo "✗ Failed to retrieve connections"
    echo "Response: $CONNECTIONS_RESPONSE"
fi
echo ""

# Test sending signaling messages
echo "Sending test signal message from client 1 to client 2..."
SIGNAL_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/signal \
  -H "Content-Type: application/json" \
  -d "{\"type\":\"offer\",\"client_id\":\"$CLIENT1_ID\",\"target_id\":\"$CLIENT2_ID\",\"payload\":{\"sdp\":\"test-sdp-offer\"}}")

if [[ $SIGNAL_RESPONSE == *"\"status\":\"message_queued\""* ]]; then
    echo "✓ Signal message queued successfully"
else
    echo "✗ Failed to queue signal message"
    echo "Response: $SIGNAL_RESPONSE"
fi
echo ""

# Test retrieving messages for a client
echo "Retrieving queued messages for client 2..."
MESSAGES_RESPONSE=$(curl -s http://localhost:8080/api/v1/messages/$CLIENT2_ID)

if [[ $MESSAGES_RESPONSE == *"\"status\":\"success\""* ]]; then
    COUNT=$(echo $MESSAGES_RESPONSE | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "✓ Retrieved $COUNT messages for client 2"
else
    echo "✗ Failed to retrieve messages"
    echo "Response: $MESSAGES_RESPONSE"
fi
echo ""

# Kill the mediatory server
echo "Stopping mediatory server..."
kill $MEDIATORY_PID 2>/dev/null || true
echo "✓ Mediatory server stopped"
echo ""

echo "=== Signaling API Tests Completed! ==="