// functions_test.go
package oas

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create types for testing
func createTypes(typeValue string) *openapi3.Types {
	t := openapi3.Types([]string{typeValue})
	return &t
}

// Helper function to create uint64 pointer for testing
func uInt64Ptr(i uint64) *uint64 {
	return &i
}

// Helper function to create a test OpenAPI document
func createTestOpenAPIDoc() *openapi3.T {
	return &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		OpenAPI: "3.0.0",
		Paths:   openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"Pet": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: createTypes("object"),
						Properties: openapi3.Schemas{
							"id": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:        createTypes("integer"),
									Description: "Unique identifier for the pet",
								},
							},
							"name": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:        createTypes("string"),
									Description: "Name of the pet",
								},
							},
						},
						Required: []string{"name"},
					},
				},
				"Order": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: createTypes("object"),
						Properties: openapi3.Schemas{
							"id": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: createTypes("integer"),
								},
							},
							"petId": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: createTypes("integer"),
								},
							},
							"status": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: createTypes("string"),
									Enum: []interface{}{"placed", "approved", "delivered"},
								},
							},
						},
					},
				},
			},
		},
	}
}

// TestConvertTypeToKCL tests the ConvertTypeToKCL function
func TestConvertTypeToKCL(t *testing.T) {
	testCases := []struct {
		name     string
		oapiType string
		format   string
		expected string
	}{
		{"string", "string", "", "str"},
		{"string with email format", "string", "email", "str"},
		{"string with date-time format", "string", "date-time", "str"},
		{"integer", "integer", "", "int"},
		{"integer with int64 format", "integer", "int64", "int"},
		{"number", "number", "", "float"},
		{"number with float format", "number", "float", "float"},
		{"boolean", "boolean", "", "bool"},
		{"array", "array", "", "[any]"},
		{"object", "object", "", "{str:any}"},
		{"null", "null", "", "None"},
		{"unknown", "unknown", "", "any"},
		{"empty", "", "", "any"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertTypeToKCL(tc.oapiType, tc.format)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGenerateConstraints tests the GenerateConstraints function
func TestGenerateConstraints(t *testing.T) {
	testCases := []struct {
		name           string
		schema         *openapi3.Schema
		fieldName      string
		useSelfPrefix  bool
		expectedPrefix string
		expectedLen    int
	}{
		{
			name: "string with minLength",
			schema: &openapi3.Schema{
				Type:      createTypes("string"),
				MinLength: 3,
			},
			fieldName:      "name",
			useSelfPrefix:  false,
			expectedPrefix: "name",
			expectedLen:    1,
		},
		{
			name: "string with maxLength",
			schema: &openapi3.Schema{
				Type:      createTypes("string"),
				MaxLength: uInt64Ptr(10),
			},
			fieldName:      "name",
			useSelfPrefix:  false,
			expectedPrefix: "name",
			expectedLen:    1,
		},
		{
			name: "string with minLength and maxLength",
			schema: &openapi3.Schema{
				Type:      createTypes("string"),
				MinLength: 3,
				MaxLength: uInt64Ptr(10),
			},
			fieldName:      "name",
			useSelfPrefix:  false,
			expectedPrefix: "name",
			expectedLen:    2,
		},
		{
			name: "string with pattern",
			schema: &openapi3.Schema{
				Type:    createTypes("string"),
				Pattern: "^[a-zA-Z0-9]+$",
			},
			fieldName:      "name",
			useSelfPrefix:  false,
			expectedPrefix: "name",
			expectedLen:    1,
		},
		{
			name: "string with enum",
			schema: &openapi3.Schema{
				Type: createTypes("string"),
				Enum: []interface{}{"foo", "bar", "baz"},
			},
			fieldName:      "status",
			useSelfPrefix:  false,
			expectedPrefix: "status",
			expectedLen:    1,
		},
		{
			name: "integer with minimum",
			schema: &openapi3.Schema{
				Type: createTypes("integer"),
				Min:  openapi3.Float64Ptr(0),
			},
			fieldName:      "age",
			useSelfPrefix:  false,
			expectedPrefix: "age",
			expectedLen:    1,
		},
		{
			name: "integer with maximum",
			schema: &openapi3.Schema{
				Type: createTypes("integer"),
				Max:  openapi3.Float64Ptr(100),
			},
			fieldName:      "age",
			useSelfPrefix:  false,
			expectedPrefix: "age",
			expectedLen:    1,
		},
		{
			name: "number with exclusiveMinimum",
			schema: &openapi3.Schema{
				Type:         createTypes("number"),
				Min:          openapi3.Float64Ptr(0),
				ExclusiveMin: true,
			},
			fieldName:      "price",
			useSelfPrefix:  true,
			expectedPrefix: "self",
			expectedLen:    1,
		},
		{
			name: "array with minItems",
			schema: &openapi3.Schema{
				Type:     createTypes("array"),
				MinItems: 1,
			},
			fieldName:      "items",
			useSelfPrefix:  false,
			expectedPrefix: "items",
			expectedLen:    1,
		},
		{
			name: "array with maxItems",
			schema: &openapi3.Schema{
				Type:     createTypes("array"),
				MaxItems: uInt64Ptr(10),
			},
			fieldName:      "items",
			useSelfPrefix:  false,
			expectedPrefix: "items",
			expectedLen:    1,
		},
		{
			name: "array with uniqueItems",
			schema: &openapi3.Schema{
				Type:        createTypes("array"),
				UniqueItems: true,
			},
			fieldName:      "tags",
			useSelfPrefix:  false,
			expectedPrefix: "tags",
			expectedLen:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			constraints := GenerateConstraints(tc.schema, tc.fieldName, tc.useSelfPrefix)
			assert.Len(t, constraints, tc.expectedLen)
			if len(constraints) > 0 {
				for _, constraint := range constraints {
					assert.Contains(t, constraint, tc.expectedPrefix)
				}
			}
		})
	}
}

// TestFormatDocumentation tests the FormatDocumentation function
func TestFormatDocumentation(t *testing.T) {
	testCases := []struct {
		name     string
		schema   *openapi3.Schema
		expected string
	}{
		{
			name: "with title and description",
			schema: &openapi3.Schema{
				Title:       "Test Schema",
				Description: "This is a test schema",
			},
			expected: "Test Schema\n\nThis is a test schema",
		},
		{
			name: "with title only",
			schema: &openapi3.Schema{
				Title: "Test Schema",
			},
			expected: "Test Schema",
		},
		{
			name: "with description only",
			schema: &openapi3.Schema{
				Description: "This is a test schema",
			},
			expected: "This is a test schema",
		},
		{
			name: "with multiline description",
			schema: &openapi3.Schema{
				Description: "This is a test schema\nWith multiple lines",
			},
			expected: "This is a test schema\nWith multiple lines",
		},
		{
			name:     "empty schema",
			schema:   &openapi3.Schema{},
			expected: "",
		},
		{
			name:     "nil schema",
			schema:   nil,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatDocumentation(tc.schema)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGenerateSchemas tests the GenerateSchemas function with different OpenAPI documents
func TestGenerateSchemas(t *testing.T) {
	testCases := []struct {
		name        string
		createDoc   func() *openapi3.T
		outputDir   string
		packageName string
		version     OpenAPIVersion
		expectedLen int
	}{
		{
			name:        "Simple API",
			createDoc:   createTestOpenAPIDoc,
			outputDir:   "testdata/simple/output",
			packageName: "simple",
			version:     OpenAPIV3,
			expectedLen: 2, // Pet and Order
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean output directory
			os.RemoveAll(tc.outputDir)
			os.MkdirAll(tc.outputDir, 0755)

			// Create test document
			doc := tc.createDoc()

			// Generate KCL schemas
			err := GenerateSchemas(doc, tc.outputDir, tc.packageName, tc.version)
			require.NoError(t, err)

			// Check if the output files exist
			files, err := os.ReadDir(tc.outputDir)
			require.NoError(t, err)

			// At minimum we expect a main.k plus a schema file for each schema
			assert.GreaterOrEqual(t, len(files), tc.expectedLen)

			// Check that main.k exists
			assert.FileExists(t, filepath.Join(tc.outputDir, "main.k"))

			// Check that expected schema files exist
			for schemaName := range doc.Components.Schemas {
				assert.FileExists(t, filepath.Join(tc.outputDir, schemaName+".k"))
			}
		})
	}
}

// TestGenerateKCLSchema tests the GenerateKCLSchema function
func TestGenerateKCLSchema(t *testing.T) {
	// Create a sample OpenAPI document
	doc := createTestOpenAPIDoc()
	require.NotNil(t, doc)
	require.NotNil(t, doc.Components)
	require.NotNil(t, doc.Components.Schemas)

	// Get the Pet schema
	schemas := doc.Components.Schemas
	schema, ok := schemas["Pet"]
	require.True(t, ok, "Pet schema not found in test document")
	require.NotNil(t, schema)

	// Generate KCL schema
	result, err := GenerateKCLSchema("TestSchema", schema, schemas, OpenAPIV3, doc)
	require.NoError(t, err)
	assert.Contains(t, result, "schema TestSchema:")
	assert.Contains(t, result, "name: str")
	assert.Contains(t, result, "id?: int")
}

// TestCheckIfNeedsRegexImport tests the CheckIfNeedsRegexImport function
func TestCheckIfNeedsRegexImport(t *testing.T) {
	testCases := []struct {
		name     string
		schema   *openapi3.SchemaRef
		expected bool
	}{
		{
			name: "schema with pattern",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:    createTypes("string"),
					Pattern: "^[a-zA-Z0-9]+$",
				},
			},
			expected: true,
		},
		{
			name: "schema with email format",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   createTypes("string"),
					Format: "email",
				},
			},
			expected: true,
		},
		{
			name: "schema with no regex requirements",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: createTypes("integer"),
				},
			},
			expected: false,
		},
		{
			name: "nested schema with pattern",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: createTypes("object"),
					Properties: openapi3.Schemas{
						"email": &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:   createTypes("string"),
								Format: "email",
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "array schema with pattern items",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: createTypes("array"),
					Items: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    createTypes("string"),
							Pattern: "^[a-zA-Z0-9]+$",
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checkIfNeedsRegexImport(tc.schema)
			assert.Equal(t, tc.expected, result)
		})
	}
}
