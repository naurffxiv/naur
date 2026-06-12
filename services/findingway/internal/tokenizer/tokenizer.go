package tokenizer

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/redis/go-redis/v9"
)

type Tokenizer struct {
	rdb *redis.Client
}

type Token struct {
	String string
	Count  int
}

var (
	raidplanRe = regexp.MustCompile(`(?:https?://)?raidplan\.io/plan/([^#\s]+)(?:#\S+)?`)
	httpUrlRe  = regexp.MustCompile(`https?://\S+`)
	bareUrlRe  = regexp.MustCompile(`\b\w[\w.-]+\.[a-z]{2,}/\S*`)
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
	// Take first path segment only
	if i := strings.IndexByte(path, '/'); i >= 0 {
		path = path[:i]
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
	"not": true, "with": true, "this": true, "have": true, "lets": true,
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
// Returns a zero Time for old entries so callers can filter them out.
func parseEntry(entry string, fallbackDayNumber int) (ts time.Time, timestamp string, description string) {
	if idx := strings.IndexByte(entry, '\t'); idx >= 0 {
		unix, err := strconv.ParseInt(entry[:idx], 10, 64)
		if err == nil {
			t := time.Unix(unix, 0).UTC()
			return t, t.Format(time.DateTime), entry[idx+1:]
		}
	}
	// Fallback for old data without a timestamp
	day := time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC).AddDate(0, 0, fallbackDayNumber)
	return time.Time{}, day.Format(time.DateOnly), entry
}

func NowToInt() int {
	startDate := time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC)
	dayCount := time.Since(startDate).Hours() / 24
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

	var splitRegex = []string{
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
	}

	fullRegex := strings.Join(splitRegex, `|`)
	re := regexp.MustCompile(fullRegex)
	result := re.Split(listing, -1)

	var resTokens []string
	for _, raw := range result {
		if token := normalizeToken(raw); token != "" {
			resTokens = append(resTokens, token)
		}
	}

	return resTokens, nil
}

func (t *Tokenizer) Init() {
	redisPw, ok := os.LookupEnv("REDIS_PASSWORD")
	redisUser, userOk := os.LookupEnv("REDIS_USER")

	if !userOk {
		panic("You must supply a REDIS_USER to start!")
	}
	if !ok {
		panic("You must supply a REDIS_PASSWORD to start!")
	}

	cert, err := tls.LoadX509KeyPair("naur.crt", "naur.key")
	if err != nil {
		panic(fmt.Errorf("Error loading certificates %s", err))
	}

	caCert, err := os.ReadFile("hyddwn-ca.crt")
	if err != nil {
		panic(fmt.Errorf("Error loading CA certificate %s", err))
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		panic("Failed to append CA certificate")
	}

	t.rdb = redis.NewClient(&redis.Options{
		Addr:     "redis.hyddwn.net:6380",
		Username: redisUser,
		Password: redisPw,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caPool,
		},
	})

	ctx := context.Background()
	_, err = t.rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}

func (t *Tokenizer) TokenizeListings(listings *ffxiv.Listings) {
	currentDayNumber := NowToInt()
	ctx := context.Background()

	pfDescriptions := []string{}

	scopedListings := listings.ForDutiesAndDataCentres(
		[]string{"HighEndDuty", "Dancing Mad (Ultimate)"},
		[]string{"Aether", "Crystal", "Dynamis", "Primal"})

	seenKey := fmt.Sprintf("seen:%d", currentDayNumber)
	seenExists, err := t.rdb.Exists(ctx, seenKey).Result()
	if err != nil {
		panic(err)
	}

	var added int64
	for _, item := range scopedListings.Listings {
		added, err = t.rdb.SAdd(ctx, seenKey, item.Id).Result()
		if err != nil {
			panic(err)
		}
		if added == 0 {
			continue // already counted today
		}

		pfDescriptions = append(pfDescriptions, item.Description)
	}

	if seenExists == 0 {
		t.rdb.Expire(ctx, seenKey, 24*32*time.Hour)
	}

	// Store description list into redis
	descriptionKey := fmt.Sprintf("descriptions:%d", currentDayNumber)
	todayExists, err := t.rdb.Exists(ctx, descriptionKey).Result()
	if err != nil {
		panic(err)
	}

	for _, description := range pfDescriptions {
		entry := fmt.Sprintf("%d\t%s", time.Now().Unix(), description)
		err = t.rdb.RPush(ctx, descriptionKey, entry).Err()
		if err != nil {
			panic(err)
		}
	}

	if todayExists == 0 {
		t.rdb.Expire(ctx, descriptionKey, 24*32*time.Hour)
	}
}

func (t *Tokenizer) GatherTokens(lookback int) []Token {

	tokenSum := make(map[string]int)
	ctx := context.Background()

	todayDayNumber := NowToInt()

	for i := range lookback {
		prevDayNumber := todayDayNumber - i

		descriptions, err := t.rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", prevDayNumber), 0, -1).Result()
		if err != nil {
			panic(err)
		}

		for _, entry := range descriptions {
			_, _, description := parseEntry(entry, prevDayNumber)
			tokens, _ := splitListingIntoTokens(description)
			for _, token := range tokens {
				tokenSum[token] += 1
			}
		}
	}

	var res []Token

	for k, v := range tokenSum {
		res = append(res, Token{k, v})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Count > res[j].Count
	})

	return res
}

func (t *Tokenizer) GatherListingCount(lookback int) int64 {
	ctx := context.Background()
	todayDayNumber := NowToInt()
	var total int64
	for i := range lookback {
		count, err := t.rdb.SCard(ctx, fmt.Sprintf("seen:%d", todayDayNumber-i)).Result()
		if err != nil {
			panic(err)
		}
		total += count
	}
	return total
}

func (t *Tokenizer) CreateCsv(lookback int, buf *bytes.Buffer) {

	todayDayNumber := NowToInt()
	csvwriter := csv.NewWriter(buf)

	err := csvwriter.Write([]string{"Timestamp", "Description"})
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	for i := range lookback {
		prevDayNumber := todayDayNumber - i

		getResult, err := t.rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", prevDayNumber), 0, -1).Result()

		if err != nil {
			panic(err)
		}

		for _, entry := range getResult {
			_, timestampStr, description := parseEntry(entry, prevDayNumber)
			err = csvwriter.Write([]string{timestampStr, strings.ReplaceAll(description, "\n", "")})
			if err != nil {
				panic(err)
			}
		}
	}

	csvwriter.Flush()
}
