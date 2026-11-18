package main

import (
	"fmt"
	"time"

	"github.com/Veraticus/findingway/internal/board"
	"github.com/Veraticus/findingway/internal/ffxiv"
)

const (
	// loopInterval is how often the listings are scraped and posted.
	loopInterval = 3 * time.Minute
	// tokenPostEvery is how many successful passes apart the token report
	// is posted: every 20th pass, starting with the first, is hourly at the
	// 3-minute loop interval.
	tokenPostEvery = 20
)

// cycle is the main loop's work, one pass per run. Its dependencies are
// funcs so that tests can fake them.
type cycle struct {
	// scrape fetches the current Listings.
	scrape func() (*ffxiv.Listings, error)
	// channels are each refreshed with refresh on every successful pass.
	channels []*board.Channel
	refresh  func(channel *board.Channel, listings *ffxiv.Listings, now time.Time) error
	// record adds the pass's Listings to the description log.
	record func(listings *ffxiv.Listings, now time.Time) error
	// postTokens posts the token report.
	postTokens func(now time.Time) error
	now        func() time.Time
	// once ends the loop after the first successful pass.
	once bool

	// passes counts the successful passes so far.
	passes int
}

// run does one pass: scrape, refresh every channel, record the Listings
// and, on every tokenPostEvery-th successful pass, post the token report.
// It returns how long to wait before the next pass (the loop interval less
// the time spent refreshing channels), or done when once is set and the
// pass succeeded.
//
// A failed scrape skips the rest of the pass: it does not count as a pass,
// does not end a once run, and waits the full loop interval before the
// next try.
func (c *cycle) run() (wait time.Duration, done bool) {
	fmt.Printf("Scraping source...\n")
	listings, err := c.scrape()
	if err != nil {
		fmt.Printf("Scraper error: %v\n", err)
		return loopInterval, false
	}
	fmt.Printf("Got %v listings.\n", len(listings.Listings))
	fmt.Printf("Sending to %v channels...\n", len(c.channels))

	wait = loopInterval
	for _, channel := range c.channels {
		startTime := c.now()
		fmt.Printf("Refreshing Discord for %v (%v)...\n", channel.Name, channel.Duty)
		if err := c.refresh(channel, listings, startTime); err != nil {
			fmt.Printf("Discord error refreshing %v: %v\n", channel.Name, err)
		}
		wait -= c.now().Sub(startTime)
	}

	if err := c.record(listings, c.now()); err != nil {
		fmt.Printf("Error recording listing descriptions: %v\n", err)
	}

	if c.passes%tokenPostEvery == 0 {
		fmt.Println("Sending tokens to discord")
		if err := c.postTokens(c.now()); err != nil {
			fmt.Printf("Error posting tokens: %v\n", err)
		}
	}

	if c.once {
		return 0, true
	}
	c.passes++
	return wait, false
}
