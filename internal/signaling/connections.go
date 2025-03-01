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

// ConnectionStatus represents the current state of a connection
type ConnectionStatus string

const (
	// StatusInitiated indicates a connection request has been initiated
	StatusInitiated ConnectionStatus = "initiated"

	// StatusNegotiating indicates peers are exchanging connection information
	StatusNegotiating ConnectionStatus = "negotiating"

	// StatusEstablished indicates the connection has been successfully established
	StatusEstablished ConnectionStatus = "established"

	// StatusFailed indicates the connection attempt failed
	StatusFailed ConnectionStatus = "failed"

	// StatusClosed indicates the connection was established but is now closed
	StatusClosed ConnectionStatus = "closed"
)

// ConnectionRequest represents a request to connect to another peer
type ConnectionRequest struct {
	SourceID      string                 `json:"source_id"`
	TargetID      string                 `json:"target_id"`
	Timestamp     time.Time              `json:"timestamp"`
	ConnectionID  string                 `json:"connection_id"`
	SourceAddress string                 `json:"source_address,omitempty"`
	TargetAddress string                 `json:"target_address,omitempty"`
	Status        ConnectionStatus       `json:"status"`
	LastUpdated   time.Time              `json:"last_updated"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectionRegistry manages active connection requests between clients.
// It provides thread-safe operations for storing, retrieving, and removing
// connection requests between peers.
type ConnectionRegistry struct {
	mu              sync.RWMutex
	connections     map[string]*ConnectionRequest
	logger          *utils.Logger
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewConnectionRegistry creates a new connection registry
func NewConnectionRegistry(logger *utils.Logger) *ConnectionRegistry {
	registry := &ConnectionRegistry{
		connections:     make(map[string]*ConnectionRequest),
		logger:          logger,
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup routine
	go registry.startPeriodicCleanup()

	return registry
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

	if req.Status == "" {
		req.Status = StatusInitiated
	}

	now := time.Now()
	req.Timestamp = now
	req.LastUpdated = now

	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}

	r.connections[req.ConnectionID] = req

	r.logger.WithFields(map[string]interface{}{
		"connection_id": req.ConnectionID,
		"source_id":     req.SourceID,
		"target_id":     req.TargetID,
		"status":        req.Status,
	}).Info("Connection registered")

	return nil
}

// UpdateConnectionStatus updates the status of a connection
func (r *ConnectionRegistry) UpdateConnectionStatus(connectionID string, status ConnectionStatus) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.connections[connectionID]
	if !exists {
		return false
	}

	conn.Status = status
	conn.LastUpdated = time.Now()

	r.logger.WithFields(map[string]interface{}{
		"connection_id": connectionID,
		"source_id":     conn.SourceID,
		"target_id":     conn.TargetID,
		"status":        status,
	}).Info("Connection status updated")

	return true
}

// UpdateConnectionError sets an error message for a connection
func (r *ConnectionRegistry) UpdateConnectionError(connectionID string, errorMsg string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.connections[connectionID]
	if !exists {
		return false
	}

	conn.Status = StatusFailed
	conn.ErrorMessage = errorMsg
	conn.LastUpdated = time.Now()

	r.logger.WithFields(map[string]interface{}{
		"connection_id": connectionID,
		"error":         errorMsg,
	}).Info("Connection marked as failed")

	return true
}

// UpdateConnectionMetadata adds or updates metadata for a connection
func (r *ConnectionRegistry) UpdateConnectionMetadata(connectionID string, key string, value interface{}) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.connections[connectionID]
	if !exists {
		return false
	}

	if conn.Metadata == nil {
		conn.Metadata = make(map[string]interface{})
	}

	conn.Metadata[key] = value
	conn.LastUpdated = time.Now()

	return true
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

// GetConnectionsByStatus returns all connections with the specified status
func (r *ConnectionRegistry) GetConnectionsByStatus(status ConnectionStatus) []*ConnectionRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ConnectionRequest
	for _, conn := range r.connections {
		if conn.Status == status {
			result = append(result, conn)
		}
	}
	return result
}

// RemoveConnection removes a connection from the registry
func (r *ConnectionRegistry) RemoveConnection(connectionID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.connections[connectionID]; exists {
		delete(r.connections, connectionID)
		r.logger.WithFields(map[string]interface{}{
			"connection_id": connectionID,
		}).Info("Connection removed")
		return true
	}
	return false
}

// CleanupStaleConnections removes connections older than the specified duration
func (r *ConnectionRegistry) CleanupStaleConnections(maxAge time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-maxAge)
	count := 0

	// Add debug logging
	r.logger.Debugf("Running cleanup with cutoff time: %v", cutoff)
	r.logger.Debugf("Initial connection count: %d", len(r.connections))

	// Store IDs to delete to avoid map modification during iteration
	toDelete := make([]string, 0)

	for id, conn := range r.connections {
		shouldDelete := false

		// Different cleanup policies based on connection status
		switch conn.Status {
		case StatusEstablished:
			// For established connections, check last updated time
			if conn.LastUpdated.Before(cutoff) {
				shouldDelete = true
				r.logger.Debugf("Will remove established connection %s: last updated %v before cutoff %v",
					id, conn.LastUpdated, cutoff)
			}
		case StatusFailed, StatusClosed:
			// Failed/closed connections removed after 1 hour
			failedCutoff := now.Add(-1 * time.Hour)
			if conn.LastUpdated.Before(failedCutoff) {
				shouldDelete = true
				r.logger.Debugf("Will remove failed/closed connection %s: last updated %v before failedCutoff %v",
					id, conn.LastUpdated, failedCutoff)
			}
		default:
			// For initiated/negotiating statuses, use timestamp
			if conn.Timestamp.Before(cutoff) {
				shouldDelete = true
				r.logger.Debugf("Will remove connection %s: timestamp %v before cutoff %v",
					id, conn.Timestamp, cutoff)
			}
		}

		if shouldDelete {
			toDelete = append(toDelete, id)
		}
	}

	// Now actually delete the connections
	for _, id := range toDelete {
		r.logger.Debugf("Deleting connection %s", id)
		delete(r.connections, id)
		count++
	}

	if count > 0 {
		r.logger.Infof("Cleaned up %d stale connection requests", count)
	}

	return count
}

// startPeriodicCleanup runs a periodic task to clean up stale connections
func (r *ConnectionRegistry) startPeriodicCleanup() {
	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Remove connections older than 30 minutes
			r.CleanupStaleConnections(30 * time.Minute)
		case <-r.stopCleanup:
			return
		}
	}
}

// Stop halts the background cleanup routine
func (r *ConnectionRegistry) Stop() {
	close(r.stopCleanup)
}

// GetConnectionStats returns statistics about the connections
func (r *ConnectionRegistry) GetConnectionStats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := map[string]int{
		"total":       len(r.connections),
		"initiated":   0,
		"negotiating": 0,
		"established": 0,
		"failed":      0,
		"closed":      0,
	}

	for _, conn := range r.connections {
		if count, exists := stats[string(conn.Status)]; exists {
			stats[string(conn.Status)] = count + 1
		}
	}

	return stats
}

// Add connection handling methods to the Handlers type
func (h *Handlers) InitConnectionHandlers() {
	// Create connection registry if it doesn't exist
	if h.connections == nil {
		h.connections = NewConnectionRegistry(h.logger)
	}
}

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
		Status:    StatusInitiated,
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
			"status":        conn.Status,
			"timestamp":     conn.Timestamp,
			"last_updated":  conn.LastUpdated,
			"is_initiator":  conn.SourceID == clientID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"connections": response,
	})
}

// Add a new handler for updating connection status
func (h *Handlers) UpdateConnectionStatus(c *gin.Context) {
	var req struct {
		ConnectionID string           `json:"connection_id" binding:"required"`
		Status       ConnectionStatus `json:"status" binding:"required"`
		ErrorMessage string           `json:"error_message,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid_request",
			"detail": err.Error(),
		})
		return
	}

	success := false
	if req.ErrorMessage != "" {
		success = h.connections.UpdateConnectionError(req.ConnectionID, req.ErrorMessage)
	} else {
		success = h.connections.UpdateConnectionStatus(req.ConnectionID, req.Status)
	}

	if !success {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "connection_not_found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "updated",
		"connection_id": req.ConnectionID,
	})
}
