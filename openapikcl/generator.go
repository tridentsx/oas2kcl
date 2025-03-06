package openapikcl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

// Add debugMode variable if it doesn't exist
var debugMode bool

// SchemaType represents the type of schema being processed
type SchemaType int

const (
	SchemaTypeUnknown SchemaType = iota
	SchemaTypeOpenAPI2
	SchemaTypeOpenAPI3
	SchemaTypeOpenAPI31
	SchemaTypeJSONSchema
)

// String returns the string representation of the schema type
func (st SchemaType) String() string {
	switch st {
	case SchemaTypeOpenAPI2:
		return "OpenAPI 2.0 (Swagger)"
	case SchemaTypeOpenAPI3:
		return "OpenAPI 3.0"
	case SchemaTypeOpenAPI31:
		return "OpenAPI 3.1"
	case SchemaTypeJSONSchema:
		return "JSON Schema"
	default:
		return "Unknown"
	}
}

// DetectSchemaType determines the type of schema from a document
// This could be either an OpenAPI document or a JSON Schema document
func DetectSchemaType(doc *openapi3.T, rawSchema map[string]interface{}) SchemaType {
	// Check if it's an OpenAPI document
	if doc != nil {
		// Check for OpenAPI version
		if doc.OpenAPI != "" {
			// OpenAPI 3.x
			if strings.HasPrefix(doc.OpenAPI, "3.1") {
				return SchemaTypeOpenAPI31
			}
			return SchemaTypeOpenAPI3
		}

		// Check for Swagger version (OpenAPI 2.0)
		if rawSchema != nil {
			if _, hasSwagger := rawSchema["swagger"]; hasSwagger {
				return SchemaTypeOpenAPI2
			}
		}
	}

	// Check for JSON Schema
	if rawSchema != nil {
		if schemaURL, hasSchema := rawSchema["$schema"]; hasSchema {
			schemaStr, ok := schemaURL.(string)
			if ok && (strings.Contains(schemaStr, "json-schema.org") ||
				strings.Contains(schemaStr, "schema.json")) {
				return SchemaTypeJSONSchema
			}
		}
	}

	return SchemaTypeUnknown
}

// GenerateKCLSchemas generates KCL schemas from either an OpenAPI spec or a JSON Schema
func GenerateKCLSchemas(doc *openapi3.T, outputDir string, packageName string, version OpenAPIVersion, rawSchema map[string]interface{}) error {
	log.Printf("starting KCL schema generation")

	// Determine the schema type
	schemaType := DetectSchemaType(doc, rawSchema)
	log.Printf("detected schema type: %s", schemaType)

	// Handle different schema types
	switch schemaType {
	case SchemaTypeOpenAPI2, SchemaTypeOpenAPI3, SchemaTypeOpenAPI31:
		return generateOpenAPISchemas(doc, outputDir, packageName, version)
	case SchemaTypeJSONSchema:
		return generateJSONSchemas(rawSchema, outputDir, packageName)
	default:
		return fmt.Errorf("unable to determine schema type or unsupported schema type")
	}
}

// formatSchemaName ensures the schema name is properly formatted for KCL
func formatSchemaName(name string) string {
	// If name is empty, return empty
	if name == "" {
		return ""
	}

	// Ensure the first character is uppercase (PascalCase)
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// extractSchemaName extracts the schema name from a reference string
func extractSchemaName(ref string) string {
	// For "#/components/schemas/Pet" or "#/definitions/Pet", return "Pet"
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// writeKCLSchemaFile writes a KCL schema to a file
func writeKCLSchemaFile(outputDir, name, content string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the schema file
	schemaPath := filepath.Join(outputDir, name+".k")
	if err := os.WriteFile(schemaPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	log.Printf("Schema %s written to %s", name, schemaPath)
	return nil
}

// generateMainFile creates a main.k file that imports all schemas
func generateMainFile(outputDir string, packageName string, schemas []string) error {
	log.Printf("generating main.k file with %d schema imports", len(schemas))

	// Create the main.k file
	mainFilePath := filepath.Join(outputDir, "main.k")
	mainFile, err := os.Create(mainFilePath)
	if err != nil {
		return fmt.Errorf("failed to create main.k file: %w", err)
	}
	defer mainFile.Close()

	// Write package comment
	mainFile.WriteString(fmt.Sprintf("# KCL schemas generated from %s\n\n", packageName))

	// Create validation schema without importing regex
	mainFile.WriteString("# No imports needed for schemas in the same directory\n")
	mainFile.WriteString("# This avoids circular dependency issues\n\n")

	// Add validation schema
	mainFile.WriteString("schema ValidationSchema:\n")
	mainFile.WriteString("    # This schema can be used to validate instances\n")
	mainFile.WriteString("    # Example: myInstance: SomeSchema\n")
	mainFile.WriteString("    _ignore?: bool = True # Empty schema\n")

	return nil
}

// GenerateTestMainK creates a test-specific main.k file
func GenerateTestMainK(outputDir string, schemas []string) error {
	// Create the main.k file
	mainFilePath := filepath.Join(outputDir, "main.k")
	mainFile, err := os.Create(mainFilePath)
	if err != nil {
		return fmt.Errorf("failed to create main.k file: %w", err)
	}
	defer mainFile.Close()

	// Write header
	mainFile.WriteString("# Test validation schema\n\n")

	// No imports needed
	mainFile.WriteString("# No schema imports needed - schemas in same directory\n\n")

	// Create a simple test schema
	mainFile.WriteString("schema TestSchema:\n")
	mainFile.WriteString("    # A simple test schema\n")
	mainFile.WriteString("    name?: str\n")
	mainFile.WriteString("    value?: int\n\n")

	// Create a validation schema
	mainFile.WriteString("schema ValidationSchema:\n")
	mainFile.WriteString("    # Add test instances here\n")
	mainFile.WriteString("    test: TestSchema = {name: \"test\", value: 42}\n")

	// Add a comment about how to use the schemas
	if len(schemas) > 0 {
		mainFile.WriteString("\n# Available schemas: \n")
		for _, schema := range schemas {
			mainFile.WriteString(fmt.Sprintf("# - %s\n", schema))
		}
	}

	return nil
}

// SpecFormat represents the format of a specification
type SpecFormat int

const (
	UnknownSpec SpecFormat = iota
	OpenAPISpec
	JSONSchemaSpec
)

// detectSpecFormat identifies if a document is OpenAPI or JSON Schema
func detectSpecFormat(data []byte) SpecFormat {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return UnknownSpec
	}

	// Check for OpenAPI indicators
	if _, hasOpenAPI := doc["openapi"]; hasOpenAPI {
		return OpenAPISpec
	}
	if _, hasSwagger := doc["swagger"]; hasSwagger {
		return OpenAPISpec
	}

	// Check for JSON Schema indicators
	if schemaURL, hasSchema := doc["$schema"]; hasSchema {
		schemaStr, ok := schemaURL.(string)
		if ok && (strings.Contains(schemaStr, "json-schema.org") ||
			strings.Contains(schemaStr, "schema.json")) {
			return JSONSchemaSpec
		}
	}

	return UnknownSpec
}

// GenerateKCL is a convenience function to generate KCL schemas from a file
func GenerateKCL(schemaFilePath, outputDir, packageName string) error {
	// Read the schema file
	data, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Detect spec format
	format := detectSpecFormat(data)

	// Process based on format
	switch format {
	case OpenAPISpec:
		// Load as OpenAPI
		doc, version, err := LoadOpenAPISchema(schemaFilePath, LoadOptions{
			FlattenSpec: true,
			SkipRemote:  false,
			MaxDepth:    100,
		})
		if err != nil {
			return fmt.Errorf("failed to load OpenAPI schema: %w", err)
		}

		// Generate KCL from OpenAPI
		return GenerateKCLSchemas(doc, outputDir, packageName, version, nil)

	case JSONSchemaSpec:
		// Parse JSON Schema
		var rawSchema map[string]interface{}
		if err := json.Unmarshal(data, &rawSchema); err != nil {
			return fmt.Errorf("failed to parse JSON Schema: %w", err)
		}

		// Generate KCL from JSON Schema
		return generateJSONSchemas(rawSchema, outputDir, packageName)

	default:
		return fmt.Errorf("unknown or unsupported specification format")
	}
}
