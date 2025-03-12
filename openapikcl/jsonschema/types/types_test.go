package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetKCLType(t *testing.T) {
	testCases := []struct {
		name     string
		schema   map[string]interface{}
		expected string
	}{
		{
			name:     "string type",
			schema:   map[string]interface{}{"type": "string"},
			expected: "str",
		},
		{
			name:     "integer type",
			schema:   map[string]interface{}{"type": "integer"},
			expected: "int",
		},
		{
			name:     "number type",
			schema:   map[string]interface{}{"type": "number"},
			expected: "float",
		},
		{
			name:     "boolean type",
			schema:   map[string]interface{}{"type": "boolean"},
			expected: "bool",
		},
		{
			name:     "null type",
			schema:   map[string]interface{}{"type": "null"},
			expected: "None",
		},
		{
			name:     "object type with title",
			schema:   map[string]interface{}{"type": "object", "title": "Person"},
			expected: "Person",
		},
		{
			name:     "object type without title",
			schema:   map[string]interface{}{"type": "object"},
			expected: "dict[str, any]",
		},
		{
			name:     "array of strings",
			schema:   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			expected: "[str]",
		},
		{
			name:     "multiple types",
			schema:   map[string]interface{}{"type": []interface{}{"string", "number", "null"}},
			expected: "any",
		},
		{
			name:     "no type specified",
			schema:   map[string]interface{}{},
			expected: "any",
		},
		{
			name:     "date-time format",
			schema:   map[string]interface{}{"type": "string", "format": "date-time"},
			expected: "str",
		},
		{
			name:     "ipv4 format",
			schema:   map[string]interface{}{"type": "string", "format": "ipv4"},
			expected: "str",
		},
		{
			name:     "uri format",
			schema:   map[string]interface{}{"type": "string", "format": "uri"},
			expected: "str",
		},
		{
			name:     "email format",
			schema:   map[string]interface{}{"type": "string", "format": "email"},
			expected: "str",
		},
		{
			name:     "uuid format",
			schema:   map[string]interface{}{"type": "string", "format": "uuid"},
			expected: "str",
		},
		{
			name:     "hostname format",
			schema:   map[string]interface{}{"type": "string", "format": "hostname"},
			expected: "str",
		},
		{
			name:     "ipv6 format",
			schema:   map[string]interface{}{"type": "string", "format": "ipv6"},
			expected: "str",
		},
		{
			name:     "json-pointer format",
			schema:   map[string]interface{}{"type": "string", "format": "json-pointer"},
			expected: "str",
		},
		{
			name:     "relative-json-pointer format",
			schema:   map[string]interface{}{"type": "string", "format": "relative-json-pointer"},
			expected: "str",
		},
		{
			name:     "iri format",
			schema:   map[string]interface{}{"type": "string", "format": "iri"},
			expected: "str",
		},
		{
			name:     "iri-reference format",
			schema:   map[string]interface{}{"type": "string", "format": "iri-reference"},
			expected: "str",
		},
		{
			name:     "uri-reference format",
			schema:   map[string]interface{}{"type": "string", "format": "uri-reference"},
			expected: "str",
		},
		{
			name:     "duration format",
			schema:   map[string]interface{}{"type": "string", "format": "duration"},
			expected: "str",
		},
		{
			name:     "date format",
			schema:   map[string]interface{}{"type": "string", "format": "date"},
			expected: "str",
		},
		{
			name:     "time format",
			schema:   map[string]interface{}{"type": "string", "format": "time"},
			expected: "str",
		},
		{
			name:     "regex format",
			schema:   map[string]interface{}{"type": "string", "format": "regex"},
			expected: "str",
		},
		{
			name:     "idn-email format",
			schema:   map[string]interface{}{"type": "string", "format": "idn-email"},
			expected: "str",
		},
		{
			name:     "idn-hostname format",
			schema:   map[string]interface{}{"type": "string", "format": "idn-hostname"},
			expected: "str",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetKCLType(tc.schema)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatSchemaName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty name",
			input:    "",
			expected: "Schema",
		},
		{
			name:     "Simple name",
			input:    "pet",
			expected: "Pet",
		},
		{
			name:     "Already capitalized",
			input:    "Pet",
			expected: "Pet",
		},
		{
			name:     "Name with spaces",
			input:    "pet store",
			expected: "PetStore",
		},
		{
			name:     "Name with special characters",
			input:    "pet-store_API@123",
			expected: "PetstoreAPI123",
		},
		{
			name:     "Single character",
			input:    "a",
			expected: "A",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatSchemaName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractSchemaName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON Schema ref",
			input:    "#/definitions/Pet",
			expected: "Pet",
		},
		{
			name:     "OpenAPI ref",
			input:    "#/components/schemas/Pet",
			expected: "Pet",
		},
		{
			name:     "File path",
			input:    "schemas/pet.json",
			expected: "Pet",
		},
		{
			name:     "URL path",
			input:    "https://example.com/schemas/pet.json",
			expected: "Pet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractSchemaName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
