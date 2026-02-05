package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/**
 * As a general note, we do not mock any file system interactions here due to the complexity
 * of doing so in Go. Instead, we focus on testing the logic of the configuration parsing
 * and retrieval methods directly.
 *
 * If you decide you want to try, please increment this counter of hours wasted: 2
 */

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
	if result == nil {
		t.Fatal("Expected encounter config, got nil")
	}
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
	if result == nil {
		t.Fatal("Expected encounter config, got nil")
	}
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

func TestParseEncounterConfig(t *testing.T) {
	cfg := &BotConfig{}
	data := []byte(`{
		"ids": [1, 2, 3],
		"name": "Test Raid",
		"difficulty": "Savage",
		"roles": [
			{"name": "DPS", "type": "Prog", "color": "#FF0000"}
		]
	}`)

	err := cfg.parseEncounterConfig("test.json", data)

	assert.NoError(t, err)
	assert.NotNil(t, cfg.Encounters)
	assert.Equal(t, 1, len(cfg.Encounters))

	encounter, ok := cfg.Encounters["Test Raid"]
	assert.True(t, ok)
	assert.Equal(t, []int{1, 2, 3}, encounter.IDs)
	assert.Equal(t, DifficultySavage, encounter.Difficulty)
	assert.Len(t, encounter.Roles, 1)
	assert.Equal(t, "DPS", encounter.Roles[0].Name)
}

func TestParseEncounterConfig_InvalidJSON(t *testing.T) {
	cfg := &BotConfig{}
	data := []byte(`{invalid json}`)

	err := cfg.parseEncounterConfig("test.json", data)

	assert.Error(t, err)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	assert.Contains(t, err.Error(), "error unmarshaling encounter config file")
}

func TestParseEncounterConfig_InitializesNilMap(t *testing.T) {
	cfg := &BotConfig{}
	data := []byte(`{
		"ids": [1],
		"name": "Test",
		"difficulty": "Extreme",
		"roles": []
	}`)

	assert.Nil(t, cfg.Encounters)

	err := cfg.parseEncounterConfig("test.json", data)

	assert.NoError(t, err)
	assert.NotNil(t, cfg.Encounters)
}

func TestParseConfigFile_Directory(t *testing.T) {
	cfg := &BotConfig{}
	info := &mockFileInfo{isDir: true}

	err := cfg.parseConfigFile("/path/to/dir", info, nil)

	assert.NoError(t, err)
	assert.Nil(t, cfg.Encounters)
}

func TestParseConfigFile_NonJSONFile(t *testing.T) {
	cfg := &BotConfig{}
	info := &mockFileInfo{isDir: false, ext: ".yaml"}

	err := cfg.parseConfigFile("/path/to/config.yaml", info, nil)

	assert.NoError(t, err)
	assert.Nil(t, cfg.Encounters)
}

func TestParseConfigFile_ReadError(t *testing.T) {
	cfg := &BotConfig{}
	info := &mockFileInfo{isDir: false, ext: ".json"}

	err := cfg.parseConfigFile("/path/to/file.json", info, os.ErrPermission)

	assert.Error(t, err)
	assert.Equal(t, os.ErrPermission, err)
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

type mockFileInfo struct {
	isDir bool
	ext   string
}

func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) Name() string       { return "test" + m.ext }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) Sys() interface{}   { return nil }
