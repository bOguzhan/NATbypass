// File: test/nat/test_client.go

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/bOguzhan/NATbypass/internal/discovery"
	"github.com/bOguzhan/NATbypass/internal/nat"
)

func main() {
	// Get environment variables
	natTypeStr := os.Getenv("NAT_TYPE")
	protocol := os.Getenv("PROTOCOL")
	peerType := os.Getenv("PEER_TYPE")

	// Convert NAT type string to discovery.NATType
	var natType discovery.NATType
	switch natTypeStr {
	case "full-cone":
		natType = discovery.NATFullCone
	case "address-restricted-cone":
		natType = discovery.NATAddressRestrictedCone
	case "port-restricted-cone":
		natType = discovery.NATPortRestrictedCone
	case "symmetric":
		natType = discovery.NATSymmetric
	default:
		log.Fatalf("Unknown NAT type: %s", natTypeStr)
	}

	// Determine local and remote addresses
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		log.Fatalf("Failed to resolve local address: %v", err)
	}

	var remoteAddr *net.UDPAddr
	if peerType == "initiator" {
		remoteAddr, err = net.ResolveUDPAddr("udp", "peer2:12345")
	} else {
		remoteAddr, err = net.ResolveUDPAddr("udp", "peer1:12345")
	}
	if err != nil {
		log.Fatalf("Failed to resolve remote address: %v", err)
	}

	// Create NAT traversal strategy factory
	factory := nat.NewStrategyFactory()

	// Select appropriate strategy
	var remoteNatType discovery.NATType
	switch peerType {
	case "initiator":
		fmt.Println("Running as initiator")
		remoteNatType = getRemoteNATType("peer2:12345")
	case "responder":
		fmt.Println("Running as responder")
		remoteNatType = getRemoteNATType("peer1:12345")
	default:
		log.Fatalf("Unknown peer type: %s", peerType)
	}

	strategy := factory.SelectStrategy(natType, remoteNatType, protocol)
	if strategy == nil {
		log.Fatalf("No suitable strategy found")
	}

	fmt.Printf("Selected strategy: %s\n", strategy.GetName())

	// Attempt to establish connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := strategy.EstablishConnection(ctx, localAddr, remoteAddr)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	fmt.Printf("Connection established! Local: %s, Remote: %s\n",
		conn.LocalAddr().String(), conn.RemoteAddr().String())

	// Send test data
	if peerType == "initiator" {
		_, err = conn.Write([]byte("Hello from initiator"))
		if err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}

		// Wait for response
		buf := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Failed to receive data: %v", err)
		}

		fmt.Printf("Received: %s\n", string(buf[:n]))
	} else {
		// Wait for data
		buf := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Failed to receive data: %v", err)
		}

		fmt.Printf("Received: %s\n", string(buf[:n]))

		// Send response
		_, err = conn.Write([]byte("Hello from responder"))
		if err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}
	}

	fmt.Println("Test completed successfully!")
}

// Mock function to get remote NAT type
// In a real implementation, this would use the signaling channel
func getRemoteNATType(address string) discovery.NATType {
	// For the test, we're using environment variables
	switch os.Getenv("NAT_TYPE") {
	case "full-cone":
		return discovery.NATFullCone
	case "address-restricted-cone":
		return discovery.NATAddressRestrictedCone
	case "port-restricted-cone":
		return discovery.NATPortRestrictedCone
	case "symmetric":
		return discovery.NATSymmetric
	default:
		return discovery.NATUnknown
	}
}
