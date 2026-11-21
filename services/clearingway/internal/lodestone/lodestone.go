package lodestone

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	defaultBaseURL      = "https://na.finalfantasyxiv.com/lodestone"
	defaultTimeout      = 30 * time.Second
	userAgent           = "Clearingway"
	searchPageBatchSize = 5
)

var (
	ErrCharacterNotFound       = errors.New("character not found on the Lodestone")
	ErrMultipleCharactersFound = errors.New("multiple characters found on the Lodestone")
	ErrUnexpectedStatus        = errors.New("unexpected response status from the Lodestone")
	ErrUnexpectedPage          = errors.New("unexpected page structure from the Lodestone")
)

type HTTPClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type Lodestone struct {
	BaseURL    string
	HTTPClient HTTPClientInterface
}

// Init - Initializes the Lodestone struct
func Init() *Lodestone {
	return &Lodestone{
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: defaultTimeout},
	}
}

// SearchCharacterLodestoneID - Searches the Lodestone for a character by name on a given world
func (ls *Lodestone) SearchCharacterLodestoneID(ctx context.Context, name, world string) (int64, error) {
	firstPage, err := ls.fetchSearchPage(ctx, name, world, 1)
	if err != nil {
		return 0, err
	}

	ids := firstPage.MatchingIDs
	for start := 2; len(ids) == 0 && start <= firstPage.TotalPages; start += searchPageBatchSize {
		end := min(start+searchPageBatchSize-1, firstPage.TotalPages)
		ids, err = ls.fetchSearchPages(ctx, name, world, start, end)
		if err != nil {
			return 0, err
		}
	}

	switch len(ids) {
	case 0:
		return 0, fmt.Errorf("%w: %s (%s)", ErrCharacterNotFound, name, world)
	case 1:
		return ids[0], nil
	default:
		return 0, fmt.Errorf("%w: %s (%s)", ErrMultipleCharactersFound, name, world)
	}
}

// IsCharacterOwnedByDiscordUser - Checks if the character's Lodestone bio contains the ownership code for a Discord user
func (ls *Lodestone) IsCharacterOwnedByDiscordUser(ctx context.Context, lodestoneId int64, discordId string) (bool, error) {
	doc, err := ls.get(ctx, fmt.Sprintf("/character/%d/", lodestoneId), nil)
	if err != nil {
		return false, err
	}

	bio, err := parseCharacterBio(doc)
	if err != nil {
		return false, err
	}

	return strings.Contains(bio, OwnershipCode(discordId)), nil
}

// CharacterURL - Returns the Lodestone profile URL for a character
func (ls *Lodestone) CharacterURL(lodestoneId int64) string {
	return fmt.Sprintf("%s/character/%d/", ls.BaseURL, lodestoneId)
}

// ProfileSettingsURL - Returns the URL where users can edit their Lodestone bio
func (ls *Lodestone) ProfileSettingsURL() string {
	return ls.BaseURL + "/my/setting/profile/"
}

// fetchSearchPage - Fetches and parses a single character search results page
func (ls *Lodestone) fetchSearchPage(ctx context.Context, name, world string, page int) (*searchPage, error) {
	query := url.Values{
		"q":         {name},
		"worldname": {world},
	}
	if page > 1 {
		query.Set("page", fmt.Sprint(page))
	}

	doc, err := ls.get(ctx, "/character/", query)
	if err != nil {
		return nil, err
	}
	return parseSearchPage(doc, name)
}

// fetchSearchPages - Concurrently fetches search pages start to end (inclusive) and returns de-duplicated matching IDs
func (ls *Lodestone) fetchSearchPages(ctx context.Context, name, world string, start, end int) ([]int64, error) {
	count := end - start + 1
	pages := make([]*searchPage, count)
	errs := make([]error, count)

	var wg sync.WaitGroup
	for i := range count {
		wg.Go(func() {
			pages[i], errs[i] = ls.fetchSearchPage(ctx, name, world, start+i)
		})
	}
	wg.Wait()

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	var ids []int64
	seen := make(map[int64]struct{})
	for _, page := range pages {
		for _, id := range page.MatchingIDs {
			// Results can shift between pages while paginating, so de-duplicate
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// get - Performs a GET request against the Lodestone and parses the response body as HTML
func (ls *Lodestone) get(ctx context.Context, path string, query url.Values) (*goquery.Document, error) {
	reqURL := ls.BaseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Lodestone request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := ls.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error requesting %s: %w", reqURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrCharacterNotFound
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: %s for %s", ErrUnexpectedStatus, resp.Status, reqURL)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing Lodestone response: %w", err)
	}
	return doc, nil
}
