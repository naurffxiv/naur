package fflogs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Veraticus/clearingway/internal/ffxiv"
	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeFflogs serves a fixed GraphQL "data" payload and records the last query.
func fakeFflogs(t *testing.T, data string) (*Fflogs, *string) {
	t.Helper()
	var lastQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		lastQuery = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":`+data+`}`)
	}))
	t.Cleanup(ts.Close)
	return &Fflogs{graphqlClient: graphql.NewClient(ts.URL, ts.Client())}, &lastQuery
}

func gilgameshChar() *ffxiv.Character {
	return &ffxiv.Character{FirstName: "Alex", LastName: "Smith", World: "Gilgamesh"}
}

func TestSetCharacterLodestoneID(t *testing.T) {
	f, query := fakeFflogs(t, `{"characterData":{"character":{"lodestoneID":12345}}}`)
	char := gilgameshChar()

	require.NoError(t, f.SetCharacterLodestoneID(char))

	assert.Equal(t, 12345, char.LodestoneID)
	assert.Contains(t, *query, `name: \"Alex Smith\", serverSlug: \"Gilgamesh\", serverRegion: \"NA\"`)
}

func TestSetCharacterLodestoneIDCharacterNotFound(t *testing.T) {
	f, _ := fakeFflogs(t, `{"characterData":{"character":null}}`)

	err := f.SetCharacterLodestoneID(gilgameshChar())

	assert.EqualError(t, err, "character Alex Smith (Gilgamesh) not found in fflogs")
}

func TestGetRankingsForCharacter(t *testing.T) {
	f, query := fakeFflogs(t, `{"characterData":{"character":{
		"rdpsZ1068Pstandard":{"totalKills":2,"metric":"rdps","partition":1,"ranks":[{"rankPercent":75,"spec":"Samurai","startTime":1000,"report":{"code":"r1","fightID":3}}]},
		"hpsZ1068Pstandard":{"totalKills":2,"metric":"hps","partition":1,"ranks":[{"rankPercent":5,"spec":"Samurai","startTime":1000,"report":{"code":"r1","fightID":3}}]},
		"rdpsZ1068Pnonstandard":{"totalKills":0,"metric":"rdps","partition":1,"ranks":[]},
		"hpsZ1068Pnonstandard":null,
		"rdpsZ9999Pstandard":{"error":"Invalid encounter id specified."}
	}}}`)

	rankings, err := f.GetRankingsForCharacter([]*RankingToGet{{IDs: []int{1068, 9999}, Difficulty: 100}}, gilgameshChar())
	require.NoError(t, err)

	assert.Contains(t, *query, "rdpsZ1068Pstandard:encounterRankings(encounterID: 1068, difficulty: 100, metric: rdps)")
	require.Contains(t, rankings.Rankings, 1068)
	assert.NotContains(t, rankings.Rankings, 9999)

	r := rankings.Rankings[1068]
	assert.Equal(t, 2, r.TotalKills)
	require.Len(t, r.Ranks, 1)
	assert.InDelta(t, 75, r.Ranks[0].DPSPercent, 0.001)
	assert.InDelta(t, 5, r.Ranks[0].HPSPercent, 0.001)
}

func TestGetRankingsForCharacterReturnsFflogsErrors(t *testing.T) {
	f, _ := fakeFflogs(t, `{"characterData":{"character":{"rdpsZ1Pstandard":{"error":"Something broke"}}}}`)

	_, err := f.GetRankingsForCharacter([]*RankingToGet{{IDs: []int{1}, Difficulty: 100}}, gilgameshChar())

	assert.EqualError(t, err, "received error from fflogs for encounter 1: Something broke")
}

func TestGetProgForReportKeepsMatchingFightsWithTheCharacter(t *testing.T) {
	f, query := fakeFflogs(t, `{"reportData":{"report":{
		"fights":[
			{"id":1,"kill":false,"difficulty":100,"encounterID":1068,"lastPhaseAsAbsoluteIndex":3,"friendlyPlayers":[10,11]},
			{"id":2,"kill":false,"difficulty":100,"encounterID":1068,"lastPhaseAsAbsoluteIndex":4,"friendlyPlayers":[11]},
			{"id":3,"kill":true,"difficulty":101,"encounterID":1068,"lastPhaseAsAbsoluteIndex":5,"friendlyPlayers":[10]},
			{"id":4,"kill":true,"difficulty":100,"encounterID":2000,"lastPhaseAsAbsoluteIndex":5,"friendlyPlayers":[10]}
		],
		"masterData":{"actors":[{"id":10,"name":"Alex Smith","server":"Gilgamesh"},{"id":11,"name":"Other","server":"Gilgamesh"}]}
	}}}`)

	fights, err := f.GetProgForReport("abc", []*RankingToGet{{IDs: []int{1068}, Difficulty: 100}}, gilgameshChar())
	require.NoError(t, err)

	assert.Contains(t, *query, `report(code: \"abc\")`)
	require.Len(t, fights.Fights, 1)
	assert.Equal(t, Fight{ID: 1, EncounterID: 1068, LastPhaseIndex: 3, ReportID: "abc"}, *fights.Fights[0])
}

func TestGetProgForReportCharacterNotInReport(t *testing.T) {
	f, _ := fakeFflogs(t, `{"reportData":{"report":{"fights":[],"masterData":{"actors":[{"id":1,"name":"Alex Smith","server":"Siren"}]}}}}`)

	_, err := f.GetProgForReport("abc", nil, gilgameshChar())

	assert.EqualError(t, err, "could not find character Alex Smith (Gilgamesh) in report abc")
}
