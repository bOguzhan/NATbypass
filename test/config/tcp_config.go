package config

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

// Add this to the ServerConfig struct in config.go
type ServerConfig struct {
	Mediatory struct {
		Host string `yaml:"host" env:"MEDIATORY_HOST"`
		Port int    `yaml:"port" env:"MEDIATORY_PORT"`
	} `yaml:"mediatory"`
	Application struct {
		Host string `yaml:"host" env:"APPLICATION_HOST"`
		Port int    `yaml:"port" env:"APPLICATION_PORT"`
	} `yaml:"application"`
	TCP TCPServerConfig `yaml:"tcp"` // Add this line
}
