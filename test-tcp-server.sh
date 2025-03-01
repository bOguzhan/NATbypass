#!/bin/bash
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Testing TCP Server Implementation ===${NC}"
echo ""

# Check if the TCP server implementation exists
if [ ! -f "internal/nat/tcp_server.go" ]; then
    echo -e "${RED}TCP server implementation not found in internal/nat/tcp_server.go${NC}"
    echo -e "${YELLOW}Make sure you've created the TCP server implementation first${NC}"
    exit 1
fi

# Clean any old test binaries to avoid stale code issues
echo -e "${YELLOW}Cleaning up previous test artifacts...${NC}"
rm -f bin/tcp_server_test

# Build the tests
echo -e "${YELLOW}Building TCP server tests...${NC}"
go test -c -o bin/tcp_server_test github.com/bOguzhan/NATbypass/internal/nat
echo -e "${GREEN}✓ TCP server test built successfully${NC}"
echo ""

# Run the TCP server tests
echo -e "${YELLOW}Running TCP server unit tests...${NC}"
go test -v github.com/bOguzhan/NATbypass/internal/nat -run TestTCPServer
TEST_EXIT_CODE=$?

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ TCP server tests passed${NC}"
else
    echo -e "${RED}✗ TCP server tests failed with exit code $TEST_EXIT_CODE${NC}"
fi
echo ""

# Build the standalone TCP server if it doesn't exist
if [ -f "test/tcp/main.go" ]; then
    echo -e "${YELLOW}Building TCP server standalone test...${NC}"
    mkdir -p bin
    rm -f bin/tcp_standalone_test  # Remove old binary if it exists
    go build -o bin/tcp_standalone_test test/tcp/main.go
    echo -e "${GREEN}✓ TCP server standalone test built successfully${NC}"
    echo -e "${YELLOW}You can run the standalone TCP server with:${NC}"
    echo -e "${GREEN}./bin/tcp_standalone_test${NC}"
    echo ""
fi

echo -e "${YELLOW}=== TCP Server Tests Completed! ===${NC}"
exit $TEST_EXIT_CODE