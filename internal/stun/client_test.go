// internal/stun/client_test.go
package stun

import (
	"testing"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/sirupsen/logrus"
)

func TestSTUNClient(t *testing.T) {
	utilsLogger := utils.NewLogger("stun-test", "info")

	// Create a logrus logger for the STUN client
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Use Google's public STUN server for testing
	client := NewClient(logger, "stun.l.google.com:19302", 5, 3)

	// Test public address discovery
	addr, err := client.DiscoverPublicAddress()
	if err != nil {
		t.Skipf("STUN test skipped: %v (this may fail in CI environments)", err)
		return
	}

	// Check if we got valid results
	if addr.IP == nil {
		t.Error("Expected non-nil IP address")
	}

	if addr.Port == 0 {
		t.Error("Expected non-zero port")
	}

	t.Logf("Discovered public address: %s:%d", addr.IP.String(), addr.Port)
}
