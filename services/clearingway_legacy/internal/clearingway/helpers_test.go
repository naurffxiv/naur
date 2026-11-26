package clearingway

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/fflogs"
	"github.com/Veraticus/clearingway/internal/ffxiv"
	"github.com/stretchr/testify/require"
)

// dps builds a DPS rank for a fight starting at startMs on the given job.
func dps(startMs int, job string, percent float64) *fflogs.Rank {
	return &fflogs.Rank{
		StartTime: startMs, Job: ffxiv.Jobs[job],
		DPSParseFound: true, DPSPercent: percent,
		Report: fflogs.Report{Code: "r", FightId: startMs},
	}
}

// hps builds an HPS-only rank.
func hps(startMs int, job string, percent float64) *fflogs.Rank {
	return &fflogs.Rank{
		StartTime: startMs, Job: ffxiv.Jobs[job],
		HPSParseFound: true, HPSPercent: percent,
		Report: fflogs.Report{Code: "r", FightId: startMs},
	}
}

// rankings maps encounter IDs to rankings.
func rankings(byID map[int]*fflogs.Ranking) *fflogs.Rankings {
	return &fflogs.Rankings{Rankings: byID}
}

func cleared(kills int, ranks ...*fflogs.Rank) *fflogs.Ranking {
	return &fflogs.Ranking{TotalKills: kills, Ranks: ranks}
}

func encounter(name string, ids ...int) *Encounter {
	return &Encounter{Name: name, Ids: ids, Difficulty: "Ultimate"}
}

func applies(t *testing.T, roles *Roles, name string, opts *ShouldApplyOpts) (bool, string) {
	t.Helper()
	role := roles.FindByName(name)
	require.NotNil(t, role, "role %q", name)
	return role.ShouldApply(opts)
}

// matching returns the names of roles in rs whose rule applies.
func matching(rs *Roles, opts *ShouldApplyOpts) []string {
	names := []string{}
	for _, r := range rs.Roles {
		if ok, _ := r.ShouldApply(opts); ok {
			names = append(names, r.Name)
		}
	}
	return names
}

var (
	top = encounter("The Omega Protocol (Ultimate)", 1068, 1077)
	dsr = encounter("Dragonsong's Reprise (Ultimate)", 1065, 1076)
)
