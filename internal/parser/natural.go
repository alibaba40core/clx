package parser

import "strings"

// nlPhrases is a bounded list of natural-language substrings.
var nlPhrases = []string{
	"find all",
	"show me",
	"list all",
	"modified today",
	"search for",
	"look for",
	"how do i",
	"can you",
}

var shellMetacharacters = []string{
	"|", ";", "&&", "||", ">", "<", "`", "$(", "${",
}

func isNaturalLanguage(raw string, tokens []string) bool {
	if len(tokens) == 0 {
		return false
	}
	if hasShellMetacharacters(raw) {
		return false
	}

	lower := strings.ToLower(raw)
	hasPhrase := false
	for _, p := range nlPhrases {
		if strings.Contains(lower, p) {
			hasPhrase = true
			break
		}
	}

	// Phrase match disambiguates shell verbs used in NL (e.g. "find all files ...").
	if hasPhrase && len(tokens) >= 2 {
		return true
	}

	if len(tokens) >= 3 && !isShellVerb(tokens[0]) {
		return true
	}
	return false
}

func hasShellMetacharacters(raw string) bool {
	for _, m := range shellMetacharacters {
		if strings.Contains(raw, m) {
			return true
		}
	}
	return false
}
