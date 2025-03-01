// cmd/application-server/main.go
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/bOguzhan/NATbypass/internal/config"
)

func main() {
	// Determine config path - default to configs/config.yaml
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Configure logger
	log := config.ConfigureLogger(cfg.Servers.Application.LogLevel)
	log.Info("Starting Application Server...")

	// Start UDP server
	serverAddr := fmt.Sprintf("%s:%d",
		cfg.Servers.Application.Host,
		cfg.Servers.Application.Port)

	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to start UDP server: %v", err)
	}
	defer conn.Close()

	log.Infof("Application Server listening on UDP %s", serverAddr)

	// Initialize and start TCP server
	tcpServer := nat.NewTCPServer(&config.TCPServerConfig{
		Host:              appConfig.TCP.Host,
		Port:              appConfig.TCP.Port,
		ConnectionTimeout: appConfig.TCP.ConnectionTimeout,
		MaxConnections:    appConfig.TCP.MaxConnections,
		BufferSize:        appConfig.TCP.BufferSize,
	}, logger)

	if err := tcpServer.Start(ctx); err != nil {
		logger.Fatalf("Failed to start TCP server: %v", err)
	}

	// Add TCP server to graceful shutdown
	defer func() {
		if err := tcpServer.Stop(); err != nil {
			logger.Errorf("Error stopping TCP server: %v", err)
		}
	}()

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
