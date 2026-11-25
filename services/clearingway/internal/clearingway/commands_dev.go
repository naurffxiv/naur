package clearingway

import (
	"clearingway/internal/ffxiv"
	"clearingway/internal/lodestone"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const lodestoneCommandTimeout = 60 * time.Second

type FFLogsInterface interface {
	CheckCharacterClear(ctx context.Context, lodestoneId, encounterId int64) (bool, error)
}

type LodestoneInterface interface {
	SearchCharacterLodestoneID(ctx context.Context, name, world string) (int64, error)
	IsCharacterOwnedByDiscordUser(ctx context.Context, lodestoneId int64, discordId string) (bool, error)
	CharacterURL(lodestoneId int64) string
	ProfileSettingsURL() string
}

// GetClearsCommand - Returns a command that retrieves if a lodestone ID has cleared an encounter.
func (cw *Clearingway) GetClearsCommand() Command {
	return Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "getclears",
			Description: "Retrieves the number of clears for a specified player.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "lodestone_id",
					Description: "The Lodestone ID of the player to retrieve clears for.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "encounter_id",
					Description: "The encounter ID to check for",
					Required:    true,
				},
			},
		},
		Handler: func(session Session, inter *discordgo.InteractionCreate) error {
			// Extract options
			var lodestoneId int64
			var encounterId int64

			for _, option := range inter.ApplicationCommandData().Options {
				switch option.Name {
				case "lodestone_id":
					lodestoneId = option.IntValue()
				case "encounter_id":
					encounterId = option.IntValue()
				}
			}

			// Validate options
			if lodestoneId == 0 || encounterId == 0 {
				return session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid options provided. Please ensure both lodestoneId and encounterId are provided and valid.",
					},
				})
			}

			// Retrieve clears from FFLogs
			ctx := context.TODO()
			clears, err := cw.FFLogs.CheckCharacterClear(ctx, lodestoneId, encounterId)
			if err != nil {
				log.Printf("Error checking character clears: %v", err)
				return session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error retrieving player clears. Please try again later.",
					},
				})
			}

			// Respond with the number of clears
			responseContent := "Player has no clears for the specified encounter."
			if clears {
				responseContent = "Player has at least 1 clear for the specified encounter!"
			}

			return session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: responseContent,
				},
			})
		},
	}
}

// LodestoneCommand - Returns a command that searches the Lodestone for a character and checks if the invoking user owns it.
func (cw *Clearingway) LodestoneCommand() Command {
	return Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "lodestone",
			Description: "Searches the Lodestone for a character and checks if you own it.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "first_name",
					Description: "The character's first name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "last_name",
					Description: "The character's last name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "world",
					Description: "The character's world",
					Required:    true,
				},
			},
		},
		Handler: func(session Session, inter *discordgo.InteractionCreate) error {
			// Extract options
			var firstName, lastName, world string
			for _, option := range inter.ApplicationCommandData().Options {
				switch option.Name {
				case "first_name":
					firstName = option.StringValue()
				case "last_name":
					lastName = option.StringValue()
				case "world":
					world = option.StringValue()
				}
			}

			// Validate options
			char, err := ffxiv.NewCharacter(firstName, lastName, world)
			if err != nil {
				return session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: invalidCharacterMessage(err, world),
					},
				})
			}

			// Defer the response since Lodestone requests can exceed Discord's 3 second response window
			if err := session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			}); err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), lodestoneCommandTimeout)
			defer cancel()

			content := cw.lodestoneLookup(ctx, char, interactionUserID(inter))
			_, err = session.InteractionResponseEdit(inter.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return err
		},
	}
}

// lodestoneLookup - Searches for the character and checks ownership and returns a message describing the result
func (cw *Clearingway) lodestoneLookup(ctx context.Context, char *ffxiv.Character, discordId string) string {
	lodestoneId, err := cw.Lodestone.SearchCharacterLodestoneID(ctx, char.Name(), char.World)
	switch {
	case errors.Is(err, lodestone.ErrCharacterNotFound):
		return fmt.Sprintf("No character found on the Lodestone for `%s (%s)`. If you recently renamed or transferred, it can take up to a day for the Lodestone to update.", char.Name(), char.World)
	case errors.Is(err, lodestone.ErrMultipleCharactersFound):
		return fmt.Sprintf("Multiple characters named `%s` were found on %s. Please make sure the name is exact.", char.Name(), char.World)
	case err != nil:
		log.Printf("Error searching Lodestone for %s (%s): %v", char.Name(), char.World, err)
		return "Error searching the Lodestone. Please try again later."
	}
	char.LodestoneID = lodestoneId

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found `%s (%s)` on the Lodestone!\n", char.Name(), char.World)
	fmt.Fprintf(&sb, "Lodestone ID: `%d`\n%s\n\n", char.LodestoneID, cw.Lodestone.CharacterURL(char.LodestoneID))

	owned, err := cw.Lodestone.IsCharacterOwnedByDiscordUser(ctx, char.LodestoneID, discordId)
	switch {
	case err != nil:
		log.Printf("Error checking Lodestone ownership of %d: %v", char.LodestoneID, err)
		sb.WriteString("Error verifying ownership. Please try again later.")
	case owned:
		fmt.Fprintf(&sb, "Ownership verified for <@%s>.", discordId)
	default:
		fmt.Fprintf(&sb, "Could not verify ownership. Add the following code to your Lodestone bio and try again:\n**%s**\nYou can edit your bio at %s",
			lodestone.OwnershipCode(discordId), cw.Lodestone.ProfileSettingsURL())
	}

	return sb.String()
}

// invalidCharacterMessage - Returns a user-facing message for character validation errors
func invalidCharacterMessage(err error, world string) string {
	if errors.Is(err, ffxiv.ErrInvalidWorld) {
		return fmt.Sprintf("`%s` is not a valid world! Make sure you spelled your world name properly.", world)
	}
	return fmt.Sprintf("Invalid character name! First and last names must each be at least %d characters.", ffxiv.MinNameLength)
}

// interactionUserID - Returns the ID of the user who triggered the interaction
func interactionUserID(inter *discordgo.InteractionCreate) string {
	if inter.Member != nil && inter.Member.User != nil {
		return inter.Member.User.ID
	}
	if inter.User != nil {
		return inter.User.ID
	}
	return ""
}
