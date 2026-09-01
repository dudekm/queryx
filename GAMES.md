# Supported Games

This page tracks every game/server that QueryX can query, grouped by the
protocol it uses. Pass the value in the **`type`** column to the library
(`client.Query(ctx, "<type>", host, port)`) or the CLI (`queryx -type <type>`).

**Legend:** ✅ implemented & registered &nbsp;·&nbsp; ⬜ planned (constant exists, protocol not wired yet)

**Status:** 52 games implemented across 8 protocols.

> Ports listed are the default **query** ports used when no port is supplied.
> They are defined in each protocol package (see `internal/protocol/**`).

---

## Minecraft (Server List Ping, TCP)

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| Minecraft: Java Edition | `minecraft` | 25565 | ✅ |
| Minecraft: Bedrock Edition | `minecraftbedrock` | 19132 | ⬜ |

## Source Engine (A2S, UDP)

Handles both Source (`0x49`) and GoldSrc (`0x6d`) responses, with automatic
challenge-response handling.

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| Counter-Strike 1.6 | `cs16` | 27015 | ✅ |
| Counter-Strike: Source | `cssource` | 27015 | ✅ |
| Counter-Strike 2 | `cs2` | 27015 | ✅ |
| Team Fortress 2 | `tf2` | 27015 | ✅ |
| Left 4 Dead | `l4d` | 27015 | ✅ |
| Left 4 Dead 2 | `l4d2` | 27015 | ✅ |
| Garry's Mod | `gmod` | 27015 | ✅ |
| Black Mesa | `blackmesa` | 27015 | ✅ |
| Day of Infamy | `dayofinfamy` | 27015 | ✅ |
| Insurgency | `insurgency` | 27015 | ✅ |
| Insurgency: Sandstorm | `insurgencysandstorm` | 27015 | ✅ |
| Killing Floor 2 | `killingfloor2` | 27015 | ✅ |
| Rust | `rust` | 28015 | ✅ |
| ARK: Survival Evolved | `ark` | 27015 | ✅ |
| ARK: Survival Ascended | `arkascended` | 27015 | ✅ |
| ATLAS | `atlas` | 27015 | ✅ |
| Conan Exiles | `conanexiles` | 27015 | ✅ |
| 7 Days to Die | `7daystodie` | 26900 | ✅ |
| Barotrauma | `barotrauma` | 27015 | ✅ |
| Hell Let Loose | `hellletloose` | 27015 | ✅ |
| Post Scriptum | `postscriptum` | 27015 | ✅ |
| Squad | `squad` | 27015 | ✅ |
| Rising Storm 2: Vietnam | `risingstorm2` | 27015 | ✅ |
| Avorion | `avorion` | 27015 | ✅ |
| Empyrion - Galactic Survival | `empyrion` | 30000 | ✅ |
| Stationeers | `stationeers` | 27015 | ✅ |
| Space Engineers | `spaceengineers` | 27015 | ✅ |
| Hurtworld | `hurtworld` | 12871 | ✅ |
| ICARUS | `icarus` | 17777 | ✅ |
| Enshrouded | `enshrouded` | 15636 | ✅ |
| V Rising | `vrising` | 27015 | ✅ |
| Unturned | `unturned` | 27015 | ✅ |
| The Forest | `theforest` | 27015 | ✅ |
| No One Survived | `noonesurvived` | 27015 | ✅ |
| Miscreated | `miscreated` | 27015 | ✅ |
| DeadPoly | `deadpoly` | 27015 | ✅ |
| Dysterra | `dysterra` | 27015 | ✅ |
| Subsistence | `subsistence` | 27015 | ✅ |
| PixARK | `pixark` | 27015 | ✅ |
| Valheim | `valheim` | 2456 | ✅ |

## GameSpy (Query Protocol, UDP)

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| ARMA 2 | `arma2` | 2302 | ✅ |
| ARMA 3 | `arma3` | 2302 | ✅ |
| DayZ | `dayz` | 2302 | ✅ |
| Day of Dragons | `dayofdragons` | 7777 | ✅ |

## CFX.re (HTTP: `/info.json`, `/players.json`, `/dynamic.json`)

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| FiveM | `fivem` | 30120 | ✅ |
| RedM | `redm` | 30120 | ✅ |
| Alt:V | `altv` | 7788 | ✅ |

## SA-MP (San Andreas Multiplayer, UDP)

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| SA-MP | `samp` | 7777 | ✅ |

## Multi Theft Auto (ASE Protocol, UDP)

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| Multi Theft Auto | `mta` | 22003 | ✅ |

## TeamSpeak 3 (ServerQuery)

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| TeamSpeak 3 | `teamspeak` | 10011 | ✅ |

## Hytale (HyQuery Protocol)

| Game | `type` | Default Port | Status |
|------|--------|--------------|--------|
| Hytale | `hytale` | 5520 | ✅ |

---

## Planned / Not Yet Implemented

These `type` constants exist for forward-compatibility but do not yet have a
registered protocol implementation.

| Game | `type` | Intended Protocol | Status |
|------|--------|-------------------|--------|
| Minecraft: Bedrock Edition | `minecraftbedrock` | RakNet (UDP) | ⬜ |
| Discord | `discord` | Discord Widget API (HTTP) | ⬜ |

---

## Adding a New Game

See the **Adding a New Game Protocol** section in [`CLAUDE.md`](CLAUDE.md) and
the **Adding New Protocol** walkthrough in [`README.md`](README.md). In short:

1. Implement the `Protocol` interface in `internal/protocol/<game>/`.
2. Register it in `register.go` → `RegisterDefaultProtocols()`.
3. Add the `Server<Game>` constant in `types.go`.
4. Add unit tests (mock transport) and an integration test.
5. Add a row to this file.
