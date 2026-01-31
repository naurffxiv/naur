package discord

import (
	"testing"
	"time"

	"os"

	"github.com/Veraticus/findingway/internal/ffxiv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartDiscord(t *testing.T) {
	token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		t.Skip("DISCORD_TOKEN not set, skipping integration test")
	}

	disc := &Discord{Token: token}
	err := disc.Start()

	assert.NoError(t, err)
}

func TestPostListings(t *testing.T) {
	token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
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

	// testing in #staff-actions
	listErr := disc.PostListings("1174350271304958032", &ffxivListings, "Dragonsong's Reprise (Ultimate)", "Aether")
	assert.NoError(t, listErr)
}
