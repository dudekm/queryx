package transport

import (
	"context"
	"time"
)

// MockTransport is a mock implementation of Transport for testing
type MockTransport struct {
	UDPResponses     map[string][]byte
	TCPResponses     map[string][]byte
	HTTPResponses    map[string][]byte
	UDPError         error
	TCPError         error
	HTTPError        error
	Latency          time.Duration
	UDPResponseQueue map[string][][]byte // For multiple sequential responses
	udpCallCount     map[string]int
}

// NewMockTransport creates a new mock transport
func NewMockTransport() *MockTransport {
	return &MockTransport{
		UDPResponses:     make(map[string][]byte),
		TCPResponses:     make(map[string][]byte),
		HTTPResponses:    make(map[string][]byte),
		UDPResponseQueue: make(map[string][][]byte),
		udpCallCount:     make(map[string]int),
	}
}

// SendUDP simulates sending UDP data
func (m *MockTransport) SendUDP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	if m.Latency > 0 {
		select {
		case <-time.After(m.Latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.UDPError != nil {
		return nil, m.UDPError
	}

	// Check if we have a response queue for sequential responses
	if queue, ok := m.UDPResponseQueue[addr]; ok && len(queue) > 0 {
		callIdx := m.udpCallCount[addr]
		m.udpCallCount[addr]++

		if callIdx < len(queue) {
			return queue[callIdx], nil
		}
		// If we've exhausted the queue, return the last response
		return queue[len(queue)-1], nil
	}

	if response, ok := m.UDPResponses[addr]; ok {
		return response, nil
	}

	return []byte{}, nil
}

// SendTCP simulates sending TCP data
func (m *MockTransport) SendTCP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	if m.Latency > 0 {
		select {
		case <-time.After(m.Latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.TCPError != nil {
		return nil, m.TCPError
	}

	if response, ok := m.TCPResponses[addr]; ok {
		return response, nil
	}

	return []byte{}, nil
}

// SendHTTP simulates sending HTTP request
func (m *MockTransport) SendHTTP(ctx context.Context, url string) ([]byte, error) {
	if m.Latency > 0 {
		select {
		case <-time.After(m.Latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.HTTPError != nil {
		return nil, m.HTTPError
	}

	if response, ok := m.HTTPResponses[url]; ok {
		return response, nil
	}

	return []byte{}, nil
}
