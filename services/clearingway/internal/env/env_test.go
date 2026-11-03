package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func cleanupEnv() {
	_ = os.Unsetenv("ENV")
	_ = os.Unsetenv("CONFIG_PATH")
	_ = os.Unsetenv("FFLOGS_CLIENT_ID")
	_ = os.Unsetenv("FFLOGS_CLIENT_SECRET")
	_ = os.Unsetenv("DISCORD_TOKEN")
	_ = os.Unsetenv("DISCORD_TEST_GUILD_ID")
}

func TestLoadEnv_Success_Development(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "development")
	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "test-client-id")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "test-client-secret")
	_ = os.Setenv("DISCORD_TOKEN", "test-discord-token")
	_ = os.Setenv("DISCORD_TEST_GUILD_ID", "test-guild-id")

	env, err := LoadEnv()

	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, Development, env.ENV)
	assert.Equal(t, "/etc/clearingway/config.yaml", env.CONFIG_PATH)
	assert.Equal(t, "test-client-id", env.FFLOGS_CLIENT_ID)
	assert.Equal(t, "test-client-secret", env.FFLOGS_CLIENT_SECRET)
	assert.Equal(t, "test-discord-token", env.DISCORD_TOKEN)
	assert.Equal(t, "test-guild-id", env.DISCORD_TEST_GUILD_ID)
}

func TestLoadEnv_Success_Production(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "production")
	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/prod-config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "prod-client-id")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "prod-client-secret")
	_ = os.Setenv("DISCORD_TOKEN", "prod-discord-token")

	env, err := LoadEnv()

	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, Production, env.ENV)
	assert.Equal(t, "/etc/clearingway/prod-config.yaml", env.CONFIG_PATH)
	assert.Equal(t, "prod-client-id", env.FFLOGS_CLIENT_ID)
	assert.Equal(t, "prod-client-secret", env.FFLOGS_CLIENT_SECRET)
	assert.Equal(t, "prod-discord-token", env.DISCORD_TOKEN)
	assert.Equal(t, "", env.DISCORD_TEST_GUILD_ID)
}

func TestLoadEnv_Error_ENVNotSet(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "test-client-id")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "test-client-secret")
	_ = os.Setenv("DISCORD_TOKEN", "test-discord-token")

	env, err := LoadEnv()

	assert.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "ENV not set in environment")
}

func TestLoadEnv_Error_InvalidENV(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "invalid")
	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "test-client-id")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "test-client-secret")
	_ = os.Setenv("DISCORD_TOKEN", "test-discord-token")

	env, err := LoadEnv()

	assert.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "invalid ENV value")
}

func TestLoadEnv_Error_CONFIG_PATHNotSet(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "development")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "test-client-id")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "test-client-secret")
	_ = os.Setenv("DISCORD_TOKEN", "test-discord-token")

	env, err := LoadEnv()

	assert.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "CONFIG_PATH not set in environment")
}

func TestLoadEnv_Error_FFLOGS_CLIENT_IDNotSet(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "development")
	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "test-client-secret")
	_ = os.Setenv("DISCORD_TOKEN", "test-discord-token")

	env, err := LoadEnv()

	assert.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "FFLOGS_CLIENT_ID not set in environment")
}

func TestLoadEnv_Error_FFLOGS_CLIENT_SECRETNotSet(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "development")
	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "test-client-id")
	_ = os.Setenv("DISCORD_TOKEN", "test-discord-token")

	env, err := LoadEnv()

	assert.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "FFLOGS_CLIENT_SECRET not set in environment")
}

func TestLoadEnv_Error_DISCORD_TOKENNotSet(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "development")
	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "test-client-id")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "test-client-secret")

	env, err := LoadEnv()

	assert.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "DISCORD_TOKEN not set in environment")
}

func TestLoadEnv_OptionalDISCORD_TEST_GUILD_IDNotSet(t *testing.T) {
	defer cleanupEnv()

	_ = os.Setenv("ENV", "development")
	_ = os.Setenv("CONFIG_PATH", "/etc/clearingway/config.yaml")
	_ = os.Setenv("FFLOGS_CLIENT_ID", "test-client-id")
	_ = os.Setenv("FFLOGS_CLIENT_SECRET", "test-client-secret")
	_ = os.Setenv("DISCORD_TOKEN", "test-discord-token")

	env, err := LoadEnv()

	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, "", env.DISCORD_TEST_GUILD_ID)
}

func TestEnvType_Values(t *testing.T) {
	assert.Equal(t, EnvType("development"), Development)
	assert.Equal(t, EnvType("production"), Production)
}
