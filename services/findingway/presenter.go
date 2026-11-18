package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/tokenizer"

	"github.com/bwmarrin/discordgo"
)

// maxReportedTokens is how many of the most frequent tokens are posted.
const maxReportedTokens = 75

// tokensMessage lists the report's most frequent tokens over its lookback
// days.
func tokensMessage(report *tokenizer.Report, at time.Time) *discordgo.MessageSend {
	var content strings.Builder
	fmt.Fprintf(&content, "**Last %d days of Tokens (%d listings scanned) as of %s\n**", report.Lookback, report.ListingCount, at.Format(time.DateTime))
	for _, token := range report.Tokens[:min(maxReportedTokens, len(report.Tokens))] {
		fmt.Fprintf(&content, "`%s` %d\n", token.String, token.Count)
	}

	return &discordgo.MessageSend{
		Content: content.String(),
	}
}

// descriptionCsvMessage attaches the report's description CSV.
func descriptionCsvMessage(csv []byte, at time.Time) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Content: fmt.Sprintf("**Description List %s**", at.Format(time.DateTime)),
		File: &discordgo.File{
			Name:        fmt.Sprintf("pf_descriptions_%s.csv", at.Format(time.DateTime)),
			ContentType: "csv",
			Reader:      bytes.NewReader(csv),
		},
	}
}
