// Package mumble implements the Mumble voice-server UDP ping query.
//
// The client sends a 12-byte request (a 0 type field plus an 8-byte identifier)
// and the server replies with 24 bytes: version, the echoed identifier, the
// current user count, the maximum user count and the allowed bandwidth. All
// integers are big-endian (network byte order).
package mumble

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
	// defaultPort is the default Mumble UDP (and TCP) port.
	defaultPort = 64738

	// pingType is the request type field value for a ping.
	pingType = 0

	// requestLen / responseLen are the fixed packet sizes.
	requestLen  = 12
	responseLen = 24

	// pingIdent is an arbitrary identifier echoed back by the server.
	pingIdent = 0x0123456789ABCDEF
)

// ServerInfo holds the parsed Mumble ping response.
type ServerInfo struct {
	Version          string `json:"version"` // e.g. "1.4.230"
	VersionMajor     int    `json:"versionMajor"`
	VersionMinor     int    `json:"versionMinor"`
	VersionPatch     int    `json:"versionPatch"`
	Users            int    `json:"users"`
	MaxUsers         int    `json:"maxUsers"`
	AllowedBandwidth int    `json:"allowedBandwidth"` // bits per second
}

// Protocol implements the Mumble UDP ping query protocol.
type Protocol struct {
	protocol.BaseProtocol
}

// NewProtocol creates a new Mumble protocol handler.
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{BaseProtocol: protocol.NewBaseProtocol(t, gameName)}
}

// Query sends a ping and parses the 24-byte response.
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	request := buildPing()

	pingStart := time.Now()
	response, err := p.Transport.SendUDP(ctx, addr, request)
	pingMs := int(time.Since(pingStart).Round(time.Millisecond).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("failed to send ping: %w", err)
	}

	info, err := parsePong(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ping response: %w", err)
	}

	return &protocol.QueryResult{
		Online:     true,
		Name:       "", // Mumble's UDP ping does not expose the server name
		NumPlayers: info.Users,
		MaxPlayers: info.MaxUsers,
		Players:    []protocol.Player{},
		Version:    info.Version,
		Ping:       pingMs,
		Raw:        info,
	}, nil
}

// buildPing builds the 12-byte ping request.
func buildPing() []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.BigEndian, uint32(pingType))
	_ = binary.Write(buf, binary.BigEndian, uint64(pingIdent))
	return buf.Bytes()
}

// parsePong parses the 24-byte ping response.
//
// Layout: [uint32 version][uint64 ident][uint32 users][uint32 maxUsers][uint32 bandwidth]
func parsePong(data []byte) (*ServerInfo, error) {
	if len(data) < responseLen {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	version := binary.BigEndian.Uint32(data[0:4])
	major := int((version >> 16) & 0xFF)
	minor := int((version >> 8) & 0xFF)
	patch := int(version & 0xFF)

	return &ServerInfo{
		Version:          fmt.Sprintf("%d.%d.%d", major, minor, patch),
		VersionMajor:     major,
		VersionMinor:     minor,
		VersionPatch:     patch,
		Users:            int(binary.BigEndian.Uint32(data[12:16])),
		MaxUsers:         int(binary.BigEndian.Uint32(data[16:20])),
		AllowedBandwidth: int(binary.BigEndian.Uint32(data[20:24])),
	}, nil
}

// Name returns the protocol name.
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (Mumble UDP ping)", p.GameName)
}

// DefaultPort returns the default Mumble UDP port.
func (p *Protocol) DefaultPort() int { return defaultPort }

// SupportsSRV indicates Mumble does not use SRV records here.
func (p *Protocol) SupportsSRV() bool { return false }

// SRVService returns an empty string (not used).
func (p *Protocol) SRVService() string { return "" }
