package clearingway

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestPingCommand(t *testing.T) {
	cmd := (&Clearingway{}).PingCommand()

	assert.Equal(t, "ping", cmd.ApplicationCommand.Name)
	assert.Equal(t, "Responds with Pong!", cmd.ApplicationCommand.Description)
	assert.NotNil(t, cmd.Handler)

	t.Run("handler responds with Pong!", func(t *testing.T) {
		var capturedResponse *discordgo.InteractionResponse

		mockSession := &MockSession{
			InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
				capturedResponse = response
				return nil
			},
		}

		cmdData := &discordgo.ApplicationCommandInteractionData{
			Name: "ping",
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
		assert.NotNil(t, capturedResponse)
		assert.Equal(t, discordgo.InteractionResponseChannelMessageWithSource, capturedResponse.Type)
		assert.Equal(t, "Pong!", capturedResponse.Data.Content)
	})

	t.Run("handler returns error when session returns error", func(t *testing.T) {
		expectedError := assert.AnError
		mockSession := &MockSession{
			InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
				return expectedError
			},
		}

		cmdData := &discordgo.ApplicationCommandInteractionData{
			Name: "ping",
		}

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				ID:   "test-interaction-id",
				Type: discordgo.InteractionApplicationCommand,
				Data: cmdData,
			},
		}

		err := cmd.Handler(mockSession, interaction)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}
