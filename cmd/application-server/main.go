// cmd/application-server/main.go
package main

import (
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	log.Info("Starting Application Server...")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start UDP server
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to start UDP server: %v", err)
	}
	defer conn.Close()

	log.Infof("Application Server listening on UDP port %s", port)

	// Basic UDP packet handling loop
	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Errorf("Error reading UDP packet: %v", err)
			continue
		}

		log.Infof("Received %d bytes from %s", n, clientAddr.String())

		// Echo the data back for now
		if _, err := conn.WriteToUDP(buffer[:n], clientAddr); err != nil {
			log.Errorf("Error sending UDP packet: %v", err)
		}
	}
}
