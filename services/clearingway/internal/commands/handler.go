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
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
	Handler     func(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// NewHandler - Creates a new CommandHandler with an empty command registry.
func NewHandler() *CommandHandler {
	return &CommandHandler{
		commands: make(map[string]Command),
	}
}

// Register - Adds a command to the handler's registry.
func (h *CommandHandler) Register(cmd Command) {
	h.commands[cmd.Name] = cmd
}

// HandleInteraction - Processes an incoming Discord interaction and executes the appropriate command.
func (h *CommandHandler) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd, ok := h.commands[i.ApplicationCommandData().Name]

	if !ok {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Command not found",
			},
		})
		return
	}

	err := cmd.Handler(s, i)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error executing command: %v", err),
			},
		})
		return
	}
}

// RegisterAll - Sends all registered commands to the Discord API for the specified guild.
// Pass an empty string for guildID to register commands globally (takes up to 1 hour to propagate).
func (h *CommandHandler) RegisterAll(s *discordgo.Session, guildID string) error {
	for _, cmd := range h.commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options:     cmd.Options,
		})
		if err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
	}
	return nil
}

// GetCommand - Retrieves a command by name from the registry.
func (h *CommandHandler) GetCommand(name string) (Command, bool) {
	cmd, ok := h.commands[name]
	return cmd, ok
}
