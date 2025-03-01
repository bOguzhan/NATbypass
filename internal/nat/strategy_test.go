package nat

import (
	"context"
	"net"
	"testing"

	"github.com/bOguzhan/NATbypass/internal/discovery"
	"github.com/stretchr/testify/assert"
)

func TestStrategyFactory(t *testing.T) {
	factory := NewStrategyFactory()

	// Test factory initialization
	assert.NotNil(t, factory)
	assert.NotEmpty(t, factory.strategies)

	// Test all strategies were registered
	strategies := factory.GetAvailableStrategies()
	assert.Len(t, strategies, 4)

	strategyNames := make(map[string]bool)
	for _, s := range strategies {
		strategyNames[s.GetName()] = true
	}

	assert.True(t, strategyNames["UDP Hole Punching"])
	assert.True(t, strategyNames["TCP Simultaneous Open"])
	assert.True(t, strategyNames["UDP Relaying"])
	assert.True(t, strategyNames["TCP Relaying"])
}

func TestStrategyRetrieval(t *testing.T) {
	factory := NewStrategyFactory()

	// Test retrieving specific strategy by type
	udpStrategy, err := factory.GetStrategyByType(UDPHolePunching)
	assert.NoError(t, err)
	assert.Equal(t, "UDP Hole Punching", udpStrategy.GetName())
	assert.Equal(t, "udp", udpStrategy.GetProtocol())

	tcpStrategy, err := factory.GetStrategyByType(TCPSimultaneousOpen)
	assert.NoError(t, err)
	assert.Equal(t, "TCP Simultaneous Open", tcpStrategy.GetName())
	assert.Equal(t, "tcp", tcpStrategy.GetProtocol())

	// Test error for non-existent strategy
	_, err = factory.GetStrategyByType("non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrStrategyNotFound, err)
}

func TestStrategySelection(t *testing.T) {
	factory := NewStrategyFactory()

	testCases := []struct {
		localNAT         discovery.NATType
		remoteNAT        discovery.NATType
		preferredProto   string
		expectedStrategy string
	}{
		// Full cone to full cone should select UDP hole punching
		{discovery.NATFullCone, discovery.NATFullCone, "", "UDP Hole Punching"},

		// Symmetric NAT to symmetric NAT should prefer relaying
		{discovery.NATSymmetric, discovery.NATSymmetric, "", "UDP Relaying"},

		// Protocol preference should be respected when possible
		{discovery.NATFullCone, discovery.NATFullCone, "tcp", "TCP Simultaneous Open"},

		// Protocol preference should fall back when NAT types are incompatible
		{discovery.NATSymmetric, discovery.NATSymmetric, "tcp", "TCP Relaying"},

		// Address restricted to port restricted should use UDP
		{discovery.NATAddressRestrictedCone, discovery.NATPortRestrictedCone, "", "UDP Hole Punching"},

		// Protocol preference should override NAT compatibility
		{discovery.NATFullCone, discovery.NATAddressRestrictedCone, "tcp", "TCP Simultaneous Open"},
	}

	for i, tc := range testCases {
		strategy := factory.SelectStrategy(tc.localNAT, tc.remoteNAT, tc.preferredProto)
		assert.NotNil(t, strategy, "Test case %d should find a strategy", i)
		assert.Equal(t, tc.expectedStrategy, strategy.GetName(), "Test case %d failed", i)
	}
}

func TestSuccessRateEstimation(t *testing.T) {
	factory := NewStrategyFactory()

	// Get the strategies
	udpStrategy, _ := factory.GetStrategyByType(UDPHolePunching)
	tcpStrategy, _ := factory.GetStrategyByType(TCPSimultaneousOpen)
	udpRelay, _ := factory.GetStrategyByType(UDPRelaying)
	tcpRelay, _ := factory.GetStrategyByType(TCPRelaying)

	// Test UDP hole punching success rates
	// Full cone to anything should have high success
	assert.InDelta(t, 0.95, udpStrategy.EstimateSuccessRate(discovery.NATFullCone, discovery.NATPortRestrictedCone), 0.01)

	// Symmetric to symmetric should have low success
	assert.InDelta(t, 0.10, udpStrategy.EstimateSuccessRate(discovery.NATSymmetric, discovery.NATSymmetric), 0.01)

	// Test TCP simultaneous open success rates
	// Generally lower than UDP
	assert.Less(t,
		tcpStrategy.EstimateSuccessRate(discovery.NATPortRestrictedCone, discovery.NATPortRestrictedCone),
		udpStrategy.EstimateSuccessRate(discovery.NATPortRestrictedCone, discovery.NATPortRestrictedCone))

	// Relaying should have consistently high success rates
	assert.GreaterOrEqual(t, udpRelay.EstimateSuccessRate(discovery.NATSymmetric, discovery.NATSymmetric), 0.95)
	assert.GreaterOrEqual(t, tcpRelay.EstimateSuccessRate(discovery.NATSymmetric, discovery.NATSymmetric), 0.95)
}

// Mock implementation for TraversalStrategy interface
type mockTraversalStrategy struct {
	name     string
	protocol string
	success  float64
}

func (m *mockTraversalStrategy) EstablishConnection(ctx context.Context, localAddr, remoteAddr *net.UDPAddr) (net.Conn, error) {
	return nil, nil
}

func (m *mockTraversalStrategy) GetProtocol() string {
	return m.protocol
}

func (m *mockTraversalStrategy) GetName() string {
	return m.name
}

func (m *mockTraversalStrategy) EstimateSuccessRate(localNATType, remoteNATType discovery.NATType) float64 {
	return m.success
}

func TestCustomStrategyRegistration(t *testing.T) {
	factory := &StrategyFactory{
		strategies: make(map[StrategyType]TraversalStrategy),
	}

	// Register a custom strategy
	customStrategy := &mockTraversalStrategy{
		name:     "Custom Strategy",
		protocol: "custom",
		success:  1.0,
	}

	customType := StrategyType("custom-strategy")
	factory.registerStrategy(customType, customStrategy)

	// Retrieve the strategy
	retrieved, err := factory.GetStrategyByType(customType)
	assert.NoError(t, err)
	assert.Equal(t, customStrategy.GetName(), retrieved.GetName())

	// Strategy should be selected when it has highest success rate
	factory.registerStrategy(UDPHolePunching, &mockTraversalStrategy{
		name:     "UDP Mock",
		protocol: "udp",
		success:  0.5,
	})

	selectedStrategy := factory.SelectStrategy(discovery.NATFullCone, discovery.NATFullCone, "")
	assert.Equal(t, "Custom Strategy", selectedStrategy.GetName())
}
