package networking

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockNetworkHandler implements NetworkHandler for testing
type mockNetworkHandler struct {
	initialized     bool
	running         bool
	sentPackets     []*Packet
	receiveCallback func(*Packet) error
	listeningAddrs  []net.Addr
	initializeError error
	startError      error
	stopError       error
	sendError       error
}

func newMockNetworkHandler() *mockNetworkHandler {
	return &mockNetworkHandler{
		sentPackets:    make([]*Packet, 0),
		listeningAddrs: []net.Addr{&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}},
	}
}

func (m *mockNetworkHandler) Initialize(config map[string]interface{}) error {
	m.initialized = true
	return m.initializeError
}

func (m *mockNetworkHandler) Start(ctx context.Context) error {
	m.running = true
	return m.startError
}

func (m *mockNetworkHandler) Stop() error {
	m.running = false
	return m.stopError
}

func (m *mockNetworkHandler) Send(packet *Packet) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentPackets = append(m.sentPackets, packet)
	return nil
}

func (m *mockNetworkHandler) RegisterReceiveCallback(callback func(*Packet) error) {
	m.receiveCallback = callback
}

func (m *mockNetworkHandler) GetListeningAddresses() []net.Addr {
	return m.listeningAddrs
}

// TestConnectionTrackerBasicOperations tests the basic operations of the BaseConnectionTracker
func TestConnectionTrackerBasicOperations(t *testing.T) {
	tracker := NewBaseConnectionTracker()

	// Create test addresses
	sourceAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.2"), Port: 12345}
	targetAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.3"), Port: 54321}

	// Test adding a connection
	connID, err := tracker.AddConnection(sourceAddr, targetAddr, UDP)
	assert.NoError(t, err)
	assert.NotEmpty(t, connID)

	// Test getting a connection
	retrievedSource, retrievedTarget, retrievedType, err := tracker.GetConnection(connID)
	assert.NoError(t, err)
	assert.Equal(t, sourceAddr, retrievedSource)
	assert.Equal(t, targetAddr, retrievedTarget)
	assert.Equal(t, UDP, retrievedType)

	// Test updating connection state
	err = tracker.UpdateConnectionState(connID, string(ConnectionStateEstablished))
	assert.NoError(t, err)

	// Test listing connections
	ids, err := tracker.ListConnections()
	assert.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, connID, ids[0])

	// Test removing a connection
	err = tracker.RemoveConnection(connID)
	assert.NoError(t, err)

	// Verify connection is removed
	_, _, _, err = tracker.GetConnection(connID)
	assert.Error(t, err)

	// Test operations on non-existent connections
	err = tracker.UpdateConnectionState("invalid-id", string(ConnectionStateEstablished))
	assert.Error(t, err)

	err = tracker.RemoveConnection("invalid-id")
	assert.Error(t, err)
}

// TestNetworkFactoryOperations tests the operations of the BaseNetworkFactory
func TestNetworkFactoryOperations(t *testing.T) {
	factory := NewNetworkFactory()

	// Register mock handlers
	factory.RegisterHandler(UDP, func(config map[string]interface{}) (NetworkHandler, error) {
		return newMockNetworkHandler(), nil
	})

	factory.RegisterHandler(TCP, func(config map[string]interface{}) (NetworkHandler, error) {
		return newMockNetworkHandler(), nil
	})

	// Test creating handlers
	udpHandler, err := factory.CreateHandler(UDP, nil)
	assert.NoError(t, err)
	assert.NotNil(t, udpHandler)

	tcpHandler, err := factory.CreateHandler(TCP, nil)
	assert.NoError(t, err)
	assert.NotNil(t, tcpHandler)

	// Test creating handler for unregistered type
	_, err = factory.CreateHandler("invalid", nil)
	assert.Error(t, err)

	// Test creating tracker
	tracker, err := factory.CreateTracker(nil)
	assert.NoError(t, err)
	assert.NotNil(t, tracker)

	// Mock NATPunchStrategy implementation
	mockStrategy := &mockNATPunchStrategy{
		name:     "MockStrategy",
		priority: 10,
		canHandle: func(src, tgt string) bool {
			return true
		},
	}

	// Test registering and retrieving NAT punch strategies
	factory.RegisterNATPunchStrategy(UDP, mockStrategy)

	retrievedStrategy := factory.GetNATPunchStrategy(UDP, "FullCone", "FullCone")
	assert.NotNil(t, retrievedStrategy)
	assert.Equal(t, mockStrategy.GetName(), retrievedStrategy.GetName())

	// Test with non-existent strategy
	nonExistentStrategy := factory.GetNATPunchStrategy(TCP, "FullCone", "FullCone")
	assert.Nil(t, nonExistentStrategy)
}

// mockNATPunchStrategy implements NATPunchStrategy for testing
type mockNATPunchStrategy struct {
	name      string
	priority  int
	canHandle func(string, string) bool
}

func (m *mockNATPunchStrategy) CanHandle(sourceNATType, targetNATType string) bool {
	return m.canHandle(sourceNATType, targetNATType)
}

func (m *mockNATPunchStrategy) InitiatePunch(ctx context.Context, sourceAddr, targetAddr net.Addr, connType ConnectionType) error {
	return nil
}

func (m *mockNATPunchStrategy) GetName() string {
	return m.name
}

func (m *mockNATPunchStrategy) GetPriority() int {
	return m.priority
}
