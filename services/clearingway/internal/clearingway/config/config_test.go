package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Loading is tested through InitBotConfig against real folders: the committed configs/ and t.TempDir() trees.

func TestBotConfig_GetEncounters(t *testing.T) {
	cfg := &BotConfig{
		Encounters: map[string]*EncounterConfig{
			"test": {
				IDs:  []int{1},
				Name: "Test Encounter",
			},
		},
	}

	encounters := cfg.GetEncounters()

	assert.NotNil(t, encounters)
	assert.Equal(t, 1, len(encounters))
	assert.Equal(t, "Test Encounter", encounters["test"].Name)
}

func TestBotConfig_GetEncounters_Nil(t *testing.T) {
	cfg := &BotConfig{}

	encounters := cfg.GetEncounters()

	assert.Nil(t, encounters)
}

func TestBotConfig_GetEncounterByID_Found(t *testing.T) {
	cfg := &BotConfig{
		Encounters: map[string]*EncounterConfig{
			"test-encounter": {
				IDs:        []int{101, 102, 103},
				Name:       "Test Encounter",
				Difficulty: DifficultySavage,
				Roles: []EncounterRole{
					{Name: "Tank", Type: RoleTypeProg},
				},
			},
		},
	}

	result := cfg.GetEncounterByID(102)

	assert.NotNil(t, result)

	//goland:noinspection GoMaybeNil
	assert.Equal(t, "Test Encounter", result.Name)
}

func TestBotConfig_GetEncounterByID_NotFound(t *testing.T) {
	cfg := &BotConfig{
		Encounters: map[string]*EncounterConfig{
			"test-encounter": {
				IDs:  []int{101, 102},
				Name: "Test Encounter",
			},
		},
	}

	result := cfg.GetEncounterByID(999)

	assert.Nil(t, result)
}

func TestBotConfig_GetEncounterByID_EmptyConfig(t *testing.T) {
	cfg := &BotConfig{}

	result := cfg.GetEncounterByID(1)

	assert.Nil(t, result)
}

func TestBotConfig_GetEncounterByName_Found(t *testing.T) {
	cfg := &BotConfig{
		Encounters: map[string]*EncounterConfig{
			"p12s": {
				IDs:  []int{1006},
				Name: "P12S",
			},
		},
	}

	result := cfg.GetEncounterByName("p12s")

	assert.NotNil(t, result)

	//goland:noinspection GoMaybeNil
	assert.Equal(t, "P12S", result.Name)
}

func TestBotConfig_GetEncounterByName_NotFound(t *testing.T) {
	cfg := &BotConfig{
		Encounters: map[string]*EncounterConfig{
			"p12s": {
				IDs:  []int{1006},
				Name: "P12S",
			},
		},
	}

	result := cfg.GetEncounterByName("nonexistent")

	assert.Nil(t, result)
}

func TestBotConfig_GetEncounterByName_NilEncounters(t *testing.T) {
	cfg := &BotConfig{}

	result := cfg.GetEncounterByName("test")

	assert.Nil(t, result)
}

func TestInitBotConfig_LoadsEveryCommittedEncounter(t *testing.T) {
	configsDir := filepath.Join("..", "..", "..", "configs")
	cfg, err := InitBotConfig(configsDir)
	require.NoError(t, err)

	files, err := filepath.Glob(filepath.Join(configsDir, "*", "*.json"))
	require.NoError(t, err)
	assert.Len(t, cfg.GetEncounters(), len(files), "every committed encounter file loads")

	// One of each Difficulty, so a loader that drops a whole tier fails here.
	want := map[string]Difficulty{
		"The Omega Protocol (Ultimate)": DifficultyUltimate,
		"M8S":                           DifficultySavage,
		"The Windward Wilds (Extreme)":  DifficultyExtreme,
	}
	for name, difficulty := range want {
		encounter := cfg.GetEncounterByName(name)
		if assert.NotNil(t, encounter, "encounter %q was not loaded", name) {
			assert.Equal(t, difficulty, encounter.Difficulty, "difficulty of %q", name)
		}
	}
}

// writeConfig - Writes contents to dir/relPath, creating any folders on the way.
func writeConfig(t *testing.T, dir, relPath, contents string) {
	t.Helper()
	path := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
}

func TestInitBotConfig_DifficultyComesFromTheFileNotTheFolder(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "savages/misfiled.json", `{"ids": [1], "name": "Misfiled", "difficulty": "Ultimate", "roles": []}`)
	writeConfig(t, dir, "anything/nested/deeper.json", `{"ids": [2], "name": "Deeper", "difficulty": "Extreme", "roles": []}`)

	cfg, err := InitBotConfig(dir)

	require.NoError(t, err)
	assert.Equal(t, DifficultyUltimate, cfg.GetEncounterByName("Misfiled").Difficulty)
	assert.Equal(t, DifficultyExtreme, cfg.GetEncounterByName("Deeper").Difficulty)
	assert.Equal(t, "Deeper", cfg.GetEncounterByID(2).Name)
}

func TestInitBotConfig_SkipsNonJSONFiles(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "examples/encounter.json.example", `{invalid`)
	writeConfig(t, dir, "notes.yaml", `not: json`)

	cfg, err := InitBotConfig(dir)

	require.NoError(t, err)
	assert.Empty(t, cfg.GetEncounters())
}

func TestInitBotConfig_Rejects(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantErr string
	}{
		{
			name:    "invalid JSON",
			files:   map[string]string{"a.json": `{invalid json}`},
			wantErr: "error unmarshaling encounter config file",
		},
		{
			name:    "missing difficulty",
			files:   map[string]string{"a.json": `{"ids": [1], "name": "A", "roles": []}`},
			wantErr: `has difficulty ""`,
		},
		{
			name:    "unknown difficulty",
			files:   map[string]string{"a.json": `{"ids": [1], "name": "A", "difficulty": "Normal", "roles": []}`},
			wantErr: `has difficulty "Normal"`,
		},
		{
			name:    "missing name",
			files:   map[string]string{"a.json": `{"ids": [1], "difficulty": "Savage", "roles": []}`},
			wantErr: "has no name",
		},
		{
			name: "duplicate name",
			files: map[string]string{
				"a.json": `{"ids": [1], "name": "Same", "difficulty": "Savage", "roles": []}`,
				"b.json": `{"ids": [2], "name": "Same", "difficulty": "Savage", "roles": []}`,
			},
			wantErr: `reuses the encounter name "Same"`,
		},
		{
			name: "duplicate FFLogs ID",
			files: map[string]string{
				"a.json": `{"ids": [1, 7], "name": "A", "difficulty": "Savage", "roles": []}`,
				"b.json": `{"ids": [7], "name": "B", "difficulty": "Savage", "roles": []}`,
			},
			wantErr: "reuses FFLogs encounter ID 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for path, contents := range tt.files {
				writeConfig(t, dir, path, contents)
			}

			cfg, err := InitBotConfig(dir)

			assert.Nil(t, cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestInitBotConfig_MissingDirectory(t *testing.T) {
	cfg, err := InitBotConfig(filepath.Join(t.TempDir(), "does-not-exist"))

	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestDifficulty_Values(t *testing.T) {
	assert.Equal(t, Difficulty("Savage"), DifficultySavage)
	assert.Equal(t, Difficulty("Extreme"), DifficultyExtreme)
	assert.Equal(t, Difficulty("Ultimate"), DifficultyUltimate)
}

func TestRoleType_Values(t *testing.T) {
	assert.Equal(t, RoleType("Cleared"), RoleTypeCleared)
	assert.Equal(t, RoleType("Prog"), RoleTypeProg)
	assert.Equal(t, RoleType("Reclear"), RoleTypeRecleared)
	assert.Equal(t, RoleType("C4X"), RoleTypeC4X)
	assert.Equal(t, RoleType("Name Color"), RoleTypeNameColor)
}

func TestEncounterConfig_JSON(t *testing.T) {
	jsonData := `{
		"ids": [1, 2],
		"name": "Test",
		"difficulty": "Savage",
		"roles": [
			{"name": "Tank", "type": "Prog", "color": "#000000"}
		]
	}`

	var config EncounterConfig
	err := json.Unmarshal([]byte(jsonData), &config)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, config.IDs)
	assert.Equal(t, "Test", config.Name)
	assert.Equal(t, DifficultySavage, config.Difficulty)
	assert.Len(t, config.Roles, 1)
	assert.Equal(t, "Tank", config.Roles[0].Name)
}

func TestEncounterRole_JSON(t *testing.T) {
	jsonData := `{"name": "Healer", "type": "Reclear", "color": "#FFFFFF"}`

	var role EncounterRole
	err := json.Unmarshal([]byte(jsonData), &role)

	assert.NoError(t, err)
	assert.Equal(t, "Healer", role.Name)
	assert.Equal(t, RoleTypeRecleared, role.Type)
	assert.Equal(t, "#FFFFFF", role.Color)
}
