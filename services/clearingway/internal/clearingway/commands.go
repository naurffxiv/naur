package clearingway

import (
	"github.com/bwmarrin/discordgo"
)

// PingCommand - Returns a ping command that responds with "Pong!".
func (cw *Clearingway) PingCommand() Command {
	return Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Responds with Pong!",
		},
		Handler: func(session Session, inter *discordgo.InteractionCreate) error {
			err := session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
			return err // nil if successful, error otherwise
		},
	}
}
