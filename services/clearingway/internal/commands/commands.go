package commands

import (
	"github.com/bwmarrin/discordgo"
)

// PingCommand - Returns a ping command that responds with "Pong!".
func PingCommand() Command {
	return Command{
		Name:        "ping",
		Description: "Responds with pong",
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
