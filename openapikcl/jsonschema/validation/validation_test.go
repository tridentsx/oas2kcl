package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateConstraints(t *testing.T) {
	testCases := []struct {
		name       string
		propSchema map[string]interface{}
		propName   string
		contains   []string
	}{
		{
			name: "String with min and max length",
			propSchema: map[string]interface{}{
				"type":      "string",
				"minLength": 3,
				"maxLength": 20,
			},
			propName: "username",
			contains: []string{
				"check username == None or len(username) >= 3",
				"check username == None or len(username) <= 20",
			},
		},
		{
			name: "String with pattern",
			propSchema: map[string]interface{}{
				"type":    "string",
				"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			},
			propName: "email",
			contains: []string{
				"# Regex pattern:",
				"check email == None or regex.match",
			},
		},
		{
			name: "String with email format",
			propSchema: map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
			propName: "email",
			contains: []string{
				"# Email validation for email",
				"check email == None or regex.match",
			},
		},
		{
			name: "String with date-time format",
			propSchema: map[string]interface{}{
				"type":   "string",
				"format": "date-time",
			},
			propName: "timestamp",
			contains: []string{
				"# Date-time validation for timestamp",
				"check timestamp == None or regex.match",
			},
		},
		{
			name: "String with URI format",
			propSchema: map[string]interface{}{
				"type":   "string",
				"format": "uri",
			},
			propName: "website",
			contains: []string{
				"# URI validation for website",
				"check website == None or regex.match",
			},
		},
		{
			name: "String with UUID format",
			propSchema: map[string]interface{}{
				"type":   "string",
				"format": "uuid",
			},
			propName: "id",
			contains: []string{
				"# UUID validation for id",
				"check id == None or regex.match",
			},
		},
		{
			name: "String with all constraints",
			propSchema: map[string]interface{}{
				"type":      "string",
				"minLength": 5,
				"maxLength": 50,
				"pattern":   "^[a-z0-9]+$",
				"format":    "hostname",
			},
			propName: "domain",
			contains: []string{
				"check domain == None or len(domain) >= 5",
				"check domain == None or len(domain) <= 50",
				"# Regex pattern: ^[a-z0-9]+$",
				"# Hostname validation for domain",
			},
		},
		{
			name: "String with enum",
			propSchema: map[string]interface{}{
				"type": "string",
				"enum": []interface{}{"pending", "active", "suspended"},
			},
			propName: "status",
			contains: []string{
				"check status == None or status in",
			},
		},
		{
			name: "Number with range constraints",
			propSchema: map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
				"maximum": 120,
			},
			propName: "age",
			contains: []string{
				"check age == None or age >= 0",
				"check age == None or age <= 120",
			},
		},
		{
			name: "Array of strings with constraints",
			propSchema: map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string", "minLength": 2, "maxLength": 10},
				"minItems":    1,
				"maxItems":    5,
				"uniqueItems": true,
			},
			propName: "tags",
			contains: []string{
				"check tags == None or len(tags) >= 1",
				"check tags == None or len(tags) <= 5",
				"check tags == None or len(tags) == len(set(tags))",
				"check tags == None or all item in tags { len(item) >= 2 }",
				"check tags == None or all item in tags { len(item) <= 10 }",
			},
		},
		{
			name: "Array of strings with pattern",
			propSchema: map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string", "pattern": "^\\d{3}-\\d{2}-\\d{4}$"},
			},
			propName: "ssns",
			contains: []string{
				"# Each item should match pattern: ^\\d{3}-\\d{2}-\\d{4}$",
				"check ssns == None or all item in ssns { regex.match",
			},
		},
		{
			name: "Array of strings with email format",
			propSchema: map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string", "format": "email"},
			},
			propName: "contacts",
			contains: []string{
				"# Each item should be a valid email format",
				"check contacts == None or all item in contacts { regex.match",
			},
		},
		{
			name: "Array of strings with enum",
			propSchema: map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string", "enum": []interface{}{"admin", "user", "guest"}},
			},
			propName: "roles",
			contains: []string{
				"check roles == None or all item in roles { item in [\"admin\", \"user\", \"guest\"] }",
			},
		},
		{
			name: "Array of numbers with range constraints",
			propSchema: map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "number", "minimum": 0, "maximum": 100},
			},
			propName: "scores",
			contains: []string{
				"check scores == None or all item in scores {",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateConstraints(tc.propSchema, tc.propName)
			for _, expectedStr := range tc.contains {
				assert.Contains(t, result, expectedStr)
			}
		})
	}
}

func TestCheckIfNeedsRegexImport(t *testing.T) {
	testCases := []struct {
		name          string
		schema        map[string]interface{}
		expectedRegex bool
	}{
		{
			name: "No regex needed",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"age": map[string]interface{}{
						"type": "integer",
					},
				},
			},
			expectedRegex: false,
		},
		{
			name: "Regex needed for pattern",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"email": map[string]interface{}{
						"type":    "string",
						"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					},
				},
			},
			expectedRegex: true,
		},
		{
			name: "Regex needed for email format",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"email": map[string]interface{}{
						"type":   "string",
						"format": "email",
					},
				},
			},
			expectedRegex: true,
		},
		{
			name: "Regex needed for date-time format",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"created": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			expectedRegex: true,
		},
		{
			name: "Regex needed for URI format",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"website": map[string]interface{}{
						"type":   "string",
						"format": "uri",
					},
				},
			},
			expectedRegex: true,
		},
		{
			name: "Regex needed for array item pattern",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"phone_numbers": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":    "string",
							"pattern": "^\\d{3}-\\d{3}-\\d{4}$",
						},
					},
				},
			},
			expectedRegex: true,
		},
		{
			name: "Regex needed for array item format",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"emails": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":   "string",
							"format": "email",
						},
					},
				},
			},
			expectedRegex: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckIfNeedsRegexImport(tc.schema)
			assert.Equal(t, tc.expectedRegex, result)
		})
	}
}

func TestGenerateRequiredPropertyChecks(t *testing.T) {
	// Schema with required properties
	schemaWithRequired := map[string]interface{}{
		"properties": map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type": "integer",
			},
		},
		"required": []interface{}{"username", "email"},
	}

	// Schema without required properties
	schemaWithoutRequired := map[string]interface{}{
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}

	// Schema with empty required array
	schemaWithEmptyRequired := map[string]interface{}{
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []interface{}{},
	}

	result := GenerateRequiredPropertyChecks(schemaWithRequired)
	assert.Contains(t, result, "check:")
	assert.Contains(t, result, "username != None")
	assert.Contains(t, result, "email != None")

	// Should not contain checks for non-required properties
	assert.NotContains(t, result, "age != None")

	// Empty result for schemas without required properties
	assert.Empty(t, GenerateRequiredPropertyChecks(schemaWithoutRequired))
	assert.Empty(t, GenerateRequiredPropertyChecks(schemaWithEmptyRequired))
}
