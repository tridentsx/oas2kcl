package openapikcl

import (
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadOpenAPISchema reads and validates an OpenAPI schema file
func LoadOpenAPISchema(filePath string) (*openapi3.T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, err
	}

	// Validate the OpenAPI document
	if err := doc.Validate(loader.Context); err != nil {
		return nil, err
	}

	return doc, nil
}

