package clearingway

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncounterInitDefaultRoles(t *testing.T) {
	e := &Encounter{}
	e.Init(&ConfigEncounter{Name: "AAC M4 (Savage)", Ids: []int{100}, Difficulty: "Savage", DefaultRoles: true})

	assert.Equal(t, 1, e.RequiredKillsToClear)
	assert.Equal(t, 101, e.DifficultyInt())
	want := map[RoleType][2]string{
		PfRole:      {"AAC M4 (Savage)-PF", "Wants to PF AAC M4 (Savage)."},
		ReclearRole: {"AAC M4 (Savage)-Reclear", "Wants to reclear AAC M4 (Savage)."},
		ParseRole:   {"AAC M4 (Savage)-Parse", "Wants to parse AAC M4 (Savage)."},
		ClearedRole: {"AAC M4 (Savage)-Cleared", "Cleared AAC M4 (Savage)."},
	}
	require.Len(t, e.Roles, len(want))
	for roleType, nameAndDescription := range want {
		assert.Equal(t, nameAndDescription[0], e.Roles[roleType].Name)
		assert.Equal(t, nameAndDescription[1], e.Roles[roleType].Description)
	}
}

func TestEncounterInitConfigRolesOverrideDefaults(t *testing.T) {
	e := &Encounter{}
	e.Init(&ConfigEncounter{
		Name: "TOP", Ids: []int{1068}, Difficulty: "Ultimate", DefaultRoles: true, RequiredKillsToClear: 2,
		ConfigRoles: []*ConfigRole{
			{Type: "Cleared", Name: "cleared top", Color: 0x3498db, Hoist: true, Mention: true},
			{Type: "Name Color", Name: "top color"},
		},
	})

	cleared := e.Roles[ClearedRole]
	assert.Equal(t, "cleared top", cleared.Name)
	assert.Equal(t, 0x3498db, cleared.Color)
	assert.True(t, cleared.Hoist)
	assert.True(t, cleared.Mention)
	assert.Equal(t, "Likes TOP's color.", e.Roles[ColorRole].Description)
	assert.Equal(t, 2, e.RequiredKillsToClear)
	assert.Equal(t, 100, e.DifficultyInt())
}

func TestEncounterClearedRoleNeedsRequiredKills(t *testing.T) {
	e := &Encounter{}
	e.Init(&ConfigEncounter{Name: "TOP", Ids: []int{1068, 1077}, DefaultRoles: true, RequiredKillsToClear: 2})
	opts := func(kills int) *ShouldApplyOpts {
		return &ShouldApplyOpts{Rankings: rankings(map[int]*fflogs.Ranking{1077: cleared(kills, dps(1700000000000, "Samurai", 50))})}
	}

	ok, msg := e.Roles[ClearedRole].ShouldApply(opts(1))
	assert.False(t, ok)
	assert.Equal(t, "Has not cleared TOP.", msg)

	ok, msg = e.Roles[ClearedRole].ShouldApply(opts(3))
	assert.True(t, ok)
	assert.Equal(t, "Cleared `TOP` at least **2** times (**3** times total):\n     `SAM` on <t:1700000000:F> (https://www.fflogs.com/reports/r#fight=1700000000000).", msg)
}

func TestEncounterInitBuildsProgRoles(t *testing.T) {
	e := &Encounter{}
	e.Init(&ConfigEncounter{Name: "TOP", DefaultRoles: true, ConfigProg: []*ConfigRole{{Name: "P1"}, {Name: "P2"}}})

	require.NotNil(t, e.ProgRoles)
	assert.Equal(t, "Reached phase 2 (P2) in prog.", e.ProgRoles.Roles[1].Description)
	assert.Len(t, (&Encounters{Encounters: []*Encounter{e}}).Roles().Roles, 6)
}

func TestEncountersClearsAndBestWorstRanks(t *testing.T) {
	es := &Encounters{Encounters: []*Encounter{top, dsr}}
	rs := rankings(map[int]*fflogs.Ranking{
		1077: cleared(1, dps(1, "Samurai", 40), dps(2, "Samurai", 90)),
		1065: cleared(0, dps(3, "Samurai", 99)),
	})

	assert.Equal(t, []string{top.Name}, es.Clears(rs).Names())

	enc, best := es.BestDPSRank(rs)
	assert.Same(t, top, enc)
	assert.InDelta(t, 90, best.DPSPercent, 0.001)

	enc, worst := es.WorstDPSRank(rs)
	assert.Same(t, top, enc)
	assert.InDelta(t, 40, worst.DPSPercent, 0.001)

	enc, best = (&Encounters{}).BestDPSRank(rs)
	assert.Nil(t, enc)
	assert.Nil(t, best)
}

func TestEncountersAddAndForName(t *testing.T) {
	es := &Encounters{}
	es.Add(top)
	es.Add(&Encounter{Name: top.Name})
	es.Add(dsr)

	assert.Equal(t, []string{top.Name, dsr.Name}, es.Names())
	assert.Same(t, dsr, es.ForName(dsr.Name))
	assert.Nil(t, es.ForName("missing"))
}

func TestEncounterRanksAndFightsFilterByEncounterIDs(t *testing.T) {
	rs := rankings(map[int]*fflogs.Ranking{1068: cleared(1, dps(1, "Samurai", 1)), 1077: cleared(1, dps(2, "Samurai", 2)), 1: cleared(1, dps(3, "Samurai", 3))})
	fights := &fflogs.Fights{Fights: []*fflogs.Fight{{ID: 1, EncounterID: 1068}, {ID: 2, EncounterID: 9}}}

	assert.Len(t, top.Ranks(rs), 2)
	got := top.Fights(fights)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].ID)
}

func TestEncounterCompleteNumber(t *testing.T) {
	assert.Equal(t, "19", (&Encounter{TotalWeaponsAvailable: 19}).CompleteNumber())
}
