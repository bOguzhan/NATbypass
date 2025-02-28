// pkg/networking/discovery.go
package networking

import (
	"context"
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

	// Create context with timeout for the whole operation
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(config.TimeoutSeconds)*time.Second,
	)
	defer cancel()

	// Try multiple times if needed
	var lastErr error
	for i := 0; i < config.RetryCount; i++ {
		// Create a STUN client
		c, err := stun.Dial("udp", config.Server)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect to STUN server: %w", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Don't defer here as we want to close before the next iteration
		// Create a channel to receive the STUN response
		responseChan := make(chan interface{}, 1)

		// Create message with transaction ID
		message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

		// Send the STUN request with a callback
		if err := c.Do(message, func(res stun.Event) {
			if res.Error != nil {
				responseChan <- res.Error
				return
			}

			var xorAddr stun.XORMappedAddress
			if err := xorAddr.GetFrom(res.Message); err != nil {
				responseChan <- err
				return
			}

			responseChan <- &PublicAddress{
				IP:   xorAddr.IP,
				Port: xorAddr.Port,
			}
		}); err != nil {
			lastErr = err
			c.Close()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Wait for the response or timeout
		select {
		case resp := <-responseChan:
			c.Close() // Close connection after getting response

			// Check if we got an error
			if err, ok := resp.(error); ok {
				lastErr = err
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Check if we got an address
			if addr, ok := resp.(*PublicAddress); ok {
				return addr, nil
			}

		case <-ctx.Done():
			// Timeout occurred
			c.Close()
			lastErr = ctx.Err()
			break
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
