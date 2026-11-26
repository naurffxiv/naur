package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMainHelper runs main() when invoked as a subprocess by the tests below.
func TestMainHelper(t *testing.T) {
	if os.Getenv("CLEARINGWAY_RUN_MAIN") != "1" {
		t.Skip("subprocess helper")
	}
	main()
}

func runMain(t *testing.T, env ...string) string {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainHelper$")
	cmd.Env = append([]string{"CLEARINGWAY_RUN_MAIN=1", "PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}, env...)
	out, err := cmd.CombinedOutput()
	assert.Error(t, err, "main should exit with an error")
	return string(out)
}

func TestMainRequiresCredentials(t *testing.T) {
	tests := map[string][]string{
		"You must supply a DISCORD_TOKEN to start!":        nil,
		"You must supply a FFLOGS_CLIENT_ID to start!":     {"DISCORD_TOKEN=t"},
		"You must supply a FFLOGS_CLIENT_SECRET to start!": {"DISCORD_TOKEN=t", "FFLOGS_CLIENT_ID=id"},
	}
	for want, env := range tests {
		t.Run(want, func(t *testing.T) {
			assert.Contains(t, runMain(t, env...), want)
		})
	}
}
