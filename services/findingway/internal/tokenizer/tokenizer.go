// Package tokenizer keeps the description log: the descriptions of tracked
// Listings stored per day in Redis, and the tokens reported from them.
package tokenizer

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/redis/go-redis/v9"
)

type Token struct {
	String string
	Count  int
}

var (
	raidplanRe = regexp.MustCompile(`(?:https?://)?raidplan\.io/plan/([^#\s]+)(?:#\S+)?`)
	httpUrlRe  = regexp.MustCompile(`https?://\S+`)
	bareUrlRe  = regexp.MustCompile(`\b\w[\w.-]+\.[a-z]{2,}/\S*`)
	splitRe    = regexp.MustCompile(strings.Join([]string{
		`\|`,   // |
		`\|\|`, // ||
		`\/`,   // /
		`\/\/`, // //
		`->`,   // ->
		`for `, // for
		` `,    // space
		`to `,  // to
		`&`,    // &
		`\+`,   // +
		"\n",   // newline
	}, `|`))
)

// urlToToken extracts the most meaningful token from a URL:
// - a path segment if present (e.g. pastebin.com/7fs57PyQ → 7fs57PyQ)
// - the hostname with dots stripped if there is no meaningful path (e.g. kefkab.in/ → kefkabin)
func urlToToken(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimRight(url, ".,!?;:)>")

	slashIdx := strings.IndexByte(url, '/')
	if slashIdx < 0 {
		return strings.ReplaceAll(url, ".", "")
	}

	hostname := url[:slashIdx]
	path := url[slashIdx+1:]

	// Strip fragment
	if i := strings.IndexByte(path, '#'); i >= 0 {
		path = path[:i]
	}
	// Take first path segment only, skipping a generic "plan" segment (e.g. raidplan.io/plan/<code>)
	segments := strings.Split(path, "/")
	path = segments[0]
	if path == "plan" && len(segments) > 1 {
		path = segments[1]
	}

	if len(path) >= 2 {
		return path
	}
	return strings.ReplaceAll(hostname, ".", "")
}

var stopWords = map[string]bool{
	"and": true, "the": true, "if": true, "do": true, "on": true,
	"to": true, "in": true, "of": true, "a": true, "is": true,
	"it": true, "no": true, "at": true, "we": true, "for": true,
	"th": true, "or": true, "be": true, "as": true, "by": true,
	// pure english filler
	"not": true, "with": true, "this": true, "have": true, "lets": true, "are": true,
	"let's": true, "please": true, "but": true, "can": true, "some": true,
	"get": true, "out": true, "come": true, "time": true,
	// vague action words
	"doing": true, "trying": true, "tell": true, "retell": true,
	"need": true, "solve": true,
}

// normalizeToken lowercases, trims punctuation, and returns an empty string
// if the token should be discarded (too short, a stop word, or URL debris).
func normalizeToken(raw string) string {
	// Drop URL debris from old Redis data (https:, raidplan.io, pastebin.com, etc.)
	if httpUrlRe.MatchString(raw) || bareUrlRe.MatchString(raw) {
		return ""
	}
	token := strings.ToLower(raw)
	token = strings.Trim(token, " .,!?;:()[]{}'\"`#")
	if len(token) < 2 || stopWords[token] {
		return ""
	}
	return token
}

// parseEntry splits a stored Redis entry into its timestamp and description.
// New entries are stored as "<unix_ts>\t<description>"; old entries are bare descriptions.
func parseEntry(entry string, fallbackDayNumber int) (timestamp string, description string) {
	if idx := strings.IndexByte(entry, '\t'); idx >= 0 {
		unix, err := strconv.ParseInt(entry[:idx], 10, 64)
		if err == nil {
			return time.Unix(unix, 0).UTC().Format(time.DateTime), entry[idx+1:]
		}
	}
	// Fallback for old data without a timestamp
	day := time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC).AddDate(0, 0, fallbackDayNumber)
	return day.Format(time.DateOnly), entry
}

// dayOf numbers the day containing now, counted from the same 1900 epoch as
// the stored keys (seen:<day>, descriptions:<day>). Days roll over at UTC
// midnight.
func dayOf(now time.Time) int {
	startDate := time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC)
	dayCount := now.Sub(startDate).Hours() / 24
	return int(dayCount)
}

func splitListingIntoTokens(listing string) ([]string, error) {
	// Extract raidplan code as a bare token, drop the rest of the URL.
	// Scheme is optional to catch bare raidplan.io/plan/CODE links.
	listing = raidplanRe.ReplaceAllStringFunc(listing, func(m string) string {
		return raidplanRe.FindStringSubmatch(m)[1]
	})
	// For https:// URLs, extract path code or hostname (pastebin, tinyurl, kefkab.in, etc.)
	listing = httpUrlRe.ReplaceAllStringFunc(listing, urlToToken)
	// Same for bare domain links without a scheme
	listing = bareUrlRe.ReplaceAllStringFunc(listing, urlToToken)

	result := splitRe.Split(listing, -1)

	var resTokens []string
	for _, raw := range result {
		if token := normalizeToken(raw); token != "" {
			resTokens = append(resTokens, token)
		}
	}

	return resTokens, nil
}

// keyExpiry is how long each day's keys are kept.
const keyExpiry = 24 * 32 * time.Hour

func seenKey(day int) string         { return fmt.Sprintf("seen:%d", day) }
func descriptionsKey(day int) string { return fmt.Sprintf("descriptions:%d", day) }

// DescriptionLog records the descriptions of newly seen Listings for the
// tracked duties and DataCentres, one Redis set and list per day, and
// reports on them.
type DescriptionLog struct {
	rdb         redis.UniversalClient
	duties      []string
	dataCentres []string
}

// NewDescriptionLog returns a DescriptionLog over rdb that records Listings
// for duties in dataCentres.
func NewDescriptionLog(rdb redis.UniversalClient, duties []string, dataCentres []string) *DescriptionLog {
	return &DescriptionLog{rdb: rdb, duties: duties, dataCentres: dataCentres}
}

// Report is what the DescriptionLog holds over a lookback window.
type Report struct {
	// Lookback is how many days, ending with the day reported on, the
	// Report covers.
	Lookback int
	// Tokens counts each token across the window's descriptions, most
	// frequent first.
	Tokens []Token
	// ListingCount is how many Listings were seen in the window.
	ListingCount int64
	// CSV lists each description with its timestamp.
	CSV []byte
}

// Record stores the description of each tracked Listing not already seen
// today or yesterday, as of now. It makes three Redis round trips however
// many Listings there are.
func (l *DescriptionLog) Record(ctx context.Context, listings *ffxiv.Listings, now time.Time) error {
	today := dayOf(now)
	todaySeenKey := seenKey(today)
	prevSeenKey := seenKey(today - 1)
	descriptionKey := descriptionsKey(today)
	scoped := listings.ForDutiesAndDataCentres(l.duties, l.dataCentres).Listings

	// Whether today's keys exist yet, and which Listings were seen yesterday.
	var seenExists, descriptionsExist *redis.IntCmd
	seenYesterday := make([]*redis.BoolCmd, len(scoped))
	if _, err := l.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		seenExists = pipe.Exists(ctx, todaySeenKey)
		descriptionsExist = pipe.Exists(ctx, descriptionKey)
		for i, item := range scoped {
			seenYesterday[i] = pipe.SIsMember(ctx, prevSeenKey, item.Id)
		}
		return nil
	}); err != nil {
		return err
	}

	// Mark the rest seen today; SAdd adds nothing for a Listing already
	// counted today.
	var unseen []*ffxiv.Listing
	var added []*redis.IntCmd
	if _, err := l.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, item := range scoped {
			if seenYesterday[i].Val() {
				continue
			}
			unseen = append(unseen, item)
			added = append(added, pipe.SAdd(ctx, todaySeenKey, item.Id))
		}
		if seenExists.Val() == 0 {
			pipe.Expire(ctx, todaySeenKey, keyExpiry)
		}
		return nil
	}); err != nil {
		return err
	}

	// Store the new Listings' descriptions.
	_, err := l.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, item := range unseen {
			if added[i].Val() == 0 {
				continue
			}
			pipe.RPush(ctx, descriptionKey, fmt.Sprintf("%d\t%s", now.Unix(), item.Description))
		}
		if descriptionsExist.Val() == 0 {
			pipe.Expire(ctx, descriptionKey, keyExpiry)
		}
		return nil
	})
	return err
}

// Report gathers the tokens, Listing count and CSV for the lookback days
// ending with the day containing now, reading every day in one round trip.
func (l *DescriptionLog) Report(ctx context.Context, lookback int, now time.Time) (*Report, error) {
	today := dayOf(now)
	report := &Report{Lookback: lookback}
	tokenSum := make(map[string]int)

	var buf bytes.Buffer
	csvwriter := csv.NewWriter(&buf)
	if err := csvwriter.Write([]string{"Timestamp", "Description"}); err != nil {
		return nil, err
	}

	// Each day's seen count and descriptions, in one round trip.
	counts := make([]*redis.IntCmd, lookback)
	descriptions := make([]*redis.StringSliceCmd, lookback)
	if _, err := l.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i := range lookback {
			counts[i] = pipe.SCard(ctx, seenKey(today-i))
			descriptions[i] = pipe.LRange(ctx, descriptionsKey(today-i), 0, -1)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	for i := range lookback {
		day := today - i
		report.ListingCount += counts[i].Val()

		for _, entry := range descriptions[i].Val() {
			timestamp, description := parseEntry(entry, day)

			tokens, _ := splitListingIntoTokens(description)
			for _, token := range tokens {
				tokenSum[token] += 1
			}

			if err := csvwriter.Write([]string{timestamp, strings.ReplaceAll(description, "\n", "")}); err != nil {
				return nil, err
			}
		}
	}

	csvwriter.Flush()
	if err := csvwriter.Error(); err != nil {
		return nil, err
	}
	report.CSV = buf.Bytes()

	for k, v := range tokenSum {
		report.Tokens = append(report.Tokens, Token{k, v})
	}
	sort.Slice(report.Tokens, func(i, j int) bool {
		return report.Tokens[i].Count > report.Tokens[j].Count
	})

	return report, nil
}
