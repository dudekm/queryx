package hytale

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// Default Hytale server port
	defaultPort = 5520

	// Magic bytes for HyQuery protocol
	requestMagic  = "HYQUERY\x00"
	responseMagic = "HYREPLY\x00"

	// Query types
	queryTypeBasic = 0x00
	queryTypeFull  = 0x01
)

// Protocol implements HyQuery protocol for Hytale servers
type Protocol struct {
	transport transport.Transport
	gameName  string
}

// NewProtocol creates a new Hytale protocol handler
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{
		transport: t,
		gameName:  gameName,
	}
}

// Query queries a Hytale server using HyQuery protocol and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	// Build HyQuery request packet (basic query)
	request := buildQueryPacket(queryTypeBasic)

	// Send via UDP and measure network latency (ping)
	pingStart := time.Now()
	response, err := p.transport.SendUDP(ctx, addr, request)
	ping := time.Since(pingStart)

	if err != nil {
		return nil, fmt.Errorf("failed to send HyQuery request: %w", err)
	}

	// Parse response
	result, err := parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HyQuery response: %w", err)
	}

	result.Ping = ping
	return result, nil
}

// buildQueryPacket builds a HyQuery request packet
func buildQueryPacket(queryType byte) []byte {
	packet := make([]byte, 9) // 8 bytes magic + 1 byte query type
	copy(packet[0:8], requestMagic)
	packet[8] = queryType
	return packet
}

// parseResponse parses a HyQuery response packet
func parseResponse(data []byte) (*protocol.QueryResult, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Verify magic bytes
	magic := string(data[0:8])
	if magic != responseMagic {
		return nil, fmt.Errorf("invalid response magic: %q", magic)
	}

	// Read response type
	responseType := data[8]
	offset := 9

	// Read server name
	serverName, offset, err := readString(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read server name: %w", err)
	}

	// Read MOTD
	motd, offset, err := readString(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read MOTD: %w", err)
	}

	// Read online players (uint32 little-endian)
	if offset+4 > len(data) {
		return nil, fmt.Errorf("insufficient data for online players")
	}
	onlinePlayers := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read max players (uint32 little-endian)
	if offset+4 > len(data) {
		return nil, fmt.Errorf("insufficient data for max players")
	}
	maxPlayers := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read port (uint32 little-endian)
	if offset+4 > len(data) {
		return nil, fmt.Errorf("insufficient data for port")
	}
	port := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read version
	version, offset, err := readString(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// Build result
	result := &protocol.QueryResult{
		Online:     true,
		Name:       serverName,
		NumPlayers: int(onlinePlayers),
		MaxPlayers: int(maxPlayers),
		Version:    version,
		Extra:      make(map[string]interface{}),
	}

	// Add MOTD to extra
	if motd != "" {
		result.Extra["motd"] = motd
	}
	result.Extra["port"] = port
	result.Extra["responseType"] = responseType

	// If full query response, parse players and plugins
	if responseType == queryTypeFull && offset < len(data) {
		// Parse players
		players, offset, err := parsePlayers(data, offset)
		if err == nil && len(players) > 0 {
			result.Players = players
		}

		// Parse plugins (if present)
		if offset < len(data) {
			plugins, _, err := parsePlugins(data, offset)
			if err == nil && len(plugins) > 0 {
				result.Extra["plugins"] = plugins
			}
		}
	}

	return result, nil
}

// readString reads a length-prefixed UTF-8 string (uint16 little-endian length)
func readString(data []byte, offset int) (string, int, error) {
	if offset+2 > len(data) {
		return "", offset, fmt.Errorf("insufficient data for string length at offset %d", offset)
	}

	length := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(length) > len(data) {
		return "", offset, fmt.Errorf("insufficient data for string content at offset %d (need %d bytes)", offset, length)
	}

	str := string(data[offset : offset+int(length)])
	offset += int(length)

	return str, offset, nil
}

// parsePlayers parses player list from full query response
func parsePlayers(data []byte, offset int) ([]protocol.Player, int, error) {
	if offset+4 > len(data) {
		return nil, offset, fmt.Errorf("insufficient data for player count")
	}

	playerCount := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	players := make([]protocol.Player, 0, playerCount)

	for i := uint32(0); i < playerCount; i++ {
		// Read player name
		name, newOffset, err := readString(data, offset)
		if err != nil {
			return players, offset, fmt.Errorf("failed to read player %d name: %w", i, err)
		}
		offset = newOffset

		// Skip UUID (16 bytes)
		if offset+16 > len(data) {
			return players, offset, fmt.Errorf("insufficient data for player %d UUID", i)
		}
		offset += 16

		players = append(players, protocol.Player{
			Name: name,
		})
	}

	return players, offset, nil
}

// parsePlugins parses plugin list from full query response
func parsePlugins(data []byte, offset int) ([]string, int, error) {
	if offset+4 > len(data) {
		return nil, offset, fmt.Errorf("insufficient data for plugin count")
	}

	pluginCount := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	plugins := make([]string, 0, pluginCount)

	for i := uint32(0); i < pluginCount; i++ {
		name, newOffset, err := readString(data, offset)
		if err != nil {
			return plugins, offset, fmt.Errorf("failed to read plugin %d name: %w", i, err)
		}
		offset = newOffset

		plugins = append(plugins, name)
	}

	return plugins, offset, nil
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (HyQuery)", p.gameName)
}

// DefaultPort returns the default Hytale server port
func (p *Protocol) DefaultPort() int {
	return defaultPort
}

// SupportsSRV indicates that Hytale does not use SRV records
func (p *Protocol) SupportsSRV() bool {
	return false
}

// SRVService returns empty string (not used)
func (p *Protocol) SRVService() string {
	return ""
}

// buildFullQueryPacket builds a HyQuery full query request packet
func buildFullQueryPacket() []byte {
	return buildQueryPacket(queryTypeFull)
}

// parseResponseBuffer is a helper for testing
func parseResponseBuffer(buf *bytes.Buffer) (*protocol.QueryResult, error) {
	return parseResponse(buf.Bytes())
}
