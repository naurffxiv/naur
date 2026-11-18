// Package discord is the adapter between findingway and the Discord API.
package discord

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Veraticus/findingway/internal/tokenizer"

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

func tokensHeader(lookback int, listingCount int64, at time.Time) string {
	return fmt.Sprintf("**Last %d days of Tokens (%d listings scanned) as of %s\n**", lookback, listingCount, at.Format(time.DateTime))
}

// TokensMessage lists the top 75 tokens seen over the lookback days.
func TokensMessage(tokens []tokenizer.Token, lookback int, listingCount int64, at time.Time) *discordgo.MessageSend {
	tokenString := tokensHeader(lookback, listingCount, at)
	tokenSlice := tokens[:min(75, len(tokens))]
	for _, token := range tokenSlice {
		tokenString = tokenString + fmt.Sprintf("`%s` %d\n", token.String, token.Count)
	}

	return &discordgo.MessageSend{
		Content: tokenString,
	}
}

// DescriptionCsvMessage attaches the description CSV.
func DescriptionCsvMessage(buf *bytes.Buffer, at time.Time) *discordgo.MessageSend {
	file := discordgo.File{
		Name:        fmt.Sprintf("pf_descriptions_%s.csv", at.Format(time.DateTime)),
		ContentType: "csv",
		Reader:      bytes.NewReader(buf.Bytes()),
	}

	return &discordgo.MessageSend{
		Content: fmt.Sprintf("**Description List %s**", at.Format(time.DateTime)),
		File:    &file,
	}
}
