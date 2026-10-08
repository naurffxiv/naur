package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/discord"
	"github.com/Veraticus/findingway/internal/scraper"
	"github.com/Veraticus/findingway/internal/tokenizer"

	"gopkg.in/yaml.v2"
)

func main() {
	const lookback = 2

	tok := &tokenizer.Tokenizer{}
	tok.Init()

	if _, ok := os.LookupEnv("TOKENS_ONLY"); ok {
		tokens := tok.GatherTokens(lookback)
		count := tok.GatherListingCount(2)
		fmt.Printf("%d listings scanned over last %d days\n\n", count, lookback)
		for _, t := range tokens {
			fmt.Printf("%-30s %d\n", t.String, t.Count)
		}
		return
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
	tokenizer := tok

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

		tokenizer.TokenizeListings(listings)

		// Output values every 1 hours
		if loopCount%20 == 0 {
			fmt.Println("Sending tokens to discord")

			err = d.CleanChannel("1510722864851189981")
			if err != nil {
				fmt.Printf("Error cleaning token channel: %s\n", err)
			}

			tokens := tokenizer.GatherTokens(lookback)
			listingCount := tokenizer.GatherListingCount(lookback)
			err = d.PostTokens("1510722864851189981", tokens, listingCount)
			if err != nil {
				fmt.Printf("Error posting tokens: %s\n", err)
			}

			// csv
			var buf bytes.Buffer
			tokenizer.CreateCsv(lookback, &buf)
			err = d.PostDescriptionCsv("1510722864851189981", &buf)
			if err != nil {
				fmt.Printf("Error posting csv: %s\n", err)
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
