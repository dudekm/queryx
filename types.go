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

	// Source Engine Games
	ServerTF2           ServerType = "tf2"
	ServerL4D           ServerType = "l4d"
	ServerL4D2          ServerType = "l4d2"
	ServerGMod          ServerType = "gmod"
	ServerBlackMesa     ServerType = "blackmesa"
	ServerDayOfInfamy   ServerType = "dayofinfamy"
	ServerInsurgency    ServerType = "insurgency"
	ServerInsurgencySS  ServerType = "insurgencysandstorm"
	ServerKillingFloor2 ServerType = "killingfloor2"

	// Games Using A2S Protocol
	ServerARK            ServerType = "ark"
	ServerARKAscended    ServerType = "arkascended"
	ServerATLAS          ServerType = "atlas"
	ServerConanExiles    ServerType = "conanexiles"
	Server7DaysToDie     ServerType = "7daystodie"
	ServerBarotrauma     ServerType = "barotrauma"
	ServerHellLetLoose   ServerType = "hellletloose"
	ServerPostScriptum   ServerType = "postscriptum"
	ServerSquad          ServerType = "squad"
	ServerRisingStorm2   ServerType = "risingstorm2"
	ServerAvorion        ServerType = "avorion"
	ServerEmpyrion       ServerType = "empyrion"
	ServerStationeers    ServerType = "stationeers"
	ServerSpaceEngineers ServerType = "spaceengineers"
	ServerHurtworld      ServerType = "hurtworld"
	ServerICARUS         ServerType = "icarus"
	ServerEnshrouded     ServerType = "enshrouded"
	ServerVRising        ServerType = "vrising"
	ServerUnturned       ServerType = "unturned"
	ServerTheForest      ServerType = "theforest"
	ServerNoOneSurvived  ServerType = "noonesurvived"
	ServerMiscreated     ServerType = "miscreated"
	ServerDeadPoly       ServerType = "deadpoly"
	ServerDysterra       ServerType = "dysterra"
	ServerSubsistence    ServerType = "subsistence"
	ServerPixARK         ServerType = "pixark"
	ServerValheim        ServerType = "valheim"

	// GameSpy Protocol Games
	ServerARMA2        ServerType = "arma2"
	ServerARMA3        ServerType = "arma3"
	ServerDayZ         ServerType = "dayz"
	ServerDayOfDragons ServerType = "dayofdragons"

	// CFX.re HTTP Protocol Games
	ServerRedM ServerType = "redm"
	ServerAltV ServerType = "altv"
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
	Online     bool                   `json:"online"`
	Name       string                 `json:"name"`
	Map        string                 `json:"map,omitempty"`
	NumPlayers int                    `json:"numPlayers"`        // Current number of players online
	MaxPlayers int                    `json:"maxPlayers"`        // Maximum number of players
	Players    []Player               `json:"players,omitempty"` // Detailed list of players (if available)
	Bots       int                    `json:"bots,omitempty"`    // Number of bots
	Type       string                 `json:"type"`              // Server type (minecraft, cs2, etc.)
	Version    string                 `json:"version,omitempty"` // Server version
	Ping       time.Duration          `json:"ping"`
	Connect    string                 `json:"connect,omitempty"`  // Connection string (host:port)
	Password   bool                   `json:"password,omitempty"` // Whether server requires password
	Extra      map[string]interface{} `json:"extra,omitempty"`
	Raw        interface{}            `json:"raw,omitempty"` // Raw server response as parsed JSON (for protocols that support it)
	QueriedAt  time.Time              `json:"queriedAt"`
}

// Player represents a single player on the server
type Player struct {
	Name     string        `json:"name"`
	Score    int           `json:"score,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}
