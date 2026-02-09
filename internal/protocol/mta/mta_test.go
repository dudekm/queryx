package mta

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
)

func TestProtocol_Name(t *testing.T) {
	p := NewProtocol(nil, "Multi Theft Auto")
	assert.Equal(t, "Multi Theft Auto (ASE)", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	p := NewProtocol(nil, "Multi Theft Auto")
	assert.Equal(t, 22003, p.DefaultPort())
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil, "Multi Theft Auto")
	assert.False(t, p.SupportsSRV())
}

func TestParseASEResponse_EYE1(t *testing.T) {
	// Build mock MTA ASE (EYE1) response
	response := buildMockASEResponse(
		"Test MTA Server",
		"Race",
		"race-dusty",
		"1.6.0-9.21261.0",
		25, // players
		32, // max players
		false,
	)

	result, err := parseASEResponse(response)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test MTA Server", result.Name)
	assert.Equal(t, "race-dusty", result.Map)
	assert.Equal(t, 25, result.NumPlayers)
	assert.Equal(t, 32, result.MaxPlayers)
	assert.Equal(t, "1.6.0-9.21261.0", result.Version)
	assert.False(t, result.Password)
	assert.Equal(t, "Race", result.Extra["gamemode"])
}

func TestParseASEResponse_WithPassword(t *testing.T) {
	response := buildMockASEResponse(
		"Private Server",
		"Freeroam",
		"map-name",
		"1.6.0",
		5,
		16,
		true, // password protected
	)

	result, err := parseASEResponse(response)

	assert.NoError(t, err)
	assert.True(t, result.Password)
}

func TestParseASEResponse_InvalidMagic(t *testing.T) {
	invalidResponse := []byte("INVALID")

	result, err := parseASEResponse(invalidResponse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid magic bytes")
}

func TestParseASEResponse_TooShort(t *testing.T) {
	shortResponse := []byte("EYE")

	result, err := parseASEResponse(shortResponse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "response too short")
}

func TestReadLengthPrefixedString(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedStr string
		expectError bool
	}{
		{
			name:        "simple string with null terminator",
			data:        []byte{0x06, 'H', 'e', 'l', 'l', 'o', 0x00},
			expectedStr: "Hello",
			expectError: false,
		},
		{
			name:        "empty string",
			data:        []byte{0x00},
			expectedStr: "",
			expectError: false,
		},
		{
			name:        "string without null terminator",
			data:        []byte{0x05, 'H', 'e', 'l', 'l', 'o'},
			expectedStr: "Hello",
			expectError: false,
		},
		{
			name:        "no data",
			data:        []byte{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			str, err := readLengthPrefixedString(reader)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStr, str)
			}
		})
	}
}

func TestProtocol_Query_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPError = assert.AnError

	p := NewProtocol(mockTransport, "Multi Theft Auto")
	ctx := context.Background()

	result, err := p.Query(ctx, "127.0.0.1:22003")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send MTA ASE query")
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "Multi Theft Auto")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Query(ctx, "127.0.0.1:22003")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestASEPortCalculation(t *testing.T) {
	// Test that ASE port is calculated correctly (game port + 123)
	tests := []struct {
		gamePort int
		asePort  int
	}{
		{22003, 22126},
		{22004, 22127},
		{30000, 30123},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			calculated := tt.gamePort + asePortOffset
			assert.Equal(t, tt.asePort, calculated)
		})
	}
}

// Helper function to build mock MTA ASE response
func buildMockASEResponse(serverName, gamemode, mapName, version string, players, maxPlayers int, password bool) []byte {
	buf := &bytes.Buffer{}

	// Magic "EYE1"
	buf.WriteString(magicEYE1)

	// Game type length + "mta" + null terminator
	buf.WriteByte(4)
	buf.WriteString("mta\x00")

	// Port string (e.g., "22003" + null)
	portStr := "22003\x00"
	buf.WriteByte(byte(len(portStr)))
	buf.WriteString(portStr)

	// Server name
	writeLengthPrefixedString(buf, serverName)

	// Game type/mode
	writeLengthPrefixedString(buf, gamemode)

	// Map name
	writeLengthPrefixedString(buf, mapName)

	// Version
	writeLengthPrefixedString(buf, version)

	// Password flag
	if password {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	// Player count (2 bytes, little-endian)
	buf.WriteByte(byte(players & 0xFF))
	buf.WriteByte(byte((players >> 8) & 0xFF))

	// Max players (2 bytes, little-endian)
	buf.WriteByte(byte(maxPlayers & 0xFF))
	buf.WriteByte(byte((maxPlayers >> 8) & 0xFF))

	return buf.Bytes()
}

func writeLengthPrefixedString(buf *bytes.Buffer, s string) {
	// Length includes null terminator
	buf.WriteByte(byte(len(s) + 1))
	buf.WriteString(s)
	buf.WriteByte(0) // null terminator
}
