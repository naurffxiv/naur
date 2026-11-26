package discord

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/Veraticus/findingway/internal/tokenizer"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanChannelDeletesOnlyTheBotsMessages(t *testing.T) {
	fake := &fakeDiscord{messages: []*discordgo.Message{
		{ID: "1", Author: &discordgo.User{ID: "bot"}},
		{ID: "2", Author: &discordgo.User{ID: "someone"}},
		{ID: "3", Author: &discordgo.User{ID: "bot"}},
	}}

	require.NoError(t, fake.bot(t).CleanChannel("c1"))

	require.Len(t, fake.requests, 2)
	del := fake.requests[1]
	assert.Equal(t, "POST /channels/c1/messages/bulk-delete", del.Method+" "+del.Path)
	assert.Equal(t, []any{"1", "3"}, del.JSON["messages"])
}

func TestCleanChannelWithoutBotMessagesDeletesNothing(t *testing.T) {
	fake := &fakeDiscord{messages: []*discordgo.Message{{ID: "2", Author: &discordgo.User{ID: "someone"}}}}

	require.NoError(t, fake.bot(t).CleanChannel("c1"))

	assert.Len(t, fake.requests, 1)
}

func listing(id, updated string) *ffxiv.Listing {
	return &ffxiv.Listing{
		Id: id, Duty: "Dancing Mad (Ultimate)", DataCentre: "Aether",
		Creator: "Creator " + id, Description: "desc " + id, Updated: updated, Expires: "in 50 minutes",
	}
}

func embedOf(r request) map[string]any {
	return r.JSON["embeds"].([]any)[0].(map[string]any)
}

func TestPostListingsSendsHeaderThenBatchesOfFive(t *testing.T) {
	fake := &fakeDiscord{}
	ls := &ffxiv.Listings{}
	for i := 0; i < 7; i++ {
		ls.Add(listing(fmt.Sprint(i), "a minute ago"))
	}
	ls.Add(&ffxiv.Listing{Id: "other", Duty: "Dancing Mad (Ultimate)", DataCentre: "Light", Updated: "now"})

	require.NoError(t, fake.bot(t).PostListings("c1", ls, "Dancing Mad (Ultimate)", "Aether"))

	posts := fake.posts("c1")
	require.Len(t, posts, 3, "header, then 5 listings, then 2")
	header := embedOf(posts[0])
	assert.Equal(t, "Dancing Mad (Ultimate) PFs (Aether)", header["title"])
	assert.True(t, strings.HasPrefix(header["description"].(string), "Found 7 listings <t:"))
	assert.Len(t, embedOf(posts[1])["fields"], 15, "three fields per listing")
	assert.Len(t, embedOf(posts[2])["fields"], 6)

	first := embedOf(posts[1])["fields"].([]any)
	assert.Equal(t, "Creator 0", first[0].(map[string]any)["name"])
	assert.Equal(t, "```desc 0```", first[1].(map[string]any)["value"])
}

func TestPostListingsDropsStaleListingsWhenFreshOnesExist(t *testing.T) {
	fake := &fakeDiscord{}
	ls := &ffxiv.Listings{Listings: []*ffxiv.Listing{listing("fresh", "a minute ago"), listing("stale", "10 minutes ago")}}

	require.NoError(t, fake.bot(t).PostListings("c1", ls, "Dancing Mad (Ultimate)", "Aether"))

	posts := fake.posts("c1")
	assert.Contains(t, embedOf(posts[0])["description"], "Found 1 listings")
	fields := embedOf(posts[1])["fields"].([]any)
	assert.Equal(t, "Creator fresh", fields[0].(map[string]any)["name"])
}

func TestPostListingsWithNoListingsSendsOnlyTheHeader(t *testing.T) {
	fake := &fakeDiscord{}

	require.NoError(t, fake.bot(t).PostListings("c1", &ffxiv.Listings{}, "Dancing Mad (Ultimate)", "Aether"))

	posts := fake.posts("c1")
	require.Len(t, posts, 1)
	assert.Contains(t, embedOf(posts[0])["description"], "Found 0 listings")
}

func TestPostListingsRejectsUnparseableUpdateTimes(t *testing.T) {
	fake := &fakeDiscord{}
	ls := &ffxiv.Listings{Listings: []*ffxiv.Listing{listing("1", "yesterday-ish")}}

	err := fake.bot(t).PostListings("c1", ls, "Dancing Mad (Ultimate)", "Aether")

	assert.ErrorContains(t, err, "could not find most recently updated duty")
	assert.Empty(t, fake.requests)
}

func TestPostTokensListsAtMost75Tokens(t *testing.T) {
	fake := &fakeDiscord{}
	tokens := []tokenizer.Token{}
	for i := 0; i < 80; i++ {
		tokens = append(tokens, tokenizer.Token{String: fmt.Sprintf("t%d", i), Count: 80 - i})
	}

	require.NoError(t, fake.bot(t).PostTokens("c1", tokens, 7, 1234))

	content := fake.posts("c1")[0].JSON["content"].(string)
	assert.True(t, strings.HasPrefix(content, "**Last 7 days of Tokens (1234 listings scanned) as of "))
	assert.Contains(t, content, "`t0` 80\n")
	assert.Contains(t, content, "`t74` 6\n")
	assert.NotContains(t, content, "`t75`")
}

func TestPostDescriptionCsvAttachesTheCsv(t *testing.T) {
	fake := &fakeDiscord{}

	require.NoError(t, fake.bot(t).PostDescriptionCsv("c1", bytes.NewBufferString("Timestamp,Description\n2026-01-01,prog\n")))

	post := fake.posts("c1")[0]
	assert.True(t, strings.HasPrefix(post.ContentType, "multipart/form-data"))
	assert.Contains(t, post.Raw, `filename="pf_descriptions_`)
	assert.Contains(t, post.Raw, "Timestamp,Description\n2026-01-01,prog\n")
	assert.Contains(t, post.Raw, "**Description List ")
}

func TestPostErrorsAreReturned(t *testing.T) {
	failing := &Discord{Session: (&fakeDiscord{}).bot(t).Session}
	failing.Session.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("network down")
	})}

	assert.ErrorContains(t, failing.PostTokens("c1", nil, 1, 0), "network down")
	assert.ErrorContains(t, failing.PostListings("c1", &ffxiv.Listings{}, "d", "dc"), "could not send header")
	assert.ErrorContains(t, failing.CleanChannel("c1"), "could not list messages")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
