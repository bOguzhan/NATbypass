package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

const (
	// Packet header size (type + length fields)
	HeaderSize = 5

	// Maximum packet size
	MaxPacketSize = 65507 // Maximum theoretical UDP packet size
)

// PacketType defines the type of packet
type PacketType byte

const (
	PacketTypeUnknown           PacketType = 0
	PacketTypeRegistration      PacketType = 1
	PacketTypeRegistrationAck   PacketType = 2
	PacketTypeHolePunch         PacketType = 3
	PacketTypeHolePunchResponse PacketType = 4
	PacketTypeHolePunchAck      PacketType = 5
	PacketTypeData              PacketType = 6
	PacketTypeKeepAlive         PacketType = 7
	PacketTypeError             PacketType = 8
)

// Packet represents a protocol packet
type Packet struct {
	Type       PacketType
	Payload    []byte
	SourceAddr net.Addr // Not serialized, used internally
}

// ParsePacket parses a byte slice into a Packet
func ParsePacket(data []byte) (*Packet, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("packet too small")
	}

	packetType := PacketType(data[0])

	// Read payload length (4 bytes)
	payloadLength := binary.BigEndian.Uint32(data[1:5])

	if uint32(len(data)) < HeaderSize+payloadLength {
		return nil, errors.New("packet payload incomplete")
	}

	packet := &Packet{
		Type:    packetType,
		Payload: data[HeaderSize : HeaderSize+payloadLength],
	}

	return packet, nil
}

// Serialize converts a Packet into a byte slice
func (p *Packet) Serialize() ([]byte, error) {
	payloadLen := len(p.Payload)
	if payloadLen > MaxPacketSize-HeaderSize {
		return nil, errors.New("payload too large")
	}

	buf := new(bytes.Buffer)

	// Write packet type (1 byte)
	err := buf.WriteByte(byte(p.Type))
	if err != nil {
		return nil, err
	}

	// Write payload length (4 bytes)
	err = binary.Write(buf, binary.BigEndian, uint32(payloadLen))
	if err != nil {
		return nil, err
	}

	// Write payload
	_, err = buf.Write(p.Payload)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
