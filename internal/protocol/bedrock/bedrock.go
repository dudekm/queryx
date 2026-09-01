// Package bedrock implements the Minecraft Bedrock Edition server query using
// the RakNet "unconnected ping/pong" exchange over UDP.
package bedrock

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// defaultPort is the default Bedrock (RakNet) UDP port.
	defaultPort = 19132

	// idUnconnectedPing / idUnconnectedPong are the RakNet packet IDs.
	idUnconnectedPing = 0x01
	idUnconnectedPong = 0x1C

	// clientGUID is an arbitrary client identifier sent with the ping.
	clientGUID = 0x0000000012345678
)

// rakNetMagic is the fixed 16-byte "offline message data ID" RakNet uses to
// recognise unconnected packets.
var rakNetMagic = []byte{
	0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE,
	0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78,
}

// ServerInfo holds all fields parsed from the Bedrock pong MOTD string.
// The MOTD is a ';'-separated string; every field is preserved here (and in
// Raw) so nothing protocol-specific is lost.
type ServerInfo struct {
	Edition         string   `json:"edition"`
	MOTDLine1       string   `json:"motdLine1"`
	ProtocolVersion string   `json:"protocolVersion"`
	VersionName     string   `json:"versionName"`
	PlayerCount     int      `json:"playerCount"`
	MaxPlayers      int      `json:"maxPlayers"`
	ServerID        string   `json:"serverId"`
	MOTDLine2       string   `json:"motdLine2"`
	GameMode        string   `json:"gameMode"`
	GameModeNumeric string   `json:"gameModeNumeric"`
	PortIPv4        string   `json:"portIPv4"`
	PortIPv6        string   `json:"portIPv6"`
	Fields          []string `json:"fields"` // raw ';'-split fields
}

// Protocol implements the Minecraft Bedrock (RakNet) query protocol.
type Protocol struct {
	protocol.BaseProtocol
}

// NewProtocol creates a new Bedrock protocol handler.
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{BaseProtocol: protocol.NewBaseProtocol(t, gameName)}
}

// Query sends an unconnected ping and parses the pong response.
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	request := buildUnconnectedPing()

	pingStart := time.Now()
	response, err := p.Transport.SendUDP(ctx, addr, request)
	pingMs := int(time.Since(pingStart).Round(time.Millisecond).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("failed to send unconnected ping: %w", err)
	}

	info, err := parsePong(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pong: %w", err)
	}

	name := info.MOTDLine1
	if info.MOTDLine2 != "" {
		name = strings.TrimSpace(info.MOTDLine1 + " " + info.MOTDLine2)
	}

	return &protocol.QueryResult{
		Online:     true,
		Name:       name,
		Map:        "", // Bedrock does not report a level/map name
		NumPlayers: info.PlayerCount,
		MaxPlayers: info.MaxPlayers,
		Players:    []protocol.Player{},
		Version:    info.VersionName,
		Ping:       pingMs,
		Raw:        info,
	}, nil
}

// buildUnconnectedPing builds the 0x01 unconnected ping packet.
func buildUnconnectedPing() []byte {
	buf := &bytes.Buffer{}
	buf.WriteByte(idUnconnectedPing)
	// Time since start (any monotonic-ish value); the server echoes it back.
	_ = binary.Write(buf, binary.BigEndian, uint64(time.Now().UnixMilli()))
	buf.Write(rakNetMagic)
	_ = binary.Write(buf, binary.BigEndian, uint64(clientGUID))
	return buf.Bytes()
}

// parsePong parses a 0x1C unconnected pong packet.
//
// Layout: [0x1C][int64 time][int64 serverGUID][16-byte magic][uint16 strLen][MOTD string]
func parsePong(data []byte) (*ServerInfo, error) {
	// 1 (id) + 8 (time) + 8 (guid) + 16 (magic) + 2 (len) = 35 byte header.
	const headerLen = 1 + 8 + 8 + 16 + 2
	if len(data) < headerLen {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}
	if data[0] != idUnconnectedPong {
		return nil, fmt.Errorf("unexpected packet id: 0x%02X", data[0])
	}

	strLen := int(binary.BigEndian.Uint16(data[33:35]))
	motd := data[headerLen:]
	// Trust the declared length when present, but never read past the buffer.
	if strLen > 0 && strLen <= len(motd) {
		motd = motd[:strLen]
	}

	return parseMOTD(string(motd)), nil
}

// parseMOTD splits the ';'-separated MOTD string into a ServerInfo.
func parseMOTD(motd string) *ServerInfo {
	fields := strings.Split(motd, ";")
	info := &ServerInfo{Fields: fields}

	get := func(i int) string {
		if i < len(fields) {
			return fields[i]
		}
		return ""
	}

	info.Edition = get(0)
	info.MOTDLine1 = get(1)
	info.ProtocolVersion = get(2)
	info.VersionName = get(3)
	info.PlayerCount = atoiSafe(get(4))
	info.MaxPlayers = atoiSafe(get(5))
	info.ServerID = get(6)
	info.MOTDLine2 = get(7)
	info.GameMode = get(8)
	info.GameModeNumeric = get(9)
	info.PortIPv4 = get(10)
	info.PortIPv6 = get(11)

	return info
}

// atoiSafe parses an int, returning 0 on any error.
func atoiSafe(s string) int {
	if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return v
	}
	return 0
}

// Name returns the protocol name.
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (Bedrock/RakNet)", p.GameName)
}

// DefaultPort returns the default Bedrock UDP port.
func (p *Protocol) DefaultPort() int { return defaultPort }

// SupportsSRV indicates Bedrock does not use SRV records.
func (p *Protocol) SupportsSRV() bool { return false }

// SRVService returns an empty string (not used).
func (p *Protocol) SRVService() string { return "" }
