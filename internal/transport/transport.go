package transport

import "context"

// Transport defines the interface for network communication
type Transport interface {
	SendUDP(ctx context.Context, addr string, data []byte) ([]byte, error)
	SendTCP(ctx context.Context, addr string, data []byte) ([]byte, error)
	SendHTTP(ctx context.Context, url string) ([]byte, error)
}
