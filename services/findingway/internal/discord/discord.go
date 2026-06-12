package discord

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/Veraticus/findingway/internal/tokenizer"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Token string

	Session  *discordgo.Session
	Channels []*Channel `yaml:"channels"`
}

type Channel struct {
	Name        string   `yaml:"name"`
	ID          string   `yaml:"id"`
	Duty        string   `yaml:"duty"`
	DataCentres []string `yaml:"dataCentres"`
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

func (d *Discord) CleanChannel(channelId string) error {
	messages, err := d.Session.ChannelMessages(channelId, 100, "", "", "")
	if err != nil {
		return fmt.Errorf("could not list messages: %w", err)
	}
	messageIds := []string{}
	botId := d.Session.State.User.ID
	for _, message := range messages {
		if message.Author.ID == botId {
			messageIds = append(messageIds, message.ID)
		}
	}
	if len(messageIds) == 0 {
		return nil
	}
	err = d.Session.ChannelMessagesBulkDelete(channelId, messageIds)
	if err != nil {
		return fmt.Errorf("could not bulk delete messages: %w", err)
	}

	return nil
}

func (d *Discord) PostListings(channelId string, listings *ffxiv.Listings, duty string, dataCentre string) error {
	scopedListings := listings.ForDutyAndDataCentre(duty, dataCentre)

	mostRecent, err := scopedListings.MostRecentUpdated()
	if err != nil {
		return fmt.Errorf("could not find most recently updated duty: %w", err)
	}
	if mostRecent != nil {
		mostRecentUpdated, err := mostRecent.UpdatedAt()
		if err != nil {
			return fmt.Errorf("could not find most recently updatedAt: %w", err)
		}
		if mostRecentUpdated.After(time.Now().Add(-4 * time.Minute)) {
			scopedListings, err = scopedListings.UpdatedWithinLast(4 * time.Minute)
			if err != nil {
				return fmt.Errorf("could not find most recently listings: %w", err)
			}
		}
	}

	headerEmbed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s PFs (%v)", duty, dataCentre),
		Type:  discordgo.EmbedTypeRich,
		Color: 0x6600ff,
		Description: fmt.Sprintf(
			"Found %v listings %v",
			len(scopedListings.Listings),
			fmt.Sprintf("<t:%v:R>", time.Now().Unix()),
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text: strings.Repeat("\u3000", 20),
		},
	}
	headerMessageSend := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{headerEmbed},
	}
	_, err = d.Session.ChannelMessageSendComplex(channelId, headerMessageSend)
	if err != nil {
		return fmt.Errorf("could not send header: %w", err)
	}

	fields := []*discordgo.MessageEmbedField{}
	for i, listing := range scopedListings.Listings {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   listing.Creator,
			Value:  listing.PartyDisplay(),
			Inline: true,
		})
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   listing.GetTags(),
			Value:  listing.GetDescription(),
			Inline: true,
		})
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   listing.GetExpires(),
			Value:  listing.GetUpdated(),
			Inline: true,
		})

		// Send a message every 5 listings
		if (i+1)%5 == 0 {
			err = d.sendMessage(channelId, fields)
			if err != nil {
				return fmt.Errorf("could not send message: %w", err)
			}
			fields = []*discordgo.MessageEmbedField{}
		}
	}

	// Ensure we send any remaining messages
	if len(fields) != 0 {
		err = d.sendMessage(channelId, fields)
		if err != nil {
			return fmt.Errorf("could not send message: %w", err)
		}
	}

	return nil
}

func (d *Discord) sendMessage(channelId string, fields []*discordgo.MessageEmbedField) error {
	embed := &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Color:  0x6600ff,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: strings.Repeat("\u3000", 20),
		},
	}
	messageSend := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	}
	_, err := d.Session.ChannelMessageSendComplex(channelId, messageSend)
	if err != nil {
		return err
	}

	return nil
}

func (d *Discord) PostTokens(channelId string, tokens []tokenizer.Token, listingCount int64) error {

	tokenString := fmt.Sprintf("**Last 2 days of Tokens (%d listings scanned) as of %s\n**", listingCount, time.Now().Format(time.DateTime))
	tokenSlice := tokens[:min(75, len(tokens))]
	for _, token := range tokenSlice {
		tokenString = tokenString + fmt.Sprintf("`%s` %d\n", token.String, token.Count)
	}

	messageSend := &discordgo.MessageSend{
		Content: tokenString,
	}
	_, err := d.Session.ChannelMessageSendComplex(channelId, messageSend)
	if err != nil {
		return err
	}

	return nil
}

func (d *Discord) PostDescriptionCsv(channelId string, buf *bytes.Buffer) error {

	reader := bytes.NewReader(buf.Bytes())
	file := discordgo.File{
		Name:        fmt.Sprintf("pf_descriptions_%s.csv", time.Now().Format(time.DateTime)),
		ContentType: "csv",
		Reader:      reader,
	}

	messageSend := &discordgo.MessageSend{
		Content: fmt.Sprintf("**Description List %s**", time.Now().Format(time.DateTime)),
		File:    &file,
	}

	_, err := d.Session.ChannelMessageSendComplex(channelId, messageSend)
	if err != nil {
		return err
	}

	return nil
}
