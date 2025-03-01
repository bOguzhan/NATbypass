package main

import (
	"fmt"
	"log"
	"net"

	"github.com/pion/stun"
)

func testMain() {
	log.Println("Testing STUN discovery...")

	// Use Google's public STUN server
	stunServer := "stun.l.google.com:19302"

	c, err := stun.Dial("udp", stunServer)
	if err != nil {
		log.Fatalf("Failed to dial STUN server: %v", err)
	}
	defer c.Close()

	var ip net.IP
	var port int

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	err = c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			log.Fatalf("Failed to process STUN response: %v", res.Error)
		}
		if res.Message.Type.Class == stun.ClassSuccessResponse {
			var xorAddr stun.XORMappedAddress
			if err := xorAddr.GetFrom(res.Message); err != nil {
				log.Fatalf("Failed to get XOR mapped address: %v", err)
			}
			ip = xorAddr.IP
			port = xorAddr.Port
		}
	})
	if err != nil {
		log.Fatalf("Failed to discover public address: %v", err)
	}

	fmt.Printf("Your public IP is: %s\n", ip.String())
	fmt.Printf("Your public port is: %d\n", port)
}
