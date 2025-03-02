package openapikcl

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPI2Conversion tests conversion of OpenAPI 2.0 schemas
func TestOpenAPI2Conversion(t *testing.T) {
	// Skip if running in CI without proper access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test in CI")
	}

	// Validate KCL is installed
	_, err := exec.LookPath("kcl")
	if err != nil {
		t.Skip("KCL not found in PATH, skipping KCL validation test")
	}

	// Test files
	tests := []struct {
		name             string
		inputFile        string
		expectedSchemas  []string
		expectedValidKCL bool
	}{
		{
			name:             "Simple Petstore v2",
			inputFile:        filepath.Join("testdata", "input", "petstore_v2.json"),
			expectedSchemas:  []string{"Pet", "PetInput", "ErrorModel"},
			expectedValidKCL: true,
		},
		{
			name:             "Complex API v2",
			inputFile:        filepath.Join("testdata", "input", "complex_v2.json"),
			expectedSchemas:  []string{"BaseObject", "Product", "Category", "Order", "OrderItem", "Customer", "Address", "ApiResponse", "Mixed", "Multi"},
			expectedValidKCL: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary output directory
			tempDir, err := ioutil.TempDir("", "kcl-test-")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Load and process the OpenAPI schema
			doc, version, err := LoadOpenAPISchema(tc.inputFile, LoadOptions{
				FlattenSpec: true,
				SkipRemote:  true,
			})
			require.NoError(t, err)
			assert.Equal(t, OpenAPIV2, version, "Should detect OpenAPI 2.0 version")

			// Generate KCL schemas
			err = GenerateKCLSchemas(doc, tempDir, "test", version)
			require.NoError(t, err)

			// Check if all expected schema files were generated
			for _, schemaName := range tc.expectedSchemas {
				schemaPath := filepath.Join(tempDir, schemaName+".k")
				_, err := os.Stat(schemaPath)
				assert.NoError(t, err, "Schema file %s should have been created", schemaName)
			}

			// Run KCL validation
			if tc.expectedValidKCL {
				// Instead of validating all schemas at once with main.k,
				// validate each schema individually to avoid circular references
				var validationSuccess = true

				files, err := ioutil.ReadDir(tempDir)
				require.NoError(t, err)

				for _, file := range files {
					if !file.IsDir() && strings.HasSuffix(file.Name(), ".k") {
						// Skip main.k file
						if file.Name() == "main.k" {
							continue
						}

						// Create validation file for this schema
						schemaName := strings.TrimSuffix(file.Name(), ".k")
						validationFile := filepath.Join(tempDir, "validate_"+schemaName+".k")

						var validationContent strings.Builder
						validationContent.WriteString("# Validation file for " + schemaName + "\n")
						validationContent.WriteString("import regex\n")                // Standard import
						validationContent.WriteString("import " + schemaName + "\n\n") // Import the schema
						validationContent.WriteString("schema Validation:\n")
						validationContent.WriteString("    dummy: str = \"test\"\n")

						err = ioutil.WriteFile(validationFile, []byte(validationContent.String()), 0644)
						require.NoError(t, err)

						// Run KCL validation for this schema
						cmd := exec.Command("kcl", validationFile)
						cmd.Dir = tempDir
						output, err := cmd.CombinedOutput()

						if err != nil {
							t.Logf("KCL validation failed for schema %s: %s", schemaName, string(output))
							validationSuccess = false
						}
					}
				}

				// Assert that all validations passed
				assert.True(t, validationSuccess, "KCL validation should succeed")
			}
		})
	}
}
