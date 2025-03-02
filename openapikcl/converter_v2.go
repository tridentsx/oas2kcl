package openapikcl

import (
	"log"
)

// This file contains code for handling OpenAPI 2.0 (Swagger) specifics.
// Since we're using kin-openapi's conversion to OpenAPI 3.0 (openapi2conv.ToV3),
// we don't need to directly handle most of the OpenAPI 2.0 specific structures.
// The conversion takes care of mapping Swagger's structures to OpenAPI 3.0.

// However, we keep this file as a placeholder for any Swagger-specific handling
// that might be needed in the future, particularly for edge cases where the
// automatic conversion doesn't provide the desired result.

// IsSwaggerVersion checks if the given version is OpenAPI 2.0
func IsSwaggerVersion(version OpenAPIVersion) bool {
	return version == OpenAPIV2
}

// HandleSwaggerSpecifics performs any special processing needed for OpenAPI 2.0
// specifications after they've been converted to OpenAPI 3.0
func HandleSwaggerSpecifics(version OpenAPIVersion) {
	if !IsSwaggerVersion(version) {
		return
	}

	log.Printf("processing OpenAPI 2.0 specific features")
	// Currently no special handling is needed as the conversion
	// provided by kin-openapi is sufficient for our purposes
}

// If we need any specific adaptations for OpenAPI 2.0, they would go here.
// For now, we rely on the conversion to OpenAPI 3.0 and then use our
// existing OpenAPI 3.0 processing code.
