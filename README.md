# QueryX

Universal Go library for querying game servers (Minecraft, Counter-Strike, etc.).

## 🎮 Supported Games

- ✅ **Minecraft Java Edition**
- ✅ **Counter-Strike 2**
- ✅ **Counter-Strike 1.6**
- ✅ **Counter-Strike: Source**

## 📦 Installation

```bash
go get github.com/dudekm/queryx
```

## 🚀 Quick Start - Library

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
    fmt.Printf("Players: %d/%d\n", result.Players.Online, result.MaxPlayers)
    fmt.Printf("Version: %s\n", result.Version)
    fmt.Printf("Ping: %v\n", result.Ping)
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

# JSON output
./queryx -host hypixel.net -json

# Debug mode
./queryx -host hypixel.net -debug

# Custom timeout
./queryx -host hypixel.net -timeout 10s
```

### CLI Flags

```
-host string       Server hostname (required)
-port int          Server port (default: game-specific)
-type string       Game type (default "minecraft")
                   Options: minecraft, cs2, cs16, cssource
-timeout duration  Query timeout (default 5s)
-debug            Enable debug logging
-json             Output as JSON
-version          Show version
```

## 🔨 Building After Changes

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
│       └── source/           # Source Engine (CS2, CS 1.6, CS:S)
└── examples/
    ├── library/              # Library usage
    └── protocol_comparison/  # Protocol comparison demo
```

## 🔧 Adding New Protocol

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
- [ ] Phase 4: More protocols (Rust, TeamSpeak, FiveM, SA-MP)
- [ ] Phase 5: Advanced features (A2S_PLAYER, A2S_RULES, connection pooling)

## 📄 License

MIT

---

**Made with Go 1.23+**
