// pkg/networking/utils.go
package networking

import (
	"fmt"
	"net"
	"time"
)

// GetLocalIP returns the non-loopback local IP of the host
func GetLocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, addr := range addrs {
		// Check if the address is IP network
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}

	return nil, fmt.Errorf("no suitable local IP address found")
}

// CheckUDPPort tests if a UDP port is available for listening
func CheckUDPPort(port int) bool {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// CreateUDPListener creates a UDP listener on the specified address
func CreateUDPListener(address string, timeout time.Duration) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", address, err)
	}

	// Set deadline if timeout is provided
	var conn *net.UDPConn
	if timeout > 0 {
		// Create a connection with timeout
		dialer := net.Dialer{Timeout: timeout}
		baseConn, err := dialer.Dial("udp", address)
		if err != nil {
			return nil, fmt.Errorf("failed to create UDP connection: %w", err)
		}

		if udpConn, ok := baseConn.(*net.UDPConn); ok {
			conn = udpConn
		} else {
			baseConn.Close()
			return nil, fmt.Errorf("failed to convert to UDP connection type")
		}
	} else {
		// Create standard connection without timeout
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to listen on UDP %s: %w", address, err)
		}
	}

	return conn, nil
}

// SendUDPPacket sends a UDP packet to the specified address
func SendUDPPacket(conn *net.UDPConn, targetAddr *net.UDPAddr, data []byte) error {
	_, err := conn.WriteToUDP(data, targetAddr)
	if err != nil {
		return fmt.Errorf("failed to send UDP packet: %w", err)
	}
	return nil
}
