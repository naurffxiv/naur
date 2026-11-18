package scraper

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Veraticus/findingway/internal/ffxiv"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"
)

func newTestServer() (*httptest.Server, error) {
	mux := http.NewServeMux()

	listings, err := os.ReadFile("testdata/listings.html")
	if err != nil {
		return nil, err
	}

	mux.HandleFunc("/listings", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(listings)
	})

	return httptest.NewServer(mux), nil
}

func TestGetListings(t *testing.T) {
	ts, err := newTestServer()
	assert.NoError(t, err)
	defer ts.Close()

	s := &Scraper{Url: ts.URL}
	listings, err := s.Scrape()
	assert.NoError(t, err)

	assert.Equal(t, 656, len(listings.Listings))

	listing := listings.Listings[100]
	assert.Equal(t, listing.Duty, "The Unending Coil of Bahamut (Ultimate)")
	assert.Equal(t, listing.Party[0].Job, ffxiv.AST)
	assert.True(t, listing.Party[0].Filled)
}

func TestScrapeReturnsErrorWhenPageCannotBeFetched(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := &Scraper{Url: ts.URL}
	listings, err := s.Scrape()

	assert.Error(t, err)
	assert.Nil(t, listings)
}

func TestScrapeReturnsNoListingsForEmptyPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><body><div id="listings" class="list"></div></body></html>`))
	}))
	defer ts.Close()

	s := &Scraper{Url: ts.URL}
	listings, err := s.Scrape()

	assert.NoError(t, err)
	assert.Empty(t, listings.Listings)
}

const twoListingsPage = `<html><body><div id="listings" class="list">
<div class="listing" data-id="bad" data-centre="Aether"></div>
<div class="listing" data-id="good" data-centre="Primal"></div>
</div></body></html>`

func servePage(t *testing.T, page string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(page))
	}))
	t.Cleanup(ts.Close)
	return ts
}

// failingFor parses listings normally except those whose data-id is in ids.
func failingFor(ids ...string) func(e *colly.HTMLElement) (*ffxiv.Listing, error) {
	return func(e *colly.HTMLElement) (*ffxiv.Listing, error) {
		for _, id := range ids {
			if e.Attr("data-id") == id {
				return nil, errors.New("unparseable " + id)
			}
		}
		return parseListing(e)
	}
}

func TestScrapeSkipsListingsThatFailToParse(t *testing.T) {
	ts := servePage(t, twoListingsPage)

	s := &Scraper{Url: ts.URL, parse: failingFor("bad")}
	listings, err := s.Scrape()

	assert.NoError(t, err)
	if assert.Len(t, listings.Listings, 1) {
		assert.Equal(t, "good", listings.Listings[0].Id)
		assert.Equal(t, "Primal", listings.Listings[0].DataCentre)
	}
}

func TestScrapeReturnsErrorWhenNoListingParses(t *testing.T) {
	ts := servePage(t, twoListingsPage)

	s := &Scraper{Url: ts.URL, parse: failingFor("bad", "good")}
	listings, err := s.Scrape()

	assert.ErrorContains(t, err, "none of 2 listings parsed: unparseable good")
	assert.Nil(t, listings)
}
