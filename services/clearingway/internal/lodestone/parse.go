package lodestone

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type searchPage struct {
	MatchingIDs []int64
	TotalPages  int
}

// parseSearchPage - Extracts the Lodestone IDs of entries matching the given name along with the total page count
func parseSearchPage(doc *goquery.Document, name string) (*searchPage, error) {
	page := &searchPage{}

	var parseErr error
	doc.Find(".ldst__window .entry").EachWithBreak(func(_ int, entry *goquery.Selection) bool {
		entryName := strings.TrimSpace(entry.Find(".entry__name").Text())
		if !namesEqual(entryName, name) {
			return true
		}

		href, _ := entry.Find(".entry__link").Attr("href")
		var id int64
		if _, err := fmt.Sscanf(href, "/lodestone/character/%d/", &id); err != nil {
			parseErr = fmt.Errorf("%w: could not parse character link %q: %w", ErrUnexpectedPage, href, err)
			return false
		}
		page.MatchingIDs = append(page.MatchingIDs, id)
		return true
	})
	if parseErr != nil {
		return nil, parseErr
	}

	totalPages, err := parseTotalPages(doc)
	if err != nil {
		return nil, err
	}
	page.TotalPages = totalPages

	return page, nil
}

// parseTotalPages - Reads the "Page X of Y" pager
func parseTotalPages(doc *goquery.Document) (int, error) {
	pager := doc.Find(".ldst__window ul.btn__pager .btn__pager__current").First()
	if pager.Length() == 0 {
		return 1, nil
	}

	var currentPage, totalPages int
	text := strings.TrimSpace(pager.Text())
	if _, err := fmt.Sscanf(text, "Page %d of %d", &currentPage, &totalPages); err != nil {
		return 0, fmt.Errorf("%w: could not parse pager %q: %w", ErrUnexpectedPage, text, err)
	}
	return totalPages, nil
}

// parseCharacterBio - Extracts the self-introduction text from a character profile page
func parseCharacterBio(doc *goquery.Document) (string, error) {
	profile := doc.Find(".character__content.selected")
	if profile.Length() == 0 {
		return "", fmt.Errorf("%w: could not find character profile", ErrUnexpectedPage)
	}
	return strings.TrimSpace(profile.Find(".character__selfintroduction").Text()), nil
}

// namesEqual - Case-insensitively compares two character names
func namesEqual(a, b string) bool {
	normalise := func(s string) string { return strings.ReplaceAll(s, "’", "'") }
	return strings.EqualFold(normalise(a), normalise(b))
}
