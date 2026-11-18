package board

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

const duty = "Dragonsong's Reprise (Ultimate)"

// call is one call the Board made on the fake Discord port.
type call struct {
	op      string // "clean" or "send"
	channel string
	message *discordgo.MessageSend
}

type fakeDiscord struct {
	calls    []call
	cleanErr error
	sendErr  func(message *discordgo.MessageSend) error
}

func (f *fakeDiscord) Clean(channelID string) error {
	f.calls = append(f.calls, call{op: "clean", channel: channelID})
	return f.cleanErr
}

func (f *fakeDiscord) Send(channelID string, message *discordgo.MessageSend) error {
	f.calls = append(f.calls, call{op: "send", channel: channelID, message: message})
	if f.sendErr != nil {
		return f.sendErr(message)
	}
	return nil
}

func (f *fakeDiscord) sends() []*discordgo.MessageSend {
	var messages []*discordgo.MessageSend
	for _, c := range f.calls {
		if c.op == "send" {
			messages = append(messages, c.message)
		}
	}
	return messages
}

func channel(dataCentres ...string) *Channel {
	return &Channel{Name: "test", ID: "chan", Duty: duty, DataCentres: dataCentres}
}

func listing(creator string, dataCentre string, updated string) *ffxiv.Listing {
	return &ffxiv.Listing{
		Id:         creator,
		Creator:    creator,
		Duty:       duty,
		DataCentre: dataCentre,
		Updated:    updated,
	}
}

func listingsOf(ls ...*ffxiv.Listing) *ffxiv.Listings {
	return &ffxiv.Listings{Listings: ls}
}

// creators returns the creator of every Listing shown in a listings message.
func creators(message *discordgo.MessageSend) []string {
	var names []string
	fields := message.Embeds[0].Fields
	for i := 0; i < len(fields); i += 3 {
		names = append(names, fields[i].Name)
	}
	return names
}

func TestRefreshCleansChannelBeforeSending(t *testing.T) {
	fake := &fakeDiscord{}

	err := New(fake).Refresh(channel("Aether", "Primal"), listingsOf(listing("a", "Aether", "now")), now)

	require.NoError(t, err)
	require.NotEmpty(t, fake.calls)
	assert.Equal(t, call{op: "clean", channel: "chan"}, fake.calls[0])
	for _, c := range fake.calls[1:] {
		assert.Equal(t, "send", c.op)
		assert.Equal(t, "chan", c.channel)
	}
}

func TestRefreshKeepsSendingWhenCleanFails(t *testing.T) {
	cleanErr := errors.New("clean failed")
	fake := &fakeDiscord{cleanErr: cleanErr}

	err := New(fake).Refresh(channel("Aether"), listingsOf(listing("a", "Aether", "now")), now)

	assert.ErrorIs(t, err, cleanErr)
	assert.Len(t, fake.sends(), 2)
}

func TestRefreshShowsOnlyFreshListingsWhenAnyAreFresh(t *testing.T) {
	rows := []struct {
		listing *ffxiv.Listing
		wants   time.Time
		fresh   bool
	}{
		{listing: listing("one-second", "Aether", "a second ago"), wants: now.Add(-1 * time.Second), fresh: true},
		{listing: listing("one-minute", "Aether", "a minute ago"), wants: now.Add(-1 * time.Minute), fresh: true},
		{listing: listing("one-hour", "Aether", "an hour ago"), wants: now.Add(-1 * time.Hour)},
		{listing: listing("37-seconds", "Aether", "37 seconds ago"), wants: now.Add(-37 * time.Second), fresh: true},
		{listing: listing("22-minutes", "Aether", "22 minutes ago"), wants: now.Add(-22 * time.Minute)},
		{listing: listing("2-hours", "Aether", "2 hours ago"), wants: now.Add(-2 * time.Hour)},
	}

	var all ffxiv.Listings
	var wantShown []string
	for _, row := range rows {
		updatedAt, err := row.listing.UpdatedAt(now)
		require.NoError(t, err)
		assert.Equal(t, row.wants, updatedAt, row.listing.Updated)

		all.Listings = append(all.Listings, row.listing)
		if row.fresh {
			wantShown = append(wantShown, row.listing.Creator)
		}
	}

	fake := &fakeDiscord{}
	err := New(fake).Refresh(channel("Aether"), &all, now)

	require.NoError(t, err)
	sends := fake.sends()
	require.Len(t, sends, 2)
	assert.Equal(t, fmt.Sprintf("Found 3 listings <t:%d:R>", now.Unix()), sends[0].Embeds[0].Description)
	assert.Equal(t, wantShown, creators(sends[1]))
}

func TestRefreshShowsAllListingsWhenNoneAreFresh(t *testing.T) {
	fake := &fakeDiscord{}
	listings := listingsOf(
		listing("stale-a", "Aether", "4 minutes ago"),
		listing("stale-b", "Aether", "22 minutes ago"),
		listing("stale-c", "Aether", "2 hours ago"),
	)

	err := New(fake).Refresh(channel("Aether"), listings, now)

	require.NoError(t, err)
	sends := fake.sends()
	require.Len(t, sends, 2)
	assert.Equal(t, []string{"stale-a", "stale-b", "stale-c"}, creators(sends[1]))
}

func TestRefreshBatchesFiveListingsPerEmbed(t *testing.T) {
	for _, tt := range []struct {
		count     int
		wantSizes []int
	}{
		{count: 0, wantSizes: nil},
		{count: 4, wantSizes: []int{4}},
		{count: 5, wantSizes: []int{5}},
		{count: 6, wantSizes: []int{5, 1}},
		{count: 12, wantSizes: []int{5, 5, 2}},
	} {
		t.Run(fmt.Sprint(tt.count), func(t *testing.T) {
			var listings ffxiv.Listings
			for i := range tt.count {
				listings.Listings = append(listings.Listings, listing(fmt.Sprintf("c%02d", i), "Aether", "now"))
			}
			fake := &fakeDiscord{}

			err := New(fake).Refresh(channel("Aether"), &listings, now)

			require.NoError(t, err)
			sends := fake.sends()
			require.Len(t, sends, 1+len(tt.wantSizes))
			var gotSizes []int
			var gotCreators []string
			for _, message := range sends[1:] {
				gotSizes = append(gotSizes, len(creators(message)))
				gotCreators = append(gotCreators, creators(message)...)
			}
			assert.Equal(t, tt.wantSizes, gotSizes)
			for i, creator := range gotCreators {
				assert.Equal(t, fmt.Sprintf("c%02d", i), creator, "listings keep their order")
			}
		})
	}
}

func TestRefreshScopesListingsToDutyAndDataCentre(t *testing.T) {
	otherDuty := listing("other-duty", "Aether", "now")
	otherDuty.Duty = "The Omega Protocol (Ultimate)"
	listings := listingsOf(
		listing("aether-1", "Aether", "now"),
		listing("crystal-1", "Crystal", "now"),
		listing("primal-1", "Primal", "now"),
		listing("aether-2", "Aether", "now"),
		otherDuty,
	)
	fake := &fakeDiscord{}

	err := New(fake).Refresh(channel("Aether", "Primal", "Dynamis"), listings, now)

	require.NoError(t, err)
	sends := fake.sends()
	require.Len(t, sends, 5)

	assert.Equal(t, duty+" PFs (Aether)", sends[0].Embeds[0].Title)
	assert.Equal(t, fmt.Sprintf("Found 2 listings <t:%d:R>", now.Unix()), sends[0].Embeds[0].Description)
	assert.Equal(t, []string{"aether-1", "aether-2"}, creators(sends[1]))

	assert.Equal(t, duty+" PFs (Primal)", sends[2].Embeds[0].Title)
	assert.Equal(t, []string{"primal-1"}, creators(sends[3]))

	assert.Equal(t, duty+" PFs (Dynamis)", sends[4].Embeds[0].Title)
	assert.Equal(t, fmt.Sprintf("Found 0 listings <t:%d:R>", now.Unix()), sends[4].Embeds[0].Description)
}

func TestRefreshReportsErrorsFromEveryDataCentre(t *testing.T) {
	sendErr := errors.New("discord down")
	fake := &fakeDiscord{sendErr: func(message *discordgo.MessageSend) error {
		title := message.Embeds[0].Title
		if strings.Contains(title, "Aether") || strings.Contains(title, "Primal") {
			return sendErr
		}
		return nil
	}}

	err := New(fake).Refresh(channel("Aether", "Crystal", "Primal"), listingsOf(listing("a", "Crystal", "now")), now)

	require.Error(t, err)
	assert.ErrorIs(t, err, sendErr)
	assert.Contains(t, err.Error(), "Aether: could not send header")
	assert.Contains(t, err.Error(), "Primal: could not send header")
	assert.NotContains(t, err.Error(), "Crystal")

	var titles []string
	for _, message := range fake.sends() {
		if title := message.Embeds[0].Title; title != "" {
			titles = append(titles, title)
		}
	}
	assert.Equal(t, []string{duty + " PFs (Aether)", duty + " PFs (Crystal)", duty + " PFs (Primal)"}, titles,
		"a failing DataCentre does not stop the next one")
}

func TestRefreshReportsUnparseableListingAndPostsOtherDataCentres(t *testing.T) {
	fake := &fakeDiscord{}
	listings := listingsOf(
		listing("broken", "Aether", "sometime"),
		listing("primal-1", "Primal", "now"),
	)

	err := New(fake).Refresh(channel("Aether", "Primal"), listings, now)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Aether: could not find fresh listings")
	sends := fake.sends()
	require.Len(t, sends, 2)
	assert.Equal(t, duty+" PFs (Primal)", sends[0].Embeds[0].Title)
}

func TestRefreshBuildsTodaysEmbeds(t *testing.T) {
	l := &ffxiv.Listing{
		Creator:     "Some Body",
		Duty:        duty,
		DataCentre:  "Aether",
		Tags:        "[Practice]",
		Description: "p6 prog",
		Expires:     "in 50 minutes",
		Updated:     "now",
		Party:       []*ffxiv.Slot{{Filled: true, Job: ffxiv.AST}, ffxiv.NewSlot()},
	}
	fake := &fakeDiscord{}

	err := New(fake).Refresh(channel("Aether"), listingsOf(l), now)

	require.NoError(t, err)
	sends := fake.sends()
	require.Len(t, sends, 2)
	footer := &discordgo.MessageEmbedFooter{Text: strings.Repeat("　", 20)}

	assert.Equal(t, &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{
		Title:       duty + " PFs (Aether)",
		Type:        discordgo.EmbedTypeRich,
		Color:       0x6600ff,
		Description: fmt.Sprintf("Found 1 listings <t:%d:R>", now.Unix()),
		Footer:      footer,
	}}}, sends[0])

	assert.Equal(t, &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{
		Type:  discordgo.EmbedTypeRich,
		Color: 0x6600ff,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Some Body", Value: l.PartyDisplay(), Inline: true},
			{Name: "[Practice]", Value: "```p6 prog```", Inline: true},
			{Name: "<:hourglass:991379574187372655> in 50 minutes", Value: "<:stopwatch:991379573000388758> now", Inline: true},
		},
		Footer: footer,
	}}}, sends[1])
}
