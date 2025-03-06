// generator_test.go
package openapikcl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectSpecFormat verifies that the spec format auto-detection works correctly
func TestDetectSpecFormat(t *testing.T) {
	testCases := []struct {
		name           string
		content        []byte
		expectedFormat SpecFormat
	}{
		{
			name:           "OpenAPI 3.0.x",
			content:        []byte(`{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`),
			expectedFormat: OpenAPISpec,
		},
		{
			name:           "Swagger 2.0",
			content:        []byte(`{"swagger": "2.0", "info": {"title": "Test API", "version": "1.0.0"}}`),
			expectedFormat: OpenAPISpec,
		},
		{
			name:           "JSON Schema Draft-07",
			content:        []byte(`{"$schema": "http://json-schema.org/draft-07/schema#", "type": "object"}`),
			expectedFormat: JSONSchemaSpec,
		},
		{
			name:           "JSON Schema Draft-2019-09",
			content:        []byte(`{"$schema": "https://json-schema.org/draft/2019-09/schema", "type": "object"}`),
			expectedFormat: JSONSchemaSpec,
		},
		{
			name:           "Unknown Format",
			content:        []byte(`{"type": "object", "properties": {"foo": {"type": "string"}}}`),
			expectedFormat: UnknownSpec,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			format := detectSpecFormat(tc.content)
			assert.Equal(t, tc.expectedFormat, format)
		})
	}
}

// TestGenerateKCL tests the core generation functionality
func TestGenerateKCL(t *testing.T) {
	// Skip if running in CI without tempdir access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_TEMPDIR_TESTS") != "" {
		t.Skip("Skipping test requiring tempdir in CI")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "kcl-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test auto-detection and generation
	testInputs := []string{
		"./testdata/oas/input/petstore.json",
	}

	for _, inputFile := range testInputs {
		t.Run(filepath.Base(inputFile), func(t *testing.T) {
			// Run the generation
			err := GenerateKCL(inputFile, tempDir, "test")
			require.NoError(t, err)

			// Check if main.k was created
			mainKPath := filepath.Join(tempDir, "main.k")
			assert.FileExists(t, mainKPath)

			// Clean up temp dir contents between tests
			files, err := os.ReadDir(tempDir)
			require.NoError(t, err)
			for _, file := range files {
				os.Remove(filepath.Join(tempDir, file.Name()))
			}
		})
	}
}
