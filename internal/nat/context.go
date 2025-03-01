package nat

import (
	"time"

	"github.com/bOguzhan/NATbypass/internal/discovery"
)

// TraversalContext contains information needed for NAT traversal
type TraversalContext struct {
	// Local peer information
	LocalID      string
	LocalNATType discovery.NATType
	LocalAddr    string

	// Remote peer information
	RemoteID      string
	RemoteNATType discovery.NATType
	RemoteAddr    string

	// Configuration
	PreferredProtocol string
	Timeout           time.Duration
	MaxRetries        int

	// Callback functions for events
	OnStateChange func(state TraversalState)
	OnLogMessage  func(level string, message string)
}

// TraversalState represents the current state of a traversal attempt
type TraversalState string

const (
	// TraversalInitialized means the traversal attempt is initialized but not started
	TraversalInitialized TraversalState = "initialized"

	// TraversalInProgress means the traversal attempt is in progress
	TraversalInProgress TraversalState = "in-progress"

	// TraversalSucceeded means the traversal attempt succeeded
	TraversalSucceeded TraversalState = "succeeded"

	// TraversalFailed means the traversal attempt failed
	TraversalFailed TraversalState = "failed"

	// TraversalCancelled means the traversal attempt was cancelled
	TraversalCancelled TraversalState = "cancelled"
)

// NewTraversalContext creates a new traversal context with default values
func NewTraversalContext(localID, remoteID string) *TraversalContext {
	return &TraversalContext{
		LocalID:           localID,
		RemoteID:          remoteID,
		LocalNATType:      discovery.NATUnknown,
		RemoteNATType:     discovery.NATUnknown,
		Timeout:           30 * time.Second,
		MaxRetries:        5,
		PreferredProtocol: "", // No preference by default
		OnStateChange:     func(state TraversalState) {},
		OnLogMessage:      func(level, message string) {},
	}
}
