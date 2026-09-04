# Tokenizer Improvements

Changes to `tokenizer.go`, centered on `splitListingIntoTokens` plus two changes
in `TokenizeListings`.

**Scope note:** The tokenizer only needs to produce a readable 3-hour pulse for
Discord. The raw `descriptions:<day>` list is preserved in Redis, so anything
requiring normalization, alias-merging, or trend detection is deferred to the
weekly LLM-assisted analysis. These changes are deliberately lean and do not try
to capture everything.

---

## 1. Pre-process URLs before splitting

Extract raidplan codes as canonical tokens and strip all other URLs. Eliminates
`plan`, `https:`, `raidplan.io` from the top tokens and merges fragment variants
like `lpsbjecjdb0xoloz#2` and `lpsbjecjdb0xoloz#14` into a single token.

**Note:** The scheme (`https://`) is optional — players paste bare
`raidplan.io/plan/CODE` links constantly, and a scheme-required regex would
regenerate `raidplan.io` + `plan` + code as three tokens. The fragment matcher
is widened to `#\S+` so a non-numeric `#...` does not leak a stray token.

```go
// Extract raidplan code as a bare token, drop the rest of the URL.
// Scheme is optional to catch bare raidplan.io/plan/CODE links.
raidplanRe := regexp.MustCompile(`(?:https?://)?raidplan\.io/plan/([A-Za-z0-9]+)(?:#\S+)?`)
listing = raidplanRe.ReplaceAllStringFunc(listing, func(m string) string {
    return raidplanRe.FindStringSubmatch(m)[1]
})

// Strip all other URLs entirely (tinyurl, pastebin, etc.)
listing = regexp.MustCompile(`https?://\S+`).ReplaceAllString(listing, "")
```

This must run **before** the lowercase pass below (see §3 ordering note).

---

## 2. Don't split on bare `-`

Currently `-` is a split delimiter, which breaks useful compound tokens like
`d1-d4`. The fix is simply to **remove** `-` from the `splitRegex` slice.

In the `splitRegex` slice, delete:

```go
`-`,    // -
```

**Do not** replace it with `` ` - ` `` (spaced hyphen). That branch would be dead
code: Go's regexp is leftmost-first, and the single-space ` ` alternative sits
earlier in the alternation, so any `a - b` splits on the first space before the
` - ` branch is ever tried. Removing `-` is the entire fix — compounds like
`d1-d4` stay intact, and spaced hyphens still split on the surrounding spaces.

---

## 3. Normalize, trim, and filter tokens (consolidated loop)

Combine lowercasing, punctuation trimming, length filtering, and stop-word
filtering into a single loop so ordering is unambiguous.

**Ordering:** lowercase → trim both ends → check length/stopword. Trimming both
ends (not just trailing) also merges wrapped tokens like `(prog)`, `"fresh"`,
and `g2)`.

```go
var stopWords = map[string]bool{
    "and": true, "the": true, "if": true, "do": true, "on": true,
    "to": true, "in": true, "of": true, "a": true, "is": true,
    "it": true, "no": true, "at": true, "we": true, "for": true,
    "th": true, "or": true, "be": true, "as": true, "by": true,
}

var resTokens []string
for _, raw := range result {
    token := strings.ToLower(raw)
    token = strings.Trim(token, " .,!?;:()[]{}\"'")
    if len(token) < 2 || stopWords[token] {
        continue
    }
    resTokens = append(resTokens, token)
}
```

**Tradeoff to accept:** `len(token) < 2` drops every bare number, so standalone
checkpoint/phase mentions (the lone `2`, `3`, `1`) disappear from the snapshot.
This is acceptable under the weekly-LLM model — numbered context is recovered
from raw text weekly — but it is a deliberate signal loss, not free.

**Minor collision note:** raidplan codes may be case-sensitive. The `ToLower`
pass can theoretically fold two distinct plans into one key. At current listing
volume the odds are negligible, so no action is required unless you want to
exempt code-shaped tokens from lowercasing.

---

## 4. Fix dedup so the corpus isn't corrupted (in `TokenizeListings`)

`PrevParsedPfIds` is in-memory and only one scrape deep. On any process restart
it re-`RPush`es every current listing's description into `descriptions:<day>`,
so the **weekly LLM analysis ingests duplicates** and token counts inflate. This
corrupts the corpus, not just the snapshot, so it's worth fixing.

Replace the in-memory `slices.Contains(t.PrevParsedPfIds, ...)` check with a
per-day Redis seen-set. `SAdd` returns the number of newly added members, so a
return of 0 means "already counted today" — even across restarts.

```go
seenKey := fmt.Sprintf("seen:%d", currentDayNumber)
seenExists, err := t.rdb.Exists(ctx, seenKey).Result()
if err != nil {
    panic(err)
}

for _, item := range scopedListings.Listings {
    added, err := t.rdb.SAdd(ctx, seenKey, item.Id).Result()
    if err != nil {
        panic(err)
    }
    if added == 0 {
        continue // already counted today
    }

    pfDescriptions = append(pfDescriptions, item.Description)
    tokenList, _ := splitListingIntoTokens(item.Description)
    for _, token := range tokenList {
        tokenMap[token] += 1
    }
}

if seenExists == 0 {
    t.rdb.Expire(ctx, seenKey, 24*32*time.Hour)
}
```

This de-dupes both the token counts and the description list in one move. The
`PrevParsedPfIds` field and its `parsedListingIds` tracking can be removed.

---

## 5. Post the denominator alongside the table

A raw `graven 77` is uninterpretable without knowing whether 90 or 900 listings
were scanned. Emit a "N listings scanned" line with the Discord post so the
3-hour snapshot is self-explanatory without waiting on the weekly pass.

Count distinct listings scanned (over the lookback window) and include it in the
output header. The exact wiring depends on the Discord-posting code, but the
count is already available as the size of the per-day `seen:<day>` set(s) or the
length of the gathered description list.

---

## Expected result

Before:
```
plan          125
https:        123
raidplan.io   121
graven         77
prog.          10
enrage.         4
w,              4
```

After (e.g. header: `312 listings scanned`):
```
graven              77
lpsbjecjdb0xoloz    54   ← codes consolidated, fragments merged
prog                52
fresh               45
p2                  30
cleanup             25
cw                  20
enrage               6
```
