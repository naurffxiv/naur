// Package discord is the adapter between findingway and the Discord API.
package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Token string

	Session *discordgo.Session
}

func (d *Discord) Start() error {
	s, err := discordgo.New("Bot " + d.Token)
	if err != nil {
		return fmt.Errorf("could not start Discord: %w", err)
	}
	s.ShouldRetryOnRateLimit = false

	err = s.Open()
	if err != nil {
		return fmt.Errorf("could not open Discord session: %w", err)
	}

	d.Session = s
	return nil
}

// Clean bulk-deletes the bot's own messages among the channel's latest 100.
func (d *Discord) Clean(channelID string) error {
	messages, err := d.Session.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		return fmt.Errorf("could not list messages: %w", err)
	}
	messageIDs := []string{}
	botID := d.Session.State.User.ID
	for _, message := range messages {
		if message.Author.ID == botID {
			messageIDs = append(messageIDs, message.ID)
		}
	}
	if len(messageIDs) == 0 {
		return nil
	}
	err = d.Session.ChannelMessagesBulkDelete(channelID, messageIDs)
	if err != nil {
		return fmt.Errorf("could not bulk delete messages: %w", err)
	}

	return nil
}

// Send posts one message to the channel.
func (d *Discord) Send(channelID string, message *discordgo.MessageSend) error {
	_, err := d.Session.ChannelMessageSendComplex(channelID, message)
	return err
}
