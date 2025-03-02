package openapikcl

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadOpenAPISchema(t *testing.T) {
	tests := []struct {
		name                string
		file                string
		expectedVersion     OpenAPIVersion
		shouldError         bool
		expectedSchemaCount int
	}{
		{
			name:                "Load OpenAPI 3.0",
			file:                "petstore.json",
			expectedVersion:     OpenAPIV3,
			shouldError:         false,
			expectedSchemaCount: 3, // Pet, Pets, Error
		},
		{
			name:                "Load OpenAPI 2.0",
			file:                "petstore_v2.json",
			expectedVersion:     OpenAPIV2,
			shouldError:         false,
			expectedSchemaCount: 3, // Pet, PetInput, ErrorModel
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join("testdata", "input", tc.file)
			doc, version, err := LoadOpenAPISchema(filePath, LoadOptions{
				FlattenSpec: false,
			})

			if tc.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedVersion, version)
			assert.NotNil(t, doc)

			// For OpenAPI 2.0, schemas are converted from definitions
			if version == OpenAPIV2 {
				assert.NotNil(t, doc.Components)
				assert.NotNil(t, doc.Components.Schemas)
				assert.Equal(t, tc.expectedSchemaCount, len(doc.Components.Schemas))
			}
		})
	}
}

func TestLoadOpenAPISchemaWithFlattening(t *testing.T) {
	tests := []struct {
		name                string
		file                string
		expectedVersion     OpenAPIVersion
		shouldError         bool
		expectedSchemaCount int
	}{
		{
			name:                "Flatten OpenAPI 3.0",
			file:                "petstore.json",
			expectedVersion:     OpenAPIV3,
			shouldError:         false,
			expectedSchemaCount: 3, // Pet, Pets, Error
		},
		{
			name:                "Flatten OpenAPI 2.0",
			file:                "petstore_v2.json",
			expectedVersion:     OpenAPIV2,
			shouldError:         false,
			expectedSchemaCount: 3, // Pet, PetInput, ErrorModel
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join("testdata", "input", tc.file)
			doc, version, err := LoadOpenAPISchema(filePath, LoadOptions{
				FlattenSpec: true,
				SkipRemote:  true,
				MaxDepth:    5,
			})

			if tc.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedVersion, version)
			assert.NotNil(t, doc)

			// Check schema count
			if doc != nil && doc.Components != nil && doc.Components.Schemas != nil {
				assert.Equal(t, tc.expectedSchemaCount, len(doc.Components.Schemas),
					"Expected %d schemas but got %d", tc.expectedSchemaCount, len(doc.Components.Schemas))
			}
		})
	}
}
