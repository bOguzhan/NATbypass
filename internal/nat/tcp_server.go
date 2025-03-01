package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/bOguzhan/NATbypass/internal/config"
)

// TCPServer represents a TCP server for NAT traversal operations
type TCPServer struct {
	config        *config.TCPServerConfig
	listener      net.Listener
	connections   map[string]*TCPConnection
	connectionsMu sync.RWMutex
	logger        *logrus.Logger
	stopChan      chan struct{}
}

// TCPConnection represents a TCP connection with its metadata
type TCPConnection struct {
	ID         string
	Conn       net.Conn
	RemoteAddr net.Addr
	LocalAddr  net.Addr
	CreatedAt  time.Time
	LastActive time.Time
}

// NewTCPServer creates a new TCPServer instance
func NewTCPServer(cfg *config.TCPServerConfig, logger *logrus.Logger) *TCPServer {
	return &TCPServer{
		config:      cfg,
		connections: make(map[string]*TCPConnection),
		logger:      logger,
		stopChan:    make(chan struct{}),
	}
}

// Start initializes and starts the TCP server
func (s *TCPServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Infof("Starting TCP server on %s", addr)

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}

	go s.acceptLoop(ctx)
	go s.maintenanceLoop(ctx)

	return nil
}

// acceptLoop continuously accepts new TCP connections
func (s *TCPServer) acceptLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.logger.Errorf("Error accepting TCP connection: %v", err)
				continue
			}

			connID := uuid.New().String()
			tcpConn := &TCPConnection{
				ID:         connID,
				Conn:       conn,
				RemoteAddr: conn.RemoteAddr(),
				LocalAddr:  conn.LocalAddr(),
				CreatedAt:  time.Now(),
				LastActive: time.Now(),
			}

			s.connectionsMu.Lock()
			s.connections[connID] = tcpConn
			s.connectionsMu.Unlock()

			s.logger.Infof("New TCP connection established: %s -> %s (ID: %s)",
				conn.RemoteAddr().String(), conn.LocalAddr().String(), connID)

			go s.handleConnection(ctx, tcpConn)
		}
	}
}

// handleConnection manages an individual TCP connection
func (s *TCPServer) handleConnection(ctx context.Context, conn *TCPConnection) {
	defer func() {
		conn.Conn.Close()
		s.connectionsMu.Lock()
		delete(s.connections, conn.ID)
		s.connectionsMu.Unlock()
		s.logger.Infof("TCP connection closed: %s (ID: %s)", conn.RemoteAddr.String(), conn.ID)
	}()

	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			// Set read deadline to prevent blocking indefinitely
			if err := conn.Conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
				s.logger.Errorf("Failed to set read deadline: %v", err)
				return
			}

			n, err := conn.Conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is a timeout, which is expected when no data is received
					continue
				}
				s.logger.Debugf("Read error from %s: %v", conn.RemoteAddr.String(), err)
				return
			}

			// Update last active time
			conn.LastActive = time.Now()
			s.handlePacket(conn, buffer[:n])
		}
	}
}

// handlePacket processes a packet received over TCP
func (s *TCPServer) handlePacket(conn *TCPConnection, data []byte) {
	// TODO: Implement protocol-specific packet handling and validation
	s.logger.Debugf("Received %d bytes from %s", len(data), conn.RemoteAddr.String())

	// Basic packet validation
	if len(data) < 4 {
		s.logger.Warnf("Received invalid packet (too short) from %s", conn.RemoteAddr.String())
		return
	}

	// Example packet processing - to be replaced with actual protocol implementation
	packetType := data[0]
	s.logger.Debugf("Processed packet of type %d from %s", packetType, conn.RemoteAddr.String())

	// Placeholder for actual packet handling logic
	switch packetType {
	case 0x01: // Registration packet
		s.logger.Infof("Registration packet from %s", conn.RemoteAddr.String())
		// TODO: Implement registration logic
	case 0x02: // Hole punching packet
		s.logger.Infof("Hole punching packet from %s", conn.RemoteAddr.String())
		// TODO: Implement hole punching logic
	case 0x03: // Keep-alive packet
		s.logger.Debugf("Keep-alive packet from %s", conn.RemoteAddr.String())
		// Send acknowledgment
		conn.Conn.Write([]byte{0x03, 0x01})
	default:
		s.logger.Warnf("Unknown packet type %d from %s", packetType, conn.RemoteAddr.String())
	}
}

// maintenanceLoop periodically cleans up stale connections
func (s *TCPServer) maintenanceLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.cleanupStaleConnections()
		}
	}
}

// cleanupStaleConnections removes inactive connections
func (s *TCPServer) cleanupStaleConnections() {
	threshold := time.Now().Add(-time.Duration(s.config.ConnectionTimeout) * time.Second)

	s.connectionsMu.Lock()
	defer s.connectionsMu.Unlock()

	for id, conn := range s.connections {
		if conn.LastActive.Before(threshold) {
			s.logger.Infof("Closing stale TCP connection: %s (ID: %s, inactive for %v)",
				conn.RemoteAddr.String(), id, time.Since(conn.LastActive))
			conn.Conn.Close()
			delete(s.connections, id)
		}
	}
}

// ForceCleanup forces an immediate cleanup of stale connections
// This is mainly used for testing purposes
func (s *TCPServer) ForceCleanup() {
	s.cleanupStaleConnections()
}

// GetActiveConnections returns the count of currently active connections
func (s *TCPServer) GetActiveConnections() int {
	s.connectionsMu.RLock()
	defer s.connectionsMu.RUnlock()
	return len(s.connections)
}

// SendTo sends data to a specific connection by ID
func (s *TCPServer) SendTo(connectionID string, data []byte) error {
	s.connectionsMu.RLock()
	conn, exists := s.connections[connectionID]
	s.connectionsMu.RUnlock()

	if !exists {
		return fmt.Errorf("connection %s not found", connectionID)
	}

	_, err := conn.Conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send data to %s: %w", connectionID, err)
	}

	return nil
}

// Stop gracefully shuts down the TCP server
func (s *TCPServer) Stop() error {
	s.logger.Info("Stopping TCP server...")

	// Only close the channel if it hasn't been closed already
	select {
	case <-s.stopChan:
		// Channel already closed, do nothing
	default:
		close(s.stopChan)
	}

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("error closing TCP listener: %w", err)
		}
	}

	// Close all active connections
	s.connectionsMu.Lock()
	defer s.connectionsMu.Unlock()

	for id, conn := range s.connections {
		s.logger.Debugf("Closing TCP connection: %s (ID: %s)", conn.RemoteAddr.String(), id)
		conn.Conn.Close()
	}

	s.connections = make(map[string]*TCPConnection)
	s.logger.Info("TCP server stopped successfully")
	return nil
}
