package networking

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConnectionState represents the state of a connection
type ConnectionState string

const (
	// ConnectionStateNew indicates a newly created connection
	ConnectionStateNew ConnectionState = "new"
	// ConnectionStateInitiating indicates a connection in the process of being established
	ConnectionStateInitiating ConnectionState = "initiating"
	// ConnectionStateEstablished indicates a fully established connection
	ConnectionStateEstablished ConnectionState = "established"
	// ConnectionStateFailed indicates a failed connection
	ConnectionStateFailed ConnectionState = "failed"
	// ConnectionStateClosed indicates a closed connection
	ConnectionStateClosed ConnectionState = "closed"
)

// Connection represents a tracked connection
type Connection struct {
	ID            string
	SourceAddr    net.Addr
	TargetAddr    net.Addr
	Type          ConnectionType
	State         ConnectionState
	CreatedAt     time.Time
	LastUpdatedAt time.Time
	Metadata      map[string]interface{}
}

// BaseConnectionTracker provides a basic implementation of the ConnectionTracker interface
type BaseConnectionTracker struct {
	connections map[string]*Connection
	mu          sync.RWMutex
}

// NewBaseConnectionTracker creates a new BaseConnectionTracker
func NewBaseConnectionTracker() *BaseConnectionTracker {
	return &BaseConnectionTracker{
		connections: make(map[string]*Connection),
	}
}

// AddConnection registers a new connection
func (t *BaseConnectionTracker) AddConnection(sourceAddr net.Addr, targetAddr net.Addr, connType ConnectionType) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := uuid.New().String()
	now := time.Now()

	t.connections[id] = &Connection{
		ID:            id,
		SourceAddr:    sourceAddr,
		TargetAddr:    targetAddr,
		Type:          connType,
		State:         ConnectionStateNew,
		CreatedAt:     now,
		LastUpdatedAt: now,
		Metadata:      make(map[string]interface{}),
	}

	return id, nil
}

// GetConnection retrieves a connection by ID
func (t *BaseConnectionTracker) GetConnection(connID string) (sourceAddr net.Addr, targetAddr net.Addr, connType ConnectionType, err error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	conn, exists := t.connections[connID]
	if !exists {
		return nil, nil, "", errors.New("connection not found")
	}

	return conn.SourceAddr, conn.TargetAddr, conn.Type, nil
}

// UpdateConnectionState updates the state of a connection
func (t *BaseConnectionTracker) UpdateConnectionState(connID string, state string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	conn, exists := t.connections[connID]
	if !exists {
		return errors.New("connection not found")
	}

	conn.State = ConnectionState(state)
	conn.LastUpdatedAt = time.Now()

	return nil
}

// RemoveConnection removes a tracked connection
func (t *BaseConnectionTracker) RemoveConnection(connID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.connections[connID]; !exists {
		return errors.New("connection not found")
	}

	delete(t.connections, connID)
	return nil
}

// ListConnections returns all active connections
func (t *BaseConnectionTracker) ListConnections() ([]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ids := make([]string, 0, len(t.connections))
	for id := range t.connections {
		ids = append(ids, id)
	}

	return ids, nil
}
