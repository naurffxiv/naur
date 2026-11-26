package ffxiv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func listingsFixture() *Listings {
	return &Listings{Listings: []*Listing{
		{Id: "1", Duty: "Dancing Mad (Ultimate)", DataCentre: "Aether"},
		{Id: "2", Duty: "Dancing Mad (Ultimate)", DataCentre: "Light"},
		{Id: "3", Duty: "The Omega Protocol (Ultimate)", DataCentre: "Aether"},
		{Id: "4", Duty: "Futures Rewritten (Ultimate)", DataCentre: "Primal"},
	}}
}

func ids(ls *Listings) []string {
	out := []string{}
	for _, l := range ls.Listings {
		out = append(out, l.Id)
	}
	return out
}

func TestForDutyAndDataCentre(t *testing.T) {
	assert.Equal(t, []string{"1"}, ids(listingsFixture().ForDutyAndDataCentre("Dancing Mad (Ultimate)", "Aether")))
	assert.Empty(t, ids(listingsFixture().ForDutyAndDataCentre("Dancing Mad (Ultimate)", "Primal")))
}

func TestForDutiesAndDataCentres(t *testing.T) {
	got := listingsFixture().ForDutiesAndDataCentres(
		[]string{"Dancing Mad (Ultimate)", "Futures Rewritten (Ultimate)"},
		[]string{"Aether", "Primal"},
	)

	assert.Equal(t, []string{"1", "4"}, ids(got))
}

func TestListingsAddIgnoresDuplicateIDs(t *testing.T) {
	ls := &Listings{}

	ls.Add(&Listing{Id: "1", Description: "first"})
	ls.Add(&Listing{Id: "1", Description: "second"})
	ls.Add(&Listing{Id: "2"})

	assert.Equal(t, []string{"1", "2"}, ids(ls))
	assert.Equal(t, "first", ls.Listings[0].Description)
}

func TestPartyDisplayShowsJobsForFilledSlotsAndRolesForOpenOnes(t *testing.T) {
	l := &Listing{Party: []*Slot{
		{Filled: true, Job: SAM},
		{Roles: Roles{Roles: []Role{Healer}}},
		{Roles: Roles{Roles: []Role{DPS, Tank}}},
	}}

	assert.Equal(t, "<:SAM:1277726922419474482> <:healer:1312855996791656449> <:tankdps:1312855993742262424> ", l.PartyDisplay())
}

func TestListingDisplayFields(t *testing.T) {
	l := &Listing{Expires: "in 5 minutes", Updated: "now", Description: "prog p5"}

	assert.Equal(t, "<:hourglass:991379574187372655> in 5 minutes", l.GetExpires())
	assert.Equal(t, "<:stopwatch:991379573000388758> now", l.GetUpdated())
	assert.Equal(t, "```prog p5```", l.GetDescription())
	assert.Equal(t, "_ _", l.GetTags(), "empty tags still render as a blank line")
	assert.Equal(t, "[Duty Complete]", (&Listing{Tags: "[Duty Complete]"}).GetTags())
}

func TestNewSlotStartsWithNoRoles(t *testing.T) {
	s := NewSlot()

	assert.Empty(t, s.Roles.Roles)
	assert.False(t, s.Filled)
}

func TestJobFromAbbreviation(t *testing.T) {
	for abbreviation, want := range map[string]Job{"GNB": GNB, "SGE": SGE, "AST": AST, "VPR": VPR, "PCT": PCT, "ARC": ARC, "THM": THM} {
		assert.Equal(t, want, JobFromAbbreviation(abbreviation), abbreviation)
	}
	assert.Equal(t, Unknown, JobFromAbbreviation("FSH"))
	assert.Equal(t, Unknown, JobFromAbbreviation("gnb"))
}

func TestEveryJobHasItsOwnEmoji(t *testing.T) {
	seen := map[string]Job{}
	for j := GNB; j < Unknown; j++ {
		emoji := j.Emoji()
		assert.NotEqual(t, Unknown.Emoji(), emoji, "job %d falls back to the unknown emoji", j)
		assert.NotContains(t, seen, emoji, "job %d reuses job %d's emoji", j, seen[emoji])
		seen[emoji] = j
	}
}

func TestRolesEmoji(t *testing.T) {
	tests := []struct {
		roles []Role
		want  string
	}{
		{[]Role{DPS}, "<:dps:1312855997697626185>"},
		{[]Role{Healer}, "<:healer:1312855996791656449>"},
		{[]Role{Tank}, "<:tank:1312855995910848573>"},
		{[]Role{DPS, Healer}, "<:healerdps:1312855994971193384>"},
		{[]Role{DPS, Tank}, "<:tankdps:1312855993742262424>"},
		{[]Role{Healer, Tank}, "<:tankhealer:1312855992777576493>"},
		// the scraper always appends roles in DPS, Healer, Tank order
		{[]Role{DPS, Healer, Tank}, "<:tankhealerdps:1312855992198758470>"},
		{[]Role{Empty}, "<:AnyRole:1312855998981214268>"},
		{[]Role{}, "<:AnyRole:1312855998981214268>"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Roles{Roles: tt.roles}.Emoji(), "%v", tt.roles)
	}
}
