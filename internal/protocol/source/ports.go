package source

import "strings"

// DefaultPorts maps game identifiers to their default query ports
var DefaultPorts = map[string]int{
	// Valve Source Engine Games
	"tf2":                 27015,
	"l4d":                 27015,
	"l4d2":                27015,
	"gmod":                27015,
	"blackmesa":           27015,
	"dayofinfamy":         27015,
	"insurgency":          27015,
	"insurgencysandstorm": 27015,
	"killingfloor2":       27015,

	// Counter-Strike Games
	"cs16":     27015,
	"cssource": 27015,
	"cs2":      27015,

	// Survival Games Using A2S
	"ark":         27015,
	"arkascended": 27015,
	"atlas":       27015,
	"conanexiles": 27015,
	"7daystodie":  26900,
	"rust":        28015,

	// Co-op/Simulation Games
	"barotrauma":   27015,
	"hellletloose": 27015,
	"postscriptum": 27015,
	"squad":        27015,
	"risingstorm2": 27015,

	// Space/Sandbox Games
	"avorion":        27015,
	"empyrion":       30000,
	"stationeers":    27015,
	"spaceengineers": 27015,

	// Other Survival/Sandbox
	"hurtworld":     12871,
	"icarus":        17777,
	"enshrouded":    15636,
	"vrising":       27015,
	"unturned":      27015,
	"theforest":     27015,
	"noonesurvived": 27015,
	"miscreated":    27015,
	"deadpoly":      27015,
	"dysterra":      27015,
	"subsistence":   27015,
	"pixark":        27015,
	"valheim":       2456,
}

// GetDefaultPort returns the default port for a game, falling back to 27015
func GetDefaultPort(gameName string) int {
	// Normalize game name to lowercase for lookup
	normalized := strings.ToLower(gameName)

	// Try direct lookup
	if port, ok := DefaultPorts[normalized]; ok {
		return port
	}

	// Try removing spaces and special characters
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, "-", "")

	if port, ok := DefaultPorts[normalized]; ok {
		return port
	}

	// Default Source Engine port
	return 27015
}
