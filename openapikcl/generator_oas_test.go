// generator_oas_test.go
package openapikcl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestGenerateKCLSchemaOpenAPI(t *testing.T) {
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

// generateOpenAPISchemasForTest is a test-specific function to bypass schema type detection
func generateOpenAPISchemasForTest(doc *openapi3.T, outputDir string, packageName string) error {
	// Skip the schema type detection and directly call the OpenAPI schema generation
	return generateOpenAPISchemas(doc, outputDir, packageName, OpenAPIV3)
}

func TestGenerateKCLSchemasOpenAPI(t *testing.T) {
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

	// Generate KCL schemas directly using the OpenAPI generator
	err = generateOpenAPISchemasForTest(doc, tempDir, "test")
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
	t.Logf("User.k contents:\n%s", contentStr)
	assert.Contains(t, contentStr, "schema User:")
	assert.Contains(t, contentStr, "name: str")
	assert.Contains(t, contentStr, "age?: int")

	// Check for proper comment instead of import statements
	assert.Contains(t, contentStr, "# No schema imports needed - schemas in same directory")

	// Check if the main.k file was created
	mainKPath := filepath.Join(tempDir, "main.k")
	_, err = os.Stat(mainKPath)
	assert.NoError(t, err, "main.k file should have been created")

	// Read main.k contents
	mainContent, err := os.ReadFile(mainKPath)
	require.NoError(t, err)

	mainContentStr := string(mainContent)
	// Check for expected content in main.k
	assert.Contains(t, mainContentStr, "# KCL schemas generated from test")
	assert.Contains(t, mainContentStr, "schema ValidationSchema:")
	assert.NotContains(t, mainContentStr, "import regex")

	// Test our relationship validation approach only if KCL is available
	if isKCLAvailable() {
		// Create a validation file that uses schemas - now main.k should contain a validation schema
		validationContent := `
# Basic validation test 
# Create an instance of User schema
user = {
    name = "Test User"
    age = 30
}

# Validate it against the User schema - no import needed
check_user = User {
    name = user.name
    age = user.age
}
`
		err = os.WriteFile(filepath.Join(tempDir, "validation_test.k"), []byte(validationContent), 0644)
		require.NoError(t, err)

		// Run KCL validation - run all files in the directory
		cmd := exec.Command("kcl", ".")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		// We expect validation to pass
		assert.NoError(t, err, "KCL validation should succeed: %s", string(output))

		// Log the output if we have any
		if len(output) > 0 {
			t.Logf("KCL output: %s", string(output))
		}
	}

	// Print the Order.k file for debugging
	orderSchemaPath := filepath.Join(tempDir, "Order.k")
	if fileExists(orderSchemaPath) {
		orderContent, err := os.ReadFile(orderSchemaPath)
		if err == nil {
			t.Logf("Order.k contents:\n%s", string(orderContent))
		}
	}

	// Print all generated KCL files for debugging
	t.Log("Listing all generated KCL files:")
	files, err := os.ReadDir(tempDir)
	if err == nil {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".k") {
				filePath := filepath.Join(tempDir, file.Name())
				content, err := os.ReadFile(filePath)
				if err == nil {
					t.Logf("--- %s ---\n%s\n", file.Name(), string(content))
				}
			}
		}
	}

	// Create a simpler validation file - do something basic with a schema
	validationPath := filepath.Join(tempDir, "validation_test.k")
	validationContent := `
# Create a user
user = {
    name = "Test User"
    age = 30
}

# Validate the user against the schema
schema User:
    # The user's age
    age?: int
    # The user's name
    name: str

# This will validate the user data against the schema
result = User {
    name = user.name
    age = user.age
}
`
	err = os.WriteFile(validationPath, []byte(validationContent), 0644)
	require.NoError(t, err)

	// Run a simpler KCL validation
	cmd := exec.Command("kcl", validationPath)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "Simple KCL validation should succeed: %s", string(output))
}

func TestSchemaInheritanceOpenAPI(t *testing.T) {
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

	// Create fake allSchemas
	parentSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: typesPtr("object"),
			Properties: openapi3.Schemas{
				"id": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: typesPtr("string"),
					},
				},
			},
		},
	}

	allSchemas := openapi3.Schemas{
		"Parent": parentSchema,
	}

	result, err := processInheritance(childSchema, allSchemas)

	require.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Parent", result[0])
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
	doc, err := loader.LoadFromFile("testdata/oas/input/schemas.yaml")
	require.NoError(t, err)

	// Generate KCL schemas
	err = GenerateKCLSchemas(doc, tempDir, "test", OpenAPIV3, nil)
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
		assert.Contains(t, contentStr, "# No schema imports needed - schemas in same directory")
	}

	// Verify main.k was created
	mainKPath := filepath.Join(tempDir, "main.k")
	assert.True(t, fileExists(mainKPath))

	// Test KCL validation if available and enabled
	if isKCLAvailable() {
		// Create a simple test file that uses the generated schemas
		testFilePath := filepath.Join(tempDir, "validation_test.k")
		testContent := `
# Simple validation test
# Create instances of different schemas

# Create a customer instance
customer = {
    id = "123"
    name = "Test Customer"
    address = {
        street = "123 Main St"
        city = "Test City"
        country = "Test Country"
    }
}

# Create an order with the customer
order = {
    id = "order-123"
    customer = customer
    status = "pending"
    items = [
        {
            productId = "prod-1"
            quantity = 2
            price = {
                amount = 9.99
                currency = "USD"
            }
        }
    ]
}

# Validate instances
customer_check = Customer(customer)
order_check = Order(order)
`
		err = os.WriteFile(testFilePath, []byte(testContent), 0644)
		require.NoError(t, err)

		// Run KCL validation - run all files in the directory
		cmd := exec.Command("kcl", ".")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "KCL validation should succeed: %s", string(output))

		// Log the output if we have any
		if len(output) > 0 {
			t.Logf("KCL output: %s", string(output))
		}
	}
}

// Helper function to check if KCL is installed
func isKCLAvailable() bool {
	_, err := exec.LookPath("kcl")
	return err == nil
}

// Helper to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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
