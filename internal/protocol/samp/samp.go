package samp

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// Default SA-MP server port
	defaultPort = 7777

	// Magic identifier
	magic = "SAMP"

	// Opcodes
	opcodeInfo     = 'i' // 0x69 - Server information
	opcodeRules    = 'r' // 0x72 - Server rules
	opcodePlayers  = 'c' // 0x63 - Player list
	opcodeDetailed = 'd' // 0x64 - Detailed player info
	opcodePing     = 'p' // 0x70 - Ping
)

// Protocol implements SA-MP Query Protocol
type Protocol struct {
	protocol.BaseProtocol
}

// NewProtocol creates a new SA-MP protocol handler
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{
		BaseProtocol: protocol.NewBaseProtocol(t, gameName),
	}
}

// Query queries a SA-MP server and returns the result
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

	// Resolve IP address
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %w", err)
	}

	var ipv4 net.IP
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			ipv4 = ip4
			break
		}
	}

	if ipv4 == nil {
		return nil, fmt.Errorf("no IPv4 address found for host")
	}

	// Build info request
	request := buildPacket(ipv4, port, opcodeInfo)

	// Send request and measure ping
	pingStart := time.Now()
	response, err := p.Transport.SendUDP(ctx, addr, request)
	ping := time.Since(pingStart)

	if err != nil {
		return nil, fmt.Errorf("failed to send SA-MP query: %w", err)
	}

	// Parse response
	result, err := parseInfoResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SA-MP response: %w", err)
	}

	result.Ping = ping
	return result, nil
}

// buildPacket builds a SA-MP query packet with 11-byte header
func buildPacket(ip net.IP, port int, opcode byte) []byte {
	packet := make([]byte, 11)

	// Magic "SAMP"
	copy(packet[0:4], magic)

	// IP address (4 bytes)
	copy(packet[4:8], ip.To4())

	// Port (2 bytes, little-endian)
	packet[8] = byte(port & 0xFF)
	packet[9] = byte((port >> 8) & 0xFF)

	// Opcode
	packet[10] = opcode

	return packet
}

// SAMPInfo contains all data from SA-MP server info response
type SAMPInfo struct {
	Password   bool   `json:"password"`
	NumPlayers int    `json:"numPlayers"`
	MaxPlayers int    `json:"maxPlayers"`
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Language   string `json:"language"`
}

// parseInfoResponse parses SA-MP info response (opcode 'i')
func parseInfoResponse(data []byte) (*protocol.QueryResult, error) {
	if len(data) < 11 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Verify header
	if string(data[0:4]) != magic {
		return nil, fmt.Errorf("invalid magic: %q", data[0:4])
	}

	if data[10] != opcodeInfo {
		return nil, fmt.Errorf("unexpected opcode: %d", data[10])
	}

	offset := 11

	// Parse all fields into struct
	info := &SAMPInfo{}

	// Password protected (1 byte)
	if offset >= len(data) {
		return nil, fmt.Errorf("insufficient data for password flag")
	}
	info.Password = data[offset] != 0
	offset++

	// Player count (2 bytes, little-endian)
	if offset+2 > len(data) {
		return nil, fmt.Errorf("insufficient data for player count")
	}
	info.NumPlayers = int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Max players (2 bytes, little-endian)
	if offset+2 > len(data) {
		return nil, fmt.Errorf("insufficient data for max players")
	}
	info.MaxPlayers = int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Hostname (4-byte length + string)
	hostname, newOffset, err := readString32(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read hostname: %w", err)
	}
	info.Hostname = hostname
	offset = newOffset

	// Gamemode (4-byte length + string)
	gamemode, newOffset, err := readString32(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read gamemode: %w", err)
	}
	info.Gamemode = gamemode
	offset = newOffset

	// Language (4-byte length + string)
	language, _, err := readString32(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read language: %w", err)
	}
	info.Language = language

	// Build result - put entire struct in Raw
	result := &protocol.QueryResult{
		Online:     true,
		Name:       info.Hostname,
		Map:        info.Gamemode,
		NumPlayers: info.NumPlayers,
		MaxPlayers: info.MaxPlayers,
		Players:    []protocol.Player{}, // Initialize as empty array, not nil
		Password:   info.Password,
		Raw:        info, // ALL data in single struct
	}

	return result, nil
}

// readString32 reads a string with 4-byte length prefix (little-endian)
func readString32(data []byte, offset int) (string, int, error) {
	if offset+4 > len(data) {
		return "", offset, fmt.Errorf("insufficient data for string length at offset %d", offset)
	}

	length := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	if offset+length > len(data) {
		return "", offset, fmt.Errorf("insufficient data for string content at offset %d (need %d bytes)", offset, length)
	}

	str := string(data[offset : offset+length])
	offset += length

	return str, offset, nil
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (SA-MP)", p.GameName)
}

// DefaultPort returns the default SA-MP port
func (p *Protocol) DefaultPort() int {
	return defaultPort
}

// SupportsSRV indicates that SA-MP does not use SRV records
func (p *Protocol) SupportsSRV() bool {
	return false
}

// SRVService returns empty string (not used)
func (p *Protocol) SRVService() string {
	return ""
}
