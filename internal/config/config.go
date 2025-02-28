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

// Config is the root configuration structure
type Config struct {
	Servers struct {
		Mediatory   ServerConfig `yaml:"mediatory"`
		Application ServerConfig `yaml:"application"`
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
			config.Servers.Application.Port = p
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
