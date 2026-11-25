package lodestone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEntry struct {
	Name string
	ID   int64
}

// searchPageHTML - Builds a minimal Lodestone search results page
func searchPageHTML(entries []testEntry, currentPage, totalPages int) string {
	var sb strings.Builder
	sb.WriteString(`<html><body><div class="ldst__window">`)
	if totalPages > 0 {
		fmt.Fprintf(&sb, `<ul class="btn__pager"><li class="btn__pager__current">Page %d of %d</li></ul>`, currentPage, totalPages)
	}
	for _, e := range entries {
		fmt.Fprintf(&sb, `<div class="entry"><a href="/lodestone/character/%d/" class="entry__link"><p class="entry__name">%s</p></a></div>`, e.ID, e.Name)
	}
	sb.WriteString(`</div></body></html>`)
	return sb.String()
}

// profilePageHTML - Builds a minimal Lodestone character profile page
func profilePageHTML(bio string) string {
	return fmt.Sprintf(`<html><body><div class="character__content selected"><div class="character__selfintroduction">%s</div></div></body></html>`, bio)
}

func htmlResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func NewTestLodestone(doFunc func(req *http.Request) (*http.Response, error)) *Lodestone {
	return &Lodestone{
		BaseURL:    "https://lodestone.test/lodestone",
		HTTPClient: &MockHTTPClient{DoFunc: doFunc},
	}
}

// pageNumber - Returns the requested search page number (defaults to 1)
func pageNumber(req *http.Request) int {
	page, err := strconv.Atoi(req.URL.Query().Get("page"))
	if err != nil {
		return 1
	}
	return page
}

// ============== SearchCharacterLodestoneID ==============

func TestSearchCharacterLodestoneID_FoundOnFirstPage(t *testing.T) {
	var requests int
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		requests++
		return htmlResponse(http.StatusOK, searchPageHTML([]testEntry{
			{Name: "Alicja Testerson", ID: 111},
			{Name: "Alicja Tester", ID: 222},
		}, 1, 1)), nil
	})

	id, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.NoError(t, err)
	assert.Equal(t, int64(222), id)
	assert.Equal(t, 1, requests)
}

func TestSearchCharacterLodestoneID_RequestParameters(t *testing.T) {
	var capturedReq *http.Request
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		capturedReq = req
		return htmlResponse(http.StatusOK, searchPageHTML([]testEntry{{Name: "Y'shtola Rhul", ID: 1}}, 1, 1)), nil
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Y'shtola Rhul", "Gilgamesh")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, capturedReq.Method)
	assert.Equal(t, "/lodestone/character/", capturedReq.URL.Path)
	assert.Equal(t, "Y'shtola Rhul", capturedReq.URL.Query().Get("q"))
	assert.Equal(t, "Gilgamesh", capturedReq.URL.Query().Get("worldname"))
	assert.Empty(t, capturedReq.URL.Query().Get("page"))
	assert.Equal(t, userAgent, capturedReq.Header.Get("User-Agent"))
}

func TestSearchCharacterLodestoneID_CaseAndApostropheInsensitive(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, searchPageHTML([]testEntry{{Name: "Y’shtola Rhul", ID: 42}}, 1, 1)), nil
	})

	id, err := ls.SearchCharacterLodestoneID(context.Background(), "y'shtola rhul", "Gilgamesh")

	assert.NoError(t, err)
	assert.Equal(t, int64(42), id)
}

func TestSearchCharacterLodestoneID_NoResults(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		// No pager is rendered when there are no results
		return htmlResponse(http.StatusOK, searchPageHTML(nil, 0, 0)), nil
	})

	id, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, ErrCharacterNotFound)
	assert.Equal(t, int64(0), id)
}

func TestSearchCharacterLodestoneID_OnlyPartialMatches(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, searchPageHTML([]testEntry{{Name: "Alicja Testerson", ID: 1}}, 1, 1)), nil
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, ErrCharacterNotFound)
}

func TestSearchCharacterLodestoneID_MultipleMatches(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, searchPageHTML([]testEntry{
			{Name: "Alicja Tester", ID: 1},
			{Name: "Alicja Tester", ID: 2},
		}, 1, 1)), nil
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, ErrMultipleCharactersFound)
}

func TestSearchCharacterLodestoneID_StopsAtBatchContainingMatch(t *testing.T) {
	const totalPages = 20
	const matchPage = 8 // Second batch (pages 7-11)

	var mu sync.Mutex
	requestedPages := map[int]bool{}
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		page := pageNumber(req)
		mu.Lock()
		requestedPages[page] = true
		mu.Unlock()

		entries := []testEntry{{Name: "Someone Else", ID: int64(page * 100)}}
		if page == matchPage {
			entries = append(entries, testEntry{Name: "Alicja Tester", ID: 999})
		}
		return htmlResponse(http.StatusOK, searchPageHTML(entries, page, totalPages)), nil
	})

	id, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.NoError(t, err)
	assert.Equal(t, int64(999), id)
	// Page 1, then batch 2-6, then batch 7-11
	for page := 1; page <= 11; page++ {
		assert.True(t, requestedPages[page], "expected page %d to be requested", page)
	}
	for page := 12; page <= totalPages; page++ {
		assert.False(t, requestedPages[page], "expected page %d not to be requested", page)
	}
}

func TestSearchCharacterLodestoneID_ExhaustsAllPages(t *testing.T) {
	const totalPages = 7

	var mu sync.Mutex
	requestedPages := map[int]bool{}
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		page := pageNumber(req)
		mu.Lock()
		requestedPages[page] = true
		mu.Unlock()
		return htmlResponse(http.StatusOK, searchPageHTML([]testEntry{{Name: "Someone Else", ID: 1}}, page, totalPages)), nil
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, ErrCharacterNotFound)
	assert.Len(t, requestedPages, totalPages)
	assert.False(t, requestedPages[totalPages+1])
}

func TestSearchCharacterLodestoneID_DuplicateAcrossPagesInBatch(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		page := pageNumber(req)
		var entries []testEntry
		// Results shifted between pages so the same character appears twice
		if page == 2 || page == 3 {
			entries = []testEntry{{Name: "Alicja Tester", ID: 555}}
		}
		return htmlResponse(http.StatusOK, searchPageHTML(entries, page, 3)), nil
	})

	id, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.NoError(t, err)
	assert.Equal(t, int64(555), id)
}

func TestSearchCharacterLodestoneID_ErrorInBatch(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		page := pageNumber(req)
		if page == 3 {
			return htmlResponse(http.StatusServiceUnavailable, "maintenance"), nil
		}
		return htmlResponse(http.StatusOK, searchPageHTML(nil, page, 4)), nil
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, ErrUnexpectedStatus)
	assert.Contains(t, err.Error(), "503")
}

func TestSearchCharacterLodestoneID_HTTPError(t *testing.T) {
	expectedErr := errors.New("connection refused")
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return nil, expectedErr
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, expectedErr)
}

func TestSearchCharacterLodestoneID_MalformedLink(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, `<div class="ldst__window"><div class="entry"><a href="/somewhere/else" class="entry__link"><p class="entry__name">Alicja Tester</p></a></div></div>`), nil
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, ErrUnexpectedPage)
}

func TestSearchCharacterLodestoneID_MalformedPager(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, `<div class="ldst__window"><ul class="btn__pager"><li class="btn__pager__current">Seite 1 von 3</li></ul></div>`), nil
	})

	_, err := ls.SearchCharacterLodestoneID(context.Background(), "Alicja Tester", "Gilgamesh")

	assert.ErrorIs(t, err, ErrUnexpectedPage)
}

// ============== IsCharacterOwnedByDiscordUser ==============

func TestIsCharacterOwnedByDiscordUser_Owned(t *testing.T) {
	var capturedPath string
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		capturedPath = req.URL.Path
		return htmlResponse(http.StatusOK, profilePageHTML("Hello! "+OwnershipCode("123456789")+" Bye!")), nil
	})

	owned, err := ls.IsCharacterOwnedByDiscordUser(context.Background(), 12345, "123456789")

	assert.NoError(t, err)
	assert.True(t, owned)
	assert.Equal(t, "/lodestone/character/12345/", capturedPath)
}

func TestIsCharacterOwnedByDiscordUser_NotOwned(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, profilePageHTML(OwnershipCode("someone-else"))), nil
	})

	owned, err := ls.IsCharacterOwnedByDiscordUser(context.Background(), 12345, "123456789")

	assert.NoError(t, err)
	assert.False(t, owned)
}

func TestIsCharacterOwnedByDiscordUser_EmptyBio(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, profilePageHTML("")), nil
	})

	owned, err := ls.IsCharacterOwnedByDiscordUser(context.Background(), 12345, "123456789")

	assert.NoError(t, err)
	assert.False(t, owned)
}

func TestIsCharacterOwnedByDiscordUser_CharacterNotFound(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusNotFound, "not found"), nil
	})

	owned, err := ls.IsCharacterOwnedByDiscordUser(context.Background(), 12345, "123456789")

	assert.ErrorIs(t, err, ErrCharacterNotFound)
	assert.False(t, owned)
}

func TestIsCharacterOwnedByDiscordUser_MissingProfile(t *testing.T) {
	ls := NewTestLodestone(func(req *http.Request) (*http.Response, error) {
		return htmlResponse(http.StatusOK, "<html><body>Under maintenance</body></html>"), nil
	})

	owned, err := ls.IsCharacterOwnedByDiscordUser(context.Background(), 12345, "123456789")

	assert.ErrorIs(t, err, ErrUnexpectedPage)
	assert.False(t, owned)
}

// ============== OwnershipCode ==============

func TestCharacterURL(t *testing.T) {
	ls := Init()
	assert.Equal(t, "https://na.finalfantasyxiv.com/lodestone/character/12345/", ls.CharacterURL(12345))
}

func TestProfileSettingsURL(t *testing.T) {
	ls := Init()
	assert.Equal(t, "https://na.finalfantasyxiv.com/lodestone/my/setting/profile/", ls.ProfileSettingsURL())
}

func TestOwnershipCode_MatchesLegacyFormat(t *testing.T) {
	// adler32("123456789") = 0x091E01DE = 152961502
	assert.Equal(t, "clearingway-152961502", OwnershipCode("123456789"))
}

func TestOwnershipCode_Deterministic(t *testing.T) {
	assert.Equal(t, OwnershipCode("abc"), OwnershipCode("abc"))
	assert.NotEqual(t, OwnershipCode("abc"), OwnershipCode("abd"))
}
