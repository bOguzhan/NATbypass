package nat

import (
	"context"
	"net"

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
