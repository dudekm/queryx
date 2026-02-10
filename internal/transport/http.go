package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPTransport handles HTTP communication
type HTTPTransport struct {
	client *http.Client
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(timeout time.Duration) *HTTPTransport {
	return &HTTPTransport{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SendHTTP sends an HTTP GET request and returns the response body
func (t *HTTPTransport) SendHTTP(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", "QueryX/1.0")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %w", err)
	}

	return body, nil
}

// SendUDP is not implemented for HTTPTransport
func (t *HTTPTransport) SendUDP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("UDP not supported by HTTPTransport")
}

// SendTCP is not implemented for HTTPTransport
func (t *HTTPTransport) SendTCP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("TCP not supported by HTTPTransport")
}
