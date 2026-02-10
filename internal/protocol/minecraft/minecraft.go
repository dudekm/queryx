package minecraft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	defaultPort = 25565
	srvService  = "minecraft"
)

// Protocol implements the Minecraft Java Edition Server List Ping protocol
type Protocol struct {
	protocol.BaseProtocol
}

// NewProtocol creates a new Minecraft protocol handler
func NewProtocol(t transport.Transport) *Protocol {
	return &Protocol{
		BaseProtocol: protocol.NewBaseProtocol(t, "Minecraft Java Edition"),
	}
}

// Query queries a Minecraft server and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	return p.QueryWithHostname(ctx, addr, "")
}

// QueryWithHostname queries a Minecraft server with original hostname for SNI
func (p *Protocol) QueryWithHostname(ctx context.Context, addr string, hostname string) (*protocol.QueryResult, error) {
	// Build handshake packet with original hostname if provided
	var handshake []byte
	var err error
	if hostname != "" {
		handshake, err = p.buildHandshakePacketWithHostname(addr, hostname)
	} else {
		handshake, err = p.buildHandshakePacket(addr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to build handshake: %w", err)
	}

	// Build status request packet
	statusRequest := p.buildStatusRequestPacket()

	// Combine packets
	request := append(handshake, statusRequest...)

	// Send via TCP and measure network latency (ping)
	pingStart := time.Now()
	response, err := p.Transport.SendTCP(ctx, addr, request)
	pingDuration := time.Since(pingStart)
	pingMs := float64(pingDuration.Microseconds()) / 1000.0

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Parse response
	result, err := p.parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result.Ping = pingMs // Set network latency in milliseconds
	return result, nil
}

// buildHandshakePacketWithHostname builds the handshake packet with custom hostname
func (p *Protocol) buildHandshakePacketWithHostname(addr string, hostname string) ([]byte, error) {
	buf := &bytes.Buffer{}

	// Packet ID (0x00 for handshake)
	if err := writeVarInt(buf, 0x00); err != nil {
		return nil, err
	}

	// Protocol version (use a widely supported version like 47 for 1.8.x, or recent 763 for 1.20.x)
	// For status queries, any protocol version should work
	if err := writeVarInt(buf, 47); err != nil {
		return nil, err
	}

	// Server address (use provided hostname, not resolved IP)
	if err := writeString(buf, hostname); err != nil {
		return nil, err
	}

	// Server port
	_, port, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}
	buf.WriteByte(byte(port >> 8))
	buf.WriteByte(byte(port & 0xFF))

	// Next state (1 for status)
	if err := writeVarInt(buf, 1); err != nil {
		return nil, err
	}

	// Prepend packet length
	packetData := buf.Bytes()
	finalBuf := &bytes.Buffer{}
	if err := writeVarInt(finalBuf, len(packetData)); err != nil {
		return nil, err
	}
	finalBuf.Write(packetData)

	return finalBuf.Bytes(), nil
}

// buildHandshakePacket builds the handshake packet
func (p *Protocol) buildHandshakePacket(addr string) ([]byte, error) {
	buf := &bytes.Buffer{}

	// Packet ID (0x00 for handshake)
	if err := writeVarInt(buf, 0x00); err != nil {
		return nil, err
	}

	// Protocol version (use a widely supported version like 47 for 1.8.x, or recent 763 for 1.20.x)
	// For status queries, any protocol version should work
	if err := writeVarInt(buf, 47); err != nil {
		return nil, err
	}

	// Server address (without port)
	host, _, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}
	if err := writeString(buf, host); err != nil {
		return nil, err
	}

	// Server port
	_, port, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}
	buf.WriteByte(byte(port >> 8))
	buf.WriteByte(byte(port & 0xFF))

	// Next state (1 for status)
	if err := writeVarInt(buf, 1); err != nil {
		return nil, err
	}

	// Prepend packet length
	packetData := buf.Bytes()
	finalBuf := &bytes.Buffer{}
	if err := writeVarInt(finalBuf, len(packetData)); err != nil {
		return nil, err
	}
	finalBuf.Write(packetData)

	return finalBuf.Bytes(), nil
}

// buildStatusRequestPacket builds the status request packet
func (p *Protocol) buildStatusRequestPacket() []byte {
	buf := &bytes.Buffer{}
	writeVarInt(buf, 1) // Packet length
	writeVarInt(buf, 0) // Packet ID (0x00 for status request)
	return buf.Bytes()
}

// parseResponse parses the server response
func (p *Protocol) parseResponse(data []byte) (*protocol.QueryResult, error) {
	reader := bytes.NewReader(data)

	// Read packet length
	_, err := readVarInt(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %w", err)
	}

	// Read packet ID
	packetID, err := readVarInt(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}

	if packetID != 0x00 {
		return nil, fmt.Errorf("unexpected packet ID: %d", packetID)
	}

	// Read JSON length
	jsonLength, err := readVarInt(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON length: %w", err)
	}

	// Read JSON data
	jsonData := make([]byte, jsonLength)
	if _, err := io.ReadFull(reader, jsonData); err != nil {
		return nil, fmt.Errorf("failed to read JSON: %w", err)
	}

	// Parse JSON
	var response serverResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to QueryResult
	result := p.convertToQueryResult(&response)

	// Parse as generic map for Raw field - ALL protocol-specific data goes here
	var rawParsed map[string]interface{}
	json.Unmarshal(jsonData, &rawParsed)
	result.Raw = rawParsed // Full Minecraft JSON response (includes favicon, protocol version, etc.)

	return result, nil
}

// convertToQueryResult converts server response to QueryResult
func (p *Protocol) convertToQueryResult(resp *serverResponse) *protocol.QueryResult {
	result := &protocol.QueryResult{
		Online:     true,
		Name:       cleanMOTD(resp.Description),
		NumPlayers: resp.Players.Online,
		MaxPlayers: resp.Players.Max,
		Players:    []protocol.Player{}, // Initialize as empty array, not nil
		Bots:       0,                   // Minecraft servers don't have bots
		Version:    resp.Version.Name,
		Password:   false, // Minecraft Java doesn't expose password info via query
	}

	// Add player list if available
	if len(resp.Players.Sample) > 0 {
		result.Players = make([]protocol.Player, len(resp.Players.Sample))
		for i, p := range resp.Players.Sample {
			result.Players[i] = protocol.Player{
				Name: p.Name,
			}
		}
	}

	// Note: All protocol-specific data (protocol version, favicon, full JSON)
	// is added to Raw field in parseResponse(), not here

	return result
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return "Minecraft Java Edition"
}

// DefaultPort returns the default Minecraft port
func (p *Protocol) DefaultPort() int {
	return defaultPort
}

// SupportsSRV indicates that Minecraft supports SRV records
func (p *Protocol) SupportsSRV() bool {
	return true
}

// SRVService returns the SRV service name
func (p *Protocol) SRVService() string {
	return srvService
}
