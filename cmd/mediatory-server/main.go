// cmd/mediatory-server/main.go
package main

import (
	"fmt"
	"os"

	"github.com/bOguzhan/NATbypass/internal/config"
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
	log := config.ConfigureLogger(cfg.Servers.Mediatory.LogLevel)
	log.Info("Starting Mediatory Server...")

	// Set Gin mode based on log level
	if cfg.Servers.Mediatory.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Set up Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Use middleware to log requests
	router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		log.Infof("Request: %s %s", method, path)

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "mediatory-server",
		})
	})

	// Start HTTP server
	serverAddr := fmt.Sprintf("%s:%d",
		cfg.Servers.Mediatory.Host,
		cfg.Servers.Mediatory.Port)

	log.Infof("Mediatory Server listening on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
