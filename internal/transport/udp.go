package transport

import (
	"context"
	"fmt"
	"net"
)

// UDPTransport handles UDP communication
type UDPTransport struct {
	bufferSize int
}

// NewUDPTransport creates a new UDP transport
func NewUDPTransport() *UDPTransport {
	return &UDPTransport{
		bufferSize: 4096,
	}
}

// SendUDP sends data via UDP and returns the response
func (t *UDPTransport) SendUDP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address %s: %w", addr, err)
	}

	// Create dialer with context
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "udp", udpAddr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP %s: %w", addr, err)
	}
	defer conn.Close()

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set deadline: %w", err)
		}
	}

	// Send data
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write UDP data: %w", err)
	}

	// Read response
	buffer := make([]byte, t.bufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read UDP response: %w", err)
	}

	return buffer[:n], nil
}

// SendTCP is not implemented for UDPTransport
func (t *UDPTransport) SendTCP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("TCP not supported by UDPTransport")
}

// SendHTTP is not implemented for UDPTransport
func (t *UDPTransport) SendHTTP(ctx context.Context, url string) ([]byte, error) {
	return nil, fmt.Errorf("HTTP not supported by UDPTransport")
}
