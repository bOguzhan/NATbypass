// cmd/mediatory-server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bOguzhan/NATbypass/internal/config"
	"github.com/bOguzhan/NATbypass/internal/signaling"
	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/gin-gonic/gin"
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
	logger := utils.NewLogger("mediatory-server", cfg.Servers.Mediatory.LogLevel)
	logger.Info("Starting Mediatory Server...")

	// Set Gin mode based on log level
	if cfg.Servers.Mediatory.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create and configure server
	server := signaling.NewServer(logger)

	// Create handlers and set server reference
	handlers := server.GetHandlers()
	handlers.SetServer(server)
	handlers.SetConfig(cfg) // Set the configuration

	// Start HTTP server in a goroutine
	serverAddr := fmt.Sprintf("%s:%d",
		cfg.Servers.Mediatory.Host,
		cfg.Servers.Mediatory.Port)

	go func() {
		logger.Infof("Mediatory Server listening on %s", serverAddr)
		if err := server.Start(serverAddr); err != nil {
			if err != http.ErrServerClosed {
				logger.Errorf("Failed to start server: %v", err)
			}
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-quit
	logger.Info("Shutting down server...")

	// Give server 5 seconds to shut down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}
