package queryx

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests test the entire flow: Client -> Resolver -> Protocol -> Parser -> QueryResult
// Only the network transport is mocked with real server responses
// This allows refactoring internals while maintaining the public API contract

// Helper: Build Minecraft server list ping response
func buildMinecraftResponse(name string, version string, protocol int, onlinePlayers int, maxPlayers int) []byte {
	// Build JSON response
	response := map[string]interface{}{
		"version": map[string]interface{}{
			"name":     version,
			"protocol": protocol,
		},
		"players": map[string]interface{}{
			"max":    maxPlayers,
			"online": onlinePlayers,
		},
		"description": name,
	}

	jsonBytes, _ := json.Marshal(response)

	// Build Minecraft packet: [length][packet_id=0x00][json_length][json]
	buf := &bytes.Buffer{}

	// Write packet ID (0x00)
	writeMinecraftVarInt(buf, 0)

	// Write JSON length
	writeMinecraftVarInt(buf, len(jsonBytes))

	// Write JSON
	buf.Write(jsonBytes)

	// Prepend packet length
	packetData := buf.Bytes()
	finalBuf := &bytes.Buffer{}
	writeMinecraftVarInt(finalBuf, len(packetData))
	finalBuf.Write(packetData)

	return finalBuf.Bytes()
}

func writeMinecraftVarInt(buf *bytes.Buffer, value int) {
	uvalue := uint32(value)
	for {
		if (uvalue & 0xFFFFFF80) == 0 {
			buf.WriteByte(byte(uvalue))
			return
		}
		buf.WriteByte(byte(uvalue&0x7F | 0x80))
		uvalue >>= 7
	}
}

// Helper: Build Source Engine A2S_INFO response
func buildSourceEngineResponse(name, mapName, game, version string, players, maxPlayers, bots byte) []byte {
	buf := &bytes.Buffer{}

	// Magic header
	binary.Write(buf, binary.LittleEndian, int32(-1))

	// Packet type (A2S_INFO = 0x49)
	buf.WriteByte(0x49)

	// Protocol version
	buf.WriteByte(0x11)

	// Server name (null-terminated)
	buf.WriteString(name)
	buf.WriteByte(0x00)

	// Map (null-terminated)
	buf.WriteString(mapName)
	buf.WriteByte(0x00)

	// Folder (null-terminated)
	buf.WriteString("csgo")
	buf.WriteByte(0x00)

	// Game (null-terminated)
	buf.WriteString(game)
	buf.WriteByte(0x00)

	// Game ID (uint16)
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

	// Version (null-terminated)
	buf.WriteString(version)
	buf.WriteByte(0x00)

	return buf.Bytes()
}

func TestIntegration_Minecraft_FullFlow(t *testing.T) {
	// Build realistic Minecraft response
	minecraftResponse := buildMinecraftResponse(
		"Hypixel Network",
		"Hypixel BETA",
		-1,     // protocol
		35000,  // online players
		200000, // max players
	)

	// Setup mock transport with real response
	mockTransport := transport.NewMockTransport()
	mockTransport.TCPResponses["127.0.0.1:25565"] = minecraftResponse

	// Create client with ONLY transport mocked
	// Resolver, Protocol, Parser are all REAL
	client := NewClientWithDefaults(
		WithTransport(mockTransport),
	)

	// INPUT: Query Minecraft server
	ctx := context.Background()
	result, err := client.Query(ctx, ServerMinecraft, "127.0.0.1", nil)

	// OUTPUT: Verify public API contract
	require.NoError(t, err, "Query should succeed")
	require.NotNil(t, result, "Result should not be nil")

	// Assert output contract
	assert.True(t, result.Online, "Server should be online")
	assert.Equal(t, "Hypixel Network", result.Name, "Server name should match")
	assert.Equal(t, "minecraft", result.Type, "Type should be minecraft")
	assert.Equal(t, 35000, result.NumPlayers, "Should have 35000 players online")
	assert.Equal(t, 200000, result.MaxPlayers, "Max players should be 200000")
	assert.Equal(t, 0, result.Bots, "Minecraft has no bots")
	assert.Equal(t, "Hypixel BETA", result.Version, "Version should match")
	assert.GreaterOrEqual(t, result.Ping, time.Duration(0), "Ping should be non-negative")
	assert.NotZero(t, result.QueriedAt, "QueriedAt should be set")

	// You can refactor internal/* all you want - this test will still pass!
}

func TestIntegration_CounterStrike_FullFlow(t *testing.T) {
	// Build realistic CS2 response
	cs2Response := buildSourceEngineResponse(
		"★ [CS2] BUCURESTI.FAIRSIDE.RO",
		"de_mirage",
		"Counter-Strike 2",
		"1.41.3.5",
		20, // players
		32, // max
		0,  // bots
	)

	// Setup mock transport with real response
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["188.212.102.57:27015"] = cs2Response

	// Create client with ONLY transport mocked
	client := NewClientWithDefaults(
		WithTransport(mockTransport),
	)

	// INPUT: Query CS2 server
	ctx := context.Background()
	result, err := client.Query(ctx, ServerCS2, "188.212.102.57", nil)

	// OUTPUT: Verify public API contract
	require.NoError(t, err, "Query should succeed")
	require.NotNil(t, result, "Result should not be nil")

	// Assert output contract
	assert.True(t, result.Online, "Server should be online")
	assert.Contains(t, result.Name, "BUCURESTI.FAIRSIDE.RO", "Server name should contain BUCURESTI")
	assert.Equal(t, "cs2", result.Type, "Type should be cs2")
	assert.Equal(t, "de_mirage", result.Map, "Map should be de_mirage")
	assert.Equal(t, 20, result.NumPlayers, "Should have 20 players")
	assert.Equal(t, 32, result.MaxPlayers, "Max players should be 32")
	assert.Equal(t, 0, result.Bots, "Should have 0 bots")
	assert.Equal(t, "1.41.3.5", result.Version, "Version should match")
	assert.False(t, result.Password, "Server should be public")
	assert.GreaterOrEqual(t, result.Ping, time.Duration(0), "Ping should be non-negative")

	// Check Raw contains protocol-specific data
	rawMap, ok := result.Raw.(map[string]interface{})
	assert.True(t, ok, "Raw should be a map")
	assert.Equal(t, uint16(730), rawMap["gameID"], "GameID should be 730")
	assert.Equal(t, true, rawMap["vac"], "VAC should be enabled")
	assert.Equal(t, "Counter-Strike 2", rawMap["game"], "Game name should match")

	// You can refactor Protocol internals - this test still passes!
}

func TestIntegration_CounterStrike_WithChallenge(t *testing.T) {
	// Challenge response
	challengeBuf := &bytes.Buffer{}
	binary.Write(challengeBuf, binary.LittleEndian, int32(-1))
	challengeBuf.WriteByte(0x41)                                  // Challenge type
	binary.Write(challengeBuf, binary.LittleEndian, int32(12345)) // Challenge number

	// Final response after challenge
	finalResponse := buildSourceEngineResponse(
		"Test Server",
		"de_dust2",
		"Counter-Strike 2",
		"1.0.0",
		10, // players
		16, // max
		2,  // bots
	)

	// Setup mock transport with challenge flow
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponseQueue["127.0.0.1:27015"] = [][]byte{
		challengeBuf.Bytes(), // First call returns challenge
		finalResponse,        // Second call returns actual data
	}

	// Create client
	client := NewClientWithDefaults(
		WithTransport(mockTransport),
	)

	// INPUT: Query server that requires challenge
	ctx := context.Background()
	result, err := client.Query(ctx, ServerCS2, "127.0.0.1", nil)

	// OUTPUT: Verify it handles challenge transparently
	require.NoError(t, err, "Should handle challenge automatically")
	assert.True(t, result.Online)
	assert.Equal(t, "Test Server", result.Name)
	assert.Equal(t, 10, result.NumPlayers)
	assert.Equal(t, 2, result.Bots)

	// Internal challenge handling can be refactored freely!
}

func TestIntegration_CustomPort(t *testing.T) {
	// Mock response
	mockResponse := buildSourceEngineResponse(
		"Test",
		"map",
		"Game",
		"v1",
		5,  // players
		20, // max
		1,  // bot
	)

	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:27016"] = mockResponse

	client := NewClientWithDefaults(
		WithTransport(mockTransport),
	)

	// INPUT: Query with custom port
	port := 27016
	result, err := client.Query(context.Background(), ServerCS2, "127.0.0.1", &port)

	// OUTPUT: Should use custom port
	require.NoError(t, err)
	assert.Equal(t, "Test", result.Name)
	assert.Equal(t, 5, result.NumPlayers)
}

func TestIntegration_ServerOffline(t *testing.T) {
	// Mock transport with timeout error (simulates offline server)
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPError = context.DeadlineExceeded

	client := NewClientWithDefaults(
		WithTransport(mockTransport),
		WithTimeout(100*time.Millisecond),
	)

	// INPUT: Query offline server (use IP to avoid DNS resolution)
	result, err := client.Query(context.Background(), ServerCS2, "192.0.2.1", nil)

	// OUTPUT: Should return error (timeout or connection error)
	require.Error(t, err)
	assert.Nil(t, result)
	// Error could be timeout or network error, just verify it failed
}

func TestIntegration_UnsupportedServerType(t *testing.T) {
	client := NewClientWithDefaults()

	// INPUT: Query with unsupported type
	result, err := client.Query(context.Background(), "unknowngame", "example.com", nil)

	// OUTPUT: Should return unsupported game error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestIntegration_MinecraftWithSRV(t *testing.T) {
	// This would test DNS SRV resolution + Minecraft protocol
	// but we'd need to mock DNS, so skip for now or create separate DNS integration tests
	t.Skip("SRV resolution requires DNS mocking")
}

// TestIntegration_APIContract verifies that the output structure matches expected format
func TestIntegration_APIContract(t *testing.T) {
	mockResponse := buildSourceEngineResponse(
		"Server",
		"map",
		"Game",
		"v1",
		15, // players
		32, // max
		0,  // bots
	)

	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:27015"] = mockResponse

	client := NewClientWithDefaults(WithTransport(mockTransport))
	result, err := client.Query(context.Background(), ServerCS2, "127.0.0.1", nil)

	require.NoError(t, err)

	// Verify API contract - these fields MUST exist and have correct types
	assert.IsType(t, true, result.Online, "Online must be bool")
	assert.IsType(t, "", result.Name, "Name must be string")
	assert.IsType(t, "", result.Map, "Map must be string")
	assert.IsType(t, 0, result.NumPlayers, "NumPlayers must be int")
	assert.IsType(t, 0, result.MaxPlayers, "MaxPlayers must be int")
	assert.IsType(t, []Player{}, result.Players, "Players must be []Player")
	assert.IsType(t, 0, result.Bots, "Bots must be int")
	assert.IsType(t, "", result.Type, "Type must be string")
	assert.IsType(t, "", result.Version, "Version must be string")
	assert.IsType(t, time.Duration(0), result.Ping, "Ping must be time.Duration")
	assert.IsType(t, "", result.Connect, "Connect must be string")
	assert.IsType(t, false, result.Password, "Password must be bool")
	// Raw is interface{} - contains ALL protocol-specific data (map[string]interface{} or custom struct)
	// For Source Engine (CS2), it contains parsed server data
	assert.NotNil(t, result.Raw, "Raw must not be nil")
	assert.IsType(t, time.Time{}, result.QueriedAt, "QueriedAt must be time.Time")

	// This test ensures the public API never breaks!
}
