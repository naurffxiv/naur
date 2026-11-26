package clearingway

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/stretchr/testify/assert"
)

func killsOpts(enc *Encounter, killsByID map[int]int) *ShouldApplyOpts {
	byID := map[int]*fflogs.Ranking{}
	for id, kills := range killsByID {
		byID[id] = cleared(kills)
	}
	return &ShouldApplyOpts{Encounters: &Encounters{Encounters: []*Encounter{enc}}, Rankings: rankings(byID)}
}

func TestRelevantRepetitionRoles(t *testing.T) {
	enc := encounter("The Omega Protocol (Ultimate)", 1068, 1077)
	enc.TotalWeaponsAvailable = 19
	roles := RelevantRepetitionRoles(&Encounters{Encounters: []*Encounter{enc}})

	assert.Equal(t, []string{"Limbo"}, matching(roles, killsOpts(enc, map[int]int{1068: 1})))
	assert.Empty(t, matching(roles, killsOpts(enc, map[int]int{1068: 18})))
	assert.Equal(t, []string{"Complete"}, matching(roles, killsOpts(enc, map[int]int{1068: 10, 1077: 9})))
	assert.Equal(t, []string{"Please Do Other Content", "Complete"}, matching(roles, killsOpts(enc, map[int]int{1068: 60, 1077: 40})))
}

func TestRelevantRepetitionRolesOnlyCountUltimatesForPleaseDoOtherContent(t *testing.T) {
	savage := &Encounter{Name: "AAC Heavyweight M4 (Savage)", Ids: []int{100}, Difficulty: "Savage"}
	roles := RelevantRepetitionRoles(&Encounters{Encounters: []*Encounter{savage}})

	assert.Len(t, roles.Roles, 1, "encounters without weapons get no Limbo or Complete role")
	assert.Empty(t, matching(roles, killsOpts(savage, map[int]int{100: 500})))
}

func TestUltimateRepetitionRoles(t *testing.T) {
	roles := UltimateRepetitionRoles()
	topEnc := UltimateEncounters.ForName("The Omega Protocol (Ultimate)")

	assert.Len(t, roles.Roles, 2*len(UltimateEncounters.Encounters))
	assert.Equal(t, []string{"Alpha Limbo"}, matching(roles, killsOpts(topEnc, map[int]int{1068: 1})))
	assert.Equal(t, []string{"Alpha"}, matching(roles, killsOpts(topEnc, map[int]int{1068: 19})))
	assert.Empty(t, matching(roles, killsOpts(topEnc, map[int]int{1068: 2})))
}
