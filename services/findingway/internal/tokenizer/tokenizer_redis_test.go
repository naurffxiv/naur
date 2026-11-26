package tokenizer

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestTokenizer(t *testing.T) *Tokenizer {
	t.Helper()
	tok, _ := newTestTokenizerWithServer(t)
	return tok
}

func newTestTokenizerWithServer(t *testing.T) (*Tokenizer, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	return &Tokenizer{rdb: newRedisClient(&redis.Options{Addr: mr.Addr()})}, mr
}

func silentServerAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var conns []net.Conn
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			conns = append(conns, conn)
			mu.Unlock()
		}
	}()

	t.Cleanup(func() {
		_ = ln.Close()
		mu.Lock()
		defer mu.Unlock()
		for _, conn := range conns {
			_ = conn.Close()
		}
	})
	return ln.Addr().String()
}

func dancingMadListing(id string, description string) *ffxiv.Listing {
	return &ffxiv.Listing{
		Id:          id,
		Duty:        "Dancing Mad (Ultimate)",
		DataCentre:  "Aether",
		Description: description,
	}
}

func TestTokenizeListingsSkipsListingSeenYesterday(t *testing.T) {
	tok := newTestTokenizer(t)
	ctx := context.Background()
	today := NowToInt()

	if err := tok.rdb.SAdd(ctx, fmt.Sprintf("seen:%d", today-1), "carryover").Err(); err != nil {
		t.Fatal(err)
	}

	err := tok.TokenizeListings(ctx, &ffxiv.Listings{Listings: []*ffxiv.Listing{
		dancingMadListing("carryover", "graven prog"),
		dancingMadListing("fresh", "enrage prog"),
	}})
	if err != nil {
		t.Fatal(err)
	}

	count, err := tok.GatherListingCount(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("GatherListingCount(2) = %d, want 2", count)
	}

	stored, err := tok.rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", today), 0, -1).Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 {
		t.Errorf("stored %d descriptions today, want 1: %v", len(stored), stored)
	}
}

func TestTokenizeListingsCountsListingOncePerDay(t *testing.T) {
	tok := newTestTokenizer(t)
	ctx := context.Background()
	listings := &ffxiv.Listings{Listings: []*ffxiv.Listing{dancingMadListing("repeat", "graven prog")}}

	for range 2 {
		if err := tok.TokenizeListings(ctx, listings); err != nil {
			t.Fatal(err)
		}
	}

	count, err := tok.GatherListingCount(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("GatherListingCount(1) = %d, want 1", count)
	}
}

func TestTokenizerReturnsErrorsWithinDeadlineWhenRedisIsUnresponsive(t *testing.T) {
	tok := &Tokenizer{rdb: newRedisClient(&redis.Options{Addr: silentServerAddr(t)})}
	listings := &ffxiv.Listings{Listings: []*ffxiv.Listing{
		dancingMadListing("one", "graven prog"),
		dancingMadListing("two", "enrage prog"),
	}}

	operations := map[string]func(ctx context.Context) error{
		"Ping": tok.Ping,
		"TokenizeListings": func(ctx context.Context) error {
			return tok.TokenizeListings(ctx, listings)
		},
		"GatherTokens": func(ctx context.Context) error {
			_, err := tok.GatherTokens(ctx, 2)
			return err
		},
		"GatherListingCount": func(ctx context.Context) error {
			_, err := tok.GatherListingCount(ctx, 2)
			return err
		},
		"CreateCsv": func(ctx context.Context) error {
			return tok.CreateCsv(ctx, 2, &bytes.Buffer{})
		},
	}

	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()

			start := time.Now()
			err := operation(ctx)
			elapsed := time.Since(start)

			if err == nil {
				t.Fatal("expected an error from an unresponsive Redis")
			}
			if elapsed > 2*time.Second {
				t.Errorf("returned after %v, want within the context deadline", elapsed)
			}
		})
	}
}

func TestInitReturnsErrorWithoutRedisCredentials(t *testing.T) {
	for _, name := range []string{"REDIS_USER", "REDIS_PASSWORD"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}

	tok := &Tokenizer{}
	if err := tok.Init(); err == nil {
		t.Fatal("expected an error when Redis credentials are missing")
	}
}

func TestTokenizeListingsSetsExpiryOnlyOnFirstWriteOfTheDay(t *testing.T) {
	tok, mr := newTestTokenizerWithServer(t)
	ctx := context.Background()
	today := NowToInt()
	seenKey := fmt.Sprintf("seen:%d", today)
	descriptionKey := fmt.Sprintf("descriptions:%d", today)
	ttl := 24 * 32 * time.Hour

	if err := tok.TokenizeListings(ctx, &ffxiv.Listings{Listings: []*ffxiv.Listing{
		dancingMadListing("first", "graven prog"),
	}}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{seenKey, descriptionKey} {
		if got := mr.TTL(key); got != ttl {
			t.Errorf("TTL(%s) after first write = %v, want %v", key, got, ttl)
		}
	}

	mr.FastForward(time.Hour)
	if err := tok.TokenizeListings(ctx, &ffxiv.Listings{Listings: []*ffxiv.Listing{
		dancingMadListing("second", "enrage prog"),
	}}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{seenKey, descriptionKey} {
		if got := mr.TTL(key); got != ttl-time.Hour {
			t.Errorf("TTL(%s) after second write = %v, want %v (not refreshed)", key, got, ttl-time.Hour)
		}
	}
}

func TestTokenizeListingsStoresEveryFreshDescription(t *testing.T) {
	tok := newTestTokenizer(t)
	ctx := context.Background()

	if err := tok.TokenizeListings(ctx, &ffxiv.Listings{Listings: []*ffxiv.Listing{
		dancingMadListing("one", "graven prog"),
		dancingMadListing("two", "enrage prog"),
		dancingMadListing("three", "p5 prog"),
	}}); err != nil {
		t.Fatal(err)
	}

	tokens, err := tok.GatherTokens(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0] != (Token{"prog", 3}) {
		t.Errorf("GatherTokens(1)[0] = %v, want prog counted once per stored description", tokens[0])
	}
}

// seedDescriptions stores entries for today and earlier days, plus one for
// tomorrow that a lookback must never include.
func seedDescriptions(t *testing.T, tok *Tokenizer) {
	t.Helper()
	ctx := context.Background()
	today := NowToInt()
	days := map[int][]string{
		today:     {"graven"},
		today - 1: {"enrage", "enrage", "enrage"},
		today + 1: {"future"},
	}
	for day, entries := range days {
		for _, entry := range entries {
			if err := tok.rdb.RPush(ctx, fmt.Sprintf("descriptions:%d", day), entry).Err(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestGatherTokensCountsEveryEntryInLookbackWindow(t *testing.T) {
	tok := newTestTokenizer(t)
	seedDescriptions(t, tok)

	tokens, err := tok.GatherTokens(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]int{}
	for _, token := range tokens {
		got[token.String] = token.Count
	}
	want := map[string]int{"graven": 1, "enrage": 3}
	if !maps.Equal(got, want) {
		t.Errorf("GatherTokens(2) = %v, want %v", got, want)
	}
	if tokens[0].String != "enrage" {
		t.Errorf("GatherTokens(2)[0] = %q, want most frequent token first", tokens[0].String)
	}
}

func TestCreateCsvWritesEveryEntryInLookbackWindow(t *testing.T) {
	tok := newTestTokenizer(t)
	seedDescriptions(t, tok)

	var buf bytes.Buffer
	if err := tok.CreateCsv(context.Background(), 2, &buf); err != nil {
		t.Fatal(err)
	}

	csv := buf.String()
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 5 {
		t.Fatalf("CreateCsv(2) wrote %d lines, want header + 4 rows:\n%s", len(lines), csv)
	}
	if strings.Contains(csv, "future") {
		t.Errorf("CreateCsv(2) included an entry from tomorrow:\n%s", csv)
	}
}
