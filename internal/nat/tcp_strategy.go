package nat

import (
	"context"
	"net"
	"time"

	"github.com/bOguzhan/NATbypass/internal/discovery"
)

// TCPSimultaneousOpenStrategy implements TCP simultaneous open NAT traversal
type TCPSimultaneousOpenStrategy struct {
	// Configuration options for TCP simultaneous open
	connTimeout time.Duration
	retryDelay  time.Duration
	maxRetries  int
}

// newTCPSimultaneousOpenStrategy creates a new TCP simultaneous open strategy
func newTCPSimultaneousOpenStrategy() *TCPSimultaneousOpenStrategy {
	return &TCPSimultaneousOpenStrategy{
		connTimeout: 3 * time.Second,
		retryDelay:  500 * time.Millisecond,
		maxRetries:  5,
	}
}

// GetProtocol returns the network protocol used by this strategy
func (s *TCPSimultaneousOpenStrategy) GetProtocol() string {
	return "tcp"
}

// GetName returns the descriptive name of this strategy
func (s *TCPSimultaneousOpenStrategy) GetName() string {
	return "TCP Simultaneous Open"
}

// EstimateSuccessRate returns estimated success rate based on NAT types
func (s *TCPSimultaneousOpenStrategy) EstimateSuccessRate(localNATType, remoteNATType discovery.NATType) float64 {
	// Success rates based on combinations of NAT types

	// Full cone to full cone has decent success rate
	if localNATType == discovery.NATFullCone && remoteNATType == discovery.NATFullCone {
		return 0.95 // Increased from 0.80
	}

	// Full cone to restricted has moderate success
	if localNATType == discovery.NATFullCone || remoteNATType == discovery.NATFullCone {
		return 0.85 // Increased from 0.60
	}

	// Address restricted to address restricted has lower success than UDP
	if localNATType == discovery.NATAddressRestrictedCone && remoteNATType == discovery.NATAddressRestrictedCone {
		return 0.70 // Increased from 0.50
	}

	// Port restricted NATs have low success with TCP simultaneous open
	if localNATType == discovery.NATPortRestrictedCone || remoteNATType == discovery.NATPortRestrictedCone {
		return 0.40 // Increased from 0.30
	}

	// Symmetric NAT rarely works with TCP simultaneous open
	if localNATType == discovery.NATSymmetric || remoteNATType == discovery.NATSymmetric {
		return 0.05 // Keep as is
	}

	// Default case - low-moderate chance
	return 0.35 // Increased from 0.25
}

// EstablishConnection attempts to establish a direct peer-to-peer connection
func (s *TCPSimultaneousOpenStrategy) EstablishConnection(
	ctx context.Context,
	localAddr,
	remoteAddr *net.UDPAddr,
) (net.Conn, error) {
	// This is a simplified implementation
	// TCP simultaneous open requires precise timing and multiple connection attempts
	// For UDP addresses, we'd need to convert to TCP addresses

	// Convert UDP addresses to TCP addresses
	localTCPAddr := &net.TCPAddr{
		IP:   localAddr.IP,
		Port: localAddr.Port,
		Zone: localAddr.Zone,
	}
	remoteTCPAddr := &net.TCPAddr{
		IP:   remoteAddr.IP,
		Port: remoteAddr.Port,
		Zone: remoteAddr.Zone,
	}

	// Create a TCP listener bound to the local address
	listener, err := net.ListenTCP("tcp", localTCPAddr)
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	// In a real implementation, we would:
	// 1. Start a goroutine to listen for incoming connections
	// 2. Attempt to connect to the remote address
	// 3. Handle race conditions and retry mechanisms
	// 4. Implement timing strategies for synchronizing connection attempts

	// For now, we'll just attempt to connect
	// This is oversimplified and would need proper implementation

	dialer := net.Dialer{
		Timeout:   s.connTimeout,
		LocalAddr: localTCPAddr,
	}

	conn, err := dialer.DialContext(ctx, "tcp", remoteTCPAddr.String())
	if err != nil {
		// In real implementation: retry with different timing, handle errors
		return nil, err
	}

	return conn, nil
}
