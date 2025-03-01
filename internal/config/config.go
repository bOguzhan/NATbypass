// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ServerConfig holds configuration for a server component
type ServerConfig struct {
	Port     int    `yaml:"port"`
	Host     string `yaml:"host"`
	LogLevel string `yaml:"log_level"`
}

// StunConfig holds configuration for STUN operations
type StunConfig struct {
	Server         string `yaml:"server"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	RetryCount     int    `yaml:"retry_count"`
}

// ConnectionConfig holds NAT traversal configuration parameters
type ConnectionConfig struct {
	HolePunchAttempts     int `yaml:"hole_punch_attempts"`
	HolePunchTimeoutMs    int `yaml:"hole_punch_timeout_ms"`
	KeepAliveIntervalSecs int `yaml:"keep_alive_interval_seconds"`
}

// TCPServerConfig holds configuration for a TCP server
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

// ApplicationConfig holds configuration for the application
type ApplicationConfig struct {
	// TCP server configuration
	TCP TCPServerConfig `yaml:"tcp"`
}

// Config is the root configuration structure
type Config struct {
	Servers struct {
		Mediatory   ServerConfig      `yaml:"mediatory"`
		Application ApplicationConfig `yaml:"application"`
	} `yaml:"servers"`
	Stun       StunConfig       `yaml:"stun"`
	Connection ConnectionConfig `yaml:"connection"`
}

// LoadConfig loads configuration from a yaml file with environment variable overrides
func LoadConfig(configPath string) (*Config, error) {
	// Default configuration
	config := &Config{}

	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML into config struct
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Override with environment variables if present
	if port := os.Getenv("MEDIATORY_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Servers.Mediatory.Port = p
		}
	}

	if port := os.Getenv("APPLICATION_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Servers.Application.TCP.Port = p
		}
	}

	if stunServer := os.Getenv("STUN_SERVER"); stunServer != "" {
		config.Stun.Server = stunServer
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
