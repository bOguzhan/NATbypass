package main

import (
	"fmt"
	"log"

	"github.com/bOguzhan/NATbypass/pkg/networking"
)

func main() {
	log.Println("Testing STUN discovery...")

	// Use Google's public STUN server
	stunServer := "stun.l.google.com:19302"

	addr, err := networking.DiscoverPublicAddress(stunServer)
	if err != nil {
		log.Fatalf("Failed to discover public address: %v", err)
	}

	fmt.Printf("Your public IP is: %s\n", addr.IP.String())
	fmt.Printf("Your public port is: %d\n", addr.Port)
}
