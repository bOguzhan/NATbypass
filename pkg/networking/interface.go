package networking

import (
	"context"
	"net"
)

// ConnectionType represents the type of connection (TCP or UDP)
type ConnectionType string

const (
	// TCP connection type
	TCP ConnectionType = "tcp"
	// UDP connection type
	UDP ConnectionType = "udp"
)

// Packet represents a protocol-agnostic network packet
type Packet struct {
	Data       []byte
	SourceAddr net.Addr
	TargetAddr net.Addr
	ConnType   ConnectionType
	Metadata   map[string]interface{}
}

// NetworkHandler defines the interface for handling network connections
type NetworkHandler interface {
	// Initialize sets up the network handler with configuration
	Initialize(config map[string]interface{}) error

	// Start begins listening for incoming connections
	Start(ctx context.Context) error

	// Stop gracefully shuts down the network handler
	Stop() error

	// Send transmits a packet to the specified address
	Send(packet *Packet) error

	// RegisterReceiveCallback registers a callback for received packets
	RegisterReceiveCallback(callback func(*Packet) error)

	// GetListeningAddresses returns the addresses this handler is listening on
	GetListeningAddresses() []net.Addr
}

// ConnectionTracker defines the interface for tracking connections
type ConnectionTracker interface {
	// AddConnection registers a new connection
	AddConnection(sourceAddr net.Addr, targetAddr net.Addr, connType ConnectionType) (string, error)

	// GetConnection retrieves a connection by ID
	GetConnection(connID string) (sourceAddr net.Addr, targetAddr net.Addr, connType ConnectionType, err error)

	// UpdateConnectionState updates the state of a connection
	UpdateConnectionState(connID string, state string) error

	// RemoveConnection removes a tracked connection
	RemoveConnection(connID string) error

	// ListConnections returns all active connections
	ListConnections() ([]string, error)
}

// NATPunchStrategy defines the interface for NAT traversal strategies
type NATPunchStrategy interface {
	// CanHandle returns true if this strategy can handle the given NAT types
	CanHandle(sourceNATType, targetNATType string) bool

	// InitiatePunch begins the NAT hole-punching process
	InitiatePunch(ctx context.Context, sourceAddr, targetAddr net.Addr, connType ConnectionType) error

	// GetName returns the name of the strategy
	GetName() string

	// GetPriority returns the priority of the strategy (higher is better)
	GetPriority() int
}

// NetworkFactory is responsible for creating protocol-specific implementations
type NetworkFactory interface {
	// CreateHandler creates a new protocol-specific NetworkHandler
	CreateHandler(connType ConnectionType, config map[string]interface{}) (NetworkHandler, error)

	// CreateTracker creates a new ConnectionTracker
	CreateTracker(config map[string]interface{}) (ConnectionTracker, error)

	// GetNATPunchStrategy returns the appropriate NAT punch strategy
	GetNATPunchStrategy(connType ConnectionType, sourceNATType, targetNATType string) NATPunchStrategy
}
