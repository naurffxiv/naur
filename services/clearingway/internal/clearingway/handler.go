package clearingway

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Session - Interface representing the methods of discordgo.Session used by CommandHandler.
// This allows for easier testing and mocking.
type Session interface {
	InteractionRespond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	InteractionResponseEdit(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ApplicationCommandBulkOverwrite(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error)
	User(userID string, options ...discordgo.RequestOption) (*discordgo.User, error)
}

// CommandHandler - Manages Discord slash commands.
type CommandHandler struct {
	commands map[string]Command
}

// Command - Represents a Discord slash command with its metadata and handler.
type Command struct {
	ApplicationCommand *discordgo.ApplicationCommand
	Handler            func(s Session, i *discordgo.InteractionCreate) error
}

// NewCommandHandler - Creates a new CommandHandler with an empty command registry.
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		commands: make(map[string]Command),
	}
}

// Register - Adds a command to the handler's registry.
func (cmdHandler *CommandHandler) Register(cmd Command) {
	// Command validation - panics are used to halt startup on invalid commands
	if cmd.ApplicationCommand.Name == "" {
		panic("Cannot register command: ApplicationCommand.Name is empty")
	}
	if cmd.Handler == nil {
		panic("Cannot register command: Handler is nil")
	}

	cmdHandler.commands[cmd.ApplicationCommand.Name] = cmd
}

// HandleInteraction - Processes an incoming Discord interaction and executes the appropriate command.
func (cmdHandler *CommandHandler) HandleInteraction(session Session, inter *discordgo.InteractionCreate) error {
	// Ensure the interaction is a command
	if inter.Type != discordgo.InteractionApplicationCommand {
		err := session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Internal Error",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to respond to invalid interaction type: %w", err)
		}
		return errors.New("invalid interaction type")
	}

	// Extract command data
	cmdData := inter.ApplicationCommandData()

	// Lookup command handler
	cmd, ok := cmdHandler.commands[cmdData.Name]
	if !ok {
		err := session.InteractionRespond(inter.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Command not found",
			},
		})
		return err // nil if successful, error otherwise
	}

	// Execute command handler
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
func (cmdHandler *CommandHandler) RegisterAll(session Session, guildID string) error {
	commands := make([]*discordgo.ApplicationCommand, 0, len(cmdHandler.commands))
	for _, cmd := range cmdHandler.commands {
		commands = append(commands, cmd.ApplicationCommand)
	}

	// Get current user to obtain application ID
	user, err := session.User("@me")
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	_, err = session.ApplicationCommandBulkOverwrite(user.ID, guildID, commands)
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
