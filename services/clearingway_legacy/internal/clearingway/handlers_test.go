package clearingway

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func flexGuild() *ConfigGuild {
	return &ConfigGuild{Name: "Test", ConfigRoles: &ConfigRoles{RelevantFlexing: true, RelevantParsing: true, Legend: true}}
}

func TestRemoveRoleCommands(t *testing.T) {
	run := map[string]func(*Clearingway, *discordgo.Session, *discordgo.InteractionCreate){
		"uncomfy":   (*Clearingway).Uncomfy,
		"uncolor":   (*Clearingway).Uncolor,
		"unflex":    (*Clearingway).Unflex,
		"removeall": (*Clearingway).RemoveAll,
	}
	tests := []struct {
		command      string
		memberRoles  []string
		wantRemoved  []string
		wantFollowUp string
	}{
		{"uncomfy", []string{"NA's Comfiest", "Nice", "Gold"}, []string{"-NA's Comfiest"}, "_ _\n__Uncomfy roles:__\n⮕ Removed!\n"},
		{"uncomfy", []string{"Gold"}, []string{}, "You do not have any uncomfy roles!"},
		{"uncolor", []string{"Gold", "Nice"}, []string{"-Gold"}, "_ _\n__Parsing roles:__\n⮕ Removed!\n"},
		{"uncolor", []string{"Nice"}, []string{}, "You do not have any parsing roles!"},
		{"unflex", []string{"Nice", "Gold"}, []string{"-Nice"}, "_ _\n__Misc. parsing roles:__\n⮕ Removed!\n"},
		{"removeall", []string{"Nice", "Gold", "The Legend", "unrelated"}, []string{"-Gold", "-Nice", "-The Legend"}, "_ _\n__Clearingway-related roles:__\n⮕ Removed!\n"},
		{"removeall", []string{"unrelated"}, []string{}, "You do not have any Clearingway-related roles!"},
	}
	for _, tt := range tests {
		t.Run(tt.command+" "+strings.Join(tt.memberRoles, ","), func(t *testing.T) {
			fake := &fakeDiscord{memberRoles: tt.memberRoles}
			c, s := botWithGuild(t, fake, flexGuild())
			i := command(tt.command, nil)

			run[tt.command](c, s, i)

			assert.ElementsMatch(t, tt.wantRemoved, fake.roleChanges())
			replies := fake.replies()
			require.Len(t, replies, 2)
			assert.Equal(t, tt.wantFollowUp, replies[1])
		})
	}
}

func TestInteractionCreateRoutesRemoveMenuButtons(t *testing.T) {
	fake := &fakeDiscord{memberRoles: []string{"Gold"}}
	c, s := botWithGuild(t, fake, flexGuild())

	c.InteractionCreate(s, component("menuRemove removeColor"))

	assert.Equal(t, []string{"-Gold"}, fake.roleChanges())
}

func TestInteractionCreateIgnoresMalformedCustomIDs(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, flexGuild())

	c.InteractionCreate(s, component("menuRemove"))
	c.InteractionCreate(s, component("nonsense removeColor"))
	c.InteractionCreate(s, component("menuEncounter encounterProcess pf"))

	assert.Empty(t, fake.Calls())
}

func TestHandlersIgnoreUnconfiguredGuilds(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, flexGuild())
	i := command("uncomfy", nil)
	i.GuildID = "other"

	c.Uncomfy(s, i)
	c.RemoveAll(s, i)
	c.Roles(s, i)

	assert.Empty(t, fake.Calls())
}

func TestRolesCommandListsEncountersAndUnskippedRoles(t *testing.T) {
	fake := &fakeDiscord{}
	config := flexGuild()
	config.ConfigEncounters = []*ConfigEncounter{{Name: "TOP", Ids: []int{1068}, DefaultRoles: true}}
	c, s := botWithGuild(t, fake, config)

	c.Roles(s, command("roles", nil))

	all := strings.Join(fake.replies(), "")
	assert.Contains(t, all, "⮕ TOP\n")
	assert.Contains(t, all, "__**Gold**__\n⮕ DPS parse is 100 in a relevant encounter.\n\n")
	assert.NotContains(t, all, "__**The Man**__", "skipped roles are hidden")
	for _, reply := range fake.replies() {
		assert.LessOrEqual(t, len(reply), 2000, "Discord rejects messages over 2000 characters")
	}
}

func TestToggleReclear(t *testing.T) {
	config := &ConfigGuild{Name: "Test", ConfigEncounters: []*ConfigEncounter{{Name: "TOP", Ids: []int{1068}, DefaultRoles: true}}}
	tests := []struct {
		memberRoles []string
		wantChange  []string
		wantReply   string
	}{
		{[]string{"TOP-Cleared"}, []string{"+TOP-Reclear"}, "Successfully added role: <@&TOP-Reclear>"},
		{[]string{"TOP-Cleared", "TOP-Reclear"}, []string{"-TOP-Reclear"}, "Successfully removed role: <@&TOP-Reclear>"},
		{[]string{}, []string{}, "You do not have the required role: <@&TOP-Cleared>"},
	}
	for _, tt := range tests {
		fake := &fakeDiscord{}
		c, s := botWithGuild(t, fake, config)

		c.ToggleReclear(s, command("reclears", tt.memberRoles, stringOption("ultimate", "TOP")))

		assert.Equal(t, tt.wantChange, fake.roleChanges())
		assert.Equal(t, tt.wantReply, fake.replies()[1])
	}
}

func autocomplete(world string) *discordgo.InteractionCreate {
	i := command("clears", nil, stringOption("world", world))
	i.Type = discordgo.InteractionApplicationCommandAutocomplete
	return i
}

func TestAutocompleteSuggestsMatchingWorlds(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, flexGuild())

	c.Autocomplete(s, autocomplete("Gilg"))

	assert.Equal(t, []string{"Gilgamesh"}, fake.autocompleteChoices(t))
}

func TestAutocompleteWithNoInputStillAnswersWithinDiscordsLimit(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, flexGuild())

	c.Autocomplete(s, autocomplete(""))

	choices := fake.autocompleteChoices(t)
	assert.NotEmpty(t, choices)
	assert.LessOrEqual(t, len(choices), 25, "Discord accepts at most 25 autocomplete choices")
}
