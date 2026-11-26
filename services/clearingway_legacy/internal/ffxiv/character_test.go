package ffxiv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharactersInitRejectsShortNames(t *testing.T) {
	cs := &Characters{Characters: map[string]*Character{}}

	_, err := cs.Init("Gilgamesh", "A", "Smith")
	assert.EqualError(t, err, "first name must be at least two characters")

	_, err = cs.Init("Gilgamesh", "Alex", "S")
	assert.EqualError(t, err, "last name must be at least two characters")

	assert.Empty(t, cs.Characters)
}

func TestCharactersInitTitleCasesWorldAndReusesCharacter(t *testing.T) {
	cs := &Characters{Characters: map[string]*Character{}}

	first, err := cs.Init("gilgamesh", "alex", "smith")
	require.NoError(t, err)
	second, err := cs.Init("gilgamesh", "alex", "smith")
	require.NoError(t, err)

	assert.Equal(t, "Gilgamesh", first.World)
	assert.Same(t, first, second)
	assert.Len(t, cs.Characters, 1)
}

func TestCharacterName(t *testing.T) {
	c := &Character{FirstName: "jean-luc", LastName: "d’arc"}

	assert.Equal(t, "Jean-Luc D'arc", c.Name())
}

func TestCharacterUpdatedRecently(t *testing.T) {
	assert.True(t, (&Character{LastUpdateTime: time.Now().Add(-4 * time.Minute)}).UpdatedRecently())
	assert.False(t, (&Character{LastUpdateTime: time.Now().Add(-6 * time.Minute)}).UpdatedRecently())
	assert.False(t, (&Character{}).UpdatedRecently())
}

func TestCharacterLodestoneSlugDependsOnlyOnDiscordID(t *testing.T) {
	a := &Character{FirstName: "Alex", LastName: "Smith"}
	b := &Character{FirstName: "Other", LastName: "Person"}

	assert.Equal(t, "clearingway-152961502", a.LodestoneSlug("123456789"))
	assert.Equal(t, a.LodestoneSlug("123456789"), b.LodestoneSlug("123456789"))
	assert.NotEqual(t, a.LodestoneSlug("123456789"), a.LodestoneSlug("987654321"))
}

func TestCharacterPhysicalDatacenter(t *testing.T) {
	assert.Same(t, NA, (&Character{World: "Gilgamesh"}).PhysicalDatacenter())
	assert.Nil(t, (&Character{World: "Atlantis"}).PhysicalDatacenter())
}
