# test-setup.sh
#!/bin/bash
set -e  # Exit on error

echo "=== Testing P2P NAT Traversal System Setup ==="
echo ""

# Check Go installation
echo "Checking Go installation..."
go version
echo "✓ Go is installed"
echo ""

# Initialize Go modules if not already initialized
if [ ! -f "go.mod" ]; then
  echo "Initializing Go modules..."
  go mod init github.com/bOguzhan/NATbypass
  echo "✓ Go modules initialized"
else
  echo "✓ Go modules already initialized"
fi

# Ensure dependencies are downloaded
echo "Downloading dependencies..."
go mod tidy
echo "✓ Dependencies downloaded"
echo ""

# Build the mediatory server
echo "Building mediatory server..."
go build -o bin/mediatory-server cmd/mediatory-server/main.go
echo "✓ Mediatory server built successfully"
echo ""

# Build the application server
echo "Building application server..."
go build -o bin/application-server cmd/application-server/main.go
echo "✓ Application server built successfully"
echo ""

# Build the STUN test utility
echo "Building STUN test program..."
go build -o bin/stun_test ./test/main.go  # Changed from test/stun_test.go to test/main.go
echo "✓ STUN test program built successfully"
echo ""

# Test loading configuration
echo "Testing configuration system..."
go test -v ./test/config
echo "✓ Configuration system test passed"
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

# Kill the mediatory server
echo "Stopping mediatory server..."
kill $MEDIATORY_PID
echo "✓ Mediatory server stopped"
echo ""

echo "=== All tests completed successfully! ==="
echo "You can now run the STUN test manually with: bin/stun_test"
echo "Run the integrated test suite with: go test ./..."
echo ""

# Run each test script
./test-mediatory.sh
./test-signaling-api.sh
./test-udp-server.sh
./test-tcp-server.sh  # Add this line to include TCP server tests

echo "All tests completed successfully!"

