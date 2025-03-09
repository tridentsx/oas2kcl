package testpkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tridentsx/oas2kcl/openapikcl"
)

// TestBasicDetectFormat tests the format detection functionality
func TestBasicDetectFormat(t *testing.T) {
	// Test JSON Schema detection
	jsonSchema := `{"$schema": "http://json-schema.org/draft-07/schema#", "type": "object"}`
	format, err := openapikcl.DetectSpecFormat([]byte(jsonSchema))
	require.NoError(t, err)
	assert.Equal(t, openapikcl.SpecFormatJSONSchema, format)

	// Test OpenAPI 3.0 detection
	openAPI3 := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}, "paths": {}}`
	format, err = openapikcl.DetectSpecFormat([]byte(openAPI3))
	require.NoError(t, err)
	assert.Equal(t, openapikcl.SpecFormatOpenAPIV3, format)
}
