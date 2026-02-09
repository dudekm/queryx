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
	assert.Nil(t, result.Raw) // Source Engine doesn't return JSON
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

// TestSourceProtocol_AllGames tests all games using Source Engine A2S protocol
func TestSourceProtocol_AllGames(t *testing.T) {
	games := []struct {
		name         string
		gameName     string
		expectedPort int
	}{
		// Valve Source Engine Games
		{"Team Fortress 2", "tf2", 27015},
		{"Left 4 Dead", "l4d", 27015},
		{"Left 4 Dead 2", "l4d2", 27015},
		{"Garry's Mod", "gmod", 27015},
		{"Black Mesa", "blackmesa", 27015},
		{"Day of Infamy", "dayofinfamy", 27015},
		{"Insurgency", "insurgency", 27015},
		{"Insurgency: Sandstorm", "insurgencysandstorm", 27015},
		{"Killing Floor 2", "killingfloor2", 27015},

		// Counter-Strike Games
		{"Counter-Strike 1.6", "cs16", 27015},
		{"Counter-Strike: Source", "cssource", 27015},
		{"Counter-Strike 2", "cs2", 27015},

		// Survival Games Using A2S
		{"ARK: Survival Evolved", "ark", 27015},
		{"ARK: Survival Ascended", "arkascended", 27015},
		{"ATLAS", "atlas", 27015},
		{"Conan Exiles", "conanexiles", 27015},
		{"7 Days to Die", "7daystodie", 26900},
		{"Rust", "rust", 28015},

		// Co-op/Tactical Games
		{"Barotrauma", "barotrauma", 27015},
		{"Hell Let Loose", "hellletloose", 27015},
		{"Post Scriptum", "postscriptum", 27015},
		{"Squad", "squad", 27015},
		{"Rising Storm 2: Vietnam", "risingstorm2", 27015},

		// Space/Sandbox Games
		{"Avorion", "avorion", 27015},
		{"Empyrion - Galactic Survival", "empyrion", 30000},
		{"Stationeers", "stationeers", 27015},
		{"Space Engineers", "spaceengineers", 27015},

		// Other Survival/Sandbox
		{"Hurtworld", "hurtworld", 12871},
		{"ICARUS", "icarus", 17777},
		{"Enshrouded", "enshrouded", 15636},
		{"V Rising", "vrising", 27015},
		{"Unturned", "unturned", 27015},
		{"The Forest", "theforest", 27015},
		{"No One Survived", "noonesurvived", 27015},
		{"Miscreated", "miscreated", 27015},
		{"DeadPoly", "deadpoly", 27015},
		{"Dysterra", "dysterra", 27015},
		{"Subsistence", "subsistence", 27015},
		{"PixARK", "pixark", 27015},
		{"Valheim", "valheim", 2456},
	}

	for _, game := range games {
		t.Run(game.name, func(t *testing.T) {
			mockTransport := transport.NewMockTransport()

			// Build mock response
			response := buildMockA2SInfoResponse(
				"Test "+game.name+" Server",
				"test_map",
				game.name,
				5,  // players
				10, // max players
				0,  // bots
			)

			mockTransport.UDPResponses["test:27015"] = response

			proto := NewProtocol(mockTransport, game.gameName)
			result, err := proto.Query(context.Background(), "test:27015")

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Online)
			assert.Equal(t, "Test "+game.name+" Server", result.Name)
			assert.Equal(t, "test_map", result.Map)
			assert.Equal(t, 5, result.NumPlayers)
			assert.Equal(t, 10, result.MaxPlayers)
			assert.Equal(t, game.expectedPort, proto.DefaultPort())
		})
	}
}

// TestGetDefaultPort tests the port lookup function
func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		gameName     string
		expectedPort int
	}{
		{"tf2", 27015},
		{"Team Fortress 2", 27015},
		{"rust", 28015},
		{"Rust", 28015},
		{"7daystodie", 26900},
		{"empyrion", 30000},
		{"hurtworld", 12871},
		{"icarus", 17777},
		{"valheim", 2456},
		{"unknown_game", 27015}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.gameName, func(t *testing.T) {
			port := GetDefaultPort(tt.gameName)
			assert.Equal(t, tt.expectedPort, port)
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
