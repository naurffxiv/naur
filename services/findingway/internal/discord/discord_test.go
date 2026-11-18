package discord

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/tokenizer"

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

func TestTokensHeader(t *testing.T) {
	at := time.Date(2026, 9, 25, 12, 30, 0, 0, time.UTC)

	got := tokensHeader(5, 42, at)

	assert.Equal(t, "**Last 5 days of Tokens (42 listings scanned) as of 2026-09-25 12:30:00\n**", got)
}

func TestTokensMessageListsAtMost75Tokens(t *testing.T) {
	at := time.Date(2026, 9, 25, 12, 30, 0, 0, time.UTC)
	tokens := make([]tokenizer.Token, 80)
	for i := range tokens {
		tokens[i] = tokenizer.Token{String: "tok", Count: i}
	}

	msg := TokensMessage(tokens, 2, 7, at)

	assert.True(t, strings.HasPrefix(msg.Content, tokensHeader(2, 7, at)))
	assert.Equal(t, 75, strings.Count(msg.Content, "`tok`"))
}

func TestDescriptionCsvMessageAttachesCsv(t *testing.T) {
	at := time.Date(2026, 9, 25, 12, 30, 0, 0, time.UTC)

	msg := DescriptionCsvMessage(bytes.NewBufferString("Timestamp,Description\n"), at)

	assert.Equal(t, "**Description List 2026-09-25 12:30:00**", msg.Content)
	require.NotNil(t, msg.File)
	assert.Equal(t, "pf_descriptions_2026-09-25 12:30:00.csv", msg.File.Name)
	body, err := io.ReadAll(msg.File.Reader)
	require.NoError(t, err)
	assert.Equal(t, "Timestamp,Description\n", string(body))
}
