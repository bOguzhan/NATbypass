package nat

import (
    "github.com/bOguzhan/NATbypass/internal/config"
)

// InitTCPServerConfig initializes a TCP server config, ensuring compatibility
func initTCPServerConfig(cfg *config.TCPServerConfig) {
    if cfg != nil {
        // Create Host/Port aliases based on ListenHost/ListenPort
        if cfg.Host == "" {
            cfg.Host = cfg.ListenHost
        }
        if cfg.Port == 0 {
            cfg.Port = cfg.ListenPort
        }
    }
}

// NewTCPServerWithConfig creates a new TCP server with configuration
func NewTCPServerWithConfig(cfg *config.TCPServerConfig) *TCPServer {
    initTCPServerConfig(cfg)
    return nil // This is a stub - the actual implementation is in tcp_server.go
}
