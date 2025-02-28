// internal/signaling/connections.go
package signaling

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/gin-gonic/gin"
)

// ConnectionRequest represents a request to connect to another peer
type ConnectionRequest struct {
	SourceID      string    `json:"source_id"`
	TargetID      string    `json:"target_id"`
	Timestamp     time.Time `json:"timestamp"`
	ConnectionID  string    `json:"connection_id"`
	SourceAddress string    `json:"source_address,omitempty"`
	TargetAddress string    `json:"target_address,omitempty"`
}

// ConnectionRegistry manages active connection requests between clients.
// It provides thread-safe operations for storing, retrieving, and removing
// connection requests between peers.
type ConnectionRegistry struct {
	mu          sync.RWMutex
	connections map[string]*ConnectionRequest
	logger      *utils.Logger
}

// NewConnectionRegistry creates a new connection registry
func NewConnectionRegistry(logger *utils.Logger) *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]*ConnectionRequest),
		logger:      logger,
	}
}

// RegisterConnection adds a new connection request to the registry
func (r *ConnectionRegistry) RegisterConnection(req *ConnectionRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if req.ConnectionID == "" {
		var err error
		req.ConnectionID, err = utils.GenerateSessionID()
		if err != nil {
			return fmt.Errorf("failed to generate connection ID: %w", err)
		}
	}

	req.Timestamp = time.Now()
	r.connections[req.ConnectionID] = req

	r.logger.WithFields(map[string]interface{}{
		"connection_id": req.ConnectionID,
		"source_id":     req.SourceID,
		"target_id":     req.TargetID,
	}).Info("Connection registered")

	return nil
}

// GetConnection retrieves a connection request by ID
func (r *ConnectionRegistry) GetConnection(connectionID string) (*ConnectionRequest, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conn, exists := r.connections[connectionID]
	return conn, exists
}

// GetConnectionsByClient finds all connections for a specific client
func (r *ConnectionRegistry) GetConnectionsByClient(clientID string) []*ConnectionRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ConnectionRequest
	for _, conn := range r.connections {
		if conn.SourceID == clientID || conn.TargetID == clientID {
			result = append(result, conn)
		}
	}
	return result
}

// RemoveConnection removes a connection from the registry
func (r *ConnectionRegistry) RemoveConnection(connectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.connections[connectionID]; exists {
		delete(r.connections, connectionID)
		r.logger.WithFields(map[string]interface{}{
			"connection_id": connectionID,
		}).Info("Connection removed")
	}
}

// CleanupStaleConnections removes connections older than the specified duration
func (r *ConnectionRegistry) CleanupStaleConnections(maxAge time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	count := 0

	for id, conn := range r.connections {
		if conn.Timestamp.Before(cutoff) {
			delete(r.connections, id)
			count++
		}
	}

	if count > 0 {
		r.logger.Infof("Cleaned up %d stale connection requests", count)
	}

	return count
}

// Add connection handling methods to the Handlers type
func (h *Handlers) InitConnectionHandlers() {
	// Create connection registry if it doesn't exist
	if h.connections == nil {
		h.connections = NewConnectionRegistry(h.logger)

		// Start a periodic cleanup goroutine
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				h.connections.CleanupStaleConnections(30 * time.Minute)
			}
		}()
	}
}

// Add this field to the Handlers type in handlers.go
// connections *ConnectionRegistry

// RequestConnection initiates a connection request to another client
func (h *Handlers) RequestConnection(c *gin.Context) {
	var req struct {
		SourceID   string `json:"source_id" binding:"required"`
		TargetID   string `json:"target_id" binding:"required"`
		SourceIP   string `json:"source_ip,omitempty"`
		SourcePort int    `json:"source_port,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid_request",
			"detail": err.Error(),
		})
		return
	}

	// Validate client IDs
	if !utils.ValidateID(req.SourceID, 32) || !utils.ValidateID(req.TargetID, 32) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid_client_id",
		})
		return
	}

	// Check if source client exists if server is available
	if h.server != nil {
		if _, exists := h.server.GetClient(req.SourceID); !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "source_client_not_found",
			})
			return
		}

		if _, exists := h.server.GetClient(req.TargetID); !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "target_client_not_found",
			})
			return
		}
	}

	// Create and register the connection request
	connReq := &ConnectionRequest{
		SourceID:  req.SourceID,
		TargetID:  req.TargetID,
		Timestamp: time.Now(),
	}

	if req.SourceIP != "" && req.SourcePort > 0 {
		connReq.SourceAddress = fmt.Sprintf("%s:%d", req.SourceIP, req.SourcePort)
	}

	if err := h.connections.RegisterConnection(connReq); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to register connection")

		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "connection_registration_failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "connection_registered",
		"connection_id": connReq.ConnectionID,
		"timestamp":     connReq.Timestamp,
	})
}

// GetActiveConnections returns all active connection requests for a client
func (h *Handlers) GetActiveConnections(c *gin.Context) {
	clientID := c.Param("client_id")

	if !utils.ValidateID(clientID, 32) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid_client_id",
		})
		return
	}

	connections := h.connections.GetConnectionsByClient(clientID)

	// Format for response
	response := make([]gin.H, 0, len(connections))
	for _, conn := range connections {
		response = append(response, gin.H{
			"connection_id": conn.ConnectionID,
			"source_id":     conn.SourceID,
			"target_id":     conn.TargetID,
			"timestamp":     conn.Timestamp,
			"is_initiator":  conn.SourceID == clientID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"connections": response,
	})
}
