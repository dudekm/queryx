package minecraft

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanMOTD_String(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "simple string",
			input:    "A Minecraft Server",
			expected: "A Minecraft Server",
		},
		{
			name:     "string with color codes",
			input:    "§aGreen §bBlue §cRed",
			expected: "Green Blue Red",
		},
		{
			name: "complex object with text",
			input: map[string]interface{}{
				"text": "Server MOTD",
			},
			expected: "Server MOTD",
		},
		{
			name: "object with extra array",
			input: map[string]interface{}{
				"extra": []interface{}{
					map[string]interface{}{"text": "Hello "},
					map[string]interface{}{"text": "World"},
				},
			},
			expected: "Hello World",
		},
		{
			// Regression: empty top-level "text" must not short-circuit before
			// the real MOTD in "extra" is read.
			name: "empty text with extra",
			input: map[string]interface{}{
				"text": "",
				"extra": []interface{}{
					map[string]interface{}{"text": "Hello "},
					map[string]interface{}{"text": "World"},
				},
			},
			expected: "Hello World",
		},
		{
			name: "non-empty text combined with extra",
			input: map[string]interface{}{
				"text": "A ",
				"extra": []interface{}{
					map[string]interface{}{"text": "B "},
					map[string]interface{}{"text": "C"},
				},
			},
			expected: "A B C",
		},
		{
			// Elements of "extra" can be bare strings, not only objects.
			name: "extra with string elements",
			input: map[string]interface{}{
				"text":  "",
				"extra": []interface{}{"Hello ", "World"},
			},
			expected: "Hello World",
		},
		{
			// "extra" nests: a component in "extra" may carry its own "extra".
			name: "nested extra",
			input: map[string]interface{}{
				"text": "",
				"extra": []interface{}{
					map[string]interface{}{
						"text": "",
						"extra": []interface{}{
							map[string]interface{}{"text": "Hel"},
							map[string]interface{}{"text": "lo"},
						},
					},
					map[string]interface{}{"text": " World"},
				},
			},
			expected: "Hello World",
		},
		{
			// Shape observed in the wild (blokowo.pl): empty top "text",
			// per-character coloured components in nested "extra", legacy "§"
			// codes mixed in.
			name: "real-world empty text, deep extra, legacy codes",
			input: map[string]interface{}{
				"text": "",
				"extra": []interface{}{
					map[string]interface{}{
						"text": "",
						"extra": []interface{}{
							map[string]interface{}{"color": "yellow", "text": "B"},
							map[string]interface{}{"color": "yellow", "text": "L"},
							map[string]interface{}{"color": "yellow", "text": "O"},
						},
					},
					map[string]interface{}{"color": "white", "text": " §7- Survival"},
				},
			},
			expected: "BLO - Survival",
		},
		{
			// An array MOTD with no wrapping object.
			name:     "top-level array",
			input:    []interface{}{map[string]interface{}{"text": "Foo "}, "Bar"},
			expected: "Foo Bar",
		},
		{
			name:     "unknown type",
			input:    123,
			expected: "Unknown",
		},
		{
			name:     "nil description",
			input:    nil,
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMOTD(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripColorCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no color codes",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "with color codes",
			input:    "§aGreen §bBlue §cRed",
			expected: "Green Blue Red",
		},
		{
			name:     "multiple consecutive codes",
			input:    "§a§lBold Green",
			expected: "Bold Green",
		},
		{
			name:     "trailing whitespace",
			input:    "  Text  ",
			expected: "Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripColorCodes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAddr(t *testing.T) {
	tests := []struct {
		name         string
		addr         string
		expectedHost string
		expectedPort int
		expectError  bool
	}{
		{
			name:         "valid address",
			addr:         "127.0.0.1:25565",
			expectedHost: "127.0.0.1",
			expectedPort: 25565,
			expectError:  false,
		},
		{
			name:         "hostname with port",
			addr:         "hypixel.net:25565",
			expectedHost: "hypixel.net",
			expectedPort: 25565,
			expectError:  false,
		},
		{
			name:         "custom port",
			addr:         "localhost:30000",
			expectedHost: "localhost",
			expectedPort: 30000,
			expectError:  false,
		},
		{
			name:        "missing port",
			addr:        "example.com",
			expectError: true,
		},
		{
			name:        "invalid port",
			addr:        "example.com:abc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := parseAddr(tt.addr)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHost, host)
				assert.Equal(t, tt.expectedPort, port)
			}
		})
	}
}
