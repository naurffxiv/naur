package fflogs

import (
	"testing"

	"github.com/Veraticus/clearingway/internal/ffxiv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func standard(metric Metric, kills int, ranks ...*Rank) *Ranking {
	return &Ranking{Metric: metric, TotalKills: kills, Partition: 1, Ranks: ranks}
}

func rank(start int, spec string, percent float64) *Rank {
	return &Rank{StartTime: start, Spec: spec, RankPercent: percent}
}

func TestRankingsAddMergesDPSAndHPSForTheSameFight(t *testing.T) {
	rs := &Rankings{Rankings: map[int]*Ranking{}}

	require.NoError(t, rs.Add(1, standard(Dps, 3, rank(1000, "WhiteMage", 80))))
	require.NoError(t, rs.Add(1, standard(Hps, 3, rank(1000, "WhiteMage", 95), rank(2000, "Sage", 50))))

	r := rs.Rankings[1]
	assert.Equal(t, 3, r.TotalKills, "kills are counted from the hps response only")
	require.Len(t, r.Ranks, 2)

	same := r.Ranks[0]
	assert.True(t, same.DPSParseFound)
	assert.True(t, same.HPSParseFound)
	assert.InDelta(t, 80, same.DPSPercent, 0.001)
	assert.InDelta(t, 95, same.HPSPercent, 0.001)
	assert.Equal(t, "WHM", same.Job.Abbreviation)

	added := r.Ranks[1]
	assert.False(t, added.DPSParseFound)
	assert.True(t, added.HPSParseFound)
	assert.Equal(t, "SGE", added.Job.Abbreviation)
}

func TestRankingsAddSkipsNonstandardCopiesOfStandardRankings(t *testing.T) {
	rs := &Rankings{Rankings: map[int]*Ranking{}}

	copyOfStandard := &Ranking{Metric: Dps, Partition: 1, Nonstandard: true, Ranks: []*Rank{rank(1000, "Paladin", 50)}}
	require.NoError(t, rs.Add(1, copyOfStandard))
	assert.Empty(t, rs.Rankings)

	realNonstandard := &Ranking{Metric: Dps, Partition: 2, Nonstandard: true, Ranks: []*Rank{rank(1000, "Paladin", 50)}}
	require.NoError(t, rs.Add(1, realNonstandard))
	assert.Len(t, rs.Rankings[1].Ranks, 1)
}

func TestRankingsAddRejectsUnknownJob(t *testing.T) {
	rs := &Rankings{Rankings: map[int]*Ranking{}}

	err := rs.Add(1, standard(Dps, 0, rank(1000, "Freelancer", 50)))

	assert.ErrorContains(t, err, "could not find job Freelancer")
}

func TestRankingCleared(t *testing.T) {
	assert.True(t, (&Ranking{TotalKills: 1}).Cleared())
	assert.False(t, (&Ranking{}).Cleared())
}

func TestRankingBestAndWorstRanks(t *testing.T) {
	r := &Ranking{Ranks: []*Rank{
		{StartTime: 3, DPSParseFound: true, DPSPercent: 50, HPSParseFound: true, HPSPercent: 10},
		{StartTime: 1, DPSParseFound: true, DPSPercent: 90},
		{StartTime: 2, HPSParseFound: true, HPSPercent: 70},
	}}

	assert.InDelta(t, 90, r.BestDPSRank().DPSPercent, 0.001)
	assert.InDelta(t, 50, r.WorstDPSRank().DPSPercent, 0.001)
	assert.InDelta(t, 70, r.BestHPSRank().HPSPercent, 0.001)
	assert.InDelta(t, 10, r.WorstHPSRank().HPSPercent, 0.001)

	empty := &Ranking{}
	assert.Nil(t, empty.BestDPSRank())
	assert.Nil(t, empty.WorstDPSRank())
	assert.Nil(t, empty.BestHPSRank())
	assert.Nil(t, empty.WorstHPSRank())
}

func TestRankingRanksByTimeLeavesOriginalOrder(t *testing.T) {
	r := &Ranking{Ranks: []*Rank{{StartTime: 3}, {StartTime: 1}, {StartTime: 2}}}

	sorted := r.RanksByTime()

	assert.Equal(t, []int{1, 2, 3}, []int{sorted[0].StartTime, sorted[1].StartTime, sorted[2].StartTime})
	assert.Equal(t, 3, r.Ranks[0].StartTime)
}

func TestRankStrings(t *testing.T) {
	r := &Rank{
		StartTime:  1700000000123,
		DPSPercent: 99.456,
		HPSPercent: 12.3,
		Job:        &ffxiv.Job{Abbreviation: "SAM"},
		Report:     Report{Code: "abc123", FightId: 7},
	}

	assert.Equal(t, 1700000000, r.UnixTime())
	assert.Equal(t, "99.46", r.DPSPercentString())
	assert.Equal(t, "12.30", r.HPSPercentString())
	assert.Equal(t, "https://www.fflogs.com/reports/abc123#fight=7", r.Report.Url())
	assert.Equal(t,
		"Best parse was **99.46** with `SAM` in `TOP` on <t:1700000000:F> (https://www.fflogs.com/reports/abc123#fight=7).\nUse `/uncolor` if you don't want this role.",
		r.BestDPSParseString("TOP"))
}
