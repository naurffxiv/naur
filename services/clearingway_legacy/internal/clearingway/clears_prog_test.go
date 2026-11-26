package clearingway

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/Veraticus/clearingway/internal/ffxiv"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// externalSites answers fflogs and Lodestone requests made through http.DefaultTransport,
// which both the fflogs client and the Lodestone scraper use.
type externalSites struct {
	discord      *fakeDiscord
	graphql      func(query string) string // returns the GraphQL "data" payload
	lodestoneBio string
}

func (e *externalSites) RoundTrip(r *http.Request) (*http.Response, error) {
	respond := func(contentType, body string) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{contentType}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	}
	switch {
	case r.URL.Host == "discord.com":
		return e.discord.RoundTrip(r)
	case r.URL.Path == "/oauth/token":
		return respond("application/json", `{"access_token":"t","token_type":"Bearer","expires_in":3600}`)
	case r.URL.Path == "/api/v2/client":
		raw, _ := io.ReadAll(r.Body)
		return respond("application/json", `{"data":`+e.graphql(string(raw))+`}`)
	case strings.HasPrefix(r.URL.Path, "/lodestone/character/"):
		return respond("text/html", `<html><body><div class="character__content selected"><div class="character__selfintroduction">`+e.lodestoneBio+`</div></div></body></html>`)
	}
	return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(&bytes.Buffer{}), Request: r}, nil
}

func useExternalSites(t *testing.T, sites *externalSites) {
	t.Helper()
	old := http.DefaultTransport
	http.DefaultTransport = sites
	t.Cleanup(func() { http.DefaultTransport = old })
}

// lodestoneIDOr answers the lodestone-ID lookup and delegates every other query.
func lodestoneIDOr(other func(string) string) func(string) string {
	return func(query string) string {
		if strings.Contains(query, "lodestoneID") {
			return `{"characterData":{"character":{"lodestoneID":42}}}`
		}
		return other(query)
	}
}

func ownerBio() string {
	return (&ffxiv.Character{}).LodestoneSlug("u1")
}

func topGuild() *ConfigGuild {
	return &ConfigGuild{
		Name:             "Test",
		ConfigRoles:      &ConfigRoles{RelevantParsing: true},
		ConfigEncounters: []*ConfigEncounter{{Name: "TOP", Ids: []int{1068}, Difficulty: "Ultimate", ConfigRoles: []*ConfigRole{{Type: "Cleared", Name: "cleared top"}}}},
	}
}

func TestClearsGivesEarnedRolesAndRemovesStaleOnes(t *testing.T) {
	fake := &fakeDiscord{memberRoles: []string{"Pink"}}
	useExternalSites(t, &externalSites{discord: fake, lodestoneBio: "hi " + ownerBio(), graphql: lodestoneIDOr(func(string) string {
		return `{"characterData":{"character":{
			"rdpsZ1068Pstandard":{"totalKills":0,"metric":"rdps","partition":1,"ranks":[{"rankPercent":100,"spec":"Samurai","startTime":1700000000000,"report":{"code":"r","fightID":1}}]},
			"hpsZ1068Pstandard":{"totalKills":2,"metric":"hps","partition":1,"ranks":[]}
		}}}`
	})})
	c, s := botWithGuild(t, fake, topGuild())
	c.Fflogs = fflogs.Init("id", "secret")

	c.Clears(s, command("clears", nil, stringOption("world", "gilgamesh"), stringOption("first-name", "alex"), stringOption("last-name", "smith")))

	assert.ElementsMatch(t, []string{"+cleared top", "+Gold", "-Pink"}, fake.roleChanges())
	replies := strings.Join(fake.replies(), "\n")
	assert.Contains(t, replies, "Finding `Alex Smith (Gilgamesh)` in the Lodestone...")
	assert.Contains(t, replies, "Finished analysis for `Alex Smith (Gilgamesh)`.")
	assert.Contains(t, replies, "__Adding role: **cleared top**__")
	assert.Contains(t, replies, "__Removing role: **Pink**__")
}

func TestClearsAsksForProofOfOwnership(t *testing.T) {
	fake := &fakeDiscord{}
	useExternalSites(t, &externalSites{discord: fake, lodestoneBio: "not mine", graphql: lodestoneIDOr(func(string) string { return `{}` })})
	c, s := botWithGuild(t, fake, topGuild())
	c.Fflogs = fflogs.Init("id", "secret")

	c.Clears(s, command("clears", nil, stringOption("world", "Gilgamesh"), stringOption("first-name", "Alex"), stringOption("last-name", "Smith")))

	assert.Empty(t, fake.roleChanges())
	last := fake.replies()[len(fake.replies())-1]
	assert.Contains(t, last, "I could not verify your ownership of `Alex Smith (Gilgamesh)`!")
	assert.Contains(t, last, "**"+ownerBio()+"**")
}

func TestClearsAndProgRejectBadInput(t *testing.T) {
	tests := []struct {
		name    string
		command string
		options map[string]string
		want    string
	}{
		{"clears missing name", "clears", map[string]string{"world": "Gilgamesh"}, "`/clears` command failed! Please input your world, first name, and last name."},
		{"clears bad world", "clears", map[string]string{"world": "Atlantis", "first-name": "Alex", "last-name": "Smith"}, "`Atlantis` is not a valid world! Make sure you spelled your world name properly."},
		{"prog missing report", "prog", map[string]string{"world": "Gilgamesh", "first-name": "Alex", "last-name": "Smith"}, "`/prog` command failed! Make sure you input your world, first name, last name, and fflogs report URL or ID."},
		{"prog bad world", "prog", map[string]string{"world": "Atlantis", "first-name": "Alex", "last-name": "Smith", "report-id": "abc"}, "`Atlantis` is not a valid world! Make sure you spelled your world name properly."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeDiscord{}
			c, s := botWithGuild(t, fake, topGuild())
			options := []*discordgo.ApplicationCommandInteractionDataOption{}
			for name, value := range tt.options {
				options = append(options, stringOption(name, value))
			}
			i := command(tt.command, nil, options...)

			map[string]func(*discordgo.Session, *discordgo.InteractionCreate){"clears": c.Clears, "prog": c.Prog}[tt.command](s, i)

			replies := fake.replies()
			require.Len(t, replies, 2)
			assert.Equal(t, tt.want, replies[1])
			assert.Empty(t, fake.roleChanges())
		})
	}
}

func TestProgGivesTheFurthestProgRole(t *testing.T) {
	fake := &fakeDiscord{memberRoles: []string{"P1"}}
	useExternalSites(t, &externalSites{discord: fake, lodestoneBio: ownerBio(), graphql: lodestoneIDOr(func(string) string {
		return `{"reportData":{"report":{
			"fights":[{"id":3,"kill":false,"difficulty":100,"encounterID":1068,"lastPhaseAsAbsoluteIndex":1,"friendlyPlayers":[10]}],
			"masterData":{"actors":[{"id":10,"name":"Alex Smith","server":"Gilgamesh"}]}
		}}}`
	})})
	config := topGuild()
	config.ConfigEncounters[0].ConfigProg = []*ConfigRole{{Name: "P1"}, {Name: "P2"}, {Name: "P3"}}
	c, s := botWithGuild(t, fake, config)
	c.Fflogs = fflogs.Init("id", "secret")

	c.Prog(s, command("prog", nil,
		stringOption("world", "Gilgamesh"), stringOption("first-name", "Alex"), stringOption("last-name", "Smith"),
		stringOption("report-id", "https://www.fflogs.com/reports/abc#fight=3")))

	assert.Equal(t, []string{"+P2", "-P1"}, fake.roleChanges())
	assert.Contains(t, strings.Join(fake.replies(), "\n"), "Your furthest prog point is now `P2` (Phase **2**).")
}
