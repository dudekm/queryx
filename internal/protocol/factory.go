package protocol

import (
	"errors"
	"sync"
)

var (
	// ErrUnsupportedGame is returned when a game type is not registered
	ErrUnsupportedGame = errors.New("unsupported game type")
)

// Factory manages protocol registration and retrieval
type Factory struct {
	protocols map[string]Protocol
	mu        sync.RWMutex
}

// NewFactory creates a new protocol factory
func NewFactory() *Factory {
	return &Factory{
		protocols: make(map[string]Protocol),
	}
}

// Register registers a protocol for a specific game type
func (f *Factory) Register(gameType string, protocol Protocol) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.protocols[gameType] = protocol
}

// Get retrieves a protocol for a specific game type
func (f *Factory) Get(gameType string) (Protocol, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	protocol, ok := f.protocols[gameType]
	if !ok {
		return nil, ErrUnsupportedGame
	}

	return protocol, nil
}

// List returns all registered game types
func (f *Factory) List() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]string, 0, len(f.protocols))
	for gameType := range f.protocols {
		types = append(types, gameType)
	}

	return types
}

// Unregister removes a protocol from the factory
func (f *Factory) Unregister(gameType string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.protocols, gameType)
}
