// File: internal/nat/config_init.go

package nat

import (
	"github.com/bOguzhan/NATbypass/internal/config"
)

func init() {
	// This is a temporary fix to ensure TCPServerConfig compatibility
	// It will be executed when the package is imported
	defaultConfig := config.DefaultConfig()

	// Make sure Host and Port are properly initialized
	defaultConfig.TCP.Host = defaultConfig.TCP.ListenHost
	defaultConfig.TCP.Port = defaultConfig.TCP.ListenPort
	defaultConfig.UDP.Host = defaultConfig.UDP.ListenHost
	defaultConfig.UDP.Port = defaultConfig.UDP.ListenPort
}
