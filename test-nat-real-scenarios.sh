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

# Build test container using our NAT simulation Dockerfile
echo -e "${YELLOW}Building test container...${NC}"
docker build -t nat-test -f test/nat/Dockerfile.nat .
echo -e "${GREEN}✓ Test container built${NC}"
echo ""

# Function to run a test scenario
run_test_scenario() {
    local nat_type_1=$1
    local nat_type_2=$2
    local protocol=$3
    
    echo -e "${YELLOW}Testing $protocol traversal with $nat_type_1 NAT to $nat_type_2 NAT...${NC}"
    
    # Start first container with NAT type 1
    docker run -d --name peer1 --cap-add=NET_ADMIN --network nat-test-network \
        -e NAT_TYPE="$nat_type_1" \
        -e PROTOCOL="$protocol" \
        -e PEER_TYPE="initiator" \
        -e REMOTE_NAT_TYPE="$nat_type_2" \
        nat-test
    
    # Start second container with NAT type 2
    docker run -d --name peer2 --cap-add=NET_ADMIN --network nat-test-network \
        -e NAT_TYPE="$nat_type_2" \
        -e PROTOCOL="$protocol" \
        -e PEER_TYPE="responder" \
        -e REMOTE_NAT_TYPE="$nat_type_1" \
        nat-test
    
    # Wait for test to complete
    sleep 15
    
    # Check results
    PEER1_EXIT=$(docker inspect --format='{{.State.ExitCode}}' peer1)
    PEER2_EXIT=$(docker inspect --format='{{.State.ExitCode}}' peer2)
    
    # Get logs for debugging
    echo "Peer 1 logs:"
    docker logs peer1
    echo "Peer 2 logs:"
    docker logs peer2
    
    # Cleanup
    docker rm -f peer1 peer2 > /dev/null
    
    if [ "$PEER1_EXIT" -eq 0 ] && [ "$PEER2_EXIT" -eq 0 ]; then
        echo -e "${GREEN}✓ $protocol traversal succeeded between $nat_type_1 and $nat_type_2 NAT${NC}"
        return 0
    else
        echo -e "${RED}✗ $protocol traversal failed between $nat_type_1 and $nat_type_2 NAT (Exit codes: peer1=$PEER1_EXIT, peer2=$PEER2_EXIT)${NC}"
        return 1
    fi
}

# Test different NAT traversal scenarios
TESTS_PASSED=0
TESTS_TOTAL=0

# For faster testing during development, test a subset of combinations first
echo -e "${YELLOW}Running critical test scenarios first...${NC}"

# Critical tests - test full-cone to full-cone first (should always succeed)
TESTS_TOTAL=$((TESTS_TOTAL + 1))
if run_test_scenario "full-cone" "full-cone" "udp"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi
echo ""

# Test full-cone to symmetric (should work with UDP hole punching)
TESTS_TOTAL=$((TESTS_TOTAL + 1))
if run_test_scenario "full-cone" "symmetric" "udp"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi
echo ""

# Test symmetric to symmetric (hardest case, should use relaying)
TESTS_TOTAL=$((TESTS_TOTAL + 1))
if run_test_scenario "symmetric" "symmetric" "udp"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi
echo ""

# Only run full test matrix if critical tests pass or if FULL_TEST=1
if [ "$TESTS_PASSED" -eq 3 ] || [ "$FULL_TEST" = "1" ]; then
    echo -e "${YELLOW}Critical tests passed. Running full test matrix...${NC}"
    
    for protocol in "udp" "tcp"; do
        for nat1 in "full-cone" "address-restricted-cone" "port-restricted-cone" "symmetric"; do
            for nat2 in "full-cone" "address-restricted-cone" "port-restricted-cone" "symmetric"; do
                # Skip combinations we already tested
                if [ "$protocol" = "udp" ] && [ "$nat1" = "full-cone" ] && [ "$nat2" = "full-cone" ]; then
                    continue
                fi
                if [ "$protocol" = "udp" ] && [ "$nat1" = "full-cone" ] && [ "$nat2" = "symmetric" ]; then
                    continue
                fi
                if [ "$protocol" = "udp" ] && [ "$nat1" = "symmetric" ] && [ "$nat2" = "symmetric" ]; then
                    continue
                fi
                
                TESTS_TOTAL=$((TESTS_TOTAL + 1))
                if run_test_scenario "$nat1" "$nat2" "$protocol"; then
                    TESTS_PASSED=$((TESTS_PASSED + 1))
                fi
                echo ""
            done
        done
    done
else
    echo -e "${YELLOW}Critical tests failed. Skipping full test matrix.${NC}"
fi

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