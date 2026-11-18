package main

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Veraticus/findingway/internal/tokenizer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var reportedAt = time.Date(2026, 9, 25, 12, 30, 0, 0, time.UTC)

func TestTokensMessageStartsWithHeader(t *testing.T) {
	report := &tokenizer.Report{
		Lookback:     5,
		Tokens:       []tokenizer.Token{{String: "prog", Count: 3}},
		ListingCount: 42,
	}

	msg := tokensMessage(report, reportedAt)

	assert.Equal(t, "**Last 5 days of Tokens (42 listings scanned) as of 2026-09-25 12:30:00\n**`prog` 3\n", msg.Content)
}

func TestTokensMessageListsAtMost75Tokens(t *testing.T) {
	report := &tokenizer.Report{Lookback: 2, Tokens: make([]tokenizer.Token, 80), ListingCount: 7}
	for i := range report.Tokens {
		report.Tokens[i] = tokenizer.Token{String: "tok", Count: i}
	}

	msg := tokensMessage(report, reportedAt)

	assert.True(t, strings.HasPrefix(msg.Content, "**Last 2 days of Tokens (7 listings scanned) as of 2026-09-25 12:30:00\n**"))
	assert.Equal(t, 75, strings.Count(msg.Content, "`tok`"))
}

func TestDescriptionCsvMessageAttachesCsv(t *testing.T) {
	msg := descriptionCsvMessage([]byte("Timestamp,Description\n"), reportedAt)

	assert.Equal(t, "**Description List 2026-09-25 12:30:00**", msg.Content)
	require.NotNil(t, msg.File)
	assert.Equal(t, "pf_descriptions_2026-09-25 12:30:00.csv", msg.File.Name)
	body, err := io.ReadAll(msg.File.Reader)
	require.NoError(t, err)
	assert.Equal(t, "Timestamp,Description\n", string(body))
}
