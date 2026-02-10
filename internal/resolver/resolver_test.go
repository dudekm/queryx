package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_String(t *testing.T) {
	addr := &Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}

	assert.Equal(t, "127.0.0.1:25565", addr.String())
}

func TestDefaultResolver_DirectIP(t *testing.T) {
	resolver := NewDefaultResolver(0)
	ctx := context.Background()

	tests := []struct {
		name string
		host string
		port int
	}{
		{"IPv4", "127.0.0.1", 25565},
		{"IPv6", "::1", 25565},
		{"IPv4 with dots", "192.168.1.1", 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := resolver.Resolve(ctx, tt.host, tt.port, "")
			assert.NoError(t, err)
			assert.NotNil(t, addr)
			assert.Equal(t, tt.host, addr.IP)
			assert.Equal(t, tt.port, addr.Port)
			assert.Equal(t, tt.host, addr.Hostname)
		})
	}
}

func TestDefaultResolver_Cache(t *testing.T) {
	resolver := NewDefaultResolver(60)
	ctx := context.Background()

	// First resolve (should be cached after this)
	addr1, err1 := resolver.Resolve(ctx, "127.0.0.1", 25565, "")
	assert.NoError(t, err1)

	// Second resolve (should hit cache)
	addr2, err2 := resolver.Resolve(ctx, "127.0.0.1", 25565, "")
	assert.NoError(t, err2)

	assert.Equal(t, addr1.IP, addr2.IP)
	assert.Equal(t, addr1.Port, addr2.Port)
}

func TestDefaultResolver_NoCacheWhenDisabled(t *testing.T) {
	resolver := NewDefaultResolver(0) // Cache disabled
	assert.Nil(t, resolver.cache)
}

func TestMockResolver_Success(t *testing.T) {
	mock := NewMockResolver()
	mock.Addresses["example.com"] = &Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}

	ctx := context.Background()
	addr, err := mock.Resolve(ctx, "example.com", 25565, "")

	assert.NoError(t, err)
	assert.NotNil(t, addr)
	assert.Equal(t, "127.0.0.1", addr.IP)
	assert.Equal(t, 25565, addr.Port)
}

func TestMockResolver_Error(t *testing.T) {
	mock := NewMockResolver()
	mock.Error = errors.New("dns resolution failed")

	ctx := context.Background()
	addr, err := mock.Resolve(ctx, "example.com", 25565, "")

	assert.Error(t, err)
	assert.Nil(t, addr)
	assert.Equal(t, "dns resolution failed", err.Error())
}

func TestMockResolver_HostNotFound(t *testing.T) {
	mock := NewMockResolver()

	ctx := context.Background()
	addr, err := mock.Resolve(ctx, "nonexistent.com", 25565, "")

	assert.Error(t, err)
	assert.Nil(t, addr)
}

func TestMockResolver_MultipleHosts(t *testing.T) {
	mock := NewMockResolver()
	mock.Addresses["host1.com"] = &Address{IP: "127.0.0.1", Port: 25565, Hostname: "host1.com"}
	mock.Addresses["host2.com"] = &Address{IP: "127.0.0.2", Port: 25566, Hostname: "host2.com"}

	ctx := context.Background()

	addr1, err1 := mock.Resolve(ctx, "host1.com", 25565, "")
	assert.NoError(t, err1)
	assert.Equal(t, "127.0.0.1", addr1.IP)

	addr2, err2 := mock.Resolve(ctx, "host2.com", 25566, "")
	assert.NoError(t, err2)
	assert.Equal(t, "127.0.0.2", addr2.IP)
}

func TestDefaultResolver_Localhost(t *testing.T) {
	resolver := NewDefaultResolver(0)
	ctx := context.Background()

	// Resolve localhost
	addr, err := resolver.Resolve(ctx, "localhost", 25565, "")
	assert.NoError(t, err)
	assert.NotNil(t, addr)
	assert.Equal(t, 25565, addr.Port)
	assert.Equal(t, "localhost", addr.Hostname)
	// IP should be either 127.0.0.1 or ::1
	assert.NotEmpty(t, addr.IP)
}

func TestDefaultResolver_WithCustomPort(t *testing.T) {
	resolver := NewDefaultResolver(0)
	ctx := context.Background()

	tests := []struct {
		name string
		port int
	}{
		{"standard port", 25565},
		{"custom port", 8080},
		{"high port", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := resolver.Resolve(ctx, "127.0.0.1", tt.port, "")
			assert.NoError(t, err)
			assert.Equal(t, tt.port, addr.Port)
		})
	}
}

func TestDefaultResolver_InvalidHost(t *testing.T) {
	resolver := NewDefaultResolver(0)
	ctx := context.Background()

	// Try to resolve non-existent domain (should fail or timeout)
	addr, err := resolver.Resolve(ctx, "this-domain-definitely-does-not-exist-12345.invalid", 25565, "")
	// Either error or timeout expected
	if err == nil {
		// If no error, addr should still be valid with some IP
		assert.NotNil(t, addr)
	} else {
		assert.Nil(t, addr)
	}
}

func TestDefaultResolver_SRVServiceEmpty(t *testing.T) {
	resolver := NewDefaultResolver(0)
	ctx := context.Background()

	// When srvService is empty, should skip SRV lookup and do regular resolution
	addr, err := resolver.Resolve(ctx, "127.0.0.1", 25565, "")
	assert.NoError(t, err)
	assert.NotNil(t, addr)
	assert.Equal(t, "127.0.0.1", addr.IP)
	assert.Equal(t, 25565, addr.Port)
}
