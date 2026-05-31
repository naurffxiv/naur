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
	tokenizer := &tokenizer.Tokenizer{}
	tokenizer.Init()

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

		// Output values every 3 hours
		if loopCount%60 == 0 {
			fmt.Println("Sending tokens to discord")

			tokens := tokenizer.GatherTokens(7)
			err = d.PostTokens("TODO", tokens)
			if err != nil {
				fmt.Printf("Error posting tokens: %s\n", err)
			}

			// csv
			var buf bytes.Buffer
			tokenizer.CreateCsv(7, &buf)
			err = d.PostDescriptionCsv("TODO", &buf)
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
