package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/board"
	"github.com/Veraticus/findingway/internal/discord"
	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/Veraticus/findingway/internal/scraper"
	"github.com/Veraticus/findingway/internal/tokenizer"

	"gopkg.in/yaml.v2"
)

const (
	lookback       = 2
	redisTimeout   = 10 * time.Second
	tokenChannelID = "1510722864851189981"
)

// config is the shape of ./config.yaml.
type config struct {
	Channels []*board.Channel `yaml:"channels"`
}

func main() {
	tok := &tokenizer.Tokenizer{}
	if err := tok.Init(); err != nil {
		panic(fmt.Errorf("could not configure the tokenizer: %w", err))
	}

	if _, ok := os.LookupEnv("TOKENS_ONLY"); ok {
		if err := printTokens(tok); err != nil {
			fmt.Printf("Error reading tokens: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := withRedisTimeout(tok.Ping); err != nil {
		fmt.Printf("Redis unreachable at startup, continuing without it: %v\n", err)
	}

	discordToken, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		panic("You must supply a DISCORD_TOKEN to start!")
	}
	once, ok := os.LookupEnv("ONCE")
	if !ok {
		once = "false"
	}
	discordToken = strings.TrimSpace(discordToken)

	d := &discord.Discord{
		Token: discordToken,
	}

	configFile, err := os.ReadFile("./config.yaml")
	if err != nil {
		panic(fmt.Errorf("could not read config.yaml: %w", err))
	}
	var cfg config
	if err := yaml.Unmarshal(configFile, &cfg); err != nil {
		panic(err)
	}

	err = d.Start()
	defer func() { _ = d.Session.Close() }()
	if err != nil {
		panic(fmt.Errorf("could not instantiate Discord: %w", err))
	}

	scraper := &scraper.Scraper{Url: "https://xivpf.com"}
	listingBoard := board.New(d)

	c := &cycle{
		scrape:   scraper.Scrape,
		channels: cfg.Channels,
		refresh:  listingBoard.Refresh,
		record: func(listings *ffxiv.Listings, _ time.Time) error {
			return withRedisTimeout(func(ctx context.Context) error {
				return tok.TokenizeListings(ctx, listings)
			})
		},
		postTokens: func(now time.Time) error {
			return postTokens(d, tok, now)
		},
		now:  time.Now,
		once: once != "false",
	}

	fmt.Printf("Starting findingway...\n")
	for {
		wait, done := c.run()
		if done {
			return
		}
		fmt.Printf("Sleeping for %v...\n", wait)
		time.Sleep(wait)
	}
}

func withRedisTimeout(operation func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	return operation(ctx)
}

func printTokens(tok *tokenizer.Tokenizer) error {
	return withRedisTimeout(func(ctx context.Context) error {
		tokens, err := tok.GatherTokens(ctx, lookback)
		if err != nil {
			return err
		}
		count, err := tok.GatherListingCount(ctx, lookback)
		if err != nil {
			return err
		}

		fmt.Printf("%d listings scanned over last %d days\n\n", count, lookback)
		for _, t := range tokens {
			fmt.Printf("%-30s %d\n", t.String, t.Count)
		}
		return nil
	})
}

func postTokens(d *discord.Discord, tok *tokenizer.Tokenizer, now time.Time) error {
	var tokens []tokenizer.Token
	var listingCount int64
	var buf bytes.Buffer
	err := withRedisTimeout(func(ctx context.Context) error {
		var err error
		if tokens, err = tok.GatherTokens(ctx, lookback); err != nil {
			return err
		}
		if listingCount, err = tok.GatherListingCount(ctx, lookback); err != nil {
			return err
		}
		return tok.CreateCsv(ctx, lookback, &buf)
	})
	if err != nil {
		return err
	}

	if err := d.Clean(tokenChannelID); err != nil {
		fmt.Printf("Error cleaning token channel: %s\n", err)
	}
	if err := d.Send(tokenChannelID, discord.TokensMessage(tokens, lookback, listingCount, now)); err != nil {
		return err
	}
	if err := d.Send(tokenChannelID, discord.DescriptionCsvMessage(&buf, now)); err != nil {
		return fmt.Errorf("could not post csv: %w", err)
	}
	return nil
}
