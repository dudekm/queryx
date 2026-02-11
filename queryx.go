// Package queryx provides a universal library for querying game servers.
//
// QueryX supports multiple game server protocols including Minecraft, Counter-Strike,
// TeamSpeak, Discord, and more. It provides a clean, testable API with built-in
// DNS resolution, caching, and multiple transport protocols (UDP/TCP/HTTP).
//
// Example usage:
//
//	client := queryx.NewClient(
//	    queryx.WithTimeout(5 * time.Second),
//	    queryx.WithDebug(true),
//	)
//
//	result, err := client.Query(ctx, queryx.GameMinecraft, "mc.example.com", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Players: %d/%d\n", result.Players.Online, result.MaxPlayers)
package queryx

import (
	"context"
	"fmt"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/resolver"
	"github.com/dudekm/queryx/internal/transport"
)

// Client is the main QueryX client for querying game servers
type Client struct {
	resolver  resolver.Resolver
	factory   *protocol.Factory
	transport transport.Transport
	logger    Logger
	timeout   time.Duration
}

// Option is a functional option for configuring the Client
type Option func(*Client)

// WithTimeout sets the default timeout for queries
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithLogger sets the logger for the client
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithTransport sets a custom transport implementation
func WithTransport(t transport.Transport) Option {
	return func(c *Client) {
		c.transport = t
	}
}

// WithResolver sets a custom resolver implementation
func WithResolver(r resolver.Resolver) Option {
	return func(c *Client) {
		c.resolver = r
	}
}

// WithDebug enables debug logging using ConsoleLogger
func WithDebug(debug bool) Option {
	return func(c *Client) {
		if debug {
			c.logger = NewConsoleLogger()
		}
	}
}

// WithDNSCache enables DNS caching with the specified TTL in seconds
func WithDNSCache(ttlSeconds int) Option {
	return func(c *Client) {
		c.resolver = resolver.NewDefaultResolver(ttlSeconds)
	}
}

// NewClient creates a new QueryX client with the given options
func NewClient(opts ...Option) *Client {
	client := &Client{
		logger:  &NoOpLogger{},
		timeout: 5 * time.Second,
		factory: protocol.NewFactory(),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Set defaults if not provided
	if client.transport == nil {
		client.transport = transport.NewDefaultTransport(client.timeout)
	}

	if client.resolver == nil {
		client.resolver = resolver.NewDefaultResolver(0) // No cache by default
	}

	return client
}

// Query queries a game server and returns the result
func (c *Client) Query(ctx context.Context, gameType GameType, host string, port *int) (*QueryResult, error) {
	c.logger.Debug("Starting query", F("gameType", gameType), F("host", host))

	// Get protocol from factory
	proto, err := c.factory.Get(string(gameType))
	if err != nil {
		if err == protocol.ErrUnsupportedGame {
			return nil, NewQueryError(gameType, host, ErrUnsupportedGame)
		}
		return nil, NewQueryError(gameType, host, err)
	}

	// Determine port
	defaultPort := proto.DefaultPort()
	if port != nil {
		defaultPort = *port
	}

	// Resolve hostname
	srvService := ""
	if proto.SupportsSRV() {
		srvService = proto.SRVService()
	}

	addr, err := c.resolver.Resolve(ctx, host, defaultPort, srvService)
	if err != nil {
		c.logger.Error("DNS resolution failed", F("host", host), F("error", err))
		return nil, NewQueryError(gameType, host, fmt.Errorf("%w: %v", ErrDNSResolution, err))
	}

	c.logger.Debug("Resolved address", F("address", addr.String()))

	// Create context with timeout if not already set
	queryCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	// Query the server
	var protocolResult *protocol.QueryResult

	// Check if protocol supports QueryWithHostname (for SNI/virtual hosting)
	type hostnameAware interface {
		QueryWithHostname(ctx context.Context, addr string, hostname string) (*protocol.QueryResult, error)
	}

	if hostnameProto, ok := proto.(hostnameAware); ok {
		// Use original hostname for protocols that need it (like Minecraft)
		protocolResult, err = hostnameProto.QueryWithHostname(queryCtx, addr.String(), host)
	} else {
		// Use standard Query for protocols that don't need hostname
		protocolResult, err = proto.Query(queryCtx, addr.String())
	}

	if err != nil {
		c.logger.Error("Query failed", F("address", addr.String()), F("error", err))
		return nil, NewQueryError(gameType, host, err)
	}

	// Use domain model directly (no conversion needed - SOLID/DDD pattern)
	// Just set the Type field for the application layer
	protocolResult.Type = string(gameType)

	c.logger.Info("Query successful",
		F("address", addr.String()),
		F("ping", protocolResult.Ping),
		F("online", protocolResult.Online),
	)

	return protocolResult, nil
}

// QueryWithOptions queries a game server using QueryInput options
func (c *Client) QueryWithOptions(ctx context.Context, input QueryInput) (*QueryResult, error) {
	// Use timeout from input if provided
	if input.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, input.Timeout)
		defer cancel()
	}

	return c.Query(ctx, input.ServerType, input.Host, input.Port)
}

// QueryVerbose queries a game server and returns both the result and detailed diagnostics
// This is useful for debugging, monitoring, or displaying detailed DNS/query information
func (c *Client) QueryVerbose(ctx context.Context, gameType GameType, host string, port *int) (*VerboseQueryResult, error) {
	queryStart := time.Now()
	c.logger.Debug("Starting verbose query", F("gameType", gameType), F("host", host))

	// Get protocol from factory
	proto, err := c.factory.Get(string(gameType))
	if err != nil {
		if err == protocol.ErrUnsupportedGame {
			return nil, NewQueryError(gameType, host, ErrUnsupportedGame)
		}
		return nil, NewQueryError(gameType, host, err)
	}

	// Determine port
	defaultPort := proto.DefaultPort()
	if port != nil {
		defaultPort = *port
	}

	// Prepare diagnostics
	diagnostics := &QueryDiagnostics{
		Timestamp: queryStart,
		Input: QueryInput{
			ServerType: gameType,
			Host:       host,
			Port:       port,
			Timeout:    c.timeout,
		},
		QueryMetrics: QueryMetrics{
			Protocol: proto.Name(),
		},
	}

	// Resolve hostname with diagnostics
	dnsStart := time.Now()
	srvService := ""
	if proto.SupportsSRV() {
		srvService = proto.SRVService()
	}

	var addr *resolver.Address
	var resDiag *resolver.ResolutionDiagnostics

	// Check if resolver supports verbose mode
	if verboseResolver, ok := c.resolver.(resolver.VerboseResolver); ok {
		addr, resDiag, err = verboseResolver.ResolveWithDiagnostics(ctx, host, defaultPort, srvService)
	} else {
		// Fallback to regular resolve
		addr, err = c.resolver.Resolve(ctx, host, defaultPort, srvService)
		resDiag = &resolver.ResolutionDiagnostics{
			InputHostname:  host,
			ResolvedIP:     addr.IP,
			ResolvedPort:   addr.Port,
			SRVRecordFound: false,
			SRVRecords:     []resolver.SRVRecordInfo{},
			ARecords:       []string{addr.IP},
			AAAARecords:    []string{},
		}
	}

	dnsLatency := time.Since(dnsStart)
	diagnostics.QueryMetrics.DNSLatencyMs = int(dnsLatency.Milliseconds())

	if err != nil {
		c.logger.Error("DNS resolution failed", F("host", host), F("error", err))
		diagnostics.QueryMetrics.Success = false
		return &VerboseQueryResult{
			Result:      nil,
			Diagnostics: diagnostics,
		}, NewQueryError(gameType, host, fmt.Errorf("%w: %v", ErrDNSResolution, err))
	}

	// Map resolver diagnostics to public API
	diagnostics.Resolution = DNSResolution{
		InputHostname:  resDiag.InputHostname,
		ResolvedIP:     resDiag.ResolvedIP,
		ResolvedPort:   resDiag.ResolvedPort,
		SRVRecordFound: resDiag.SRVRecordFound,
		ARecords:       resDiag.ARecords,
		AAAARecords:    resDiag.AAAARecords,
	}

	// Convert SRV records
	for _, srvRec := range resDiag.SRVRecords {
		diagnostics.Resolution.SRVRecords = append(diagnostics.Resolution.SRVRecords, SRVRecord{
			Target:   srvRec.Target,
			Port:     srvRec.Port,
			Priority: srvRec.Priority,
			Weight:   srvRec.Weight,
		})
	}

	c.logger.Debug("Resolved address", F("address", addr.String()))

	// Create context with timeout if not already set
	queryCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	// Query the server
	queryOnlyStart := time.Now()
	var protocolResult *protocol.QueryResult

	// Check if protocol supports QueryWithHostname (for SNI/virtual hosting)
	type hostnameAware interface {
		QueryWithHostname(ctx context.Context, addr string, hostname string) (*protocol.QueryResult, error)
	}

	if hostnameProto, ok := proto.(hostnameAware); ok {
		// Use original hostname for protocols that need it (like Minecraft)
		protocolResult, err = hostnameProto.QueryWithHostname(queryCtx, addr.String(), host)
	} else {
		// Use standard Query for protocols that don't need hostname
		protocolResult, err = proto.Query(queryCtx, addr.String())
	}

	queryOnlyLatency := time.Since(queryOnlyStart)
	diagnostics.QueryMetrics.QueryLatencyMs = int(queryOnlyLatency.Milliseconds())

	if err != nil {
		c.logger.Error("Query failed", F("address", addr.String()), F("error", err))
		diagnostics.QueryMetrics.Success = false
		return &VerboseQueryResult{
			Result:      nil,
			Diagnostics: diagnostics,
		}, NewQueryError(gameType, host, err)
	}

	// Set query metrics
	diagnostics.QueryMetrics.Success = true
	diagnostics.QueryMetrics.LatencyMs = protocolResult.Ping

	// Try to extract protocol version from Raw data if available
	if rawMap, ok := protocolResult.Raw.(map[string]interface{}); ok {
		if version, ok := rawMap["version"].(map[string]interface{}); ok {
			if protocolVersion, ok := version["protocol"].(float64); ok {
				diagnostics.QueryMetrics.ProtocolVersion = int(protocolVersion)
			}
		}
	}

	// Use domain model directly (no conversion needed - SOLID/DDD pattern)
	protocolResult.Type = string(gameType)

	c.logger.Info("Verbose query successful",
		F("address", addr.String()),
		F("ping", protocolResult.Ping),
		F("online", protocolResult.Online),
		F("dns_latency_ms", diagnostics.QueryMetrics.DNSLatencyMs),
		F("query_latency_ms", diagnostics.QueryMetrics.QueryLatencyMs),
	)

	return &VerboseQueryResult{
		Result:      protocolResult,
		Diagnostics: diagnostics,
	}, nil
}

// QuickQuery is a convenience function for quick one-off queries
// It creates a new client with default settings and performs the query
func QuickQuery(gameType GameType, host string) (*QueryResult, error) {
	client := NewClient()
	return client.Query(context.Background(), gameType, host, nil)
}
