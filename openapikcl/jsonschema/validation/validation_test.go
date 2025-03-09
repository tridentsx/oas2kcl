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
		contains   string
	}{
		{
			name: "String with min and max length",
			propSchema: map[string]interface{}{
				"type":      "string",
				"minLength": 3,
				"maxLength": 20,
			},
			propName: "username",
			contains: "check username == None or len(username) >= 3",
		},
		{
			name: "String with pattern",
			propSchema: map[string]interface{}{
				"type":    "string",
				"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			},
			propName: "email",
			contains: "# Regex pattern:",
		},
		{
			name: "String with enum",
			propSchema: map[string]interface{}{
				"type": "string",
				"enum": []interface{}{"pending", "active", "suspended"},
			},
			propName: "status",
			contains: "check status == None or status in",
		},
		{
			name: "Number with range constraints",
			propSchema: map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
				"maximum": 120,
			},
			propName: "age",
			contains: "check age == None or age >= 0",
		},
		{
			name: "Array with constraints",
			propSchema: map[string]interface{}{
				"type":        "array",
				"minItems":    1,
				"maxItems":    10,
				"uniqueItems": true,
			},
			propName: "tags",
			contains: "check tags == None or len(tags) >= 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateConstraints(tc.propSchema, tc.propName)
			assert.Contains(t, result, tc.contains)
		})
	}
}

func TestCheckIfNeedsRegexImport(t *testing.T) {
	// Schema with regex pattern
	schemaWithRegex := map[string]interface{}{
		"properties": map[string]interface{}{
			"email": map[string]interface{}{
				"type":    "string",
				"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			},
		},
	}

	// Schema without regex pattern
	schemaWithoutRegex := map[string]interface{}{
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}

	assert.True(t, CheckIfNeedsRegexImport(schemaWithRegex))
	assert.False(t, CheckIfNeedsRegexImport(schemaWithoutRegex))
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
