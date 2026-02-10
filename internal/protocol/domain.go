package protocol

import "time"

// QueryResult represents the result of a server query (Domain Entity)
// Single source of truth - used by all protocols and application layers
type QueryResult struct {
	Online     bool          `json:"online"`
	Name       string        `json:"name"`
	Map        string        `json:"map"`        // Map name (always present, empty string if not available)
	NumPlayers int           `json:"numPlayers"` // Current number of players online
	MaxPlayers int           `json:"maxPlayers"` // Maximum number of players
	Players    []Player      `json:"players"`    // Detailed list of players (empty array if not available)
	Bots       int           `json:"bots"`       // Number of bots (0 if not available)
	Type       string        `json:"type"`       // Server type (minecraft, cs2, etc.)
	Version    string        `json:"version"`    // Server version (empty string if not available)
	Ping       time.Duration `json:"ping"`
	Connect    string        `json:"connect"`       // Connection string (host:port)
	Password   bool          `json:"password"`      // Whether server requires password (false if not available)
	Raw        interface{}   `json:"raw,omitempty"` // ALL protocol-specific data (full response, game tags, mods, etc.)
}

// Player represents a single player on the server (Domain Entity)
type Player struct {
	Name     string        `json:"name"`
	Score    int           `json:"score,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}
