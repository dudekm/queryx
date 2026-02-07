package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// TCPTransport handles TCP communication
type TCPTransport struct {
	bufferSize int
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport() *TCPTransport {
	return &TCPTransport{
		bufferSize: 4096,
	}
}

// SendTCP sends data via TCP and returns the response
func (t *TCPTransport) SendTCP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	// Resolve TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP address %s: %w", addr, err)
	}

	// Create dialer with context
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", tcpAddr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP %s: %w", addr, err)
	}
	defer conn.Close()

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set deadline: %w", err)
		}
	}

	// Send data and measure network latency
	sendTime := time.Now()
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write TCP data: %w", err)
	}

	// Read response - read all available data
	var result bytes.Buffer
	buffer := make([]byte, t.bufferSize)
	firstRead := true

	for {
		// After first read, set a short read timeout to detect end of data
		if !firstRead {
			conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		}

		n, err := conn.Read(buffer)

		// Measure ping on first data received
		if n > 0 && firstRead {
			_ = time.Since(sendTime) // Network latency measured here
		}
		if n > 0 {
			result.Write(buffer[:n])
			firstRead = false
		}

		if err != nil {
			if err == io.EOF {
				// EOF is normal after reading all data
				break
			}
			// For timeout after getting data, that's OK - we got all data
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && result.Len() > 0 {
				break
			}
			// For other errors, check if we have data
			if result.Len() > 0 {
				// Got some data, return it
				break
			}
			return nil, fmt.Errorf("failed to read TCP response: %w", err)
		}
	}

	if result.Len() == 0 {
		return nil, fmt.Errorf("no data received from server")
	}

	return result.Bytes(), nil
}

// SendUDP is not implemented for TCPTransport
func (t *TCPTransport) SendUDP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("UDP not supported by TCPTransport")
}

// SendHTTP is not implemented for TCPTransport
func (t *TCPTransport) SendHTTP(ctx context.Context, url string) ([]byte, error) {
	return nil, fmt.Errorf("HTTP not supported by TCPTransport")
}
