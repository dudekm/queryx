package hytale

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
)

func TestProtocol_Name(t *testing.T) {
	p := NewProtocol(nil, "Hytale")
	assert.Equal(t, "Hytale (HyQuery)", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	p := NewProtocol(nil, "Hytale")
	assert.Equal(t, 5520, p.DefaultPort())
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil, "Hytale")
	assert.False(t, p.SupportsSRV())
}

func TestBuildQueryPacket(t *testing.T) {
	tests := []struct {
		name      string
		queryType byte
		expected  []byte
	}{
		{
			name:      "basic query",
			queryType: queryTypeBasic,
			expected:  []byte("HYQUERY\x00\x00"),
		},
		{
			name:      "full query",
			queryType: queryTypeFull,
			expected:  []byte("HYQUERY\x00\x01"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := buildQueryPacket(tt.queryType)
			assert.Equal(t, tt.expected, packet)
			assert.Equal(t, 9, len(packet))
		})
	}
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock response
	response := buildMockBasicResponse(
		"Test Hytale Server",
		"Welcome to our server!",
		10, // online players
		32, // max players
		5520,
		"2026.01.22",
	)

	mockTransport.UDPResponses["test:5520"] = response

	p := NewProtocol(mockTransport, "Hytale")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:5520")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test Hytale Server", result.Name)
	assert.Equal(t, 10, result.NumPlayers)
	assert.Equal(t, 32, result.MaxPlayers)
	assert.Equal(t, "2026.01.22", result.Version)
	assert.Equal(t, "Welcome to our server!", result.Extra["motd"])
	assert.Equal(t, uint32(5520), result.Extra["port"])
}

func TestProtocol_Query_FullResponse(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock full response with players
	response := buildMockFullResponse(
		"Test Server",
		"MOTD",
		3, // online players
		20,
		5520,
		"1.0.0",
		[]string{"Player1", "Player2", "Player3"},
		[]string{"HyQuery", "TestPlugin"},
	)

	mockTransport.UDPResponses["test:5520"] = response

	p := NewProtocol(mockTransport, "Hytale")
	result, err := p.Query(context.Background(), "test:5520")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.NumPlayers)
	assert.Equal(t, 3, len(result.Players))
	assert.Equal(t, "Player1", result.Players[0].Name)
	assert.Equal(t, "Player2", result.Players[1].Name)
	assert.Equal(t, "Player3", result.Players[2].Name)

	plugins, ok := result.Extra["plugins"].([]string)
	assert.True(t, ok)
	assert.Equal(t, 2, len(plugins))
	assert.Equal(t, "HyQuery", plugins[0])
	assert.Equal(t, "TestPlugin", plugins[1])
}

func TestProtocol_Query_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPError = assert.AnError

	p := NewProtocol(mockTransport, "Hytale")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:5520")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send HyQuery request")
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "Hytale")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Query(ctx, "test:5520")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseResponse_InvalidMagic(t *testing.T) {
	invalidResponse := []byte("INVALID\x00\x00")

	result, err := parseResponse(invalidResponse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid response magic")
}

func TestParseResponse_TooShort(t *testing.T) {
	shortResponse := []byte("HYREPLY")

	result, err := parseResponse(shortResponse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "response too short")
}

func TestReadString(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		offset      int
		expectedStr string
		expectedOff int
		expectError bool
	}{
		{
			name:        "simple string",
			data:        append([]byte{0x05, 0x00}, []byte("Hello")...),
			offset:      0,
			expectedStr: "Hello",
			expectedOff: 7,
			expectError: false,
		},
		{
			name:        "empty string",
			data:        []byte{0x00, 0x00},
			offset:      0,
			expectedStr: "",
			expectedOff: 2,
			expectError: false,
		},
		{
			name:        "UTF-8 string",
			data:        append([]byte{0x0C, 0x00}, []byte("Привет")...), // 12 bytes in UTF-8
			offset:      0,
			expectedStr: "Привет",
			expectedOff: 14, // 2 (length) + 12 (data)
			expectError: false,
		},
		{
			name:        "insufficient data for length",
			data:        []byte{0x05},
			offset:      0,
			expectError: true,
		},
		{
			name:        "insufficient data for content",
			data:        []byte{0x10, 0x00, 0x41, 0x42}, // says 16 bytes, has 2
			offset:      0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, offset, err := readString(tt.data, tt.offset)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStr, str)
				assert.Equal(t, tt.expectedOff, offset)
			}
		})
	}
}

func TestParsePlayers(t *testing.T) {
	// Build player data: count (4 bytes) + player entries
	buf := &bytes.Buffer{}

	// Player count: 2
	binary.Write(buf, binary.LittleEndian, uint32(2))

	// Player 1: "Alice" + UUID (16 bytes)
	writeString(buf, "Alice")
	buf.Write(make([]byte, 16)) // UUID

	// Player 2: "Bob" + UUID (16 bytes)
	writeString(buf, "Bob")
	buf.Write(make([]byte, 16)) // UUID

	players, _, err := parsePlayers(buf.Bytes(), 0)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(players))
	assert.Equal(t, "Alice", players[0].Name)
	assert.Equal(t, "Bob", players[1].Name)
}

func TestParsePlugins(t *testing.T) {
	// Build plugin data: count (4 bytes) + plugin names
	buf := &bytes.Buffer{}

	// Plugin count: 3
	binary.Write(buf, binary.LittleEndian, uint32(3))

	// Plugins
	writeString(buf, "HyQuery")
	writeString(buf, "WorldEdit")
	writeString(buf, "Essentials")

	plugins, _, err := parsePlugins(buf.Bytes(), 0)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(plugins))
	assert.Equal(t, "HyQuery", plugins[0])
	assert.Equal(t, "WorldEdit", plugins[1])
	assert.Equal(t, "Essentials", plugins[2])
}

// Helper functions for building mock responses

func buildMockBasicResponse(serverName, motd string, online, max, port uint32, version string) []byte {
	buf := &bytes.Buffer{}

	// Magic
	buf.WriteString(responseMagic)

	// Response type (basic)
	buf.WriteByte(queryTypeBasic)

	// Server name
	writeString(buf, serverName)

	// MOTD
	writeString(buf, motd)

	// Online players
	binary.Write(buf, binary.LittleEndian, online)

	// Max players
	binary.Write(buf, binary.LittleEndian, max)

	// Port
	binary.Write(buf, binary.LittleEndian, port)

	// Version
	writeString(buf, version)

	return buf.Bytes()
}

func buildMockFullResponse(serverName, motd string, online, max, port uint32, version string, players, plugins []string) []byte {
	buf := &bytes.Buffer{}

	// Magic
	buf.WriteString(responseMagic)

	// Response type (full)
	buf.WriteByte(queryTypeFull)

	// Server name
	writeString(buf, serverName)

	// MOTD
	writeString(buf, motd)

	// Online players
	binary.Write(buf, binary.LittleEndian, online)

	// Max players
	binary.Write(buf, binary.LittleEndian, max)

	// Port
	binary.Write(buf, binary.LittleEndian, port)

	// Version
	writeString(buf, version)

	// Players
	binary.Write(buf, binary.LittleEndian, uint32(len(players)))
	for _, player := range players {
		writeString(buf, player)
		buf.Write(make([]byte, 16)) // UUID placeholder
	}

	// Plugins
	binary.Write(buf, binary.LittleEndian, uint32(len(plugins)))
	for _, plugin := range plugins {
		writeString(buf, plugin)
	}

	return buf.Bytes()
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}
