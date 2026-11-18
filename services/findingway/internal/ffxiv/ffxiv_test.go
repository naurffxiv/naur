package ffxiv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fixedNow = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

func TestUpdatedAt(t *testing.T) {
	now := fixedNow

	tests := []struct {
		listing *Listing
		wants   time.Time
	}{
		{
			listing: &Listing{Updated: "a second ago"},
			wants:   now.Add(time.Duration(-1) * time.Second),
		},
		{
			listing: &Listing{Updated: "a minute ago"},
			wants:   now.Add(time.Duration(-1) * time.Minute),
		},
		{
			listing: &Listing{Updated: "an hour ago"},
			wants:   now.Add(time.Duration(-1) * time.Hour),
		},
		{
			listing: &Listing{Updated: "37 seconds ago"},
			wants:   now.Add(time.Duration(-37) * time.Second),
		},
		{
			listing: &Listing{Updated: "22 minutes ago"},
			wants:   now.Add(time.Duration(-22) * time.Minute),
		},
		{
			listing: &Listing{Updated: "2 hours ago"},
			wants:   now.Add(time.Duration(-2) * time.Hour),
		},
	}

	for _, tt := range tests {
		updatedAt, err := tt.listing.UpdatedAt(now)
		assert.Nil(t, err)
		assert.Equal(t, tt.wants, updatedAt)
	}
}

func TestUpdatedWithinLast(t *testing.T) {
	expected1 := &Listing{Updated: "a second ago"}
	expected2 := &Listing{Updated: "40 seconds ago"}
	expected3 := &Listing{Updated: "a minute ago"}
	expected4 := &Listing{Updated: "2 minutes ago"}
	bad1 := &Listing{Updated: "5 minutes ago"}
	bad2 := &Listing{Updated: "22 minutes ago"}
	bad3 := &Listing{Updated: "an hour ago"}
	bad4 := &Listing{Updated: "3 hours ago"}

	listings := &Listings{
		Listings: []*Listing{
			bad4,
			expected1,
			bad3,
			expected4,
			expected2,
			bad2,
			bad1,
			expected3,
		},
	}

	mostRecentListings, err := listings.UpdatedWithinLast(3*time.Minute, fixedNow)
	assert.Nil(t, err)
	assert.Equal(t, []*Listing{expected1, expected4, expected2, expected3}, mostRecentListings.Listings)
}

func TestUpdatedAtNowAndBlankMeanNow(t *testing.T) {
	for _, updated := range []string{"", "now"} {
		updatedAt, err := (&Listing{Updated: updated}).UpdatedAt(fixedNow)
		assert.NoError(t, err)
		assert.Equal(t, fixedNow, updatedAt, "updated %q", updated)
	}
}

func TestUpdatedAtRejectsUnknownText(t *testing.T) {
	_, err := (&Listing{Updated: "sometime"}).UpdatedAt(fixedNow)
	assert.ErrorContains(t, err, "failed to parse time sometime")
}

func TestUpdatedWithinLastFailsOnUnparseableListing(t *testing.T) {
	listings := &Listings{Listings: []*Listing{{Updated: "now"}, {Updated: "sometime"}}}

	_, err := listings.UpdatedWithinLast(time.Minute, fixedNow)
	assert.Error(t, err)
}

func TestForDutyAndDataCentreKeepsOnlyMatchingListings(t *testing.T) {
	match := &Listing{Duty: "TOP", DataCentre: "Aether"}
	listings := &Listings{Listings: []*Listing{
		{Duty: "TOP", DataCentre: "Primal"},
		match,
		{Duty: "DSR", DataCentre: "Aether"},
	}}

	assert.Equal(t, []*Listing{match}, listings.ForDutyAndDataCentre("TOP", "Aether").Listings)
	assert.Empty(t, listings.ForDutyAndDataCentre("FRU", "Aether").Listings)
}

func TestForDutiesAndDataCentresKeepsListingsInAnyPair(t *testing.T) {
	topAether := &Listing{Duty: "TOP", DataCentre: "Aether"}
	dsrPrimal := &Listing{Duty: "DSR", DataCentre: "Primal"}
	listings := &Listings{Listings: []*Listing{
		topAether,
		{Duty: "FRU", DataCentre: "Aether"},
		dsrPrimal,
		{Duty: "TOP", DataCentre: "Chaos"},
	}}

	got := listings.ForDutiesAndDataCentres([]string{"TOP", "DSR"}, []string{"Aether", "Primal"})
	assert.Equal(t, []*Listing{topAether, dsrPrimal}, got.Listings)
}

func TestAddIgnoresListingWithSeenId(t *testing.T) {
	first := &Listing{Id: "1", Creator: "first"}
	listings := &Listings{}

	listings.Add(first)
	listings.Add(&Listing{Id: "1", Creator: "duplicate"})
	listings.Add(&Listing{Id: "2"})

	assert.Len(t, listings.Listings, 2)
	assert.Same(t, first, listings.Listings[0])
}

func TestPartyDisplayShowsJobsForFilledSlotsAndRolesForOpenOnes(t *testing.T) {
	listing := &Listing{Party: []*Slot{
		{Filled: true, Job: JobFromAbbreviation("AST")},
		{Roles: Roles{Roles: []Role{Tank}}},
		{Roles: Roles{Roles: []Role{DPS, Healer}}},
		NewSlot(),
	}}

	assert.Equal(t,
		AST.Emoji()+" "+
			"<:tank:1312855995910848573> "+
			"<:healerdps:1312855994971193384> "+
			"<:AnyRole:1312855998981214268> ",
		listing.PartyDisplay())
}

func TestDisplayFieldsDecorateListingText(t *testing.T) {
	listing := &Listing{Expires: "in 5 minutes", Updated: "now", Description: "farm"}

	assert.Equal(t, "<:hourglass:991379574187372655> in 5 minutes", listing.GetExpires())
	assert.Equal(t, "<:stopwatch:991379573000388758> now", listing.GetUpdated())
	assert.Equal(t, "```farm```", listing.GetDescription())
	assert.Equal(t, "_ _", listing.GetTags())

	listing.Tags = "[Practice]"
	assert.Equal(t, "[Practice]", listing.GetTags())
}

func TestJobFromAbbreviation(t *testing.T) {
	assert.Equal(t, AST, JobFromAbbreviation("AST"))
	assert.Equal(t, "<:AST:1277726729695137936>", AST.Emoji())
	assert.Equal(t, PCT, JobFromAbbreviation("PCT"))
	assert.Equal(t, Unknown, JobFromAbbreviation("FSH"))
	assert.Equal(t, "<:FSH:1317911119901233162>", Unknown.Emoji())
}
