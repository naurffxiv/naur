package clearingway

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func progEncounter() *Encounter {
	e := &Encounter{}
	e.Init(&ConfigEncounter{Name: "TOP", DefaultRoles: true, ConfigProg: []*ConfigRole{{Name: "P1"}, {Name: "P2"}, {Name: "P3"}}})
	return e
}

func progOpts(existing []*Role, fights ...*fflogs.Fight) *ShouldApplyOpts {
	return &ShouldApplyOpts{Fights: &fflogs.Fights{Fights: fights}, ExistingRoles: &Roles{Roles: existing}}
}

func names(rs []*Role) []string {
	out := []string{}
	for _, r := range rs {
		out = append(out, r.Name)
	}
	return out
}

func TestProgRolesNeedFights(t *testing.T) {
	ok, msg, _, _ := progEncounter().ProgRoles.ShouldApply(progOpts(nil))

	assert.False(t, ok)
	assert.Equal(t, "No valid fights found in provided report!", msg)
}

func TestProgRolesGiveFurthestPhaseAndRemoveLowerOnes(t *testing.T) {
	e := progEncounter()

	ok, msg, add, remove := e.ProgRoles.ShouldApply(progOpts(nil, &fflogs.Fight{ID: 4, LastPhaseIndex: 1}, &fflogs.Fight{ID: 5, LastPhaseIndex: 2}))

	require.True(t, ok)
	assert.Equal(t, []string{"P3"}, names(add))
	assert.Equal(t, []string{"P1", "P2"}, names(remove))
	assert.Equal(t, "⮕ Fight 5\nYour furthest prog point is now `P3` (Phase **3**).\n", msg)
}

func TestProgRolesKillGivesClearedRole(t *testing.T) {
	e := progEncounter()

	ok, msg, add, remove := e.ProgRoles.ShouldApply(progOpts(nil, &fflogs.Fight{ID: 9, LastPhaseIndex: 2, Kill: true}))

	require.True(t, ok)
	assert.Equal(t, []string{"TOP-Cleared"}, names(add))
	assert.Equal(t, []string{"P1", "P2", "P3"}, names(remove))
	assert.Contains(t, msg, "`TOP-Cleared` (**Cleared**)")
}

func TestProgRolesNeverGoBackwards(t *testing.T) {
	e := progEncounter()

	ok, msg, _, _ := e.ProgRoles.ShouldApply(progOpts([]*Role{e.ProgRoles.Roles[2]}, &fflogs.Fight{ID: 1, LastPhaseIndex: 0}))
	assert.False(t, ok)
	assert.Contains(t, msg, "You already have a prog role further than the furthest prog in this report!")

	ok, msg, _, _ = e.ProgRoles.ShouldApply(progOpts([]*Role{e.ProgRoles.Roles[1]}, &fflogs.Fight{ID: 1, LastPhaseIndex: 1}))
	assert.False(t, ok)
	assert.Contains(t, msg, "is the same as the furthest prog point in this report")

	ok, _, _, _ = e.ProgRoles.ShouldApply(progOpts([]*Role{e.Roles[ClearedRole]}, &fflogs.Fight{ID: 1, LastPhaseIndex: 2}))
	assert.False(t, ok, "a cleared player keeps the cleared role")
}
