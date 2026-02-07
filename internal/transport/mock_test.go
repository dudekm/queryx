package transport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockTransport_UDP(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockTransport)
		addr          string
		data          []byte
		expectedResp  []byte
		expectedError error
	}{
		{
			name: "successful UDP response",
			setupMock: func(m *MockTransport) {
				m.UDPResponses["127.0.0.1:25565"] = []byte("response")
			},
			addr:          "127.0.0.1:25565",
			data:          []byte("request"),
			expectedResp:  []byte("response"),
			expectedError: nil,
		},
		{
			name: "UDP error",
			setupMock: func(m *MockTransport) {
				m.UDPError = errors.New("connection refused")
			},
			addr:          "127.0.0.1:25565",
			data:          []byte("request"),
			expectedResp:  nil,
			expectedError: errors.New("connection refused"),
		},
		{
			name: "no response configured",
			setupMock: func(m *MockTransport) {
				// No setup
			},
			addr:          "127.0.0.1:25565",
			data:          []byte("request"),
			expectedResp:  []byte{},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockTransport()
			tt.setupMock(mock)

			ctx := context.Background()
			resp, err := mock.SendUDP(ctx, tt.addr, tt.data)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestMockTransport_TCP(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockTransport)
		addr          string
		data          []byte
		expectedResp  []byte
		expectedError error
	}{
		{
			name: "successful TCP response",
			setupMock: func(m *MockTransport) {
				m.TCPResponses["127.0.0.1:25565"] = []byte("tcp response")
			},
			addr:          "127.0.0.1:25565",
			data:          []byte("request"),
			expectedResp:  []byte("tcp response"),
			expectedError: nil,
		},
		{
			name: "TCP error",
			setupMock: func(m *MockTransport) {
				m.TCPError = errors.New("connection timeout")
			},
			addr:          "127.0.0.1:25565",
			data:          []byte("request"),
			expectedResp:  nil,
			expectedError: errors.New("connection timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockTransport()
			tt.setupMock(mock)

			ctx := context.Background()
			resp, err := mock.SendTCP(ctx, tt.addr, tt.data)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestMockTransport_HTTP(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockTransport)
		url           string
		expectedResp  []byte
		expectedError error
	}{
		{
			name: "successful HTTP response",
			setupMock: func(m *MockTransport) {
				m.HTTPResponses["http://example.com"] = []byte("http response")
			},
			url:           "http://example.com",
			expectedResp:  []byte("http response"),
			expectedError: nil,
		},
		{
			name: "HTTP error",
			setupMock: func(m *MockTransport) {
				m.HTTPError = errors.New("404 not found")
			},
			url:           "http://example.com",
			expectedResp:  nil,
			expectedError: errors.New("404 not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockTransport()
			tt.setupMock(mock)

			ctx := context.Background()
			resp, err := mock.SendHTTP(ctx, tt.url)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestMockTransport_Latency(t *testing.T) {
	mock := NewMockTransport()
	mock.Latency = 50 * time.Millisecond
	mock.UDPResponses["127.0.0.1:25565"] = []byte("delayed response")

	ctx := context.Background()
	start := time.Now()
	resp, err := mock.SendUDP(ctx, "127.0.0.1:25565", []byte("request"))
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, []byte("delayed response"), resp)
	assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
}

func TestMockTransport_ContextCancellation(t *testing.T) {
	mock := NewMockTransport()
	mock.Latency = 1 * time.Second
	mock.UDPResponses["127.0.0.1:25565"] = []byte("response")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := mock.SendUDP(ctx, "127.0.0.1:25565", []byte("request"))
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}
