package openapikcl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// LoadOptions configures the loading process
type LoadOptions struct {
	FlattenSpec bool
	SkipRemote  bool
	MaxDepth    int
}

// LoadOpenAPISchema is now version-aware
func LoadOpenAPISchema(filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	log.Printf("reading OpenAPI schema file: %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("error reading file: %w", err)
	}

	// Check file extension to determine format
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".yaml" || ext == ".yml" {
		log.Printf("detected YAML format from file extension: %s", ext)
	} else if ext == ".json" {
		log.Printf("detected JSON format from file extension: %s", ext)
	} else {
		log.Printf("no specific format identified from extension: %s, will attempt auto-detection", ext)
	}

	// Detect version
	version, err := DetectOpenAPIVersion(data)
	if err != nil {
		return nil, "", err
	}

	switch version {
	case OpenAPIV2:
		return loadOpenAPIV2Schema(data, filePath, opts)
	case OpenAPIV3:
		return loadOpenAPIV3Schema(data, filePath, opts)
	case OpenAPIV31:
		return loadOpenAPIV31Schema(data, filePath, opts)
	default:
		return nil, "", fmt.Errorf("unsupported OpenAPI version: %s", version)
	}
}

// Separate functions for each version
func loadOpenAPIV3Schema(data []byte, filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	log.Print("parsing OpenAPI schema")
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		log.Printf("error parsing schema: %v", err)
		return nil, OpenAPIV3, err
	}

	log.Print("validating OpenAPI document")
	if err := doc.Validate(loader.Context); err != nil {
		log.Printf("schema validation failed: %v", err)
		return nil, OpenAPIV3, err
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
			return nil, OpenAPIV3, err
		}
		doc = flatDoc
		log.Print("successfully flattened OpenAPI specification")
	}

	log.Print("successfully loaded and validated OpenAPI schema")
	return doc, OpenAPIV3, nil
}

func loadOpenAPIV2Schema(data []byte, filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	// Load as OpenAPI 2.0
	swagger := &openapi2.T{}
	if err := json.Unmarshal(data, swagger); err != nil {
		return nil, OpenAPIV2, fmt.Errorf("error parsing OpenAPI 2.0 document: %w", err)
	}

	// Convert to OpenAPI 3.0
	loader := openapi3.NewLoader()
	doc, err := openapi2conv.ToV3(swagger)
	if err != nil {
		return nil, OpenAPIV2, fmt.Errorf("error converting OpenAPI 2.0 to 3.0: %w", err)
	}

	// Ensure OpenAPI version is set properly after conversion
	if doc.OpenAPI == "" {
		doc.OpenAPI = "3.0.0"
	}

	// Skip validation for OpenAPI 2.0 documents
	// They don't always pass OpenAPI 3.0 validation due to conversion differences
	log.Print("skipping validation for converted OpenAPI 2.0 document")

	// Optionally we can log loader settings for debugging
	log.Printf("using loader with context: %v", loader.Context != nil)

	// Handle flattening if needed
	if opts.FlattenSpec {
		log.Print("starting specification flattening process")
		flattener := NewFlattener(FlattenOptions{
			BaseDir:    filepath.Dir(filePath),
			MaxDepth:   opts.MaxDepth,
			SkipRemote: opts.SkipRemote,
		}, doc)

		flatDoc, err := flattener.FlattenSpec()
		if err != nil {
			return nil, OpenAPIV2, fmt.Errorf("error flattening specification: %w", err)
		}
		doc = flatDoc
	}

	return doc, OpenAPIV2, nil
}

func loadOpenAPIV31Schema(data []byte, filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	// Placeholder for future OpenAPI 3.1 support
	return nil, OpenAPIV31, fmt.Errorf("OpenAPI 3.1 support not yet implemented")
}
