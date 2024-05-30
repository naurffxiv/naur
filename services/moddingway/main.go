package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/naurffxiv/moddingway/internal/discord"
)

func main() {
	// TODO: change back to DISCORD_TOKEN before pushing
	discordToken, ok := os.LookupEnv("DISCORD_TOKEN2")
	if !ok {
		panic("You must supply a DISCORD_TOKEN to start!")
	}
	discordToken = strings.TrimSpace(discordToken)

	d := &discord.Discord{
		Token: discordToken,
	}

	fmt.Printf("Starting Discord...\n")
	err := d.Start()
	if err != nil {
		panic(fmt.Errorf("Could not instantiate Discord: %w", err))
	}
	defer d.Session.Close()
	start(d)
}

func start(d *discord.Discord) {
	d.Session.AddHandler(d.AddCommands)
	err := d.Session.Open()
	if err != nil {
		panic(fmt.Errorf("Could not open Discord session: %f", err))
	}

	d.Session.AddHandler(d.InteractionCreate)

	fmt.Println("Moddingway is ready. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	d.Session.Close()
}
