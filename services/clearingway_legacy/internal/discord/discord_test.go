package discord

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recorder struct {
	paths  []string
	bodies []map[string]any
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	var body map[string]any
	raw, _ := io.ReadAll(req.Body)
	_ = json.Unmarshal(raw, &body)
	r.paths = append(r.paths, req.Method+" "+strings.TrimPrefix(req.URL.Path, "/api/v9"))
	r.bodies = append(r.bodies, body)
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{}`)), Request: req}, nil
}

func session(t *testing.T, rt http.RoundTripper) *discordgo.Session {
	t.Helper()
	s, err := discordgo.New("Bot test")
	require.NoError(t, err)
	s.Client = &http.Client{Transport: rt}
	return s
}

var interaction = &discordgo.Interaction{ID: "i1", AppID: "app", Token: "tok"}

func TestStartInteractionRepliesEphemerally(t *testing.T) {
	rec := &recorder{}

	require.NoError(t, StartInteraction(session(t, rec), interaction, "hello"))

	assert.Equal(t, []string{"POST /interactions/i1/tok/callback"}, rec.paths)
	data := rec.bodies[0]["data"].(map[string]any)
	assert.Equal(t, "hello", data["content"])
	assert.Equal(t, float64(discordgo.MessageFlagsEphemeral), data["flags"])
}

func TestContinueInteractionSendsEphemeralFollowUp(t *testing.T) {
	rec := &recorder{}

	require.NoError(t, ContinueInteraction(session(t, rec), interaction, "more"))

	assert.Equal(t, []string{"POST /webhooks/app/tok"}, rec.paths)
	assert.Equal(t, "more", rec.bodies[0]["content"])
	assert.Equal(t, float64(discordgo.MessageFlagsEphemeral), rec.bodies[0]["flags"])
}

func TestStartCreatesSessionWithGuildIntents(t *testing.T) {
	d := &Discord{Token: "abc"}

	require.NoError(t, d.Start())

	assert.Equal(t, "Bot abc", d.Session.Token)
	assert.Equal(t, discordgo.IntentsGuilds|discordgo.IntentsGuildMessages, d.Session.Identify.Intents)
}
