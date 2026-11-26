package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/Veraticus/findingway/internal/discord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestMainHelper(t *testing.T) {
	if os.Getenv("FINDINGWAY_RUN_MAIN") != "1" {
		t.Skip("subprocess helper")
	}
	main()
}

func TestMainFailsWithoutRedisCredentials(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainHelper$")
	cmd.Dir = t.TempDir()
	cmd.Env = []string{"FINDINGWAY_RUN_MAIN=1", "PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}

	out, err := cmd.CombinedOutput()

	assert.Error(t, err)
	assert.Contains(t, string(out), "could not configure the tokenizer: REDIS_USER is not set")
}

func TestCommittedConfigParsesIntoChannels(t *testing.T) {
	raw, err := os.ReadFile("config.yaml")
	require.NoError(t, err)
	d := &discord.Discord{}

	require.NoError(t, yaml.Unmarshal(raw, &d))

	require.NotEmpty(t, d.Channels)
	for _, c := range d.Channels {
		assert.NotEmpty(t, c.ID, c.Name)
		assert.NotEmpty(t, c.Duty, c.Name)
		assert.NotEmpty(t, c.DataCentres, c.Name)
	}
}
