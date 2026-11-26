package fflogs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFightsAddIgnoresDuplicateFightsFromTheSameReport(t *testing.T) {
	fs := &Fights{}

	fs.Add(&Fight{ID: 1, ReportID: "a"})
	fs.Add(&Fight{ID: 1, ReportID: "a"})
	fs.Add(&Fight{ID: 1, ReportID: "b"})

	assert.Len(t, fs.Fights, 2)
}

func TestFightsFurthestFight(t *testing.T) {
	assert.Nil(t, (&Fights{}).FurthestFight())

	prog := &Fights{Fights: []*Fight{{ID: 1, LastPhaseIndex: 2}, {ID: 2, LastPhaseIndex: 4}, {ID: 3, LastPhaseIndex: 3}}}
	assert.Equal(t, 2, prog.FurthestFight().ID)

	withKill := &Fights{Fights: []*Fight{{ID: 1, LastPhaseIndex: 5}, {ID: 2, LastPhaseIndex: 1, Kill: true}}}
	assert.Equal(t, 2, withKill.FurthestFight().ID)
}

func TestFightReportURL(t *testing.T) {
	assert.Equal(t, "https://www.fflogs.com/reports/abc#fight=4", (&Fight{ReportID: "abc", ID: 4}).ReportURL())
}
