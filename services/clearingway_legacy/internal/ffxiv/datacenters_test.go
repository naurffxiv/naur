package ffxiv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhysicalDatacenterForWorld(t *testing.T) {
	tests := map[string]*PhysicalDatacenter{
		"Gilgamesh": NA,
		"Balmung":   NA,
		"Atlantis":  nil,
		"gilgamesh": nil, // world names are case-sensitive
	}
	for world, want := range tests {
		t.Run(world, func(t *testing.T) {
			assert.Same(t, want, PhysicalDatacenterForWorld(world))
		})
	}
}

func TestPhysicalDatacenterForAbbreviation(t *testing.T) {
	assert.Same(t, NA, PhysicalDatacenterForAbbreviation("NA"))
	assert.Same(t, EU, PhysicalDatacenterForAbbreviation("EU"))
	assert.Same(t, OC, PhysicalDatacenterForAbbreviation("OC"))
	assert.Same(t, JP, PhysicalDatacenterForAbbreviation("JP"))
	assert.Nil(t, PhysicalDatacenterForAbbreviation("na"))
}

func TestIsWorld(t *testing.T) {
	assert.True(t, IsWorld("Gilgamesh"))
	assert.False(t, IsWorld("Atlantis"))
	assert.False(t, IsWorld(""))
}

func TestEveryWorldBelongsToExactlyOnePhysicalDatacenter(t *testing.T) {
	seen := map[string]string{}
	for _, pdc := range AllPhysicalDatacenters() {
		for _, ldc := range pdc.LogicalDatacenters {
			for name := range ldc.Worlds {
				if other, ok := seen[name]; ok {
					t.Errorf("world %s is in both %s and %s", name, other, pdc.Abbreviation)
				}
				seen[name] = pdc.Abbreviation
			}
		}
	}
	assert.Len(t, AllWorlds(), len(seen))
}
