package scraper

import (
	"fmt"
	"strings"

	"github.com/Veraticus/findingway/internal/ffxiv"

	"github.com/gocolly/colly/v2"
)

type Scraper struct {
	Url string

	// parse turns one listing element into a Listing; nil means
	// parseListing. Tests replace it to exercise parse failures, which
	// colly's Unmarshal cannot produce for a Listing.
	parse func(e *colly.HTMLElement) (*ffxiv.Listing, error)
}

// Scrape collects the Listings from the listings page. A Listing that fails
// to parse is logged and skipped; an error is returned only when the page
// cannot be fetched or none of its Listings parse.
func (s *Scraper) Scrape() (*ffxiv.Listings, error) {
	listings := &ffxiv.Listings{}

	// The collector is synchronous, so these are final once Visit returns.
	var found, failed int
	var lastErr error

	parse := s.parse
	if parse == nil {
		parse = parseListing
	}

	c := colly.NewCollector()

	c.OnHTML("#listings.list .listing", func(e *colly.HTMLElement) {
		found++
		listing, err := parse(e)
		if err != nil {
			fmt.Printf("Skipping listing that could not be parsed: %v\n", err)
			failed++
			lastErr = err
			return
		}
		listings.Add(listing)
	})

	if err := c.Visit(s.Url + "/listings"); err != nil {
		return nil, fmt.Errorf("could not scrape listings: %w", err)
	}
	if found > 0 && failed == found {
		return nil, fmt.Errorf("could not scrape listings: none of %d listings parsed: %w", found, lastErr)
	}

	return listings, nil
}

// parseListing reads one Listing from its listing element.
func parseListing(e *colly.HTMLElement) (*ffxiv.Listing, error) {
	listing := &ffxiv.Listing{Party: []*ffxiv.Slot{}}

	// We can unmarshal a fair amount of information
	if err := e.Unmarshal(listing); err != nil {
		return nil, err
	}

	// Get attributes which are unmarshall-able
	listing.DataCentre = e.Attr("data-centre")
	listing.PfCategory = e.Attr("data-pf-category")
	listing.Id = e.Attr("data-id")

	// Get everything else that isn't easily inferred; first description
	description := e.ChildText(".left .description")
	description = strings.TrimSpace(strings.ReplaceAll(description, listing.Tags, ""))
	listing.Description = description

	// Then the party list
	e.ForEach(".party .slot", func(s int, p *colly.HTMLElement) {
		slot := ffxiv.NewSlot()
		class := p.Attr("class")

		if strings.Contains(class, "dps") {
			slot.Roles.Roles = append(slot.Roles.Roles, ffxiv.DPS)
		}

		if strings.Contains(class, "healer") {
			slot.Roles.Roles = append(slot.Roles.Roles, ffxiv.Healer)
		}

		if strings.Contains(class, "tank") {
			slot.Roles.Roles = append(slot.Roles.Roles, ffxiv.Tank)
		}

		if strings.Contains(class, "empty") {
			slot.Roles.Roles = append(slot.Roles.Roles, ffxiv.Empty)
		}

		if strings.Contains(class, "filled") {
			slot.Filled = true
			slot.Job = ffxiv.JobFromAbbreviation(p.Attr("title"))
		}

		listing.Party = append(listing.Party, slot)
	})

	return listing, nil
}
