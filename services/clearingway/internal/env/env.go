package env

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	CONFIG_PATH          string
	FFLOGS_CLIENT_ID     string
	FFLOGS_CLIENT_SECRET string
	DISCORD_TOKEN        string
	ENV                  EnvType
}

type EnvType string

const (
	Development EnvType = "development"
	Production  EnvType = "production"
)

// LoadEnv - Loads environment variables from either the system or a .env file
func LoadEnv() (*Env, error) {
	// Attempt to load .env file (ignore error if it doesn't exist)
	// This will not override existing environment variables
	_ = godotenv.Load()

	// -------------- LOAD ENV VARIABLES --------------
	// Load ENV, validate, and convert to EnvType
	envTypeStr, ok := os.LookupEnv("ENV")
	if !ok {
		return nil, errors.New("ENV not set in environment")
	}
	var envType EnvType
	switch envTypeStr {
	case string(Development):
		envType = Development
	case string(Production):
		envType = Production
	default:
		return nil, errors.New("invalid ENV value; must be 'development' or 'production'")
	}

	// Load other required env variables
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		return nil, errors.New("CONFIG_PATH not set in environment")
	}
	fflogsClientId, ok := os.LookupEnv("FFLOGS_CLIENT_ID")
	if !ok {
		return nil, errors.New("FFLOGS_CLIENT_ID not set in environment")
	}
	fflogsClientSecret, ok := os.LookupEnv("FFLOGS_CLIENT_SECRET")
	if !ok {
		return nil, errors.New("FFLOGS_CLIENT_SECRET not set in environment")
	}
	discordToken, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		return nil, errors.New("DISCORD_TOKEN not set in environment")
	}

	// -------------- RETURN ENV --------------
	return &Env{
		CONFIG_PATH:          configPath,
		FFLOGS_CLIENT_ID:     fflogsClientId,
		FFLOGS_CLIENT_SECRET: fflogsClientSecret,
		DISCORD_TOKEN:        discordToken,
		ENV:                  envType,
	}, nil
}
