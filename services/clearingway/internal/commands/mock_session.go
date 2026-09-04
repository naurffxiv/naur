package commands

import (
	"github.com/bwmarrin/discordgo"
)

type MockSession struct {
	InteractionRespondFunc              func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	InteractionResponseEditFunc         func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ApplicationCommandBulkOverwriteFunc func(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error)
	UserFunc                            func(userID string, options ...discordgo.RequestOption) (*discordgo.User, error)
}

func (m *MockSession) InteractionRespond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	if m.InteractionRespondFunc != nil {
		return m.InteractionRespondFunc(interaction, response, options...)
	}
	return nil
}

func (m *MockSession) InteractionResponseEdit(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if m.InteractionResponseEditFunc != nil {
		return m.InteractionResponseEditFunc(interaction, edit, options...)
	}
	return nil, nil
}

func (m *MockSession) ApplicationCommandBulkOverwrite(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	if m.ApplicationCommandBulkOverwriteFunc != nil {
		return m.ApplicationCommandBulkOverwriteFunc(appID, guildID, commands, options...)
	}
	return nil, nil
}

func (m *MockSession) User(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
	if m.UserFunc != nil {
		return m.UserFunc(userID, options...)
	}
	return nil, nil
}

type SessionInterface interface {
	InteractionRespond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	InteractionResponseEdit(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ApplicationCommandBulkOverwrite(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error)
	User(userID string, options ...discordgo.RequestOption) (*discordgo.User, error)
}

var _ SessionInterface = (*MockSession)(nil)
