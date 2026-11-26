package clearingway

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMenuInitBuildsMainMenuButtons(t *testing.T) {
	m := &Menu{}
	m.Init(&ConfigMenu{Name: "main", Type: "menuMain", Title: "Welcome", ConfigButtons: []*ConfigButton{
		{Label: "Verify", MenuType: "menuVerify", MenuName: "menuVerify"},
		{Label: "Remove", Style: 4, MenuType: "menuRemove", MenuName: "menuRemove"},
		{Label: "Prog", MenuType: "menuEncounter", MenuName: "prog"},
		{Label: "", MenuType: "menuVerify", MenuName: "menuVerify"},
		{Label: "Unknown", MenuType: "menuMain", MenuName: "x"},
		{Label: "No target"},
	}})

	ids := []string{}
	for _, b := range m.Buttons {
		ids = append(ids, fmt.Sprintf("%s|%d|%s", b.Label, b.Style, b.CustomID))
	}
	assert.Equal(t, []string{
		"Verify|1|menuVerify menu",
		"Remove|4|menuRemove menu",
		"Prog|1|menuEncounter menu prog",
	}, ids)
}

func TestPopulateButtonsDedupesAndWrapsRowsOfFive(t *testing.T) {
	buttons := []discordgo.Button{}
	for i := 0; i < 30; i++ {
		buttons = append(buttons, discordgo.Button{CustomID: fmt.Sprint(i)})
	}
	buttons = append([]discordgo.Button{{CustomID: "0"}}, buttons...)

	rows := PopulateButtons(buttons)

	require.Len(t, rows, 5, "Discord allows at most 5 rows of 5 buttons")
	for _, row := range rows {
		assert.Len(t, row.(discordgo.ActionsRow).Components, 5)
	}
	assert.Equal(t, "0", rows[0].(discordgo.ActionsRow).Components[0].(discordgo.Button).CustomID)
	assert.Equal(t, "1", rows[0].(discordgo.ActionsRow).Components[1].(discordgo.Button).CustomID)
}

func menuGuild() *ConfigGuild {
	return &ConfigGuild{
		Name:             "Test",
		ConfigRoles:      &ConfigRoles{Menu: true, RelevantFlexing: true},
		ConfigEncounters: []*ConfigEncounter{{Name: "TOP", Ids: []int{1068}, DefaultRoles: true}},
		ConfigMenus: []*ConfigMenu{
			{Name: "pf", Type: "menuEncounter", Title: "PF roles", Description: "Pick PF roles", RoleType: []string{"PF"}, RequireClear: true, MultiSelect: true},
			{Name: "extra", Type: "menuMain", Title: "Extra", ConfigButtons: []*ConfigButton{{Label: "PF", MenuType: "menuEncounter", MenuName: "pf"}}},
		},
		ConfigMenuOrder: []ConfigMenuOrder{{Name: "all", Menus: []string{"menuMain", "extra"}}},
	}
}

func TestMenuMainSendPostsTheMenuToTheChannel(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, menuGuild())

	c.MenuMainSend(s, command("menu", nil, stringOption("menu", "menuMain")))

	posts := fake.callsTo(http.MethodPost, "/channels/c1/messages")
	require.Len(t, posts, 1)
	embed := posts[0].Body["embeds"].([]any)[0].(map[string]any)
	assert.Equal(t, "Welcome to Test", embed["title"])
	assert.Equal(t, []string{"Sent menu message(s)."}, fake.replies())
}

func TestMenuMainSendGroupPostsEveryMenuInTheGroup(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, menuGuild())

	c.MenuMainSend(s, command("menu", nil, stringOption("menu", "group all")))

	assert.Len(t, fake.callsTo(http.MethodPost, "/channels/c1/messages"), 2)
}

func TestMenuMainSendRejectsUnknownMenus(t *testing.T) {
	for option, want := range map[string]string{
		"nope":       "Unable to find menu.",
		"group nope": "Unable to find menu group.",
		"pf":         "Menu is not of type menuMain.",
	} {
		fake := &fakeDiscord{}
		c, s := botWithGuild(t, fake, menuGuild())

		c.MenuMainSend(s, command("menu", nil, stringOption("menu", option)))

		assert.Empty(t, fake.callsTo(http.MethodPost, "/channels/c1/messages"), option)
		assert.Equal(t, want, fake.replies()[0], option)
	}
}

func TestMenuAutocompleteListsMenusAndGroups(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, menuGuild())
	i := command("menu", nil, stringOption("menu", "gro"))
	i.Type = discordgo.InteractionApplicationCommandAutocomplete

	c.MenuAutocomplete(s, i)

	assert.Equal(t, []string{"group all"}, fake.autocompleteChoices(t))
}

func TestRemoveMenuOffersRemoveButtons(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, menuGuild())
	c.Guilds.Guilds["g1"].Menus.Menus["menuRemove"].FinalizeButtons() // done by DiscordReady

	c.InteractionCreate(s, component("menuRemove menu"))

	calls := fake.Calls()
	require.Len(t, calls, 1)
	data := calls[0].Body["data"].(map[string]any)
	assert.Equal(t, float64(discordgo.MessageFlagsEphemeral), data["flags"])
	row := data["components"].([]any)[0].(map[string]any)["components"].([]any)
	assert.Equal(t, "menuRemove removeAll", row[0].(map[string]any)["custom_id"])
	assert.Equal(t, "menuRemove removeFlex", row[1].(map[string]any)["custom_id"])
}

func TestVerifyMenuOpensTheCharacterModal(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, menuGuild())

	c.InteractionCreate(s, component("menuVerify menu"))

	calls := fake.Calls()
	require.Len(t, calls, 1)
	assert.Equal(t, float64(discordgo.InteractionResponseModal), calls[0].Body["type"])
	assert.Equal(t, "menuVerify clearsModal", calls[0].Body["data"].(map[string]any)["custom_id"])
}

func TestVerifyModalSubmissionRunsClears(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := botWithGuild(t, fake, menuGuild())
	input := func(id, value string) discordgo.MessageComponent {
		return &discordgo.ActionsRow{Components: []discordgo.MessageComponent{&discordgo.TextInput{CustomID: id, Value: value}}}
	}
	i := component("")
	i.Type = discordgo.InteractionModalSubmit
	i.Data = discordgo.ModalSubmitInteractionData{CustomID: "menuVerify clearsModal", Components: []discordgo.MessageComponent{
		input("firstName", "Alex"), input("lastName", "Smith"), input("world", "Atlantis"),
	}}

	c.InteractionCreate(s, i)

	assert.Equal(t, []string{"Received character info...", "`Atlantis` is not a valid world! Make sure you spelled your world name properly."}, fake.replies())
}

func encounterMenuBot(t *testing.T, fake *fakeDiscord) (*Clearingway, *discordgo.Session) {
	t.Helper()
	c, s := botWithGuild(t, fake, menuGuild())
	g := c.Guilds.Guilds["g1"]
	menu := g.Menus.Menus["pf"]
	menu.MenuEncounterInit(g.Encounters, menu.AdditionalData.RoleType)
	return c, s
}

func TestEncounterMenuPreselectsRolesTheMemberHas(t *testing.T) {
	fake := &fakeDiscord{}
	c, s := encounterMenuBot(t, fake)
	i := component("menuEncounter menu pf")
	i.Member.Roles = []string{"TOP-PF"}

	c.InteractionCreate(s, i)

	calls := fake.Calls()
	require.Len(t, calls, 1)
	data := calls[0].Body["data"].(map[string]any)
	assert.Contains(t, data["embeds"].([]any)[0].(map[string]any)["description"], "Pick PF roles\n### Available roles\n- <@&TOP-PF>")
	selectMenu := data["components"].([]any)[0].(map[string]any)["components"].([]any)[0].(map[string]any)
	assert.Equal(t, "menuEncounter encounterProcess pf 0", selectMenu["custom_id"])
	option := selectMenu["options"].([]any)[0].(map[string]any)
	assert.Equal(t, "TOP-PF", option["value"])
	assert.Equal(t, true, option["default"])
}

func processEncounterMenu(t *testing.T, memberRoles, selected []string) (*fakeDiscord, string) {
	t.Helper()
	fake := &fakeDiscord{}
	c, s := encounterMenuBot(t, fake)
	i := component("menuEncounter encounterProcess pf 0")
	i.Member.Roles = memberRoles
	i.Data = discordgo.MessageComponentInteractionData{CustomID: "menuEncounter encounterProcess pf 0", Values: selected}

	c.InteractionCreate(s, i)

	edits := fake.callsTo(http.MethodPatch, "/webhooks/app/tok/messages/@original")
	require.Len(t, edits, 1)
	return fake, edits[0].Body["content"].(string)
}

func TestEncounterMenuProcess(t *testing.T) {
	tests := []struct {
		name        string
		memberRoles []string
		selected    []string
		wantChanges []string
		wantReply   string
	}{
		{"needs the cleared role", nil, []string{"TOP-PF"}, []string{}, "You do not have the required role: <@&TOP-Cleared>\n"},
		{"adds when cleared", []string{"TOP-Cleared"}, []string{"TOP-PF"}, []string{"+TOP-PF"}, "Successfully added role: <@&TOP-PF>\n"},
		{"removes when deselected", []string{"TOP-Cleared", "TOP-PF"}, nil, []string{"-TOP-PF"}, "Successfully removed role: <@&TOP-PF>\n"},
		{"no change", []string{"TOP-Cleared", "TOP-PF"}, []string{"TOP-PF"}, []string{}, "Nothing has been done!\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake, reply := processEncounterMenu(t, tt.memberRoles, tt.selected)
			assert.Equal(t, tt.wantChanges, fake.roleChanges())
			assert.Equal(t, tt.wantReply, reply)
		})
	}
}
