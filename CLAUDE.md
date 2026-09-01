# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

QueryX is a universal Go library for querying game servers. It provides a clean, protocol-agnostic API for querying servers across 52 games using 8 different protocols (Minecraft, Source Engine, GameSpy, CFX.re, SA-MP, MTA, TeamSpeak, Hytale).

The library is designed with testability, extensibility, and maintainability in mind.

The full, authoritative list of supported games (with `type` keys, protocols,
default ports and implementation status) lives in [`GAMES.md`](GAMES.md). Update
that table whenever a game is added or changes status.

## Build & Test Commands

### Building

```bash
# Build everything (library + CLI)
go build ./...

# Build CLI tool
go build -o queryx ./cmd/queryx

# Build for different platforms
GOOS=linux go build -o queryx ./cmd/queryx
GOOS=windows go build -o queryx.exe ./cmd/queryx
```

### Testing

```bash
# Run all tests (unit + integration)
go test ./...

# Run only unit tests (fast, excludes integration tests)
go test ./... -short

# Run only integration tests
go test -v -run TestIntegration

# Test specific package
go test ./internal/protocol/minecraft -v
go test ./internal/protocol/source -v
go test ./internal/transport -v
go test ./internal/resolver -v

# Test with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run single test
go test -v -run TestProtocol_Query_Success
go test -v -run TestIntegration_Minecraft_FullFlow
```

### Docker (local dev/test without a host Go toolchain)

A multi-stage `Dockerfile`, a `docker-compose.yml`, and a `Makefile` provide a
containerized workflow. The compose services mount the source and cache the Go
module/build caches in named volumes.

```bash
# Via Makefile (see `make help` for all targets)
make docker-test         # full test suite in a container
make docker-test-short   # unit tests only
make docker-lint         # golangci-lint
make docker-dev          # interactive shell
make docker-build        # build runtime CLI image (queryx:local)
make docker-run ARGS="-type rust -host rust.example.com"

# Via docker compose directly
docker compose run --rm test
docker compose run --rm test-short
docker compose run --rm lint
docker compose build queryx
docker compose run --rm queryx -type fivem -host fivem.example.com

# Build/run the image by hand (override Go version if needed)
docker build --build-arg GO_VERSION=1.27 -t queryx:local .
docker run --rm queryx:local -version
```

The runtime image is a minimal Alpine layer with only the static binary and CA
certificates, running as a non-root user.

## Architecture

### Core Design Principles

1. **Protocol-agnostic API**: Users interact through `Client.Query()` regardless of game type
2. **Protocol factory pattern**: Protocols are registered in a factory and retrieved by game type string
3. **Dependency injection**: All components (transport, resolver, protocols) are injected for testability
4. **Layered architecture**: Clear separation between public API, protocols, transport, and DNS resolution

### Design Principles & Best Practices

This codebase follows **SOLID**, **KISS**, and **DRY** principles:

#### SOLID Principles
- **Single Responsibility**: Each protocol handles only its own game protocol logic
- **Open/Closed**: Easy to add new protocols without modifying existing code (just implement `Protocol` interface)
- **Liskov Substitution**: All protocols are interchangeable via the `Protocol` interface
- **Interface Segregation**: Small, focused interfaces (`Transport`, `Resolver`, `Protocol`)
- **Dependency Inversion**: High-level Client depends on abstractions (interfaces), not concrete implementations

#### KISS (Keep It Simple, Stupid)
- Straightforward factory pattern for protocol registration
- Simple interface contracts with minimal methods
- Clear separation of concerns between layers

#### DRY (Don't Repeat Yourself)
- Shared `Protocol` interface eliminates protocol-specific client code
- Common transport layer reused across all protocols
- DNS resolver logic centralized in one place

### Unified QueryResult Contract

**CRITICAL: Consistent Input/Output Across All Protocols**

Regardless of the game or protocol (Minecraft, CS2, ARMA, TeamSpeak), the `QueryResult` structure MUST remain consistent:

```go
type QueryResult struct {
    Online     bool        // Server online status
    Name       string      // Server name/hostname
    Map        string      // Current map name
    NumPlayers int         // Current player count
    MaxPlayers int         // Maximum player slots
    Players    []Player    // Detailed player list (if available)
    Bots       int         // Number of bots
    Type       string      // Server type (minecraft, cs2, etc.)
    Version    string      // Server version
    Ping       int         // Network latency in milliseconds (e.g., 587)
    Password   bool        // Password required flag
    Raw        interface{} // ALL protocol-specific data goes HERE
}
```

#### Standard Fields (Use These)

When implementing a protocol, **always map to these standard fields**:
- `Online` - Is the server responding?
- `Name` - Server name/title
- `Map` - Current map name (if applicable)
- `NumPlayers` - How many players are online
- `MaxPlayers` - Maximum player capacity
- `Players` - Detailed player list (name, score, duration)
- `Bots` - Number of bots (if protocol distinguishes them)
- `Version` - Server version string
- `Ping` - Network round-trip time in milliseconds (int, e.g., 587)
- `Password` - Does server require password?

#### Protocol-Specific Data Goes to `Raw`

**All protocol-specific fields that don't fit the standard schema MUST go into the `Raw` field.**

Examples:
- Minecraft: Full JSON response with favicon, sample players → `Raw`
- Source Engine: Game tags, server ID, VAC status, app ID → `Raw`
- TeamSpeak: Channel info, codec details, virtual server info → `Raw`
- FiveM: Resources, game build, server endpoints → `Raw`
- GameSpy: All key-value pairs that aren't standard fields → `Raw`

**Why this matters:**
1. **Consistent API**: Users can query any game and expect the same fields
2. **Easy switching**: Change from CS2 to Rust without code changes
3. **Predictable**: Same input format (gameType, host, port) → same output format
4. **Extensible**: Protocol-specific data still accessible via `Raw` for power users
5. **Simple**: Only ONE place for protocol-specific data - `Raw` (not `Extra`, not custom fields)

**When adding a new protocol:**
- ✅ DO: Map server response to standard `QueryResult` fields
- ✅ DO: Put ALL protocol-specific data in `Raw` field (as parsed JSON, struct, or map)
- ✅ DO: Use zero values for missing standard fields (empty string, 0, false, nil slice)
- ❌ DON'T: Add new custom fields to `QueryResult` structure
- ❌ DON'T: Use `Extra` field (deprecated - use `Raw` instead)
- ❌ DON'T: Skip standard fields just because protocol doesn't provide them

**What goes in `Raw`:**
- Full server response (JSON, parsed struct, map[string]interface{})
- Protocol-specific metadata (game tags, mod lists, resource lists)
- Extended player info (avatars, teams, clans, K/D ratios)
- Server rules, settings, custom data
- Anything that's unique to this game/protocol

### Architecture Layers

```
┌─────────────────────────────────────────┐
│  Public API (queryx.go)                 │
│  - Client, Query(), QueryWithOptions()  │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Protocol Factory (protocol/factory.go) │
│  - Register/Get protocols by game type  │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Protocol Implementations (internal/)   │
│  - minecraft/, source/, gamespy/        │
│  - Each implements Protocol interface   │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  DNS Resolver (resolver/)               │
│  - SRV record support (Minecraft)       │
│  - DNS caching with TTL                 │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Transport Layer (transport/)           │
│  - SendUDP(), SendTCP(), SendHTTP()     │
│  - Mock transport for testing           │
└─────────────────────────────────────────┘
```

### Key Components

#### Protocol Interface (`internal/protocol/protocol.go`)

All game protocols implement this interface:

```go
type Protocol interface {
    Query(ctx context.Context, addr string) (*QueryResult, error)
    Name() string
    DefaultPort() int
    SupportsSRV() bool
    SRVService() string
}
```

Some protocols support `QueryWithHostname()` for SNI/virtual hosting (e.g., Minecraft).

#### Protocol Factory (`internal/protocol/factory.go`)

Thread-safe registry for protocol implementations. Protocols are registered at client initialization in `register.go`.

#### Transport Layer (`internal/transport/`)

Abstraction for network communication:
- `transport.go`: Interface definition
- `default.go`: Real UDP/TCP/HTTP implementation
- `mock.go`: Mock implementation for testing

The mock transport uses maps to store pre-defined responses for different addresses, enabling deterministic testing.

#### DNS Resolver (`internal/resolver/`)

Handles DNS resolution with:
- SRV record support (for Minecraft's `_minecraft._tcp.domain`)
- Optional caching with configurable TTL
- Mock resolver for testing

### Protocol Implementations

Each protocol lives in `internal/protocol/<name>/`:

- **Minecraft** (`minecraft/`): Server List Ping protocol over TCP
  - Uses VarInt encoding for packet framing
  - Supports SNI via QueryWithHostname()
  - Parses JSON response format

- **Source Engine** (`source/`): A2S protocol over UDP
  - Used by CS 1.6, CS:S, CS2, TF2, Rust, ARK, etc. (40+ games)
  - Handles both GoldSrc (0x6d) and Source Engine (0x49) response formats
  - Supports challenge-response flow
  - Port detection for GoldSrc servers (CS 1.6, HL1)

- **GameSpy** (`gamespy/`): GameSpy Query Protocol over UDP
  - Used by ARMA 2/3, DayZ
  - Key-value pair response format

- **CFX.re** (`cfxre/`): HTTP-based protocol
  - Used by FiveM, RedM, Alt:V
  - JSON responses over HTTP

- **SA-MP** (`samp/`): San Andreas Multiplayer protocol
- **Multi Theft Auto** (`mta/`): ASE protocol
- **TeamSpeak** (`teamspeak/`): TeamSpeak 3 ServerQuery

### Client Initialization Flow

1. `NewClient()` creates client with default timeout (5s)
2. Apply functional options (WithTimeout, WithDebug, WithTransport, etc.)
3. `RegisterDefaultProtocols()` registers all 60+ game protocols in the factory
4. Client is ready to query any supported game

### Query Execution Flow

1. User calls `Client.Query(ctx, GameType, host, port)` - **Same input format for all games**
2. Client retrieves protocol from factory
3. DNS resolver resolves hostname (with SRV support if protocol enables it)
4. Client calls protocol's `Query()` or `QueryWithHostname()` method
5. Protocol builds game-specific request packet
6. Protocol sends request via transport layer (UDP/TCP/HTTP)
7. Protocol parses binary/JSON response into `protocol.QueryResult` - **Maps to standard fields**
8. Client converts `protocol.QueryResult` → `queryx.QueryResult` (public API type)
9. Result returned to user - **Same output format for all games**

**Key principle**: Different protocols, different wire formats, but **unified input/output** at the API level.

## Testing Strategy

Two complementary layers, both network-free (the transport is always mocked, so
tests are fast and deterministic — no real servers are contacted):

### 1. Unit Tests — components in isolation

- Test individual components with all dependencies mocked (transport, resolver, protocols).
- Fast, focused tests for each component.
- Located in each package: `minecraft_test.go`, `source_test.go`, `cfxre_test.go`, etc.
- Cover parsing edge cases, error paths, ports, and value objects.

### 2. Integration / End-to-End Tests — the public contract

These are **black-box E2E tests of the public API**: they pin the
**input → output** contract so that *whatever changes inside* `internal/*`,
the same input yields the same `QueryResult`.

- **Input** is always the same shape: `client.Query(ctx, <ServerType>, host, port)`.
- **Output** is always the same shape: a `QueryResult` with the standard fields.
- Only the **network transport** is mocked, fed with realistic raw server bytes/JSON.
- Everything else (resolver, factory, protocol, parser, API mapping) is the **real** code.
- Located in `integration_test.go` at the project root.

**Why this matters:** you can refactor or rewrite any protocol, the transport,
the resolver, or the factory, and as long as the public contract holds, these
tests stay green. `TestIntegration_APIContract` additionally asserts the exact
field types of `QueryResult` so the universal API can never silently drift.

Every new game MUST ship with both: unit tests for its protocol package and an
`integration_test.go` E2E case (see `TestIntegration_Rust_FullFlow` and
`TestIntegration_FiveM_FullFlow` for templates).

### Writing Tests

**For protocol implementations:**
```go
// Use mock transport with real server response bytes
mockTransport := transport.NewMockTransport()
mockTransport.UDPResponses["127.0.0.1:27015"] = realServerResponseBytes

proto := source.NewProtocol(mockTransport, "CS2")
result, err := proto.Query(ctx, "127.0.0.1:27015")
```

**For integration tests:**
```go
// Test entire flow with mocked transport
client := NewClientWithDefaults(WithTransport(mockTransport))
result, err := client.Query(ctx, GameCS2, "server.com", nil)
// Assert on public QueryResult fields
```

### Test Helpers

- `buildMinecraftResponse()`: Builds valid Minecraft packet bytes
- `buildSourceEngineResponse()`: Builds valid Source Engine A2S_INFO response
- Mock transport automatically returns responses for specific addresses

## Adding a New Game Protocol

Follow this pattern (documented in README, but key points):

1. Create `internal/protocol/<game>/<game>.go`
2. Implement `Protocol` interface
3. Register in `register.go` → `RegisterDefaultProtocols()`
4. Add `Server<Game>` constant in `types.go`
5. Write unit tests with mock transport
6. Add integration test in `integration_test.go`

**Important**: Study existing implementations (minecraft, source) for patterns on:
- Binary packet parsing with `bytes.Buffer` and `encoding/binary`
- Error handling and validation
- Ping measurement timing
- Response format parsing

**CRITICAL - Mapping to QueryResult**:
- **Always map to standard fields**: `Online`, `Name`, `Map`, `NumPlayers`, `MaxPlayers`, `Version`, `Password`, `Players`, `Bots`, `Ping`
- **Put protocol-specific data ONLY in `Raw`**: Everything that doesn't fit standard fields goes to `Raw` - no exceptions
- **Don't use `Extra` field**: It exists for backward compatibility but should NOT be used in new protocols
- **Use zero values for missing fields**: If protocol doesn't provide map name, set `Map: ""` (empty string, not nil)
- **Measure ping properly**: Record network latency in milliseconds as int (convert from time.Duration using `int(duration.Round(time.Millisecond).Milliseconds())`)

Example mapping:
```go
// Parse protocol response
serverName := parseServerName(response)
playerCount := parsePlayerCount(response)

// Measure ping and convert to milliseconds
pingStart := time.Now()
response, err := transport.SendUDP(ctx, addr, request)
pingDuration := time.Since(pingStart)
pingMs := int(pingDuration.Round(time.Millisecond).Milliseconds())

// Parse ALL protocol-specific data into Raw
// This can be: full JSON, parsed struct, map[string]interface{}, etc.
rawData := map[string]interface{}{
    "fullResponse": parseFullJSON(response),
    "gameTags":     parseGameTags(response),
    "modList":      parseModList(response),
    "serverRules":  parseServerRules(response),
    // ... any other protocol-specific fields
}

// Map to standard QueryResult
return &protocol.QueryResult{
    Online:     true,
    Name:       serverName,              // standard field
    Map:        parseMapName(response),  // standard field (or "" if not available)
    NumPlayers: playerCount,             // standard field
    MaxPlayers: parseMaxPlayers(response), // standard field
    Players:    parsePlayers(response),  // standard field (or nil if not available)
    Bots:       parseBots(response),     // standard field (or 0 if not available)
    Version:    parseVersion(response),  // standard field (or "" if not available)
    Ping:       pingMs,                  // network latency in milliseconds (int)
    Password:   parsePasswordFlag(response), // standard field (or false if unknown)
    Raw:        rawData,                 // ALL protocol-specific data → ONLY Raw
}, nil
```

## Code Organization

### Public API (`*.go` at root)
- `queryx.go`: Main client and Query() method
- `types.go`: Public types (QueryResult, GameType constants)
- `errors.go`: Error types and helpers
- `logger.go`: Logger interface and implementations
- `register.go`: Protocol registration

### Internal Packages
- `internal/protocol/`: Protocol interface + implementations
- `internal/transport/`: Network transport abstraction
- `internal/resolver/`: DNS resolution with caching

### CLI Tool
- `cmd/queryx/main.go`: Command-line interface

### Tests
- `*_test.go`: Unit tests (per package)
- `integration_test.go`: Full flow integration tests

## Common Tasks

### Running the CLI locally
```bash
go run ./cmd/queryx -type minecraft -host hypixel.net
go run ./cmd/queryx -type cs2 -host server.com -port 27015 -debug
```

### Debugging protocol issues
```bash
# Enable debug logging
go run ./cmd/queryx -host server.com -debug

# Or in code
client := queryx.NewClientWithDefaults(queryx.WithDebug(true))
```

### Testing protocol changes
```bash
# Test specific protocol
go test ./internal/protocol/source -v -run TestProtocol_Query

# Test integration
go test -v -run TestIntegration_CounterStrike
```

## Important Implementation Notes

### Universal API Contract (CRITICAL)

**Every protocol implementation MUST:**
1. Return the same `QueryResult` structure with standard fields
2. Map protocol responses to standard fields: `Online`, `Name`, `Map`, `NumPlayers`, `MaxPlayers`, `Version`, `Password`, `Ping`, `Connect`, `Players`, `Bots`
3. Put ALL protocol-specific data in the `Raw` field (not `Extra`, not custom fields - ONLY `Raw`)
4. Never add custom fields to `QueryResult` - this breaks the universal API contract
5. Use zero values for missing standard fields (empty string, 0, false) - don't skip them

**Example: Different protocols, same output structure**
```go
// Minecraft query
result, _ := client.Query(ctx, GameMinecraft, "hypixel.net", nil)
fmt.Println(result.Name, result.NumPlayers, result.MaxPlayers) // Works

// CS2 query - SAME fields available
result, _ := client.Query(ctx, GameCS2, "server.com", nil)
fmt.Println(result.Name, result.NumPlayers, result.MaxPlayers) // Works identically

// ARMA 3 query - SAME fields available
result, _ := client.Query(ctx, GameARMA3, "arma.com", nil)
fmt.Println(result.Name, result.NumPlayers, result.MaxPlayers) // Works identically
```

This consistency is the **core value proposition** of QueryX. Protect it.

**Note on `Extra` vs `Raw` fields:**
- `Extra` field exists in the struct for backward compatibility but should NOT be used
- Use ONLY `Raw` for protocol-specific data
- Having two fields for the same purpose (Extra and Raw) is confusing and redundant
- Future versions may deprecate `Extra` completely

### Source Engine Protocol Specifics

- **GoldSrc vs Source**: Response packet type differs (0x6d vs 0x49)
- **Challenge-response**: Some servers require challenge number in request
- **Port detection**: GoldSrc servers need port adjustment (queryPort vs gamePort)
- These details are handled in `internal/protocol/source/source.go`

### Minecraft Protocol Specifics

- **VarInt encoding**: Packet lengths and fields use VarInt encoding
- **SNI support**: QueryWithHostname() sends original hostname for virtual hosting
- **SRV records**: Supports `_minecraft._tcp.domain` SRV lookup
- TCP-based with packet framing

### Error Handling

- DNS errors: Wrapped with `ErrDNSResolution`
- Unsupported games: `ErrUnsupportedGame`
- Network errors: Propagated from transport layer
- All errors include context via `QueryError` type

### Performance Considerations

- DNS caching reduces latency for repeated queries (opt-in via `WithDNSCache()`)
- Context cancellation supported throughout the stack
- Default 5-second timeout (configurable via `WithTimeout()`)

## Version & Dependencies

- Go 1.27+ required
- Only external dependency: `github.com/stretchr/testify` (for testing)
- Standard library used for networking (`net`, `context`, `encoding/binary`)
