package main

import (
	"fmt"
	"net"

	"github.com/pion/stun"
)

// PublicAddress represents public IP and port discovered via STUN
type PublicAddress struct {
	IP   net.IP
	Port int
}

// DiscoverPublicAddress uses a STUN server to discover the public IP and port
func DiscoverPublicAddress(stunServer string) (*PublicAddress, error) {
	c, err := stun.Dial("udp", stunServer)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to STUN server: %w", err)
	}
	defer c.Close()

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
