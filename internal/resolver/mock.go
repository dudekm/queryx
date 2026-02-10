package resolver

import (
	"context"
	"net"
)

// MockResolver is a mock implementation of Resolver for testing
type MockResolver struct {
	Addresses map[string]*Address
	Error     error
}

// NewMockResolver creates a new mock resolver
func NewMockResolver() *MockResolver {
	return &MockResolver{
		Addresses: make(map[string]*Address),
	}
}

// Resolve returns a pre-configured address or error
func (m *MockResolver) Resolve(ctx context.Context, host string, defaultPort int, srvService string) (*Address, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	addr, ok := m.Addresses[host]
	if !ok {
		return nil, &net.DNSError{
			Err:  "no such host",
			Name: host,
		}
	}

	return addr, nil
}
