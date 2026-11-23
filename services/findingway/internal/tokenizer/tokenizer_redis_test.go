package tokenizer

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestTokenizer(t *testing.T) *Tokenizer {
	t.Helper()
	mr := miniredis.RunT(t)
	return &Tokenizer{rdb: newRedisClient(&redis.Options{Addr: mr.Addr()})}
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
