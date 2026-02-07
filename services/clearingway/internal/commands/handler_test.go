package commands

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.commands)
	assert.Equal(t, 0, len(handler.commands))
}

func TestCommandHandler_Register(t *testing.T) {
	handler := NewHandler()

	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Test command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return nil
		},
	}

	handler.Register(cmd)

	retrievedCmd, ok := handler.GetCommand("test")
	assert.True(t, ok)
	assert.Equal(t, "test", retrievedCmd.ApplicationCommand.Name)
}

func TestCommandHandler_Register_Overwrite(t *testing.T) {
	handler := NewHandler()

	originalCmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Original command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return errors.New("original")
		},
	}

	updatedCmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Updated command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return nil
		},
	}

	handler.Register(originalCmd)
	handler.Register(updatedCmd)

	retrievedCmd, ok := handler.GetCommand("test")
	require.True(t, ok)
	assert.Equal(t, "Updated command", retrievedCmd.ApplicationCommand.Description)

	err := retrievedCmd.Handler(nil, nil)
	assert.NoError(t, err)
}

func TestCommandHandler_Register_NoName(t *testing.T) {
	handler := NewHandler()

	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "",
			Description: "Invalid command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return nil
		},
	}

	assert.Panics(t, func() {
		handler.Register(cmd)
	})
}

func TestCommandHandler_Register_NilHandler(t *testing.T) {
	handler := NewHandler()

	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "invalid",
			Description: "Invalid command",
		},
		Handler: nil,
	}

	assert.Panics(t, func() {
		handler.Register(cmd)
	})
}

func TestCommandHandler_GetCommand(t *testing.T) {
	handler := NewHandler()

	t.Run("returns command when exists", func(t *testing.T) {
		cmd := Command{
			ApplicationCommand: &discordgo.ApplicationCommand{
				Name:        "ping",
				Description: "Ping command",
			},
			Handler: func(s Session, i *discordgo.InteractionCreate) error {
				return nil
			},
		}

		handler.Register(cmd)

		retrievedCmd, ok := handler.GetCommand("ping")
		assert.True(t, ok)
		assert.Equal(t, "ping", retrievedCmd.ApplicationCommand.Name)
	})

	t.Run("returns false when command does not exist", func(t *testing.T) {
		retrievedCmd, ok := handler.GetCommand("unknown")
		assert.False(t, ok)
		assert.Equal(t, Command{}, retrievedCmd)
	})
}

func TestCommandHandler_GetCommand_UnknownCommand(t *testing.T) {
	handler := NewHandler()

	retrievedCmd, ok := handler.GetCommand("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, Command{}, retrievedCmd)
}

func TestCommandHandler_HandleInteraction_UnknownCommand(t *testing.T) {
	handler := NewHandler()

	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			assert.Equal(t, "Command not found", response.Data.Content)
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "unknown-command",
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := handler.HandleInteraction(mockSession, interaction)
	assert.NoError(t, err)
}

func TestCommandHandler_HandleInteraction_KnownCommand_Success(t *testing.T) {
	handler := NewHandler()

	handlerExecuted := false
	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Ping command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			handlerExecuted = true
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
			return err
		},
	}
	handler.Register(cmd)

	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			assert.Equal(t, "Pong!", response.Data.Content)
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "ping",
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := handler.HandleInteraction(mockSession, interaction)
	assert.NoError(t, err)
	assert.True(t, handlerExecuted)
}

func TestCommandHandler_HandleInteraction_KnownCommand_HandlerError(t *testing.T) {
	handler := NewHandler()

	handlerError := errors.New("handler error")
	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "failing",
			Description: "Failing command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return handlerError
		},
	}
	handler.Register(cmd)

	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			assert.Contains(t, response.Data.Content, "Error executing command:")
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "failing",
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := handler.HandleInteraction(mockSession, interaction)
	assert.NoError(t, err)
}

func TestCommandHandler_HandleInteraction_HandlerErrorEditFallback(t *testing.T) {
	handler := NewHandler()

	handlerError := errors.New("handler error")
	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "failing",
			Description: "Failing command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return handlerError
		},
	}
	handler.Register(cmd)

	respondError := errors.New("respond failed")
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			return respondError
		},
		InteractionResponseEditFunc: func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
			assert.Contains(t, *edit.Content, "Error executing command:")
			return nil, nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "failing",
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := handler.HandleInteraction(mockSession, interaction)
	assert.NoError(t, err)
}

func TestCommandHandler_HandleInteraction_HandlerErrorEditFallbackError(t *testing.T) {
	handler := NewHandler()

	handlerError := errors.New("handler error")
	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "failing",
			Description: "Failing command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return handlerError
		},
	}
	handler.Register(cmd)

	respondError := errors.New("respond failed")
	editError := errors.New("edit failed")
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			return respondError
		},
		InteractionResponseEditFunc: func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
			return nil, editError
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "failing",
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := handler.HandleInteraction(mockSession, interaction)
	assert.Error(t, err)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	assert.Contains(t, err.Error(), "failed to respond with error")
	assert.Contains(t, err.Error(), "edit failed")
}

func TestCommandHandler_HandleInteraction_InvalidInteractionType(t *testing.T) {
	handler := NewHandler()

	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			assert.Equal(t, "Internal Error", response.Data.Content)
			return nil
		},
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionPing,
		},
	}

	err := handler.HandleInteraction(mockSession, interaction)
	assert.Error(t, err)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	assert.Contains(t, err.Error(), "invalid interaction type")
}

func TestCommandHandler_HandleInteraction_InvalidInteractionType_RespondError(t *testing.T) {
	handler := NewHandler()

	respondError := errors.New("respond failed")
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			return respondError
		},
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionPing,
		},
	}

	err := handler.HandleInteraction(mockSession, interaction)
	assert.Error(t, err)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	assert.Contains(t, err.Error(), "failed to respond to invalid interaction type")
	assert.Contains(t, err.Error(), "respond failed")
}

func TestCommandHandler_RegisterAll_Success(t *testing.T) {
	handler := NewHandler()

	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Test command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return nil
		},
	}
	handler.Register(cmd)

	user := &discordgo.User{
		ID: "test-app-id",
	}

	mockSession := &MockSession{
		ApplicationCommandBulkOverwriteFunc: func(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
			assert.Equal(t, "test-app-id", appID)
			assert.Equal(t, "test-guild-id", guildID)
			assert.Len(t, commands, 1)
			assert.Equal(t, "test", commands[0].Name)
			return nil, nil
		},
		UserFunc: func(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
			return user, nil
		},
	}

	err := handler.RegisterAll(mockSession, "test-guild-id")
	assert.NoError(t, err)
}

func TestCommandHandler_RegisterAll_Error(t *testing.T) {
	handler := NewHandler()

	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Test command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return nil
		},
	}
	handler.Register(cmd)

	expectedError := errors.New("API error")
	mockSession := &MockSession{
		ApplicationCommandBulkOverwriteFunc: func(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
			return nil, expectedError
		},
		UserFunc: func(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
			return &discordgo.User{ID: "test-app-id"}, nil
		},
	}

	err := handler.RegisterAll(mockSession, "test-guild-id")
	assert.Error(t, err)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	assert.Contains(t, err.Error(), "failed to bulk register commands")
	assert.Contains(t, err.Error(), "API error")
}

func TestCommandHandler_RegisterAll_UserError(t *testing.T) {
	handler := NewHandler()

	cmd := Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Test command",
		},
		Handler: func(s Session, i *discordgo.InteractionCreate) error {
			return nil
		},
	}
	handler.Register(cmd)

	userError := errors.New("user fetch error")
	mockSession := &MockSession{
		UserFunc: func(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
			return nil, userError
		},
	}

	err := handler.RegisterAll(mockSession, "test-guild-id")
	assert.Error(t, err)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	assert.Contains(t, err.Error(), "failed to get current user")
	assert.Contains(t, err.Error(), "user fetch error")
}

func TestCommandHandler_RegisterAll_EmptyCommands(t *testing.T) {
	handler := NewHandler()

	user := &discordgo.User{
		ID: "test-app-id",
	}

	var capturedCommands []*discordgo.ApplicationCommand
	mockSession := &MockSession{
		ApplicationCommandBulkOverwriteFunc: func(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
			capturedCommands = commands
			return nil, nil
		},
		UserFunc: func(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
			return user, nil
		},
	}

	err := handler.RegisterAll(mockSession, "")
	assert.NoError(t, err)
	assert.Len(t, capturedCommands, 0)
}
