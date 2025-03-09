package openapikcl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tridentsx/oas2kcl/openapikcl"
	"github.com/tridentsx/oas2kcl/openapikcl/oas"
)

// TestDetectSpecFormat consolidates spec format detection tests
func TestDetectSpecFormat(t *testing.T) {
	testCases := []struct {
		name           string
		data           string
		expectedFormat openapikcl.SpecFormat
		expectError    bool
	}{
		// JSON Schema detection tests
		{
			name:           "JSON Schema with $schema property",
			data:           `{"$schema": "http://json-schema.org/draft-07/schema#", "type": "object"}`,
			expectedFormat: openapikcl.SpecFormatJSONSchema,
			expectError:    false,
		},
		{
			name:           "JSON Schema with type property",
			data:           `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			expectedFormat: openapikcl.SpecFormatJSONSchema,
			expectError:    false,
		},
		// OpenAPI detection tests
		{
			name:           "OpenAPI 3.0 with openapi property",
			data:           `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}, "paths": {}}`,
			expectedFormat: openapikcl.SpecFormatOpenAPIV3,
			expectError:    false,
		},
		{
			name:           "OpenAPI 2.0/Swagger with swagger property",
			data:           `{"swagger": "2.0", "info": {"title": "Test API", "version": "1.0.0"}, "paths": {}}`,
			expectedFormat: openapikcl.SpecFormatOpenAPIV2,
			expectError:    false,
		},
		// Error case
		{
			name:           "Unknown format",
			data:           `{"unknown": "format"}`,
			expectedFormat: openapikcl.SpecFormatUnknown,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			format, err := openapikcl.DetectSpecFormat([]byte(tc.data))

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedFormat, format)
		})
	}
}

// TestDetectOpenAPIVersion tests OpenAPI version detection
func TestDetectOpenAPIVersion(t *testing.T) {
	testCases := []struct {
		name            string
		json            string
		expectedVersion oas.OpenAPIVersion
		expectError     bool
	}{
		{
			name:            "OpenAPI 2.0",
			json:            `{"swagger": "2.0"}`,
			expectedVersion: oas.OpenAPIV2,
			expectError:     false,
		},
		{
			name:            "OpenAPI 3.0.0",
			json:            `{"openapi": "3.0.0"}`,
			expectedVersion: oas.OpenAPIV3,
			expectError:     false,
		},
		{
			name:            "OpenAPI 3.0.1",
			json:            `{"openapi": "3.0.1"}`,
			expectedVersion: oas.OpenAPIV3,
			expectError:     false,
		},
		{
			name:            "OpenAPI 3.0.2",
			json:            `{"openapi": "3.0.2"}`,
			expectedVersion: oas.OpenAPIV3,
			expectError:     false,
		},
		{
			name:            "OpenAPI 3.0.3",
			json:            `{"openapi": "3.0.3"}`,
			expectedVersion: oas.OpenAPIV3,
			expectError:     false,
		},
		{
			name:            "OpenAPI 3.1.0",
			json:            `{"openapi": "3.1.0"}`,
			expectedVersion: oas.OpenAPIV31,
			expectError:     false,
		},
		{
			name:            "Invalid document",
			json:            `{"invalid": true}`,
			expectedVersion: "",
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version, err := oas.DetectOpenAPIVersion([]byte(tc.json))

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedVersion, version)
		})
	}
}
