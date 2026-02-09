package teamspeak

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/transport"
)

const (
	// Default TeamSpeak ServerQuery port
	defaultPort = 10011
)

// Protocol implements TeamSpeak 3 ServerQuery Protocol
type Protocol struct {
	transport transport.Transport
	gameName  string
}

// NewProtocol creates a new TeamSpeak protocol handler
func NewProtocol(t transport.Transport, gameName string) *Protocol {
	return &Protocol{
		transport: t,
		gameName:  gameName,
	}
}

// Query queries a TeamSpeak 3 server using ServerQuery and returns the result
func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	// Parse address to ensure we have host:port format
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		// If no port specified, add default port
		host = addr
		portStr = strconv.Itoa(defaultPort)
		addr = net.JoinHostPort(host, portStr)
	}

	// Connect to TeamSpeak ServerQuery (TCP)
	pingStart := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TeamSpeak server: %w", err)
	}
	defer conn.Close()

	// Set read deadline
	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetDeadline(deadline)
	} else {
		conn.SetDeadline(time.Now().Add(10 * time.Second))
	}

	reader := bufio.NewReader(conn)

	// Read welcome message (2 lines: welcome text + TS3)
	_, err = reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read welcome message: %w", err)
	}

	_, err = reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read TS3 line: %w", err)
	}

	// Select first virtual server (use sid=1)
	_, err = conn.Write([]byte("use sid=1\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to send use command: %w", err)
	}

	// Read use command response
	_, err = reader.ReadString('\n') // Response line
	if err != nil {
		return nil, fmt.Errorf("failed to read use response: %w", err)
	}

	// Send serverinfo command
	_, err = conn.Write([]byte("serverinfo\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to send serverinfo command: %w", err)
	}

	// Read serverinfo response
	infoLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read serverinfo response: %w", err)
	}

	ping := time.Since(pingStart)

	// Parse serverinfo response
	result, err := parseServerInfo(infoLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse serverinfo: %w", err)
	}

	result.Ping = ping

	// Send quit command
	conn.Write([]byte("quit\n"))

	return result, nil
}

// parseServerInfo parses TeamSpeak serverinfo response
func parseServerInfo(line string) (*protocol.QueryResult, error) {
	line = strings.TrimSpace(line)

	// Parse key=value pairs separated by spaces
	data := make(map[string]string)
	pairs := strings.Fields(line)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := unescapeValue(parts[1])
			data[key] = value
		}
	}

	// Extract required fields
	serverName := data["virtualserver_name"]
	if serverName == "" {
		return nil, fmt.Errorf("missing server name in response")
	}

	// Parse client counts
	clientsOnline := parseInt(data["virtualserver_clientsonline"])
	maxClients := parseInt(data["virtualserver_maxclients"])

	// Subtract query clients (usually 1)
	queryClients := parseInt(data["virtualserver_queryclientsonline"])
	actualClients := clientsOnline - queryClients
	if actualClients < 0 {
		actualClients = clientsOnline
	}

	result := &protocol.QueryResult{
		Online:     true,
		Name:       serverName,
		NumPlayers: actualClients,
		MaxPlayers: maxClients,
		Version:    data["virtualserver_version"],
		Password:   data["virtualserver_flag_password"] == "1",
		Extra:      make(map[string]interface{}),
	}

	// Add extra information
	if platform := data["virtualserver_platform"]; platform != "" {
		result.Extra["platform"] = platform
	}
	if uptime := data["virtualserver_uptime"]; uptime != "" {
		result.Extra["uptime"] = uptime
	}
	if created := data["virtualserver_created"]; created != "" {
		result.Extra["created"] = created
	}
	if codec := data["virtualserver_codec_encryption_mode"]; codec != "" {
		result.Extra["codec_encryption"] = codec
	}
	if channels := data["virtualserver_channelsonline"]; channels != "" {
		result.Extra["channels_online"] = parseInt(channels)
	}

	return result, nil
}

// parseInt safely parses an integer from string
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}

// unescapeValue unescapes TeamSpeak ServerQuery escaped values
func unescapeValue(s string) string {
	s = strings.ReplaceAll(s, "\\s", " ")
	s = strings.ReplaceAll(s, "\\p", "|")
	s = strings.ReplaceAll(s, "\\/", "/")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// Name returns the protocol name
func (p *Protocol) Name() string {
	return fmt.Sprintf("%s (ServerQuery)", p.gameName)
}

// DefaultPort returns the default TeamSpeak ServerQuery port
func (p *Protocol) DefaultPort() int {
	return defaultPort
}

// SupportsSRV indicates that TeamSpeak does not use SRV records for ServerQuery
func (p *Protocol) SupportsSRV() bool {
	return false
}

// SRVService returns empty string (not used)
func (p *Protocol) SRVService() string {
	return ""
}
