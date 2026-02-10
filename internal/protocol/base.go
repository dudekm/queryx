package protocol

import "github.com/dudekm/queryx/internal/transport"

// BaseProtocol provides common fields for protocol implementations (Infrastructure)
// Embed this in your protocol struct to avoid duplication (DRY principle)
type BaseProtocol struct {
	Transport transport.Transport
	GameName  string
}

// NewBaseProtocol creates a new base protocol with transport and game name (Factory)
func NewBaseProtocol(t transport.Transport, gameName string) BaseProtocol {
	return BaseProtocol{
		Transport: t,
		GameName:  gameName,
	}
}
