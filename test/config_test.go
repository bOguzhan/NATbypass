// test/config_test.go
package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/bOguzhan/NATbypass/internal/config"
)

func TestConfigLoading(t *testing.T) {
	// Test loading from file
	cfg, err := config.LoadConfig("../configs/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify default values from config.yaml
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

	cfg, err = config.LoadConfig("../configs/config.yaml")
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
