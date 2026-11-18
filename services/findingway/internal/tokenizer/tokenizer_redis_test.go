package tokenizer

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

const dancingMad = "Dancing Mad (Ultimate)"

var trackedDataCentres = []string{"Aether", "Crystal", "Dynamis", "Primal"}

// now is 2026-11-17 12:00 UTC, day 46373 counted from the 1900 epoch.
var now = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

const today = 46373

func newTestLog(t *testing.T) (*DescriptionLog, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), ContextTimeoutEnabled: true})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewDescriptionLog(rdb, []string{dancingMad}, trackedDataCentres), mr, rdb
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
		Duty:        dancingMad,
		DataCentre:  "Aether",
		Description: description,
	}
}

func listingsOf(ls ...*ffxiv.Listing) *ffxiv.Listings {
	return &ffxiv.Listings{Listings: ls}
}

func TestDayOfMatchesStoredKeyNumbering(t *testing.T) {
	tests := []struct {
		at   time.Time
		want int
	}{
		{now, today},
		{time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC), today},
		{time.Date(2026, 11, 16, 23, 59, 59, 0, time.UTC), today - 1},
		{time.Date(2026, 9, 25, 9, 25, 57, 0, time.UTC), 46320},
	}

	for _, tt := range tests {
		if got := dayOf(tt.at); got != tt.want {
			t.Errorf("dayOf(%v) = %d, want %d", tt.at, got, tt.want)
		}
	}
}

func TestRecordStoresTimestampedDescriptionsWith32DayExpiry(t *testing.T) {
	log, mr, rdb := newTestLog(t)
	ctx := context.Background()

	if err := log.Record(ctx, listingsOf(dancingMadListing("one", "graven prog")), now); err != nil {
		t.Fatal(err)
	}

	stored, err := rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", today), 0, -1).Result()
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%d\tgraven prog", now.Unix())
	if len(stored) != 1 || stored[0] != want {
		t.Errorf("descriptions:%d = %q, want [%q]", today, stored, want)
	}
	for _, key := range []string{fmt.Sprintf("seen:%d", today), fmt.Sprintf("descriptions:%d", today)} {
		if ttl := mr.TTL(key); ttl != 32*24*time.Hour {
			t.Errorf("TTL(%s) = %v, want %v", key, ttl, 32*24*time.Hour)
		}
	}
}

func TestRecordOnlyStoresTrackedDutiesAndDataCentres(t *testing.T) {
	log, _, rdb := newTestLog(t)
	ctx := context.Background()
	otherDuty := dancingMadListing("other-duty", "uwu prog")
	otherDuty.Duty = "The Weapon's Refrain (Ultimate)"
	otherDataCentre := dancingMadListing("other-dc", "jp prog")
	otherDataCentre.DataCentre = "Elemental"

	err := log.Record(ctx, listingsOf(otherDuty, otherDataCentre, dancingMadListing("tracked", "graven prog")), now)
	if err != nil {
		t.Fatal(err)
	}

	seen, err := rdb.SMembers(ctx, fmt.Sprintf("seen:%d", today)).Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != "tracked" {
		t.Errorf("seen:%d = %v, want [tracked]", today, seen)
	}
}

func TestRecordSkipsListingSeenYesterday(t *testing.T) {
	log, _, rdb := newTestLog(t)
	ctx := context.Background()

	if err := rdb.SAdd(ctx, fmt.Sprintf("seen:%d", today-1), "carryover").Err(); err != nil {
		t.Fatal(err)
	}

	err := log.Record(ctx, listingsOf(
		dancingMadListing("carryover", "graven prog"),
		dancingMadListing("fresh", "enrage prog"),
	), now)
	if err != nil {
		t.Fatal(err)
	}

	report, err := log.Report(ctx, 2, now)
	if err != nil {
		t.Fatal(err)
	}
	if report.ListingCount != 2 {
		t.Errorf("ListingCount = %d, want 2", report.ListingCount)
	}

	stored, err := rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", today), 0, -1).Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 {
		t.Errorf("stored %d descriptions today, want 1: %v", len(stored), stored)
	}
}

func TestRecordCountsListingOncePerDay(t *testing.T) {
	log, _, _ := newTestLog(t)
	ctx := context.Background()
	listings := listingsOf(dancingMadListing("repeat", "graven prog"))

	for _, at := range []time.Time{now, now.Add(3 * time.Minute)} {
		if err := log.Record(ctx, listings, at); err != nil {
			t.Fatal(err)
		}
	}

	report, err := log.Report(ctx, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if report.ListingCount != 1 {
		t.Errorf("ListingCount = %d, want 1", report.ListingCount)
	}
	if report.Lookback != 1 {
		t.Errorf("Lookback = %d, want 1", report.Lookback)
	}
}

func TestRecordStoresOnlyNewListingsFromAMixedBatchAndKeepsTheFirstExpiry(t *testing.T) {
	log, mr, rdb := newTestLog(t)
	ctx := context.Background()

	if err := rdb.SAdd(ctx, fmt.Sprintf("seen:%d", today-1), "carryover").Err(); err != nil {
		t.Fatal(err)
	}
	if err := log.Record(ctx, listingsOf(dancingMadListing("repeat", "first pass")), now); err != nil {
		t.Fatal(err)
	}
	mr.FastForward(time.Hour)

	err := log.Record(ctx, listingsOf(
		dancingMadListing("new-1", "graven prog"),
		dancingMadListing("carryover", "seen yesterday"),
		dancingMadListing("repeat", "seen today"),
		dancingMadListing("new-2", "enrage prog"),
	), now)
	if err != nil {
		t.Fatal(err)
	}

	stored, err := rdb.LRange(ctx, fmt.Sprintf("descriptions:%d", today), 0, -1).Result()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		fmt.Sprintf("%d\tfirst pass", now.Unix()),
		fmt.Sprintf("%d\tgraven prog", now.Unix()),
		fmt.Sprintf("%d\tenrage prog", now.Unix()),
	}
	if fmt.Sprint(stored) != fmt.Sprint(want) {
		t.Errorf("descriptions:%d = %q, want %q", today, stored, want)
	}
	for _, key := range []string{fmt.Sprintf("seen:%d", today), fmt.Sprintf("descriptions:%d", today)} {
		if ttl := mr.TTL(key); ttl != 32*24*time.Hour-time.Hour {
			t.Errorf("TTL(%s) = %v, want the first Record's expiry less an hour", key, ttl)
		}
	}
}

func TestRecordAcrossMidnight(t *testing.T) {
	log, _, rdb := newTestLog(t)
	ctx := context.Background()
	beforeMidnight := time.Date(2026, 11, 16, 23, 59, 0, 0, time.UTC)
	afterMidnight := time.Date(2026, 11, 17, 0, 1, 0, 0, time.UTC)

	if err := log.Record(ctx, listingsOf(dancingMadListing("late", "graven prog")), beforeMidnight); err != nil {
		t.Fatal(err)
	}
	err := log.Record(ctx, listingsOf(
		dancingMadListing("late", "graven prog"),
		dancingMadListing("early", "enrage prog"),
	), afterMidnight)
	if err != nil {
		t.Fatal(err)
	}

	for key, want := range map[string][]string{
		fmt.Sprintf("seen:%d", today-1): {"late"},
		fmt.Sprintf("seen:%d", today):   {"early"},
	} {
		got, err := rdb.SMembers(ctx, key).Result()
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) || got[0] != want[0] {
			t.Errorf("%s = %v, want %v", key, got, want)
		}
	}
	for key, want := range map[string]string{
		fmt.Sprintf("descriptions:%d", today-1): fmt.Sprintf("%d\tgraven prog", beforeMidnight.Unix()),
		fmt.Sprintf("descriptions:%d", today):   fmt.Sprintf("%d\tenrage prog", afterMidnight.Unix()),
	} {
		got, err := rdb.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s = %q, want [%q]", key, got, want)
		}
	}

	todayOnly, err := log.Report(ctx, 1, afterMidnight)
	if err != nil {
		t.Fatal(err)
	}
	if todayOnly.ListingCount != 1 {
		t.Errorf("Report(1).ListingCount = %d, want 1", todayOnly.ListingCount)
	}
	bothDays, err := log.Report(ctx, 2, afterMidnight)
	if err != nil {
		t.Fatal(err)
	}
	if bothDays.ListingCount != 2 {
		t.Errorf("Report(2).ListingCount = %d, want 2", bothDays.ListingCount)
	}
}

func TestReportGathersTokensCountAndCsv(t *testing.T) {
	log, _, rdb := newTestLog(t)
	ctx := context.Background()
	yesterday := now.Add(-24 * time.Hour)

	if err := log.Record(ctx, listingsOf(dancingMadListing("a", "graven prog")), yesterday); err != nil {
		t.Fatal(err)
	}
	if err := log.Record(ctx, listingsOf(dancingMadListing("b", "enrage prog\nreclear")), now); err != nil {
		t.Fatal(err)
	}
	// Legacy entries carry no timestamp; the CSV falls back to the day.
	if err := rdb.RPush(ctx, fmt.Sprintf("descriptions:%d", today), "prog legacy").Err(); err != nil {
		t.Fatal(err)
	}
	// Outside a lookback of 2.
	if err := log.Record(ctx, listingsOf(dancingMadListing("old", "ancient prog")), now.Add(-48*time.Hour)); err != nil {
		t.Fatal(err)
	}

	report, err := log.Report(ctx, 2, now)
	if err != nil {
		t.Fatal(err)
	}

	if report.ListingCount != 2 {
		t.Errorf("ListingCount = %d, want 2", report.ListingCount)
	}

	wantTokens := map[string]int{"prog": 3, "graven": 1, "enrage": 1, "reclear": 1, "legacy": 1}
	gotTokens := map[string]int{}
	for _, token := range report.Tokens {
		gotTokens[token.String] = token.Count
	}
	if fmt.Sprint(gotTokens) != fmt.Sprint(wantTokens) {
		t.Errorf("tokens = %v, want %v", gotTokens, wantTokens)
	}
	if report.Tokens[0] != (Token{"prog", 3}) {
		t.Errorf("first token = %v, want the most frequent {prog 3}", report.Tokens[0])
	}

	wantCsv := "Timestamp,Description\n" +
		now.Format(time.DateTime) + ",enrage progreclear\n" +
		"2026-11-17,prog legacy\n" +
		yesterday.Format(time.DateTime) + ",graven prog\n"
	if string(report.CSV) != wantCsv {
		t.Errorf("CSV =\n%s\nwant\n%s", report.CSV, wantCsv)
	}
}

func TestDescriptionLogReturnsErrorsWithinDeadlineWhenRedisIsUnresponsive(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: silentServerAddr(t), ContextTimeoutEnabled: true})
	t.Cleanup(func() { _ = rdb.Close() })
	log := NewDescriptionLog(rdb, []string{dancingMad}, trackedDataCentres)
	listings := listingsOf(
		dancingMadListing("one", "graven prog"),
		dancingMadListing("two", "enrage prog"),
	)

	operations := map[string]func(ctx context.Context) error{
		"Record": func(ctx context.Context) error {
			return log.Record(ctx, listings, now)
		},
		"Report": func(ctx context.Context) error {
			_, err := log.Report(ctx, 2, now)
			return err
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
