package clearingway

import (
	"clearingway/internal/lodestone"
	"context"
	"errors"
	"fmt"
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

// ============== LodestoneCommand ==============

type MockLodestone struct {
	SearchCharacterLodestoneIDFunc    func(ctx context.Context, name, world string) (int64, error)
	IsCharacterOwnedByDiscordUserFunc func(ctx context.Context, lodestoneId int64, discordId string) (bool, error)
}

func (m *MockLodestone) SearchCharacterLodestoneID(ctx context.Context, name, world string) (int64, error) {
	if m.SearchCharacterLodestoneIDFunc != nil {
		return m.SearchCharacterLodestoneIDFunc(ctx, name, world)
	}
	return 0, nil
}

func (m *MockLodestone) IsCharacterOwnedByDiscordUser(ctx context.Context, lodestoneId int64, discordId string) (bool, error) {
	if m.IsCharacterOwnedByDiscordUserFunc != nil {
		return m.IsCharacterOwnedByDiscordUserFunc(ctx, lodestoneId, discordId)
	}
	return false, nil
}

func (m *MockLodestone) CharacterURL(lodestoneId int64) string {
	return fmt.Sprintf("https://lodestone.test/character/%d/", lodestoneId)
}

func (m *MockLodestone) ProfileSettingsURL() string {
	return "https://lodestone.test/my/setting/profile/"
}

// lodestoneInteraction - Builds a guild /lodestone interaction from user 123456789
func lodestoneInteraction(firstName, lastName, world string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:     "test-interaction-id",
			Type:   discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{User: &discordgo.User{ID: "123456789"}},
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "lodestone",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Name: "first_name", Type: discordgo.ApplicationCommandOptionString, Value: firstName},
					{Name: "last_name", Type: discordgo.ApplicationCommandOptionString, Value: lastName},
					{Name: "world", Type: discordgo.ApplicationCommandOptionString, Value: world},
				},
			},
		},
	}
}

// lodestoneTestSession - Returns a mock session recording response types and the final edited content
type lodestoneTestSession struct {
	MockSession
	ResponseTypes []discordgo.InteractionResponseType
	Content       string
	Edited        bool
}

func newLodestoneTestSession() *lodestoneTestSession {
	s := &lodestoneTestSession{}
	s.InteractionRespondFunc = func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
		s.ResponseTypes = append(s.ResponseTypes, response.Type)
		if response.Data != nil {
			s.Content = response.Data.Content
		}
		return nil
	}
	s.InteractionResponseEditFunc = func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
		s.Edited = true
		s.Content = *edit.Content
		return nil, nil
	}
	return s
}

func TestLodestoneCommand_CommandMetadata(t *testing.T) {
	cw := &Clearingway{}
	cmd := cw.LodestoneCommand()

	assert.Equal(t, "lodestone", cmd.ApplicationCommand.Name)
	require.Len(t, cmd.ApplicationCommand.Options, 3)

	for i, name := range []string{"first_name", "last_name", "world"} {
		assert.Equal(t, name, cmd.ApplicationCommand.Options[i].Name)
		assert.True(t, cmd.ApplicationCommand.Options[i].Required)
		assert.Equal(t, discordgo.ApplicationCommandOptionString, cmd.ApplicationCommand.Options[i].Type)
	}
}

func TestLodestoneCommand_Handler_FoundAndOwned(t *testing.T) {
	cw := &Clearingway{
		Lodestone: &MockLodestone{
			SearchCharacterLodestoneIDFunc: func(ctx context.Context, name, world string) (int64, error) {
				// Inputs are normalised before searching
				assert.Equal(t, "Alicja Tester", name)
				assert.Equal(t, "Gilgamesh", world)
				return 12345, nil
			},
			IsCharacterOwnedByDiscordUserFunc: func(ctx context.Context, lodestoneId int64, discordId string) (bool, error) {
				assert.Equal(t, int64(12345), lodestoneId)
				assert.Equal(t, "123456789", discordId)
				return true, nil
			},
		},
	}
	session := newLodestoneTestSession()

	err := cw.LodestoneCommand().Handler(session, lodestoneInteraction("alicja", "TESTER", "gilgamesh"))

	assert.NoError(t, err)
	assert.Equal(t, []discordgo.InteractionResponseType{discordgo.InteractionResponseDeferredChannelMessageWithSource}, session.ResponseTypes)
	assert.True(t, session.Edited)
	assert.Contains(t, session.Content, "Found `Alicja Tester (Gilgamesh)`")
	assert.Contains(t, session.Content, "`12345`")
	assert.Contains(t, session.Content, "https://lodestone.test/character/12345/")
	assert.Contains(t, session.Content, "Ownership verified for <@123456789>")
}

func TestLodestoneCommand_Handler_FoundNotOwned(t *testing.T) {
	cw := &Clearingway{
		Lodestone: &MockLodestone{
			SearchCharacterLodestoneIDFunc: func(ctx context.Context, name, world string) (int64, error) {
				return 12345, nil
			},
			IsCharacterOwnedByDiscordUserFunc: func(ctx context.Context, lodestoneId int64, discordId string) (bool, error) {
				return false, nil
			},
		},
	}
	session := newLodestoneTestSession()

	err := cw.LodestoneCommand().Handler(session, lodestoneInteraction("Alicja", "Tester", "Gilgamesh"))

	assert.NoError(t, err)
	assert.Contains(t, session.Content, "Could not verify ownership")
	assert.Contains(t, session.Content, lodestone.OwnershipCode("123456789"))
	assert.Contains(t, session.Content, "https://lodestone.test/my/setting/profile/")
}

func TestLodestoneCommand_Handler_OwnershipError(t *testing.T) {
	cw := &Clearingway{
		Lodestone: &MockLodestone{
			SearchCharacterLodestoneIDFunc: func(ctx context.Context, name, world string) (int64, error) {
				return 12345, nil
			},
			IsCharacterOwnedByDiscordUserFunc: func(ctx context.Context, lodestoneId int64, discordId string) (bool, error) {
				return false, errors.New("boom")
			},
		},
	}
	session := newLodestoneTestSession()

	err := cw.LodestoneCommand().Handler(session, lodestoneInteraction("Alicja", "Tester", "Gilgamesh"))

	assert.NoError(t, err)
	assert.Contains(t, session.Content, "Found `Alicja Tester (Gilgamesh)`")
	assert.Contains(t, session.Content, "Error verifying ownership")
}

func TestLodestoneCommand_Handler_SearchErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"not found", fmt.Errorf("%w: x", lodestone.ErrCharacterNotFound), "No character found on the Lodestone for `Alicja Tester (Gilgamesh)`"},
		{"multiple", fmt.Errorf("%w: x", lodestone.ErrMultipleCharactersFound), "Multiple characters named `Alicja Tester` were found on Gilgamesh"},
		{"other", errors.New("boom"), "Error searching the Lodestone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownershipChecked := false
			cw := &Clearingway{
				Lodestone: &MockLodestone{
					SearchCharacterLodestoneIDFunc: func(ctx context.Context, name, world string) (int64, error) {
						return 0, tt.err
					},
					IsCharacterOwnedByDiscordUserFunc: func(ctx context.Context, lodestoneId int64, discordId string) (bool, error) {
						ownershipChecked = true
						return false, nil
					},
				},
			}
			session := newLodestoneTestSession()

			err := cw.LodestoneCommand().Handler(session, lodestoneInteraction("Alicja", "Tester", "Gilgamesh"))

			assert.NoError(t, err)
			assert.True(t, session.Edited)
			assert.Contains(t, session.Content, tt.expected)
			assert.False(t, ownershipChecked)
		})
	}
}

func TestLodestoneCommand_Handler_InvalidWorld(t *testing.T) {
	cw := &Clearingway{
		Lodestone: &MockLodestone{
			SearchCharacterLodestoneIDFunc: func(ctx context.Context, name, world string) (int64, error) {
				t.Fatal("Lodestone should not be searched for an invalid world")
				return 0, nil
			},
		},
	}
	session := newLodestoneTestSession()

	err := cw.LodestoneCommand().Handler(session, lodestoneInteraction("Alicja", "Tester", "Narnia"))

	assert.NoError(t, err)
	assert.Equal(t, []discordgo.InteractionResponseType{discordgo.InteractionResponseChannelMessageWithSource}, session.ResponseTypes)
	assert.False(t, session.Edited)
	assert.Equal(t, "`Narnia` is not a valid world! Make sure you spelled your world name properly.", session.Content)
}

func TestLodestoneCommand_Handler_InvalidName(t *testing.T) {
	cw := &Clearingway{Lodestone: &MockLodestone{}}
	session := newLodestoneTestSession()

	err := cw.LodestoneCommand().Handler(session, lodestoneInteraction("A", "Tester", "Gilgamesh"))

	assert.NoError(t, err)
	assert.False(t, session.Edited)
	assert.Equal(t, "Invalid character name! First and last names must each be at least 2 characters.", session.Content)
}

func TestLodestoneCommand_Handler_DeferError(t *testing.T) {
	expectedErr := errors.New("discord unavailable")
	cw := &Clearingway{Lodestone: &MockLodestone{}}
	session := &MockSession{
		InteractionRespondFunc: func(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
			return expectedErr
		},
	}

	err := cw.LodestoneCommand().Handler(session, lodestoneInteraction("Alicja", "Tester", "Gilgamesh"))

	assert.ErrorIs(t, err, expectedErr)
}

func TestInteractionUserID(t *testing.T) {
	guild := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{Member: &discordgo.Member{User: &discordgo.User{ID: "guild-user"}}}}
	dm := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{User: &discordgo.User{ID: "dm-user"}}}
	none := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{}}

	assert.Equal(t, "guild-user", interactionUserID(guild))
	assert.Equal(t, "dm-user", interactionUserID(dm))
	assert.Empty(t, interactionUserID(none))
}
