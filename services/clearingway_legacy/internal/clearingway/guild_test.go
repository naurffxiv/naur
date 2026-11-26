package clearingway

import (
	"os"
	"testing"

	"github.com/Veraticus/clearingway/internal/ffxiv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCommittedConfigInitializesEveryGuild(t *testing.T) {
	raw, err := os.ReadFile("../../config.yaml")
	require.NoError(t, err)
	c := &Clearingway{Config: &Config{}}
	require.NoError(t, yaml.Unmarshal(raw, c.Config))

	c.Init()

	require.NotEmpty(t, c.Guilds.Guilds)
	for _, g := range c.Guilds.Guilds {
		for _, e := range g.Encounters.Encounters {
			require.NotNil(t, e.Roles[ClearedRole], "%s: %s has no cleared role", g.Name, e.Name)
		}
		seen := map[string]bool{}
		for _, r := range g.AllRoles() {
			if r.Skip || r.Type == LimboRole || r.Type == CompleteRole {
				continue // limbo and complete roles share a name across encounters
			}
			assert.False(t, seen[r.Name], "%s: duplicate role name %q", g.Name, r.Name)
			seen[r.Name] = true
		}
	}
	assert.Len(t, c.AllWorlds, len(ffxiv.AllWorlds()))
}

func TestGuildInitReconfiguresRoles(t *testing.T) {
	g := &Guild{}
	g.Init(&ConfigGuild{
		Name: "Test", GuildId: "1",
		ConfigRoles:               &ConfigRoles{Legend: true, RelevantFlexing: true, Datacenter: true},
		ConfigPhysicalDatacenters: []*ConfigPhysicalDatacenter{{Name: "NA", LogicalDatacenters: []*ConfigLogicalDatacenter{{From: "Aether", To: "aether", Color: 1, Hoist: true}}}},
		ConfigReconfigureRoles: []*ConfigReconfigureRoles{
			{From: "Nice", To: "nice", Color: 2},
			{From: "The Legend", Skip: true},
			{From: "The Man", DontSkip: true},
		},
	})

	roles := &Roles{Roles: g.AllRoles()}
	require.NotNil(t, roles.FindByName("nice"))
	assert.Equal(t, 2, roles.FindByName("nice").Color)
	assert.Nil(t, roles.FindByName("Nice"))
	assert.True(t, roles.FindByName("The Legend").Skip)
	assert.False(t, roles.FindByName("The Man").Skip)
	aether := roles.FindByName("aether")
	require.NotNil(t, aether)
	assert.True(t, aether.Hoist)
	assert.Equal(t, 1, aether.Color)

	assert.True(t, g.LegendEnabled)
	assert.False(t, g.RelevantParsingEnabled)
	assert.Nil(t, g.RelevantParsingRoles)
	assert.False(t, g.IsProgEnabled())
}

func TestDatacenterRolesMatchTheCharactersWorld(t *testing.T) {
	pds := &PhysicalDatacenters{PhysicalDatacenters: map[string]*PhysicalDatacenter{}}
	pds.Init([]*ConfigPhysicalDatacenter{{Name: "NA"}})
	opts := &ShouldApplyOpts{Character: &ffxiv.Character{FirstName: "Alex", LastName: "Smith", World: "Gilgamesh"}}

	assert.Equal(t, []string{"Aether"}, matching(pds.AllRoles(), opts))
	_, msg := applies(t, pds.AllRoles(), "Aether", opts)
	assert.Equal(t, "Character `Alex Smith (Gilgamesh)` is in the Aether datacenter in North America.", msg)
}

func TestAchievementInitAppliesConfigRoles(t *testing.T) {
	a := &Achievement{}
	a.Init(&ConfigAchievement{Title: "Bring It On", ConfigRoles: []*ConfigRole{{Type: "Cleared", Name: "brought it", Description: "d", Color: 3}}})
	as := &Achievements{Achievements: []*Achievement{a}}

	role := as.Roles().FindByName("brought it")
	require.NotNil(t, role)
	assert.Equal(t, ClearedRole, role.Type)
	assert.Equal(t, "d", role.Description)
	assert.Equal(t, 3, role.Color)
}

func TestCleanReportID(t *testing.T) {
	for in, want := range map[string]string{
		"abc123":                                        "abc123",
		"https://www.fflogs.com/reports/abc123":         "abc123",
		"https://www.fflogs.com/reports/abc123/":        "abc123",
		"https://www.fflogs.com/reports/abc123#fight=4": "abc123",
	} {
		assert.Equal(t, want, CleanReportId(in), in)
	}
}

func TestCleanWorld(t *testing.T) {
	assert.Equal(t, "Gilgamesh", cleanWorld(" Gilgamesh! "))
}
