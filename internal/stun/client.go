// internal/stun/client.go
package stun

import (
	"fmt"
	"net"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/pion/stun"
)

// Client represents a STUN client
type Client struct {
	Logger  *utils.Logger
	Server  string
	Timeout int // in seconds
	Retries int
}

// NewClient creates a new STUN client
func NewClient(logger *utils.Logger, server string, timeout int, retries int) *Client {
	return &Client{
		Logger:  logger,
		Server:  server,
		Timeout: timeout,
		Retries: retries,
	}
}

// DiscoverPublicAddress discovers the public IP and port using STUN
func (c *Client) DiscoverPublicAddress() (*net.UDPAddr, error) {
	c.Logger.Debugf("Discovering public address using STUN server: %s", c.Server)

	// Create a connection to the STUN server
	conn, err := net.Dial("udp4", c.Server)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to STUN server: %w", err)
	}
	defer conn.Close()

	// Set a timeout for the request
	conn.SetDeadline(time.Now().Add(time.Duration(c.Timeout) * time.Second))

	// Create a STUN message
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// Send the message
	if _, err := conn.Write(message.Raw); err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %w", err)
	}

	// Create a buffer to receive the response
	buffer := make([]byte, 1024)

	// Receive the response
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to receive STUN response: %w", err)
	}

	// Parse the response
	response := &stun.Message{Raw: buffer[:n]}
	if err := response.Decode(); err != nil {
		return nil, fmt.Errorf("failed to decode STUN message: %w", err)
	}

	// Extract the XOR-MAPPED-ADDRESS attribute
	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(response); err != nil {
		return nil, fmt.Errorf("failed to extract XOR-MAPPED-ADDRESS: %w", err)
	}

	c.Logger.Infof("Discovered public address: %s", xorAddr.String())

	return &net.UDPAddr{
		IP:   xorAddr.IP,
		Port: xorAddr.Port,
	}, nil
}

// DetermineNATType determines the type of NAT based on RFC 3489
// This is a simplified version - a complete implementation would require
// multiple STUN tests as described in the RFC
func (c *Client) DetermineNATType() (string, error) {
	// TODO: Implement proper NAT type detection based on RFC 3489 or RFC 5780
	// For now, we're just returning "Unknown"

	c.Logger.Debug("NAT type detection not fully implemented yet")
	return "Unknown", nil
}
