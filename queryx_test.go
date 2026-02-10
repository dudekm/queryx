package queryx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/protocol"
	"github.com/dudekm/queryx/internal/resolver"
	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
)

// mockProtocol is a mock protocol for testing
type mockProtocol struct {
	queryResult *protocol.QueryResult
	queryError  error
}

func (m *mockProtocol) Query(ctx context.Context, addr string) (*protocol.QueryResult, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.queryResult, nil
}

func (m *mockProtocol) Name() string       { return "mock" }
func (m *mockProtocol) DefaultPort() int   { return 25565 }
func (m *mockProtocol) SupportsSRV() bool  { return false }
func (m *mockProtocol) SRVService() string { return "" }

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient()

	assert.NotNil(t, client)
	assert.NotNil(t, client.logger)
	assert.NotNil(t, client.transport)
	assert.NotNil(t, client.resolver)
	assert.NotNil(t, client.factory)
	assert.Equal(t, 5*time.Second, client.timeout)
}

func TestNewClient_WithOptions(t *testing.T) {
	logger := NewConsoleLogger()
	mockTransport := transport.NewMockTransport()
	mockResolver := resolver.NewMockResolver()

	client := NewClient(
		WithTimeout(10*time.Second),
		WithLogger(logger),
		WithTransport(mockTransport),
		WithResolver(mockResolver),
	)

	assert.Equal(t, 10*time.Second, client.timeout)
	assert.Equal(t, logger, client.logger)
	assert.Equal(t, mockTransport, client.transport)
	assert.Equal(t, mockResolver, client.resolver)
}

func TestWithDebug(t *testing.T) {
	client := NewClient(WithDebug(true))
	assert.IsType(t, &ConsoleLogger{}, client.logger)

	client2 := NewClient(WithDebug(false))
	assert.IsType(t, &NoOpLogger{}, client2.logger)
}

func TestWithDNSCache(t *testing.T) {
	client := NewClient(WithDNSCache(60))
	assert.NotNil(t, client.resolver)
}

func TestClient_Query_Success(t *testing.T) {
	// Setup mocks
	mockTransport := transport.NewMockTransport()
	mockResolver := resolver.NewMockResolver()
	mockResolver.Addresses["example.com"] = &resolver.Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}

	// Create client
	client := NewClient(
		WithTransport(mockTransport),
		WithResolver(mockResolver),
	)

	// Register mock protocol
	mockProto := &mockProtocol{
		queryResult: &protocol.QueryResult{
			Online:     true,
			Name:       "Test Server",
			NumPlayers: 5,
			MaxPlayers: 20,
			Bots:       0,
		},
	}
	client.factory.Register(string(GameMinecraft), mockProto)

	// Execute query
	ctx := context.Background()
	result, err := client.Query(ctx, GameMinecraft, "example.com", nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test Server", result.Name)
	assert.Equal(t, 5, result.NumPlayers)
	assert.Equal(t, "minecraft", result.Type)
	assert.GreaterOrEqual(t, result.Ping, float64(0))
}

func TestClient_Query_UnsupportedGame(t *testing.T) {
	client := NewClient()

	ctx := context.Background()
	result, err := client.Query(ctx, GameMinecraft, "example.com", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrUnsupportedGame))

	var queryErr *QueryError
	assert.True(t, errors.As(err, &queryErr))
	assert.Equal(t, GameMinecraft, queryErr.GameType)
}

func TestClient_Query_DNSResolutionFailed(t *testing.T) {
	mockResolver := resolver.NewMockResolver()
	mockResolver.Error = errors.New("dns failed")

	client := NewClient(WithResolver(mockResolver))

	mockProto := &mockProtocol{
		queryResult: &protocol.QueryResult{Online: true},
	}
	client.factory.Register(string(GameMinecraft), mockProto)

	ctx := context.Background()
	result, err := client.Query(ctx, GameMinecraft, "example.com", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrDNSResolution))
}

func TestClient_Query_ProtocolError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockResolver := resolver.NewMockResolver()
	mockResolver.Addresses["example.com"] = &resolver.Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}

	client := NewClient(
		WithTransport(mockTransport),
		WithResolver(mockResolver),
	)

	mockProto := &mockProtocol{
		queryError: ErrTimeout,
	}
	client.factory.Register(string(GameMinecraft), mockProto)

	ctx := context.Background()
	result, err := client.Query(ctx, GameMinecraft, "example.com", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrTimeout))
}

func TestClient_Query_CustomPort(t *testing.T) {
	mockResolver := resolver.NewMockResolver()
	customPort := 30000
	mockResolver.Addresses["example.com"] = &resolver.Address{
		IP:       "127.0.0.1",
		Port:     customPort,
		Hostname: "example.com",
	}

	client := NewClient(WithResolver(mockResolver))

	mockProto := &mockProtocol{
		queryResult: &protocol.QueryResult{Online: true},
	}
	client.factory.Register(string(GameMinecraft), mockProto)

	ctx := context.Background()
	result, err := client.Query(ctx, GameMinecraft, "example.com", &customPort)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClient_Query_WithTimeout(t *testing.T) {
	client := NewClient(WithTimeout(100 * time.Millisecond))

	mockProto := &mockProtocol{
		queryResult: &protocol.QueryResult{Online: true},
	}
	client.factory.Register(string(GameMinecraft), mockProto)

	mockResolver := resolver.NewMockResolver()
	mockResolver.Addresses["example.com"] = &resolver.Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}
	client.resolver = mockResolver

	ctx := context.Background()
	result, err := client.Query(ctx, GameMinecraft, "example.com", nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestQueryWithOptions(t *testing.T) {
	mockResolver := resolver.NewMockResolver()
	mockResolver.Addresses["example.com"] = &resolver.Address{
		IP:       "127.0.0.1",
		Port:     25565,
		Hostname: "example.com",
	}

	client := NewClient(WithResolver(mockResolver))

	mockProto := &mockProtocol{
		queryResult: &protocol.QueryResult{Online: true},
	}
	client.factory.Register(string(GameMinecraft), mockProto)

	input := QueryInput{
		ServerType: GameMinecraft,
		Host:       "example.com",
		Timeout:    2 * time.Second,
	}

	ctx := context.Background()
	result, err := client.QueryWithOptions(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestQuickQuery(t *testing.T) {
	// QuickQuery should fail for unsupported game since no protocols are registered
	result, err := QuickQuery(GameMinecraft, "example.com")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrUnsupportedGame))
}
