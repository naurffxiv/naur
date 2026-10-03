package config

type Difficulty string

const (
	DifficultySavage   Difficulty = "Savage"
	DifficultyExtreme  Difficulty = "Extreme"
	DifficultyUltimate Difficulty = "Ultimate"
)

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
