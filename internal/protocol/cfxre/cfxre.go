package cfxre

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// Default HTTP ports for CFX.re games
	defaultPortFiveM = 30120
	defaultPortRedM  = 30120
	defaultPortAltV  = 7788
)

// Protocol implements CFX.re HTTP Query Protocol
// Used by FiveM, RedM, and Alt:V
type Protocol struct {
	protocol.BaseProtocol
}

// NewProtocol creates a new CFX.re protocol handler
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{
		BaseProtocol: protocol.NewBaseProtocol(t, gameName),
	}
}

// InfoResponse represents the /info.json endpoint response
type InfoResponse struct {
	Vars struct {
		SvHostname          string `json:"sv_hostname"`
		SvMaxClients        string `json:"sv_maxclients"`
		SvProjectName       string `json:"sv_projectName"`
		SvProjectDesc       string `json:"sv_projectDesc"`
		Tags                string `json:"tags"`
		GameName            string `json:"gamename"`
		MapName             string `json:"mapname"`
		GameType            string `json:"gametype"`
		SvScriptHookAllowed string `json:"sv_scriptHookAllowed"`
	} `json:"vars"`
	Server    string   `json:"server"`
	Resources []string `json:"resources"`
}

// PlayersResponse represents a single player from /players.json
type PlayersResponse struct {
	Name        string   `json:"name"`
	ID          int      `json:"id"`
	Identifiers []string `json:"identifiers"`
	Ping        int      `json:"ping"`
}

// Query queries a CFX.re server and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	// CFX.re servers expose HTTP endpoints
	// Ensure we have http:// prefix
	baseURL := addr
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		baseURL = "http://" + addr
	}

	// Fetch /info.json
	pingStart := time.Now()
	infoData, err := p.Transport.SendHTTP(ctx, baseURL+"/info.json")
	pingDuration := time.Since(pingStart)
	pingMs := float64(pingDuration.Microseconds()) / 1000.0

	if err != nil {
		return nil, fmt.Errorf("failed to fetch /info.json: %w", err)
	}

	var info InfoResponse
	if err := json.Unmarshal(infoData, &info); err != nil {
		return nil, fmt.Errorf("failed to parse info response: %w", err)
	}

	// Fetch /players.json
	playersData, err := p.Transport.SendHTTP(ctx, baseURL+"/players.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /players.json: %w", err)
	}

	var players []PlayersResponse
	if err := json.Unmarshal(playersData, &players); err != nil {
		return nil, fmt.Errorf("failed to parse players response: %w", err)
	}

	// Parse max players (default to 32 if not set or invalid)
	maxPlayers := 32
	if info.Vars.SvMaxClients != "" {
		if val, err := strconv.Atoi(info.Vars.SvMaxClients); err == nil {
			maxPlayers = val
		}
	}

	// Build result with ALL data in raw
	result := &protocol.QueryResult{
		Online:     true,
		Name:       info.Vars.SvHostname,
		Map:        info.Vars.MapName,
		NumPlayers: len(players),
		MaxPlayers: maxPlayers,
		Players:    []protocol.Player{}, // Initialize as empty array, not nil
		Version:    info.Server,
		Ping:       pingMs,
		Raw: map[string]interface{}{
			"info":    info,    // Full /info.json response
			"players": players, // Full /players.json response
		},
	}

	// Convert players to protocol.Player format
	if len(players) > 0 {
		result.Players = make([]protocol.Player, len(players))
		for i, p := range players {
			result.Players[i] = protocol.Player{
				Name: p.Name,
			}
		}
	}

	return result, nil
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (CFX.re)", p.GameName)
}

// DefaultPort returns the default HTTP port for the game
func (p *Protocol) DefaultPort() int {
	return GetDefaultPort(p.GameName)
}

// SupportsSRV indicates that CFX.re does not use SRV records
func (p *Protocol) SupportsSRV() bool {
	return false
}

// SRVService returns empty string (not used)
func (p *Protocol) SRVService() string {
	return ""
}

// GetDefaultPort returns the default port for a CFX.re game
func GetDefaultPort(gameName string) int {
	normalized := strings.ToLower(gameName)
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, "-", "")

	switch normalized {
	case "altv":
		return defaultPortAltV
	case "fivem", "redm":
		return defaultPortFiveM
	default:
		return defaultPortFiveM
	}
}
