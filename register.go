package queryx

import (
	"github.com/dudekm/queryx/internal/protocol/minecraft"
	"github.com/dudekm/queryx/internal/protocol/source"
)

// RegisterDefaultProtocols registers all default protocols with the client
func (c *Client) RegisterDefaultProtocols() {
	// Register Minecraft Java Edition
	minecraftProto := minecraft.NewProtocol(c.transport)
	c.factory.Register(string(GameMinecraft), minecraftProto)

	// Register Counter-Strike 1.6 (GoldSrc/Source Engine)
	cs16Proto := source.NewProtocol(c.transport, "Counter-Strike 1.6")
	c.factory.Register(string(GameCS16), cs16Proto)

	// Register Counter-Strike: Source
	cssProto := source.NewProtocol(c.transport, "Counter-Strike: Source")
	c.factory.Register(string(GameCSSource), cssProto)

	// Register Counter-Strike 2
	cs2Proto := source.NewProtocol(c.transport, "Counter-Strike 2")
	c.factory.Register(string(GameCS2), cs2Proto)
}

// NewClientWithDefaults creates a new client with all default protocols registered
func NewClientWithDefaults(opts ...Option) *Client {
	client := NewClient(opts...)
	client.RegisterDefaultProtocols()
	return client
}
