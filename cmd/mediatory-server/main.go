// cmd/mediatory-server/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	// Set up Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Add custom logger middleware
	router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		logger.WithFields(map[string]interface{}{
			"method": method,
			"path":   path,
			"client": c.ClientIP(),
		}).Info("Request")

		c.Next()

		logger.WithFields(map[string]interface{}{
			"method": method,
			"path":   path,
			"status": c.Writer.Status(),
		}).Debug("Response")
	})

	// Set up the signaling handlers
	handlers := signaling.NewHandlers(logger)
	handlers.SetupRoutes(router)

	// Start HTTP server
	serverAddr := fmt.Sprintf("%s:%d",
		cfg.Servers.Mediatory.Host,
		cfg.Servers.Mediatory.Port)

	// Setup graceful shutdown
	go func() {
		logger.Infof("Mediatory Server listening on %s", serverAddr)
		if err := router.Run(serverAddr); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
}
