package main

import (
	"clearingway/internal/clearingway"
	"clearingway/internal/env"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// ============= LOAD ENV VARIABLES ==============
	loadedEnv, err := env.LoadEnv()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// ============= INITIALIZE/START CLEARINGWAY BOT ==============
	clearingwayBot, err := clearingway.NewBotInstance(loadedEnv)
	if err != nil {
		log.Fatalf("Error initializing Clearingway bot: %v", err)
	}
	if err := clearingwayBot.Start(); err != nil {
		log.Fatalf("Error starting Clearingway bot: %v", err)
	}

	// ============= WAIT FOR TERMINATION SIGNAL ==============
	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// ============= STOP CLEARINGWAY BOT ==============
	if err = clearingwayBot.Stop(); err != nil {
		log.Fatalf("Error stopping Clearingway bot: %v", err)
	}
}
