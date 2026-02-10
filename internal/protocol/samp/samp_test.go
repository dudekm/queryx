package samp

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
)

func TestProtocol_Name(t *testing.T) {
	p := NewProtocol(nil, "SA-MP")
	assert.Equal(t, "SA-MP (SA-MP)", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	p := NewProtocol(nil, "SA-MP")
	assert.Equal(t, 7777, p.DefaultPort())
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil, "SA-MP")
	assert.False(t, p.SupportsSRV())
}

func TestBuildPacket(t *testing.T) {
	ip := net.ParseIP("127.0.0.1").To4()
	port := 7777

	packet := buildPacket(ip, port, opcodeInfo)

	assert.Equal(t, 11, len(packet))
	assert.Equal(t, "SAMP", string(packet[0:4]))
	assert.Equal(t, ip, net.IP(packet[4:8]))
	assert.Equal(t, byte(port&0xFF), packet[8])
	assert.Equal(t, byte((port>>8)&0xFF), packet[9])
	assert.Equal(t, byte(opcodeInfo), packet[10])
}

func TestParseInfoResponse(t *testing.T) {
	// Build mock SA-MP info response
	response := buildMockInfoResponse(
		"Test SA-MP Server",
		"Roleplay",
		"English",
		50,  // players
		100, // max players
		false,
	)

	result, err := parseInfoResponse(response)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test SA-MP Server", result.Name)
	assert.Equal(t, "Roleplay", result.Map)
	assert.Equal(t, 50, result.NumPlayers)
	assert.Equal(t, 100, result.MaxPlayers)
	assert.False(t, result.Password)

	// Check Raw contains protocol-specific data
	rawMap, ok := result.Raw.(map[string]interface{})
	assert.True(t, ok, "Raw should be a map")
	assert.Equal(t, "English", rawMap["language"])
	assert.Equal(t, "Roleplay", rawMap["gamemode"])
}

func TestParseInfoResponse_WithPassword(t *testing.T) {
	response := buildMockInfoResponse(
		"Private Server",
		"Freeroam",
		"Czech",
		10,
		50,
		true, // password protected
	)

	result, err := parseInfoResponse(response)

	assert.NoError(t, err)
	assert.True(t, result.Password)
}

func TestParseInfoResponse_InvalidMagic(t *testing.T) {
	invalidResponse := []byte("INVALID" + string(make([]byte, 4)))

	result, err := parseInfoResponse(invalidResponse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid magic")
}

func TestParseInfoResponse_TooShort(t *testing.T) {
	shortResponse := []byte("SAMP")

	result, err := parseInfoResponse(shortResponse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "response too short")
}

func TestReadString32(t *testing.T) {
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
			data:        append([]byte{0x05, 0x00, 0x00, 0x00}, []byte("Hello")...),
			offset:      0,
			expectedStr: "Hello",
			expectedOff: 9,
			expectError: false,
		},
		{
			name:        "empty string",
			data:        []byte{0x00, 0x00, 0x00, 0x00},
			offset:      0,
			expectedStr: "",
			expectedOff: 4,
			expectError: false,
		},
		{
			name:        "UTF-8 string",
			data:        append([]byte{0x0C, 0x00, 0x00, 0x00}, []byte("Привет")...), // 12 bytes in UTF-8
			offset:      0,
			expectedStr: "Привет",
			expectedOff: 16, // 4 (length) + 12 (data)
			expectError: false,
		},
		{
			name:        "insufficient data for length",
			data:        []byte{0x05, 0x00},
			offset:      0,
			expectError: true,
		},
		{
			name:        "insufficient data for content",
			data:        []byte{0x10, 0x00, 0x00, 0x00, 0x41, 0x42},
			offset:      0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, offset, err := readString32(tt.data, tt.offset)

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

func TestProtocol_Query_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPError = assert.AnError

	p := NewProtocol(mockTransport, "SA-MP")
	ctx := context.Background()

	result, err := p.Query(ctx, "127.0.0.1:7777")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send SA-MP query")
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "SA-MP")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Query(ctx, "127.0.0.1:7777")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// Helper function to build mock SA-MP info response
func buildMockInfoResponse(hostname, gamemode, language string, players, maxPlayers int, password bool) []byte {
	data := make([]byte, 0, 256)

	// Magic "SAMP"
	data = append(data, []byte("SAMP")...)

	// IP (127.0.0.1)
	data = append(data, 127, 0, 0, 1)

	// Port (7777 in little-endian)
	data = append(data, byte(7777&0xFF), byte((7777>>8)&0xFF))

	// Opcode (info)
	data = append(data, opcodeInfo)

	// Password flag
	if password {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}

	// Player count (2 bytes, little-endian)
	data = append(data, byte(players&0xFF), byte((players>>8)&0xFF))

	// Max players (2 bytes, little-endian)
	data = append(data, byte(maxPlayers&0xFF), byte((maxPlayers>>8)&0xFF))

	// Hostname (4-byte length + string)
	data = appendString32(data, hostname)

	// Gamemode (4-byte length + string)
	data = appendString32(data, gamemode)

	// Language (4-byte length + string)
	data = appendString32(data, language)

	return data
}

func appendString32(data []byte, s string) []byte {
	// Length (4 bytes, little-endian)
	length := uint32(len(s))
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, length)
	data = append(data, lengthBytes...)

	// String content
	data = append(data, []byte(s)...)

	return data
}
