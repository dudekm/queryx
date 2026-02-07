package resolver

import (
	"sync"
	"time"
)

// Cache provides DNS caching with TTL support
type Cache struct {
	entries map[string]*cacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

type cacheEntry struct {
	addr      *Address
	expiresAt time.Time
}

// NewCache creates a new DNS cache with the specified TTL in seconds
func NewCache(ttlSeconds int) *Cache {
	return &Cache{
		entries: make(map[string]*cacheEntry),
		ttl:     time.Duration(ttlSeconds) * time.Second,
	}
}

// Get retrieves an address from the cache
// Returns nil if the entry doesn't exist or has expired
func (c *Cache) Get(host string) *Address {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[host]
	if !ok {
		return nil
	}

	// Check if entry has expired
	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.addr
}

// Set stores an address in the cache
func (c *Cache) Set(host string, addr *Address) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[host] = &cacheEntry{
		addr:      addr,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// cleanup removes expired entries (called periodically)
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for host, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, host)
		}
	}
}
