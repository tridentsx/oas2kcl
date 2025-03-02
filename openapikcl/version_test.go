package openapikcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectOpenAPIVersion(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		expectedVersion OpenAPIVersion
		shouldError     bool
	}{
		{
			name:            "OpenAPI 2.0",
			json:            `{"swagger": "2.0"}`,
			expectedVersion: OpenAPIV2,
			shouldError:     false,
		},
		{
			name:            "OpenAPI 3.0.0",
			json:            `{"openapi": "3.0.0"}`,
			expectedVersion: OpenAPIV3,
			shouldError:     false,
		},
		{
			name:            "OpenAPI 3.0.1",
			json:            `{"openapi": "3.0.1"}`,
			expectedVersion: OpenAPIV3,
			shouldError:     false,
		},
		{
			name:            "OpenAPI 3.0.2",
			json:            `{"openapi": "3.0.2"}`,
			expectedVersion: OpenAPIV3,
			shouldError:     false,
		},
		{
			name:            "OpenAPI 3.0.3",
			json:            `{"openapi": "3.0.3"}`,
			expectedVersion: OpenAPIV3,
			shouldError:     false,
		},
		{
			name:            "OpenAPI 3.1.0",
			json:            `{"openapi": "3.1.0"}`,
			expectedVersion: OpenAPIV31,
			shouldError:     false,
		},
		{
			name:            "Invalid document",
			json:            `{"invalid": "json"}`,
			expectedVersion: "",
			shouldError:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			version, err := DetectOpenAPIVersion([]byte(tc.json))

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVersion, version)
			}
		})
	}
}
