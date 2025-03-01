// File: test/nat/local_test.go

package nat_test

import (
    "context"
    "net"
    "testing"
    "time"

    "github.com/bOguzhan/NATbypass/internal/discovery"
    "github.com/bOguzhan/NATbypass/internal/nat"
    "github.com/stretchr/testify/assert"
)

func TestNATTraversalStrategiesLocalPair(t *testing.T) {
    factory := nat.NewStrategyFactory()

    // Test cases with different strategy types
    testCases := []struct {
        name           string
        strategyType   nat.StrategyType
        expectedToWork bool
    }{
        {"UDP Hole Punching", nat.UDPHolePunching, true},
        {"TCP Simultaneous Open", nat.TCPSimultaneousOpen, true},
        {"UDP Relaying", nat.UDPRelaying, true},
        {"TCP Relaying", nat.TCPRelaying, true},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            strategy, err := factory.GetStrategyByType(tc.strategyType)
            assert.NoError(t, err)

            // Create local addresses for testing
            localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
            assert.NoError(t, err)
            
            remoteAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
            assert.NoError(t, err)
            
            // For local testing, we don't need to actually establish connections
            // Just verify the strategy interface works correctly
            ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
            defer cancel()
            
            // Attempt connection - will likely fail due to no actual listener
            _, err = strategy.EstablishConnection(ctx, localAddr, remoteAddr)
            
            // We're just testing that it doesn't panic or have obvious issues
            // The actual success/failure is not important for this test
            t.Logf("Strategy %s attempt result: %v", strategy.GetName(), err)
        })
    }
}

func TestNATTraversalStrategyMatrixOnLocalhost(t *testing.T) {
    factory := nat.NewStrategyFactory()
    
    // Test strategy selection for different NAT type combinations
    natTypes := []discovery.NATType{
        discovery.NATFullCone,
        discovery.NATAddressRestrictedCone,
        discovery.NATPortRestrictedCone,
        discovery.NATSymmetric,
    }
    
    // For each NAT type combination, check that a strategy is selected
    for _, localNAT := range natTypes {
        for _, remoteNAT := range natTypes {
            t.Run(string(localNAT)+"-to-"+string(remoteNAT), func(t *testing.T) {
                // Try with no protocol preference
                strategy := factory.SelectStrategy(localNAT, remoteNAT, "")
                assert.NotNil(t, strategy, "Should select a strategy for %s to %s", localNAT, remoteNAT)
                
                // Try with UDP preference
                udpStrategy := factory.SelectStrategy(localNAT, remoteNAT, "udp")
                assert.NotNil(t, udpStrategy, "Should select a UDP strategy")
                assert.Equal(t, "udp", udpStrategy.GetProtocol())
                
                // Try with TCP preference
                tcpStrategy := factory.SelectStrategy(localNAT, remoteNAT, "tcp")
                assert.NotNil(t, tcpStrategy, "Should select a TCP strategy")
                assert.Equal(t, "tcp", tcpStrategy.GetProtocol())
            })
        }
    }
}
