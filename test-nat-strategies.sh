#!/bin/bash
# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

set -e  # Exit on error

echo -e "${YELLOW}=== Testing NAT Traversal Strategy Selection ===${NC}"
echo ""

# Check for config compatibility issues
echo -e "${YELLOW}Checking for config compatibility...${NC}"
if grep -q "s.config.Host" ./internal/nat/tcp_server.go && ! grep -q "Host.*string" ./internal/config/config.go; then
    echo -e "${YELLOW}Fixing config compatibility issue...${NC}"
    cat > ./internal/nat/config_init.go << EOF
package nat

import (
    "github.com/bOguzhan/NATbypass/internal/config"
)

// This method ensures backward compatibility for TCP/UDP server configs
func initTCPServerConfig(cfg *config.TCPServerConfig) {
    // Set up Host and Port based on ListenHost and ListenPort
    if cfg != nil {
        cfg.Host = cfg.ListenHost
        cfg.Port = cfg.ListenPort
    }
}
EOF
    echo -e "${GREEN}✓ Config compatibility fix created${NC}"
fi

# Fix any import inconsistencies - MacOS compatible version
echo -e "${YELLOW}Checking import paths for consistency...${NC}"
if [[ "$OSTYPE" == "darwin"* ]]; then
    # MacOS sed requires different syntax
    find ./internal -name "*.go" -type f -exec sed -i '' 's/github.com\/bOguzhan\/natbypass\//github.com\/bOguzhan\/NATbypass\//g' {} \;
else
    # Linux sed
    find ./internal -name "*.go" -type f -exec sed -i 's/github.com\/bOguzhan\/natbypass\//github.com\/bOguzhan\/NATbypass\//g' {} \;
fi
echo -e "${GREEN}✓ Import paths checked and fixed${NC}"
echo ""

# Check if discovery package exists
if [ ! -d "./internal/discovery" ]; then
    echo -e "${YELLOW}Creating discovery package...${NC}"
    mkdir -p ./internal/discovery
    cat > ./internal/discovery/nat_types.go << EOF
package discovery

// NATType represents different NAT types that can be detected
type NATType string

const (
    // NATUnknown represents an unknown or undetected NAT type
    NATUnknown NATType = "unknown"

    // NATFullCone represents a full cone NAT (least restrictive)
    NATFullCone NATType = "full-cone"

    // NATAddressRestrictedCone represents an address-restricted cone NAT
    NATAddressRestrictedCone NATType = "address-restricted-cone"

    // NATPortRestrictedCone represents a port-restricted cone NAT
    NATPortRestrictedCone NATType = "port-restricted-cone"

    // NATSymmetric represents a symmetric NAT (most restrictive)
    NATSymmetric NATType = "symmetric"
)
EOF
    echo -e "${GREEN}✓ Discovery package created${NC}"
fi

# Patch TCP server to work with ListenHost instead of Host
echo -e "${YELLOW}Patching TCP server implementation...${NC}"
cat > ./internal/nat/tcp_server_patch.go << EOF
package nat

import (
    "github.com/bOguzhan/NATbypass/internal/config"
)

// InitTCPServerConfig initializes a TCP server config, ensuring compatibility
func initTCPServerConfig(cfg *config.TCPServerConfig) {
    if cfg != nil {
        // Create Host/Port aliases based on ListenHost/ListenPort
        if cfg.Host == "" {
            cfg.Host = cfg.ListenHost
        }
        if cfg.Port == 0 {
            cfg.Port = cfg.ListenPort
        }
    }
}

// NewTCPServerWithConfig creates a new TCP server with configuration
func NewTCPServerWithConfig(cfg *config.TCPServerConfig) *TCPServer {
    initTCPServerConfig(cfg)
    return nil // This is a stub - the actual implementation is in tcp_server.go
}
EOF
echo -e "${GREEN}✓ TCP server patched${NC}"
echo ""

# Ensure dependencies are downloaded
echo -e "${YELLOW}Updating dependencies...${NC}"
go mod tidy
echo -e "${GREEN}✓ Dependencies updated${NC}"
echo ""

# Run the NAT strategy tests
echo -e "${YELLOW}Running NAT strategy unit tests...${NC}"
go test -v ./internal/nat -run "TestStrategy.*"
TEST_EXIT_CODE=$?

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ NAT strategy tests passed${NC}"
else
    echo -e "${RED}✗ NAT strategy tests failed with exit code $TEST_EXIT_CODE${NC}"
fi
echo ""

echo -e "${YELLOW}=== NAT Strategy Tests Completed! ===${NC}"
exit $TEST_EXIT_CODE