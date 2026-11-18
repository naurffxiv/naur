package config

import "slices"

type Difficulty string

const (
	DifficultySavage   Difficulty = "Savage"
	DifficultyExtreme  Difficulty = "Extreme"
	DifficultyUltimate Difficulty = "Ultimate"
)

// difficulties lists every known Difficulty.
var difficulties = []Difficulty{DifficultySavage, DifficultyExtreme, DifficultyUltimate}

// Valid reports whether d is a known Difficulty.
func (d Difficulty) Valid() bool {
	return slices.Contains(difficulties, d)
}

type RoleType string

const (
	RoleTypeCleared   RoleType = "Cleared"
	RoleTypeProg      RoleType = "Prog"
	RoleTypeRecleared RoleType = "Reclear"
	RoleTypeC4X       RoleType = "C4X"
	RoleTypeNameColor RoleType = "Name Color"
)

type EncounterRole struct {
	Name  string   `json:"name"`
	Type  RoleType `json:"type"`
	Color string   `json:"color"`
}

type EncounterConfig struct {
	IDs        []int           `json:"ids"`
	Name       string          `json:"name"`
	Difficulty Difficulty      `json:"difficulty"`
	Roles      []EncounterRole `json:"roles"`
}
