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

// cleanMOTD extracts plain text from MOTD (Message of the Day).
//
// A Minecraft MOTD can be a plain string or a chat component. A chat component
// is an object with an optional "text" string and an optional "extra" array of
// further components, which nest arbitrarily; elements of "extra" may themselves
// be plain strings. Real servers routinely put the whole MOTD inside "extra"
// with an empty top-level "text" (e.g. {"text":"","extra":[...]}) and colour
// each segment (even each character) with its own nested component. Extracting
// the text therefore has to walk the whole tree, not just read a flat "text".
func cleanMOTD(description interface{}) string {
	switch description.(type) {
	case string, map[string]interface{}, []interface{}:
		return stripColorCodes(componentText(description, 0))
	default:
		// nil, numbers, or any shape we do not recognise as a component.
		return "Unknown"
	}
}

// maxComponentDepth bounds the recursion over a chat component tree, guarding
// against a pathologically deep (or hostile) payload.
const maxComponentDepth = 64

// componentText recursively concatenates the visible text of a chat component,
// which may be a string, a component object ("text" plus a nested "extra"
// array), or an array of components. Colour/formatting metadata carried on
// sibling keys ("color", "bold", …) is intentionally ignored; only text is
// collected. Legacy "§" colour codes are stripped by the caller.
func componentText(v interface{}, depth int) string {
	if depth > maxComponentDepth {
		return ""
	}
	switch c := v.(type) {
	case string:
		return c
	case []interface{}:
		var b strings.Builder
		for _, e := range c {
			b.WriteString(componentText(e, depth+1))
		}
		return b.String()
	case map[string]interface{}:
		var b strings.Builder
		if text, ok := c["text"].(string); ok {
			b.WriteString(text)
		}
		if extra, ok := c["extra"]; ok {
			b.WriteString(componentText(extra, depth+1))
		}
		return b.String()
	default:
		return ""
	}
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
