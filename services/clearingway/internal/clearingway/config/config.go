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

// parseEncounterConfig - Parses an encounter config file and adds it to the BotConfig
func (cfg *BotConfig) parseEncounterConfig(path string, data []byte) error {
	// Unmarshal encounter config
	var encounterConfig EncounterConfig
	if err := json.Unmarshal(data, &encounterConfig); err != nil {
		return fmt.Errorf("error unmarshaling encounter config file %s: %w", path, err)
	}

	// Init map if nil and add encounter config
	if cfg.Encounters == nil {
		cfg.Encounters = make(map[string]*EncounterConfig)
	}
	cfg.Encounters[encounterConfig.Name] = &encounterConfig
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

	// Read config file
	file, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error reading config file %s: %s", path, err)
	}

	// Determine config type based on directory structure
	lastDirInPath := filepath.Base(filepath.Dir(path))
	switch lastDirInPath {
	case "ultimates":
	case "savages":
	case "extremes":
		if err := cfg.parseEncounterConfig(path, file); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown config file type for file %s", path)
	}

	return nil
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
