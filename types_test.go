package queryx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGameTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		gameType GameType
		expected string
	}{
		{"Minecraft Java", GameMinecraft, "minecraft"},
		{"Minecraft Bedrock", GameMinecraftBedrock, "minecraft_bedrock"},
		{"CS 1.6", GameCS16, "cs16"},
		{"CS Source", GameCSSource, "cssource"},
		{"CS2", GameCS2, "cs2"},
		{"Rust", GameRust, "rust"},
		{"FiveM", GameFiveM, "fivem"},
		{"SA-MP", GameSAMP, "samp"},
		{"TeamSpeak", GameTeamSpeak, "teamspeak"},
		{"Discord", GameDiscord, "discord"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.gameType))
		})
	}
}

func TestQueryInputZeroValue(t *testing.T) {
	var input QueryInput
	assert.Equal(t, GameType(""), input.ServerType)
	assert.Equal(t, "", input.Host)
	assert.Nil(t, input.Port)
	assert.Equal(t, time.Duration(0), input.Timeout)
	assert.Nil(t, input.Options)
}

func TestQueryResultZeroValue(t *testing.T) {
	var result QueryResult
	assert.False(t, result.Online)
	assert.Equal(t, "", result.Name)
	assert.Equal(t, 0, result.NumPlayers)
	assert.Equal(t, 0, result.Bots)
	assert.Nil(t, result.Players)
	assert.Nil(t, result.Raw)
}

func TestPlayerZeroValue(t *testing.T) {
	var player Player
	assert.Equal(t, "", player.Name)
	assert.Equal(t, 0, player.Score)
	assert.Equal(t, float64(0), player.Duration)
}
