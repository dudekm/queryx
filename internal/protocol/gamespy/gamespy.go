package gamespy

import (
	"context"
	"fmt"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// Default GameSpy query port
	defaultPort = 2302
)

// Protocol implements GameSpy Query Protocol
// Used by ARMA 2, ARMA 3, DayZ, Day of Dragons, and many other games
type Protocol struct {
	transport transport.Transport
	gameName  string
}

// NewProtocol creates a new GameSpy protocol handler
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{
		transport: t,
		gameName:  gameName,
	}
}

// Query queries a GameSpy server and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	// Build GameSpy query packet
	request := buildQueryPacket()

	// Send via UDP and measure network latency (ping)
	pingStart := time.Now()
	response, err := p.transport.SendUDP(ctx, addr, request)
	ping := time.Since(pingStart)

	if err != nil {
		return nil, fmt.Errorf("failed to send GameSpy query: %w", err)
	}

	// Parse key-value response
	data, err := parseKeyValue(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GameSpy response: %w", err)
	}

	// Build result from parsed data
	result := &protocol.QueryResult{
		Online:     true,
		Name:       getString(data, "hostname"),
		Map:        getString(data, "mapname"),
		NumPlayers: getInt(data, "numplayers"),
		MaxPlayers: getInt(data, "maxplayers"),
		Version:    getString(data, "gamever"),
		Password:   getInt(data, "password") == 1,
		Ping:       ping,
		Raw:        data, // ALL key-value pairs from GameSpy response
	}

	return result, nil
}

// buildQueryPacket builds a GameSpy query packet
func buildQueryPacket() []byte {
	// GameSpy query format: \status\
	// Some servers also accept: \info\
	return []byte("\\status\\")
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (GameSpy)", p.gameName)
}

// DefaultPort returns the default GameSpy port
func (p *Protocol) DefaultPort() int {
	// Use game-specific ports
	return GetDefaultPort(p.gameName)
}

// SupportsSRV indicates that GameSpy does not use SRV records
func (p *Protocol) SupportsSRV() bool {
	return false
}

// SRVService returns empty string (not used)
func (p *Protocol) SRVService() string {
	return ""
}
