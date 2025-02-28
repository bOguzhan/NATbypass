// test/config/config_test.go
package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/bOguzhan/NATbypass/internal/config"
)

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file for testing
	tempConfig := `
servers:
  mediatory:
    port: 8080
    host: "0.0.0.0"
    log_level: "info"
  application:
    port: 9000
    host: "0.0.0.0"
    log_level: "info"
stun:
  server: "stun.l.google.com:19302"
  timeout_seconds: 5
  retry_count: 3
connection:
  hole_punch_attempts: 5
  hole_punch_timeout_ms: 500
  keep_alive_interval_seconds: 30
`
	// Create a temporary file with the test configuration
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	if _, err := tmpfile.Write([]byte(tempConfig)); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test loading from temporary file
	cfg, err := config.LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify values from the temporary config
	expectedMediatoryPort := 8080
	if cfg.Servers.Mediatory.Port != expectedMediatoryPort {
		t.Errorf("Expected mediatory port to be %d, got %d", expectedMediatoryPort, cfg.Servers.Mediatory.Port)
	}

	// Verify STUN configuration
	expectedStunServer := "stun.l.google.com:19302"
	if cfg.Stun.Server != expectedStunServer {
		t.Errorf("Expected STUN server to be %s, got %s", expectedStunServer, cfg.Stun.Server)
	}

	// Test environment variable override
	testPort := 9090
	os.Setenv("MEDIATORY_PORT", fmt.Sprintf("%d", testPort))
	defer os.Unsetenv("MEDIATORY_PORT")

	cfg, err = config.LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load configuration with environment variable: %v", err)
	}

	if cfg.Servers.Mediatory.Port != testPort {
		t.Errorf("Expected mediatory port to be overridden to %d, got %d", testPort, cfg.Servers.Mediatory.Port)
	}

	// Test logger configuration
	logger := config.ConfigureLogger("debug")
	if logger == nil {
		t.Error("Logger should not be nil")
	}
}
