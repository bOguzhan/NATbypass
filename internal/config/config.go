// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

// ServersConfig contains configuration for different servers
type ServersConfig struct {
	Mediatory   ServerConfig `yaml:"mediatory"`
	Application ServerConfig `yaml:"application"`
}

// STUNConfig contains STUN server related configuration
type STUNConfig struct {
	Server  string        `yaml:"server"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
	Retries int           `yaml:"retries"`
}

// SignalingConfig contains signaling server related configuration
type SignalingConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ConnTTL         time.Duration `yaml:"conn_ttl"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

// TCPServerConfig contains TCP server related configuration for NAT traversal
type TCPServerConfig struct {
	ListenHost        string        `yaml:"listen_host"`
	ListenPort        int           `yaml:"listen_port"`
	Host              string        `yaml:"-"` // Not in YAML, just for code compatibility
	Port              int           `yaml:"-"` // Not in YAML, just for code compatibility
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	MaxConnections    int           `yaml:"max_connections"`
	BufferSize        int           `yaml:"buffer_size"`
}

// UDPServerConfig contains UDP server related configuration for NAT traversal
type UDPServerConfig struct {
	ListenHost       string        `yaml:"listen_host"`
	ListenPort       int           `yaml:"listen_port"`
	Host             string        `yaml:"-"` // Alias for ListenHost
	Port             int           `yaml:"-"` // Alias for ListenPort
	PacketBufferSize int           `yaml:"packet_buffer_size"`
	IdleTimeout      time.Duration `yaml:"idle_timeout"`
	MaxPacketSize    int           `yaml:"max_packet_size"`
}

// TraversalConfig contains NAT traversal related configuration
type TraversalConfig struct {
	PreferredProtocol string        `yaml:"preferred_protocol"`
	Timeout           time.Duration `yaml:"timeout"`
	MaxRetries        int           `yaml:"max_retries"`
	RelayServer       string        `yaml:"relay_server"`
	RelayPort         int           `yaml:"relay_port"`
}

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Servers   ServersConfig   `yaml:"servers"` // For newer code structure
	STUN      STUNConfig      `yaml:"stun"`
	Signaling SignalingConfig `yaml:"signaling"`
	TCP       TCPServerConfig `yaml:"tcp"`
	UDP       UDPServerConfig `yaml:"udp"`
	Traversal TraversalConfig `yaml:"traversal"`
}

// LoadConfig loads configuration from a yaml file with environment variable overrides
func LoadConfig(configPath string) (*Config, error) {
	// Create default config
	config := DefaultConfig()

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
			config.Server.Port = p
		}
	}

	if port := os.Getenv("APPLICATION_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if stunServer := os.Getenv("STUN_SERVER"); stunServer != "" {
		config.STUN.Server = stunServer
	}

	// Initialize aliases for backward compatibility
	config.TCP.Host = config.TCP.ListenHost
	config.TCP.Port = config.TCP.ListenPort
	config.UDP.Host = config.UDP.ListenHost
	config.UDP.Port = config.UDP.ListenPort

	return config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:     "0.0.0.0",
			Port:     8080,
			LogLevel: "info",
		},
		Servers: ServersConfig{
			Mediatory: ServerConfig{
				Host:     "0.0.0.0",
				Port:     8081,
				LogLevel: "info",
			},
			Application: ServerConfig{
				Host:     "0.0.0.0",
				Port:     8082,
				LogLevel: "info",
			},
		},
		STUN: STUNConfig{
			Server:  "stun.l.google.com",
			Port:    19302,
			Timeout: 5 * time.Second,
			Retries: 3,
		},
		Signaling: SignalingConfig{
			Host:            "0.0.0.0",
			Port:            8081,
			ConnTTL:         5 * time.Minute,
			CleanupInterval: 1 * time.Minute,
		},
		TCP: TCPServerConfig{
			ListenHost:        "0.0.0.0",
			ListenPort:        0,
			ConnectionTimeout: 30 * time.Second,
			IdleTimeout:       2 * time.Minute,
			MaxConnections:    100,
			BufferSize:        4096,
		},
		UDP: UDPServerConfig{
			ListenHost:       "0.0.0.0",
			ListenPort:       0,
			PacketBufferSize: 4096,
			IdleTimeout:      2 * time.Minute,
			MaxPacketSize:    1500,
		},
		Traversal: TraversalConfig{
			PreferredProtocol: "",
			Timeout:           30 * time.Second,
			MaxRetries:        5,
			RelayServer:       "",
			RelayPort:         3478,
		},
	}

	// Set up aliases for backward compatibility
	cfg.TCP.Host = cfg.TCP.ListenHost
	cfg.TCP.Port = cfg.TCP.ListenPort
	cfg.UDP.Host = cfg.UDP.ListenHost
	cfg.UDP.Port = cfg.UDP.ListenPort

	return cfg
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
