// Package quake3 implements the idTech3 (Quake III engine) "getstatus" query
// over UDP, used by Quake III Arena and many idTech3-derived games.
package quake3

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// defaultPort is the default idTech3 UDP port.
	defaultPort = 27960
)

// connlessHeader is the 4-byte 0xFF prefix on every connectionless packet.
var connlessHeader = []byte{0xFF, 0xFF, 0xFF, 0xFF}

// ServerInfo holds the parsed statusResponse data.
type ServerInfo struct {
	Vars    map[string]string `json:"vars"`    // all cvars from the info string
	Players []PlayerInfo      `json:"players"` // parsed player lines
}

// PlayerInfo is a single player line from the status response.
type PlayerInfo struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Ping  int    `json:"ping"`
}

// Protocol implements the Quake3 getstatus query protocol.
type Protocol struct {
	protocol.BaseProtocol
}

// NewProtocol creates a new Quake3 protocol handler.
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{BaseProtocol: protocol.NewBaseProtocol(t, gameName)}
}

// Query sends a getstatus request and parses the statusResponse.
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	request := append(append([]byte{}, connlessHeader...), []byte("getstatus\n")...)

	pingStart := time.Now()
	response, err := p.Transport.SendUDP(ctx, addr, request)
	pingMs := int(time.Since(pingStart).Round(time.Millisecond).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("failed to send getstatus: %w", err)
	}

	info, err := parseStatusResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse statusResponse: %w", err)
	}

	maxPlayers := atoiSafe(info.Vars["sv_maxclients"])
	if maxPlayers == 0 {
		maxPlayers = atoiSafe(info.Vars["sv_maxplayers"])
	}

	players := make([]protocol.Player, len(info.Players))
	for i, pl := range info.Players {
		players[i] = protocol.Player{Name: pl.Name, Score: pl.Score}
	}

	return &protocol.QueryResult{
		Online:     true,
		Name:       info.Vars["sv_hostname"],
		Map:        info.Vars["mapname"],
		NumPlayers: len(info.Players),
		MaxPlayers: maxPlayers,
		Players:    players,
		Version:    firstNonEmpty(info.Vars["version"], info.Vars["gamename"]),
		Ping:       pingMs,
		Password:   info.Vars["g_needpass"] == "1" || info.Vars["needpass"] == "1",
		Raw:        info,
	}, nil
}

// parseStatusResponse parses a connectionless statusResponse packet.
//
// Layout: [0xFF 0xFF 0xFF 0xFF]statusResponse\n\key\value\...\n
//
//	score ping "name"\n (one line per player)
func parseStatusResponse(data []byte) (*ServerInfo, error) {
	if len(data) < 4 || !bytes.Equal(data[:4], connlessHeader) {
		return nil, fmt.Errorf("missing connectionless header")
	}
	body := string(data[4:])
	body = strings.TrimPrefix(body, "statusResponse\n")

	lines := strings.Split(body, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty status response")
	}

	info := &ServerInfo{Vars: parseInfoString(lines[0]), Players: []PlayerInfo{}}

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if pl, ok := parsePlayerLine(line); ok {
			info.Players = append(info.Players, pl)
		}
	}
	return info, nil
}

// parseInfoString parses a "\key\value\key\value" cvar string.
func parseInfoString(s string) map[string]string {
	vars := map[string]string{}
	parts := strings.Split(strings.TrimPrefix(s, "\\"), "\\")
	for i := 0; i+1 < len(parts); i += 2 {
		vars[parts[i]] = parts[i+1]
	}
	return vars
}

// parsePlayerLine parses a `score ping "name"` player line.
func parsePlayerLine(line string) (PlayerInfo, bool) {
	// Split off the quoted name first.
	first := strings.Index(line, "\"")
	last := strings.LastIndex(line, "\"")
	name := ""
	nums := line
	if first >= 0 && last > first {
		name = line[first+1 : last]
		nums = strings.TrimSpace(line[:first])
	}
	scoreping := strings.Fields(nums)
	if len(scoreping) < 2 {
		return PlayerInfo{}, false
	}
	return PlayerInfo{
		Name:  name,
		Score: atoiSafe(scoreping[0]),
		Ping:  atoiSafe(scoreping[1]),
	}, true
}

func atoiSafe(s string) int {
	if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return v
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// Name returns the protocol name.
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (idTech3/getstatus)", p.GameName)
}

// DefaultPort returns the default idTech3 UDP port.
func (p *Protocol) DefaultPort() int { return defaultPort }

// SupportsSRV indicates idTech3 does not use SRV records.
func (p *Protocol) SupportsSRV() bool { return false }

// SRVService returns an empty string (not used).
func (p *Protocol) SRVService() string { return "" }
