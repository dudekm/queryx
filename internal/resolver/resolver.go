package resolver

import (
	"context"
	"fmt"
	"net"
	"strconv"
)

// Address represents a resolved network address
type Address struct {
	IP       string
	Port     int
	Hostname string
}

// String returns the address in host:port format
func (a *Address) String() string {
	return net.JoinHostPort(a.IP, strconv.Itoa(a.Port))
}

// Resolver defines the interface for DNS resolution
type Resolver interface {
	Resolve(ctx context.Context, host string, defaultPort int, srvService string) (*Address, error)
}

// DefaultResolver implements DNS resolution with SRV record support and caching
type DefaultResolver struct {
	cache    *Cache
	resolver *net.Resolver
}

// NewDefaultResolver creates a new default resolver with optional cache TTL
func NewDefaultResolver(cacheTTL int) *DefaultResolver {
	var cache *Cache
	if cacheTTL > 0 {
		cache = NewCache(cacheTTL)
	}

	return &DefaultResolver{
		cache:    cache,
		resolver: net.DefaultResolver,
	}
}

// Resolve resolves a hostname to an IP address with port
// Resolution logic:
// 1. Check if host is already an IP address
// 2. Check cache if enabled
// 3. Try SRV lookup if srvService is provided
// 4. Fall back to A/AAAA lookup
// 5. Cache the result if caching is enabled
func (r *DefaultResolver) Resolve(ctx context.Context, host string, defaultPort int, srvService string) (*Address, error) {
	// Check if host is already an IP address
	if ip := net.ParseIP(host); ip != nil {
		return &Address{
			IP:       host,
			Port:     defaultPort,
			Hostname: host,
		}, nil
	}

	// Check cache
	if r.cache != nil {
		if addr := r.cache.Get(host); addr != nil {
			return addr, nil
		}
	}

	var addr *Address
	var err error

	// Try SRV lookup if service is provided
	if srvService != "" {
		addr, err = r.resolveSRV(ctx, host, srvService, defaultPort)
		if err == nil && addr != nil {
			// Cache and return SRV result
			if r.cache != nil {
				r.cache.Set(host, addr)
			}
			return addr, nil
		}
	}

	// Fall back to A/AAAA lookup
	addr, err = r.resolveHost(ctx, host, defaultPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %s: %w", host, err)
	}

	// Cache the result
	if r.cache != nil {
		r.cache.Set(host, addr)
	}

	return addr, nil
}

// resolveSRV performs SRV record lookup
func (r *DefaultResolver) resolveSRV(ctx context.Context, host string, service string, defaultPort int) (*Address, error) {
	_, addrs, err := r.resolver.LookupSRV(ctx, service, "tcp", host)
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("no SRV records found")
	}

	// Use the first SRV record
	srvAddr := addrs[0]

	// Resolve the SRV target to an IP
	ips, err := r.resolver.LookupHost(ctx, srvAddr.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SRV target %s: %w", srvAddr.Target, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IPs found for SRV target %s", srvAddr.Target)
	}

	return &Address{
		IP:       ips[0],
		Port:     int(srvAddr.Port),
		Hostname: host,
	}, nil
}

// resolveHost performs A/AAAA record lookup
func (r *DefaultResolver) resolveHost(ctx context.Context, host string, defaultPort int) (*Address, error) {
	ips, err := r.resolver.LookupHost(ctx, host)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IPs found for host %s", host)
	}

	return &Address{
		IP:       ips[0],
		Port:     defaultPort,
		Hostname: host,
	}, nil
}
