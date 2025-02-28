package networking

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHandlerRegistration tests handler registration and creation
func TestHandlerRegistration(t *testing.T) {
	factory := NewNetworkFactory()

	// Register UDP handler
	factory.RegisterHandler(UDP, func(config map[string]interface{}) (NetworkHandler, error) {
		return newMockNetworkHandler(), nil
	})

	// Test creating valid handler
	handler, err := factory.CreateHandler(UDP, nil)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Test creating handler that returns error
	factory.RegisterHandler(TCP, func(config map[string]interface{}) (NetworkHandler, error) {
		return nil, errors.New("mock error")
	})

	handler, err = factory.CreateHandler(TCP, nil)
	assert.Error(t, err)
	assert.Nil(t, handler)
}

// TestStrategySelection tests the NAT punch strategy selection
func TestStrategySelection(t *testing.T) {
	factory := NewNetworkFactory()

	// Create strategies with different priorities
	lowPriority := &mockNATPunchStrategy{
		name:     "LowPriority",
		priority: 1,
		canHandle: func(src, tgt string) bool {
			return true
		},
	}

	highPriority := &mockNATPunchStrategy{
		name:     "HighPriority",
		priority: 10,
		canHandle: func(src, tgt string) bool {
			return true
		},
	}

	incompatible := &mockNATPunchStrategy{
		name:     "Incompatible",
		priority: 100,
		canHandle: func(src, tgt string) bool {
			return false
		},
	}

	// Register strategies for UDP
	factory.RegisterNATPunchStrategy(UDP, lowPriority)
	factory.RegisterNATPunchStrategy(UDP, highPriority)
	factory.RegisterNATPunchStrategy(UDP, incompatible)

	// Should select high priority strategy
	strategy := factory.GetNATPunchStrategy(UDP, "FullCone", "FullCone")
	assert.NotNil(t, strategy)
	assert.Equal(t, "HighPriority", strategy.GetName())

	// Test no compatible strategies
	noMatchStrategy := &mockNATPunchStrategy{
		name:     "NoMatch",
		priority: 1,
		canHandle: func(src, tgt string) bool {
			return false
		},
	}

	factory2 := NewNetworkFactory()
	factory2.RegisterNATPunchStrategy(UDP, noMatchStrategy)

	strategy = factory2.GetNATPunchStrategy(UDP, "FullCone", "FullCone")
	assert.Nil(t, strategy)
}

// TestMultiStrategyNATTraversal tests multiple NAT traversal strategies
func TestMultiStrategyNATTraversal(t *testing.T) {
	factory := NewNetworkFactory()

	// Create strategies with different NAT type handling
	fullConeStrategy := &mockNATPunchStrategy{
		name:     "FullConeStrategy",
		priority: 10,
		canHandle: func(src, tgt string) bool {
			return src == "FullCone" && tgt == "FullCone"
		},
	}

	restrictedStrategy := &mockNATPunchStrategy{
		name:     "RestrictedStrategy",
		priority: 5,
		canHandle: func(src, tgt string) bool {
			return src == "Restricted" || tgt == "Restricted"
		},
	}

	// Register strategies
	factory.RegisterNATPunchStrategy(UDP, fullConeStrategy)
	factory.RegisterNATPunchStrategy(UDP, restrictedStrategy)

	// Test strategy selection based on NAT types
	strategy := factory.GetNATPunchStrategy(UDP, "FullCone", "FullCone")
	assert.NotNil(t, strategy)
	assert.Equal(t, "FullConeStrategy", strategy.GetName())

	strategy = factory.GetNATPunchStrategy(UDP, "Restricted", "PortRestricted")
	assert.NotNil(t, strategy)
	assert.Equal(t, "RestrictedStrategy", strategy.GetName())

	// No strategy should match for Symmetric
	strategy = factory.GetNATPunchStrategy(UDP, "Symmetric", "Symmetric")
	assert.Nil(t, strategy)
}
