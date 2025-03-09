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
		{
			name: "Array Properties",
			schema: map[string]interface{}{
				"title": "Product",
				"type":  "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Product ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Product name",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Product tags",
						"items": map[string]interface{}{
							"type": "string",
						},
						"minItems":    1,
						"uniqueItems": true,
					},
					"variants": map[string]interface{}{
						"type":        "array",
						"description": "Product variants",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id": map[string]interface{}{
									"type": "string",
								},
								"size": map[string]interface{}{
									"type": "string",
								},
								"color": map[string]interface{}{
									"type": "string",
								},
								"price": map[string]interface{}{
									"type":    "number",
									"minimum": 0,
								},
							},
							"required": []interface{}{"id", "price"},
						},
					},
					"sizes": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []interface{}{"S", "M", "L", "XL"},
						},
					},
					"prices": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":    "number",
							"minimum": 0,
							"maximum": 1000,
						},
					},
					"images": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":   "string",
							"format": "uri",
						},
						"minItems": 1,
						"maxItems": 10,
					},
				},
				"required": []interface{}{"id", "name"},
			},
			expected: "variants?: [TestArrayPropertiesVariantsItem]",
		},
		{
			name: "String Constraint Schema",
			schema: map[string]interface{}{
				"type":  "object",
				"title": "StringConstraints",
				"properties": map[string]interface{}{
					"username": map[string]interface{}{
						"type":        "string",
						"minLength":   3,
						"maxLength":   50,
						"description": "Username with length constraints",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"format":      "email",
						"description": "Email address",
					},
					"website": map[string]interface{}{
						"type":        "string",
						"format":      "uri",
						"description": "Website URL",
					},
					"pattern_field": map[string]interface{}{
						"type":        "string",
						"pattern":     "^[A-Z][a-z]+$",
						"description": "Field with regex pattern",
					},
					"uuid_field": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "UUID field",
					},
					"date_field": map[string]interface{}{
						"type":        "string",
						"format":      "date",
						"description": "Date field (YYYY-MM-DD)",
					},
					"datetime_field": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "DateTime field",
					},
					"combined_constraints": map[string]interface{}{
						"type":        "string",
						"minLength":   8,
						"maxLength":   100,
						"pattern":     "^[a-zA-Z0-9]+$",
						"description": "Field with multiple constraints",
					},
					"tags": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":      "string",
							"minLength": 2,
							"maxLength": 20,
						},
						"minItems":    1,
						"uniqueItems": true,
						"description": "Tags with string constraints",
					},
					"emails": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":   "string",
							"format": "email",
						},
						"description": "List of email addresses",
					},
					"patterns": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":    "string",
							"pattern": "^[A-Z]{2}\\d{4}$",
						},
						"description": "List of pattern-constrained strings",
					},
				},
				"required": []interface{}{"username", "email"},
			},
			expected: "schema StringConstraints:",
		},
	}

	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "schema-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new schema generator
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

			// For array properties test case, check the specific array validations
			if tc.name == "Array Properties" {
				// Check simple array type
				assert.Contains(t, schemaContent, "tags?: [str]", "Simple array should be typed properly")

				// Check complex array with object items (should have nested schema)
				assert.Contains(t, schemaContent, "variants?: [TestArrayPropertiesVariantsItem]",
					"Array of objects should reference nested schema")

				// Verify nested item schema for array of objects
				assert.Contains(t, schemaContent, "schema TestArrayPropertiesVariantsItem:",
					"Should generate schema for array item objects")

				// Check array constraints
				assert.Contains(t, schemaContent, "check tags == None or len(tags) >= 1",
					"Should validate minItems constraint")
				assert.Contains(t, schemaContent, "check tags == None or len(tags) == len(set(tags))",
					"Should validate uniqueItems constraint")

				// Check array of strings with enum
				assert.Contains(t, schemaContent, "sizes?: [str]", "Array of enum strings should be typed properly")
				assert.Contains(t, schemaContent, "all item in sizes {",
					"Should validate items against enum values")
				assert.Contains(t, schemaContent, `item in ["S", "M", "L", "XL"]`,
					"Should check each item against the enum values")

				// Check array of numbers with min/max
				assert.Contains(t, schemaContent, "prices?: [float]",
					"Array of numbers should be typed properly")
				assert.Contains(t, schemaContent, "all item in prices {",
					"Should validate numeric constraints for items")
				assert.Contains(t, schemaContent, "item >= 0",
					"Should check minimum value constraint for numbers")
				assert.Contains(t, schemaContent, "item <= 1000",
					"Should check maximum value constraint for numbers")

				// Check array size constraints
				assert.Contains(t, schemaContent, "check images == None or len(images) >= 1",
					"Should validate minItems constraint")
				assert.Contains(t, schemaContent, "check images == None or len(images) <= 10",
					"Should validate maxItems constraint")
			}

			// For string constraints test case, check the specific string constraints
			if tc.name == "String Constraint Schema" {
				// Verify the schema name
				assert.Contains(t, schemaContent, "schema StringConstraints:",
					"Should generate schema for string constraints")

				// Verify required properties
				assert.Contains(t, schemaContent, "username: str", "Required property should not have ? modifier")
				assert.Contains(t, schemaContent, "email: str", "Required property should not have ? modifier")

				// Verify optional properties
				assert.Contains(t, schemaContent, "website?: str", "Optional property should have ? modifier")
				assert.Contains(t, schemaContent, "pattern_field?: str", "Optional property should have ? modifier")
				assert.Contains(t, schemaContent, "uuid_field?: str", "Optional property should have ? modifier")
				assert.Contains(t, schemaContent, "date_field?: str", "Optional property should have ? modifier")
				assert.Contains(t, schemaContent, "datetime_field?: str", "Optional property should have ? modifier")
				assert.Contains(t, schemaContent, "combined_constraints?: str", "Optional property should have ? modifier")
				assert.Contains(t, schemaContent, "tags?: [str]", "Array of string constraints should be typed properly")
				assert.Contains(t, schemaContent, "emails?: [str]", "Array of email addresses should be typed properly")
				assert.Contains(t, schemaContent, "patterns?: [str]", "Array of pattern-constrained strings should be typed properly")

				// Verify length constraints
				assert.Contains(t, schemaContent, "check username == None or len(username) >= 3",
					"Should validate minLength constraint for username")
				assert.Contains(t, schemaContent, "check username == None or len(username) <= 50",
					"Should validate maxLength constraint for username")
				assert.Contains(t, schemaContent, "check combined_constraints == None or len(combined_constraints) >= 8",
					"Should validate minLength constraint for combined_constraints")
				assert.Contains(t, schemaContent, "check combined_constraints == None or len(combined_constraints) <= 100",
					"Should validate maxLength constraint for combined_constraints")

				// Verify array constraints
				assert.Contains(t, schemaContent, "check tags == None or len(tags) >= 1",
					"Should validate minItems constraint for tags")
				assert.Contains(t, schemaContent, "check tags == None or len(tags) == len(set(tags))",
					"Should validate uniqueItems constraint for tags")
				assert.Contains(t, schemaContent, "check tags == None or all item in tags { len(item) >= 2 }",
					"Should validate minLength constraint for each tag")
				assert.Contains(t, schemaContent, "check tags == None or all item in tags { len(item) <= 20 }",
					"Should validate maxLength constraint for each tag")

				// Verify pattern constraints (as comments or imports)
				assert.Contains(t, schemaContent, "# Regex pattern: ^[A-Z][a-z]+$",
					"Should include regex pattern for pattern_field")
				assert.Contains(t, schemaContent, "# Each item should match pattern: ^[A-Z]{2}\\d{4}$",
					"Should include regex pattern for patterns")

				// Verify format constraints
				assert.Contains(t, schemaContent, "# Email validation for email",
					"Should include email format validation for email")
				assert.Contains(t, schemaContent, "# URI validation for website",
					"Should include URI format validation for website")
				assert.Contains(t, schemaContent, "# UUID validation for uuid_field",
					"Should include UUID format validation for uuid_field")
				assert.Contains(t, schemaContent, "# Date validation for date_field",
					"Should include date format validation for date_field")
				assert.Contains(t, schemaContent, "# Date-time validation for datetime_field",
					"Should include date-time format validation for datetime_field")
				assert.Contains(t, schemaContent, "# Each item should be a valid email format",
					"Should include email format validation for emails")

				// Check for check blocks
				assert.Contains(t, schemaContent, "check:", "Schema should include validation checks")
				assert.Contains(t, schemaContent, "username != None", "Schema should verify required property is not None")
				assert.Contains(t, schemaContent, "email != None", "Schema should verify required property is not None")
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
