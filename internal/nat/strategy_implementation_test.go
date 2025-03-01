// File: internal/nat/strategy_implementation_test.go

package nat

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bOguzhan/NATbypass/internal/discovery"
	"github.com/stretchr/testify/assert"
)

func TestStrategyImplementations(t *testing.T) {
	factory := NewStrategyFactory()
	strategies := factory.GetAvailableStrategies()

	// Check that all strategies implement the interface correctly
	for _, strategy := range strategies {
		t.Run(strategy.GetName(), func(t *testing.T) {
			// Test protocol is non-empty
			assert.NotEmpty(t, strategy.GetProtocol(), "Protocol should not be empty")

			// Test name is non-empty
			assert.NotEmpty(t, strategy.GetName(), "Name should not be empty")

			// Test success rate is between 0 and 1
			for _, local := range []discovery.NATType{
				discovery.NATFullCone,
				discovery.NATAddressRestrictedCone,
				discovery.NATPortRestrictedCone,
				discovery.NATSymmetric,
			} {
				for _, remote := range []discovery.NATType{
					discovery.NATFullCone,
					discovery.NATAddressRestrictedCone,
					discovery.NATPortRestrictedCone,
					discovery.NATSymmetric,
				} {
					rate := strategy.EstimateSuccessRate(local, remote)
					assert.GreaterOrEqual(t, rate, 0.0, "Success rate should be >= 0")
					assert.LessOrEqual(t, rate, 1.0, "Success rate should be <= 1")
				}
			}
		})
	}
}

func TestEstablishConnectionInterface(t *testing.T) {
	// Setup test UDP addresses
	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	assert.NoError(t, err)

	remoteAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	assert.NoError(t, err)

	// Test each strategy's EstablishConnection interface
	factory := NewStrategyFactory()
	for _, stratType := range []StrategyType{
		UDPHolePunching,
		TCPSimultaneousOpen,
		UDPRelaying,
		TCPRelaying,
	} {
		t.Run(string(stratType), func(t *testing.T) {
			strategy, err := factory.GetStrategyByType(stratType)
			assert.NoError(t, err)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// We just test that the call doesn't panic - we expect errors since these are stubs
			_, err = strategy.EstablishConnection(ctx, localAddr, remoteAddr)
			assert.NoError(t, err)
			// We don't assert on error as different strategies may return different errors
		})
	}
}
