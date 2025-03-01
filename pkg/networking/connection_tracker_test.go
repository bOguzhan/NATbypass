package networking

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConnectionState tests the connection states and transitions
func TestConnectionState(t *testing.T) {
	tracker := NewBaseConnectionTracker()

	sourceAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.2"), Port: 12345}
	targetAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.3"), Port: 54321}

	// Add a connection
	connID, err := tracker.AddConnection(sourceAddr, targetAddr, UDP)
	assert.NoError(t, err)

	// Test state transitions
	states := []ConnectionState{
		ConnectionStateInitiating,
		ConnectionStateEstablished,
		ConnectionStateFailed,
		ConnectionStateClosed,
	}

	for _, state := range states {
		err := tracker.UpdateConnectionState(connID, string(state))
		assert.NoError(t, err)

		_, _, _, err = tracker.GetConnection(connID)
		assert.NoError(t, err)
	}
}

// TestConcurrentAccess tests concurrent access to the connection tracker
func TestConcurrentAccess(t *testing.T) {
	tracker := NewBaseConnectionTracker()

	// Number of concurrent operations
	numOps := 100

	// Create channels for synchronization
	done := make(chan bool, numOps)

	// Run concurrent operations
	for i := 0; i < numOps; i++ {
		go func(idx int) {
			sourceAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.2"), Port: 10000 + idx}
			targetAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.3"), Port: 20000 + idx}

			// Add connection
			connID, err := tracker.AddConnection(sourceAddr, targetAddr, UDP)
			if err != nil {
				t.Error("Error adding connection:", err)
				done <- true
				return
			}

			// Update state
			err = tracker.UpdateConnectionState(connID, string(ConnectionStateEstablished))
			if err != nil {
				t.Error("Error updating state:", err)
				done <- true
				return
			}

			// Get connection
			_, _, _, err = tracker.GetConnection(connID)
			if err != nil {
				t.Error("Error getting connection:", err)
				done <- true
				return
			}

			// List connections
			_, err = tracker.ListConnections()
			if err != nil {
				t.Error("Error listing connections:", err)
				done <- true
				return
			}

			// Remove connection
			err = tracker.RemoveConnection(connID)
			if err != nil {
				t.Error("Error removing connection:", err)
			}

			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numOps; i++ {
		<-done
	}
}

// TestConnectionTrackerTimeouts tests the timestamps and ensures connections have proper timestamps
func TestConnectionTrackerTimeouts(t *testing.T) {
	tracker := NewBaseConnectionTracker()

	sourceAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.2"), Port: 12345}
	targetAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.3"), Port: 54321}

	// Add connection
	connID, err := tracker.AddConnection(sourceAddr, targetAddr, UDP)
	assert.NoError(t, err)

	// Get the connection from internal map for timestamp checking
	conn, exists := tracker.connections[connID]
	assert.True(t, exists)

	// Check timestamps
	assert.False(t, conn.CreatedAt.IsZero())
	assert.False(t, conn.LastUpdatedAt.IsZero())

	// Ensure initial timestamps are equal
	assert.Equal(t, conn.CreatedAt, conn.LastUpdatedAt)

	// Sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update state and check that LastUpdatedAt changes
	beforeUpdate := conn.LastUpdatedAt
	err = tracker.UpdateConnectionState(connID, string(ConnectionStateEstablished))
	assert.NoError(t, err)

	// Get the updated connection
	conn, exists = tracker.connections[connID]
	assert.True(t, exists)

	// Check that LastUpdatedAt was updated but CreatedAt wasn't
	assert.Equal(t, beforeUpdate, conn.CreatedAt)
	assert.True(t, conn.LastUpdatedAt.After(beforeUpdate))
}
