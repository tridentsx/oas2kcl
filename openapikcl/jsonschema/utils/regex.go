package utils

import (
	"strings"
)

// TranslateECMAToGoRegex converts an ECMA-262 regex pattern to Go regex syntax.
// This is a best-effort translation and may not cover all edge cases.
func TranslateECMAToGoRegex(pattern string) string {
	// If pattern is already using backticks, strip them
	pattern = strings.TrimPrefix(pattern, "`")
	pattern = strings.TrimSuffix(pattern, "`")

	// Common translations
	translations := map[string]string{
		// Lookaheads/lookbehinds (not supported in Go)
		"(?=":  "", // Positive lookahead
		"(?!":  "", // Negative lookahead
		"(?<=": "", // Positive lookbehind
		"(?<!": "", // Negative lookbehind

		// Named capture groups
		"(?<": "(?P<", // Convert ECMA named groups to Go named groups

		// Word boundaries
		"\\b": `\b`, // Word boundary
		"\\B": `\B`, // Non-word boundary

		// Character classes
		"\\d": `[0-9]`,        // Digits
		"\\w": `[a-zA-Z0-9_]`, // Word characters
		"\\s": `\s`,           // Whitespace (same in both)

		// Unicode properties (Go uses different syntax)
		"\\p{L}": `\pL`, // Letters
		"\\p{N}": `\pN`, // Numbers
		"\\p{P}": `\pP`, // Punctuation
		"\\p{S}": `\pS`, // Symbols
		"\\p{Z}": `\pZ`, // Separators
	}

	// Apply translations
	result := pattern
	for ecma, go_regex := range translations {
		result = strings.ReplaceAll(result, ecma, go_regex)
	}

	// Handle special cases
	result = handleSpecialCases(result)

	// Wrap in backticks to avoid having to escape backslashes
	return "`" + result + "`"
}

// handleSpecialCases handles regex patterns that need special treatment
func handleSpecialCases(pattern string) string {
	// Handle unicode categories/blocks
	// ECMA: \p{Letter} -> Go: \pL
	unicodePattern := strings.NewReplacer(
		"\\p{Letter}", "\\pL",
		"\\p{Number}", "\\pN",
		"\\p{Punctuation}", "\\pP",
		"\\p{Symbol}", "\\pS",
		"\\p{Mark}", "\\pM",
		"\\p{Separator}", "\\pZ",
	)
	pattern = unicodePattern.Replace(pattern)

	return pattern
}

// Common regex patterns translated to Go syntax
var CommonPatterns = map[string]string{
	"email":     `^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`,
	"hostname":  `^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`,
	"ipv4":      `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
	"date":      `^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])$`,
	"time":      `^(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-][01]\d:[0-5]\d)$`,
	"date-time": `^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-][01]\d:[0-5]\d)$`,
	"uuid":      `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	"uri":       `^[a-zA-Z][a-zA-Z0-9+.-]*:[^\s]*$`,
}
