package ffxiv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============== Worlds ==============

func TestAllWorlds_ContainsKnownWorlds(t *testing.T) {
	worlds := AllWorlds()

	assert.Contains(t, worlds, "Gilgamesh")
	assert.Contains(t, worlds, "Golem")
	assert.Contains(t, worlds, "Twintania")
	assert.Contains(t, worlds, "Tycoon")
	assert.Contains(t, worlds, "Ravana")
	assert.Contains(t, worlds, "Tonberry")
}

func TestAllWorlds_NoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, world := range AllWorlds() {
		assert.False(t, seen[world], "duplicate world %s", world)
		seen[world] = true
	}
}

func TestAllWorlds_ReturnsCopy(t *testing.T) {
	AllWorlds()[0] = "Narnia"
	assert.False(t, IsWorld("Narnia"))
}

func TestFindWorld(t *testing.T) {
	world, ok := FindWorld("  gilgaMESH ")
	assert.True(t, ok)
	assert.Equal(t, "Gilgamesh", world)

	world, ok = FindWorld("Narnia")
	assert.False(t, ok)
	assert.Empty(t, world)
}

func TestIsWorld(t *testing.T) {
	assert.True(t, IsWorld("Cactuar"))
	assert.True(t, IsWorld("cactuar"))
	assert.False(t, IsWorld("Narnia"))
	assert.False(t, IsWorld(""))
}

// ============== Character ==============

func TestNewCharacter_Normalises(t *testing.T) {
	char, err := NewCharacter("  alicja ", "TESTER", "gilgamesh")

	assert.NoError(t, err)
	assert.Equal(t, "Gilgamesh", char.World)
	assert.Equal(t, "Alicja", char.FirstName)
	assert.Equal(t, "Tester", char.LastName)
	assert.Equal(t, "Alicja Tester", char.Name())
	assert.Equal(t, int64(0), char.LodestoneID)
}

func TestNewCharacter_ReplacesTypographicApostrophe(t *testing.T) {
	char, err := NewCharacter("Y’shtola", "Rhul", "Gilgamesh")

	assert.NoError(t, err)
	assert.Equal(t, "Y'shtola Rhul", char.Name())
}

func TestNewCharacter_InvalidWorld(t *testing.T) {
	char, err := NewCharacter("Alicja", "Tester", "Narnia")

	assert.ErrorIs(t, err, ErrInvalidWorld)
	assert.Nil(t, char)
}

func TestNewCharacter_ShortNames(t *testing.T) {
	_, err := NewCharacter("A", "Tester", "Gilgamesh")
	assert.ErrorIs(t, err, ErrInvalidName)

	_, err = NewCharacter("Alicja", " T ", "Gilgamesh")
	assert.ErrorIs(t, err, ErrInvalidName)
}
