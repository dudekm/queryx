package cfxre

import "strings"

// Endpoint is an immutable value object describing a CFX.re server's HTTP base
// address. It normalizes a raw "host:port" (or full URL) address by ensuring a
// scheme is present and no trailing slash remains, then exposes the well-known
// CFX.re data endpoints.
//
// As a value object it has no identity: two Endpoints with the same normalized
// base URL are equal, and instances are safe to copy and share.
type Endpoint struct {
	baseURL string
}

// NewEndpoint builds an Endpoint from a raw address. Addresses without an
// explicit http:// or https:// scheme default to http:// (CFX.re servers
// expose their query endpoints over plain HTTP by default).
func NewEndpoint(addr string) Endpoint {
	base := strings.TrimSpace(addr)
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	base = strings.TrimRight(base, "/")
	return Endpoint{baseURL: base}
}

// Base returns the normalized base URL (scheme + host[:port], no trailing slash).
func (e Endpoint) Base() string { return e.baseURL }

// Info returns the URL of the /info.json endpoint (server vars, resources).
func (e Endpoint) Info() string { return e.baseURL + "/info.json" }

// Players returns the URL of the /players.json endpoint (connected players).
func (e Endpoint) Players() string { return e.baseURL + "/players.json" }

// Dynamic returns the URL of the /dynamic.json endpoint (live player counts).
func (e Endpoint) Dynamic() string { return e.baseURL + "/dynamic.json" }
