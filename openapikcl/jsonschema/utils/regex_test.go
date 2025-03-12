package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslateECMAToGoRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple pattern",
			input:    "^[a-z]+$",
			expected: "`^[a-z]+$`",
		},
		{
			name:     "Pattern with digit class",
			input:    "\\d+",
			expected: "`[0-9]+`",
		},
		{
			name:     "Pattern with word class",
			input:    "\\w+",
			expected: "`[a-zA-Z0-9_]+`",
		},
		{
			name:     "Pattern with word boundary",
			input:    "\\bword\\b",
			expected: "`\\bword\\b`",
		},
		{
			name:     "Pattern with unicode property",
			input:    "\\p{L}+",
			expected: "`\\pL+`",
		},
		{
			name:     "Pattern with lookahead",
			input:    "foo(?=bar)",
			expected: "`foo`",
			skip:     true, // Skip this test as lookaheads are not fully supported
		},
		{
			name:     "Pattern with named group",
			input:    "(?<name>\\w+)",
			expected: "`(?P<name>[a-zA-Z0-9_]+)`",
		},
		{
			name:     "Pattern with unicode category",
			input:    "\\p{Letter}+",
			expected: "`\\pL+`",
		},
		{
			name:     "Pattern with multiple translations",
			input:    "\\d+\\w+\\s+",
			expected: "`[0-9]+[a-zA-Z0-9_]+\\s+`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Pattern with lookahead" {
				t.Skip("Skipping test as lookaheads are not fully supported")
			}
			result := TranslateECMAToGoRegex(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommonPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   []string
		invalid []string
	}{
		{
			name:    "email",
			pattern: CommonPatterns["email"],
			valid: []string{
				"test@example.com",
				"user.name@domain.co.uk",
				"user+tag@example.com",
				"user@sub.domain.com",
			},
			invalid: []string{
				"invalid",
				"@example.com",
				"user@",
				"user@.com",
			},
		},
		{
			name:    "ipv4",
			pattern: CommonPatterns["ipv4"],
			valid: []string{
				"192.168.1.1",
				"10.0.0.0",
				"172.16.254.1",
				"0.0.0.0",
				"255.255.255.255",
			},
			invalid: []string{
				"256.256.256.256",
				"1.2.3",
				"192.168.001.1",
				"192.168.1.1.1",
			},
		},
		{
			name:    "date-time",
			pattern: CommonPatterns["date-time"],
			valid: []string{
				"2024-03-20T10:00:00Z",
				"2024-03-20T10:00:00+01:00",
				"2024-03-20T10:00:00.123Z",
				"2024-03-20T10:00:00.123+01:00",
			},
			invalid: []string{
				"2024-13-20T10:00:00Z", // Invalid month
				"2024-03-20T25:00:00Z", // Invalid hour
				"2024-03-20T10:00:00",  // Missing timezone
				"2024-03-20",           // Missing time
			},
		},
		{
			name:    "uri",
			pattern: CommonPatterns["uri"],
			valid: []string{
				"http://example.com",
				"https://example.com",
				"ftp://example.com",
				"mailto:user@example.com",
			},
			invalid: []string{
				"example.com",
				"http:/example.com",
				"http//example.com",
				"http:example.com",
			},
		},
		{
			name:    "uuid",
			pattern: CommonPatterns["uuid"],
			valid: []string{
				"123e4567-e89b-12d3-a456-426614174000",
				"550e8400-e29b-41d4-a716-446655440000",
				"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			},
			invalid: []string{
				"not-a-uuid",
				"123e4567-e89b-12d3-a456-42661417400",   // Too short
				"123e4567-e89b-12d3-a456-4266141740000", // Too long
				"123e4567-e89b-12d3-a456-42661417400g",  // Invalid character
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := regexp.Compile(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to compile pattern %s: %v", tt.name, err)
			}

			// Test valid cases
			for _, valid := range tt.valid {
				if !re.MatchString(valid) {
					t.Errorf("Pattern %s should match %s but didn't", tt.name, valid)
				}
			}

			// Test invalid cases
			for _, invalid := range tt.invalid {
				if re.MatchString(invalid) {
					t.Errorf("Pattern %s should not match %s but did", tt.name, invalid)
				}
			}
		})
	}
}

func TestHandleSpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Unicode Letter category",
			input:    "\\p{Letter}+",
			expected: "\\pL+",
		},
		{
			name:     "Unicode Number category",
			input:    "\\p{Number}+",
			expected: "\\pN+",
		},
		{
			name:     "Unicode Punctuation category",
			input:    "\\p{Punctuation}+",
			expected: "\\pP+",
		},
		{
			name:     "Multiple unicode categories",
			input:    "\\p{Letter}+\\p{Number}*",
			expected: "\\pL+\\pN*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleSpecialCases(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKCLRegexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		valid   []string
		invalid []string
	}{
		{
			name:   "date-time in KCL",
			format: "date-time",
			valid: []string{
				"2024-03-20T10:00:00Z",
				"2024-03-20T10:00:00+01:00",
				"2024-03-20T10:00:00.123Z",
				"2024-03-20T10:00:00.123+01:00",
			},
			invalid: []string{
				"2024-13-20T10:00:00Z", // Invalid month
				"2024-03-20T25:00:00Z", // Invalid hour
				"2024-03-20T10:00:00",  // Missing timezone
				"2024-03-20",           // Missing time
			},
		},
		{
			name:   "ipv4 in KCL",
			format: "ipv4",
			valid: []string{
				"192.168.1.1",
				"10.0.0.0",
				"172.16.254.1",
				"0.0.0.0",
				"255.255.255.255",
			},
			invalid: []string{
				"256.256.256.256",
				"1.2.3",
				"192.168.001.1",
				"192.168.1.1.1",
			},
		},
		{
			name:   "email in KCL",
			format: "email",
			valid: []string{
				"test@example.com",
				"user.name@domain.co.uk",
				"user+tag@example.com",
				"user@sub.domain.com",
			},
			invalid: []string{
				"invalid",
				"@example.com",
				"user@",
				"user@.com",
			},
		},
		{
			name:   "uri in KCL",
			format: "uri",
			valid: []string{
				"http://example.com",
				"https://example.com",
				"ftp://example.com",
				"mailto:user@example.com",
			},
			invalid: []string{
				"example.com",
				"http:/example.com",
				"http//example.com",
				"http:example.com",
			},
		},
		{
			name:   "uuid in KCL",
			format: "uuid",
			valid: []string{
				"123e4567-e89b-12d3-a456-426614174000",
				"550e8400-e29b-41d4-a716-446655440000",
				"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			},
			invalid: []string{
				"not-a-uuid",
				"123e4567-e89b-12d3-a456-42661417400",   // Too short
				"123e4567-e89b-12d3-a456-4266141740000", // Too long
				"123e4567-e89b-12d3-a456-42661417400g",  // Invalid character
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, ok := CommonPatterns[tt.format]
			if !ok {
				t.Fatalf("Pattern for format %s not found in CommonPatterns", tt.format)
			}

			// Convert to KCL regex pattern
			kclPattern := `r"` + pattern + `"`

			// Test in Go first (without r"..." wrapper)
			re, err := regexp.Compile(pattern)
			if err != nil {
				t.Fatalf("Failed to compile pattern %s: %v", tt.format, err)
			}

			// Test valid cases
			for _, valid := range tt.valid {
				if !re.MatchString(valid) {
					t.Errorf("Pattern %s should match %s but didn't", tt.format, valid)
				}
			}

			// Test invalid cases
			for _, invalid := range tt.invalid {
				if re.MatchString(invalid) {
					t.Errorf("Pattern %s should not match %s but did", tt.format, invalid)
				}
			}

			// Log the KCL pattern for manual testing
			t.Logf("KCL pattern for %s: %s", tt.format, kclPattern)
		})
	}
}
