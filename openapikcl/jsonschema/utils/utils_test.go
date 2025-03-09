package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStringValue(t *testing.T) {
	data := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
	}

	value, ok := GetStringValue(data, "string")
	assert.True(t, ok)
	assert.Equal(t, "value", value)

	value, ok = GetStringValue(data, "missing")
	assert.False(t, ok)
	assert.Equal(t, "", value)

	value, ok = GetStringValue(data, "number")
	assert.False(t, ok)
	assert.Equal(t, "", value)
}

func TestFormatLiteral(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "String literal",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "True boolean literal",
			input:    true,
			expected: "True",
		},
		{
			name:     "False boolean literal",
			input:    false,
			expected: "False",
		},
		{
			name:     "Nil literal",
			input:    nil,
			expected: "None",
		},
		{
			name:     "Integer literal",
			input:    42,
			expected: "42",
		},
		{
			name:     "Float literal",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "Array literal",
			input:    []interface{}{"a", 1, true},
			expected: `["a", 1, True]`,
		},
		{
			name:     "Map literal",
			input:    map[string]interface{}{"key": "value", "num": 42},
			expected: `{"key": "value", "num": 42}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatLiteral(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSanitizePropertyName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty name",
			input:    "",
			expected: "property",
		},
		{
			name:     "Simple name",
			input:    "name",
			expected: "name",
		},
		{
			name:     "Name with spaces",
			input:    "first name",
			expected: "first_name",
		},
		{
			name:     "Name with special characters",
			input:    "user-email@domain.com",
			expected: "user_email_domain_com",
		},
		{
			name:     "Name starting with number",
			input:    "123property",
			expected: "_123property",
		},
		{
			name:     "Reserved keyword",
			input:    "import",
			expected: "import_",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizePropertyName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
