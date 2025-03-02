// generator_test.go
package openapikcl

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a pointer to openapi3.Types value
// openapi3.Types is a slice of strings, not a single string
func typesPtr(t string) *openapi3.Types {
	result := openapi3.Types{t} // Create a slice with the single type
	return &result
}

// Helper function to create a simple OpenAPI document for testing
func createTestOpenAPIDoc() *openapi3.T {
	schemaRef := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: typesPtr("object"), // Use typesPtr instead of string literal
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:        typesPtr("string"), // Use typesPtr
						Description: "The user's name",
					},
				},
				"age": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:        typesPtr("integer"), // Use typesPtr
						Description: "The user's age",
					},
				},
			},
			Required: []string{"name"},
		},
	}

	doc := &openapi3.T{
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"User": schemaRef,
			},
		},
	}
	return doc
}

func TestGenerateKCLSchema(t *testing.T) {
	schema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        typesPtr("object"), // Use typesPtr
			Description: "Test schema",
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:      typesPtr("string"), // Use typesPtr
						MinLength: 1,
					},
				},
				"tags": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: typesPtr("array"), // Use typesPtr
						Items: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: typesPtr("string"), // Use typesPtr
							},
						},
					},
				},
			},
			Required: []string{"name"},
		},
	}

	schemas := openapi3.Schemas{
		"TestSchema": schema,
	}

	result, err := generateKCLSchema("TestSchema", schema, schemas)
	require.NoError(t, err)
	assert.Contains(t, result, "schema TestSchema:")
	assert.Contains(t, result, "name: str")
	assert.Contains(t, result, "tags: [str] | None")
}

func TestGenerateKCLSchemas(t *testing.T) {
	// Skip if running in CI without tempdir access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_TEMPDIR_TESTS") != "" {
		t.Skip("Skipping test requiring tempdir in CI")
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "kcl-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test OpenAPI document
	doc := createTestOpenAPIDoc()

	// Generate KCL schemas
	err = GenerateKCLSchemas(doc, tempDir)
	require.NoError(t, err)

	// Check if the User.k file was created
	userSchemaPath := filepath.Join(tempDir, "User.k")
	_, err = os.Stat(userSchemaPath)
	assert.NoError(t, err, "User.k file should have been created")

	// Read the contents of the file
	content, err := ioutil.ReadFile(userSchemaPath)
	require.NoError(t, err)

	// Check content
	contentStr := string(content)
	assert.Contains(t, contentStr, "schema User:")
	assert.Contains(t, contentStr, "name: str")
	assert.Contains(t, contentStr, "age: int | None")
}

func TestSchemaInheritance(t *testing.T) {
	// Create parent schema
	parentSchema := &openapi3.Schema{
		Type: typesPtr("object"), // Already using typesPtr correctly
		Properties: openapi3.Schemas{
			"id": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: typesPtr("string"), // Already using typesPtr correctly
				},
			},
		},
	}

	// Create child schema with allOf reference
	childSchema := &openapi3.Schema{
		AllOf: []*openapi3.SchemaRef{
			{
				Ref: "#/components/schemas/Parent",
			},
		},
		Properties: openapi3.Schemas{
			"name": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: typesPtr("string"), // Already using typesPtr correctly
				},
			},
		},
	}

	result, err := processInheritance(childSchema, openapi3.Schemas{
		"Parent": &openapi3.SchemaRef{Value: parentSchema},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Parent", result[0])
}

// Helper to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
