package guardrails

import (
	"regexp"
	"strings"
)

// Match dispatches to the built-in matcher for ruleID.
// Returns (matched, reason). Unknown ruleIDs always return false.
func Match(ruleID, content string) (bool, string) {
	switch ruleID {
	case "prompt-injection":
		return matchPromptInjection(content)
	case "secret-detection":
		return matchSecretDetection(content)
	case "pii-leakage-output":
		return matchPIILeakage(content)
	case "toxicity-basic":
		return matchToxicityBasic(content)
	default:
		return false, ""
	}
}

// injectionPhrases are case-insensitive substrings that signal prompt injection.
var injectionPhrases = []string{
	"ignore previous instructions",
	"ignore all previous",
	"forget your instructions",
	"disregard your instructions",
	"override your instructions",
	"you are now",
	"act as",
	"jailbreak",
	"do anything now",
	"dan mode",
}

func matchPromptInjection(content string) (bool, string) {
	lower := strings.ToLower(content)
	for _, phrase := range injectionPhrases {
		if strings.Contains(lower, phrase) {
			return true, "prompt injection phrase: " + phrase
		}
	}
	return false, ""
}

// secretPatterns holds compiled regexes for common credential formats.
var secretPatterns = []struct {
	label string
	re    *regexp.Regexp
}{
	{"aws-access-key", regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{"github-token", regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9_]{36,251}\b`)},
	{"stripe-live-key", regexp.MustCompile(`\bsk_live_[0-9a-zA-Z]{24,}\b`)},
	// Generic bearer tokens (long opaque strings after "Bearer ")
	{"bearer-token", regexp.MustCompile(`Bearer\s+[A-Za-z0-9\-._~+/]{20,}`)},
}

func matchSecretDetection(content string) (bool, string) {
	for _, p := range secretPatterns {
		if p.re.MatchString(content) {
			return true, "credential pattern matched: " + p.label
		}
	}
	return false, ""
}

// piiPatterns covers SSN, credit card numbers, and email addresses.
var piiPatterns = []struct {
	label string
	re    *regexp.Regexp
}{
	{"ssn", regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)},
	// 13-16 digit card numbers with optional spaces or dashes between groups
	{"credit-card", regexp.MustCompile(`\b(?:\d[ -]?){13,16}\b`)},
	{"email", regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`)},
}

func matchPIILeakage(content string) (bool, string) {
	for _, p := range piiPatterns {
		if p.re.MatchString(content) {
			return true, "PII entity detected: " + p.label
		}
	}
	return false, ""
}

// toxicPhrases is a minimal keyword set for obviously harmful content.
var toxicPhrases = []string{
	"how to make a bomb",
	"how to synthesize drugs",
	"child sexual abuse",
	"csam",
	"how to kill",
}

func matchToxicityBasic(content string) (bool, string) {
	lower := strings.ToLower(content)
	for _, phrase := range toxicPhrases {
		if strings.Contains(lower, phrase) {
			return true, "toxic content: " + phrase
		}
	}
	return false, ""
}
