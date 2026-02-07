package queryx

import "time"

// ServerType represents the type of server to query (games, voice servers, etc.)
type ServerType string

// GameType is an alias for ServerType (backward compatibility)
type GameType = ServerType

const (
	ServerMinecraft        ServerType = "minecraft"
	ServerMinecraftBedrock ServerType = "minecraft_bedrock"
	ServerCS16             ServerType = "cs16"
	ServerCSSource         ServerType = "cssource"
	ServerCS2              ServerType = "cs2"
	ServerRust             ServerType = "rust"
	ServerFiveM            ServerType = "fivem"
	ServerSAMP             ServerType = "samp"
	ServerTeamSpeak        ServerType = "teamspeak"
	ServerDiscord          ServerType = "discord"
)

// Backward compatibility aliases
const (
	GameMinecraft        = ServerMinecraft
	GameMinecraftBedrock = ServerMinecraftBedrock
	GameCS16             = ServerCS16
	GameCSSource         = ServerCSSource
	GameCS2              = ServerCS2
	GameRust             = ServerRust
	GameFiveM            = ServerFiveM
	GameSAMP             = ServerSAMP
	GameTeamSpeak        = ServerTeamSpeak
	GameDiscord          = ServerDiscord
)

// QueryInput contains all parameters needed to query a server
type QueryInput struct {
	ServerType ServerType
	Host       string
	Port       *int
	Timeout    time.Duration
	Options    map[string]interface{}
}

// QueryResult contains the parsed response from a server
type QueryResult struct {
	Online     bool
	Name       string
	Map        string
	NumPlayers int      // Current number of players online
	MaxPlayers int      // Maximum number of players
	Players    []Player // Detailed list of players (if available)
	Bots       int      // Number of bots
	Type       string   // Server type (minecraft, cs2, etc.)
	Version    string   // Server version
	Ping       time.Duration
	Connect    string // Connection string (host:port)
	Password   bool   // Whether server requires password
	Extra      map[string]interface{}
	Raw        []byte // Raw server response for debugging
	QueriedAt  time.Time
}

// Player represents a single player on the server
type Player struct {
	Name     string
	Score    int
	Duration time.Duration
}
