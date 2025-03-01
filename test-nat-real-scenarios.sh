# File: test-nat-real-scenarios.sh

#!/bin/bash
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

set -e  # Exit on error

echo -e "${YELLOW}=== Testing NAT Traversal with Real Network Scenarios ===${NC}"
echo ""

# Check Docker is available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is required but not found. Please install Docker.${NC}"
    exit 1
fi

# Create a Docker network for testing
echo -e "${YELLOW}Setting up test network environment...${NC}"
docker network create --subnet=172.20.0.0/16 nat-test-network || true
echo -e "${GREEN}✓ Test network created${NC}"
echo ""

# Build test container
echo -e "${YELLOW}Building test container...${NC}"
cat > Dockerfile.test << EOF
FROM golang:1.19

WORKDIR /app
COPY . .

# Install required tools
RUN go mod download
RUN go build -o /app/bin/nat-test test/nat/test_client.go

ENTRYPOINT ["/app/bin/nat-test"]
EOF

docker build -t nat-test -f Dockerfile.test .
echo -e "${GREEN}✓ Test container built${NC}"
echo ""

# Function to run a test scenario
run_test_scenario() {
    local nat_type_1=$1
    local nat_type_2=$2
    local protocol=$3
    
    echo -e "${YELLOW}Testing $protocol traversal with $nat_type_1 NAT to $nat_type_2 NAT...${NC}"
    
    # Start first container with NAT type 1
    docker run -d --name peer1 --network nat-test-network \
        -e NAT_TYPE="$nat_type_1" \
        -e PROTOCOL="$protocol" \
        -e PEER_TYPE="initiator" \
        nat-test
    
    # Start second container with NAT type 2
    docker run -d --name peer2 --network nat-test-network \
        -e NAT_TYPE="$nat_type_2" \
        -e PROTOCOL="$protocol" \
        -e PEER_TYPE="responder" \
        nat-test
    
    # Wait for test to complete
    sleep 10
    
    # Check results
    PEER1_EXIT=$(docker inspect --format='{{.State.ExitCode}}' peer1)
    PEER2_EXIT=$(docker inspect --format='{{.State.ExitCode}}' peer2)
    
    # Cleanup
    docker rm -f peer1 peer2 > /dev/null
    
    if [ "$PEER1_EXIT" -eq 0 ] && [ "$PEER2_EXIT" -eq 0 ]; then
        echo -e "${GREEN}✓ $protocol traversal succeeded between $nat_type_1 and $nat_type_2 NAT${NC}"
        return 0
    else
        echo -e "${RED}✗ $protocol traversal failed between $nat_type_1 and $nat_type_2 NAT${NC}"
        return 1
    fi
}

# Test different NAT traversal scenarios
TESTS_PASSED=0
TESTS_TOTAL=0

for protocol in "udp" "tcp"; do
    for nat1 in "full-cone" "address-restricted-cone" "port-restricted-cone" "symmetric"; do
        for nat2 in "full-cone" "address-restricted-cone" "port-restricted-cone" "symmetric"; do
            TESTS_TOTAL=$((TESTS_TOTAL + 1))
            if run_test_scenario "$nat1" "$nat2" "$protocol"; then
                TESTS_PASSED=$((TESTS_PASSED + 1))
            fi
            echo ""
        done
    done
done

# Clean up
echo -e "${YELLOW}Cleaning up test environment...${NC}"
docker network rm nat-test-network || true
echo -e "${GREEN}✓ Test environment cleaned up${NC}"
echo ""

# Report results
echo -e "${YELLOW}=== NAT Traversal Tests Summary ===${NC}"
echo -e "${GREEN}$TESTS_PASSED${NC}/${YELLOW}$TESTS_TOTAL${NC} tests passed"
echo ""

SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_TOTAL))
if [ "$SUCCESS_RATE" -ge 80 ]; then
    echo -e "${GREEN}Success rate: $SUCCESS_RATE% (Target: 80%)${NC}"
    echo -e "${GREEN}✓ NAT traversal tests succeeded!${NC}"
    exit 0
else
    echo -e "${RED}Success rate: $SUCCESS_RATE% (Target: 80%)${NC}"
    echo -e "${RED}✗ NAT traversal tests failed to meet target success rate${NC}"
    exit 1
fi