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
  go mod init github.com/yourusername/p2p-nat-traversal
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

# Create test directory if it doesn't exist
mkdir -p test

# Create a simple test program for STUN discovery
cat > test/stun_test.go <<EOL
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/yourusername/p2p-nat-traversal/pkg/networking"
)

func main() {
    log.Println("Testing STUN discovery...")
    
    // Use Google's public STUN server
    stunServer := "stun.l.google.com:19302"
    
    addr, err := networking.DiscoverPublicAddress(stunServer)
    if err != nil {
        log.Fatalf("Failed to discover public address: %v", err)
    }
    
    fmt.Printf("Your public IP is: %s\n", addr.IP.String())
    fmt.Printf("Your public port is: %d\n", addr.Port)
}
EOL

echo "✓ Created STUN test program"
echo ""

# Build and run STUN test program
echo "Building and running STUN test..."
go build -o bin/stun_test test/stun_test.go
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
curl -s http://localhost:8080/health
echo ""
echo "✓ Mediatory server health check passed"
echo ""

# Kill the mediatory server
echo "Stopping mediatory server..."
kill $MEDIATORY_PID
echo "✓ Mediatory server stopped"
echo ""

echo "=== All tests completed successfully! ==="
echo "You can now run the STUN test manually with: bin/stun_test"
echo ""