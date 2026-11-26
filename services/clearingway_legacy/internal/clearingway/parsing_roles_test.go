package clearingway

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/stretchr/testify/assert"
)

func TestRelevantParsingRolesPickTheColorOfTheBestClearedParse(t *testing.T) {
	tests := map[float64][]string{
		100:   {"Gold"},
		99.99: {"Pink"},
		99:    {"Pink"},
		98.9:  {"Orange"},
		95:    {"Orange"},
		75:    {"Purple"},
		50:    {"Blue"},
		25:    {"Green"},
		24.9:  {"Gray"},
		0:     {},
	}
	for percent, want := range tests {
		opts := &ShouldApplyOpts{
			Encounters: &Encounters{Encounters: []*Encounter{top}},
			Rankings:   rankings(map[int]*fflogs.Ranking{1068: cleared(1, dps(1000, "Samurai", percent), dps(2000, "Samurai", 0))}),
		}
		assert.Equal(t, want, matching(RelevantParsingRoles(), opts), "best parse %v", percent)
	}
}

func TestRelevantParsingRolesIgnoreUnclearedEncounters(t *testing.T) {
	opts := &ShouldApplyOpts{
		Encounters: &Encounters{Encounters: []*Encounter{top}},
		Rankings:   rankings(map[int]*fflogs.Ranking{1068: cleared(0, dps(1000, "Samurai", 100))}),
	}

	ok, msg := applies(t, RelevantParsingRoles(), "Gold", opts)

	assert.False(t, ok)
	assert.Equal(t, "No encounter or rank found.", msg)
}

func TestRelevantParsingRolesExplainTheBestParse(t *testing.T) {
	opts := &ShouldApplyOpts{
		Encounters: &Encounters{Encounters: []*Encounter{top, dsr}},
		Rankings: rankings(map[int]*fflogs.Ranking{
			1068: cleared(1, dps(1700000000000, "Samurai", 96)),
			1065: cleared(1, dps(1700000001000, "Reaper", 100)),
		}),
	}

	ok, msg := applies(t, RelevantParsingRoles(), "Gold", opts)

	assert.True(t, ok)
	assert.Equal(t, "Best parse was **100.00** with `RPR` in `Dragonsong's Reprise (Ultimate)` on <t:1700000001:F> (https://www.fflogs.com/reports/r#fight=1700000001000).\nUse `/uncolor` if you don't want this role.", msg)
}
