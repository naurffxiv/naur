package clearingway

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/stretchr/testify/assert"
)

// flexScenarios describe one cleared TOP ranking and the flex role it earns,
// named once for the relevant set and once for the ultimate set.
var flexScenarios = []struct {
	name     string
	ranks    []*fflogs.Rank
	relevant string
	ultimate string
}{
	{"zero dps parse", []*fflogs.Rank{dps(1, "Samurai", 0.4)}, "NA's Comfiest", "The Comfy Legend"},
	{"69 dps parse", []*fflogs.Rank{dps(1, "Samurai", 69.5)}, "Nice", "The Nice Legend"},
	{"zero hps parse as healer", []*fflogs.Rank{hps(1, "Sage", 0.5)}, "Chad", "The Chadding Legend"},
	{"100 hps parse as non-healer", []*fflogs.Rank{hps(1, "Paladin", 100)}, "Bloodbather", "The Bloodbathing Legend"},
	{"100 hps parse as healer", []*fflogs.Rank{hps(1, "WhiteMage", 100)}, "Overhealer", "The Overhealing Legend"},
	{"every color but gold", []*fflogs.Rank{
		dps(1, "Samurai", 99.5), dps(2, "Samurai", 96), dps(3, "Samurai", 80),
		dps(4, "Samurai", 60), dps(5, "Samurai", 30), dps(6, "Samurai", 10),
	}, "Rainbow", "The Rainbow Legend"},
}

func TestFlexingRoles(t *testing.T) {
	for _, s := range flexScenarios {
		t.Run(s.name, func(t *testing.T) {
			opts := &ShouldApplyOpts{
				Encounters: &Encounters{Encounters: []*Encounter{top}},
				Rankings:   rankings(map[int]*fflogs.Ranking{1068: cleared(1, s.ranks...)}),
			}
			assert.Equal(t, []string{s.relevant}, matching(RelevantFlexingRoles(), opts))
			assert.Equal(t, []string{s.ultimate}, matching(UltimateFlexingRoles(), opts))
		})
	}
}

func TestFlexingRolesNeedAClear(t *testing.T) {
	for _, s := range flexScenarios {
		opts := &ShouldApplyOpts{
			Encounters: &Encounters{Encounters: []*Encounter{top}},
			Rankings:   rankings(map[int]*fflogs.Ranking{1068: cleared(0, s.ranks...)}),
		}
		assert.Empty(t, matching(RelevantFlexingRoles(), opts), s.name)
		assert.Empty(t, matching(UltimateFlexingRoles(), opts), s.name)
	}
}

func TestRainbowNeedsEveryColorInOneEncounter(t *testing.T) {
	opts := &ShouldApplyOpts{
		Encounters: &Encounters{Encounters: []*Encounter{top, dsr}},
		Rankings: rankings(map[int]*fflogs.Ranking{
			1068: cleared(1, dps(1, "Samurai", 99.5), dps(2, "Samurai", 96), dps(3, "Samurai", 80)),
			1065: cleared(1, dps(4, "Samurai", 60), dps(5, "Samurai", 30), dps(6, "Samurai", 10)),
		}),
	}

	ok, _ := applies(t, RelevantFlexingRoles(), "Rainbow", opts)

	assert.False(t, ok)
}

func TestComfiestExplainsTheZeroParse(t *testing.T) {
	opts := &ShouldApplyOpts{
		Encounters: &Encounters{Encounters: []*Encounter{top}},
		Rankings:   rankings(map[int]*fflogs.Ranking{1068: cleared(1, dps(1700000000000, "Samurai", 0.4))}),
	}

	_, msg := applies(t, RelevantFlexingRoles(), "NA's Comfiest", opts)

	assert.Equal(t, "Parsed **0** (0.40) with `SAM` in `The Omega Protocol (Ultimate)` on <t:1700000000:F> (https://www.fflogs.com/reports/r#fight=1700000000000).\nUse `/uncomfy` if you don't want this role.", msg)
}
