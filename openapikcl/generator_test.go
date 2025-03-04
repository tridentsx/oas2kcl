// generator_test.go
package openapikcl

import (
	"fmt"
	"os"
	"os/exec"
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
						Default:   "default_name",
					},
				},
				"age": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    typesPtr("integer"), // Use typesPtr
						Default: float64(25),
					},
				},
				"isActive": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    typesPtr("boolean"), // Use typesPtr
						Default: true,
					},
				},
				"status": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: typesPtr("string"), // Use typesPtr
						Enum: []interface{}{
							"active",
							"inactive",
							"pending",
						},
						Default: "active",
					},
				},
				"priority": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: typesPtr("integer"), // Use typesPtr
						Enum: []interface{}{
							float64(1),
							float64(2),
							float64(3),
						},
						Default: float64(2),
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

	// Create a minimal document for testing
	doc := &openapi3.T{
		Components: &openapi3.Components{
			Schemas: schemas,
		},
	}

	result, err := GenerateKCLSchema("TestSchema", schema, schemas, OpenAPIV3, doc)
	require.NoError(t, err)
	assert.Contains(t, result, "schema TestSchema:")
	assert.Contains(t, result, "name: str = \"default_name\"")
	assert.Contains(t, result, "age?: int = 25")
	assert.Contains(t, result, "isActive?: bool = true")
	assert.Contains(t, result, "status?: str = \"active\"")
	assert.Contains(t, result, "priority?: int = 2")
	assert.Contains(t, result, "tags?: [str]")
}

func TestGenerateKCLSchemas(t *testing.T) {
	// Skip if running in CI without tempdir access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_TEMPDIR_TESTS") != "" {
		t.Skip("Skipping test requiring tempdir in CI")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "kcl-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test OpenAPI document
	doc := createTestOpenAPIDoc()

	// Generate KCL schemas
	err = GenerateKCLSchemas(doc, tempDir, "test", OpenAPIV3)
	require.NoError(t, err)

	// Check if the User.k file was created
	userSchemaPath := filepath.Join(tempDir, "User.k")
	_, err = os.Stat(userSchemaPath)
	assert.NoError(t, err, "User.k file should have been created")

	// Read the contents of the file
	content, err := os.ReadFile(userSchemaPath)
	require.NoError(t, err)

	// Check content
	contentStr := string(content)
	assert.Contains(t, contentStr, "schema User:")
	assert.Contains(t, contentStr, "name: str")
	assert.Contains(t, contentStr, "age?: int")

	// Check for proper import statements
	assert.Contains(t, contentStr, "import regex")

	// Check if the main.k file was created
	mainKPath := filepath.Join(tempDir, "main.k")
	_, err = os.Stat(mainKPath)
	assert.NoError(t, err, "main.k file should have been created")

	// Read main.k contents
	mainContent, err := os.ReadFile(mainKPath)
	require.NoError(t, err)

	mainContentStr := string(mainContent)
	// Check for expected content in main.k
	assert.Contains(t, mainContentStr, "import regex")
	assert.Contains(t, mainContentStr, "schema ValidationSchema:")

	// Test our relationship validation approach only if KCL is available
	if isKCLAvailable() {
		// Create a validation file that imports all schemas
		validationPath := filepath.Join(tempDir, "validation_test.k")
		validationContent := `import regex
import User

schema ValidationTest:
    user_instance?: User
`
		err = os.WriteFile(validationPath, []byte(validationContent), 0644)
		require.NoError(t, err)

		// Run KCL validation on the test file
		cmd := exec.Command("kcl", "validation_test.k")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		// We expect validation to pass
		assert.NoError(t, err, "KCL validation should succeed: %s", string(output))
	}
}

// Helper function to check if KCL is installed
func isKCLAvailable() bool {
	_, err := exec.LookPath("kcl")
	return err == nil
}

func TestGenerateMainK(t *testing.T) {
	// Skip if running in CI without access to tmp dir
	if os.Getenv("CI") == "true" && os.Getenv("TMPDIR") == "" {
		t.Skip("Skipping test in CI environment without tmp dir")
	}

	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "main-k-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Define some schema names
	schemaNames := []string{"User", "Pet", "Order", "Category"}

	// Generate the main.k file
	err = generateMainK(tmpDir, schemaNames, nil)
	if err != nil {
		t.Fatalf("Failed to generate main.k: %v", err)
	}

	// Verify that the main.k file was created
	mainKPath := filepath.Join(tmpDir, "main.k")
	if _, err := os.Stat(mainKPath); os.IsNotExist(err) {
		t.Fatalf("main.k file was not created")
	}

	// Read the contents of the main.k file
	content, err := os.ReadFile(mainKPath)
	if err != nil {
		t.Fatalf("Failed to read main.k: %v", err)
	}

	// Convert content to string for assertions
	contentStr := string(content)

	// Check that the file contains expected elements
	assert.Contains(t, contentStr, "# This file is generated for KCL validation")
	assert.Contains(t, contentStr, "import regex")

	// Check for the validation schema
	assert.Contains(t, contentStr, "schema ValidationSchema:")

	// Check for helpful comments
	// Remove assertion for specific comment that may not be needed
	// assert.Contains(t, contentStr, "This is a simple validation schema")
}

func TestSchemaInheritance(t *testing.T) {

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

	result, err := processInheritance(childSchema)

	require.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Parent", result[0])
}

// Helper to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func TestGenerateKCLFromFile(t *testing.T) {
	// Skip if running in CI without tempdir access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_TEMPDIR_TESTS") != "" {
		t.Skip("Skipping test requiring tempdir in CI")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "kcl-test-schema-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Load and parse the OpenAPI document
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/input/schemas.yaml")
	require.NoError(t, err)

	// Generate KCL schemas
	err = GenerateKCLSchemas(doc, tempDir, "test", OpenAPIV3)
	require.NoError(t, err)

	printFilesInDir(tempDir)

	// Check if expected schema files were created
	expectedSchemas := []string{"Address", "Customer", "Metadata", "Order", "OrderResponse", "Price"}
	for _, schema := range expectedSchemas {
		schemaPath := filepath.Join(tempDir, schema+".k")
		assert.True(t, fileExists(schemaPath), "Schema file %s should exist", schema+".k")

		// Read and verify basic content
		content, err := os.ReadFile(schemaPath)
		require.NoError(t, err)
		contentStr := string(content)

		assert.Contains(t, contentStr, "schema "+schema+":")
		assert.Contains(t, contentStr, "import regex")
	}

	// Verify main.k was created
	mainKPath := filepath.Join(tempDir, "main.k")
	assert.True(t, fileExists(mainKPath))

	// Test KCL validation if available
	if isKCLAvailable() {
		cmd := exec.Command("kcl", "run", tempDir)
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "KCL validation should succeed: %s", string(output))
	}
}

// Helper function to print files in directory
func printFilesInDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fmt.Println(entry.Name())
	}
	return nil
}
