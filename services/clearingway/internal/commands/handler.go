package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// CommandHandler - Manages Discord slash commands.
type CommandHandler struct {
	commands map[string]Command
}

// Command - Represents a Discord slash command with its metadata and handler.
type Command struct {
	ApplicationCommand *discordgo.ApplicationCommand
	Handler            func(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// NewHandler - Creates a new CommandHandler with an empty command registry.
func NewHandler() *CommandHandler {
	return &CommandHandler{
		commands: make(map[string]Command),
	}
}

// Register - Adds a command to the handler's registry.
func (cmdHandler *CommandHandler) Register(cmd Command) {
	cmdHandler.commands[cmd.ApplicationCommand.Name] = cmd
}

// HandleInteraction - Processes an incoming Discord interaction and executes the appropriate command.
func (cmdHandler *CommandHandler) HandleInteraction(session *discordgo.Session, inter *discordgo.InteractionCreate) error {
	cmd, ok := cmdHandler.commands[inter.ApplicationCommandData().Name]

	if !ok {
		err := session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Command not found",
			},
		})
		return err // nil if successful, error otherwise
	}

	err := cmd.Handler(session, inter)
	if err != nil {
		// Attempt to respond with an error message
		respondErr := session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error executing command: %v", err),
			},
		})

		// If responding fails, we've likely already responded, so try editing the response instead
		if respondErr != nil {
			content := fmt.Sprintf("Error executing command: %v", err)
			_, editErr := session.InteractionResponseEdit(inter.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})

			// If the edit also fails, return that error
			if editErr != nil {
				return fmt.Errorf("failed to respond with error: %w", editErr)
			}
		}
	}
	return nil
}

// RegisterAll - Sends all registered commands to the Discord API for the specified guild using bulk overwrite.
// Pass an empty string for guildID to register commands globally (takes up to 1 hour to propagate).
func (cmdHandler *CommandHandler) RegisterAll(session *discordgo.Session, guildID string) error {
	commands := make([]*discordgo.ApplicationCommand, 0, len(cmdHandler.commands))
	for _, cmd := range cmdHandler.commands {
		commands = append(commands, cmd.ApplicationCommand)
	}

	_, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, guildID, commands)
	if err != nil {
		return fmt.Errorf("failed to bulk register commands: %w", err)
	}
	return nil
}

// GetCommand - Retrieves a command by name from the registry.
func (cmdHandler *CommandHandler) GetCommand(name string) (Command, bool) {
	cmd, ok := cmdHandler.commands[name]
	return cmd, ok
}
