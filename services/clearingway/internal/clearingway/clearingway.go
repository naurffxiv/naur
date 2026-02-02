package clearingway

import (
	"clearingway/internal/clearingway/config"
	"clearingway/internal/commands"
	"clearingway/internal/discord"
	"clearingway/internal/env"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Clearingway struct {
	Config         *config.BotConfig
	Discord        *discord.Discord
	CommandHandler *commands.CommandHandler
}

// NewBotInstance - Initializes a new Clearingway bot instance
func NewBotInstance(loadedEnv *env.Env) (*Clearingway, error) {
	// ============== LOAD CONFIG ==============
	loadedConfig, err := config.InitBotConfig(loadedEnv.CONFIG_PATH)
	if err != nil {
		return nil, err
	}

	// ============== INITIALIZE DISCORD ==============
	discordClient, err := discord.NewSession(loadedEnv.DISCORD_TOKEN)
	if err != nil {
		return nil, err
	}

	return &Clearingway{
		Config:         loadedConfig,
		Discord:        discordClient,
		CommandHandler: initCommandHandler(discordClient.Session, loadedEnv),
	}, nil
}

// initCommandHandler - Sets up the command handler and registers commands
func initCommandHandler(s *discordgo.Session, loadedEnv *env.Env) *commands.CommandHandler {
	handler := commands.NewHandler()

	handler.Register(commands.PingCommand())

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		// If env is development, register commands to test guild only
		// If this is a blank string, commands are registered globally
		testingGuildId := ""
		if loadedEnv.ENV == env.Development {
			testingGuildId = loadedEnv.DISCORD_TEST_GUILD_ID
		}

		if err := handler.RegisterAll(s, testingGuildId); err != nil {
			log.Printf("Error registering commands: %v", err)
		}
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handler.HandleInteraction(s, i)
	})

	return handler
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
