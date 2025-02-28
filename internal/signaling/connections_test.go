// internal/signaling/connections_test.go
package signaling

import (
	"testing"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestConnectionRegistry(t *testing.T) {
	logger := utils.NewLogger("test")
	registry := NewConnectionRegistry(logger)

	// Test registering a connection
	conn := &ConnectionRequest{
		SourceID:  "source1",
		TargetID:  "target1",
		Timestamp: time.Now(),
	}

	err := registry.RegisterConnection(conn)
	assert.NoError(t, err)
	assert.NotEmpty(t, conn.ConnectionID)

	// Test retrieving a connection
	retrieved, exists := registry.GetConnection(conn.ConnectionID)
	assert.True(t, exists)
	assert.Equal(t, conn, retrieved)

	// Test getting connections by client
	conns := registry.GetConnectionsByClient("source1")
	assert.Len(t, conns, 1)
	assert.Equal(t, conn, conns[0])

	// Test cleanup of stale connections
	time.Sleep(10 * time.Millisecond)
	count := registry.CleanupStaleConnections(5 * time.Millisecond)
	assert.Equal(t, 1, count)

	// Verify connection was removed
	_, exists = registry.GetConnection(conn.ConnectionID)
	assert.False(t, exists)
}
