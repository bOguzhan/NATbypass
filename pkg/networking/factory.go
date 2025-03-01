package networking

import (
	"fmt"
	"sync"
)

// BaseNetworkFactory implements the NetworkFactory interface
type BaseNetworkFactory struct {
	handlers        map[ConnectionType]func(map[string]interface{}) (NetworkHandler, error)
	punchStrategies map[ConnectionType][]NATPunchStrategy
	mu              sync.RWMutex
}

// NewNetworkFactory creates a new BaseNetworkFactory
func NewNetworkFactory() *BaseNetworkFactory {
	return &BaseNetworkFactory{
		handlers:        make(map[ConnectionType]func(map[string]interface{}) (NetworkHandler, error)),
		punchStrategies: make(map[ConnectionType][]NATPunchStrategy),
	}
}

// RegisterHandler registers a handler constructor for a specific connection type
func (f *BaseNetworkFactory) RegisterHandler(
	connType ConnectionType,
	constructor func(map[string]interface{}) (NetworkHandler, error),
) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.handlers[connType] = constructor
}

// RegisterNATPunchStrategy registers a NAT punch strategy for a connection type
func (f *BaseNetworkFactory) RegisterNATPunchStrategy(connType ConnectionType, strategy NATPunchStrategy) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.punchStrategies[connType]; !exists {
		f.punchStrategies[connType] = []NATPunchStrategy{}
	}

	f.punchStrategies[connType] = append(f.punchStrategies[connType], strategy)
}

// CreateHandler creates a new protocol-specific NetworkHandler
func (f *BaseNetworkFactory) CreateHandler(connType ConnectionType, config map[string]interface{}) (NetworkHandler, error) {
	f.mu.RLock()
	constructor, exists := f.handlers[connType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no handler registered for connection type: %s", connType)
	}

	return constructor(config)
}

// CreateTracker creates a new ConnectionTracker
func (f *BaseNetworkFactory) CreateTracker(config map[string]interface{}) (ConnectionTracker, error) {
	// We use the base tracker implementation for all connection types
	return NewBaseConnectionTracker(), nil
}

// GetNATPunchStrategy returns the appropriate NAT punch strategy
func (f *BaseNetworkFactory) GetNATPunchStrategy(connType ConnectionType, sourceNATType, targetNATType string) NATPunchStrategy {
	f.mu.RLock()
	strategies, exists := f.punchStrategies[connType]
	f.mu.RUnlock()

	if !exists {
		return nil
	}

	// Find the highest priority strategy that can handle this NAT combination
	var bestStrategy NATPunchStrategy
	bestPriority := -1

	for _, strategy := range strategies {
		if strategy.CanHandle(sourceNATType, targetNATType) && strategy.GetPriority() > bestPriority {
			bestStrategy = strategy
			bestPriority = strategy.GetPriority()
		}
	}

	return bestStrategy
}
