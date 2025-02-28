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
	logger := utils.NewLogger("test", "info")
	registry := NewConnectionRegistry(logger)
	defer registry.Stop()

	// The issue is in our test timestamps - they need to be in the past
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	thirtyFiveMinutesAgo := time.Now().Add(-35 * time.Minute)
	twoHoursAgo := time.Now().Add(-2 * time.Hour)

	// Add some connections with different statuses
	connections := []*ConnectionRequest{
		{
			SourceID:    "source1",
			TargetID:    "target1",
			Status:      StatusInitiated,
			Timestamp:   thirtyFiveMinutesAgo,
			LastUpdated: thirtyFiveMinutesAgo, // Add LastUpdated
		},
		{
			SourceID:    "source2",
			TargetID:    "target2",
			Status:      StatusEstablished,
			Timestamp:   oneHourAgo,
			LastUpdated: thirtyFiveMinutesAgo,
		},
		{
			SourceID:    "source3",
			TargetID:    "target3",
			Status:      StatusFailed,
			Timestamp:   twoHoursAgo,
			LastUpdated: twoHoursAgo, // This is 2 hours old now (not just 1)
		},
		{
			SourceID:    "source4",
			TargetID:    "target4",
			Status:      StatusNegotiating,
			Timestamp:   time.Now(),
			LastUpdated: time.Now(),
		},
	}

	for _, conn := range connections {
		err := registry.RegisterConnection(conn)
		assert.NoError(t, err)
	}

	// Verify connections were added
	stats := registry.GetConnectionStats()
	assert.Equal(t, 4, stats["total"])

	// Run cleanup - should remove the stale connections
	count := registry.CleanupStaleConnections(30 * time.Minute)
	assert.Equal(t, 3, count) // Should remove connections 0, 1, and 2

	// Verify only the recent connection remains
	remaining := registry.GetConnectionsByClient("source4")
	assert.Len(t, remaining, 1)

	// Stats should reflect the cleanup
	stats = registry.GetConnectionStats()
	assert.Equal(t, 1, stats["total"])
}

func TestBackgroundCleanupRoutine(t *testing.T) {
	logger := utils.NewLogger("test", "info")
	registry := NewConnectionRegistry(logger)

	// Set a very short cleanup interval for testing
	registry.cleanupInterval = 100 * time.Millisecond

	// Override stopCleanup to create a new channel
	registry.stopCleanup = make(chan struct{})

	// Restart the background routine with the shorter interval
	go registry.startPeriodicCleanup()
	defer registry.Stop()

	// Add a connection that will be cleaned up
	// Make the timestamp 2 hours ago to ensure it gets cleaned up
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	conn := &ConnectionRequest{
		SourceID:    "source-temp",
		TargetID:    "target-temp",
		Status:      StatusInitiated,
		Timestamp:   twoHoursAgo,
		LastUpdated: twoHoursAgo,
	}

	registry.RegisterConnection(conn)

	// Wait longer for cleanup to run (500ms should be enough for ~5 cleanup cycles)
	time.Sleep(500 * time.Millisecond)

	// Check if connection was removed
	_, exists := registry.GetConnection(conn.ConnectionID)
	assert.False(t, exists, "Connection should have been cleaned up by background routine")
}
