package openapikcl

import (
	"log"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadOptions configures the loading process
type LoadOptions struct {
	FlattenSpec bool
	SkipRemote  bool
	MaxDepth    int
}

// LoadOpenAPISchema reads, validates, and optionally flattens an OpenAPI schema file
func LoadOpenAPISchema(filePath string, opts LoadOptions) (*openapi3.T, error) {
	log.Printf("reading OpenAPI schema file: %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("error reading file: %v", err)
		return nil, err
	}

	log.Print("parsing OpenAPI schema")
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		log.Printf("error parsing schema: %v", err)
		return nil, err
	}

	log.Print("validating OpenAPI document")
	if err := doc.Validate(loader.Context); err != nil {
		log.Printf("schema validation failed: %v", err)
		return nil, err
	}

	// Flatten the specification if requested
	if opts.FlattenSpec {
		log.Print("flattening OpenAPI specification")
		flattener := NewFlattener(FlattenOptions{
			BaseDir:    filepath.Dir(filePath),
			MaxDepth:   opts.MaxDepth,
			SkipRemote: opts.SkipRemote,
		}, doc)

		flatDoc, err := flattener.FlattenSpec()
		if err != nil {
			log.Printf("error flattening specification: %v", err)
			return nil, err
		}
		doc = flatDoc
		log.Print("successfully flattened OpenAPI specification")
	}

	log.Print("successfully loaded and validated OpenAPI schema")
	return doc, nil
}
