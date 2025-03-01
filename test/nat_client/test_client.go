// File: test/nat/test_client.go

package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bOguzhan/NATbypass/internal/discovery"
	"github.com/bOguzhan/NATbypass/internal/nat"
)

const (
	testMessage       = "HELLO_FROM_NAT_TEST"
	testResponse      = "ACK_FROM_NAT_TEST"
	signalAddr        = "localhost:8080" // In a real scenario, this would be a publicly accessible server
	connectionRetries = 5
	retryDelay        = 1 * time.Second
	testTimeout       = 10 * time.Second
)

func main() {
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("NAT traversal test client starting")

	// Get environment variables for test configuration
	natTypeStr := os.Getenv("NAT_TYPE")
	protocol := os.Getenv("PROTOCOL")
	peerType := os.Getenv("PEER_TYPE")
	remoteHost := "peer2"
	if peerType == "responder" {
		remoteHost = "peer1"
	}

	log.Printf("Configuration: NAT Type=%s, Protocol=%s, Role=%s",
		natTypeStr, protocol, peerType)

	// Convert NAT type string to discovery.NATType
	var natType discovery.NATType
	switch strings.ToLower(natTypeStr) {
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

	// Set up context with interrupt handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt, shutting down...")
		cancel()
	}()

	// Simulated NATed endpoints
	var localAddr, remoteAddr *net.UDPAddr
	var err error

	// In a real test scenario, these would be determined dynamically
	// For the Docker test, we use predefined ports
	if peerType == "initiator" {
		localAddr, err = net.ResolveUDPAddr("udp", "0.0.0.0:12345")
		remoteAddr, err = net.ResolveUDPAddr("udp", remoteHost+":12346")
	} else {
		localAddr, err = net.ResolveUDPAddr("udp", "0.0.0.0:12346")
		remoteAddr, err = net.ResolveUDPAddr("udp", remoteHost+":12345")
	}
	if err != nil {
		log.Fatalf("Failed to resolve addresses: %v", err)
	}

	// Create NAT traversal factory
	factory := nat.NewStrategyFactory()

	// Simulate remote peer's NAT type (in real scenario would come via signaling)
	var remoteNATType discovery.NATType
	switch peerType {
	case "initiator":
		// Get remote NAT type from environment or use the same as local
		remotePeerNATType := os.Getenv("REMOTE_NAT_TYPE")
		if remotePeerNATType == "" {
			remoteNATType = natType // Default to same as local for testing
		} else {
			switch strings.ToLower(remotePeerNATType) {
			case "full-cone":
				remoteNATType = discovery.NATFullCone
			case "address-restricted-cone":
				remoteNATType = discovery.NATAddressRestrictedCone
			case "port-restricted-cone":
				remoteNATType = discovery.NATPortRestrictedCone
			case "symmetric":
				remoteNATType = discovery.NATSymmetric
			default:
				remoteNATType = natType
			}
		}
	case "responder":
		// Get remote NAT type from environment or use the same as local
		remotePeerNATType := os.Getenv("REMOTE_NAT_TYPE")
		if remotePeerNATType == "" {
			remoteNATType = natType // Default to same as local for testing
		} else {
			switch strings.ToLower(remotePeerNATType) {
			case "full-cone":
				remoteNATType = discovery.NATFullCone
			case "address-restricted-cone":
				remoteNATType = discovery.NATAddressRestrictedCone
			case "port-restricted-cone":
				remoteNATType = discovery.NATPortRestrictedCone
			case "symmetric":
				remoteNATType = discovery.NATSymmetric
			default:
				remoteNATType = natType
			}
		}
	}

	// Select appropriate strategy
	log.Printf("Local NAT type: %s, Remote NAT type: %s", natType, remoteNATType)
	strategy := factory.SelectStrategy(natType, remoteNATType, protocol)
	if strategy == nil {
		log.Fatalf("No suitable strategy found for the given NAT types")
	}
	log.Printf("Selected strategy: %s", strategy.GetName())

	// Set up timeout context for connection attempt
	connCtx, connCancel := context.WithTimeout(ctx, testTimeout)
	defer connCancel()

	// Attempt to establish connection with retries
	var conn net.Conn
	for i := 0; i < connectionRetries; i++ {
		log.Printf("Connection attempt %d/%d", i+1, connectionRetries)

		conn, err = strategy.EstablishConnection(connCtx, localAddr, remoteAddr)
		if err == nil {
			break
		}

		log.Printf("Connection attempt failed: %v", err)
		if i < connectionRetries-1 {
			log.Printf("Retrying in %v", retryDelay)
			select {
			case <-time.After(retryDelay):
				// Continue to next attempt
			case <-connCtx.Done():
				log.Fatalf("Connection timeout or context canceled")
				return
			}
		}
	}

	if err != nil {
		log.Fatalf("All connection attempts failed: %v", err)
	}

	defer conn.Close()
	log.Printf("Connection established! Local: %s, Remote: %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String())

	// Communication test based on peer type
	if peerType == "initiator" {
		// Send message
		log.Printf("Sending test message: %s", testMessage)
		_, err = conn.Write([]byte(testMessage))
		if err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}

		// Wait for response
		buf := make([]byte, 1024)
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Fatalf("Failed to set read deadline: %v", err)
		}

		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Failed to receive response: %v", err)
		}

		response := string(buf[:n])
		log.Printf("Received response: %s", response)

		if response != testResponse {
			log.Fatalf("Unexpected response: expected '%s', got '%s'", testResponse, response)
		}

		log.Println("NAT traversal test completed successfully!")

	} else { // responder
		// Wait for incoming message
		buf := make([]byte, 1024)
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Fatalf("Failed to set read deadline: %v", err)
		}

		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Failed to receive message: %v", err)
		}

		message := string(buf[:n])
		log.Printf("Received message: %s", message)

		if message != testMessage {
			log.Fatalf("Unexpected message: expected '%s', got '%s'", testMessage, message)
		}

		// Send response
		log.Printf("Sending response: %s", testResponse)
		_, err = conn.Write([]byte(testResponse))
		if err != nil {
			log.Fatalf("Failed to send response: %v", err)
		}

		log.Println("NAT traversal test completed successfully!")
	}

	// Exit with success code
	os.Exit(0)
}
