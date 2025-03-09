package openapikcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tridentsx/oas2kcl/openapikcl/oas"
)

func TestLoadOpenAPISchema(t *testing.T) {
	tests := []struct {
		name                string
		filename            string
		expectedVersion     OpenAPIVersion
		shouldError         bool
		expectedSchemaCount int
	}{
		{
			name:                "Load OpenAPI 3.0",
			filename:            "testdata/oas/input/petstore.json",
			expectedVersion:     oas.OpenAPIV3,
			shouldError:         false,
			expectedSchemaCount: 3, // Pet, Pets, Error
		},
		{
			name:                "Load OpenAPI 2.0",
			filename:            "testdata/oas/input/petstore_v2.json",
			expectedVersion:     oas.OpenAPIV2,
			shouldError:         false,
			expectedSchemaCount: 3, // Pet, PetInput, ErrorModel
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.filename
			doc, version, err := LoadOpenAPISchema(filePath, LoadOptions{
				Flatten: false,
			})

			if tc.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedVersion, version)
			assert.NotNil(t, doc)

			// For OpenAPI 2.0, schemas are converted from definitions
			if version == oas.OpenAPIV2 {
				assert.NotNil(t, doc.Components)
				assert.NotNil(t, doc.Components.Schemas)
				assert.Equal(t, tc.expectedSchemaCount, len(doc.Components.Schemas))
			}
		})
	}
}

func TestLoadOpenAPISchemaWithFlattening(t *testing.T) {
	testCases := []struct {
		name        string
		filePath    string
		flatten     bool
		expectError bool
	}{
		{
			name:        "Flatten OpenAPI Schema",
			filePath:    "testdata/oas/input/petstore.yaml",
			flatten:     true,
			expectError: false,
		},
		{
			name:        "Do Not Flatten OpenAPI Schema",
			filePath:    "testdata/oas/input/petstore.yaml",
			flatten:     false,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, _, err := LoadOpenAPISchema(tc.filePath, LoadOptions{
				Flatten:            tc.flatten,
				ResolveReferences:  true,
				ValidateReferences: true,
			})

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, doc)
		})
	}
}
