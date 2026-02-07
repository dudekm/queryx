package source

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
	p := NewProtocol(nil, "CS2")
	assert.Equal(t, "CS2 (Source Engine)", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	p := NewProtocol(nil, "CS2")
	assert.Equal(t, 27015, p.DefaultPort())
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil, "CS2")
	assert.False(t, p.SupportsSRV())
}

func TestProtocol_BuildA2SInfoRequest(t *testing.T) {
	p := NewProtocol(nil, "CS2")
	request := p.buildA2SInfoRequestWithChallenge(-1)

	assert.NotNil(t, request)
	assert.Greater(t, len(request), 0)

	// Verify magic header
	reader := bytes.NewReader(request)
	var header int32
	binary.Read(reader, binary.LittleEndian, &header)
	assert.Equal(t, int32(headerMagic), header)

	// Verify packet type
	packetType, _ := reader.ReadByte()
	assert.Equal(t, byte(a2sInfoRequest), packetType)
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock response
	response := buildMockA2SInfoResponse(
		"Test CS2 Server",
		"de_dust2",
		"Counter-Strike 2",
		10, // players
		20, // max players
		2,  // bots
	)

	mockTransport.UDPResponses["127.0.0.1:27015"] = response

	p := NewProtocol(mockTransport, "CS2")
	ctx := context.Background()

	result, err := p.Query(ctx, "127.0.0.1:27015")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test CS2 Server", result.Name)
	assert.Equal(t, "de_dust2", result.Map)
	assert.Equal(t, 10, result.NumPlayers)
	assert.Equal(t, 2, result.Bots)
	assert.Equal(t, 20, result.MaxPlayers)
	assert.NotNil(t, result.Raw)
}

func TestProtocol_Query_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPError = assert.AnError

	p := NewProtocol(mockTransport, "CS2")
	ctx := context.Background()

	result, err := p.Query(ctx, "127.0.0.1:27015")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "CS2")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Query(ctx, "127.0.0.1:27015")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProtocol_Query_WithChallenge(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build challenge response
	challengeBuf := &bytes.Buffer{}
	binary.Write(challengeBuf, binary.LittleEndian, int32(headerMagic))
	challengeBuf.WriteByte(a2sChallenge)
	binary.Write(challengeBuf, binary.LittleEndian, int32(12345)) // Challenge number

	// Build final response
	finalResponse := buildMockA2SInfoResponse(
		"Test CS2 Server",
		"de_dust2",
		"Counter-Strike 2",
		10, // players
		20, // max players
		2,  // bots
	)

	// Setup response queue: first call returns challenge, second returns actual data
	addr := "127.0.0.1:27015"
	mockTransport.UDPResponseQueue[addr] = [][]byte{
		challengeBuf.Bytes(), // First call: challenge
		finalResponse,        // Second call: actual response
	}

	p := NewProtocol(mockTransport, "CS2")
	ctx := context.Background()

	result, err := p.Query(ctx, addr)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test CS2 Server", result.Name)
	assert.Equal(t, "de_dust2", result.Map)
	assert.Equal(t, 10, result.NumPlayers)
}

func TestProtocol_ParseResponse_InvalidHeader(t *testing.T) {
	p := NewProtocol(nil, "CS2")

	// Invalid magic header
	data := []byte{0x00, 0x00, 0x00, 0x00}

	result, err := p.parseA2SInfoResponse(data)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid header magic")
}

func TestProtocol_ParseResponse_InvalidPacketType(t *testing.T) {
	p := NewProtocol(nil, "CS2")

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, int32(headerMagic))
	buf.WriteByte(0xFF) // Invalid packet type

	result, err := p.parseA2SInfoResponse(buf.Bytes())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unexpected packet type")
}

func TestReadNullTerminatedString(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "simple string",
			data:     []byte("hello\x00"),
			expected: "hello",
		},
		{
			name:     "empty string",
			data:     []byte("\x00"),
			expected: "",
		},
		{
			name:     "string with spaces",
			data:     []byte("hello world\x00"),
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			result, err := readNullTerminatedString(reader)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// buildMockA2SInfoResponse builds a mock A2S_INFO response for testing
func buildMockA2SInfoResponse(serverName, mapName, game string, players, maxPlayers, bots byte) []byte {
	buf := &bytes.Buffer{}

	// Magic header
	binary.Write(buf, binary.LittleEndian, int32(headerMagic))

	// Packet type
	buf.WriteByte(a2sInfoResponse)

	// Protocol version
	buf.WriteByte(0x11)

	// Server name
	buf.WriteString(serverName)
	buf.WriteByte(0x00)

	// Map
	buf.WriteString(mapName)
	buf.WriteByte(0x00)

	// Folder
	buf.WriteString("csgo")
	buf.WriteByte(0x00)

	// Game
	buf.WriteString(game)
	buf.WriteByte(0x00)

	// Game ID
	binary.Write(buf, binary.LittleEndian, uint16(730))

	// Players
	buf.WriteByte(players)
	buf.WriteByte(maxPlayers)
	buf.WriteByte(bots)

	// Server type
	buf.WriteByte('d') // Dedicated

	// Environment
	buf.WriteByte('l') // Linux

	// Visibility
	buf.WriteByte(0x00) // Public

	// VAC
	buf.WriteByte(0x01) // Secured

	// Version
	buf.WriteString("1.0.0.0")
	buf.WriteByte(0x00)

	return buf.Bytes()
}
