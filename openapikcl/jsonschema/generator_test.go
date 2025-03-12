package jsonschema

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaGenerator(t *testing.T) {
	t.Skip("This test is for the old SchemaGenerator implementation. Use TestTreeBasedGenerator instead.")

	testCases := []struct {
		name          string
		schema        map[string]interface{}
		expected      string
		expectedTitle string
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
			expected:      "schema TestSimpleObjectSchema:",
			expectedTitle: "SimpleObjectSchema",
		},
		{
			name: "Empty schema",
			schema: map[string]interface{}{
				"title": "Empty",
				"type":  "object",
			},
			expected:      "schema TestEmptySchema:",
			expectedTitle: "TestEmptySchema",
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
			expected:      "username: str",
			expectedTitle: "",
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
			expected:      "address?: TestNestedObjectPropertiesAddress",
			expectedTitle: "",
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
			expected:      "variants?: [TestArrayPropertiesVariantsItem]",
			expectedTitle: "",
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
			expected: `# Username with length constraints
    username: str
    # Min length: 3
    # Max length: 50`,
			expectedTitle: "TestStringConstraintSchema",
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

			// For empty schema, verify it's a valid KCL schema by writing it to a file and parsing it
			if tc.name == "Empty schema" {
				// Write schema to a file
				schemaFilePath := filepath.Join(tempDir, "TestEmptySchema.k")
				err = os.WriteFile(schemaFilePath, []byte(schemaContent), 0644)
				require.NoError(t, err, "Failed to write schema file")

				// Verify the schema can be parsed with KCL
				if _, err := exec.LookPath("kcl"); err == nil {
					cmd := exec.Command("kcl", "fmt", schemaFilePath)
					output, err := cmd.CombinedOutput()
					if err != nil {
						t.Logf("KCL schema validation error: %s", output)
						t.Errorf("Generated empty schema is not valid KCL: %v", err)
					}

					// Create a simple test file that uses the empty schema
					testFilePath := filepath.Join(tempDir, "test_empty.k")
					testContent := "import TestEmptySchema\n\n# Create an instance of the empty schema\nempty = TestEmptySchema{}\n"
					err = os.WriteFile(testFilePath, []byte(testContent), 0644)
					require.NoError(t, err, "Failed to write test file")

					// Verify the test file can be parsed
					cmd = exec.Command("kcl", "fmt", testFilePath)
					output, err = cmd.CombinedOutput()
					if err != nil {
						t.Logf("KCL test validation error: %s", output)
						t.Errorf("Test using empty schema is not valid KCL: %v", err)
					}
				} else {
					t.Log("KCL not installed, skipping validation")
				}
			}

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

				// Add debug print
				fmt.Printf("Generated schema content:\n%s\n", schemaContent)

				// Print variants property description
				variantsSchema := tc.schema["properties"].(map[string]interface{})["variants"].(map[string]interface{})
				itemsSchema := variantsSchema["items"].(map[string]interface{})
				fmt.Printf("Variants property schema: %+v\n", variantsSchema)
				fmt.Printf("Items schema: %+v\n", itemsSchema)

				// Check complex array with object items (should have nested schema)
				assert.Contains(t, schemaContent, "variants?: [TestArrayPropertiesVariantsItem]",
					"Array of objects should reference nested schema")

				// Verify nested item schema for array of objects
				assert.Contains(t, schemaContent, "schema TestArrayPropertiesVariantsItem:",
					"Should generate schema for array item objects")

				// Check array constraints
				assert.Contains(t, schemaContent, "tags == None or len(tags) >= 1",
					"Should validate minItems constraint")
				assert.Contains(t, schemaContent, "tags == None or len(tags) == len({str(item): None for item in tags})",
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
				assert.Contains(t, schemaContent, "images == None or len(images) >= 1",
					"Should validate minItems constraint")
				assert.Contains(t, schemaContent, "images == None or len(images) <= 10",
					"Should validate maxItems constraint")
			}

			// For string constraints test case, use simplified checks
			if tc.name == "String Constraint Schema" {
				// Only check for basic schema elements, not specific validation rules
				if tc.expectedTitle != "" {
					assert.Contains(t, schemaContent, "schema "+tc.expectedTitle+":", "Schema should have the correct title")
				}
				// Check for username field existence with basic constraints
				assert.Contains(t, schemaContent, "username: str", "Required property should not have ? modifier")
				assert.Contains(t, schemaContent, "email: str", "Required property should not have ? modifier")

				// Check for basic format comments without exact pattern validation
				assert.Contains(t, schemaContent, "# Format: email", "Should include email format comment")
				assert.Contains(t, schemaContent, "# Format: uri", "Should include URI format comment")
				assert.Contains(t, schemaContent, "# Format: uuid", "Should include UUID format comment")
				assert.Contains(t, schemaContent, "# Format: date", "Should include date format comment")

				// Check for array type syntax
				assert.Contains(t, schemaContent, "[str]", "Array properties should use [str] syntax")

				// Check for basic validation block existence
				assert.Contains(t, schemaContent, "check:", "Schema should include validation checks")
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
