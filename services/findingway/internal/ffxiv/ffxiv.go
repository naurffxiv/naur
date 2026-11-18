package ffxiv

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Listings struct {
	Listings []*Listing
}

type Listing struct {
	DataCentre  string
	Id          string
	PfCategory  string
	Duty        string `selector:".left .duty"`
	Tags        string `selector:".left .description span"`
	TagsColor   string `selector:".left .description span" attr:"class"`
	Description string
	MinIL       string `selector:".middle .stat .value"`
	Creator     string `selector:".right .creator .text"`
	World       string `selector:".right .world .text"`
	Expires     string `selector:".right .expires .text"`
	Updated     string `selector:".right .updated .text"`
	Party       []*Slot
}

type Slot struct {
	Roles  Roles
	Job    Job
	Filled bool
}

func NewSlot() *Slot {
	return &Slot{
		Roles: Roles{Roles: []Role{}},
	}
}

func (ls *Listings) ForDutyAndDataCentre(duty string, dataCentre string) *Listings {
	listings := &Listings{Listings: []*Listing{}}

	for _, l := range ls.Listings {
		if l.Duty == duty {
			if l.DataCentre == dataCentre {
				listings.Listings = append(listings.Listings, l)
			}
		}
	}

	return listings
}

func (ls *Listings) ForDutiesAndDataCentres(duties []string, dataCentres []string) *Listings {
	listings := &Listings{Listings: []*Listing{}}

	for _, l := range ls.Listings {
		if slices.Contains(duties, l.Duty) {
			if slices.Contains(dataCentres, l.DataCentre) {
				listings.Listings = append(listings.Listings, l)
			}
		}
	}

	return listings
}

// UpdatedWithinLast returns the Listings updated within duration before now.
func (ls *Listings) UpdatedWithinLast(duration time.Duration, now time.Time) (*Listings, error) {
	listings := &Listings{Listings: []*Listing{}}

	for _, l := range ls.Listings {
		updatedAt, err := l.UpdatedAt(now)
		if err != nil {
			return nil, fmt.Errorf("could not find most recent update time: %w", err)
		}
		if now.Add(-duration).Before(updatedAt) {
			listings.Listings = append(listings.Listings, l)
		}
	}

	return listings, nil
}

func (ls *Listings) Add(l *Listing) {
	for _, existingListing := range ls.Listings {
		if existingListing.Id == l.Id {
			return
		}
	}

	ls.Listings = append(ls.Listings, l)
}

func (l *Listing) PartyDisplay() string {
	var party strings.Builder

	for _, slot := range l.Party {
		if slot.Filled {
			party.WriteString(slot.Job.Emoji() + " ")
		} else {
			party.WriteString(slot.Roles.Emoji() + " ")
		}
	}

	return party.String()

}

func (l *Listing) GetExpires() string {
	return "<:hourglass:991379574187372655> " + l.Expires
}

func (l *Listing) GetUpdated() string {
	return "<:stopwatch:991379573000388758> " + l.Updated
}

func (l *Listing) GetTags() string {
	if len(l.Tags) == 0 {
		return "_ _"
	}
	return l.Tags
}

func (l *Listing) GetDescription() string {
	return "```" + l.Description + "```"
}

var updatedSecondsRegexp = regexp.MustCompile(`(\d+) seconds ago`)
var updatedMinutesRegexp = regexp.MustCompile(`(\d+) minutes ago`)
var updatedHoursRegexp = regexp.MustCompile(`(\d+) hours ago`)

// UpdatedAt converts the relative Updated text (e.g. "5 minutes ago") to an
// absolute time measured from now.
func (l *Listing) UpdatedAt(now time.Time) (time.Time, error) {
	if l.Updated == "" {
		return now, nil
	}

	if l.Updated == "now" {
		return now, nil
	}

	if l.Updated == "a second ago" {
		return now.Add(time.Duration(-1) * time.Second), nil
	}

	if l.Updated == "a minute ago" {
		return now.Add(time.Duration(-1) * time.Minute), nil
	}

	if l.Updated == "an hour ago" {
		return now.Add(time.Duration(-1) * time.Hour), nil
	}

	match := updatedSecondsRegexp.FindStringSubmatch(l.Updated)
	if len(match) != 0 {
		seconds, err := strconv.Atoi(match[1])
		if err != nil {
			return now, fmt.Errorf("could not parse time %v: %w", l.Updated, err)
		}
		return now.Add(time.Duration(-seconds) * time.Second), nil
	}

	match = updatedMinutesRegexp.FindStringSubmatch(l.Updated)
	if len(match) != 0 {
		minutes, err := strconv.Atoi(match[1])
		if err != nil {
			return now, fmt.Errorf("could not parse time %v: %w", l.Updated, err)
		}
		return now.Add(time.Duration(-minutes) * time.Minute), nil
	}

	match = updatedHoursRegexp.FindStringSubmatch(l.Updated)
	if len(match) != 0 {
		hours, err := strconv.Atoi(match[1])
		if err != nil {
			return now, fmt.Errorf("could not parse time %v: %w", l.Updated, err)
		}
		return now.Add(time.Duration(-hours) * time.Hour), nil
	}

	return now, fmt.Errorf("failed to parse time %v", l.Updated)
}
