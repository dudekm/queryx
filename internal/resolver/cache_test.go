package resolver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetGet(t *testing.T) {
	cache := NewCache(60)
	addr := &Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}

	// Set and get
	cache.Set("example.com", addr)
	result := cache.Get("example.com")

	assert.NotNil(t, result)
	assert.Equal(t, addr.IP, result.IP)
	assert.Equal(t, addr.Port, result.Port)
	assert.Equal(t, addr.Hostname, result.Hostname)
}

func TestCache_GetNonExistent(t *testing.T) {
	cache := NewCache(60)
	result := cache.Get("nonexistent.com")
	assert.Nil(t, result)
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(1) // 1 second TTL
	addr := &Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}

	cache.Set("example.com", addr)

	// Should exist immediately
	result := cache.Get("example.com")
	assert.NotNil(t, result)

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// Should be expired
	result = cache.Get("example.com")
	assert.Nil(t, result)
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(60)
	addr1 := &Address{IP: "127.0.0.1", Port: 25565, Hostname: "example1.com"}
	addr2 := &Address{IP: "127.0.0.2", Port: 25566, Hostname: "example2.com"}

	cache.Set("example1.com", addr1)
	cache.Set("example2.com", addr2)

	// Verify entries exist
	assert.NotNil(t, cache.Get("example1.com"))
	assert.NotNil(t, cache.Get("example2.com"))

	// Clear cache
	cache.Clear()

	// Verify entries are gone
	assert.Nil(t, cache.Get("example1.com"))
	assert.Nil(t, cache.Get("example2.com"))
}

func TestCache_Cleanup(t *testing.T) {
	cache := NewCache(1)
	addr := &Address{IP: "127.0.0.1", Port: 25565, Hostname: "example.com"}

	cache.Set("example.com", addr)

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// Run cleanup
	cache.cleanup()

	// Entry should be removed from internal map
	cache.mu.RLock()
	_, exists := cache.entries["example.com"]
	cache.mu.RUnlock()

	assert.False(t, exists)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := NewCache(60)
	addr := &Address{IP: "127.0.0.1", Port: 25565, Hostname: "example.com"}

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			cache.Set("example.com", addr)
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			result := cache.Get("example.com")
			assert.NotNil(t, result)
			done <- true
		}(i)
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}
}
