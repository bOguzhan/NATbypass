// pkg/protocol/messages.go
package protocol

import (
	"encoding/json"
	"time"
)

// MessageType defines the type of message being exchanged
type MessageType string

const (
	// TypeRegister is used by clients to register with the Mediatory Server
	TypeRegister MessageType = "register"

	// TypeOffer is used to send an offer to initiate p2p connection
	TypeOffer MessageType = "offer"

	// TypeAnswer is used to respond to an offer
	TypeAnswer MessageType = "answer"

	// TypeICECandidate is used to exchange ICE candidates
	TypeICECandidate MessageType = "ice-candidate"

	// TypeKeepAlive is used to maintain NAT mappings
	TypeKeepAlive MessageType = "keep-alive"
)

// Message represents the base structure for all protocol messages
type Message struct {
	Type      MessageType     `json:"type"`
	ClientID  string          `json:"client_id"`
	TargetID  string          `json:"target_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// NewMessage creates a new message of the specified type
func NewMessage(msgType MessageType, clientID string, payload interface{}) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		ClientID:  clientID,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}, nil
}
