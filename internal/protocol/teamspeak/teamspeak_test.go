package teamspeak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocol_Name(t *testing.T) {
	p := NewProtocol(nil, "TeamSpeak 3")
	assert.Equal(t, "TeamSpeak 3 (ServerQuery)", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	p := NewProtocol(nil, "TeamSpeak 3")
	assert.Equal(t, 10011, p.DefaultPort())
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil, "TeamSpeak 3")
	assert.False(t, p.SupportsSRV())
}

func TestParseServerInfo(t *testing.T) {
	// Mock TeamSpeak serverinfo response
	response := "virtualserver_name=Test\\sTeamSpeak\\sServer virtualserver_clientsonline=15 virtualserver_maxclients=32 virtualserver_queryclientsonline=1 virtualserver_version=3.13.7 virtualserver_platform=Linux virtualserver_flag_password=0 virtualserver_uptime=123456 virtualserver_channelsonline=5"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test TeamSpeak Server", result.Name)
	assert.Equal(t, 14, result.NumPlayers) // 15 - 1 query client
	assert.Equal(t, 32, result.MaxPlayers)
	assert.Equal(t, "3.13.7", result.Version)
	assert.False(t, result.Password)

	// Check Raw contains protocol-specific data
	rawMap, ok := result.Raw.(map[string]interface{})
	assert.True(t, ok, "Raw should be a map")
	assert.Equal(t, "Linux", rawMap["platform"])
	assert.Equal(t, "123456", rawMap["uptime"])
	assert.Equal(t, 5, rawMap["channels_online"])
}

func TestParseServerInfo_WithPassword(t *testing.T) {
	response := "virtualserver_name=Private\\sServer virtualserver_clientsonline=5 virtualserver_maxclients=10 virtualserver_queryclientsonline=0 virtualserver_version=3.13.7 virtualserver_flag_password=1"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	assert.True(t, result.Password)
	assert.Equal(t, "Private Server", result.Name)
	assert.Equal(t, 5, result.NumPlayers)
}

func TestParseServerInfo_MissingServerName(t *testing.T) {
	response := "virtualserver_clientsonline=10 virtualserver_maxclients=32"

	result, err := parseServerInfo(response)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "missing server name")
}

func TestParseServerInfo_EmptyResponse(t *testing.T) {
	response := ""

	result, err := parseServerInfo(response)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUnescapeValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Test\\sServer",
			expected: "Test Server",
		},
		{
			input:    "Value\\sWith\\sSpaces",
			expected: "Value With Spaces",
		},
		{
			input:    "Path\\/To\\/Something",
			expected: "Path/To/Something",
		},
		{
			input:    "Pipe\\pSeparated",
			expected: "Pipe|Separated",
		},
		{
			input:    "Backslash\\\\Test",
			expected: "Backslash\\Test",
		},
		{
			input:    "NoEscapes",
			expected: "NoEscapes",
		},
		{
			input:    "Multiple\\sEscapes\\sand\\smore",
			expected: "Multiple Escapes and more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := unescapeValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"10", 10},
		{"123456", 123456},
		{"", 0},
		{"invalid", 0},
		{"-5", -5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseServerInfo_ComplexName(t *testing.T) {
	// Test with complex server name containing special characters
	response := "virtualserver_name=My\\sAwesome\\sServer\\p[EU] virtualserver_clientsonline=20 virtualserver_maxclients=64 virtualserver_queryclientsonline=2 virtualserver_version=3.13.7 virtualserver_flag_password=0"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	assert.Equal(t, "My Awesome Server|[EU]", result.Name)
	assert.Equal(t, 18, result.NumPlayers) // 20 - 2 query clients
	assert.Equal(t, 64, result.MaxPlayers)
}

func TestParseServerInfo_NoQueryClients(t *testing.T) {
	// Test when query clients count is 0
	response := "virtualserver_name=Test\\sServer virtualserver_clientsonline=10 virtualserver_maxclients=32 virtualserver_queryclientsonline=0 virtualserver_version=3.13.7"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	assert.Equal(t, 10, result.NumPlayers) // Should not subtract anything
}

func TestParseServerInfo_MoreQueryClientsThanTotal(t *testing.T) {
	// Edge case: more query clients than total (shouldn't happen but handle gracefully)
	response := "virtualserver_name=Test\\sServer virtualserver_clientsonline=2 virtualserver_maxclients=32 virtualserver_queryclientsonline=5 virtualserver_version=3.13.7"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	// Should fall back to showing all clients when calculation would be negative
	assert.Equal(t, 2, result.NumPlayers)
}

func TestParseServerInfo_AllExtraFields(t *testing.T) {
	response := "virtualserver_name=Full\\sServer virtualserver_clientsonline=10 virtualserver_maxclients=32 virtualserver_queryclientsonline=1 virtualserver_version=3.13.7 virtualserver_platform=Windows virtualserver_uptime=987654 virtualserver_created=1234567890 virtualserver_codec_encryption_mode=1 virtualserver_channelsonline=8 virtualserver_flag_password=0"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	assert.Equal(t, "Full Server", result.Name)

	// Check Raw contains protocol-specific data
	rawMap, ok := result.Raw.(map[string]interface{})
	assert.True(t, ok, "Raw should be a map")
	assert.Equal(t, "Windows", rawMap["platform"])
	assert.Equal(t, "987654", rawMap["uptime"])
	assert.Equal(t, "1234567890", rawMap["created"])
	assert.Equal(t, "1", rawMap["codec_encryption"])
	assert.Equal(t, 8, rawMap["channels_online"])
}

func TestParseServerInfo_MinimalResponse(t *testing.T) {
	// Test with only required fields
	response := "virtualserver_name=Minimal virtualserver_clientsonline=5 virtualserver_maxclients=10 virtualserver_queryclientsonline=0 virtualserver_version=3.13.7"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	assert.Equal(t, "Minimal", result.Name)
	assert.Equal(t, 5, result.NumPlayers)
	assert.Equal(t, 10, result.MaxPlayers)
	assert.Equal(t, "3.13.7", result.Version)
}

func TestParseServerInfo_ZeroClients(t *testing.T) {
	response := "virtualserver_name=Empty\\sServer virtualserver_clientsonline=0 virtualserver_maxclients=100 virtualserver_queryclientsonline=0 virtualserver_version=3.13.7"

	result, err := parseServerInfo(response)

	assert.NoError(t, err)
	assert.Equal(t, "Empty Server", result.Name)
	assert.Equal(t, 0, result.NumPlayers)
	assert.Equal(t, 100, result.MaxPlayers)
}
