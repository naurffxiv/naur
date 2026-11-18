package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/tokenizer"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisClientReturnsErrorWithoutRedisCredentials(t *testing.T) {
	for _, name := range []string{"REDIS_USER", "REDIS_PASSWORD"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := newRedisClient(); err == nil {
		t.Fatal("expected an error when Redis credentials are missing")
	}
}

func TestNewRedisClientReturnsErrorWithoutPasswordOrCertificates(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("REDIS_USER", "findingway")
	t.Setenv("REDIS_PASSWORD", "")
	if err := os.Unsetenv("REDIS_PASSWORD"); err != nil {
		t.Fatal(err)
	}

	_, err := newRedisClient()
	assert.EqualError(t, err, "REDIS_PASSWORD is not set")

	t.Setenv("REDIS_PASSWORD", "secret")
	_, err = newRedisClient()
	assert.ErrorContains(t, err, "error loading certificates")
}

func TestConfigTracksTodaysDuties(t *testing.T) {
	cfg, err := readConfig("config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.TrackedDuties) != 1 || cfg.TrackedDuties[0] != "Dancing Mad (Ultimate)" {
		t.Errorf("TrackedDuties = %v, want [Dancing Mad (Ultimate)]", cfg.TrackedDuties)
	}
	if want := []string{"Aether", "Crystal", "Dynamis", "Primal"}; !slices.Equal(cfg.TrackedDataCentres, want) {
		t.Errorf("TrackedDataCentres = %v, want %v", cfg.TrackedDataCentres, want)
	}
	if len(cfg.Channels) == 0 {
		t.Error("expected channels in config.yaml")
	}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadConfigRejectsEmptyTrackedLists(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "no tracked duties",
			body: "trackedDataCentres: [Aether]\n",
			want: "trackedDuties is empty",
		},
		{
			name: "no tracked data centres",
			body: "trackedDuties: [\"Dancing Mad (Ultimate)\"]\ntrackedDataCentres: []\n",
			want: "trackedDataCentres is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readConfig(writeConfig(t, tt.body))
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestReadConfigReportsUnreadableAndInvalidFiles(t *testing.T) {
	_, err := readConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.ErrorContains(t, err, "could not read")

	_, err = readConfig(writeConfig(t, "channels: {not: a list"))
	assert.ErrorContains(t, err, "could not parse")
}

type fakeReporter struct {
	report   *tokenizer.Report
	err      error
	lookback int
	at       time.Time
}

func (f *fakeReporter) Report(_ context.Context, lookback int, now time.Time) (*tokenizer.Report, error) {
	f.lookback, f.at = lookback, now
	return f.report, f.err
}

type fakeMessenger struct {
	cleanErr error
	// sendErrs are returned by successive sends.
	sendErrs []error

	cleaned []string
	sent    []*discordgo.MessageSend
}

func (f *fakeMessenger) Clean(channelID string) error {
	f.cleaned = append(f.cleaned, channelID)
	return f.cleanErr
}

func (f *fakeMessenger) Send(channelID string, message *discordgo.MessageSend) error {
	f.sent = append(f.sent, message)
	if len(f.sent) <= len(f.sendErrs) {
		return f.sendErrs[len(f.sent)-1]
	}
	return nil
}

var testReport = &tokenizer.Report{
	Lookback:     lookback,
	Tokens:       []tokenizer.Token{{String: "prog", Count: 3}},
	ListingCount: 4,
	CSV:          []byte("Timestamp,Description\n"),
}

func TestPrintTokensWritesTheReport(t *testing.T) {
	r := &fakeReporter{report: testReport}
	var out strings.Builder

	require.NoError(t, printTokens(&out, r, reportedAt))

	assert.Equal(t, lookback, r.lookback)
	assert.Equal(t, reportedAt, r.at)
	assert.Equal(t, "4 listings scanned over last 2 days\n\nprog                           3\n", out.String())
}

func TestPrintTokensReturnsReportError(t *testing.T) {
	var out strings.Builder

	err := printTokens(&out, &fakeReporter{err: errors.New("redis down")}, reportedAt)

	assert.EqualError(t, err, "redis down")
	assert.Empty(t, out.String())
}

func TestPostTokensReplacesTheTokenChannelsMessages(t *testing.T) {
	m := &fakeMessenger{}

	require.NoError(t, postTokens(m, &fakeReporter{report: testReport}, reportedAt))

	assert.Equal(t, []string{tokenChannelID}, m.cleaned)
	require.Len(t, m.sent, 2)
	assert.Equal(t, tokensMessage(testReport, reportedAt).Content, m.sent[0].Content)
	assert.Equal(t, "pf_descriptions_2026-09-25 12:30:00.csv", m.sent[1].File.Name)
}

func TestPostTokensPostsEvenWhenCleaningFails(t *testing.T) {
	m := &fakeMessenger{cleanErr: errors.New("missing permission")}

	require.NoError(t, postTokens(m, &fakeReporter{report: testReport}, reportedAt))

	assert.Len(t, m.sent, 2)
}

func TestPostTokensFailures(t *testing.T) {
	t.Run("report error posts nothing", func(t *testing.T) {
		m := &fakeMessenger{}
		err := postTokens(m, &fakeReporter{err: errors.New("redis down")}, reportedAt)
		assert.EqualError(t, err, "redis down")
		assert.Empty(t, m.cleaned)
		assert.Empty(t, m.sent)
	})

	t.Run("tokens send error skips the csv", func(t *testing.T) {
		m := &fakeMessenger{sendErrs: []error{errors.New("rate limited")}}
		err := postTokens(m, &fakeReporter{report: testReport}, reportedAt)
		assert.EqualError(t, err, "rate limited")
		assert.Len(t, m.sent, 1)
	})

	t.Run("csv send error", func(t *testing.T) {
		m := &fakeMessenger{sendErrs: []error{nil, errors.New("too large")}}
		err := postTokens(m, &fakeReporter{report: testReport}, reportedAt)
		assert.EqualError(t, err, "could not post csv: too large")
	})
}
