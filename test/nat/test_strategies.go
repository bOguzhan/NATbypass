// File: test/nat/test_strategies.go

package nat_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/bOguzhan/NATbypass/internal/discovery"
)

// MockStrategy implements a basic NAT traversal strategy for testing
type MockStrategy struct {
	name           string
	protocol       string
	successRateMap map[string]float64
}

func (s *MockStrategy) GetName() string {
	return s.name
}

func (s *MockStrategy) GetProtocol() string {
	return s.protocol
}

func (s *MockStrategy) EstablishConnection(ctx context.Context, localAddr, remoteAddr *net.UDPAddr) (net.Conn, error) {
	if s.protocol == "udp" {
		return s.establishUDP(ctx, localAddr, remoteAddr)
	}
	return s.establishTCP(ctx, localAddr, remoteAddr)
}

func (s *MockStrategy) establishUDP(ctx context.Context, localAddr, remoteAddr *net.UDPAddr) (net.Conn, error) {
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on local address: %w", err)
	}

	// Send a test packet to establish mapping
	_, err = conn.WriteToUDP([]byte("TEST_PACKET"), remoteAddr)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send test packet: %w", err)
	}

	return &UDPConnWrapper{
		UDPConn:    conn,
		remoteAddr: remoteAddr,
	}, nil
}

func (s *MockStrategy) establishTCP(ctx context.Context, localAddr, remoteAddr *net.UDPAddr) (net.Conn, error) {
	// Convert UDP addresses to TCP addresses
	localTCPAddr := &net.TCPAddr{
		IP:   localAddr.IP,
		Port: localAddr.Port,
	}

	remoteTCPAddr := &net.TCPAddr{
		IP:   remoteAddr.IP,
		Port: remoteAddr.Port,
	}

	// Try to connect
	dialer := net.Dialer{
		LocalAddr: localTCPAddr,
		Timeout:   5 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", remoteTCPAddr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to establish TCP connection: %w", err)
	}

	return conn, nil
}

func (s *MockStrategy) EstimateSuccessRate(localNATType, remoteNATType discovery.NATType) float64 {
	key := string(localNATType) + "-" + string(remoteNATType)
	if rate, exists := s.successRateMap[key]; exists {
		return rate
	}
	return 0.5 // Default: 50% chance
}

// UDPConnWrapper wraps a UDP connection to simulate a connection-oriented behavior
type UDPConnWrapper struct {
	*net.UDPConn
	remoteAddr net.Addr
}

func (c *UDPConnWrapper) Read(b []byte) (int, error) {
	n, addr, err := c.UDPConn.ReadFromUDP(b)
	if err != nil {
		return 0, err
	}

	// For testing, we accept packets from any address
	// In a real implementation, we would check addr against remoteAddr
	return n, nil
}

func (c *UDPConnWrapper) Write(b []byte) (int, error) {
	return c.UDPConn.WriteToUDP(b, c.remoteAddr.(*net.UDPAddr))
}

func (c *UDPConnWrapper) RemoteAddr() net.Addr {
	return c.remoteAddr
}
