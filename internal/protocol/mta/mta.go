package mta

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// Default MTA server port (ASE query port = game port + 123)
	defaultPort   = 22003
	asePortOffset = 123

	// Magic bytes
	magicEYE1 = "EYE1" // Full server query
	magicEYE2 = "EYE2" // Light query
	magicEYE3 = "EYE3" // XFire query

	// Query type
	queryFull  = 's' // Full ASE protocol query
	queryLight = 'b' // Light query
)

// Protocol implements Multi Theft Auto ASE Query Protocol
type Protocol struct {
	transport transport.Transport
	gameName  string
}

// NewProtocol creates a new MTA protocol handler
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{
		transport: t,
		gameName:  gameName,
	}
}

// Query queries a MTA server using ASE protocol and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	// Parse host and port
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// Calculate ASE query port (game port + 123)
	asePort := port + asePortOffset
	aseAddr := net.JoinHostPort(host, strconv.Itoa(asePort))

	// Build query packet (single byte 's')
	request := []byte{queryFull}

	// Send request and measure ping
	pingStart := time.Now()
	response, err := p.transport.SendUDP(ctx, aseAddr, request)
	ping := time.Since(pingStart)

	if err != nil {
		return nil, fmt.Errorf("failed to send MTA ASE query: %w", err)
	}

	// Parse response
	result, err := parseASEResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MTA ASE response: %w", err)
	}

	result.Ping = ping
	return result, nil
}

// parseASEResponse parses MTA ASE (EYE1) response
func parseASEResponse(data []byte) (*protocol.QueryResult, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Check magic bytes
	magic := string(data[0:4])
	if magic != magicEYE1 && magic != magicEYE2 && magic != magicEYE3 {
		return nil, fmt.Errorf("invalid magic bytes: %q", magic)
	}

	reader := bytes.NewReader(data[4:])

	// Read game type length + "mta"
	gameTypeLen, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read game type length: %w", err)
	}

	gameType := make([]byte, gameTypeLen)
	if _, err := reader.Read(gameType); err != nil {
		return nil, fmt.Errorf("failed to read game type: %w", err)
	}

	// Read port string
	portStrLen, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read port length: %w", err)
	}

	portStr := make([]byte, portStrLen)
	if _, err := reader.Read(portStr); err != nil {
		return nil, fmt.Errorf("failed to read port: %w", err)
	}

	// Read server name
	serverName, err := readLengthPrefixedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read server name: %w", err)
	}

	// Read game type/mode
	gameMode, err := readLengthPrefixedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read game mode: %w", err)
	}

	// Read map name
	mapName, err := readLengthPrefixedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read map name: %w", err)
	}

	// Read version
	version, err := readLengthPrefixedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// Read password flag
	passwordFlag, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read password flag: %w", err)
	}

	// Read player count
	playerCountBytes := make([]byte, 2)
	if _, err := reader.Read(playerCountBytes); err != nil {
		return nil, fmt.Errorf("failed to read player count: %w", err)
	}
	numPlayers := int(playerCountBytes[0]) | (int(playerCountBytes[1]) << 8)

	// Read max players
	maxPlayersBytes := make([]byte, 2)
	if _, err := reader.Read(maxPlayersBytes); err != nil {
		return nil, fmt.Errorf("failed to read max players: %w", err)
	}
	maxPlayers := int(maxPlayersBytes[0]) | (int(maxPlayersBytes[1]) << 8)

	// Build result
	rawData := map[string]interface{}{
		"gamemode": gameMode,
		"gametype": string(gameType[:len(gameType)-1]), // Remove null terminator
	}

	// Try to parse rules (remaining data)
	if reader.Len() > 2 {
		// Read rule count
		ruleCountBytes := make([]byte, 2)
		if _, err := reader.Read(ruleCountBytes); err == nil {
			ruleCount := int(ruleCountBytes[0]) | (int(ruleCountBytes[1]) << 8)

			rules := make(map[string]string)
			for i := 0; i < ruleCount && reader.Len() > 0; i++ {
				key, err := readLengthPrefixedString(reader)
				if err != nil {
					break
				}
				value, err := readLengthPrefixedString(reader)
				if err != nil {
					break
				}
				rules[key] = value
			}

			if len(rules) > 0 {
				rawData["rules"] = rules
			}
		}
	}

	result := &protocol.QueryResult{
		Online:     true,
		Name:       serverName,
		Map:        mapName,
		NumPlayers: numPlayers,
		MaxPlayers: maxPlayers,
		Version:    version,
		Password:   passwordFlag != 0,
		Raw:        rawData,
	}

	return result, nil
}

// readLengthPrefixedString reads a string with 1-byte length prefix (including null terminator)
func readLengthPrefixedString(reader *bytes.Reader) (string, error) {
	length, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	data := make([]byte, length)
	if _, err := reader.Read(data); err != nil {
		return "", err
	}

	// Remove null terminator if present
	if data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}

	return string(data), nil
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (ASE)", p.gameName)
}

// DefaultPort returns the default MTA server port
func (p *Protocol) DefaultPort() int {
	return defaultPort
}

// SupportsSRV indicates that MTA does not use SRV records
func (p *Protocol) SupportsSRV() bool {
	return false
}

// SRVService returns empty string (not used)
func (p *Protocol) SRVService() string {
	return ""
}
