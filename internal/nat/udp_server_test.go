package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bOguzhan/NATbypass/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func TestUDPServerBasicOperations(t *testing.T) {
	// Choose a port that's likely to be available for testing
	testPort := 12345
	listenAddr := "127.0.0.1:" + string(testPort)

	// Create a UDP server
	server, err := NewUDPServer(listenAddr)
	assert.NoError(t, err)
	assert.NotNil(t, server)

	// Start the server in a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = server.Start(ctx)
	assert.NoError(t, err)
	defer server.Stop()

	// Let's verify we have no connections initially
	assert.Equal(t, 0, server.GetConnectionCount())

	// We need to wait a bit for the server to fully start
	time.Sleep(100 * time.Millisecond)

	// Test sending registration packet to server
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	assert.NoError(t, err)

	clientConn, err := net.ListenUDP("udp", clientAddr)
	assert.NoError(t, err)
	defer clientConn.Close()

	// Create and send a registration packet
	clientID := "test-client-1"
	regPacket := &protocol.Packet{
		Type:    protocol.PacketTypeRegistration,
		Payload: []byte(clientID),
	}

	serverAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	assert.NoError(t, err)

	packetData, err := regPacket.Serialize()
	assert.NoError(t, err)

	_, err = clientConn.WriteToUDP(packetData, serverAddr)
	assert.NoError(t, err)

	// Read response (should be registration acknowledgment)
	buffer := make([]byte, 4096)
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := clientConn.ReadFromUDP(buffer) // Fixed: removed unused variable `addr`
	assert.NoError(t, err)

	responsePacket, err := protocol.ParsePacket(buffer[:n])
	assert.NoError(t, err)
	assert.Equal(t, protocol.PacketTypeRegistrationAck, responsePacket.Type)

	// After registration, we should have one connection
	assert.Eventually(t, func() bool {
		return server.GetConnectionCount() == 1
	}, 2*time.Second, 100*time.Millisecond)
}

func TestUDPServerHolePunching(t *testing.T) {
	// Choose ports that are likely to be available
	testPort1 := 12346
	testPort2 := 12347

	listenAddr1 := "127.0.0.1:" + string(testPort1)
	listenAddr2 := "127.0.0.1:" + string(testPort2)

	// Create two UDP servers (simulating two different hosts)
	server1, err := NewUDPServer(listenAddr1)
	assert.NoError(t, err)

	server2, err := NewUDPServer(listenAddr2)
	assert.NoError(t, err)

	// Start both servers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = server1.Start(ctx)
	assert.NoError(t, err)
	defer server1.Stop()

	err = server2.Start(ctx)
	assert.NoError(t, err)
	defer server2.Stop()

	time.Sleep(100 * time.Millisecond)

	// Set up clients on both "hosts"
	client1Addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	assert.NoError(t, err)
	client1Conn, err := net.ListenUDP("udp", client1Addr)
	assert.NoError(t, err)
	defer client1Conn.Close()

	client2Addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	assert.NoError(t, err)
	client2Conn, err := net.ListenUDP("udp", client2Addr)
	assert.NoError(t, err)
	defer client2Conn.Close()

	server1Addr, err := net.ResolveUDPAddr("udp", listenAddr1)
	assert.NoError(t, err)

	server2Addr, err := net.ResolveUDPAddr("udp", listenAddr2)
	assert.NoError(t, err)

	// Register both clients
	client1ID := "test-client-1"
	client2ID := "test-client-2"

	// Register client 1 with server 1
	regPacket1 := &protocol.Packet{
		Type:    protocol.PacketTypeRegistration,
		Payload: []byte(client1ID),
	}

	packetData, err := regPacket1.Serialize()
	assert.NoError(t, err)
	_, err = client1Conn.WriteToUDP(packetData, server1Addr)
	assert.NoError(t, err)

	// Read and verify registration acknowledgment
	buffer := make([]byte, 4096)
	client1Conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := client1Conn.ReadFromUDP(buffer)
	assert.NoError(t, err)

	responsePacket, err := protocol.ParsePacket(buffer[:n])
	assert.NoError(t, err)
	assert.Equal(t, protocol.PacketTypeRegistrationAck, responsePacket.Type)

	// Register client 2 with server 2
	regPacket2 := &protocol.Packet{
		Type:    protocol.PacketTypeRegistration,
		Payload: []byte(client2ID),
	}

	packetData, err = regPacket2.Serialize()
	assert.NoError(t, err)
	_, err = client2Conn.WriteToUDP(packetData, server2Addr)
	assert.NoError(t, err)

	// Read and verify registration acknowledgment
	client2Conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err = client2Conn.ReadFromUDP(buffer)
	assert.NoError(t, err)

	responsePacket, err = protocol.ParsePacket(buffer[:n])
	assert.NoError(t, err)
	assert.Equal(t, protocol.PacketTypeRegistrationAck, responsePacket.Type)

	// Client 1 requests hole punching to client 2
	holePunchPacket := &protocol.Packet{
		Type:    protocol.PacketTypeHolePunch,
		Payload: []byte(client2ID),
	}

	packetData, err = holePunchPacket.Serialize()
	assert.NoError(t, err)
	_, err = client1Conn.WriteToUDP(packetData, server1Addr)
	assert.NoError(t, err)

	// Client 1 should receive hole punch response with client 2's address
	client1Conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err = client1Conn.ReadFromUDP(buffer)
	assert.NoError(t, err)

	responsePacket, err = protocol.ParsePacket(buffer[:n])
	assert.NoError(t, err)
	assert.Equal(t, protocol.PacketTypeHolePunchResponse, responsePacket.Type)

	// Client 2 should receive hole punch request with client 1's address
	client2Conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err = client2Conn.ReadFromUDP(buffer)
	assert.NoError(t, err)

	responsePacket, err = protocol.ParsePacket(buffer[:n])
	assert.NoError(t, err)
	assert.Equal(t, protocol.PacketTypeHolePunch, responsePacket.Type)
}

func TestUDPHolePuncher(t *testing.T) {
	// Create hole puncher with a dynamically assigned port
	puncher, err := NewUDPHolePuncher(0) // 0 means OS will assign a port
	assert.NoError(t, err)
	defer puncher.CloseSession("test-session")

	// Test initiating a hole punch to localhost (this is a simplified test)
	localAddr := "127.0.0.1:12348"
	session, err := puncher.InitiateHolePunch(localAddr, "test-session")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.False(t, session.IsEstablished())

	// Test session retrieval
	retrievedSession, exists := puncher.GetSession("test-session")
	assert.True(t, exists)
	assert.Equal(t, session, retrievedSession)

	// Test closing session
	puncher.CloseSession("test-session")
	_, exists = puncher.GetSession("test-session")
	assert.False(t, exists)
}

func TestUDPServerConcurrentOperations(t *testing.T) {
	// Choose a port that's likely to be available for testing
	testPort := 12349
	listenAddr := fmt.Sprintf("127.0.0.1:%d", testPort)

	// Create a UDP server
	server, err := NewUDPServer(listenAddr)
	assert.NoError(t, err)

	// Start the server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = server.Start(ctx)
	assert.NoError(t, err)
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Number of concurrent clients to test with
	clientCount := 5

	var wg sync.WaitGroup
	wg.Add(clientCount)

	serverAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	assert.NoError(t, err)

	// Launch multiple clients concurrently
	for i := 0; i < clientCount; i++ {
		go func(clientNum int) {
			defer wg.Done()

			// Create client
			clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
			assert.NoError(t, err)

			clientConn, err := net.ListenUDP("udp", clientAddr)
			assert.NoError(t, err)
			defer clientConn.Close()

			// Register client
			clientID := "test-client-" + string(clientNum)
			regPacket := &protocol.Packet{
				Type:    protocol.PacketTypeRegistration,
				Payload: []byte(clientID),
			}

			packetData, err := regPacket.Serialize()
			assert.NoError(t, err)

			_, err = clientConn.WriteToUDP(packetData, serverAddr)
			assert.NoError(t, err)

			// Read and verify registration acknowledgment
			buffer := make([]byte, 4096)
			clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, _, err := clientConn.ReadFromUDP(buffer)
			assert.NoError(t, err)

			responsePacket, err := protocol.ParsePacket(buffer[:n])
			assert.NoError(t, err)
			assert.Equal(t, protocol.PacketTypeRegistrationAck, responsePacket.Type)

			// Send keep-alive packet
			keepAlivePacket := &protocol.Packet{
				Type:    protocol.PacketTypeKeepAlive,
				Payload: []byte{},
			}

			packetData, err = keepAlivePacket.Serialize()
			assert.NoError(t, err)

			_, err = clientConn.WriteToUDP(packetData, serverAddr)
			assert.NoError(t, err)

		}(i)
	}

	// Wait for all clients to complete their operations
	wg.Wait()

	// Verify that the server tracked the correct number of connections
	assert.Eventually(t, func() bool {
		return server.GetConnectionCount() == clientCount
	}, 2*time.Second, 100*time.Millisecond)
}
