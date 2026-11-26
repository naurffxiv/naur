package lodestone

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Veraticus/clearingway/internal/ffxiv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeLodestone serves HTML pages keyed by request URI and points the package at it.
func fakeLodestone(t *testing.T, pages map[string]string) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, ok := pages[r.URL.RequestURI()]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, page)
	}))
	t.Cleanup(ts.Close)

	oldLodestone, oldCharacter := lodestoneUrl, characterLodestoneUrl
	lodestoneUrl = ts.URL + "/lodestone"
	characterLodestoneUrl = ts.URL + "/lodestone/character/"
	t.Cleanup(func() { lodestoneUrl, characterLodestoneUrl = oldLodestone, oldCharacter })
}

const searchPath = "/lodestone/character/?q=Alex+Smith&worldname=Gilgamesh"

func searchPage(pager string, entries ...string) string {
	body := `<div class="ldst__window">`
	for _, e := range entries {
		body += e
	}
	if pager != "" {
		body += `<ul class="btn__pager"><li class="btn__pager__current">` + pager + `</li></ul>`
	}
	return `<html><body>` + body + `</div></body></html>`
}

func entry(name string, id int) string {
	return fmt.Sprintf(`<div class="entry"><a class="entry__link" href="/lodestone/character/%d/"><p class="entry__name">%s</p></a></div>`, id, name)
}

func alex() *ffxiv.Character {
	return &ffxiv.Character{FirstName: "Alex", LastName: "Smith", World: "Gilgamesh"}
}

func TestSetCharacterLodestoneIDSkipsLookupWhenAlreadySet(t *testing.T) {
	fakeLodestone(t, map[string]string{})
	c := alex()
	c.LodestoneID = 7

	require.NoError(t, SetCharacterLodestoneID(c))
	assert.Equal(t, 7, c.LodestoneID)
}

func TestSetCharacterLodestoneIDFindsExactNameMatch(t *testing.T) {
	fakeLodestone(t, map[string]string{
		searchPath: searchPage("Page 1 of 1", entry("Alex Smithson", 1), entry("alex smith", 42)),
	})
	c := alex()

	require.NoError(t, SetCharacterLodestoneID(c))
	assert.Equal(t, 42, c.LodestoneID)
}

func TestSetCharacterLodestoneIDSearchesEveryPage(t *testing.T) {
	fakeLodestone(t, map[string]string{
		searchPath:             searchPage("Page 1 of 2", entry("Someone Else", 1)),
		searchPath + "&page=2": searchPage("Page 2 of 2", entry("Alex Smith", 42)),
	})
	c := alex()

	require.NoError(t, SetCharacterLodestoneID(c))
	assert.Equal(t, 42, c.LodestoneID)
}

func TestSetCharacterLodestoneIDErrors(t *testing.T) {
	tests := map[string]struct {
		page string
		want string
	}{
		"no match": {
			searchPage("Page 1 of 1", entry("Someone Else", 1)),
			"no character found on the Lodestone for `Alex Smith (Gilgamesh)`!",
		},
		"several matches": {
			searchPage("Page 1 of 1", entry("Alex Smith", 1), entry("Alex Smith", 2)),
			"too many characters found for name Alex Smith (Gilgamesh)!",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fakeLodestone(t, map[string]string{searchPath: tt.page})
			c := alex()

			err := SetCharacterLodestoneID(c)

			assert.ErrorContains(t, err, tt.want)
			assert.Zero(t, c.LodestoneID)
		})
	}
}

func TestSetCharacterLodestoneIDReportsHTTPErrors(t *testing.T) {
	fakeLodestone(t, map[string]string{})

	err := SetCharacterLodestoneID(alex())

	assert.ErrorContains(t, err, "encountered search errors:\nNot Found")
}

func profilePage(bio string) string {
	return `<html><body><div class="character__content selected"><div class="character__selfintroduction">` + bio + `</div></div></body></html>`
}

func TestCharacterIsOwnedByDiscordUser(t *testing.T) {
	c := alex()
	c.LodestoneID = 42
	slug := c.LodestoneSlug("123456789")

	fakeLodestone(t, map[string]string{"/lodestone/character/42/": profilePage("Hello! " + slug)})
	owned, err := CharacterIsOwnedByDiscordUser(c, "123456789")
	require.NoError(t, err)
	assert.True(t, owned)

	owned, err = CharacterIsOwnedByDiscordUser(c, "987654321")
	require.NoError(t, err)
	assert.False(t, owned)
}

func TestGetAchievementsReadsQuotedAchievementNames(t *testing.T) {
	fakeLodestone(t, map[string]string{
		"/lodestone/character/42/achievement": `<html><body><ul>
			<li class="entry"><p class="entry__activity__txt">Alex Smith earned the achievement "The Ultimate Legend"!</p></li>
			<li class="entry"><p class="entry__activity__txt">Alex Smith earned the achievement "Bring It On"!</p></li>
		</ul></body></html>`,
	})
	c := alex()
	c.LodestoneID = 42

	achievements, err := GetAchievements(c)

	require.NoError(t, err)
	assert.Equal(t, []string{`"The Ultimate Legend"`, `"Bring It On"`}, achievements)
}
