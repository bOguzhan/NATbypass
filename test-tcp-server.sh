#!/bin/bash
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Building TCP server test...${NC}"
go build -o bin/tcp_server_test test/tcp_server_test.go

# Also build dedicated TCP server test
echo -e "${YELLOW}Building TCP server standalone test...${NC}"
go build -o bin/tcp_standalone_test test/tcp/main.go

echo -e "${YELLOW}Running TCP server test...${NC}"
if go test -v ./test -run TestTCPServer; then
    echo -e "${GREEN}TCP Server test completed successfully!${NC}"
else
    echo -e "${RED}TCP Server test failed!${NC}"
    exit 1
fi

echo -e "${YELLOW}You can also run the standalone TCP server test with:${NC}"
echo -e "${GREEN}./bin/tcp_standalone_test${NC}"

# Don't remove the built binaries so they can be used manually
echo -e "${YELLOW}Test binaries available in ./bin/${NC}"