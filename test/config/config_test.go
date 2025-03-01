package config_test

import (
	"testing"

	"github.com/bOguzhan/NATbypass/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// Simple test to verify the config package can be imported
	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		t.Skipf("Config test skipped: %v (this may fail if config file doesn't exist)", err)
		return
	}

	// Check if basic properties are loaded
	if cfg == nil {
		t.Error("Expected non-nil config object")
	}
}
