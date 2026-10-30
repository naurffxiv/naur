package tokenizer

import (
	"context"
	"fmt"
	"testing"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestTokenizer(t *testing.T) *Tokenizer {
	t.Helper()
	mr := miniredis.RunT(t)
	return &Tokenizer{rdb: redis.NewClient(&redis.Options{Addr: mr.Addr()})}
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

	tok.TokenizeListings(&ffxiv.Listings{Listings: []*ffxiv.Listing{
		dancingMadListing("carryover", "graven prog"),
		dancingMadListing("fresh", "enrage prog"),
	}})

	if got := tok.GatherListingCount(2); got != 2 {
		t.Errorf("GatherListingCount(2) = %d, want 2", got)
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
	listings := &ffxiv.Listings{Listings: []*ffxiv.Listing{dancingMadListing("repeat", "graven prog")}}

	tok.TokenizeListings(listings)
	tok.TokenizeListings(listings)

	if got := tok.GatherListingCount(1); got != 1 {
		t.Errorf("GatherListingCount(1) = %d, want 1", got)
	}
}
