package clearingway

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
)

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
