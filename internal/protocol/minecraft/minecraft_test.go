package minecraft

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dudekm/queryx/internal/transport"
	"github.com/stretchr/testify/assert"
)

func TestProtocol_Name(t *testing.T) {
	p := NewProtocol(nil)
	assert.Equal(t, "Minecraft Java Edition", p.Name())
}

func TestProtocol_DefaultPort(t *testing.T) {
	p := NewProtocol(nil)
	assert.Equal(t, 25565, p.DefaultPort())
}

func TestProtocol_SupportsSRV(t *testing.T) {
	p := NewProtocol(nil)
	assert.True(t, p.SupportsSRV())
}

func TestProtocol_SRVService(t *testing.T) {
	p := NewProtocol(nil)
	assert.Equal(t, "minecraft", p.SRVService())
}

func TestProtocol_BuildHandshakePacket(t *testing.T) {
	p := NewProtocol(nil)
	packet, err := p.buildHandshakePacket("localhost:25565")

	assert.NoError(t, err)
	assert.NotNil(t, packet)
	assert.Greater(t, len(packet), 0)
}

func TestProtocol_BuildStatusRequestPacket(t *testing.T) {
	p := NewProtocol(nil)
	packet := p.buildStatusRequestPacket()

	assert.NotNil(t, packet)
	assert.Equal(t, []byte{0x01, 0x00}, packet)
}

func TestProtocol_ConvertToQueryResult(t *testing.T) {
	p := NewProtocol(nil)

	resp := &serverResponse{
		Version: struct {
			Name     string `json:"name"`
			Protocol int    `json:"protocol"`
		}{
			Name:     "1.20.1",
			Protocol: 763,
		},
		Players: struct {
			Max    int `json:"max"`
			Online int `json:"online"`
			Sample []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"sample"`
		}{
			Max:    100,
			Online: 50,
			Sample: []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}{
				{Name: "Player1", ID: "uuid1"},
				{Name: "Player2", ID: "uuid2"},
			},
		},
		Description: "A Minecraft Server",
		Favicon:     "data:image/png;base64,abc123",
	}

	result := p.convertToQueryResult(resp)

	assert.True(t, result.Online)
	assert.Equal(t, "A Minecraft Server", result.Name)
	assert.Equal(t, "1.20.1", result.Version)
	assert.Equal(t, 50, result.NumPlayers)
	assert.Equal(t, 100, result.MaxPlayers)
	assert.Len(t, result.Players, 2)
	assert.Equal(t, "Player1", result.Players[0].Name)
	assert.Equal(t, "Player2", result.Players[1].Name)
	assert.Equal(t, 763, result.Extra["protocol"])
	assert.Equal(t, "data:image/png;base64,abc123", result.Extra["favicon"])
}

func TestProtocol_Query_Success(t *testing.T) {
	mockTransport := transport.NewMockTransport()

	// Create a valid Minecraft server response
	responseJSON := map[string]interface{}{
		"version": map[string]interface{}{
			"name":     "1.20.1",
			"protocol": 763,
		},
		"players": map[string]interface{}{
			"max":    100,
			"online": 25,
			"sample": []map[string]interface{}{
				{"name": "TestPlayer", "id": "uuid-here"},
			},
		},
		"description": "Test Server",
		"favicon":     "",
	}

	jsonBytes, _ := json.Marshal(responseJSON)

	// Build mock response packet
	responseBuf := &bytes.Buffer{}
	// Packet length
	writeVarInt(responseBuf, 1+len(jsonBytes)+5) // ID + JSON length varint + JSON
	// Packet ID
	writeVarInt(responseBuf, 0x00)
	// JSON length
	writeVarInt(responseBuf, len(jsonBytes))
	// JSON data
	responseBuf.Write(jsonBytes)

	mockTransport.TCPResponses["127.0.0.1:25565"] = responseBuf.Bytes()

	p := NewProtocol(mockTransport)
	ctx := context.Background()

	result, err := p.Query(ctx, "127.0.0.1:25565")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Online)
	assert.Equal(t, "Test Server", result.Name)
	assert.Equal(t, "1.20.1", result.Version)
	assert.Equal(t, 25, result.NumPlayers)
	assert.Equal(t, 100, result.MaxPlayers)
	assert.NotNil(t, result.Raw)
}

func TestProtocol_Query_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.TCPError = assert.AnError

	p := NewProtocol(mockTransport)
	ctx := context.Background()

	result, err := p.Query(ctx, "127.0.0.1:25565")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestProtocol_Query_Timeout(t *testing.T) {
	mockTransport := transport.NewMockTransport()
	mockTransport.Latency = 2 * time.Second

	p := NewProtocol(mockTransport)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Query(ctx, "127.0.0.1:25565")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProtocol_ParseResponse_InvalidPacketID(t *testing.T) {
	p := NewProtocol(nil)

	// Build response with wrong packet ID
	buf := &bytes.Buffer{}
	writeVarInt(buf, 2)    // packet length
	writeVarInt(buf, 0x99) // wrong packet ID

	result, err := p.parseResponse(buf.Bytes())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unexpected packet ID")
}

func TestProtocol_ParseResponse_InvalidJSON(t *testing.T) {
	p := NewProtocol(nil)

	invalidJSON := []byte("{invalid json}")

	buf := &bytes.Buffer{}
	writeVarInt(buf, 1+len(invalidJSON)+5)
	writeVarInt(buf, 0x00)
	writeVarInt(buf, len(invalidJSON))
	buf.Write(invalidJSON)

	result, err := p.parseResponse(buf.Bytes())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}
