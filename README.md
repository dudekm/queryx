# QueryX

[![Test](https://github.com/dudekm/queryx/actions/workflows/test.yml/badge.svg)](https://github.com/dudekm/queryx/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dudekm/queryx)](https://goreportcard.com/report/github.com/dudekm/queryx)
[![codecov](https://codecov.io/gh/dudekm/queryx/branch/main/graph/badge.svg)](https://codecov.io/gh/dudekm/queryx)
[![Go Reference](https://pkg.go.dev/badge/github.com/dudekm/queryx.svg)](https://pkg.go.dev/github.com/dudekm/queryx)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Universal Go library and CLI for querying game servers. One protocol-agnostic
API — `client.Query(ctx, type, host, port)` — returns the same `QueryResult`
for 52 games across 8 protocols (Minecraft, Source Engine, GameSpy, CFX.re,
SA-MP, MTA, TeamSpeak, Hytale).

## 📚 Table of Contents

- [Supported Games](#-supported-games)
- [Installation](#-installation)
- [Quick Start (Library)](#-quick-start-library)
- [Configuration](#-configuration)
- [CLI Tool](#-cli-tool)
- [Development](#-development)
- [Docker](#-docker)
- [Testing](#-testing)
- [Project Structure](#-project-structure)
- [Adding a New Game](#-adding-a-new-game)
- [Examples](#-examples)
- [Debugging](#-debugging)
- [Roadmap](#-roadmap)
- [License](#-license)

## 🎮 Supported Games

QueryX supports **52 games** across **8 protocols**. Highlights:

- ✅ **Minecraft Java Edition**
- ✅ **Counter-Strike 2 / 1.6 / Source**
- ✅ **Rust** (Source Engine A2S, default port `28015`)
- ✅ **FiveM / RedM / Alt:V** (CFX.re HTTP)
- ✅ **ARMA 2/3, DayZ** (GameSpy)
- ✅ **SA-MP, Multi Theft Auto, TeamSpeak 3, Hytale**
- ✅ …and 40+ more Source Engine titles (ARK, Squad, Valheim, 7 Days to Die, etc.)

👉 **See [`GAMES.md`](GAMES.md) for the full table** — every game with its
`type` key, protocol, default port, and implementation status.

## 📦 Installation

```bash
go get github.com/dudekm/queryx
```

## 🚀 Quick Start (Library)

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dudekm/queryx"
)

func main() {
    // Create client
    client := queryx.NewClientWithDefaults()

    // Query server
    ctx := context.Background()
    result, err := client.Query(ctx, queryx.GameMinecraft, "hypixel.net", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Server: %s\n", result.Name)
    fmt.Printf("Players: %d/%d\n", result.NumPlayers, result.MaxPlayers)
    fmt.Printf("Version: %s\n", result.Version)
    fmt.Printf("Ping: %dms\n", result.Ping)
}
```

### With Custom Port

```go
port := 25566
result, err := client.Query(ctx, queryx.GameMinecraft, "example.com", &port)
```

### With Options

```go
client := queryx.NewClientWithDefaults(
    queryx.WithTimeout(10 * time.Second),
    queryx.WithDebug(true),
)

result, err := client.Query(ctx, queryx.GameMinecraft, "mc.example.com", nil)
```

### All Supported Games

```go
// Minecraft
result, err := client.Query(ctx, queryx.GameMinecraft, "hypixel.net", nil)

// Counter-Strike 2
result, err := client.Query(ctx, queryx.GameCS2, "cs2.example.com", nil)

// Counter-Strike 1.6
result, err := client.Query(ctx, queryx.GameCS16, "cs16.example.com", nil)

// Counter-Strike: Source
result, err := client.Query(ctx, queryx.GameCSSource, "css.example.com", nil)

// Rust (Source Engine A2S, default query port 28015)
result, err := client.Query(ctx, queryx.GameRust, "rust.example.com", nil)

// FiveM (CFX.re HTTP, default port 30120)
result, err := client.Query(ctx, queryx.GameFiveM, "fivem.example.com", nil)
```

### Error Handling

```go
result, err := client.Query(ctx, queryx.GameMinecraft, "example.com", nil)
if err != nil {
    // Check for specific errors
    if errors.Is(err, queryx.ErrTimeout) {
        fmt.Println("Query timed out")
    } else if errors.Is(err, queryx.ErrServerOffline) {
        fmt.Println("Server is offline")
    }
    return
}

// Check if server is online
if !result.Online {
    fmt.Println("Server is offline")
    return
}

fmt.Printf("Server: %s\n", result.Name)
```

### Verbose Mode (Diagnostic Information)

Get detailed diagnostic information including DNS resolution, SRV records, and timing metrics:

```go
// Query with verbose diagnostics
verboseResult, err := client.QueryVerbose(ctx, queryx.GameMinecraft, "sopelmc.pl", nil)
if err != nil {
    log.Fatal(err)
}

// Access standard query result
result := verboseResult.Result
fmt.Printf("Server: %s\n", result.Name)
fmt.Printf("Players: %d/%d\n", result.NumPlayers, result.MaxPlayers)

// Access diagnostic information
diag := verboseResult.Diagnostics
fmt.Printf("\nDNS Resolution:\n")
fmt.Printf("  Input Host: %s\n", diag.Resolution.InputHostname)
fmt.Printf("  Resolved IP: %s\n", diag.Resolution.ResolvedIP)
fmt.Printf("  Resolved Port: %d\n", diag.Resolution.ResolvedPort)
fmt.Printf("  SRV Record Found: %v\n", diag.Resolution.SRVRecordFound)

fmt.Printf("\nTiming:\n")
fmt.Printf("  DNS Latency: %dms\n", diag.QueryMetrics.DNSLatencyMs)
fmt.Printf("  Query Latency: %dms\n", diag.QueryMetrics.QueryLatencyMs)
fmt.Printf("  Network Ping: %dms\n", diag.QueryMetrics.LatencyMs)

// Access SRV records if found
if len(diag.Resolution.SRVRecords) > 0 {
    fmt.Printf("\nSRV Records:\n")
    for _, srv := range diag.Resolution.SRVRecords {
        fmt.Printf("  %s:%d (priority: %d, weight: %d)\n",
            srv.Target, srv.Port, srv.Priority, srv.Weight)
    }
}
```

**Diagnostic Information Includes:**
- Input hostname and resolved IP/port
- DNS resolution details (A, AAAA, SRV records)
- Query timing (DNS lookup, server query, network ping)
- Protocol information (name, version)
- Success status

## 🧩 Configuration

The client is configured with functional options passed to
`NewClientWithDefaults(...)` (or `NewClient(...)`). All are optional; sensible
defaults apply.

| Option | Default | Description |
|--------|---------|-------------|
| `WithTimeout(d time.Duration)` | `5s` | Per-query timeout (also honored via `context`). |
| `WithDebug(enabled bool)` | `false` | Enable verbose debug logging to stdout. |
| `WithLogger(l Logger)` | no-op logger | Plug in a custom `Logger` implementation. |
| `WithDNSCache(ttlSeconds int)` | disabled | Cache DNS lookups for the given TTL. |
| `WithResolver(r resolver.Resolver)` | system resolver | Inject a custom DNS resolver (e.g. for tests). |
| `WithTransport(t transport.Transport)` | UDP/TCP/HTTP | Inject a custom transport (e.g. a mock in tests). |

```go
client := queryx.NewClientWithDefaults(
    queryx.WithTimeout(10*time.Second),
    queryx.WithDNSCache(300),  // cache DNS for 5 minutes
    queryx.WithDebug(true),
)
```

Dependency injection (`WithTransport`, `WithResolver`) is what makes the library
fully testable without touching the network — see [Testing](#-testing).

For CLI configuration, see [CLI Flags](#cli-flags).

## 🖥️ CLI Tool

### Build

```bash
go build -o queryx ./cmd/queryx
```

### Usage

```bash
# Minecraft (default)
./queryx -type minecraft -host hypixel.net

# With custom port
./queryx -type minecraft -host example.com -port 25566

# Counter-Strike 2
./queryx -type cs2 -host cs2.example.com

# Counter-Strike 1.6
./queryx -type cs16 -host cs16.example.com

# Rust (default query port 28015)
./queryx -type rust -host rust.example.com

# FiveM (CFX.re HTTP, default port 30120)
./queryx -type fivem -host fivem.example.com

# JSON output
./queryx -host hypixel.net -json

# Debug mode (logs to console)
./queryx -host hypixel.net -debug

# Verbose mode (detailed diagnostics)
./queryx -host sopelmc.pl -verbose

# Verbose with JSON
./queryx -host sopelmc.pl -verbose -json

# Custom timeout
./queryx -host hypixel.net -timeout 10s
```

### CLI Flags

```
-host string       Server hostname (required)
-port int          Server port (default: game-specific)
-type string       Game type (default "minecraft")
                   Options: minecraft, cs2, cs16, cssource, rust, fivem,
                   redm, altv, arma3, dayz, samp, mta, teamspeak, and 40+
                   more Source Engine titles (see types.go)
-timeout duration  Query timeout (default 5s)
-debug            Enable debug logging (console output)
-verbose          Show detailed diagnostic information (DNS, SRV, timing)
-json             Output as JSON
-version          Show version
```

## 🔨 Development

Prefer a fully containerized workflow? Skip straight to [Docker](#-docker) — no
local Go toolchain required. Otherwise, with Go installed locally:

```bash
# Install dependencies
go mod tidy

# Build everything
go build ./...

# Build CLI tool
go build -o queryx ./cmd/queryx

# Build for Windows
go build -o queryx.exe ./cmd/queryx

# Build for Linux
GOOS=linux go build -o queryx ./cmd/queryx
```

## 🐳 Docker

You can develop, test, lint and run QueryX entirely in Docker — no local Go
toolchain required. A multi-stage `Dockerfile`, a `compose.yaml` with
ready-made service definitions, and a `Makefile` with `docker-*` shortcuts are
included.

### Using the Makefile (recommended)

```bash
make docker-test        # run the full test suite in a container
make docker-test-short  # run unit tests only (fast)
make docker-lint        # run golangci-lint
make docker-dev         # open an interactive dev shell
make docker-build       # build the runtime CLI image (queryx:local)

# Run the CLI, passing flags via ARGS
make docker-run ARGS="-type rust -host rust.example.com"
make docker-run ARGS="-type fivem -host fivem.example.com -json"
```

Run `make help` to list every available target (local and Docker).

### Using docker compose directly

```bash
# Development & CI workflows (source mounted, module/build caches persisted)
docker compose run --rm test          # full test suite
docker compose run --rm test-short    # unit tests only
docker compose run --rm coverage      # write coverage.out
docker compose run --rm lint          # golangci-lint
docker compose run --rm dev           # interactive shell

# Build and run the CLI image
docker compose build queryx
docker compose run --rm queryx -type minecraft -host hypixel.net
```

### Building and running the image by hand

```bash
# Build (override the Go version if needed with --build-arg GO_VERSION=1.27)
docker build -t queryx:local .

# The image's entrypoint is the queryx binary
docker run --rm queryx:local -version
docker run --rm queryx:local -type cs2 -host cs2.example.com
```

The runtime image is a minimal Alpine layer containing only the statically
linked binary and CA certificates, and runs as a non-root user.

## 🧪 Testing

QueryX has two types of tests:
- **Unit tests** - Test individual components with mocks (fast, isolated)
- **Integration tests** - Test entire flow from input to output (realistic, contract-focused)

### Run All Tests

```bash
# Run all tests (unit + integration)
go test ./...

# With verbose output
go test ./... -v

# With coverage report
go test ./... -cover

# With race detector (Windows requires CGO)
CGO_ENABLED=1 go test -race ./...
```

### Run Specific Test Types

```bash
# Only unit tests (exclude integration tests)
go test ./... -v -short

# Only integration tests
go test -v -run TestIntegration

# Specific integration test
go test -v -run TestIntegration_Minecraft_FullFlow
go test -v -run TestIntegration_CounterStrike_FullFlow
```

### Test Specific Package

```bash
# Test main package (includes integration tests)
go test . -v

# Test Minecraft protocol
go test ./internal/protocol/minecraft -v

# Test Source Engine protocol (CS2, CS 1.6, CS:Source)
go test ./internal/protocol/source -v

# Test transport layer
go test ./internal/transport -v

# Test DNS resolver
go test ./internal/resolver -v
```

### Code Coverage

#### Generate Coverage Report

```bash
# Generate coverage profile
go test ./... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# Open interactive HTML coverage report
go tool cover -html=coverage.out
```

#### Coverage by Package

```bash
# Coverage for each package
go test ./... -cover

# Detailed coverage per package
go test ./internal/protocol/minecraft -coverprofile=minecraft.out
go tool cover -html=minecraft.out
```

#### Current Coverage

- **Main package**: 97.8%
- **Protocol interface**: 100%
- **Minecraft protocol**: 76.5%
- **Source Engine**: 79.6%
- **Resolver**: 50%

### Integration Tests vs Unit Tests

**Unit Tests** (existing):
```go
// Mock everything - test components in isolation
mockTransport := transport.NewMockTransport()
mockResolver := resolver.NewMockResolver()
mockProto := &mockProtocol{...}
```

**Integration Tests** (`integration_test.go`):
```go
// Mock ONLY network transport - test entire flow
mockTransport := transport.NewMockTransport()
mockTransport.UDPResponses[addr] = realServerResponse

client := NewClientWithDefaults(WithTransport(mockTransport))
result, err := client.Query(ctx, ServerCS2, "server.com", nil)

// Test input → output contract
// Refactor internals freely - test still passes!
```

### Running Tests During Development

```bash
# Quick check (fast)
go test ./... -short

# Full test suite
go test ./...

# Watch mode (requires external tool like gotestsum)
gotestsum --watch

# Test with coverage and open report
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

### Test Philosophy

1. **Unit Tests** - Test each component works correctly in isolation
2. **Integration Tests** - Test the public API contract (input → output)
3. **No Real Network Calls** - All network responses are mocked for reliability

This allows you to:
- ✅ Refactor `internal/*` packages safely
- ✅ Change protocol parsing logic without breaking tests
- ✅ Ensure public API remains stable
- ✅ Fast, deterministic test runs

## 📁 Project Structure

```
queryx/
├── queryx.go                  # Public API
├── types.go                   # Core types (QueryResult, GameType)
├── errors.go                  # Error definitions
├── logger.go                  # Logger interface
├── register.go                # Protocol registration
├── cmd/
│   └── queryx/               # CLI tool
├── internal/
│   ├── transport/            # Network layer (UDP/TCP/HTTP)
│   ├── resolver/             # DNS resolution with SRV support
│   └── protocol/             # Protocol implementations
│       ├── minecraft/        # Minecraft protocol
│       ├── source/           # Source Engine (CS2, CS 1.6, CS:S, Rust, ...)
│       ├── cfxre/            # CFX.re HTTP (FiveM, RedM, Alt:V)
│       ├── gamespy/          # GameSpy (ARMA 2/3, DayZ)
│       ├── samp/             # SA-MP
│       ├── mta/              # Multi Theft Auto (ASE)
│       └── teamspeak/        # TeamSpeak 3
├── Dockerfile                # Multi-stage build for the CLI image
├── compose.yaml        # Dev/test/lint/run services
├── Makefile                  # Local + Docker task shortcuts
└── examples/
    ├── library/              # Library usage
    └── protocol_comparison/  # Protocol comparison demo
```

## 🔧 Adding a New Game

> After implementing a game, remember to add a row to [`GAMES.md`](GAMES.md).

1. **Create protocol package**
```bash
mkdir internal/protocol/yourgame
```

2. **Implement Protocol interface** (`internal/protocol/yourgame/yourgame.go`)
```go
package yourgame

import (
    "context"
    "github.com/dudekm/queryx/internal/protocol"
    "github.com/dudekm/queryx/internal/transport"
)

type Protocol struct {
    transport transport.Transport
}

func NewProtocol(t transport.Transport) *Protocol {
    return &Protocol{transport: t}
}

func (p *Protocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
    // Implement: build request, send, parse response
    return &protocol.QueryResult{
        Online: true,
        Name:   "Server Name",
        // ... fill other fields
    }, nil
}

func (p *Protocol) Name() string        { return "Your Game" }
func (p *Protocol) DefaultPort() int    { return 12345 }
func (p *Protocol) SupportsSRV() bool   { return false }
func (p *Protocol) SRVService() string  { return "" }
```

3. **Register protocol** (in `register.go`)
```go
import "github.com/dudekm/queryx/internal/protocol/yourgame"

func (c *Client) RegisterDefaultProtocols() {
    // ... existing protocols ...

    yourGameProto := yourgame.NewProtocol(c.transport)
    c.factory.Register(string(GameYourGame), yourGameProto)
}
```

4. **Add GameType constant** (in `types.go`)
```go
const (
    GameYourGame GameType = "yourgame"
)
```

5. **Write tests** (`internal/protocol/yourgame/yourgame_test.go`)
```go
func TestProtocol_Query(t *testing.T) {
    mockTransport := transport.NewMockTransport()
    mockTransport.UDPResponses["127.0.0.1:12345"] = mockData

    p := NewProtocol(mockTransport)
    result, err := p.Query(context.Background(), "127.0.0.1:12345")

    assert.NoError(t, err)
    assert.True(t, result.Online)
}
```

6. **Add an integration (E2E) test** in `integration_test.go` asserting the
   public `QueryResult` contract, and **add a row to [`GAMES.md`](GAMES.md)**.

## 📝 Examples

See `examples/` directory:
- `examples/library/` - Basic library usage
- `examples/protocol_comparison/` - Compare different protocols

Run examples:
```bash
cd examples/library
go run main.go

cd ../protocol_comparison
go run main.go
```

## 🐛 Debugging

Enable debug logging:

```go
// In code
client := queryx.NewClientWithDefaults(queryx.WithDebug(true))
```

```bash
# In CLI
./queryx -host hypixel.net -debug
```

Inspect raw response:
```go
fmt.Printf("Raw response: %x\n", result.Raw)
```

## 📖 Documentation

- **Minecraft Protocol**: https://wiki.vg/Protocol
- **Source Engine Protocol**: https://developer.valvesoftware.com/wiki/Server_queries

## 🗺️ Roadmap

- [x] Phase 1: Foundation (types, transport, resolver, protocol interface)
- [x] Phase 2: Minecraft Java Edition
- [x] Phase 3: Counter-Strike (CS2, CS 1.6, CS:Source)
- [x] Phase 4: More protocols (Rust, TeamSpeak, FiveM, SA-MP, GameSpy, MTA)
- [ ] Phase 5: Advanced features (A2S_PLAYER, A2S_RULES, connection pooling)

## 📄 License

MIT

---

**Made with Go 1.27+**
