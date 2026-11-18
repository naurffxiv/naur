package discord

import (
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const unitTestChannelID = "1566200857207382106"

func TestStartDiscord(t *testing.T) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		t.Skip("DISCORD_TOKEN not set, skipping integration test")
	}

	disc := &Discord{Token: token}
	err := disc.Start()

	assert.NoError(t, err)
}

func TestSend(t *testing.T) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		t.Skip("DISCORD_TOKEN not set, skipping integration test")
	}

	disc := &Discord{Token: token}
	require.NoError(t, disc.Start())
	defer func() { _ = disc.Session.Close() }()

	err := disc.Send(unitTestChannelID, &discordgo.MessageSend{
		Content: "findingway smoke test " + time.Now().Format(time.DateTime),
	})
	assert.NoError(t, err)
}
