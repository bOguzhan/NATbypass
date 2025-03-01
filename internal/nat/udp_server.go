package nat

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/bOguzhan/NATbypass/pkg/protocol"
)

const (
	udpReadBufferSize  = 4096
	udpCleanupInterval = 5 * time.Minute
)

// UDPServer handles UDP connections and NAT traversal operations
type UDPServer struct {
	conn        *net.UDPConn
	listenAddr  string
	connections map[string]*UDPConnection
	mutex       sync.RWMutex
	stopChan    chan struct{}
	packetChan  chan *protocol.Packet
	cleanup     *time.Ticker
	maxIdleTime time.Duration
}

// UDPConnection represents a UDP connection with a peer
type UDPConnection struct {
	peerAddr    *net.UDPAddr
	lastActive  time.Time
	established bool
	clientID    string
}

// NewUDPServer creates a new UDP server instance
func NewUDPServer(listenAddr string) (*UDPServer, error) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	server := &UDPServer{
		conn:        conn,
		listenAddr:  listenAddr,
		connections: make(map[string]*UDPConnection),
		stopChan:    make(chan struct{}),
		packetChan:  make(chan *protocol.Packet, 100),
		cleanup:     time.NewTicker(udpCleanupInterval),
		maxIdleTime: 10 * time.Minute,
	}

	return server, nil
}

// Start begins the UDP server operation
func (s *UDPServer) Start(ctx context.Context) error {
	log.Printf("Starting UDP server on %s", s.listenAddr)

	// Start packet listener
	go s.listenPackets()

	// Start packet processor
	go s.processPackets(ctx)

	// Start connection cleanup routine
	go s.cleanupConnections(ctx)

	return nil
}

// Stop gracefully shuts down the UDP server
func (s *UDPServer) Stop() error {
	close(s.stopChan)
	s.cleanup.Stop()
	return s.conn.Close()
}

// listenPackets continuously reads incoming UDP packets
func (s *UDPServer) listenPackets() {
	buffer := make([]byte, udpReadBufferSize)

	for {
		select {
		case <-s.stopChan:
			return
		default:
			n, addr, err := s.conn.ReadFromUDP(buffer)
			if err != nil {
				if utils.IsClosedNetworkError(err) {
					return
				}
				log.Printf("Error reading from UDP: %v", err)
				continue
			}

			// Copy buffer to prevent data race with the next read
			data := make([]byte, n)
			copy(data, buffer[:n])

			packet, err := protocol.ParsePacket(data)
			if err != nil {
				log.Printf("Error parsing packet from %s: %v", addr.String(), err)
				continue
			}

			packet.SourceAddr = addr
			s.packetChan <- packet
		}
	}
}

// processPackets handles incoming packets based on their type
func (s *UDPServer) processPackets(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case packet := <-s.packetChan:
			s.handlePacket(packet)
		}
	}
}

// handlePacket processes individual packets based on their type
func (s *UDPServer) handlePacket(packet *protocol.Packet) {
	addr := packet.SourceAddr.(*net.UDPAddr)
	addrKey := addr.String()

	switch packet.Type {
	case protocol.PacketTypeRegistration:
		s.handleRegistration(packet, addr)
	case protocol.PacketTypeHolePunch:
		s.handleHolePunch(packet, addr)
	case protocol.PacketTypeData:
		s.handleDataPacket(packet, addr)
	case protocol.PacketTypeKeepAlive:
		s.updateConnectionTimestamp(addrKey)
	default:
		log.Printf("Unknown packet type from %s: %d", addrKey, packet.Type)
	}
}

// handleRegistration processes client registration packets
func (s *UDPServer) handleRegistration(packet *protocol.Packet, addr *net.UDPAddr) {
	clientID := string(packet.Payload)
	addrKey := addr.String()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create or update connection
	s.connections[addrKey] = &UDPConnection{
		peerAddr:    addr,
		lastActive:  time.Now(),
		established: true,
		clientID:    clientID,
	}

	log.Printf("Client %s registered from %s", clientID, addrKey)

	// Send acknowledgment
	response := &protocol.Packet{
		Type:    protocol.PacketTypeRegistrationAck,
		Payload: []byte("registered"),
	}
	s.sendPacket(response, addr)
}

// handleHolePunch processes NAT hole punching packets
func (s *UDPServer) handleHolePunch(packet *protocol.Packet, addr *net.UDPAddr) {
	targetID := string(packet.Payload)
	sourceAddrKey := addr.String()

	// Find target client connection
	var targetAddr *net.UDPAddr
	var targetFound bool

	s.mutex.RLock()
	for _, conn := range s.connections {
		if conn.clientID == targetID {
			targetAddr = conn.peerAddr
			targetFound = true
			break
		}
	}
	s.mutex.RUnlock()

	if !targetFound {
		// Target client not found, send error response
		response := &protocol.Packet{
			Type:    protocol.PacketTypeError,
			Payload: []byte("target client not found"),
		}
		s.sendPacket(response, addr)
		return
	}

	// Send source client's address to target client
	punchRequest := &protocol.Packet{
		Type:    protocol.PacketTypeHolePunch,
		Payload: []byte(sourceAddrKey),
	}
	s.sendPacket(punchRequest, targetAddr)

	// Send target client's address to source client
	response := &protocol.Packet{
		Type:    protocol.PacketTypeHolePunchResponse,
		Payload: []byte(targetAddr.String()),
	}
	s.sendPacket(response, addr)

	log.Printf("Initiated hole punch between %s and %s", sourceAddrKey, targetAddr.String())
}

// handleDataPacket processes data transfer packets
func (s *UDPServer) handleDataPacket(packet *protocol.Packet, addr *net.UDPAddr) {
	// In a real implementation, this might forward the data to the appropriate target
	// For now, we'll just update the connection timestamp
	s.updateConnectionTimestamp(addr.String())
}

// sendPacket sends a packet to the specified address
func (s *UDPServer) sendPacket(packet *protocol.Packet, addr *net.UDPAddr) {
	data, err := packet.Serialize()
	if err != nil {
		log.Printf("Error serializing packet: %v", err)
		return
	}

	_, err = s.conn.WriteToUDP(data, addr)
	if err != nil {
		log.Printf("Error sending packet to %s: %v", addr.String(), err)
	}
}

// updateConnectionTimestamp updates the last active time for a connection
func (s *UDPServer) updateConnectionTimestamp(addrKey string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if conn, exists := s.connections[addrKey]; exists {
		conn.lastActive = time.Now()
	}
}

// cleanupConnections periodically removes idle connections
func (s *UDPServer) cleanupConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-s.cleanup.C:
			s.mutex.Lock()
			now := time.Now()
			for addr, conn := range s.connections {
				if now.Sub(conn.lastActive) > s.maxIdleTime {
					log.Printf("Removing idle connection: %s", addr)
					delete(s.connections, addr)
				}
			}
			s.mutex.Unlock()
		}
	}
}

// GetConnectionCount returns the number of active connections
func (s *UDPServer) GetConnectionCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.connections)
}
