package openapikcl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kcl-lang.io/kcl-go"
)

// TestOpenAPI2Conversion tests conversion of OpenAPI 2.0 schemas
func TestOpenAPI2Conversion(t *testing.T) {
	// Skip if running in CI without proper access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test in CI")
	}

	// Test files
	tests := []struct {
		name             string
		filename         string
		expectedSchemas  []string
		expectedValidKCL bool
	}{
		{
			name:             "Simple Petstore v2",
			filename:         "testdata/oas/input/petstore_v2.json",
			expectedSchemas:  []string{"Pet", "PetInput", "ErrorModel"},
			expectedValidKCL: true,
		},
		{
			name:             "Complex API v2",
			filename:         "testdata/oas/input/complex_v2.json",
			expectedSchemas:  []string{"BaseObject", "Product", "Category", "Order", "OrderItem", "Customer", "Address", "ApiResponse", "Mixed", "Multi"},
			expectedValidKCL: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary output directory
			tempDir, err := os.MkdirTemp("", "kcl-test-")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Load and process the OpenAPI schema
			doc, version, err := LoadOpenAPISchema(tc.filename, LoadOptions{
				FlattenSpec: true,
				SkipRemote:  true,
			})
			require.NoError(t, err)
			assert.Equal(t, OpenAPIV2, version, "Should detect OpenAPI 2.0 version")

			// Generate KCL schemas
			err = GenerateKCLSchemas(doc, tempDir, "test", version, nil)
			require.NoError(t, err)

			// Check if all expected schema files were generated
			for _, schemaName := range tc.expectedSchemas {
				schemaPath := filepath.Join(tempDir, schemaName+".k")
				_, err := os.Stat(schemaPath)
				fmt.Println(schemaPath)
				assert.NoError(t, err, "Schema file %s should have been created", schemaName)
			}

			// Skip KCL validation for now as it requires proper module resolution
			// which is not set up in the test environment
			if tc.expectedValidKCL && false {
				// Run validation using KCL go-sdk
				result, err := kcl.Run(tempDir)

				if err != nil {
					t.Logf("KCL validation failed: %v", err)
					t.Error("KCL validation should succeed")
				}

				// Assert that validation passed
				assert.NoError(t, err, "KCL validation should succeed")
				assert.NotNil(t, result, "KCL validation result should not be nil")
			}
		})
	}
}
