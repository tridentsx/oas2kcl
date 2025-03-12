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
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema"
	"github.com/tridentsx/oas2kcl/openapikcl/oas"
	"gopkg.in/yaml.v2"
)

var debugMode bool

// SpecFormat defines the format of a specification
type SpecFormat int

const (
	// SpecFormatUnknown represents an unknown specification format
	SpecFormatUnknown SpecFormat = iota
	// SpecFormatOpenAPIV2 represents OpenAPI 2.0 (Swagger)
	SpecFormatOpenAPIV2
	// SpecFormatOpenAPIV3 represents OpenAPI 3.0
	SpecFormatOpenAPIV3
	// SpecFormatOpenAPIV31 represents OpenAPI 3.1
	SpecFormatOpenAPIV31
	// SpecFormatJSONSchema represents JSON Schema
	SpecFormatJSONSchema
)

// String returns a string representation of the SpecFormat
func (s SpecFormat) String() string {
	switch s {
	case SpecFormatOpenAPIV2:
		return "OpenAPI 2.0"
	case SpecFormatOpenAPIV3:
		return "OpenAPI 3.0"
	case SpecFormatOpenAPIV31:
		return "OpenAPI 3.1"
	case SpecFormatJSONSchema:
		return "JSON Schema"
	default:
		return "Unknown"
	}
}

// GenerateKCLSchemas generates KCL schemas from an OpenAPI document
func GenerateKCLSchemas(doc *openapi3.T, outputDir string, packageName string, version OpenAPIVersion, jsonSchemaPayload []byte) error {
	log.Printf("Starting KCL schema generation")

	// Check if we have an OpenAPI document
	if doc != nil {
		// We have an OpenAPI document, use the oas package
		return oas.GenerateSchemas(doc, outputDir, packageName, version)
	}

	// Check if we have JSON Schema data
	if jsonSchemaPayload != nil && len(jsonSchemaPayload) > 0 {
		// We have JSON Schema data, use the jsonschema package
		return jsonschema.GenerateSchemas(jsonSchemaPayload, outputDir, packageName)
	}

	return fmt.Errorf("no valid schema provided")
}

// formatSchemaName ensures the schema name is properly formatted for KCL
func formatSchemaName(name string) string {
	// If name is empty, return empty
	if name == "" {
		return ""
	}

	// Replace spaces and special characters with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove any non-alphanumeric characters except underscores
	var result strings.Builder
	for i, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			if i == 0 && unicode.IsDigit(r) {
				// KCL identifiers can't start with a digit, so prefix with underscore
				result.WriteRune('_')
			}
			result.WriteRune(r)
		}
	}

	name = result.String()

	// Ensure the first character is uppercase (PascalCase)
	runes := []rune(name)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
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

// GenerateKCL generates KCL schemas from an OpenAPI or JSON Schema file
func GenerateKCL(schemaFile string, outputDir string, packageName string, debugMode bool) error {
	// Read the schema file
	schemaData, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	// Detect the specification format
	format, err := DetectSpecFormat(schemaData)
	if err != nil {
		return fmt.Errorf("error detecting specification format: %w", err)
	}

	// Process based on the detected format
	switch format {
	case SpecFormatJSONSchema:
		log.Printf("Detected JSON Schema format, processing with tree-based JSON Schema generator")
		return jsonschema.GenerateSchemaTreeAndKCL(schemaData, outputDir, debugMode)

	case SpecFormatOpenAPIV2, SpecFormatOpenAPIV3, SpecFormatOpenAPIV31:
		log.Printf("Processing OpenAPI format (%s) with tree-based approach", format)

		// Try to parse as OpenAPI
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(schemaData)
		if err != nil {
			return fmt.Errorf("error parsing OpenAPI document: %w", err)
		}

		// Determine the OpenAPI version
		version, err := oas.DetectOpenAPIVersion(schemaData)
		if err != nil {
			return fmt.Errorf("error detecting OpenAPI version: %w", err)
		}

		// Extract JSON schemas from OpenAPI and convert to tree-based approach
		// For now, delegate to the standard OpenAPI processor
		// TODO: Implement tree-based approach for OpenAPI schemas
		return oas.GenerateSchemas(doc, outputDir, packageName, version)

	default:
		return fmt.Errorf("unsupported specification format")
	}
}

// DetectSpecFormat detects the format of the specification from raw data.
func DetectSpecFormat(data []byte) (SpecFormat, error) {
	// First, check if it's an OpenAPI specification by trying to detect the version
	openAPIVersion, err := oas.DetectOpenAPIVersion(data)
	if err == nil {
		// It's an OpenAPI spec, determine which version
		switch openAPIVersion {
		case oas.OpenAPIV2:
			return SpecFormatOpenAPIV2, nil
		case oas.OpenAPIV3:
			return SpecFormatOpenAPIV3, nil
		case oas.OpenAPIV31:
			return SpecFormatOpenAPIV31, nil
		}
	}

	// If it's not recognized as an OpenAPI spec, check for JSON schema
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal(data, &jsonSchema); err != nil {
		return SpecFormatUnknown, fmt.Errorf("failed to parse as JSON: %w", err)
	}

	// Check for JSON Schema specific patterns
	if schemaURI, ok := jsonSchema["$schema"].(string); ok {
		if strings.Contains(schemaURI, "json-schema.org") {
			return SpecFormatJSONSchema, nil
		}
	}

	// Check for common JSON Schema properties
	if _, ok := jsonSchema["properties"]; ok {
		if _, ok := jsonSchema["type"]; ok {
			if typ, ok := jsonSchema["type"].(string); ok && typ == "object" {
				return SpecFormatJSONSchema, nil
			}
		}
	}

	return SpecFormatUnknown, fmt.Errorf("unrecognized specification format")
}

// ParseSpecYAMLorJSON parses the spec as either YAML or JSON.
func ParseSpecYAMLorJSON(content []byte, out interface{}) error {
	if err := yaml.Unmarshal(content, out); err != nil {
		// Try JSON if YAML parsing fails
		if err := json.Unmarshal(content, out); err != nil {
			return fmt.Errorf("failed to parse as YAML or JSON: %w", err)
		}
	}
	return nil
}

// GenerateTestCaseOutput generates KCL schemas for a specific test case directory
func GenerateTestCaseOutput(testCaseDir string) error {
	// Check if the test case directory exists
	if _, err := os.Stat(testCaseDir); os.IsNotExist(err) {
		return fmt.Errorf("test case directory %s does not exist", testCaseDir)
	}

	// Determine the schema file path based on the directory
	var schemaFile string
	schemaFormats := []string{".json", ".yaml", ".yml"}
	for _, format := range schemaFormats {
		candidate := filepath.Join(testCaseDir, "schema"+format)
		if _, err := os.Stat(candidate); err == nil {
			schemaFile = candidate
			break
		}
	}

	if schemaFile == "" {
		return fmt.Errorf("no schema file found in test case directory %s", testCaseDir)
	}

	// Create the output directory
	outputDir := filepath.Join(testCaseDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get a simple package name from the directory name
	packageName := filepath.Base(testCaseDir)

	// Generate KCL schemas (no debug mode for test cases)
	return GenerateKCL(schemaFile, outputDir, packageName, false)
}
