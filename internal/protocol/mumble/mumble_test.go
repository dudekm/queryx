package mumble

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildPong builds a valid 24-byte Mumble ping response.
func buildPong(version uint32, users, maxUsers, bandwidth uint32) []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.BigEndian, version)
	_ = binary.Write(buf, binary.BigEndian, uint64(pingIdent)) // echoed ident
	_ = binary.Write(buf, binary.BigEndian, users)
	_ = binary.Write(buf, binary.BigEndian, maxUsers)
	_ = binary.Write(buf, binary.BigEndian, bandwidth)
	return buf.Bytes()
}

func TestProtocol_Metadata(t *testing.T) {
	p := NewProtocol(nil, "Mumble")
	assert.Equal(t, "Mumble (Mumble UDP ping)", p.Name())
	assert.Equal(t, 64738, p.DefaultPort())
	assert.False(t, p.SupportsSRV())
	assert.Empty(t, p.SRVService())
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	// version 1.4.230 -> 0x0001 04 E6
	mockTransport.UDPResponses["127.0.0.1:64738"] = buildPong(0x000104E6, 12, 100, 558000)

	p := NewProtocol(mockTransport, "Mumble")
	result, err := p.Query(context.Background(), "127.0.0.1:64738")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, 12, result.NumPlayers)
	assert.Equal(t, 100, result.MaxPlayers)
	assert.Equal(t, "1.4.230", result.Version)
	assert.GreaterOrEqual(t, result.Ping, 0)

	info, ok := result.Raw.(*ServerInfo)
	require.True(t, ok)
	assert.Equal(t, 558000, info.AllowedBandwidth)
	assert.Equal(t, 1, info.VersionMajor)
	assert.Equal(t, 4, info.VersionMinor)
	assert.Equal(t, 230, info.VersionPatch)
}

func TestProtocol_Query_TooShort(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:64738"] = []byte{0x00, 0x01, 0x02}

	p := NewProtocol(mockTransport, "Mumble")
	_, err := p.Query(context.Background(), "127.0.0.1:64738")
	assert.Error(t, err)
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "Mumble")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := p.Query(ctx, "127.0.0.1:64738")
	assert.Error(t, err)
}

func TestBuildPing(t *testing.T) {
	req := buildPing()
	assert.Len(t, req, requestLen)
	assert.Equal(t, uint32(pingType), binary.BigEndian.Uint32(req[0:4]))
}
