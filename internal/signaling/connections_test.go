// internal/signaling/connections_test.go
package signaling

import (
	"testing"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestConnectionRegistry(t *testing.T) {
	logger := utils.NewLogger("test", "info")
	registry := NewConnectionRegistry(logger)
	defer registry.Stop() // Make sure cleanup goroutine is stopped

	// Test registering a connection
	conn := &ConnectionRequest{
		SourceID:  "source1",
		TargetID:  "target1",
		Timestamp: time.Now(),
	}

	err := registry.RegisterConnection(conn)
	assert.NoError(t, err)
	assert.NotEmpty(t, conn.ConnectionID)
	assert.Equal(t, StatusInitiated, conn.Status)

	// Test retrieving a connection
	retrieved, exists := registry.GetConnection(conn.ConnectionID)
	assert.True(t, exists)
	assert.Equal(t, conn.SourceID, retrieved.SourceID)
	assert.Equal(t, conn.TargetID, retrieved.TargetID)

	// Test getting connections by client
	conns := registry.GetConnectionsByClient("source1")
	assert.Len(t, conns, 1)
	assert.Equal(t, conn.ConnectionID, conns[0].ConnectionID)

	// Test updating connection status
	success := registry.UpdateConnectionStatus(conn.ConnectionID, StatusNegotiating)
	assert.True(t, success)

	retrieved, _ = registry.GetConnection(conn.ConnectionID)
	assert.Equal(t, StatusNegotiating, retrieved.Status)

	// Test getting connections by status
	statusConns := registry.GetConnectionsByStatus(StatusNegotiating)
	assert.Len(t, statusConns, 1)
	assert.Equal(t, conn.ConnectionID, statusConns[0].ConnectionID)

	// Test updating connection error
	success = registry.UpdateConnectionError(conn.ConnectionID, "test error")
	assert.True(t, success)

	retrieved, _ = registry.GetConnection(conn.ConnectionID)
	assert.Equal(t, StatusFailed, retrieved.Status)
	assert.Equal(t, "test error", retrieved.ErrorMessage)

	// Test updating metadata
	success = registry.UpdateConnectionMetadata(conn.ConnectionID, "test_key", "test_value")
	assert.True(t, success)

	retrieved, _ = registry.GetConnection(conn.ConnectionID)
	assert.Equal(t, "test_value", retrieved.Metadata["test_key"])

	// Test connection stats
	stats := registry.GetConnectionStats()
	assert.Equal(t, 1, stats["total"])
	assert.Equal(t, 1, stats["failed"])
	assert.Equal(t, 0, stats["established"])

	// Test non-existent connection operations
	success = registry.UpdateConnectionStatus("non-existent", StatusEstablished)
	assert.False(t, success)

	success = registry.UpdateConnectionError("non-existent", "error")
	assert.False(t, success)

	success = registry.UpdateConnectionMetadata("non-existent", "key", "value")
	assert.False(t, success)

	// Test removal
	success = registry.RemoveConnection(conn.ConnectionID)
	assert.True(t, success)

	_, exists = registry.GetConnection(conn.ConnectionID)
	assert.False(t, exists)
}

func TestConnectionRegistryCleanup(t *testing.T) {
	logger := utils.NewLogger("test", "debug") // Use debug logging level
	registry := NewConnectionRegistry(logger)
	defer registry.Stop()

	// Create timestamps with explicit time values to avoid any inconsistencies
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)
	threeHoursAgo := now.Add(-3 * time.Hour)
	oneHourAndOneMin := now.Add(-61 * time.Minute)

	// Add connections with different statuses
	connections := []*ConnectionRequest{
		{
			SourceID:     "source1",
			TargetID:     "target1",
			Status:       StatusInitiated,
			Timestamp:    twoHoursAgo,
			LastUpdated:  twoHoursAgo,
			ConnectionID: "conn1",
		},
		{
			SourceID:     "source2",
			TargetID:     "target2",
			Status:       StatusEstablished,
			Timestamp:    threeHoursAgo,
			LastUpdated:  threeHoursAgo,
			ConnectionID: "conn2",
		},
		{
			SourceID:     "source3",
			TargetID:     "target3",
			Status:       StatusFailed,
			Timestamp:    oneHourAndOneMin,
			LastUpdated:  oneHourAndOneMin,
			ConnectionID: "conn3",
		},
		{
			SourceID:     "source4",
			TargetID:     "target4",
			Status:       StatusNegotiating,
			Timestamp:    now,
			LastUpdated:  now,
			ConnectionID: "conn4",
		},
	}

	// Manually add the connections to bypass automatic ID generation
	for _, conn := range connections {
		registry.connections[conn.ConnectionID] = conn
	}

	// Verify connections were added
	assert.Equal(t, 4, len(registry.connections))

	// Run cleanup - with a 30-minute cutoff
	count := registry.CleanupStaleConnections(30 * time.Minute)
	assert.Equal(t, 3, count) // Should remove conn1, conn2, conn3

	// Verify only the recent connection remains
	_, exists := registry.GetConnection("conn4")
	assert.True(t, exists, "Recent connection should still exist")
	assert.Equal(t, 1, len(registry.connections))
}

func TestBackgroundCleanupRoutine(t *testing.T) {
	logger := utils.NewLogger("test", "debug")
	registry := NewConnectionRegistry(logger)

	// Stop the original cleanup goroutine
	registry.Stop()

	// Create a new channel and restart with a very short interval
	registry.cleanupInterval = 50 * time.Millisecond
	registry.stopCleanup = make(chan struct{})

	// Add an explicitly very old connection
	now := time.Now()
	veryOldTime := now.Add(-24 * time.Hour)
	conn := &ConnectionRequest{
		SourceID:     "source-temp",
		TargetID:     "target-temp",
		Status:       StatusInitiated,
		Timestamp:    veryOldTime,
		LastUpdated:  veryOldTime,
		ConnectionID: "old-conn",
	}

	// Directly add to the map
	registry.connections["old-conn"] = conn

	// Start the background cleanup
	go registry.startPeriodicCleanup()

	// Wait for a few cleanup cycles
	time.Sleep(250 * time.Millisecond)

	// Check if connection was removed
	registry.mu.RLock() // Explicitly lock for thread safety
	_, exists := registry.connections["old-conn"]
	registry.mu.RUnlock()
	assert.False(t, exists, "Connection should have been cleaned up by background routine")

	registry.Stop() // Stop the cleanup goroutine
}
