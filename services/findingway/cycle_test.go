package main

import (
	"errors"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/board"
	"github.com/Veraticus/findingway/internal/ffxiv"

	"github.com/stretchr/testify/assert"
)

var cycleStart = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

// fakeLoop records what a cycle asks of its dependencies. Its clock moves
// on by refreshTakes during each refresh.
type fakeLoop struct {
	clock        time.Time
	refreshTakes time.Duration

	// scrapeErrs are returned by successive scrapes; past the end, scrapes
	// succeed.
	scrapeErrs []error
	recordErr  error

	scrapes       int
	refreshed     []string
	refreshedWith *ffxiv.Listings
	recorded      []time.Time
	posted        []time.Time
}

var testListings = &ffxiv.Listings{Listings: []*ffxiv.Listing{{Id: "1"}}}

func (f *fakeLoop) cycle(channels ...string) *cycle {
	var boardChannels []*board.Channel
	for _, id := range channels {
		boardChannels = append(boardChannels, &board.Channel{ID: id, Name: id})
	}
	return &cycle{
		scrape: func() (*ffxiv.Listings, error) {
			f.scrapes++
			if f.scrapes <= len(f.scrapeErrs) && f.scrapeErrs[f.scrapes-1] != nil {
				return nil, f.scrapeErrs[f.scrapes-1]
			}
			return testListings, nil
		},
		channels: boardChannels,
		refresh: func(channel *board.Channel, listings *ffxiv.Listings, now time.Time) error {
			f.refreshedWith = listings
			f.refreshed = append(f.refreshed, channel.ID)
			f.clock = f.clock.Add(f.refreshTakes)
			return nil
		},
		record: func(listings *ffxiv.Listings, now time.Time) error {
			f.recorded = append(f.recorded, now)
			return f.recordErr
		},
		postTokens: func(now time.Time) error {
			f.posted = append(f.posted, now)
			return nil
		},
		now: func() time.Time { return f.clock },
	}
}

func TestCycleRefreshesEveryChannelAndRecordsOnASuccessfulPass(t *testing.T) {
	f := &fakeLoop{clock: cycleStart, refreshTakes: 10 * time.Second}
	c := f.cycle("ucob", "uwu", "tea")

	wait, done := c.run()

	assert.False(t, done)
	assert.Equal(t, []string{"ucob", "uwu", "tea"}, f.refreshed)
	assert.Same(t, testListings, f.refreshedWith)
	assert.Equal(t, []time.Time{cycleStart.Add(30 * time.Second)}, f.recorded)
	assert.Equal(t, loopInterval-30*time.Second, wait, "the wait is the interval less the time spent refreshing")
	assert.Equal(t, 1, c.passes)
}

func TestCycleScrapeErrorSkipsThePassAndWaitsTheFullInterval(t *testing.T) {
	f := &fakeLoop{clock: cycleStart, scrapeErrs: []error{errors.New("xivpf is down")}}
	c := f.cycle("ucob")

	wait, done := c.run()

	assert.False(t, done)
	assert.Equal(t, loopInterval, wait)
	assert.Empty(t, f.refreshed)
	assert.Empty(t, f.recorded)
	assert.Empty(t, f.posted)
	assert.Equal(t, 0, c.passes, "a failed scrape is not a pass")
}

func TestCycleRecordErrorStillRefreshesBoards(t *testing.T) {
	f := &fakeLoop{clock: cycleStart, recordErr: errors.New("redis unreachable")}
	c := f.cycle("ucob", "uwu")

	_, done := c.run()

	assert.False(t, done)
	assert.Equal(t, []string{"ucob", "uwu"}, f.refreshed)
	assert.Len(t, f.recorded, 1)
	assert.Len(t, f.posted, 1, "the token report is still posted")
}

func TestCyclePostsTokensOnTheFirstPassAndEvery20thAfter(t *testing.T) {
	f := &fakeLoop{clock: cycleStart}
	c := f.cycle("ucob")

	var postedOn []int
	for pass := range 41 {
		before := len(f.posted)
		c.run()
		if len(f.posted) > before {
			postedOn = append(postedOn, pass)
		}
	}

	assert.Equal(t, []int{0, 20, 40}, postedOn)
}

func TestCycleScrapeErrorsDoNotShiftTheTokenPost(t *testing.T) {
	f := &fakeLoop{clock: cycleStart, scrapeErrs: []error{nil, errors.New("down"), errors.New("down")}}
	c := f.cycle("ucob")

	var postedOnRun []int
	for run := range 23 {
		before := len(f.posted)
		c.run()
		if len(f.posted) > before {
			postedOnRun = append(postedOnRun, run)
		}
	}

	// Runs 1 and 2 fail, so the 20th successful pass after the first is
	// run 22.
	assert.Equal(t, []int{0, 22}, postedOnRun)
	assert.Equal(t, 21, c.passes)
}

func TestCycleOnceEndsAfterTheFirstSuccessfulPass(t *testing.T) {
	f := &fakeLoop{clock: cycleStart}
	c := f.cycle("ucob")
	c.once = true

	wait, done := c.run()

	assert.True(t, done)
	assert.Zero(t, wait)
	assert.Equal(t, []string{"ucob"}, f.refreshed)
	assert.Len(t, f.posted, 1)
}

func TestCycleOnceKeepsRetryingAfterAScrapeError(t *testing.T) {
	f := &fakeLoop{clock: cycleStart, scrapeErrs: []error{errors.New("down"), errors.New("down")}}
	c := f.cycle("ucob")
	c.once = true

	var waits []time.Duration
	for {
		wait, done := c.run()
		if done {
			break
		}
		waits = append(waits, wait)
	}

	assert.Equal(t, []time.Duration{loopInterval, loopInterval}, waits)
	assert.Equal(t, 3, f.scrapes)
	assert.Equal(t, []string{"ucob"}, f.refreshed)
	assert.Len(t, f.posted, 1, "the token report is posted on the first successful pass")
}
