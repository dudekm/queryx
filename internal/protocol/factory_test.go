package protocol

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockProtocol is a mock implementation for testing
type MockProtocol struct {
	name        string
	defaultPort int
	supportsSRV bool
	srvService  string
	queryResult *QueryResult
	queryError  error
}

func (m *MockProtocol) Query(ctx context.Context, addr string) (*QueryResult, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.queryResult, nil
}

func (m *MockProtocol) Name() string {
	return m.name
}

func (m *MockProtocol) DefaultPort() int {
	return m.defaultPort
}

func (m *MockProtocol) SupportsSRV() bool {
	return m.supportsSRV
}

func (m *MockProtocol) SRVService() string {
	return m.srvService
}

func TestFactory_RegisterAndGet(t *testing.T) {
	factory := NewFactory()
	protocol := &MockProtocol{
		name:        "minecraft",
		defaultPort: 25565,
	}

	factory.Register("minecraft", protocol)

	retrieved, err := factory.Get("minecraft")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "minecraft", retrieved.Name())
}

func TestFactory_GetUnsupported(t *testing.T) {
	factory := NewFactory()

	protocol, err := factory.Get("minecraft")
	assert.Error(t, err)
	assert.Nil(t, protocol)
	assert.True(t, errors.Is(err, ErrUnsupportedGame))
}

func TestFactory_List(t *testing.T) {
	factory := NewFactory()

	protocol1 := &MockProtocol{name: "minecraft"}
	protocol2 := &MockProtocol{name: "cs16"}

	factory.Register("minecraft", protocol1)
	factory.Register("cs16", protocol2)

	types := factory.List()
	assert.Len(t, types, 2)
	assert.Contains(t, types, "minecraft")
	assert.Contains(t, types, "cs16")
}

func TestFactory_Unregister(t *testing.T) {
	factory := NewFactory()
	protocol := &MockProtocol{name: "minecraft"}

	factory.Register("minecraft", protocol)

	// Verify it exists
	_, err := factory.Get("minecraft")
	assert.NoError(t, err)

	// Unregister
	factory.Unregister("minecraft")

	// Verify it's gone
	_, err = factory.Get("minecraft")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedGame))
}

func TestFactory_ConcurrentAccess(t *testing.T) {
	factory := NewFactory()
	var wg sync.WaitGroup

	// Concurrent registration
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			protocol := &MockProtocol{name: "test"}
			factory.Register("minecraft", protocol)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			factory.Get("minecraft")
		}(i)
	}

	wg.Wait()
}

func TestFactory_ListEmpty(t *testing.T) {
	factory := NewFactory()
	types := factory.List()
	assert.Empty(t, types)
}
