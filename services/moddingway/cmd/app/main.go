package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/naurffxiv/moddingway/internal/database"
	"github.com/naurffxiv/moddingway/internal/discord"
	"github.com/naurffxiv/moddingway/internal/util"
)

func main() {
	env := util.EnvGetter{
		Ok: true,
	}

	d := &discord.Discord{}

	// configure logging
	os.Mkdir("logs", os.ModePerm) //nolint:errcheck
	logFile, _ := os.Create(filepath.Join("logs", "appLogs.log"))
	defer logFile.Close()
	multi := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(multi)
	log.Println("Logging configuration set")

	discordToken := env.GetEnv("DISCORD_TOKEN")

	debug := env.GetEnv("DEBUG")
	debug = strings.ToLower(debug)

	if debug == "true" {
		guildID := env.GetEnv("GUILD_ID")
		modLoggingChannelID := env.GetEnv("MOD_LOGGING_CHANNEL_ID")

		d.Token = discordToken
		d.GuildID = guildID
		d.ModLoggingChannelID = modLoggingChannelID
	} else {
		d.Init(discordToken)
	}

	dbArgs := database.DbInfo{
		Host:     env.GetEnv("POSTGRES_HOST"),
		Port:     env.GetEnv("POSTGRES_PORT"),
		User:     env.GetEnv("POSTGRES_USER"),
		Password: env.GetEnv("POSTGRES_PASSWORD"),
		DbName:   env.GetEnv("POSTGRES_DB"),
	}

	if !env.Ok {
		log.Panicf("You must supply a %s to start!", env.EnvName)
	}

	d.Conn = database.ConnectToDatabase(dbArgs)
	database.PopulateDatabase(d.Conn)
	
	log.Printf("Starting Discord...\n")
	err := d.Start()
	if err != nil {
		log.Panic(fmt.Errorf("Could not instantiate Discord: %w", err))
	}
	defer d.Session.Close()
	start(d)
}

// start adds all the commands and connects the bot to Discord.
// Listens for CTRL+C then terminates the connection.
func start(d *discord.Discord) {
	d.Ready.Add(1)
	d.Session.AddHandler(d.DiscordReady)
	err := d.Session.Open()
	if err != nil {
		log.Panicf("Could not open Discord session: %f", err)
	}

	d.Ready.Wait()
	d.Session.AddHandler(d.InteractionCreate)
	log.Println("Moddingway is ready. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	d.Session.Close()
}
