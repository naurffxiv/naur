package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		// This token is taken from the Discord documentation examples and is not valid
		{"filled token", "MTk4NjIyNDgzNDcxOTI1MjQ4.Cl2FMQ.ZnCjm1XVW7vRze4b7Cq4se7kKWs"},
		// Empty token is permitted because we check for nil in env.go
		{"empty token", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discord, err := NewSession(tt.token)

			assert.NoError(t, err)
			assert.NotNil(t, discord)
			assert.NotNil(t, discord.Session)
		})
	}
}
