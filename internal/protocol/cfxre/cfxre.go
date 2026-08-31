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

// defaultMaxPlayers is used when the server does not report a valid slot count.
const defaultMaxPlayers = 32

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

// DynamicResponse represents the /dynamic.json endpoint response. This is the
// canonical lightweight endpoint CFX.re servers use to advertise live player
// counts, and is more reliable than deriving counts from /players.json (which
// some servers truncate or hide).
type DynamicResponse struct {
	Clients      int    `json:"clients"`
	GameType     string `json:"gametype"`
	Hostname     string `json:"hostname"`
	MapName      string `json:"mapname"`
	SvMaxClients int    `json:"sv_maxclients"`
}

// Query queries a CFX.re server and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	// Normalize the address into an immutable endpoint value object that knows
	// how to build the well-known CFX.re data endpoints.
	endpoint := NewEndpoint(addr)

	// Fetch /info.json (required) and measure network latency (ping).
	pingStart := time.Now()
	infoData, err := p.Transport.SendHTTP(ctx, endpoint.Info())
	pingMs := int(time.Since(pingStart).Round(time.Millisecond).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /info.json: %w", err)
	}

	var info InfoResponse
	if err := json.Unmarshal(infoData, &info); err != nil {
		return nil, fmt.Errorf("failed to parse info response: %w", err)
	}

	// Fetch /players.json (required).
	playersData, err := p.Transport.SendHTTP(ctx, endpoint.Players())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /players.json: %w", err)
	}

	var players []PlayersResponse
	if err := json.Unmarshal(playersData, &players); err != nil {
		return nil, fmt.Errorf("failed to parse players response: %w", err)
	}

	// Fetch /dynamic.json (best-effort). It provides authoritative live counts;
	// servers that don't expose it simply leave the derived values in place.
	dynamic, hasDynamic := p.fetchDynamic(ctx, endpoint)

	result := &protocol.QueryResult{
		Online:     true,
		Name:       info.Vars.SvHostname,
		Map:        info.Vars.MapName,
		NumPlayers: len(players),
		MaxPlayers: parseMaxClients(info.Vars.SvMaxClients),
		Players:    toPlayers(players),
		Version:    info.Server,
		Ping:       pingMs,
		Raw: map[string]interface{}{
			"info":    info,    // Full /info.json response
			"players": players, // Full /players.json response
		},
	}

	if hasDynamic {
		p.applyDynamic(result, dynamic)
		result.Raw.(map[string]interface{})["dynamic"] = dynamic
	}

	return result, nil
}

// fetchDynamic retrieves and parses /dynamic.json. Any transport or parse
// failure is treated as "not available" so that servers without the endpoint
// still return a valid result.
func (p *Protocol) fetchDynamic(ctx context.Context, endpoint Endpoint) (DynamicResponse, bool) {
	data, err := p.Transport.SendHTTP(ctx, endpoint.Dynamic())
	if err != nil || len(data) == 0 {
		return DynamicResponse{}, false
	}

	var dynamic DynamicResponse
	if err := json.Unmarshal(data, &dynamic); err != nil {
		return DynamicResponse{}, false
	}
	return dynamic, true
}

// applyDynamic enriches the result with authoritative values from /dynamic.json,
// only overriding fields when the dynamic endpoint provides a better value.
func (p *Protocol) applyDynamic(result *protocol.QueryResult, dynamic DynamicResponse) {
	// Prefer the reported client count when it exceeds the (possibly hidden or
	// truncated) /players.json length.
	if dynamic.Clients > result.NumPlayers {
		result.NumPlayers = dynamic.Clients
	}
	if dynamic.SvMaxClients > 0 {
		result.MaxPlayers = dynamic.SvMaxClients
	}
	if result.Map == "" {
		result.Map = dynamic.MapName
	}
}

// parseMaxClients converts the string sv_maxclients var into an int, falling
// back to defaultMaxPlayers when unset or invalid.
func parseMaxClients(raw string) int {
	if raw == "" {
		return defaultMaxPlayers
	}
	if val, err := strconv.Atoi(raw); err == nil {
		return val
	}
	return defaultMaxPlayers
}

// toPlayers maps the CFX.re player list onto the unified protocol.Player slice.
// It always returns a non-nil slice to satisfy the QueryResult contract.
func toPlayers(players []PlayersResponse) []protocol.Player {
	result := make([]protocol.Player, len(players))
	for i, p := range players {
		result[i] = protocol.Player{Name: p.Name}
	}
	return result
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
