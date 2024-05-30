package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Token   string
	Session *discordgo.Session
}

func (d *Discord) Start() error {
	s, err := discordgo.New("Bot " + d.Token)
	if err != nil {
		return fmt.Errorf("Could not start Discord: %f", err)
	}

	s.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentGuildMembers | discordgo.IntentGuildModeration | discordgo.IntentGuildMessageReactions

	d.Session = s
	return nil
}
