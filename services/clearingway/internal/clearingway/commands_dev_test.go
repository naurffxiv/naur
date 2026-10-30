package clearingway

import (
	"context"
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockFFLogs struct {
	CheckCharacterClearFunc func(ctx context.Context, lodestoneId, encounterId int64) (bool, error)
}

func (m *MockFFLogs) CheckCharacterClear(ctx context.Context, lodestoneId, encounterId int64) (bool, error) {
	if m.CheckCharacterClearFunc != nil {
		return m.CheckCharacterClearFunc(ctx, lodestoneId, encounterId)
	}
	return false, nil
}

func TestGetClearsCommand_CommandMetadata(t *testing.T) {
	cw := &Clearingway{}
	cmd := cw.GetClearsCommand()

	assert.Equal(t, "getclears", cmd.ApplicationCommand.Name)
	assert.Equal(t, "Retrieves the number of clears for a specified player.", cmd.ApplicationCommand.Description)
	require.Len(t, cmd.ApplicationCommand.Options, 2)

	assert.Equal(t, "lodestone_id", cmd.ApplicationCommand.Options[0].Name)
	assert.Equal(t, "The Lodestone ID of the player to retrieve clears for.", cmd.ApplicationCommand.Options[0].Description)
	assert.True(t, cmd.ApplicationCommand.Options[0].Required)
	assert.Equal(t, discordgo.ApplicationCommandOptionInteger, cmd.ApplicationCommand.Options[0].Type)

	assert.Equal(t, "encounter_id", cmd.ApplicationCommand.Options[1].Name)
	assert.Equal(t, "The encounter ID to check for", cmd.ApplicationCommand.Options[1].Description)
	assert.True(t, cmd.ApplicationCommand.Options[1].Required)
	assert.Equal(t, discordgo.ApplicationCommandOptionInteger, cmd.ApplicationCommand.Options[1].Type)
}

func TestGetClearsCommand_Handler_HasClears(t *testing.T) {
	cw := &Clearingway{
		FFLogs: &MockFFLogs{
			CheckCharacterClearFunc: func(ctx context.Context, lodestoneId, encounterId int64) (bool, error) {
				assert.Equal(t, int64(12345), lodestoneId)
				assert.Equal(t, int64(1000), encounterId)
				return true, nil
			},
		},
	}

	cmd := cw.GetClearsCommand()

	var respondedContent string
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			respondedContent = response.Data.Content
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "getclears",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Name:  "lodestone_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(12345),
			},
			{
				Name:  "encounter_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(1000),
			},
		},
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := cmd.Handler(mockSession, interaction)
	assert.NoError(t, err)
	assert.Equal(t, "Player has at least 1 clear for the specified encounter!", respondedContent)
}

func TestGetClearsCommand_Handler_NoClears(t *testing.T) {
	cw := &Clearingway{
		FFLogs: &MockFFLogs{
			CheckCharacterClearFunc: func(ctx context.Context, lodestoneId, encounterId int64) (bool, error) {
				return false, nil
			},
		},
	}

	cmd := cw.GetClearsCommand()

	var respondedContent string
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			respondedContent = response.Data.Content
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "getclears",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Name:  "lodestone_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(12345),
			},
			{
				Name:  "encounter_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(1000),
			},
		},
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := cmd.Handler(mockSession, interaction)
	assert.NoError(t, err)
	assert.Equal(t, "Player has no clears for the specified encounter.", respondedContent)
}

func TestGetClearsCommand_Handler_InvalidOptions_ZeroLodestoneId(t *testing.T) {
	cw := &Clearingway{
		FFLogs: &MockFFLogs{},
	}

	cmd := cw.GetClearsCommand()

	var respondedContent string
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			respondedContent = response.Data.Content
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "getclears",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Name:  "lodestone_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(0),
			},
			{
				Name:  "encounter_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(1000),
			},
		},
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := cmd.Handler(mockSession, interaction)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid options provided. Please ensure both lodestoneId and encounterId are provided and valid.", respondedContent)
}

func TestGetClearsCommand_Handler_InvalidOptions_ZeroEncounterId(t *testing.T) {
	cw := &Clearingway{
		FFLogs: &MockFFLogs{},
	}

	cmd := cw.GetClearsCommand()

	var respondedContent string
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			respondedContent = response.Data.Content
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "getclears",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Name:  "lodestone_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(12345),
			},
			{
				Name:  "encounter_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(0),
			},
		},
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := cmd.Handler(mockSession, interaction)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid options provided. Please ensure both lodestoneId and encounterId are provided and valid.", respondedContent)
}

func TestGetClearsCommand_Handler_FFLogsError(t *testing.T) {
	fflogsError := errors.New("FFLogs API error")

	cw := &Clearingway{
		FFLogs: &MockFFLogs{
			CheckCharacterClearFunc: func(ctx context.Context, lodestoneId, encounterId int64) (bool, error) {
				return false, fflogsError
			},
		},
	}

	cmd := cw.GetClearsCommand()

	var respondedContent string
	mockSession := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			respondedContent = response.Data.Content
			return nil
		},
	}

	cmdData := discordgo.ApplicationCommandInteractionData{
		Name: "getclears",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Name:  "lodestone_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(12345),
			},
			{
				Name:  "encounter_id",
				Type:  discordgo.ApplicationCommandOptionInteger,
				Value: float64(1000),
			},
		},
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "test-interaction-id",
			Type: discordgo.InteractionApplicationCommand,
			Data: cmdData,
		},
	}

	err := cmd.Handler(mockSession, interaction)
	assert.NoError(t, err)
	assert.Equal(t, "Error retrieving player clears. Please try again later.", respondedContent)
}
