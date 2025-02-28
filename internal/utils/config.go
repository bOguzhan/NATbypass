// internal/utils/config.go
package utils

// Config is a simple wrapper for configuration to be used across packages
type Config struct {
	STUNServer     string
	LogLevel       string
	TimeoutSeconds int
	RetryCount     int
}

// NewDefaultConfig creates a Config with sensible defaults
func NewDefaultConfig() *Config {
	return &Config{
		STUNServer:     "stun.l.google.com:19302",
		LogLevel:       "info",
		TimeoutSeconds: 5,
		RetryCount:     3,
	}
}
