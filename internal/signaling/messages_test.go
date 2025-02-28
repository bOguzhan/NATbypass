// internal/signaling/messages_test.go
package signaling

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/bOguzhan/NATbypass/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func TestMessageQueue(t *testing.T) {
	logger := utils.NewLogger("test", "info")
	queue := NewMessageQueue(logger)

	// Test adding a message
	payload, _ := json.Marshal(map[string]interface{}{"sdp": "test offer"})
	msg := protocol.Message{
		Type:      protocol.TypeOffer,
		ClientID:  "client1",
		TargetID:  "client2",
		Timestamp: time.Now(),
		Payload:   json.RawMessage(payload),
	}

	queue.AddMessage("client2", msg)

	// Test peeking messages
	messages := queue.PeekMessages("client2")
	assert.Len(t, messages, 1)
	assert.Equal(t, msg.Type, messages[0].Type)
	assert.Equal(t, msg.ClientID, messages[0].ClientID)

	// Test getting and removing messages
	messages = queue.GetMessages("client2")
	assert.Len(t, messages, 1)
	assert.Equal(t, msg.Type, messages[0].Type)

	// Verify messages were removed
	messages = queue.PeekMessages("client2")
	assert.Empty(t, messages)

	// Test cleanup of old messages
	oldPayload, _ := json.Marshal(map[string]interface{}{"sdp": "old answer"})
	queue.AddMessage("client3", protocol.Message{
		Type:      protocol.TypeAnswer,
		ClientID:  "client3",
		TargetID:  "client4",
		Timestamp: time.Now().Add(-10 * time.Minute),
		Payload:   json.RawMessage(oldPayload),
	})

	count := queue.CleanupOldMessages(5 * time.Minute)
	assert.Equal(t, 1, count)

	messages = queue.PeekMessages("client3")
	assert.Empty(t, messages)
}
