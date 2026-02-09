package queryx

import (
	"github.com/dudekm/queryx/internal/protocol/cfxre"
	"github.com/dudekm/queryx/internal/protocol/gamespy"
	"github.com/dudekm/queryx/internal/protocol/minecraft"
	"github.com/dudekm/queryx/internal/protocol/source"
)

// RegisterDefaultProtocols registers all default protocols with the client
func (c *Client) RegisterDefaultProtocols() {
	// Register Minecraft Java Edition
	minecraftProto := minecraft.NewProtocol(c.transport)
	c.factory.Register(string(GameMinecraft), minecraftProto)

	// Register Counter-Strike Games (Source Engine)
	cs16Proto := source.NewProtocol(c.transport, "Counter-Strike 1.6")
	c.factory.Register(string(GameCS16), cs16Proto)

	cssProto := source.NewProtocol(c.transport, "Counter-Strike: Source")
	c.factory.Register(string(GameCSSource), cssProto)

	cs2Proto := source.NewProtocol(c.transport, "Counter-Strike 2")
	c.factory.Register(string(GameCS2), cs2Proto)

	// Register Valve Source Engine Games
	tf2Proto := source.NewProtocol(c.transport, "Team Fortress 2")
	c.factory.Register(string(ServerTF2), tf2Proto)

	l4dProto := source.NewProtocol(c.transport, "Left 4 Dead")
	c.factory.Register(string(ServerL4D), l4dProto)

	l4d2Proto := source.NewProtocol(c.transport, "Left 4 Dead 2")
	c.factory.Register(string(ServerL4D2), l4d2Proto)

	gmodProto := source.NewProtocol(c.transport, "Garry's Mod")
	c.factory.Register(string(ServerGMod), gmodProto)

	blackMesaProto := source.NewProtocol(c.transport, "Black Mesa")
	c.factory.Register(string(ServerBlackMesa), blackMesaProto)

	dayOfInfamyProto := source.NewProtocol(c.transport, "Day of Infamy")
	c.factory.Register(string(ServerDayOfInfamy), dayOfInfamyProto)

	insurgencyProto := source.NewProtocol(c.transport, "Insurgency")
	c.factory.Register(string(ServerInsurgency), insurgencyProto)

	insurgencySSProto := source.NewProtocol(c.transport, "Insurgency: Sandstorm")
	c.factory.Register(string(ServerInsurgencySS), insurgencySSProto)

	kf2Proto := source.NewProtocol(c.transport, "Killing Floor 2")
	c.factory.Register(string(ServerKillingFloor2), kf2Proto)

	// Register Survival Games Using A2S Protocol
	arkProto := source.NewProtocol(c.transport, "ARK: Survival Evolved")
	c.factory.Register(string(ServerARK), arkProto)

	arkAscendedProto := source.NewProtocol(c.transport, "ARK: Survival Ascended")
	c.factory.Register(string(ServerARKAscended), arkAscendedProto)

	atlasProto := source.NewProtocol(c.transport, "ATLAS")
	c.factory.Register(string(ServerATLAS), atlasProto)

	conanProto := source.NewProtocol(c.transport, "Conan Exiles")
	c.factory.Register(string(ServerConanExiles), conanProto)

	sevenDaysProto := source.NewProtocol(c.transport, "7 Days to Die")
	c.factory.Register(string(Server7DaysToDie), sevenDaysProto)

	rustProto := source.NewProtocol(c.transport, "Rust")
	c.factory.Register(string(ServerRust), rustProto)

	// Register Co-op/Tactical Games
	barotraumaProto := source.NewProtocol(c.transport, "Barotrauma")
	c.factory.Register(string(ServerBarotrauma), barotraumaProto)

	hellLetLooseProto := source.NewProtocol(c.transport, "Hell Let Loose")
	c.factory.Register(string(ServerHellLetLoose), hellLetLooseProto)

	postScriptumProto := source.NewProtocol(c.transport, "Post Scriptum")
	c.factory.Register(string(ServerPostScriptum), postScriptumProto)

	squadProto := source.NewProtocol(c.transport, "Squad")
	c.factory.Register(string(ServerSquad), squadProto)

	risingStormProto := source.NewProtocol(c.transport, "Rising Storm 2: Vietnam")
	c.factory.Register(string(ServerRisingStorm2), risingStormProto)

	// Register Space/Sandbox Games
	avorionProto := source.NewProtocol(c.transport, "Avorion")
	c.factory.Register(string(ServerAvorion), avorionProto)

	empyrionProto := source.NewProtocol(c.transport, "Empyrion - Galactic Survival")
	c.factory.Register(string(ServerEmpyrion), empyrionProto)

	stationeersProto := source.NewProtocol(c.transport, "Stationeers")
	c.factory.Register(string(ServerStationeers), stationeersProto)

	spaceEngineersProto := source.NewProtocol(c.transport, "Space Engineers")
	c.factory.Register(string(ServerSpaceEngineers), spaceEngineersProto)

	// Register Other Survival/Sandbox Games
	hurtworldProto := source.NewProtocol(c.transport, "Hurtworld")
	c.factory.Register(string(ServerHurtworld), hurtworldProto)

	icarusProto := source.NewProtocol(c.transport, "ICARUS")
	c.factory.Register(string(ServerICARUS), icarusProto)

	enshroudedProto := source.NewProtocol(c.transport, "Enshrouded")
	c.factory.Register(string(ServerEnshrouded), enshroudedProto)

	vrisingProto := source.NewProtocol(c.transport, "V Rising")
	c.factory.Register(string(ServerVRising), vrisingProto)

	unturnedProto := source.NewProtocol(c.transport, "Unturned")
	c.factory.Register(string(ServerUnturned), unturnedProto)

	theForestProto := source.NewProtocol(c.transport, "The Forest")
	c.factory.Register(string(ServerTheForest), theForestProto)

	noOneSurvivedProto := source.NewProtocol(c.transport, "No One Survived")
	c.factory.Register(string(ServerNoOneSurvived), noOneSurvivedProto)

	miscreatedProto := source.NewProtocol(c.transport, "Miscreated")
	c.factory.Register(string(ServerMiscreated), miscreatedProto)

	deadPolyProto := source.NewProtocol(c.transport, "DeadPoly")
	c.factory.Register(string(ServerDeadPoly), deadPolyProto)

	dysterraProto := source.NewProtocol(c.transport, "Dysterra")
	c.factory.Register(string(ServerDysterra), dysterraProto)

	subsistenceProto := source.NewProtocol(c.transport, "Subsistence")
	c.factory.Register(string(ServerSubsistence), subsistenceProto)

	pixarkProto := source.NewProtocol(c.transport, "PixARK")
	c.factory.Register(string(ServerPixARK), pixarkProto)

	valheimProto := source.NewProtocol(c.transport, "Valheim")
	c.factory.Register(string(ServerValheim), valheimProto)

	// Register GameSpy Protocol Games
	arma2Proto := gamespy.NewProtocol(c.transport, "ARMA 2")
	c.factory.Register(string(ServerARMA2), arma2Proto)

	arma3Proto := gamespy.NewProtocol(c.transport, "ARMA 3")
	c.factory.Register(string(ServerARMA3), arma3Proto)

	dayzProto := gamespy.NewProtocol(c.transport, "DayZ")
	c.factory.Register(string(ServerDayZ), dayzProto)

	dayOfDragonsProto := gamespy.NewProtocol(c.transport, "Day of Dragons")
	c.factory.Register(string(ServerDayOfDragons), dayOfDragonsProto)

	// Register CFX.re HTTP Protocol Games
	fivemProto := cfxre.NewProtocol(c.transport, "FiveM")
	c.factory.Register(string(ServerFiveM), fivemProto)

	redmProto := cfxre.NewProtocol(c.transport, "RedM")
	c.factory.Register(string(ServerRedM), redmProto)

	altvProto := cfxre.NewProtocol(c.transport, "Alt:V")
	c.factory.Register(string(ServerAltV), altvProto)
}

// NewClientWithDefaults creates a new client with all default protocols registered
func NewClientWithDefaults(opts ...Option) *Client {
	client := NewClient(opts...)
	client.RegisterDefaultProtocols()
	return client
}
