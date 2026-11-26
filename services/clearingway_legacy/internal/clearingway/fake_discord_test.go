package clearingway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Veraticus/clearingway/internal/discord"
	"github.com/bwmarrin/discordgo"
)

// call is one REST request the bot made to Discord.
type call struct {
	Method string
	Path   string // relative to /api/v9, without query
	Body   map[string]any
}

// fakeDiscord answers Discord REST calls in memory and records them.
type fakeDiscord struct {
	mu          sync.Mutex
	calls       []call
	memberRoles []string
	guildRoles  []*discordgo.Role
}

func (f *fakeDiscord) RoundTrip(r *http.Request) (*http.Response, error) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v9")
	var body map[string]any
	if r.Body != nil {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
	}
	f.mu.Lock()
	f.calls = append(f.calls, call{Method: r.Method, Path: path, Body: body})
	f.mu.Unlock()

	var resp any = map[string]any{}
	switch {
	case r.Method == http.MethodGet && strings.Contains(path, "/members/"):
		resp = map[string]any{"user": map[string]any{"id": "u1"}, "roles": f.memberRoles}
	case r.Method == http.MethodGet && strings.HasSuffix(path, "/roles"):
		resp = f.guildRoles
	case r.Method == http.MethodPost && strings.HasSuffix(path, "/roles"):
		resp = map[string]any{"id": "new-" + body["name"].(string), "name": body["name"]}
	case r.Method == http.MethodPatch && strings.HasSuffix(path, "/roles"),
		r.Method == http.MethodPut && strings.Contains(path, "/commands"):
		resp = []any{}
	}
	raw, _ := json.Marshal(resp)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(raw)),
		Request:    r,
	}, nil
}

// session returns a discordgo session whose REST calls go to the fake.
func (f *fakeDiscord) session(t *testing.T) *discordgo.Session {
	t.Helper()
	s, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	s.Client = &http.Client{Transport: f}
	return s
}

func (f *fakeDiscord) Calls() []call {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]call(nil), f.calls...)
}

// replies returns the text of every interaction response and follow-up, in order.
func (f *fakeDiscord) replies() []string {
	out := []string{}
	for _, c := range f.Calls() {
		switch {
		case strings.HasSuffix(c.Path, "/callback"):
			if data, ok := c.Body["data"].(map[string]any); ok {
				if content, ok := data["content"].(string); ok {
					out = append(out, content)
				}
			}
		case strings.HasPrefix(c.Path, "/webhooks/"):
			out = append(out, c.Body["content"].(string))
		}
	}
	return out
}

// roleChanges returns "+id" for each role added to a member and "-id" for each removed.
func (f *fakeDiscord) roleChanges() []string {
	out := []string{}
	for _, c := range f.Calls() {
		if !strings.Contains(c.Path, "/members/") || !strings.Contains(c.Path, "/roles/") {
			continue
		}
		id := c.Path[strings.LastIndex(c.Path, "/")+1:]
		switch c.Method {
		case http.MethodPut:
			out = append(out, "+"+id)
		case http.MethodDelete:
			out = append(out, "-"+id)
		}
	}
	return out
}

// botWithGuild returns a Clearingway with one configured guild "g1" talking to fake.
func botWithGuild(t *testing.T, fake *fakeDiscord, config *ConfigGuild) (*Clearingway, *discordgo.Session) {
	t.Helper()
	config.GuildId, config.ChannelId = "g1", "c1"
	c := &Clearingway{Config: &Config{ConfigGuilds: []*ConfigGuild{config}}}
	c.Init()
	s := fake.session(t)
	c.Discord = &discord.Discord{Session: s}
	// connect every role to a Discord role whose ID is the role name
	for _, r := range c.Guilds.Guilds["g1"].AllRoles() {
		r.DiscordRole = &discordgo.Role{ID: r.Name, Name: r.Name}
	}
	return c, s
}

func command(name string, memberRoles []string, options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i1", AppID: "app", Token: "tok", Type: discordgo.InteractionApplicationCommand,
		GuildID: "g1", ChannelID: "c1",
		Member: &discordgo.Member{User: &discordgo.User{ID: "u1"}, Roles: memberRoles},
		Data:   discordgo.ApplicationCommandInteractionData{Name: name, Options: options},
	}}
}

func component(customID string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i1", AppID: "app", Token: "tok", Type: discordgo.InteractionMessageComponent,
		GuildID: "g1", ChannelID: "c1",
		Member: &discordgo.Member{User: &discordgo.User{ID: "u1"}},
		Data:   discordgo.MessageComponentInteractionData{CustomID: customID},
	}}
}

func stringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: name, Type: discordgo.ApplicationCommandOptionString, Value: value}
}

// callsTo returns the calls made with method to path.
func (f *fakeDiscord) callsTo(method, path string) []call {
	out := []call{}
	for _, c := range f.Calls() {
		if c.Method == method && strings.TrimSuffix(c.Path, "/") == path {
			out = append(out, c)
		}
	}
	return out
}

// autocompleteChoices returns the values of the single autocomplete response.
func (f *fakeDiscord) autocompleteChoices(t *testing.T) []string {
	t.Helper()
	calls := f.Calls()
	if len(calls) != 1 {
		t.Fatalf("autocomplete must answer Discord exactly once, got %d calls", len(calls))
	}
	choices := calls[0].Body["data"].(map[string]any)["choices"].([]any)
	names := []string{}
	for _, c := range choices {
		names = append(names, c.(map[string]any)["value"].(string))
	}
	return names
}
