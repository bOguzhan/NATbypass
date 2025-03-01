// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// MediatoryConfig holds configuration for the mediatory server
type MediatoryConfig struct {
	Host     string `yaml:"host" env:"MEDIATORY_HOST" default:"0.0.0.0"`
	Port     int    `yaml:"port" env:"MEDIATORY_PORT" default:"8080"`
	LogLevel string `yaml:"log_level" env:"MEDIATORY_LOG_LEVEL" default:"info"`
}

// ApplicationConfig holds configuration for the application server
type ApplicationConfig struct {
	Host     string `yaml:"host" env:"APP_HOST" default:"0.0.0.0"`
	Port     int    `yaml:"port" env:"APP_PORT" default:"9000"`
	LogLevel string `yaml:"log_level" env:"APP_LOG_LEVEL" default:"info"`
}

// TCPServerConfig holds the configuration for the TCP server
type TCPServerConfig struct {
	// Host to bind the TCP server to
	Host string `yaml:"host" env:"TCP_SERVER_HOST" default:"0.0.0.0"`

	// Port to listen on
	Port int `yaml:"port" env:"TCP_SERVER_PORT" default:"5555"`

	// ConnectionTimeout in seconds for inactive connections
	ConnectionTimeout int `yaml:"connection_timeout" env:"TCP_CONNECTION_TIMEOUT" default:"300"`

	// MaxConnections is the maximum number of concurrent TCP connections
	MaxConnections int `yaml:"max_connections" env:"TCP_MAX_CONNECTIONS" default:"1000"`

	// BufferSize for TCP read operations
	BufferSize int `yaml:"buffer_size" env:"TCP_BUFFER_SIZE" default:"4096"`
}

// Config represents the application configuration
type Config struct {
	Servers struct {
		Mediatory   MediatoryConfig   `yaml:"mediatory"`
		Application ApplicationConfig `yaml:"application"`
		TCP         TCPServerConfig   `yaml:"tcp"` // Move TCP config to top level
	} `yaml:"servers"`
	STUN struct {
		Server         string `yaml:"server" env:"STUN_SERVER" default:"stun.l.google.com:19302"`
		TimeoutSeconds int    `yaml:"timeout_seconds" env:"STUN_TIMEOUT" default:"5"`
		RetryCount     int    `yaml:"retry_count" env:"STUN_RETRY" default:"3"`
	} `yaml:"stun"`
	Connection struct {
		HolePunchAttempts        int `yaml:"hole_punch_attempts" env:"HOLE_PUNCH_ATTEMPTS" default:"5"`
		HolePunchTimeoutMs       int `yaml:"hole_punch_timeout_ms" env:"HOLE_PUNCH_TIMEOUT_MS" default:"500"`
		KeepAliveIntervalSeconds int `yaml:"keep_alive_interval_seconds" env:"KEEP_ALIVE_INTERVAL" default:"30"`
	} `yaml:"connection"`
}

// LoadConfig loads configuration from a yaml file with environment variable overrides
func LoadConfig(configPath string) (*Config, error) {
	// Create default config
	config := &Config{}

	// Set defaults
	// ... (existing default setting code)

	// If config file exists, load it
	if configPath != "" {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables if present
	if port := os.Getenv("MEDIATORY_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Servers.Mediatory.Port = p
		}
	}

	if port := os.Getenv("APPLICATION_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Servers.Application.Port = p
		}
	}

	if stunServer := os.Getenv("STUN_SERVER"); stunServer != "" {
		config.STUN.Server = stunServer
	}

	return config, nil
}

// ConfigureLogger sets up the logger based on configuration
func ConfigureLogger(logLevel string) *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Set log level from config
	switch logLevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	return log
}
