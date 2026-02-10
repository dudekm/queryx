package cfxre

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
)

func TestProtocol_Name(t *testing.T) {
	p := NewProtocol(nil, "FiveM")
	assert.Equal(t, "FiveM (CFX.re)", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	tests := []struct {
		gameName     string
		expectedPort int
	}{
		{"FiveM", 30120},
		{"RedM", 30120},
		{"Alt:V", 7788},
	}

	for _, tt := range tests {
		t.Run(tt.gameName, func(t *testing.T) {
			p := NewProtocol(nil, tt.gameName)
			assert.Equal(t, tt.expectedPort, p.DefaultPort())
		})
	}
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil, "FiveM")
	assert.False(t, p.SupportsSRV())
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock info response
	infoResponse := InfoResponse{
		Server: "v1.0.0.5555 win32",
	}
	infoResponse.Vars.SvHostname = "Test FiveM Server"
	infoResponse.Vars.SvMaxClients = "32"
	infoResponse.Vars.MapName = "Los Santos"
	infoResponse.Vars.GameName = "gta5"
	infoResponse.Vars.Tags = "roleplay,economy"

	infoData, _ := json.Marshal(infoResponse)
	mockTransport.HTTPResponses["http://test:30120/info.json"] = infoData

	// Build mock players response
	players := []PlayersResponse{
		{Name: "Player1", ID: 1, Ping: 50},
		{Name: "Player2", ID: 2, Ping: 75},
		{Name: "Player3", ID: 3, Ping: 100},
	}
	playersData, _ := json.Marshal(players)
	mockTransport.HTTPResponses["http://test:30120/players.json"] = playersData

	p := NewProtocol(mockTransport, "FiveM")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:30120")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test FiveM Server", result.Name)
	assert.Equal(t, "Los Santos", result.Map)
	assert.Equal(t, 3, result.NumPlayers)
	assert.Equal(t, 32, result.MaxPlayers)
	assert.Equal(t, "v1.0.0.5555 win32", result.Version)
	assert.Equal(t, 3, len(result.Players))
	assert.Equal(t, "Player1", result.Players[0].Name)

	// Check Raw contains full CFX.re response
	assert.NotNil(t, result.Raw)
	rawMap, ok := result.Raw.(map[string]interface{})
	assert.True(t, ok, "Raw should be a map")
	rawInfo, ok := rawMap["info"].(InfoResponse)
	assert.True(t, ok, "Raw should contain info")
	assert.Equal(t, "roleplay,economy", rawInfo.Vars.Tags)
}

func TestProtocol_Query_WithHTTPPrefix(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock responses
	infoResponse := InfoResponse{
		Server: "v1.0.0",
	}
	infoResponse.Vars.SvHostname = "Test Server"
	infoResponse.Vars.SvMaxClients = "48"

	infoData, _ := json.Marshal(infoResponse)
	mockTransport.HTTPResponses["http://example.com:30120/info.json"] = infoData

	players := []PlayersResponse{}
	playersData, _ := json.Marshal(players)
	mockTransport.HTTPResponses["http://example.com:30120/players.json"] = playersData

	p := NewProtocol(mockTransport, "FiveM")
	ctx := context.Background()

	// Query with http:// prefix
	result, err := p.Query(ctx, "http://example.com:30120")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Server", result.Name)
	assert.Equal(t, 0, result.NumPlayers)
	assert.Equal(t, 48, result.MaxPlayers)
}

func TestProtocol_Query_InvalidMaxPlayers(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock with invalid max players
	infoResponse := InfoResponse{
		Server: "v1.0.0",
	}
	infoResponse.Vars.SvHostname = "Test Server"
	infoResponse.Vars.SvMaxClients = "invalid" // Invalid number

	infoData, _ := json.Marshal(infoResponse)
	mockTransport.HTTPResponses["http://test:30120/info.json"] = infoData

	players := []PlayersResponse{}
	playersData, _ := json.Marshal(players)
	mockTransport.HTTPResponses["http://test:30120/players.json"] = playersData

	p := NewProtocol(mockTransport, "FiveM")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:30120")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 32, result.MaxPlayers) // Should default to 32
}

func TestProtocol_Query_EmptyMaxPlayers(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock with empty max players
	infoResponse := InfoResponse{
		Server: "v1.0.0",
	}
	infoResponse.Vars.SvHostname = "Test Server"
	infoResponse.Vars.SvMaxClients = "" // Empty

	infoData, _ := json.Marshal(infoResponse)
	mockTransport.HTTPResponses["http://test:30120/info.json"] = infoData

	players := []PlayersResponse{}
	playersData, _ := json.Marshal(players)
	mockTransport.HTTPResponses["http://test:30120/players.json"] = playersData

	p := NewProtocol(mockTransport, "FiveM")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:30120")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 32, result.MaxPlayers) // Should default to 32
}

func TestProtocol_Query_InfoError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.HTTPError = assert.AnError

	p := NewProtocol(mockTransport, "FiveM")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:30120")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch /info.json")
}

func TestProtocol_Query_PlayersError(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Info succeeds
	infoResponse := InfoResponse{
		Server: "v1.0.0",
	}
	infoResponse.Vars.SvHostname = "Test Server"
	infoResponse.Vars.SvMaxClients = "32"

	infoData, _ := json.Marshal(infoResponse)
	mockTransport.HTTPResponses["http://test:30120/info.json"] = infoData

	// But players endpoint returns error (simulate via invalid JSON)
	mockTransport.HTTPResponses["http://test:30120/players.json"] = []byte("invalid json")

	p := NewProtocol(mockTransport, "FiveM")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:30120")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse players response")
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "FiveM")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Query(ctx, "test:30120")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProtocol_AllGames(t *testing.T) {
	games := []struct {
		name         string
		gameName     string
		expectedPort int
	}{
		{"FiveM", "fivem", 30120},
		{"RedM", "redm", 30120},
		{"Alt:V", "altv", 7788},
	}

	for _, game := range games {
		t.Run(game.name, func(t *testing.T) {
			mockTransport := transport.NewMockTransport()

			// Build mock responses
			infoResponse := InfoResponse{
				Server: "v1.0.0",
			}
			infoResponse.Vars.SvHostname = "Test " + game.name + " Server"
			infoResponse.Vars.SvMaxClients = "64"
			infoResponse.Vars.MapName = "test_map"

			infoData, _ := json.Marshal(infoResponse)
			mockTransport.HTTPResponses["http://test:30120/info.json"] = infoData

			players := []PlayersResponse{
				{Name: "Player1", ID: 1},
				{Name: "Player2", ID: 2},
			}
			playersData, _ := json.Marshal(players)
			mockTransport.HTTPResponses["http://test:30120/players.json"] = playersData

			proto := NewProtocol(mockTransport, game.gameName)
			result, err := proto.Query(context.Background(), "test:30120")

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Online)
			assert.Equal(t, "Test "+game.name+" Server", result.Name)
			assert.Equal(t, "test_map", result.Map)
			assert.Equal(t, 2, result.NumPlayers)
			assert.Equal(t, 64, result.MaxPlayers)
			assert.Equal(t, game.expectedPort, proto.DefaultPort())
		})
	}
}

func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		gameName     string
		expectedPort int
	}{
		{"fivem", 30120},
		{"FiveM", 30120},
		{"redm", 30120},
		{"RedM", 30120},
		{"altv", 7788},
		{"Alt:V", 7788},
		{"unknown", 30120}, // fallback to FiveM port
	}

	for _, tt := range tests {
		t.Run(tt.gameName, func(t *testing.T) {
			port := GetDefaultPort(tt.gameName)
			assert.Equal(t, tt.expectedPort, port)
		})
	}
}
