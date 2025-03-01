package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bOguzhan/NATbypass/internal/config"
	"github.com/bOguzhan/NATbypass/internal/nat"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create TCP server configuration
	tcpConfig := &config.TCPServerConfig{
		Host:              "0.0.0.0",
		Port:              5555,
		ConnectionTimeout: 300,
		MaxConnections:    1000,
		BufferSize:        4096,
	}

	// Parse command line flags for configuration override
	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "-port" && i+1 < len(os.Args) {
				var port int
				_, err := fmt.Sscanf(os.Args[i+1], "%d", &port)
				if err == nil && port > 0 && port < 65536 {
					tcpConfig.Port = port
				}
				i++
			}
		}
	}

	logger.Infof("Starting TCP server on port %d", tcpConfig.Port)

	// Create and start the server
	server := nat.NewTCPServer(tcpConfig, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		logger.Fatalf("Failed to start TCP server: %v", err)
	}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Print server info and instructions
	logger.Infof("TCP server is running on %s:%d", tcpConfig.Host, tcpConfig.Port)
	logger.Info("Press Ctrl+C to stop the server")

	// Stats reporting
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.Infof("Active connections: %d", server.GetActiveConnections())
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for signal
	<-sigCh
	logger.Info("Shutdown signal received, stopping server...")

	// Stop the server
	if err := server.Stop(); err != nil {
		logger.Errorf("Error stopping server: %v", err)
	}

	logger.Info("Server stopped successfully")
}
