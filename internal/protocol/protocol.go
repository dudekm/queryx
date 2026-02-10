package protocol

import "context"

// Protocol defines the interface that all game server protocols must implement (Domain Service)
// This is the core contract for querying game servers
type Protocol interface {
	// Query sends a query to the server and returns the parsed result
	Query(ctx context.Context, addr string) (*QueryResult, error)

	// Name returns the protocol name
	Name() string

	// DefaultPort returns the default port for this protocol
	DefaultPort() int

	// SupportsSRV indicates if this protocol supports SRV record lookup
	SupportsSRV() bool

	// SRVService returns the SRV service name (e.g., "minecraft")
	SRVService() string
}
