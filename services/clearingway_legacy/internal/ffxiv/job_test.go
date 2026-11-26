package ffxiv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobIsHealer(t *testing.T) {
	for _, name := range []string{"Astrologian", "WhiteMage", "Sage", "Scholar"} {
		assert.True(t, Jobs[name].IsHealer(), name)
	}
	for _, name := range []string{"Conjurer", "Arcanist", "Paladin", "BlackMage", "Any"} {
		assert.False(t, Jobs[name].IsHealer(), name)
	}
}

func TestJobsFullNamesAreSpelledCorrectly(t *testing.T) {
	assert.Equal(t, "Pugilist", Jobs["Pugilist"].FullName)
	assert.Equal(t, "Dark Knight", Jobs["DarkKnight"].FullName)
}
