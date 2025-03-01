package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bOguzhan/NATbypass/internal/stun"
	"github.com/bOguzhan/NATbypass/internal/utils"
)

func main() {
	serverFlag := flag.String("server", "stun.l.google.com:19302", "STUN server address")
	flag.Parse()

	// Create a logger for the STUN client
	logger := utils.NewLogger("stun-test", "info")

	// Create STUN client with all required parameters
	client := stun.NewClient(logger, *serverFlag, 5, 3)

	publicAddr, err := client.DiscoverPublicAddress()
	if err != nil {
		log.Fatalf("Failed to discover public address: %v", err)
	}

	fmt.Printf("Your public IP address is: %s\n", publicAddr.String())

	// Keep the connection open for demonstration
	fmt.Println("Press Ctrl+C to exit...")
	for {
		time.Sleep(time.Second)
	}
}
