package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/board"
	"github.com/Veraticus/findingway/internal/discord"
	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/Veraticus/findingway/internal/scraper"
	"github.com/Veraticus/findingway/internal/tokenizer"

	"github.com/bwmarrin/discordgo"
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

// The Listings whose descriptions the description log records.
var (
	trackedDuties      = []string{"Dancing Mad (Ultimate)"}
	trackedDataCentres = []string{"Aether", "Crystal", "Dynamis", "Primal"}
)

func main() {
	rdb, err := newRedisClient()
	if err != nil {
		panic(fmt.Errorf("could not configure Redis: %w", err))
	}
	defer func() { _ = rdb.Close() }()
	descriptionLog := tokenizer.NewDescriptionLog(rdb, trackedDuties, trackedDataCentres)

	if _, ok := os.LookupEnv("TOKENS_ONLY"); ok {
		if err := printTokens(os.Stdout, descriptionLog, time.Now()); err != nil {
			fmt.Printf("Error reading tokens: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := withRedisTimeout(func(ctx context.Context) error { return rdb.Ping(ctx).Err() }); err != nil {
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
		record: func(listings *ffxiv.Listings, now time.Time) error {
			return withRedisTimeout(func(ctx context.Context) error {
				return descriptionLog.Record(ctx, listings, now)
			})
		},
		postTokens: func(now time.Time) error {
			return postTokens(d, descriptionLog, now)
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

// reporter reads the description log's Report.
type reporter interface {
	Report(ctx context.Context, lookback int, now time.Time) (*tokenizer.Report, error)
}

// messenger posts to a Discord channel.
type messenger interface {
	Clean(channelID string) error
	Send(channelID string, message *discordgo.MessageSend) error
}

func printTokens(w io.Writer, descriptionLog reporter, now time.Time) error {
	return withRedisTimeout(func(ctx context.Context) error {
		report, err := descriptionLog.Report(ctx, lookback, now)
		if err != nil {
			return err
		}

		var out strings.Builder
		fmt.Fprintf(&out, "%d listings scanned over last %d days\n\n", report.ListingCount, report.Lookback)
		for _, t := range report.Tokens {
			fmt.Fprintf(&out, "%-30s %d\n", t.String, t.Count)
		}
		_, err = io.WriteString(w, out.String())
		return err
	})
}

func postTokens(d messenger, descriptionLog reporter, now time.Time) error {
	var report *tokenizer.Report
	err := withRedisTimeout(func(ctx context.Context) error {
		var err error
		report, err = descriptionLog.Report(ctx, lookback, now)
		return err
	})
	if err != nil {
		return err
	}

	if err := d.Clean(tokenChannelID); err != nil {
		fmt.Printf("Error cleaning token channel: %s\n", err)
	}
	if err := d.Send(tokenChannelID, tokensMessage(report, now)); err != nil {
		return err
	}
	if err := d.Send(tokenChannelID, descriptionCsvMessage(report.CSV, now)); err != nil {
		return fmt.Errorf("could not post csv: %w", err)
	}
	return nil
}
