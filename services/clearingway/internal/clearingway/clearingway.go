package clearingway

import (
	"clearingway/internal/clearingway/config"
	"clearingway/internal/discord"
	"clearingway/internal/env"
	"clearingway/internal/fflogs"
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Clearingway struct {
	Config         *config.BotConfig
	Discord        *discord.Discord
	FFLogs         *fflogs.FFLogs
	CommandHandler *CommandHandler
}

// NewBotInstance - Initializes a new Clearingway bot instance
func NewBotInstance(loadedEnv *env.Env) (*Clearingway, error) {
	ctx := context.Background()

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

	// ============== INITIALIZE CLEARINGWAY ==============
	cw := &Clearingway{
		Config:  loadedConfig,
		Discord: discordClient,
		FFLogs:  fflogs.Init(ctx, loadedEnv),
	}

	// ============== INITIALIZE COMMAND HANDLER ==============
	cw.CommandHandler = cw.initCommandHandler(discordClient.Session, loadedEnv)

	return cw, nil
}

// initCommandHandler - Sets up the command handler and registers commands
func (cw *Clearingway) initCommandHandler(session *discordgo.Session, loadedEnv *env.Env) *CommandHandler {
	handler := NewCommandHandler()

	handler.Register(cw.PingCommand())
	handler.Register(cw.GetClearsCommand())

	session.AddHandler(func(session *discordgo.Session, ready *discordgo.Ready) {
		// If env is development, register commands to test guild only
		// If this is a blank string, commands are registered globally
		testingGuildId := ""
		if loadedEnv.ENV == env.Development {
			testingGuildId = loadedEnv.DISCORD_TEST_GUILD_ID
		}

		if err := handler.RegisterAll(session, testingGuildId); err != nil {
			log.Printf("Error registering commands: %v", err)
		}
	})

	session.AddHandler(func(session *discordgo.Session, inter *discordgo.InteractionCreate) {
		err := handler.HandleInteraction(session, inter)
		if err != nil {
			log.Printf("Error handling interaction: %v", err)
		}
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
