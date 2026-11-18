// Package board keeps a Discord channel showing the current Listings for one
// duty, one block per DataCentre.
package board

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"

	"github.com/bwmarrin/discordgo"
)

const (
	// listingsPerEmbed is how many Listings share one Discord embed.
	listingsPerEmbed = 5
	// freshWindow: when any Listing was updated within this window, only
	// Listings updated within it are shown.
	freshWindow = 4 * time.Minute
	embedColor  = 0x6600ff
)

// Discord is the port the Board posts through.
type Discord interface {
	// Clean removes the bot's own messages from the channel.
	Clean(channelID string) error
	// Send posts one message to the channel.
	Send(channelID string, message *discordgo.MessageSend) error
}

// Channel is a Discord channel that shows the Listings for one duty across
// DataCentres.
type Channel struct {
	Name        string   `yaml:"name"`
	ID          string   `yaml:"id"`
	Duty        string   `yaml:"duty"`
	DataCentres []string `yaml:"dataCentres"`
}

// Board refreshes channels through a Discord port.
type Board struct {
	discord Discord
}

// New returns a Board that posts through discord.
func New(discord Discord) *Board {
	return &Board{discord: discord}
}

// Refresh wipes the channel and reposts the channel's Listings, one header
// plus batched embeds per DataCentre. now anchors the Listings' relative
// times. A failure in one DataCentre does not stop the others; every error
// is returned joined.
func (b *Board) Refresh(channel *Channel, listings *ffxiv.Listings, now time.Time) error {
	var errs []error

	if err := b.discord.Clean(channel.ID); err != nil {
		errs = append(errs, fmt.Errorf("could not clean channel %s: %w", channel.ID, err))
	}

	for _, dataCentre := range channel.DataCentres {
		if err := b.post(channel, listings, dataCentre, now); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", dataCentre, err))
		}
	}

	return errors.Join(errs...)
}

func (b *Board) post(channel *Channel, listings *ffxiv.Listings, dataCentre string, now time.Time) error {
	scoped, err := visibleListings(listings.ForDutyAndDataCentre(channel.Duty, dataCentre), now)
	if err != nil {
		return err
	}

	if err := b.discord.Send(channel.ID, headerMessage(channel.Duty, dataCentre, len(scoped.Listings), now)); err != nil {
		return fmt.Errorf("could not send header: %w", err)
	}

	for start := 0; start < len(scoped.Listings); start += listingsPerEmbed {
		end := min(start+listingsPerEmbed, len(scoped.Listings))
		if err := b.discord.Send(channel.ID, listingsMessage(scoped.Listings[start:end])); err != nil {
			return fmt.Errorf("could not send message: %w", err)
		}
	}

	return nil
}

// visibleListings narrows listings to the fresh window when any of them are
// fresh, and otherwise leaves them all.
func visibleListings(listings *ffxiv.Listings, now time.Time) (*ffxiv.Listings, error) {
	fresh, err := listings.UpdatedWithinLast(freshWindow, now)
	if err != nil {
		return nil, fmt.Errorf("could not find fresh listings: %w", err)
	}
	if len(fresh.Listings) == 0 {
		return listings, nil
	}
	return fresh, nil
}

func footer() *discordgo.MessageEmbedFooter {
	return &discordgo.MessageEmbedFooter{
		Text: strings.Repeat("　", 20),
	}
}

func headerMessage(duty string, dataCentre string, count int, now time.Time) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       fmt.Sprintf("%s PFs (%v)", duty, dataCentre),
			Type:        discordgo.EmbedTypeRich,
			Color:       embedColor,
			Description: fmt.Sprintf("Found %v listings <t:%v:R>", count, now.Unix()),
			Footer:      footer(),
		}},
	}
}

func listingsMessage(listings []*ffxiv.Listing) *discordgo.MessageSend {
	fields := make([]*discordgo.MessageEmbedField, 0, 3*len(listings))
	for _, listing := range listings {
		fields = append(fields,
			&discordgo.MessageEmbedField{
				Name:   listing.Creator,
				Value:  listing.PartyDisplay(),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   listing.GetTags(),
				Value:  listing.GetDescription(),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   listing.GetExpires(),
				Value:  listing.GetUpdated(),
				Inline: true,
			},
		)
	}

	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type:   discordgo.EmbedTypeRich,
			Color:  embedColor,
			Fields: fields,
			Footer: footer(),
		}},
	}
}
