package commands

import (
	"github.com/bwmarrin/discordgo"
)

// PingCommand - Returns a ping command that responds with "Pong!".
func PingCommand() Command {
	return Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Responds with Pong!",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
			return err // nil if successful, error otherwise
		},
	}
}
