package discord

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type request struct {
	Method      string
	Path        string
	ContentType string
	Raw         string
	JSON        map[string]any
}

// fakeDiscord records REST calls and returns canned channel messages.
type fakeDiscord struct {
	requests []request
	messages []*discordgo.Message
}

func (f *fakeDiscord) RoundTrip(r *http.Request) (*http.Response, error) {
	raw := []byte{}
	if r.Body != nil {
		raw, _ = io.ReadAll(r.Body)
	}
	req := request{Method: r.Method, Path: strings.TrimPrefix(r.URL.Path, "/api/v9"), ContentType: r.Header.Get("Content-Type"), Raw: string(raw)}
	_ = json.Unmarshal(raw, &req.JSON)
	f.requests = append(f.requests, req)

	var body any = map[string]any{}
	if r.Method == http.MethodGet && strings.HasSuffix(req.Path, "/messages") {
		body = f.messages
	}
	out, _ := json.Marshal(body)
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(out)), Request: r}, nil
}

func (f *fakeDiscord) bot(t *testing.T) *Discord {
	t.Helper()
	s, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	s.Client = &http.Client{Transport: f}
	s.State.User = &discordgo.User{ID: "bot"}
	return &Discord{Session: s}
}

// posts returns the JSON bodies of messages sent to a channel.
func (f *fakeDiscord) posts(channel string) []request {
	out := []request{}
	for _, r := range f.requests {
		if r.Method == http.MethodPost && r.Path == "/channels/"+channel+"/messages" {
			out = append(out, r)
		}
	}
	return out
}
