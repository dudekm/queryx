package gamespy

import (
	"context"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
)

func TestProtocol_Name(t *testing.T) {
	p := NewProtocol(nil, "ARMA 3")
	assert.Equal(t, "ARMA 3 (GameSpy)", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	tests := []struct {
		gameName     string
		expectedPort int
	}{
		{"ARMA 2", 2302},
		{"ARMA 3", 2302},
		{"DayZ", 2302},
		{"Day of Dragons", 7777},
	}

	for _, tt := range tests {
		t.Run(tt.gameName, func(t *testing.T) {
			p := NewProtocol(nil, tt.gameName)
			assert.Equal(t, tt.expectedPort, p.DefaultPort())
		})
	}
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil, "ARMA 3")
	assert.False(t, p.SupportsSRV())
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Build mock GameSpy response
	response := buildMockGameSpyResponse(
		"Test ARMA 3 Server",
		"Altis",
		"10",
		"64",
		"1.2.3.4567",
	)

	mockTransport.UDPResponses["test:2302"] = response

	p := NewProtocol(mockTransport, "ARMA 3")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:2302")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test ARMA 3 Server", result.Name)
	assert.Equal(t, "Altis", result.Map)
	assert.Equal(t, 10, result.NumPlayers)
	assert.Equal(t, 64, result.MaxPlayers)
	assert.Equal(t, "1.2.3.4567", result.Version)
}

func TestProtocol_Query_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPError = assert.AnError

	p := NewProtocol(mockTransport, "ARMA 3")
	ctx := context.Background()

	result, err := p.Query(ctx, "test:2302")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send GameSpy query")
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "ARMA 3")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Query(ctx, "test:2302")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected map[string]string
		wantErr  bool
	}{
		{
			name: "basic key-value pairs",
			data: []byte("\\hostname\\Test Server\\mapname\\Altis\\numplayers\\10\\maxplayers\\64\\"),
			expected: map[string]string{
				"hostname":   "Test Server",
				"mapname":    "Altis",
				"numplayers": "10",
				"maxplayers": "64",
			},
			wantErr: false,
		},
		{
			name: "with null terminators",
			data: []byte("\\hostname\\Test Server\x00\\mapname\\Altis\x00\\"),
			expected: map[string]string{
				"hostname": "Test Server",
				"mapname":  "Altis",
			},
			wantErr: false,
		},
		{
			name: "with whitespace",
			data: []byte("\\hostname\\  Test Server  \\mapname\\  Altis  \\"),
			expected: map[string]string{
				"hostname": "Test Server",
				"mapname":  "Altis",
			},
			wantErr: false,
		},
		{
			name:     "empty response",
			data:     []byte(""),
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid format",
			data:     []byte("hostname=test,mapname=altis"),
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseKeyValue(tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	data := map[string]string{
		"hostname": "Test Server",
		"mapname":  "Altis",
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"hostname", "Test Server"},
		{"mapname", "Altis"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := getString(data, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInt(t *testing.T) {
	data := map[string]string{
		"numplayers": "10",
		"maxplayers": "64",
		"invalid":    "not_a_number",
	}

	tests := []struct {
		key      string
		expected int
	}{
		{"numplayers", 10},
		{"maxplayers", 64},
		{"invalid", 0},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := getInt(data, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		gameName     string
		expectedPort int
	}{
		{"arma2", 2302},
		{"ARMA 2", 2302},
		{"arma3", 2302},
		{"ARMA 3", 2302},
		{"dayz", 2302},
		{"DayZ", 2302},
		{"dayofdragons", 7777},
		{"Day of Dragons", 7777},
		{"unknown_game", 2302}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.gameName, func(t *testing.T) {
			port := GetDefaultPort(tt.gameName)
			assert.Equal(t, tt.expectedPort, port)
		})
	}
}

func TestProtocol_AllGames(t *testing.T) {
	games := []struct {
		name         string
		gameName     string
		expectedPort int
	}{
		{"ARMA 2", "arma2", 2302},
		{"ARMA 3", "arma3", 2302},
		{"DayZ", "dayz", 2302},
		{"Day of Dragons", "dayofdragons", 7777},
	}

	for _, game := range games {
		t.Run(game.name, func(t *testing.T) {
			mockTransport := transport.NewMockTransport()

			// Build mock response
			response := buildMockGameSpyResponse(
				"Test "+game.name+" Server",
				"test_map",
				"5",
				"32",
				"1.0.0",
			)

			mockTransport.UDPResponses["test:2302"] = response

			proto := NewProtocol(mockTransport, game.gameName)
			result, err := proto.Query(context.Background(), "test:2302")

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Online)
			assert.Equal(t, "Test "+game.name+" Server", result.Name)
			assert.Equal(t, "test_map", result.Map)
			assert.Equal(t, 5, result.NumPlayers)
			assert.Equal(t, 32, result.MaxPlayers)
			assert.Equal(t, game.expectedPort, proto.DefaultPort())
		})
	}
}

func TestBuildQueryPacket(t *testing.T) {
	packet := buildQueryPacket()
	assert.NotNil(t, packet)
	assert.Equal(t, []byte("\\status\\"), packet)
}

// buildMockGameSpyResponse builds a mock GameSpy response for testing
func buildMockGameSpyResponse(hostname, mapname, numplayers, maxplayers, gamever string) []byte {
	response := "\\hostname\\" + hostname +
		"\\mapname\\" + mapname +
		"\\numplayers\\" + numplayers +
		"\\maxplayers\\" + maxplayers +
		"\\gamever\\" + gamever +
		"\\password\\0\\"

	return []byte(response)
}
