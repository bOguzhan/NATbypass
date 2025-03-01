// cmd/application-server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bOguzhan/NATbypass/internal/config"
	"github.com/bOguzhan/NATbypass/internal/nat"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger := logrus.New()
	logLevel, err := logrus.ParseLevel(cfg.Servers.Application.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Setup HTTP server
	router := gin.Default()
	addr := fmt.Sprintf("%s:%d", cfg.Servers.Application.Host, cfg.Servers.Application.Port)

	// Setup routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"server": "application",
		})
	})

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup TCP server
	tcpConfig := &config.TCPServerConfig{
		Host:              cfg.Servers.TCP.Host,
		Port:              cfg.Servers.TCP.Port,
		ConnectionTimeout: cfg.Servers.TCP.ConnectionTimeout,
		MaxConnections:    cfg.Servers.TCP.MaxConnections,
		BufferSize:        cfg.Servers.TCP.BufferSize,
	}
	tcpServer := nat.NewTCPServer(tcpConfig, logger)

	// Start TCP server
	logger.Info("Starting TCP server...")
	if err := tcpServer.Start(ctx); err != nil {
		logger.Fatalf("Failed to start TCP server: %v", err)
	}

	// Setup server graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Infof("Starting application server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received, stopping server...")

	// Create a deadline for shutdown operations
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Stop HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("HTTP server shutdown error: %v", err)
	}

	// Stop TCP server
	if err := tcpServer.Stop(); err != nil {
		logger.Errorf("TCP server shutdown error: %v", err)
	}

	logger.Info("Server stopped successfully")
}
