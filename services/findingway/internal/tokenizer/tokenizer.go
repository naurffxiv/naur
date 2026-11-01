package tokenizer

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"errors"
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

	result := splitRe.Split(listing, -1)

	var resTokens []string
	for _, raw := range result {
		if token := normalizeToken(raw); token != "" {
			resTokens = append(resTokens, token)
		}
	}

	return resTokens, nil
}

func newRedisClient(opts *redis.Options) *redis.Client {
	opts.ContextTimeoutEnabled = true
	return redis.NewClient(opts)
}

func (t *Tokenizer) Init() error {
	redisPw, ok := os.LookupEnv("REDIS_PASSWORD")
	redisUser, userOk := os.LookupEnv("REDIS_USER")

	if !userOk {
		return errors.New("REDIS_USER is not set")
	}
	if !ok {
		return errors.New("REDIS_PASSWORD is not set")
	}

	cert, err := tls.LoadX509KeyPair("naur.crt", "naur.key")
	if err != nil {
		return fmt.Errorf("error loading certificates: %w", err)
	}

	caCert, err := os.ReadFile("hyddwn-ca.crt")
	if err != nil {
		return fmt.Errorf("error loading CA certificate: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return errors.New("failed to append CA certificate")
	}

	t.rdb = newRedisClient(&redis.Options{
		Addr:     "redis.hyddwn.net:6380",
		Username: redisUser,
		Password: redisPw,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caPool,
		},
	})
	return nil
}

func (t *Tokenizer) Ping(ctx context.Context) error {
	return t.rdb.Ping(ctx).Err()
}

func (t *Tokenizer) TokenizeListings(ctx context.Context, listings *ffxiv.Listings) error {
	currentDayNumber := NowToInt()

	pfDescriptions := []string{}

	scopedListings := listings.ForDutiesAndDataCentres(
		[]string{"Dancing Mad (Ultimate)"},
		[]string{"Aether", "Crystal", "Dynamis", "Primal"})

	seenKey := fmt.Sprintf("seen:%d", currentDayNumber)
	prevSeenKey := fmt.Sprintf("seen:%d", currentDayNumber-1)
	seenExists, err := t.rdb.Exists(ctx, seenKey).Result()
	if err != nil {
		return err
	}

	var added int64
	for _, item := range scopedListings.Listings {
		seenYesterday, err := t.rdb.SIsMember(ctx, prevSeenKey, item.Id).Result()
		if err != nil {
			return err
		}
		if seenYesterday {
			continue
		}

		added, err = t.rdb.SAdd(ctx, seenKey, item.Id).Result()
		if err != nil {
			return err
		}
		if added == 0 {
			continue // already counted today
		}

		pfDescriptions = append(pfDescriptions, item.Description)
	}

	if seenExists == 0 {
		if err := t.rdb.Expire(ctx, seenKey, 24*32*time.Hour).Err(); err != nil {
			return err
		}
	}

	// Store description list into redis
	descriptionKey := fmt.Sprintf("descriptions:%d", currentDayNumber)
	todayExists, err := t.rdb.Exists(ctx, descriptionKey).Result()
	if err != nil {
		return err
	}

	for _, description := range pfDescriptions {
		entry := fmt.Sprintf("%d\t%s", time.Now().Unix(), description)
		err = t.rdb.RPush(ctx, descriptionKey, entry).Err()
		if err != nil {
			return err
		}
	}

	if todayExists == 0 {
		if err := t.rdb.Expire(ctx, descriptionKey, 24*32*time.Hour).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tokenizer) GatherTokens(ctx context.Context, lookback int) ([]Token, error) {

	tokenSum := make(map[string]int)

	todayDayNumber := NowToInt()

	for i := range lookback {
		prevDayNumber := todayDayNumber - i

		descriptions, err := t.rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", prevDayNumber), 0, -1).Result()
		if err != nil {
			return nil, err
		}

		for _, entry := range descriptions {
			_, description := parseEntry(entry, prevDayNumber)
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

	return res, nil
}

func (t *Tokenizer) GatherListingCount(ctx context.Context, lookback int) (int64, error) {
	todayDayNumber := NowToInt()
	var total int64
	for i := range lookback {
		count, err := t.rdb.SCard(ctx, fmt.Sprintf("seen:%d", todayDayNumber-i)).Result()
		if err != nil {
			return 0, err
		}
		total += count
	}
	return total, nil
}

func (t *Tokenizer) CreateCsv(ctx context.Context, lookback int, buf *bytes.Buffer) error {

	todayDayNumber := NowToInt()
	csvwriter := csv.NewWriter(buf)

	err := csvwriter.Write([]string{"Timestamp", "Description"})
	if err != nil {
		return err
	}

	for i := range lookback {
		prevDayNumber := todayDayNumber - i

		getResult, err := t.rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", prevDayNumber), 0, -1).Result()

		if err != nil {
			return err
		}

		for _, entry := range getResult {
			timestampStr, description := parseEntry(entry, prevDayNumber)
			err = csvwriter.Write([]string{timestampStr, strings.ReplaceAll(description, "\n", "")})
			if err != nil {
				return err
			}
		}
	}

	csvwriter.Flush()
	return csvwriter.Error()
}
