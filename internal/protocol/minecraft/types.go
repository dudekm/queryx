package minecraft

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// serverResponse represents the JSON response from a Minecraft server
type serverResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"sample"`
	} `json:"players"`
	Description interface{} `json:"description"` // Can be string or object
	Favicon     string      `json:"favicon"`
}

// cleanMOTD extracts plain text from MOTD (Message of the Day)
// Minecraft MOTD can be a string or a complex JSON object with formatting
func cleanMOTD(description interface{}) string {
	switch v := description.(type) {
	case string:
		return stripColorCodes(v)
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok {
			return stripColorCodes(text)
		}
		// Handle "extra" array
		if extra, ok := v["extra"].([]interface{}); ok {
			var parts []string
			for _, e := range extra {
				if eMap, ok := e.(map[string]interface{}); ok {
					if text, ok := eMap["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
			return stripColorCodes(strings.Join(parts, ""))
		}
	}
	return "Unknown"
}

// stripColorCodes removes Minecraft color codes (§x)
func stripColorCodes(s string) string {
	result := ""
	skip := false
	for _, c := range s {
		if c == '§' {
			skip = true
			continue
		}
		if skip {
			skip = false
			continue
		}
		result += string(c)
	}
	return strings.TrimSpace(result)
}

// parseAddr parses an address string into host and port
func parseAddr(addr string) (host string, port int, err error) {
	hostStr, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid address format: %w", err)
	}

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %w", err)
	}

	return hostStr, portInt, nil
}
