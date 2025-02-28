// pkg/networking/discovery.go
package networking

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun"
)

// PublicAddress represents public IP and port discovered via STUN
type PublicAddress struct {
	IP   net.IP
	Port int
}

// STUNConfig contains configuration for STUN operations
type STUNConfig struct {
	Server         string
	TimeoutSeconds int
	RetryCount     int
}

// DiscoverPublicAddressWithConfig uses a STUN server to discover the public IP and port with configuration
func DiscoverPublicAddressWithConfig(config STUNConfig) (*PublicAddress, error) {
	// Set default values if not provided
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 5
	}

	if config.RetryCount <= 0 {
		config.RetryCount = 3
	}

	// Set timeout
	timeout := time.Duration(config.TimeoutSeconds) * time.Second

	// Create a STUN client with timeout
	c, err := stun.Dial("udp", config.Server)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to STUN server: %w", err)
	}
	defer c.Close()

	// Set timeout on the connection
	c.SetDeadline(time.Now().Add(timeout))

	// Try multiple times if needed
	var lastErr error
	for i := 0; i < config.RetryCount; i++ {
		message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

		var xorAddr stun.XORMappedAddress
		err = c.Do(message, func(res stun.Event) {
			if res.Error != nil {
				lastErr = res.Error
				return
			}

			if err := xorAddr.GetFrom(res.Message); err != nil {
				lastErr = err
				return
			}
		})

		if err == nil && lastErr == nil {
			return &PublicAddress{
				IP:   xorAddr.IP,
				Port: xorAddr.Port,
			}, nil
		}

		// If we failed, sleep briefly before retrying
		if i < config.RetryCount-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to get STUN binding after %d attempts: %w",
			config.RetryCount, lastErr)
	}

	return nil, fmt.Errorf("failed to get STUN binding after %d attempts",
		config.RetryCount)
}

// DiscoverPublicAddress uses a STUN server to discover the public IP and port
// This is a simplified version that calls the configurable version with defaults
func DiscoverPublicAddress(stunServer string) (*PublicAddress, error) {
	config := STUNConfig{
		Server:         stunServer,
		TimeoutSeconds: 5,
		RetryCount:     3,
	}
	return DiscoverPublicAddressWithConfig(config)
}
