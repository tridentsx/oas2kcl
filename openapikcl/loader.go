package openapikcl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tridentsx/oas2kcl/openapikcl/oas"
	"gopkg.in/yaml.v3"
)

// OpenAPIVersion is an alias for oas.OpenAPIVersion
type OpenAPIVersion = oas.OpenAPIVersion

// LoadOptions contains options for loading OpenAPI schemas
type LoadOptions struct {
	Flatten            bool // Flatten the schema
	ResolveReferences  bool // Resolve external references
	ValidateReferences bool // Validate references
	MaxDepth           int  // Maximum depth for reference resolution
	SkipRemote         bool // Skip remote references
}

// LoadOpenAPISchema loads an OpenAPI schema from a file
func LoadOpenAPISchema(filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	log.Printf("reading OpenAPI schema file: %s", filePath)

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("error reading file: %w", err)
	}

	// Detect version
	version, err := oas.DetectOpenAPIVersion(data)
	if err != nil {
		return nil, "", fmt.Errorf("error detecting OpenAPI version: %w", err)
	}

	switch version {
	case oas.OpenAPIV2:
		return loadOpenAPIV2Schema(data, filePath, opts)
	case oas.OpenAPIV3:
		return loadOpenAPIV3Schema(data, filePath, opts)
	case oas.OpenAPIV31:
		return loadOpenAPIV31Schema(data, filePath, opts)
	default:
		return nil, "", fmt.Errorf("unsupported OpenAPI version: %s", version)
	}
}

// loadOpenAPIV3Schema loads an OpenAPI 3.0 schema
func loadOpenAPIV3Schema(data []byte, filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	log.Print("parsing OpenAPI schema")
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = opts.ResolveReferences

	doc, err := loader.LoadFromData(data)
	if err != nil {
		log.Printf("error parsing schema: %v", err)
		return nil, oas.OpenAPIV3, err
	}

	log.Print("validating OpenAPI document")
	if opts.ValidateReferences {
		if err := doc.Validate(loader.Context); err != nil {
			log.Printf("schema validation failed: %v", err)
			return nil, oas.OpenAPIV3, err
		}
	}

	// Flatten the specification if requested
	if opts.Flatten {
		log.Print("flattening OpenAPI specification")
		flatDoc, err := oas.FlattenDocument(doc, oas.FlattenOptions{
			BaseDir:    filepath.Dir(filePath),
			MaxDepth:   opts.MaxDepth,
			SkipRemote: opts.SkipRemote,
		})
		if err != nil {
			log.Printf("error flattening specification: %v", err)
			return nil, oas.OpenAPIV3, err
		}
		doc = flatDoc
	}

	log.Print("successfully loaded and validated OpenAPI schema")
	return doc, oas.OpenAPIV3, nil
}

// loadOpenAPIV2Schema loads an OpenAPI 2.0 schema and converts it to 3.0
func loadOpenAPIV2Schema(data []byte, filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	// Load as OpenAPI 2.0
	swagger := &openapi2.T{}
	if err := parseYAMLOrJSON(data, swagger); err != nil {
		return nil, oas.OpenAPIV2, fmt.Errorf("error parsing OpenAPI 2.0 document: %w", err)
	}

	// Convert to OpenAPI 3.0
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = opts.ResolveReferences
	doc, err := openapi2conv.ToV3(swagger)
	if err != nil {
		return nil, oas.OpenAPIV2, fmt.Errorf("error converting OpenAPI 2.0 to 3.0: %w", err)
	}

	// Mark this as originally a Swagger document (we can't use AddExtension directly)
	if doc.Extensions == nil {
		doc.Extensions = make(map[string]interface{})
	}
	doc.Extensions["x-original-swagger-version"] = "2.0"

	// Handle flattening if needed
	if opts.Flatten {
		log.Print("starting specification flattening process")
		flatDoc, err := oas.FlattenDocument(doc, oas.FlattenOptions{
			BaseDir:    filepath.Dir(filePath),
			MaxDepth:   opts.MaxDepth,
			SkipRemote: opts.SkipRemote,
		})
		if err != nil {
			return nil, oas.OpenAPIV2, fmt.Errorf("error flattening specification: %w", err)
		}
		doc = flatDoc
	}

	return doc, oas.OpenAPIV2, nil
}

// loadOpenAPIV31Schema loads an OpenAPI 3.1 schema
func loadOpenAPIV31Schema(data []byte, filePath string, opts LoadOptions) (*openapi3.T, OpenAPIVersion, error) {
	// Placeholder for future OpenAPI 3.1 support
	return nil, oas.OpenAPIV31, fmt.Errorf("OpenAPI 3.1 support not yet implemented")
}

// parseYAMLOrJSON parses the input as YAML or JSON
func parseYAMLOrJSON(data []byte, target interface{}) error {
	// Try JSON first
	if err := json.Unmarshal(data, target); err == nil {
		return nil
	}

	// If JSON fails, try YAML
	return yaml.Unmarshal(data, target)
}
