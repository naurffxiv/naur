package tokenizer

import (
	"slices"
	"testing"
)

func TestSplitListingIntoTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // tokens that must appear
		excludes []string // tokens that must NOT appear
	}{
		{
			name:     "raidplan with https and fragment",
			input:    "Check https://raidplan.io/plan/lpsbjecjdb0xoloz#2 for strats",
			contains: []string{"lpsbjecjdb0xoloz"},
			excludes: []string{"https:", "raidplan.io", "plan", "lpsbjecjdb0xoloz#2"},
		},
		{
			name:     "bare raidplan link",
			input:    "raidplan.io/plan/lpsbjecjdb0xoloz prog",
			contains: []string{"lpsbjecjdb0xoloz", "prog"},
			excludes: []string{"raidplan.io", "plan"},
		},
		{
			name:     "raidplan code with hyphens",
			input:    "https://raidplan.io/plan/-LCqw-Otdw7AKprg",
			contains: []string{"-lcqw-otdw7akprg"},
			excludes: []string{"https:", "raidplan.io", "plan"},
		},
		{
			name:     "pastebin url",
			input:    "strats here https://pastebin.com/7fs57PyQ",
			contains: []string{"7fs57pyq"},
			excludes: []string{"https:", "pastebin.com"},
		},
		{
			name:     "tinyurl",
			input:    "https://tinyurl.com/kefkabin",
			contains: []string{"kefkabin"},
			excludes: []string{"https:", "tinyurl.com"},
		},
		{
			name:     "kefkab.in bare hostname",
			input:    "https://kefkab.in/",
			contains: []string{"kefkabin"},
			excludes: []string{"https:", "kefkab.in"},
		},
		{
			name:     "compounds with hyphens preserved",
			input:    "need d1-d4 prog",
			contains: []string{"d1-d4", "prog"},
		},
		{
			name:     "stop words removed",
			input:    "we are doing prog and graven",
			contains: []string{"prog", "graven"},
			excludes: []string{"we", "and", "are"},
		},
		{
			name:     "trailing punctuation trimmed",
			input:    "prog. enrage. graven,",
			contains: []string{"prog", "enrage", "graven"},
			excludes: []string{"prog.", "enrage.", "graven,"},
		},
		{
			name:     "parentheses trimmed",
			input:    "(prog) [fresh] graven",
			contains: []string{"prog", "fresh", "graven"},
			excludes: []string{"(prog)", "[fresh]"},
		},
		{
			name:     "single chars and short tokens dropped",
			input:    "e i a prog",
			contains: []string{"prog"},
			excludes: []string{"e", "i", "a"},
		},
		{
			name:     "bare numbers dropped",
			input:    "p2 graven 2 3 p1",
			contains: []string{"p2", "graven", "p1"},
			excludes: []string{"2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := splitListingIntoTokens(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tt.contains {
				if !slices.Contains(tokens, want) {
					t.Errorf("expected token %q in output %v", want, tokens)
				}
			}
			for _, banned := range tt.excludes {
				if slices.Contains(tokens, banned) {
					t.Errorf("unexpected token %q in output %v", banned, tokens)
				}
			}
		})
	}
}

func TestUrlToToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://pastebin.com/7fs57PyQ", "7fs57PyQ"},
		{"https://tinyurl.com/kefkabin", "kefkabin"},
		{"https://kefkab.in/", "kefkabin"},
		{"https://raidplan.io/plan/lpsbjecjdb0xoloz", "lpsbjecjdb0xoloz"},
		{"pastebin.com/7fs57PyQ", "7fs57PyQ"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := urlToToken(tt.input)
			if got != tt.want {
				t.Errorf("urlToToken(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
