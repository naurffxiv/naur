package clearingway

import (
	"clearingway/internal/clearingway/config"
	"clearingway/internal/discord"
	"clearingway/internal/env"
)

type Clearingway struct {
	Config  *config.BotConfig
	Discord *discord.Discord
}

// NewBotInstance - Initializes a new Clearingway bot instance
func NewBotInstance(env *env.Env) (*Clearingway, error) {
	// ============== LOAD CONFIG ==============
	loadedConfig, err := config.InitBotConfig(env.CONFIG_PATH)
	if err != nil {
		return nil, err
	}

	// ============== INITIALIZE DISCORD ==============
	discordClient, err := discord.NewSession(env.DISCORD_TOKEN)
	if err != nil {
		return nil, err
	}

	return &Clearingway{
		Config:  loadedConfig,
		Discord: discordClient,
	}, nil
}

// Start - Opens the Discord session
func (cw *Clearingway) Start() error {
	return cw.Discord.Session.Open()
}

// Stop - Closes the Discord session
func (cw *Clearingway) Stop() error {
	return cw.Discord.Session.Close()
}

// GetConfig - Returns the bot configuration
func (cw *Clearingway) GetConfig() *config.BotConfig {
	return cw.Config
}

// GetDiscord - Returns the Discord client
func (cw *Clearingway) GetDiscord() *discord.Discord {
	return cw.Discord
}
