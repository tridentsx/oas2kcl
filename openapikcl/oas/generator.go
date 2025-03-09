// Package oas provides functionality for converting OpenAPI Schema to KCL.
package oas

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// formatSchemaName formats a schema name to be valid in KCL.
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

// extractSchemaName extracts the schema name from a reference string.
func extractSchemaName(ref string) string {
	// For "#/components/schemas/Pet" or "#/definitions/Pet", return "Pet"
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// writeKCLSchemaFile writes a KCL schema to a file.
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
		Swagger string `json:"swagger" yaml:"swagger"` // OpenAPI 2.0
		OpenAPI string `json:"openapi" yaml:"openapi"` // OpenAPI 3.x
	}

	// Try JSON first
	jsonErr := json.Unmarshal(data, &doc)
	if jsonErr == nil {
		return detectVersion(doc)
	}

	// If JSON parsing fails, try YAML
	yamlErr := yaml.Unmarshal(data, &doc)
	if yamlErr != nil {
		return "", fmt.Errorf("failed to parse document as JSON (%w) or YAML (%w)", jsonErr, yamlErr)
	}

	return detectVersion(doc)
}

// detectVersion determines the OpenAPI version from the parsed document
func detectVersion(doc struct {
	Swagger string `json:"swagger" yaml:"swagger"`
	OpenAPI string `json:"openapi" yaml:"openapi"`
}) (OpenAPIVersion, error) {
	if doc.Swagger == "2.0" {
		return OpenAPIV2, nil
	} else if doc.OpenAPI == "3.0.0" || doc.OpenAPI == "3.0.1" || doc.OpenAPI == "3.0.2" || doc.OpenAPI == "3.0.3" {
		return OpenAPIV3, nil
	} else if strings.HasPrefix(doc.OpenAPI, "3.1") {
		return OpenAPIV31, nil
	}

	return "", fmt.Errorf("unsupported OpenAPI version: swagger=%q, openapi=%q", doc.Swagger, doc.OpenAPI)
}

// IsSwaggerVersion returns true if the given version is OpenAPI 2.0 (Swagger).
func IsSwaggerVersion(version OpenAPIVersion) bool {
	return version == OpenAPIV2
}

// GenerateSchemas converts an OpenAPI schema to KCL schemas and writes them to the output directory.
func GenerateSchemas(doc *openapi3.T, outputDir string, packageName string, version OpenAPIVersion) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Collect all schema names
	var schemas []string
	if doc.Components != nil && doc.Components.Schemas != nil {
		schemas = collectSchemas(doc.Components.Schemas)
	}

	// Check if we have any schemas to generate
	if len(schemas) == 0 {
		log.Printf("no schemas found in OpenAPI document")
		return nil
	}

	// Generate KCL schemas for each schema in the OpenAPI document
	for _, schemaName := range schemas {
		schemaRef := doc.Components.Schemas[schemaName]
		kclSchema, err := GenerateKCLSchema(schemaName, schemaRef, doc.Components.Schemas, version, doc)
		if err != nil {
			return fmt.Errorf("failed to generate KCL schema for %s: %w", schemaName, err)
		}

		// Write the KCL schema to a file
		if err := writeKCLSchemaFile(outputDir, schemaName, kclSchema); err != nil {
			return fmt.Errorf("failed to write KCL schema file for %s: %w", schemaName, err)
		}
	}

	// Generate a main.k file that imports all schemas
	if err := generateMainFile(outputDir, packageName, schemas); err != nil {
		return fmt.Errorf("failed to generate main.k file: %w", err)
	}

	log.Printf("successfully generated %d KCL schemas", len(schemas))
	return nil
}

// collectSchemas collects all schema names from an OpenAPI schema collection.
func collectSchemas(schemas openapi3.Schemas) []string {
	var schemaNames []string
	for name := range schemas {
		schemaNames = append(schemaNames, name)
	}
	return schemaNames
}

// GenerateKCLSchema generates a KCL schema from an OpenAPI schema reference.
func GenerateKCLSchema(name string, schema *openapi3.SchemaRef, allSchemas openapi3.Schemas, version OpenAPIVersion, doc *openapi3.T) (string, error) {
	log.Printf("generating KCL schema for %s", name)

	// Format the schema name to be valid in KCL
	formattedName := formatSchemaName(name)

	var sb strings.Builder
	needsRegexImport := false

	// Check if the schema includes formats that require regex
	needsRegexImport = checkIfNeedsRegexImport(schema)

	// Add necessary imports
	if needsRegexImport {
		sb.WriteString("import regex\n\n")
	} else {
		sb.WriteString("# No imports needed\n\n")
	}

	// Start schema definition
	sb.WriteString(fmt.Sprintf("schema %s:\n", formattedName))

	// Add schema documentation if available
	if schema.Value != nil && (schema.Value.Title != "" || schema.Value.Description != "") {
		if schema.Value.Title != "" {
			sb.WriteString(fmt.Sprintf("    # %s\n", schema.Value.Title))
		}
		if schema.Value.Description != "" {
			description := schema.Value.Description
			lines := strings.Split(description, "\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("    # %s\n", line))
			}
		}
	}

	// Process inheritance (allOf)
	parents, err := processInheritance(schema.Value, allSchemas)
	if err != nil {
		return "", fmt.Errorf("failed to process inheritance: %w", err)
	}

	// Format parent names to be valid in KCL
	var formattedParents []string
	for _, parent := range parents {
		formattedParents = append(formattedParents, formatSchemaName(parent))
	}

	// Add mixin if there are parent schemas
	if len(formattedParents) > 0 {
		sb.WriteString(fmt.Sprintf("\n    mixin [%s]", strings.Join(formattedParents, ", ")))
	}

	// Add properties
	var requiredProps []string
	if schema.Value != nil && schema.Value.Required != nil {
		requiredProps = schema.Value.Required
	}

	var constraints []string
	var properties []*openapi3.SchemaRef

	// Collect properties from the schema
	if schema.Value != nil && schema.Value.Properties != nil {
		for propName, propSchema := range schema.Value.Properties {
			properties = append(properties, propSchema)

			// Determine if the property is required
			isRequired := false
			for _, req := range requiredProps {
				if req == propName {
					isRequired = true
					break
				}
			}

			// Get KCL field definition
			fieldType, documentation := generateFieldType(propName, propSchema, isRequired, formattedName, doc)

			// Add the field definition
			sb.WriteString(fmt.Sprintf("\n    %s%s", documentation, fieldType))

			// Add constraints for this field
			fieldConstraints := generateConstraints(propSchema.Value, propName, false)
			if len(fieldConstraints) > 0 {
				constraints = append(constraints, fieldConstraints...)
			}
		}
	}

	// If no properties, add a placeholder comment
	if len(properties) == 0 {
		sb.WriteString("\n    # No properties defined")
	}

	// Add constraints section if we have any
	if len(constraints) > 0 {
		sb.WriteString("\n\n    check:")
		for _, constraint := range constraints {
			sb.WriteString(fmt.Sprintf("\n        %s", constraint))
		}
	}

	return sb.String(), nil
}

// generateMainFile generates a main.k file that imports all schemas.
func generateMainFile(outputDir string, packageName string, schemas []string) error {
	log.Printf("generating main.k file with %d schema imports", len(schemas))

	// Create the main.k file
	mainFilePath := filepath.Join(outputDir, "main.k")
	var mainFile strings.Builder

	// Write package comment
	mainFile.WriteString(fmt.Sprintf("# KCL schemas generated from %s\n\n", packageName))

	// Import all schemas
	if len(schemas) > 0 {
		for _, schema := range schemas {
			mainFile.WriteString(fmt.Sprintf("import %s\n", schema))
		}
		mainFile.WriteString("\n")
	} else {
		mainFile.WriteString("# No schemas to import\n\n")
	}

	// Add validation schema
	mainFile.WriteString("schema ValidationSchema:\n")
	mainFile.WriteString("    # This schema can be used to validate instances\n")
	mainFile.WriteString("    # Example: myInstance: SomeSchema\n")
	mainFile.WriteString("    _ignore?: bool = True # Empty schema\n")

	// Write the main file
	return os.WriteFile(mainFilePath, []byte(mainFile.String()), 0644)
}

// processInheritance processes inheritance (allOf) in an OpenAPI schema.
func processInheritance(schema *openapi3.Schema, allSchemas openapi3.Schemas) ([]string, error) {
	var parents []string

	// Check if the schema has allOf
	if schema == nil || schema.AllOf == nil {
		return parents, nil
	}

	// Process each schema in allOf
	for _, subSchema := range schema.AllOf {
		// Check if it's a reference to another schema
		if subSchema.Ref != "" {
			// Extract schema name from reference
			refParts := strings.Split(subSchema.Ref, "/")
			parentName := refParts[len(refParts)-1]

			// Add the parent to the list
			parents = append(parents, parentName)
		}
	}

	return parents, nil
}

// generateFieldType generates a KCL field type from an OpenAPI schema.
func generateFieldType(fieldName string, fieldSchema *openapi3.SchemaRef, isRequired bool, schemaName string, doc *openapi3.T) (string, string) {
	optionalMarker := "?"
	if isRequired {
		optionalMarker = ""
	}

	// Get the field documentation
	documentation := ""
	if fieldSchema.Value != nil && fieldSchema.Value.Description != "" {
		documentation = fmt.Sprintf("# %s\n    ", fieldSchema.Value.Description)
	}

	// Handle reference to another schema
	if fieldSchema.Ref != "" {
		referencedType := extractSchemaName(fieldSchema.Ref)
		return fmt.Sprintf("%s%s: %s", fieldName, optionalMarker, formatSchemaName(referencedType)), documentation
	}

	// If no schema value, default to string
	if fieldSchema.Value == nil {
		return fmt.Sprintf("%s%s: str", fieldName, optionalMarker), documentation
	}

	// Handle array type
	if fieldSchema.Value.Type != nil && (*fieldSchema.Value.Type)[0] == "array" {
		// Get array item type
		if fieldSchema.Value.Items != nil {
			if fieldSchema.Value.Items.Ref != "" {
				// Array of references
				itemType := extractSchemaName(fieldSchema.Value.Items.Ref)
				return fmt.Sprintf("%s%s: [%s]", fieldName, optionalMarker, formatSchemaName(itemType)), documentation
			} else if fieldSchema.Value.Items.Value != nil && fieldSchema.Value.Items.Value.Type != nil {
				// Array of primitive types
				itemType := (*fieldSchema.Value.Items.Value.Type)[0]
				kclType := mapOpenAPITypeToKCL(itemType)
				return fmt.Sprintf("%s%s: [%s]", fieldName, optionalMarker, kclType), documentation
			}
		}
		// Default array of any
		return fmt.Sprintf("%s%s: [any]", fieldName, optionalMarker), documentation
	}

	// Handle object type
	if fieldSchema.Value.Type != nil && (*fieldSchema.Value.Type)[0] == "object" {
		// Inline object definition
		if fieldSchema.Value.Properties != nil && len(fieldSchema.Value.Properties) > 0 {
			// This is a complex inline object that would be better as a separate schema
			nestedName := formatSchemaName(schemaName + "_" + fieldName)
			return fmt.Sprintf("%s%s: %s", fieldName, optionalMarker, nestedName), documentation
		}
		// Simple object or empty object
		return fmt.Sprintf("%s%s: {str:any}", fieldName, optionalMarker), documentation
	}

	// Handle oneOf, anyOf, allOf
	if fieldSchema.Value.OneOf != nil && len(fieldSchema.Value.OneOf) > 0 {
		return fmt.Sprintf("%s%s: any", fieldName, optionalMarker), documentation
	}
	if fieldSchema.Value.AnyOf != nil && len(fieldSchema.Value.AnyOf) > 0 {
		return fmt.Sprintf("%s%s: any", fieldName, optionalMarker), documentation
	}
	if fieldSchema.Value.AllOf != nil && len(fieldSchema.Value.AllOf) > 0 {
		return fmt.Sprintf("%s%s: any", fieldName, optionalMarker), documentation
	}

	// Handle primitive types
	if fieldSchema.Value.Type != nil {
		typeStr := (*fieldSchema.Value.Type)[0]
		kclType := mapOpenAPITypeToKCL(typeStr)
		return fmt.Sprintf("%s%s: %s", fieldName, optionalMarker, kclType), documentation
	}

	// Default to any
	return fmt.Sprintf("%s%s: any", fieldName, optionalMarker), documentation
}

// mapOpenAPITypeToKCL maps an OpenAPI type to a KCL type.
func mapOpenAPITypeToKCL(openAPIType string) string {
	switch openAPIType {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "object":
		return "{str:any}"
	case "array":
		return "[any]"
	case "null":
		return "None"
	default:
		return "any"
	}
}

// generateConstraints generates KCL constraints for an OpenAPI schema.
func generateConstraints(schema *openapi3.Schema, fieldName string, useSelfPrefix bool) []string {
	var constraints []string
	if schema == nil {
		return constraints
	}

	// Determine the field accessor prefix
	prefix := ""
	if useSelfPrefix {
		prefix = "self."
	}
	fieldAccess := prefix + fieldName

	// Handle string constraints
	if schema.Type != nil && len(*schema.Type) > 0 && (*schema.Type)[0] == "string" {
		// MinLength constraint
		if schema.MinLength > 0 {
			constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", fieldAccess, schema.MinLength))
		}

		// MaxLength constraint
		if schema.MaxLength != nil {
			constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", fieldAccess, *schema.MaxLength))
		}

		// Pattern constraint
		if schema.Pattern != "" {
			// Escape pattern for KCL
			pattern := strings.ReplaceAll(schema.Pattern, "\\", "\\\\")
			constraints = append(constraints, fmt.Sprintf("regex.match(%s, r\"%s\")", fieldAccess, pattern))
		}

		// Format constraint
		if schema.Format != "" {
			switch schema.Format {
			case "email":
				emailPattern := `r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"`
				constraints = append(constraints, fmt.Sprintf("regex.match(%s, %s)", fieldAccess, emailPattern))
			case "date-time":
				dateTimePattern := `r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$"`
				constraints = append(constraints, fmt.Sprintf("regex.match(%s, %s)", fieldAccess, dateTimePattern))
			case "date":
				datePattern := `r"^\d{4}-\d{2}-\d{2}$"`
				constraints = append(constraints, fmt.Sprintf("regex.match(%s, %s)", fieldAccess, datePattern))
			case "time":
				timePattern := `r"^\d{2}:\d{2}:\d{2}(\.\d+)?$"`
				constraints = append(constraints, fmt.Sprintf("regex.match(%s, %s)", fieldAccess, timePattern))
			case "uuid":
				uuidPattern := `r"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"`
				constraints = append(constraints, fmt.Sprintf("regex.match(%s, %s)", fieldAccess, uuidPattern))
			case "uri":
				uriPattern := `r"^(https?|ftp)://[^\s/$.?#].[^\s]*$"`
				constraints = append(constraints, fmt.Sprintf("regex.match(%s, %s)", fieldAccess, uriPattern))
			}
		}

		// Enum constraint
		if schema.Enum != nil && len(schema.Enum) > 0 {
			values := make([]string, 0, len(schema.Enum))
			for _, e := range schema.Enum {
				if s, ok := e.(string); ok {
					values = append(values, fmt.Sprintf("\"%s\"", s))
				}
			}
			if len(values) > 0 {
				constraints = append(constraints, fmt.Sprintf("%s in [%s]", fieldAccess, strings.Join(values, ", ")))
			}
		}
	}

	// Handle numeric constraints (integer and number)
	if schema.Type != nil && len(*schema.Type) > 0 && ((*schema.Type)[0] == "integer" || (*schema.Type)[0] == "number") {
		// Minimum constraint
		if schema.Min != nil {
			if schema.ExclusiveMin {
				constraints = append(constraints, fmt.Sprintf("%s > %g", fieldAccess, *schema.Min))
			} else {
				constraints = append(constraints, fmt.Sprintf("%s >= %g", fieldAccess, *schema.Min))
			}
		}

		// Maximum constraint
		if schema.Max != nil {
			if schema.ExclusiveMax {
				constraints = append(constraints, fmt.Sprintf("%s < %g", fieldAccess, *schema.Max))
			} else {
				constraints = append(constraints, fmt.Sprintf("%s <= %g", fieldAccess, *schema.Max))
			}
		}

		// MultipleOf constraint
		if schema.MultipleOf != nil {
			// KCL doesn't have a direct way to check multiple of, so we use modulo
			if (*schema.Type)[0] == "integer" {
				constraints = append(constraints, fmt.Sprintf("%s %% %d == 0", fieldAccess, int(*schema.MultipleOf)))
			} else {
				// For floating point, we approximate using a small epsilon
				constraints = append(constraints, fmt.Sprintf("abs(%s %% %g) < 0.0000001", fieldAccess, *schema.MultipleOf))
			}
		}
	}

	// Handle array constraints
	if schema.Type != nil && len(*schema.Type) > 0 && (*schema.Type)[0] == "array" {
		// MinItems constraint
		if schema.MinItems > 0 {
			constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", fieldAccess, schema.MinItems))
		}

		// MaxItems constraint
		if schema.MaxItems != nil {
			constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", fieldAccess, *schema.MaxItems))
		}

		// UniqueItems constraint
		if schema.UniqueItems {
			constraints = append(constraints, fmt.Sprintf("len(%s) == len(unique(%s))", fieldAccess, fieldAccess))
		}
	}

	return constraints
}

// checkIfNeedsRegexImport checks if the schema requires the regex import.
func checkIfNeedsRegexImport(schema *openapi3.SchemaRef) bool {
	return checkIfNeedsRegexImportWithDepth(schema, 0)
}

// checkIfNeedsRegexImportWithDepth checks if the schema requires the regex import with a depth limit.
func checkIfNeedsRegexImportWithDepth(schema *openapi3.SchemaRef, depth int) bool {
	if schema == nil {
		return false
	}

	// Prevent stack overflow with a reasonable depth limit
	if depth > 10 {
		return false
	}

	// Check if this schema has a pattern
	if schema.Value != nil && schema.Value.Pattern != "" {
		return true
	}

	// Check if this schema has a format that requires regex
	if schema.Value != nil && schema.Value.Format != "" {
		switch schema.Value.Format {
		case "email", "date-time", "date", "time", "uuid", "uri", "hostname", "ipv4", "ipv6":
			return true
		}
	}

	// Check properties
	if schema.Value != nil && schema.Value.Properties != nil {
		for _, propSchema := range schema.Value.Properties {
			if checkIfNeedsRegexImportWithDepth(propSchema, depth+1) {
				return true
			}
		}
	}

	// Check array items
	if schema.Value != nil && schema.Value.Items != nil {
		if checkIfNeedsRegexImportWithDepth(schema.Value.Items, depth+1) {
			return true
		}
	}

	// Check oneOf, anyOf, allOf
	if schema.Value != nil {
		if schema.Value.OneOf != nil {
			for _, s := range schema.Value.OneOf {
				if checkIfNeedsRegexImportWithDepth(s, depth+1) {
					return true
				}
			}
		}

		if schema.Value.AnyOf != nil {
			for _, s := range schema.Value.AnyOf {
				if checkIfNeedsRegexImportWithDepth(s, depth+1) {
					return true
				}
			}
		}

		if schema.Value.AllOf != nil {
			for _, s := range schema.Value.AllOf {
				if checkIfNeedsRegexImportWithDepth(s, depth+1) {
					return true
				}
			}
		}
	}

	return false
}

// ConvertTypeToKCL converts an OpenAPI type to a KCL type
func ConvertTypeToKCL(oapiType, format string) string {
	if oapiType == "" {
		return "any"
	}

	// Handle primitive types
	switch oapiType {
	case "string":
		// Handle string formats
		switch format {
		case "byte", "binary":
			return "str"
		case "date", "date-time", "time", "email", "uuid", "uri":
			return "str"
		case "password":
			return "str"
		default:
			return "str"
		}
	case "number":
		switch format {
		case "float", "double":
			return "float"
		default:
			return "float"
		}
	case "integer":
		switch format {
		case "int32", "int64":
			return "int"
		default:
			return "int"
		}
	case "boolean":
		return "bool"
	case "null":
		return "None"
	case "array":
		return "[any]" // This should be refined when handling arrays
	case "object":
		return "{str:any}" // This should be refined when handling objects
	default:
		// Return a generic type for unknown types
		return "any"
	}
}

// GenerateConstraints generates KCL constraints from an OpenAPI schema
func GenerateConstraints(schema *openapi3.Schema, fieldName string, useSelfPrefix bool) []string {
	var constraints []string
	var prefix string
	if useSelfPrefix {
		prefix = "self"
	} else {
		prefix = fieldName
	}

	// Handle string constraints
	if schema.Type != nil && len(*schema.Type) > 0 && (*schema.Type)[0] == "string" {
		// MinLength constraint
		if schema.MinLength > 0 {
			constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", prefix, schema.MinLength))
		}

		// MaxLength constraint
		if schema.MaxLength != nil {
			constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", prefix, *schema.MaxLength))
		}

		// Pattern constraint
		if schema.Pattern != "" {
			// Escape backslashes for KCL's raw string format
			pattern := strings.ReplaceAll(schema.Pattern, "\\", "\\\\")
			constraints = append(constraints, fmt.Sprintf("regex.match(%s, r\"%s\")", prefix, pattern))
		}

		// Enum constraint
		if len(schema.Enum) > 0 {
			var values []string
			for _, e := range schema.Enum {
				if s, ok := e.(string); ok {
					values = append(values, fmt.Sprintf("\"%s\"", s))
				}
			}
			if len(values) > 0 {
				constraints = append(constraints, fmt.Sprintf("%s in [%s]", prefix, strings.Join(values, ", ")))
			}
		}
	}

	// Handle numeric constraints
	if schema.Type != nil && len(*schema.Type) > 0 && ((*schema.Type)[0] == "integer" || (*schema.Type)[0] == "number") {
		// Minimum constraint
		if schema.Min != nil {
			if schema.ExclusiveMin {
				constraints = append(constraints, fmt.Sprintf("%s > %v", prefix, *schema.Min))
			} else {
				constraints = append(constraints, fmt.Sprintf("%s >= %v", prefix, *schema.Min))
			}
		}

		// Maximum constraint
		if schema.Max != nil {
			if schema.ExclusiveMax {
				constraints = append(constraints, fmt.Sprintf("%s < %v", prefix, *schema.Max))
			} else {
				constraints = append(constraints, fmt.Sprintf("%s <= %v", prefix, *schema.Max))
			}
		}

		// MultipleOf constraint
		if schema.MultipleOf != nil {
			constraints = append(constraints, fmt.Sprintf("%s %% %v == 0", prefix, *schema.MultipleOf))
		}
	}

	// Handle array constraints
	if schema.Type != nil && len(*schema.Type) > 0 && (*schema.Type)[0] == "array" {
		// MinItems constraint
		if schema.MinItems > 0 {
			constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", prefix, schema.MinItems))
		}

		// MaxItems constraint
		if schema.MaxItems != nil {
			constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", prefix, *schema.MaxItems))
		}

		// UniqueItems constraint
		if schema.UniqueItems {
			constraints = append(constraints, fmt.Sprintf("len(%s) == len(unique(%s))", prefix, prefix))
		}
	}

	return constraints
}

// FormatDocumentation formats the documentation from an OpenAPI schema
func FormatDocumentation(schema *openapi3.Schema) string {
	if schema == nil || (schema.Title == "" && schema.Description == "") {
		return ""
	}

	var sb strings.Builder

	// Add title as first line if present
	if schema.Title != "" {
		sb.WriteString(schema.Title)
		if schema.Description != "" {
			sb.WriteString("\n")
		}
	}

	// Add description
	if schema.Description != "" {
		// Split description into multiple lines if needed
		lines := strings.Split(schema.Description, "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(line)
			}
		}
	}

	return sb.String()
}

// More functions from generator_oas.go should be moved here
