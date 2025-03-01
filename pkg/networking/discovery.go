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

// STUNConfig represents configuration options for STUN client
type STUNConfig struct {
	Server         string
	TimeoutSeconds int // Changed from Timeout to TimeoutSeconds
	RetryCount     int
}

// Default settings for STUN discovery
var DefaultSTUNConfig = STUNConfig{
	Server:         "stun.l.google.com:19302",
	TimeoutSeconds: 5, // Changed to match the field name
	RetryCount:     3,
}

// DiscoverPublicAddress uses a STUN server to discover the public IP and port
func DiscoverPublicAddress(stunServer string) (*PublicAddress, error) {
	config := DefaultSTUNConfig
	config.Server = stunServer
	return DiscoverPublicAddressWithConfig(config)
}

// DiscoverPublicAddressWithConfig discovers the public IP and port using a STUN server with custom config
func DiscoverPublicAddressWithConfig(config STUNConfig) (*PublicAddress, error) {
	c, err := stun.Dial("udp", config.Server)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to STUN server: %w", err)
	}
	defer c.Close()

	// Create a timeout duration from the seconds value
	c.SetRTO(time.Duration(config.TimeoutSeconds) * time.Second)

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	var xorAddr stun.XORMappedAddress
	if err = c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			err = res.Error
			return
		}

		if err = xorAddr.GetFrom(res.Message); err != nil {
			return
		}
	}); err != nil {
		return nil, fmt.Errorf("failed to get STUN binding: %w", err)
	}

	return &PublicAddress{
		IP:   xorAddr.IP,
		Port: xorAddr.Port,
	}, nil
}
