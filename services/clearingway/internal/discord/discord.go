package discord

import "github.com/bwmarrin/discordgo"

type Discord struct {
	Session *discordgo.Session
}

// NewSession - Initializes a new Discord session with the provided bot token
func NewSession(token string) (*Discord, error) {
	// "Bot " prefix is required for bot authentication
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	return &Discord{Session: session}, nil
}
