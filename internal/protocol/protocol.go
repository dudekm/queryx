package protocol

import (
	"context"
	"time"
)

// QueryResult represents the result of a server query
// This is defined here to avoid import cycles
type QueryResult struct {
	Online     bool
	Name       string
	Map        string
	NumPlayers int      // Current number of players online
	MaxPlayers int      // Maximum number of players
	Players    []Player // Detailed list of players (if available)
	Bots       int      // Number of bots
	Type       string   // Server type (minecraft, cs2, etc.)
	Version    string   // Server version
	Ping       time.Duration
	Connect    string // Connection string (host:port)
	Password   bool   // Whether server requires password
	Extra      map[string]interface{}
	Raw        []byte
	QueriedAt  time.Time
}

// Player represents a single player on the server
type Player struct {
	Name     string
	Score    int
	Duration time.Duration
}

// Protocol defines the interface that all game server protocols must implement
type Protocol interface {
	// Query sends a query to the server and returns the parsed result
	Query(ctx context.Context, addr string) (*QueryResult, error)

	// Name returns the protocol name
	Name() string

	// DefaultPort returns the default port for this protocol
	DefaultPort() int

	// SupportsSRV indicates if this protocol supports SRV record lookup
	SupportsSRV() bool

	// SRVService returns the SRV service name (e.g., "minecraft")
	SRVService() string
}
