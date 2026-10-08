package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/discord"
	"github.com/Veraticus/findingway/internal/scraper"
	"github.com/Veraticus/findingway/internal/tokenizer"

	"gopkg.in/yaml.v2"
)

const (
	lookback       = 2
	redisTimeout   = 10 * time.Second
	tokenChannelId = "1510722864851189981"
)

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

	config, err := os.ReadFile("./config.yaml")
	if err != nil {
		panic(fmt.Errorf("could not read config.yaml: %w", err))
	}
	if err := yaml.Unmarshal(config, &d); err != nil {
		panic(err)
	}

	err = d.Start()
	defer func() { _ = d.Session.Close() }()
	if err != nil {
		panic(fmt.Errorf("could not instantiate Discord: %w", err))
	}

	scraper := &scraper.Scraper{Url: "https://xivpf.com"}

	fmt.Printf("Starting findingway...\n")
	loopCount := 0
	for {
		totalWait := 3 * time.Minute
		fmt.Printf("Scraping source...\n")
		listings, err := scraper.Scrape()
		if err != nil {
			fmt.Printf("Scraper error: %v\n", err)
			continue
		}
		fmt.Printf("Got %v listings.\n", len(listings.Listings))
		fmt.Printf("Sending to %v channels...\n", len(d.Channels))

		for _, c := range d.Channels {
			startTime := time.Now()
			fmt.Printf("Cleaning Discord for %v (%v)...\n", c.Name, c.Duty)
			err = d.CleanChannel(c.ID)
			if err != nil {
				fmt.Printf("Discord error cleaning channel: %v\n", err)
			}

			fmt.Printf("Updating Discord for %v (%v)...\n", c.Name, c.Duty)
			for _, dataCentre := range c.DataCentres {
				err = d.PostListings(c.ID, listings, c.Duty, dataCentre)
			}
			if err != nil {
				fmt.Printf("Discord error updating messages: %v\n", err)
			}
			endTime := time.Now()
			duration := endTime.Sub(startTime)
			totalWait -= duration
		}

		err = withRedisTimeout(func(ctx context.Context) error {
			return tok.TokenizeListings(ctx, listings)
		})
		if err != nil {
			fmt.Printf("Error storing tokens: %v\n", err)
		}

		// Output values every 1 hours
		if loopCount%20 == 0 {
			fmt.Println("Sending tokens to discord")
			if err := postTokens(d, tok); err != nil {
				fmt.Printf("Error posting tokens: %v\n", err)
			}
		}

		if once != "false" {
			os.Exit(0)
		}
		fmt.Printf("Sleeping for %v...\n", totalWait)
		loopCount += 1
		time.Sleep(totalWait)
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

func postTokens(d *discord.Discord, tok *tokenizer.Tokenizer) error {
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

	if err := d.CleanChannel(tokenChannelId); err != nil {
		fmt.Printf("Error cleaning token channel: %s\n", err)
	}
	if err := d.PostTokens(tokenChannelId, tokens, lookback, listingCount); err != nil {
		return err
	}
	if err := d.PostDescriptionCsv(tokenChannelId, &buf); err != nil {
		return fmt.Errorf("could not post csv: %w", err)
	}
	return nil
}
