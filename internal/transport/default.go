package transport

import (
	"context"
	"time"
)

// DefaultTransport implements all transport methods (UDP, TCP, HTTP)
type DefaultTransport struct {
	udp  *UDPTransport
	tcp  *TCPTransport
	http *HTTPTransport
}

// NewDefaultTransport creates a new default transport with all protocols
func NewDefaultTransport(timeout time.Duration) *DefaultTransport {
	return &DefaultTransport{
		udp:  NewUDPTransport(),
		tcp:  NewTCPTransport(),
		http: NewHTTPTransport(timeout),
	}
}

// SendUDP sends data via UDP
func (d *DefaultTransport) SendUDP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	return d.udp.SendUDP(ctx, addr, data)
}

// SendTCP sends data via TCP
func (d *DefaultTransport) SendTCP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	return d.tcp.SendTCP(ctx, addr, data)
}

// SendHTTP sends an HTTP GET request
func (d *DefaultTransport) SendHTTP(ctx context.Context, url string) ([]byte, error) {
	return d.http.SendHTTP(ctx, url)
}
