package discord

import (
	"testing"
	"time"

	"os"

	"github.com/Veraticus/findingway/internal/ffxiv"

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

func TestPostListings(t *testing.T) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		t.Skip("DISCORD_TOKEN not set, skipping integration test")
	}

	disc := &Discord{Token: token}
	err := disc.Start()
	require.NoError(t, err)

	now := time.Now()
	listings := []struct {
		listing *ffxiv.Listing
		wants   time.Time
	}{
		{
			listing: &ffxiv.Listing{Updated: "a second ago"},
			wants:   now.Add(time.Duration(-1) * time.Second),
		},
		{
			listing: &ffxiv.Listing{Updated: "a minute ago"},
			wants:   now.Add(time.Duration(-1) * time.Minute),
		},
		{
			listing: &ffxiv.Listing{Updated: "an hour ago"},
			wants:   now.Add(time.Duration(-1) * time.Hour),
		},
		{
			listing: &ffxiv.Listing{Updated: "37 seconds ago"},
			wants:   now.Add(time.Duration(-37) * time.Second),
		},
		{
			listing: &ffxiv.Listing{Updated: "22 minutes ago"},
			wants:   now.Add(time.Duration(-22) * time.Minute),
		},
		{
			listing: &ffxiv.Listing{Updated: "2 hours ago"},
			wants:   now.Add(time.Duration(-2) * time.Hour),
		},
	}

	var ffxivListings ffxiv.Listings
	for _, item := range listings {
		ffxivListings.Listings = append(ffxivListings.Listings, item.listing)
	}

	listErr := disc.PostListings(unitTestChannelID, &ffxivListings, "Dragonsong's Reprise (Ultimate)", "Aether")
	assert.NoError(t, listErr)
}

func TestTokensHeader(t *testing.T) {
	at := time.Date(2026, 9, 25, 12, 30, 0, 0, time.UTC)

	got := tokensHeader(5, 42, at)

	assert.Equal(t, "**Last 5 days of Tokens (42 listings scanned) as of 2026-09-25 12:30:00\n**", got)
}
