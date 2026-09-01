# Supported Games

This page tracks every game/server QueryX can query. Pass the value in the
**`type`** column to the library (`client.Query(ctx, "<type>", host, port)`) or
the CLI (`queryx -type <type>`).

> **Rule:** entries in every table on this page are **always sorted
> alphabetically by game name** (case-insensitive). Keep it that way when adding
> a game. This rule is also recorded in [`CLAUDE.md`](CLAUDE.md).

**Legend:** ✅ implemented & registered &nbsp;·&nbsp; ⬜ planned (not implemented yet)

**Status:** 57 implemented · 11 protocols · 9 planned.

> Ports are the default **query** ports used when no port is supplied; they live
> in each protocol package under `internal/protocol/**`.

---

## ✅ Implemented Games

| Game | `type` | Protocol | Default Port |
|------|--------|----------|--------------|
| 7 Days to Die | `7daystodie` | Source Engine (A2S) | 26900 |
| Alt:V | `altv` | CFX.re (HTTP) | 7788 |
| ARK: Survival Ascended | `arkascended` | Source Engine (A2S) | 27015 |
| ARK: Survival Evolved | `ark` | Source Engine (A2S) | 27015 |
| ARMA 2 | `arma2` | GameSpy | 2302 |
| ARMA 3 | `arma3` | GameSpy | 2302 |
| ATLAS | `atlas` | Source Engine (A2S) | 27015 |
| Avorion | `avorion` | Source Engine (A2S) | 27015 |
| Barotrauma | `barotrauma` | Source Engine (A2S) | 27015 |
| Black Mesa | `blackmesa` | Source Engine (A2S) | 27015 |
| Conan Exiles | `conanexiles` | Source Engine (A2S) | 27015 |
| Counter-Strike 1.6 | `cs16` | Source Engine (GoldSrc) | 27015 |
| Counter-Strike 2 | `cs2` | Source Engine (A2S) | 27015 |
| Counter-Strike: Source | `cssource` | Source Engine (A2S) | 27015 |
| Day of Dragons | `dayofdragons` | GameSpy | 7777 |
| Day of Infamy | `dayofinfamy` | Source Engine (A2S) | 27015 |
| DayZ | `dayz` | GameSpy | 2302 |
| DeadPoly | `deadpoly` | Source Engine (A2S) | 27015 |
| Dysterra | `dysterra` | Source Engine (A2S) | 27015 |
| Empyrion - Galactic Survival | `empyrion` | Source Engine (A2S) | 30000 |
| Enshrouded | `enshrouded` | Source Engine (A2S) | 15636 |
| FiveM | `fivem` | CFX.re (HTTP) | 30120 |
| Garry's Mod | `gmod` | Source Engine (A2S) | 27015 |
| Hell Let Loose | `hellletloose` | Source Engine (A2S) | 27015 |
| Hurtworld | `hurtworld` | Source Engine (A2S) | 12871 |
| Hytale | `hytale` | HyQuery | 5520 |
| ICARUS | `icarus` | Source Engine (A2S) | 17777 |
| Insurgency | `insurgency` | Source Engine (A2S) | 27015 |
| Insurgency: Sandstorm | `insurgencysandstorm` | Source Engine (A2S) | 27015 |
| Killing Floor 2 | `killingfloor2` | Source Engine (A2S) | 27015 |
| Left 4 Dead | `l4d` | Source Engine (A2S) | 27015 |
| Left 4 Dead 2 | `l4d2` | Source Engine (A2S) | 27015 |
| Minecraft: Bedrock Edition | `minecraftbedrock` | Bedrock (RakNet unconnected ping) | 19132 |
| Minecraft: Java Edition | `minecraft` | Minecraft (Server List Ping) | 25565 |
| Miscreated | `miscreated` | Source Engine (A2S) | 27015 |
| Mordhau | `mordhau` | Source Engine (A2S) | 27015 |
| Multi Theft Auto | `mta` | ASE | 22003 |
| Mumble | `mumble` | Mumble (UDP ping) | 64738 |
| No One Survived | `noonesurvived` | Source Engine (A2S) | 27015 |
| PixARK | `pixark` | Source Engine (A2S) | 27015 |
| Post Scriptum | `postscriptum` | Source Engine (A2S) | 27015 |
| Project Zomboid | `projectzomboid` | Source Engine (A2S) | 16261 |
| Quake III / idTech3 | `quake3` | idTech3 (getstatus) | 27960 |
| RedM | `redm` | CFX.re (HTTP) | 30120 |
| Rising Storm 2: Vietnam | `risingstorm2` | Source Engine (A2S) | 27015 |
| Rust | `rust` | Source Engine (A2S) | 28015 |
| SA-MP | `samp` | SA-MP | 7777 |
| Space Engineers | `spaceengineers` | Source Engine (A2S) | 27015 |
| Squad | `squad` | Source Engine (A2S) | 27015 |
| Stationeers | `stationeers` | Source Engine (A2S) | 27015 |
| Subsistence | `subsistence` | Source Engine (A2S) | 27015 |
| Team Fortress 2 | `tf2` | Source Engine (A2S) | 27015 |
| TeamSpeak 3 | `teamspeak` | TeamSpeak 3 (ServerQuery) | 10011 |
| The Forest | `theforest` | Source Engine (A2S) | 27015 |
| Unturned | `unturned` | Source Engine (A2S) | 27015 |
| V Rising | `vrising` | Source Engine (A2S) | 27015 |
| Valheim | `valheim` | Source Engine (A2S) | 2456 |

## ⬜ Planned / Not Yet Implemented

A roadmap signpost. These are **deliberately not implemented** because they do
not fit QueryX's model of an *anonymous, documented, per-server* query (the
Blocker column says why) — implementing them would mean shipping credential
flows or unverified/guessed wire formats. They stay here as candidates; the
blocker is what has to be resolved first.

`discord` already has a `type` constant declared in `types.go` but no registered
protocol.

| Game | proposed `type` | Intended Protocol | Blocker |
|------|-----------------|-------------------|---------|
| BeamMP | `beammp` | HTTP (BeamMP backend) | No per-server query — only a central master list |
| Discord | `discord` | Discord Widget API (HTTP) | Keyed by guild ID, not host+port; needs a non-DNS query flow |
| Factorio | `factorio` | Matching-server API (HTTP) | Requires a Factorio account token |
| Minetest / Luanti | `minetest` | Minetest server query (UDP) | Needs a peer handshake; no simple documented request packet |
| Palworld | `palworld` | REST API (HTTP) | Requires the admin password (REST/RCON) |
| Satisfactory | `satisfactory` | Lightweight Query (UDP) | Handshake framing not verified against a real server |
| Teeworlds | `teeworlds` | Teeworlds server info (UDP) | Token-based since 0.7; exact framing unverified |
| Terraria | `terraria` | TShock REST API (HTTP) | No vanilla query; needs TShock plugin + token |
| Ventrilo | `ventrilo` | Ventrilo status (UDP) | Proprietary, obfuscated protocol |

---

## Adding a New Game

See the **Adding a New Game Protocol** section in [`CLAUDE.md`](CLAUDE.md) and
the **Adding a New Game** walkthrough in [`README.md`](README.md). In short:

1. Implement the `Protocol` interface in `internal/protocol/<game>/` (or reuse
   an existing protocol package, e.g. `source` for A2S games).
2. Register it in `register.go` → `RegisterDefaultProtocols()`.
3. Add the `Server<Game>` constant in `types.go`.
4. Add unit tests (mock transport) and an integration test.
5. Move the game from the **Planned** table to **Implemented** here —
   **keeping both tables sorted alphabetically by game name**.
