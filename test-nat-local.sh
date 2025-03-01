#!/bin/bash
# File: test-nat-local.sh

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

set -e  # Exit on error

echo -e "${YELLOW}=== Testing NAT Traversal Strategies Locally ===${NC}"
echo ""

# First ensure all files in test/nat directory use the same package
echo -e "${YELLOW}Checking package declarations consistency...${NC}"
if grep -q "^package nat$" ./test/nat/*.go 2>/dev/null; then
    echo -e "${YELLOW}Fixing package declarations...${NC}"
    # Use proper macOS syntax for sed
    find ./test/nat -name "*.go" -exec sed -i '' 's/^package nat$/package nat_test/g' {} \;
    echo -e "${GREEN}✓ Package declarations fixed${NC}"
else
    echo -e "${GREEN}✓ Package declarations are consistent${NC}"
fi
echo ""

# Check if test_client.go exists in the right directory
if [ ! -d "./test/nat_client" ]; then
    echo -e "${YELLOW}Creating nat_client directory...${NC}"
    mkdir -p ./test/nat_client
    
    # If the test_client.go file exists in test/nat, move it
    if [ -f "./test/nat/test_client.go" ]; then
        echo -e "${YELLOW}Moving test_client.go to correct location...${NC}"
        mv ./test/nat/test_client.go ./test/nat_client/
        echo -e "${GREEN}✓ test_client.go moved${NC}"
    fi
fi

# Create a simulate_nat.sh script if it doesn't exist
if [ ! -f "./test/nat/simulate_nat.sh" ]; then
    echo -e "${YELLOW}Creating simulate_nat.sh script...${NC}"
    cat > ./test/nat/simulate_nat.sh << 'EOF'
#!/bin/bash
# File: test/nat/simulate_nat.sh

# Get environment variables
NAT_TYPE=${NAT_TYPE:-"full-cone"}
PEER_TYPE=${PEER_TYPE:-"initiator"}
PROTOCOL=${PROTOCOL:-"udp"}

echo "Setting up $NAT_TYPE NAT environment for $PEER_TYPE using $PROTOCOL"

# In a real environment, this would configure iptables for NAT behavior
# For the test client, we just run it
/app/bin/nat-test
EOF
    chmod +x ./test/nat/simulate_nat.sh
    echo -e "${GREEN}✓ simulate_nat.sh created${NC}"
fi

# Update Dockerfile.nat if needed
if [ -f "./test/nat/Dockerfile.nat" ]; then
    echo -e "${YELLOW}Checking Dockerfile.nat...${NC}"
    if grep -q "test/nat/test_client.go" ./test/nat/Dockerfile.nat; then
        echo -e "${YELLOW}Updating Dockerfile.nat...${NC}"
        sed -i '' 's|test/nat/test_client.go|test/nat_client/test_client.go|g' ./test/nat/Dockerfile.nat
        echo -e "${GREEN}✓ Dockerfile.nat updated${NC}"
    else
        echo -e "${GREEN}✓ Dockerfile.nat is correct${NC}"
    fi
fi

echo -e "${YELLOW}Running local NAT traversal strategy tests...${NC}"

# Run the local NAT traversal tests
go test -v ./test/nat

TEST_EXIT_CODE=$?

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ Local NAT traversal tests passed${NC}"
else
    echo -e "${RED}✗ Local NAT traversal tests failed with exit code $TEST_EXIT_CODE${NC}"
fi

echo ""
echo -e "${YELLOW}=== Local NAT Traversal Tests Completed! ===${NC}"
exit $TEST_EXIT_CODE