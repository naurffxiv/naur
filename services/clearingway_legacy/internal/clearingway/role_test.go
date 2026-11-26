package clearingway

import (
	"net/http"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleEnsureReusesExistingRoleByName(t *testing.T) {
	fake := &fakeDiscord{}
	r := &Role{Name: "Gold"}
	existing := &discordgo.Role{ID: "42", Name: "Gold"}

	require.NoError(t, r.Ensure("g1", fake.session(t), []*discordgo.Role{{ID: "1", Name: "Pink"}, existing}))

	assert.Same(t, existing, r.DiscordRole)
	assert.Empty(t, fake.Calls())
}

func TestRoleEnsureCreatesMissingRole(t *testing.T) {
	fake := &fakeDiscord{}
	r := &Role{Name: "Gold", Color: 0xe1cc8a, Hoist: true, Mention: true}

	require.NoError(t, r.Ensure("g1", fake.session(t), nil))

	calls := fake.Calls()
	require.Len(t, calls, 1)
	assert.Equal(t, http.MethodPost, calls[0].Method)
	assert.Equal(t, "/guilds/g1/roles", calls[0].Path)
	assert.Equal(t, map[string]any{"name": "Gold", "color": float64(0xe1cc8a), "permissions": "0", "hoist": true, "mentionable": true}, calls[0].Body)
	assert.Equal(t, "new-Gold", r.DiscordRole.ID)
}

func TestRoleEnsureSkipsSkippedRoles(t *testing.T) {
	fake := &fakeDiscord{}
	r := &Role{Name: "The Man", Skip: true}

	require.NoError(t, r.Ensure("g1", fake.session(t), nil))

	assert.Nil(t, r.DiscordRole)
	assert.Empty(t, fake.Calls())
}

func TestRoleAddAndRemoveFromCharacter(t *testing.T) {
	fake := &fakeDiscord{}
	s := fake.session(t)
	r := &Role{DiscordRole: &discordgo.Role{ID: "42"}}

	require.NoError(t, r.AddToCharacter("g1", "u1", s))
	require.NoError(t, r.RemoveFromCharacter("g1", "u1", s))

	assert.Equal(t, []string{"+42", "-42"}, fake.roleChanges())
}

func TestRolesReorderSendsRolesInOrder(t *testing.T) {
	fake := &fakeDiscord{}
	rs := &Roles{Roles: []*Role{{DiscordRole: &discordgo.Role{ID: "b"}}, {DiscordRole: &discordgo.Role{ID: "a"}}}}

	require.NoError(t, rs.Reorder("g1", fake.session(t)))

	calls := fake.Calls()
	require.Len(t, calls, 1)
	assert.Equal(t, http.MethodPatch, calls[0].Method)
	assert.Equal(t, "/guilds/g1/roles", calls[0].Path)
}

func TestRolesLookups(t *testing.T) {
	gold := &Role{Name: "Gold", DiscordRole: &discordgo.Role{ID: "1"}}
	pink := &Role{Name: "Pink", DiscordRole: &discordgo.Role{ID: "2"}}
	rs := &Roles{Roles: []*Role{gold, pink}}

	assert.Same(t, pink, rs.FindByName("Pink"))
	assert.Nil(t, rs.FindByName("Blue"))
	found, i := rs.IndexOfRole(&Role{Name: "Pink"})
	assert.True(t, found)
	assert.Equal(t, 1, i)
	found, _ = rs.IndexOfRole(&Role{Name: "Blue"})
	assert.False(t, found)
	assert.Equal(t, []*Role{pink}, rs.InDiscordRoles([]string{"2", "9"}))
	assert.True(t, gold.PresentInRoles([]string{"1"}))
	assert.False(t, gold.PresentInRoles([]string{"2"}))
}

func TestRolePhase(t *testing.T) {
	assert.Equal(t, "Phase **3**", (&Role{Type: ProgRole}).Phase(3))
	assert.Equal(t, "**Cleared**", (&Role{Type: ClearedRole}).Phase(3))
}

func TestDiscordReadyEnsuresRolesAndRegistersCommands(t *testing.T) {
	readyDelay = 0
	t.Cleanup(func() { readyDelay = time.Second })
	fake := &fakeDiscord{guildRoles: []*discordgo.Role{{ID: "1", Name: "Gold"}}}
	config := &ConfigGuild{Name: "Test", ConfigRoles: &ConfigRoles{RelevantParsing: true}}
	config.GuildId, config.ChannelId = "g1", "c1"
	c := &Clearingway{Config: &Config{ConfigGuilds: []*ConfigGuild{config}}}
	c.Init()
	s := fake.session(t)

	c.DiscordReady(s, &discordgo.Ready{User: &discordgo.User{ID: "app"}, Guilds: []*discordgo.Guild{{ID: "g1"}, {ID: "unconfigured"}}})

	assert.True(t, c.Ready)
	roles := c.Guilds.Guilds["g1"].RelevantParsingRoles
	assert.Equal(t, "1", roles.FindByName("Gold").DiscordRole.ID)
	assert.Equal(t, "new-Pink", roles.FindByName("Pink").DiscordRole.ID)

	assert.Len(t, fake.callsTo(http.MethodPut, "/applications/app/guilds/g1/commands"), 1)
}
