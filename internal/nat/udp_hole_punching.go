package nat

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bOguzhan/NATbypass/pkg/protocol"
)

const (
	holePunchTimeout   = 30 * time.Second
	holePunchRetries   = 5
	holePunchDelay     = 500 * time.Millisecond
	holePunchKeepAlive = 10 * time.Second
)

// HolePunchingSession represents an active hole punching attempt
type HolePunchingSession struct {
	localAddr      *net.UDPAddr
	remoteAddr     *net.UDPAddr
	established    bool
	conn           *net.UDPConn
	sessionID      string
	lastActivity   time.Time
	mutex          sync.RWMutex
	keepAliveTimer *time.Timer
	done           chan struct{}
}

// UDPHolePuncher handles UDP hole punching operations
type UDPHolePuncher struct {
	sessions  map[string]*HolePunchingSession
	mutex     sync.RWMutex
	localPort int
	baseConn  *net.UDPConn
}

// NewUDPHolePuncher creates a new UDP hole punching manager
func NewUDPHolePuncher(localPort int) (*UDPHolePuncher, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+string(localPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &UDPHolePuncher{
		sessions:  make(map[string]*HolePunchingSession),
		localPort: localPort,
		baseConn:  conn,
	}, nil
}

// InitiateHolePunch starts a hole punching session to a remote peer
func (p *UDPHolePuncher) InitiateHolePunch(remoteAddrStr string, sessionID string) (*HolePunchingSession, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteAddrStr)
	if err != nil {
		return nil, err
	}

	// Create a new connection for this punching session
	localAddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0, // Use any available port
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	// Get the local address after binding
	localAddr = conn.LocalAddr().(*net.UDPAddr)

	session := &HolePunchingSession{
		localAddr:      localAddr,
		remoteAddr:     remoteAddr,
		established:    false,
		conn:           conn,
		sessionID:      sessionID,
		lastActivity:   time.Now(),
		keepAliveTimer: time.NewTimer(holePunchKeepAlive),
		done:           make(chan struct{}),
	}

	p.mutex.Lock()
	p.sessions[sessionID] = session
	p.mutex.Unlock()

	// Start hole punching in background
	go p.doPunchHole(session)

	// Start listener for this session
	go p.listenForSession(session)

	return session, nil
}

// doPunchHole performs the actual hole punching operation
func (p *UDPHolePuncher) doPunchHole(session *HolePunchingSession) {
	// Create punch packet
	packet := &protocol.Packet{
		Type:    protocol.PacketTypeHolePunch,
		Payload: []byte(session.sessionID),
	}

	data, err := packet.Serialize()
	if err != nil {
		log.Printf("Error serializing hole punch packet: %v", err)
		return
	}

	// Send multiple punch packets to increase chances of success
	for i := 0; i < holePunchRetries && !session.IsEstablished(); i++ {
		log.Printf("Sending hole punch packet %d/%d to %s", i+1, holePunchRetries, session.remoteAddr.String())

		_, err = session.conn.WriteToUDP(data, session.remoteAddr)
		if err != nil {
			log.Printf("Error sending hole punch packet: %v", err)
			continue
		}

		time.Sleep(holePunchDelay)
	}

	// Set a timeout for the hole punching attempt
	timer := time.NewTimer(holePunchTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		// If we reach here, the hole punch wasn't acknowledged
		if !session.IsEstablished() {
			log.Printf("Hole punching to %s timed out", session.remoteAddr.String())
			p.CloseSession(session.sessionID)
		}
	case <-session.done:
		// Session was either established or explicitly closed
		return
	}
}

// listenForSession listens for incoming packets on a specific session
func (p *UDPHolePuncher) listenForSession(session *HolePunchingSession) {
	buffer := make([]byte, udpReadBufferSize)

	// Set read deadline to handle timeout
	session.conn.SetReadDeadline(time.Now().Add(holePunchTimeout))

	for {
		n, addr, err := session.conn.ReadFromUDP(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() {
				// Check if session is established despite timeout
				if !session.IsEstablished() {
					log.Printf("Read timeout for session %s", session.sessionID)
					p.CloseSession(session.sessionID)
				}
				return
			}

			log.Printf("Error reading from UDP in session %s: %v", session.sessionID, err)
			continue
		}

		// Process the received packet
		packet, err := protocol.ParsePacket(buffer[:n])
		if err != nil {
			log.Printf("Error parsing packet: %v", err)
			continue
		}

		// Update session state
		session.UpdateActivity()

		// Handle different packet types
		switch packet.Type {
		case protocol.PacketTypeHolePunch:
			log.Printf("Received hole punch from %s", addr.String())
			// Send acknowledgment
			ack := &protocol.Packet{
				Type:    protocol.PacketTypeHolePunchAck,
				Payload: []byte("ok"),
			}
			ackData, _ := ack.Serialize()
			session.conn.WriteToUDP(ackData, addr)

			// Update session with the actual remote address
			p.updateSessionRemoteAddr(session, addr)

		case protocol.PacketTypeHolePunchAck:
			log.Printf("Received hole punch acknowledgment from %s", addr.String())
			// Update session with the actual remote address and mark as established
			p.updateSessionRemoteAddr(session, addr)
			session.SetEstablished(true)

		case protocol.PacketTypeKeepAlive:
			// Just update activity timestamp, which was done above

		case protocol.PacketTypeData:
			// In a real implementation, this would be passed to application layer
			log.Printf("Received %d bytes of data in session %s", len(packet.Payload), session.sessionID)
		}

		// If this was the first packet after timeout, remove deadline
		if session.IsEstablished() {
			session.conn.SetReadDeadline(time.Time{})
		}
	}
}

// updateSessionRemoteAddr updates the remote address of a session
func (p *UDPHolePuncher) updateSessionRemoteAddr(session *HolePunchingSession, addr *net.UDPAddr) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	// Only update if the address is different
	if session.remoteAddr.String() != addr.String() {
		log.Printf("Updating remote address for session %s: %s -> %s",
			session.sessionID, session.remoteAddr.String(), addr.String())
		session.remoteAddr = addr
	}
}

// CloseSession ends a hole punching session
func (p *UDPHolePuncher) CloseSession(sessionID string) {
	p.mutex.Lock()
	session, exists := p.sessions[sessionID]
	if exists {
		delete(p.sessions, sessionID)
	}
	p.mutex.Unlock()

	if exists {
		close(session.done)
		session.conn.Close()
	}
}

// GetSession retrieves a hole punching session by ID
func (p *UDPHolePuncher) GetSession(sessionID string) (*HolePunchingSession, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	session, exists := p.sessions[sessionID]
	return session, exists
}

// IsEstablished checks if the hole punching session is established
func (s *HolePunchingSession) IsEstablished() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.established
}

// SetEstablished updates the established state of the session
func (s *HolePunchingSession) SetEstablished(established bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.established = established

	if established {
		log.Printf("Session %s established with %s", s.sessionID, s.remoteAddr.String())
	}
}

// UpdateActivity updates the last activity timestamp
func (s *HolePunchingSession) UpdateActivity() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lastActivity = time.Now()
}

// GetRemoteAddr returns the remote address for this session
func (s *HolePunchingSession) GetRemoteAddr() *net.UDPAddr {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.remoteAddr
}

// SendData sends data over the established hole punch connection
func (s *HolePunchingSession) SendData(data []byte) error {
	if !s.IsEstablished() {
		return errors.New("session not established")
	}

	packet := &protocol.Packet{
		Type:    protocol.PacketTypeData,
		Payload: data,
	}

	packetData, err := packet.Serialize()
	if err != nil {
		return err
	}

	_, err = s.conn.WriteToUDP(packetData, s.GetRemoteAddr())
	if err == nil {
		s.UpdateActivity()
	}
	return err
}
