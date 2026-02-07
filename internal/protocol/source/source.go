package source

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// Magic header for all Source Engine packets (0xFFFFFFFF as int32 = -1)
	headerMagic = -1

	// Packet types
	a2sInfoRequest  = 0x54 // 'T'
	a2sInfoResponse = 0x49 // 'I'
	a2sChallenge    = 0x41 // 'A' - Challenge response

	// Default port for Source Engine games
	defaultPort = 27015
)

// Protocol implements Source Engine Query Protocol (A2S)
// Used by CS 1.6, CS:Source, CS2, TF2, Garry's Mod, etc.
type Protocol struct {
	transport transport.Transport
	gameName  string // "CS 1.6", "CS2", etc.
}

// NewProtocol creates a new Source Engine protocol handler
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{
		transport: t,
		gameName:  gameName,
	}
}

// Query queries a Source Engine server and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	// Build A2S_INFO request packet (with default challenge -1)
	request := p.buildA2SInfoRequestWithChallenge(-1)

	// Send via UDP and measure network latency (ping)
	pingStart := time.Now()
	response, err := p.transport.SendUDP(ctx, addr, request)
	ping := time.Since(pingStart)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check if server requires challenge
	if len(response) >= 5 {
		// Read packet type (after 4-byte header)
		packetType := response[4]

		if packetType == a2sChallenge {
			// Server sent challenge, extract it and retry
			if len(response) < 9 {
				return nil, fmt.Errorf("invalid challenge response length")
			}

			// Challenge is 4 bytes after header and packet type
			challenge := int32(binary.LittleEndian.Uint32(response[5:9]))

			// Resend request with challenge and measure ping for actual data response
			request = p.buildA2SInfoRequestWithChallenge(challenge)
			pingStart = time.Now()
			response, err = p.transport.SendUDP(ctx, addr, request)
			ping = time.Since(pingStart) // Update ping with actual data response time

			if err != nil {
				return nil, fmt.Errorf("failed to send challenge request: %w", err)
			}
		}
	}

	// Parse response
	result, err := p.parseA2SInfoResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result.Raw = response
	result.Ping = ping // Set network latency
	return result, nil
}

// buildA2SInfoRequestWithChallenge builds an A2S_INFO request packet with challenge
func (p *Protocol) buildA2SInfoRequestWithChallenge(challenge int32) []byte {
	buf := &bytes.Buffer{}

	// Write magic header (0xFFFFFFFF)
	binary.Write(buf, binary.LittleEndian, int32(headerMagic))

	// Write packet type (0x54 = 'T')
	buf.WriteByte(a2sInfoRequest)

	// Write payload: "Source Engine Query\0"
	buf.WriteString("Source Engine Query")
	buf.WriteByte(0x00)

	// Write challenge (4 bytes)
	binary.Write(buf, binary.LittleEndian, challenge)

	return buf.Bytes()
}

// parseA2SInfoResponse parses an A2S_INFO response packet
func (p *Protocol) parseA2SInfoResponse(data []byte) (*protocol.QueryResult, error) {
	reader := bytes.NewReader(data)

	// Read and verify magic header
	var header int32
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if header != headerMagic {
		return nil, fmt.Errorf("invalid header magic: %x", header)
	}

	// Read packet type
	packetType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read packet type: %w", err)
	}

	if packetType != a2sInfoResponse {
		return nil, fmt.Errorf("unexpected packet type: %x", packetType)
	}

	// Read protocol version
	protocolVersion, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol version: %w", err)
	}

	// Read null-terminated strings
	serverName, err := readNullTerminatedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read server name: %w", err)
	}

	mapName, err := readNullTerminatedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read map name: %w", err)
	}

	folder, err := readNullTerminatedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read folder: %w", err)
	}

	game, err := readNullTerminatedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read game: %w", err)
	}

	// Read game ID (2 bytes)
	var gameID uint16
	if err := binary.Read(reader, binary.LittleEndian, &gameID); err != nil {
		return nil, fmt.Errorf("failed to read game ID: %w", err)
	}

	// Read player counts
	players, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read players: %w", err)
	}

	maxPlayers, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read max players: %w", err)
	}

	bots, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read bots: %w", err)
	}

	// Read server type
	serverType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read server type: %w", err)
	}

	// Read environment
	environment, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read environment: %w", err)
	}

	// Read visibility
	visibility, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read visibility: %w", err)
	}

	// Read VAC
	vac, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read VAC: %w", err)
	}

	// Read version string
	version, err := readNullTerminatedString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// Build result
	result := &protocol.QueryResult{
		Online:     true,
		Name:       serverName,
		Map:        mapName,
		NumPlayers: int(players),
		MaxPlayers: int(maxPlayers),
		Bots:       int(bots),
		Version:    version,
		Password:   visibility != 0, // visibility 0 = public, 1 = private/password
		Extra:      make(map[string]interface{}),
	}

	// Add extra information
	result.Extra["protocol"] = protocolVersion
	result.Extra["game"] = game
	result.Extra["folder"] = folder
	result.Extra["gameID"] = gameID
	result.Extra["serverType"] = string(serverType)
	result.Extra["environment"] = string(environment)
	result.Extra["visibility"] = visibility == 0
	result.Extra["vac"] = vac == 1

	return result, nil
}

// readNullTerminatedString reads a null-terminated string from the reader
func readNullTerminatedString(r io.Reader) (string, error) {
	var buf bytes.Buffer
	b := make([]byte, 1)

	for {
		if _, err := r.Read(b); err != nil {
			return "", err
		}
		if b[0] == 0x00 {
			break
		}
		buf.WriteByte(b[0])
	}

	return buf.String(), nil
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (Source Engine)", p.gameName)
}

// DefaultPort returns the default Source Engine port
func (p *Protocol) DefaultPort() int {
	return defaultPort
}

// SupportsSRV indicates that Source Engine does not use SRV records
func (p *Protocol) SupportsSRV() bool {
	return false
}

// SRVService returns empty string (not used)
func (p *Protocol) SRVService() string {
	return ""
}
