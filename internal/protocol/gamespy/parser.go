package gamespy

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// parseKeyValue parses GameSpy key-value format: \key\value\key\value\
func parseKeyValue(data []byte) (map[string]string, error) {
	result := make(map[string]string)

	// GameSpy responses use backslash as delimiter
	// Format: \key1\value1\key2\value2\
	pairs := bytes.Split(data, []byte("\\"))

	// First element is empty (data starts with \), so start from index 1
	for i := 1; i < len(pairs)-1; i += 2 {
		if i+1 < len(pairs) {
			key := string(pairs[i])
			value := string(pairs[i+1])

			// Clean up any null terminators or whitespace
			key = strings.TrimSpace(strings.Trim(key, "\x00"))
			value = strings.TrimSpace(strings.Trim(value, "\x00"))

			if key != "" {
				result[key] = value
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid key-value pairs found in response")
	}

	return result, nil
}

// getString safely retrieves a string value from the parsed data
func getString(data map[string]string, key string) string {
	if val, ok := data[key]; ok {
		return val
	}
	return ""
}

// getInt safely retrieves an integer value from the parsed data
func getInt(data map[string]string, key string) int {
	if val, ok := data[key]; ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

// DefaultPorts maps game identifiers to their default query ports
var DefaultPorts = map[string]int{
	"arma2":        2302,
	"arma3":        2302,
	"dayz":         2302,
	"dayofdragons": 7777,
}

// GetDefaultPort returns the default port for a game, falling back to 2302
func GetDefaultPort(gameName string) int {
	normalized := strings.ToLower(gameName)
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, "-", "")

	if port, ok := DefaultPorts[normalized]; ok {
		return port
	}

	// Default GameSpy port (used by ARMA games)
	return 2302
}
