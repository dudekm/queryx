package bedrock

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

// buildPong builds a valid 0x1C unconnected pong packet wrapping the given MOTD.
func buildPong(motd string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteByte(idUnconnectedPong)
	_ = binary.Write(buf, binary.BigEndian, uint64(12345))     // time
	_ = binary.Write(buf, binary.BigEndian, uint64(987654321)) // server GUID
	buf.Write(rakNetMagic)                                     // 16-byte magic
	_ = binary.Write(buf, binary.BigEndian, uint16(len(motd))) // string length
	buf.WriteString(motd)
	return buf.Bytes()
}

const sampleMOTD = "MCPE;§bMy Server;390;1.20.10;5;10;1234567890;Bedrock level;Survival;1;19132;19133;"

func TestProtocol_Metadata(t *testing.T) {
	p := NewProtocol(nil, "Minecraft Bedrock")
	assert.Equal(t, "Minecraft Bedrock (Bedrock/RakNet)", p.Name())
	assert.Equal(t, 19132, p.DefaultPort())
	assert.False(t, p.SupportsSRV())
	assert.Empty(t, p.SRVService())
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:19132"] = buildPong(sampleMOTD)

	p := NewProtocol(mockTransport, "Minecraft Bedrock")
	result, err := p.Query(context.Background(), "127.0.0.1:19132")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "§bMy Server Bedrock level", result.Name)
	assert.Equal(t, 5, result.NumPlayers)
	assert.Equal(t, 10, result.MaxPlayers)
	assert.Equal(t, "1.20.10", result.Version)
	assert.GreaterOrEqual(t, result.Ping, 0)

	info, ok := result.Raw.(*ServerInfo)
	require.True(t, ok)
	assert.Equal(t, "MCPE", info.Edition)
	assert.Equal(t, "Survival", info.GameMode)
	assert.Equal(t, "390", info.ProtocolVersion)
}

func TestProtocol_Query_TooShort(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:19132"] = []byte{0x1C, 0x00}

	p := NewProtocol(mockTransport, "Minecraft Bedrock")
	result, err := p.Query(context.Background(), "127.0.0.1:19132")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProtocol_Query_WrongPacketID(t *testing.T) {
	pong := buildPong(sampleMOTD)
	pong[0] = 0x00 // corrupt the packet id

	mockTransport := transport.NewMockTransport()
	mockTransport.UDPResponses["127.0.0.1:19132"] = pong

	p := NewProtocol(mockTransport, "Minecraft Bedrock")
	_, err := p.Query(context.Background(), "127.0.0.1:19132")
	assert.Error(t, err)
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport, "Minecraft Bedrock")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := p.Query(ctx, "127.0.0.1:19132")
	assert.Error(t, err)
}

func TestParseMOTD_ShortField(t *testing.T) {
	// Missing trailing fields must not panic; they default to zero values.
	info := parseMOTD("MCPE;Name;390;1.20.10")
	assert.Equal(t, "Name", info.MOTDLine1)
	assert.Equal(t, 0, info.PlayerCount)
	assert.Equal(t, "", info.GameMode)
}
