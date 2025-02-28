// internal/signaling/messages.go
package signaling

import (
	"sync"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/bOguzhan/NATbypass/pkg/protocol"
)

// MessageQueue stores messages waiting to be delivered to clients
type MessageQueue struct {
	mu       sync.RWMutex
	messages map[string][]protocol.Message // Map client ID to their message queue
	logger   *utils.Logger
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(logger *utils.Logger) *MessageQueue {
	return &MessageQueue{
		messages: make(map[string][]protocol.Message),
		logger:   logger,
	}
}

// AddMessage adds a message to a client's queue
func (q *MessageQueue) AddMessage(clientID string, message protocol.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Set timestamp if not set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Create queue if it doesn't exist
	if _, exists := q.messages[clientID]; !exists {
		q.messages[clientID] = make([]protocol.Message, 0, 10)
	}

	// Add message to queue
	q.messages[clientID] = append(q.messages[clientID], message)

	q.logger.WithFields(map[string]interface{}{
		"client_id": clientID,
		"from":      message.ClientID,
		"type":      message.Type,
	}).Debug("Message queued for client")
}

// GetMessages retrieves and removes messages for a client
func (q *MessageQueue) GetMessages(clientID string) []protocol.Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	messages, exists := q.messages[clientID]
	if !exists || len(messages) == 0 {
		return nil
	}

	// Clear the queue
	delete(q.messages, clientID)

	q.logger.WithFields(map[string]interface{}{
		"client_id": clientID,
		"count":     len(messages),
	}).Debug("Messages retrieved for client")

	return messages
}

// PeekMessages retrieves messages without removing them
func (q *MessageQueue) PeekMessages(clientID string) []protocol.Message {
	q.mu.RLock()
	defer q.mu.RUnlock()

	messages, exists := q.messages[clientID]
	if !exists {
		return nil
	}

	// Return a copy of the messages
	result := make([]protocol.Message, len(messages))
	copy(result, messages)

	return result
}

// CleanupOldMessages removes messages older than maxAge
func (q *MessageQueue) CleanupOldMessages(maxAge time.Duration) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	count := 0

	for clientID, msgs := range q.messages {
		// Filter out old messages
		filtered := make([]protocol.Message, 0, len(msgs))
		for _, msg := range msgs {
			if !msg.Timestamp.Before(cutoff) {
				filtered = append(filtered, msg)
			} else {
				count++
			}
		}

		// Update the queue or remove it if empty
		if len(filtered) > 0 {
			q.messages[clientID] = filtered
		} else {
			delete(q.messages, clientID)
		}
	}

	if count > 0 {
		q.logger.Infof("Cleaned up %d old messages", count)
	}

	return count
}
