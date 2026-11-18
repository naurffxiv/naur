package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type BotConfig struct {
	Encounters map[string]*EncounterConfig `json:"encounters"`
}

// GetEncounters - Returns all encounter configurations
func (cfg *BotConfig) GetEncounters() map[string]*EncounterConfig {
	return cfg.Encounters
}

// GetEncounterByID - Retrieves an encounter configuration by its ID
func (cfg *BotConfig) GetEncounterByID(id int) *EncounterConfig {
	for _, encounter := range cfg.Encounters {
		for _, encounterID := range encounter.IDs {
			if encounterID == id {
				return encounter
			}
		}
	}
	return nil
}

// GetEncounterByName - Retrieves an encounter configuration by its name
func (cfg *BotConfig) GetEncounterByName(name string) *EncounterConfig {
	if encounter, ok := cfg.Encounters[name]; ok {
		return encounter
	}
	return nil
}

// parseEncounterConfig - Parses an encounter config file and adds it to the BotConfig.
// The encounter's own "difficulty" field decides its Difficulty; the folder it sits in does not matter.
func (cfg *BotConfig) parseEncounterConfig(path string, data []byte) error {
	var encounterConfig EncounterConfig
	if err := json.Unmarshal(data, &encounterConfig); err != nil {
		return fmt.Errorf("error unmarshaling encounter config file %s: %w", path, err)
	}

	if err := cfg.validateEncounter(path, &encounterConfig); err != nil {
		return err
	}

	if cfg.Encounters == nil {
		cfg.Encounters = make(map[string]*EncounterConfig)
	}
	cfg.Encounters[encounterConfig.Name] = &encounterConfig
	return nil
}

// validateEncounter - Rejects an encounter with no name, an unknown difficulty, or a name or FFLogs ID
// that an already loaded encounter uses.
func (cfg *BotConfig) validateEncounter(path string, encounter *EncounterConfig) error {
	if encounter.Name == "" {
		return fmt.Errorf("encounter config file %s has no name", path)
	}

	if !encounter.Difficulty.Valid() {
		return fmt.Errorf("encounter config file %s has difficulty %q, want one of %q",
			path, encounter.Difficulty, difficulties)
	}

	if _, ok := cfg.Encounters[encounter.Name]; ok {
		return fmt.Errorf("encounter config file %s reuses the encounter name %q", path, encounter.Name)
	}

	for _, id := range encounter.IDs {
		if existing := cfg.GetEncounterByID(id); existing != nil {
			return fmt.Errorf("encounter config file %s reuses FFLogs encounter ID %d from %q", path, id, existing.Name)
		}
	}

	return nil
}

// parseConfigFile - Callback function for filepath.Walk to parse config files
func (cfg *BotConfig) parseConfigFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Skip non-JSON files and directories
	if info.IsDir() || filepath.Ext(path) != ".json" {
		return nil
	}

	file, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error reading config file %s: %s", path, err)
	}

	return cfg.parseEncounterConfig(path, file)
}

// InitBotConfig - Initializes the bot configuration by loading all config files from the specified directory
func InitBotConfig(configDir string) (*BotConfig, error) {
	cfg := &BotConfig{}
	// This "walks" the config directory and processes each file using parseConfigFile
	// "Walking" means it goes recursively through all subdirectories and files
	err := filepath.Walk(configDir, cfg.parseConfigFile)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
