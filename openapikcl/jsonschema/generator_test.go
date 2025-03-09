package jsonschema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaGenerator(t *testing.T) {
	testCases := []struct {
		name     string
		schema   map[string]interface{}
		expected string
	}{
		{
			name: "Simple object schema",
			schema: map[string]interface{}{
				"title": "Person",
				"type":  "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The person's name",
					},
					"age": map[string]interface{}{
						"type":        "integer",
						"description": "The person's age",
						"minimum":     0,
					},
				},
				"required": []interface{}{"name"},
			},
			expected: "schema TestSimpleObjectSchema:",
		},
		{
			name: "Empty schema",
			schema: map[string]interface{}{
				"title": "Empty",
				"type":  "object",
			},
			expected: "# This schema has no properties defined",
		},
		{
			name: "Required vs Optional Properties",
			schema: map[string]interface{}{
				"title": "User",
				"type":  "object",
				"properties": map[string]interface{}{
					"username": map[string]interface{}{
						"type":        "string",
						"description": "The user's username",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "The user's email address",
						"format":      "email",
					},
					"firstName": map[string]interface{}{
						"type":        "string",
						"description": "The user's first name",
					},
					"lastName": map[string]interface{}{
						"type":        "string",
						"description": "The user's last name",
					},
				},
				"required": []interface{}{"username", "email"},
			},
			expected: "username: str",
		},
		{
			name: "Nested Object Properties",
			schema: map[string]interface{}{
				"title": "Customer",
				"type":  "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The customer ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The customer name",
					},
					"address": map[string]interface{}{
						"type":        "object",
						"description": "The customer address",
						"properties": map[string]interface{}{
							"street": map[string]interface{}{
								"type":        "string",
								"description": "Street address",
							},
							"city": map[string]interface{}{
								"type":        "string",
								"description": "City",
							},
							"state": map[string]interface{}{
								"type":        "string",
								"description": "State or province",
							},
							"postalCode": map[string]interface{}{
								"type":        "string",
								"description": "Postal code",
							},
							"country": map[string]interface{}{
								"type":        "string",
								"description": "Country",
							},
						},
						"required": []interface{}{"street", "city", "country"},
					},
					"contact": map[string]interface{}{
						"type":  "object",
						"title": "ContactInfo",
						"properties": map[string]interface{}{
							"email": map[string]interface{}{
								"type":   "string",
								"format": "email",
							},
							"phone": map[string]interface{}{
								"type": "string",
							},
						},
						"required": []interface{}{"email"},
					},
				},
				"required": []interface{}{"id", "name"},
			},
			expected: "address?: TestNestedObjectPropertiesAddress",
		},
	}

	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "schema-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a generator for each test case
			generator := NewSchemaGenerator(tc.schema, tempDir)

			// Generate the schema
			schemaContent, err := generator.GenerateKCLSchema(tc.schema, "Test"+tc.name)
			require.NoError(t, err)

			// Check that the expected content is present
			assert.Contains(t, schemaContent, tc.expected)

			// For the required properties test case, check specific required/optional syntax
			if tc.name == "Required vs Optional Properties" {
				assert.Contains(t, schemaContent, "username: str", "Required property should not have ? modifier")
				assert.Contains(t, schemaContent, "email: str", "Required property should not have ? modifier")
				assert.Contains(t, schemaContent, "firstName?: str", "Optional property should have ? modifier")
				assert.Contains(t, schemaContent, "lastName?: str", "Optional property should have ? modifier")

				// Check for required property validation
				assert.Contains(t, schemaContent, "check:", "Schema should include validation checks")
				assert.Contains(t, schemaContent, "username != None", "Schema should verify required property is not None")
				assert.Contains(t, schemaContent, "email != None", "Schema should verify required property is not None")
			}

			// For the nested objects test case, check for nested schema definitions
			if tc.name == "Nested Object Properties" {
				// Check for the automatically named nested schema (address)
				assert.Contains(t, schemaContent, "schema TestNestedObjectPropertiesAddress:",
					"Should generate schema for nested address object")

				// Verify properties of the nested address schema
				assert.Contains(t, schemaContent, "street: str", "Required nested property should not have ? modifier")
				assert.Contains(t, schemaContent, "city: str", "Required nested property should not have ? modifier")
				assert.Contains(t, schemaContent, "country: str", "Required nested property should not have ? modifier")
				assert.Contains(t, schemaContent, "state?: str", "Optional nested property should have ? modifier")
				assert.Contains(t, schemaContent, "postalCode?: str", "Optional nested property should have ? modifier")

				// Check for the named nested schema (contact with title ContactInfo)
				assert.Contains(t, schemaContent, "schema ContactInfo:",
					"Should use title for nested contact object schema")

				// Verify properties of the nested contact schema
				assert.Contains(t, schemaContent, "email: str", "Required nested property should not have ? modifier")
				assert.Contains(t, schemaContent, "phone?: str", "Optional nested property should have ? modifier")

				// Verify required property validation for nested schemas
				assert.Contains(t, schemaContent, "street != None", "Should validate required properties in nested schema")
				assert.Contains(t, schemaContent, "city != None", "Should validate required properties in nested schema")
				assert.Contains(t, schemaContent, "country != None", "Should validate required properties in nested schema")
				assert.Contains(t, schemaContent, "email != None", "Should validate required properties in nested schema")
			}
		})
	}
}

func TestGenerateSchemas(t *testing.T) {
	// Simple schema for testing
	schemaJSON := `{
		"title": "TestPerson",
		"type": "object",
		"properties": {
			"name": {
				"type": "string"
			},
			"age": {
				"type": "integer"
			}
		}
	}`

	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "schemas-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Generate schemas
	err = GenerateSchemas([]byte(schemaJSON), tempDir, "test")
	require.NoError(t, err)

	// Check if the expected file was created
	schemaFile := filepath.Join(tempDir, "TestPerson.k")
	_, err = os.Stat(schemaFile)
	assert.NoError(t, err, "Schema file should exist")

	// Check the content of the file
	content, err := os.ReadFile(schemaFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "schema TestPerson:")
}
