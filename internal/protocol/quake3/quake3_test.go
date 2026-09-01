package quake3

import (
	"context"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildStatusResponse(info string, playerLines ...string) []byte {
	out := append([]byte{}, connlessHeader...)
	out = append(out, []byte("statusResponse\n")...)
	out = append(out, []byte(info+"\n")...)
	for _, pl := range playerLines {
		out = append(out, []byte(pl+"\n")...)
	}
	return out
}

const sampleInfo = `\sv_hostname\Test Q3 Server\mapname\q3dm6\sv_maxclients\16\g_gametype\0\version\Q3 1.32\g_needpass\1`

func TestProtocol_Metadata(t *testing.T) {
	p := NewProtocol(nil, "Quake III")
	assert.Equal(t, "Quake III (idTech3/getstatus)", p.Name())
	assert.Equal(t, 27960, p.DefaultPort())
	assert.False(t, p.SupportsSRV())
	assert.Empty(t, p.SRVService())
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:27960"] = buildStatusResponse(
		sampleInfo,
		`10 50 "Player One"`,
		`5 80 "Player Two"`,
	)

	p := NewProtocol(mockTransport, "Quake III")
	result, err := p.Query(context.Background(), "127.0.0.1:27960")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test Q3 Server", result.Name)
	assert.Equal(t, "q3dm6", result.Map)
	assert.Equal(t, 16, result.MaxPlayers)
	assert.Equal(t, 2, result.NumPlayers)
	assert.Equal(t, "Q3 1.32", result.Version)
	assert.True(t, result.Password, "g_needpass=1 means password required")
	require.Len(t, result.Players, 2)
	assert.Equal(t, "Player One", result.Players[0].Name)
	assert.Equal(t, 10, result.Players[0].Score)

	info, ok := result.Raw.(*ServerInfo)
	require.True(t, ok)
	assert.Equal(t, "0", info.Vars["g_gametype"])
	assert.Equal(t, 80, info.Players[1].Ping)
}

func TestProtocol_Query_NoPlayers(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:27960"] = buildStatusResponse(sampleInfo)

	p := NewProtocol(mockTransport, "Quake III")
	result, err := p.Query(context.Background(), "127.0.0.1:27960")

	require.NoError(t, err)
	assert.Equal(t, 0, result.NumPlayers)
	assert.Empty(t, result.Players)
}

func TestProtocol_Query_MissingHeader(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:27960"] = []byte("statusResponse\n")

	p := NewProtocol(mockTransport, "Quake III")
	_, err := p.Query(context.Background(), "127.0.0.1:27960")
	assert.Error(t, err)
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "Quake III")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := p.Query(ctx, "127.0.0.1:27960")
	assert.Error(t, err)
}
