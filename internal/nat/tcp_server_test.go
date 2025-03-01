package nat_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/bOguzhan/NATbypass/internal/config"
	"github.com/bOguzhan/NATbypass/internal/nat"
)

func TestTCPServer(t *testing.T) {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create TCP server configuration
	tcpConfig := &config.TCPServerConfig{
		Host:              "127.0.0.1",
		Port:              8765, // Use a port that's likely to be available
		ConnectionTimeout: 5,    // Short timeout for testing
		MaxConnections:    10,
		BufferSize:        1024,
	}

	// Initialize server
	server := nat.NewTCPServer(tcpConfig, logger)

	// Start server in a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Start(ctx)
	assert.NoError(t, err, "Server should start without error")
	defer server.Stop()

	// Wait a bit for the server to start
	time.Sleep(100 * time.Millisecond)

	// Test single connection
	t.Run("TestSingleConnection", func(t *testing.T) {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tcpConfig.Host, tcpConfig.Port))
		assert.NoError(t, err, "Should connect to server")
		defer conn.Close()

		// Test sending data
		testData := []byte{0x03, 0x00, 0x01, 0x02} // Keep-alive packet
		_, err = conn.Write(testData)
		assert.NoError(t, err, "Should write to connection")

		// Test receiving response
		response := make([]byte, 2)
		err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		assert.NoError(t, err, "Should set read deadline")

		n, err := conn.Read(response)
		assert.NoError(t, err, "Should read response")
		assert.Equal(t, 2, n, "Response should be 2 bytes")
		assert.Equal(t, byte(0x03), response[0], "First byte should be packet type")
		assert.Equal(t, byte(0x01), response[1], "Second byte should be acknowledgment")

		// Verify connection count
		assert.Equal(t, 1, server.GetActiveConnections(), "Should have 1 active connection")
	})

	// Test multiple concurrent connections
	t.Run("TestMultipleConnections", func(t *testing.T) {
		const numConnections = 5
		var wg sync.WaitGroup
		wg.Add(numConnections)

		for i := 0; i < numConnections; i++ {
			go func(id int) {
				defer wg.Done()

				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tcpConfig.Host, tcpConfig.Port))
				if err != nil {
					t.Errorf("Connection %d failed: %v", id, err)
					return
				}
				defer conn.Close()

				// Send keep-alive
				_, err = conn.Write([]byte{0x03, 0x00, 0x01, 0x02})
				if err != nil {
					t.Errorf("Write on connection %d failed: %v", id, err)
					return
				}

				// Read response
				resp := make([]byte, 2)
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				if _, err := conn.Read(resp); err != nil {
					t.Errorf("Read on connection %d failed: %v", id, err)
					return
				}

				// Keep connection open for a moment
				time.Sleep(500 * time.Millisecond)
			}(i)
		}

		// Wait for all connections to be established
		time.Sleep(200 * time.Millisecond)

		// Check if we have all expected connections
		connCount := server.GetActiveConnections()
		assert.GreaterOrEqual(t, connCount, 1, "Should have at least one active connection")

		// Wait for all goroutines to complete
		wg.Wait()
	})

	// Test connection timeout
	t.Run("TestConnectionTimeout", func(t *testing.T) {
		// Lower timeout for this test
		server.Stop()
		tcpConfig.ConnectionTimeout = 1 // 1 second timeout
		server = nat.NewTCPServer(tcpConfig, logger)
		err := server.Start(ctx)
		assert.NoError(t, err, "Server should restart without error")

		// Establish a connection
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tcpConfig.Host, tcpConfig.Port))
		assert.NoError(t, err, "Should connect to server")
		defer conn.Close()

		// Send initial data
		_, err = conn.Write([]byte{0x03, 0x00, 0x01, 0x02})
		assert.NoError(t, err, "Should write to connection")

		// Read response
		resp := make([]byte, 2)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err = conn.Read(resp)
		assert.NoError(t, err, "Should read response")

		// Verify we have a connection
		assert.Equal(t, 1, server.GetActiveConnections(), "Should have 1 active connection")

		// Wait for timeout and cleanup
		time.Sleep(time.Duration(tcpConfig.ConnectionTimeout+1) * time.Second)

		// Force a cleanup cycle - since server is already of type *nat.TCPServer, no type assertion needed
		server.ForceCleanup() // Direct method call

		time.Sleep(200 * time.Millisecond) // Give a short time for cleanup to complete

		// Connection should be cleaned up by the server
		assert.Equal(t, 0, server.GetActiveConnections(), "Should have 0 active connections after timeout")
	})
}
