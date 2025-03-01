// File: test/nat/mock_strategies.go

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
	successRateMap map[string]map[string]float64
}

func NewMockStrategy(name, protocol string) *MockStrategy {
	return &MockStrategy{
		name:           name,
		protocol:       protocol,
		successRateMap: make(map[string]map[string]float64),
	}
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

func (s *MockStrategy) EstimateSuccessRate(localNATType, remoteNATType discovery.NATType) float64 {
	local := string(localNATType)
	remote := string(remoteNATType)

	if localMap, ok := s.successRateMap[local]; ok {
		if rate, ok := localMap[remote]; ok {
			return rate
		}
	}

	// Default success rates based on standard expectations
	if s.protocol == "udp" {
		if localNATType == discovery.NATSymmetric && remoteNATType == discovery.NATSymmetric {
			return 0.1 // UDP difficult with symmetric-symmetric
		}
		if localNATType == discovery.NATSymmetric || remoteNATType == discovery.NATSymmetric {
			return 0.3 // UDP with one symmetric side
		}
		return 0.9 // UDP generally works well
	} else {
		// TCP generally harder
		if localNATType == discovery.NATSymmetric || remoteNATType == discovery.NATSymmetric {
			return 0.05 // TCP very difficult with symmetric
		}
		return 0.7 // TCP generally works but less reliable than UDP
	}
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
	// For TCP, convert UDP addresses to TCP
	localTCPAddr := &net.TCPAddr{
		IP:   localAddr.IP,
		Port: localAddr.Port,
	}
	remoteTCPAddr := &net.TCPAddr{
		IP:   remoteAddr.IP,
		Port: remoteAddr.Port,
	}

	// Create a TCP connection
	d := net.Dialer{
		LocalAddr: localTCPAddr,
		Timeout:   5 * time.Second,
	}

	// Try to connect
	conn, err := d.DialContext(ctx, "tcp", remoteTCPAddr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to establish TCP connection: %w", err)
	}

	return conn, nil
}

// UDPConnWrapper wraps a UDP connection to simulate connection-oriented behavior
type UDPConnWrapper struct {
	*net.UDPConn
	remoteAddr net.Addr
}

func (c *UDPConnWrapper) Read(b []byte) (int, error) {
	n, addr, err := c.UDPConn.ReadFromUDP(b)
	if err != nil {
		return 0, err
	}
	// Only accept packets from the established remote address
	if addr.String() != c.remoteAddr.String() {
		return 0, fmt.Errorf("received packet from unexpected address: %s", addr.String())
	}
	return n, nil
}

func (c *UDPConnWrapper) Write(b []byte) (int, error) {
	return c.UDPConn.WriteToUDP(b, c.remoteAddr.(*net.UDPAddr))
}

func (c *UDPConnWrapper) RemoteAddr() net.Addr {
	return c.remoteAddr
}
