// version.go
package openapikcl

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OpenAPIVersion represents supported OpenAPI specification versions
type OpenAPIVersion string

const (
	OpenAPIV2  OpenAPIVersion = "2.0"
	OpenAPIV3  OpenAPIVersion = "3.0"
	OpenAPIV31 OpenAPIVersion = "3.1"
)

// DetectOpenAPIVersion detects the OpenAPI version from raw data
func DetectOpenAPIVersion(data []byte) (OpenAPIVersion, error) {
	// Parse just enough to get the version
	var doc struct {
		Swagger string `json:"swagger"` // OpenAPI 2.0
		OpenAPI string `json:"openapi"` // OpenAPI 3.x
	}

	if err := json.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("failed to parse document: %w", err)
	}

	if doc.Swagger == "2.0" {
		return OpenAPIV2, nil
	} else if doc.OpenAPI == "3.0.0" || doc.OpenAPI == "3.0.1" || doc.OpenAPI == "3.0.2" || doc.OpenAPI == "3.0.3" {
		return OpenAPIV3, nil
	} else if strings.HasPrefix(doc.OpenAPI, "3.1") {
		return OpenAPIV31, nil
	}

	return "", fmt.Errorf("unsupported OpenAPI version: swagger=%q, openapi=%q", doc.Swagger, doc.OpenAPI)
}
