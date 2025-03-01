package nat

import (
	"context"
	"errors"
	"net"
	"sort"

	"github.com/bOguzhan/NATbypass/internal/discovery"
)

// TraversalStrategy defines the interface for different NAT traversal techniques
type TraversalStrategy interface {
	// EstablishConnection attempts to establish a direct peer-to-peer connection
	EstablishConnection(ctx context.Context, localAddr, remoteAddr *net.UDPAddr) (net.Conn, error)

	// GetProtocol returns the network protocol used by this strategy
	GetProtocol() string

	// GetName returns the descriptive name of this strategy
	GetName() string

	// EstimateSuccessRate returns an estimated success rate for this strategy based on NAT types
	EstimateSuccessRate(localNATType, remoteNATType discovery.NATType) float64
}

// StrategyType defines the type of NAT traversal strategy
type StrategyType string

const (
	// UDPHolePunching represents standard UDP hole punching technique
	UDPHolePunching StrategyType = "udp-hole-punching"

	// TCPSimultaneousOpen represents TCP simultaneous open technique
	TCPSimultaneousOpen StrategyType = "tcp-simultaneous-open"

	// UDPRelaying represents UDP relaying via a TURN server (fallback option)
	UDPRelaying StrategyType = "udp-relaying"

	// TCPRelaying represents TCP relaying via a TURN server (fallback option)
	TCPRelaying StrategyType = "tcp-relaying"
)

// StrategySelector helps select the optimal traversal strategy based on NAT types
type StrategySelector interface {
	// SelectStrategy selects the optimal strategy based on NAT types and preferences
	SelectStrategy(localNATType, remoteNATType discovery.NATType, preferredProtocol string) TraversalStrategy

	// GetStrategyByType returns a specific strategy by its type
	GetStrategyByType(strategyType StrategyType) (TraversalStrategy, error)

	// GetAvailableStrategies returns all available strategies
	GetAvailableStrategies() []TraversalStrategy
}

// Errors related to strategy factory
var (
	ErrStrategyNotFound = errors.New("requested traversal strategy not found")
	ErrNoValidStrategy  = errors.New("no valid traversal strategy available")
)

// StrategyFactory creates and manages NAT traversal strategies
type StrategyFactory struct {
	strategies map[StrategyType]TraversalStrategy
}

// NewStrategyFactory creates a new strategy factory with all available strategies
func NewStrategyFactory() *StrategyFactory {
	factory := &StrategyFactory{
		strategies: make(map[StrategyType]TraversalStrategy),
	}

	// Register all available strategies
	factory.registerStrategy(UDPHolePunching, newUDPHolePunchingStrategy())
	factory.registerStrategy(TCPSimultaneousOpen, newTCPSimultaneousOpenStrategy())
	factory.registerStrategy(UDPRelaying, newUDPRelayingStrategy())
	factory.registerStrategy(TCPRelaying, newTCPRelayingStrategy())

	return factory
}

// registerStrategy registers a strategy with the factory
func (f *StrategyFactory) registerStrategy(strategyType StrategyType, strategy TraversalStrategy) {
	f.strategies[strategyType] = strategy
}

// GetStrategyByType retrieves a strategy by its type
func (f *StrategyFactory) GetStrategyByType(strategyType StrategyType) (TraversalStrategy, error) {
	strategy, exists := f.strategies[strategyType]
	if !exists {
		return nil, ErrStrategyNotFound
	}
	return strategy, nil
}

// GetAvailableStrategies returns all registered strategies
func (f *StrategyFactory) GetAvailableStrategies() []TraversalStrategy {
	strategies := make([]TraversalStrategy, 0, len(f.strategies))
	for _, strategy := range f.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// SelectStrategy chooses the optimal traversal strategy based on NAT types and preferences
func (f *StrategyFactory) SelectStrategy(localNATType, remoteNATType discovery.NATType, preferredProtocol string) TraversalStrategy {
	candidates := make([]struct {
		strategy    TraversalStrategy
		successRate float64
	}, 0, len(f.strategies))

	// Evaluate each strategy's success rate for the given NAT types
	for _, strategy := range f.strategies {
		// If preferred protocol is specified, filter by protocol
		if preferredProtocol != "" && strategy.GetProtocol() != preferredProtocol {
			continue
		}

		successRate := strategy.EstimateSuccessRate(localNATType, remoteNATType)
		candidates = append(candidates, struct {
			strategy    TraversalStrategy
			successRate float64
		}{strategy, successRate})
	}

	// Sort by success rate (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].successRate > candidates[j].successRate
	})

	// Return the strategy with highest success rate, or nil if none available
	if len(candidates) > 0 {
		return candidates[0].strategy
	}

	// If no preferred protocol match was found, return the best strategy overall
	if preferredProtocol != "" {
		return f.SelectStrategy(localNATType, remoteNATType, "")
	}

	// Should never happen if strategies are properly registered
	return nil
}
