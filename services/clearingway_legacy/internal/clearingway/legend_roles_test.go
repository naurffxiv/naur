package clearingway

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearsOf returns opts where the first n ultimates are cleared once.
func clearsOf(names ...string) *ShouldApplyOpts {
	byID := map[int]*fflogs.Ranking{}
	for i, name := range names {
		enc := UltimateEncounters.ForName(name)
		byID[enc.Ids[0]] = cleared(1, dps(1700000000000+i*1000, "Samurai", 50))
	}
	return &ShouldApplyOpts{Encounters: UltimateEncounters, Rankings: rankings(byID)}
}

const (
	ucob = "The Unending Coil of Bahamut (Ultimate)"
	uwu  = "The Weapon's Refrain (Ultimate)"
	tea  = "The Epic of Alexander (Ultimate)"
	dsrN = "Dragonsong's Reprise (Ultimate)"
	topN = "The Omega Protocol (Ultimate)"
	fru  = "Futures Rewritten (Ultimate)"
	dm   = "Dancing Mad (Ultimate)"
)

func TestLegendRolesCountClearedUltimates(t *testing.T) {
	all := []string{ucob, uwu, tea, dsrN, topN, fru, dm}
	want := []string{"The Legend", "The Double Legend", "The Triple Legend", "The Quad Legend", "The Penta Legend", "The Sexa Legend", "The Septuple Legend"}
	for n := 1; n <= 7; n++ {
		got := matching(LegendRoles(), clearsOf(all[:n]...))
		assert.Contains(t, got, want[n-1], "%d clears", n)
	}
	assert.Empty(t, matching(LegendRoles(), clearsOf()))
}

func TestLegendRolesForSpecificUltimates(t *testing.T) {
	assert.Equal(t, []string{"The Double Legend", "The Man"}, matching(LegendRoles(), clearsOf(ucob, uwu)))
	assert.Equal(t, []string{"The Triple Legend", "The Man", "The Myth"}, matching(LegendRoles(), clearsOf(ucob, uwu, tea)))
	assert.Equal(t, []string{"The Double Legend"}, matching(LegendRoles(), clearsOf(dsrN, topN)))
	assert.Equal(t, []string{"The Penta Legend"}, matching(LegendRoles(), clearsOf(tea, dsrN, topN, fru, dm)))
	assert.Equal(t, []string{"The Sexa Legend"}, matching(LegendRoles(), clearsOf(uwu, tea, dsrN, topN, fru, dm)))
}

func TestTheLegendMessageListsTheClear(t *testing.T) {
	_, msg := applies(t, LegendRoles(), "The Legend", clearsOf(topN))

	assert.Equal(t, "Cleared the following one Ultimate fight:\n     `The Omega Protocol (Ultimate)` with `SAM` on <t:1700000000:F> (https://www.fflogs.com/reports/r#fight=1700000000000).", msg)
}

func TestTheUncoiledLegendMessageListsAllFiveUltimates(t *testing.T) {
	ok, msg := applies(t, LegendRoles(), "The Uncoiled Legend", clearsOf(ucob, uwu, tea, dsrN, topN))

	require.True(t, ok)
	for _, name := range []string{ucob, uwu, tea, dsrN, topN} {
		assert.Contains(t, msg, "`"+name+"`")
	}
}
