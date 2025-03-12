// Package jsonschema implements the conversion of JSON Schema to KCL.
package jsonschema

import (
	"strings"

	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/types"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/utils"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/validation"
)

// Export helper functions from utility packages for backward compatibility

// GetStringValue safely extracts a string value from a map
var GetStringValue = utils.GetStringValue

// GetBoolValue safely extracts a boolean value from a map
var GetBoolValue = utils.GetBoolValue

// GetIntValue safely extracts an integer value from a map
var GetIntValue = utils.GetIntValue

// GetFloatValue safely extracts a float value from a map
var GetFloatValue = utils.GetFloatValue

// GetMapValue safely extracts a map value from a map
var GetMapValue = utils.GetMapValue

// GetArrayValue safely extracts an array value from a map
var GetArrayValue = utils.GetArrayValue

// GetKCLType returns the KCL type for a JSON Schema type
var GetKCLType = types.GetKCLType

// GetSchemaType extracts the type from a JSON Schema
var GetSchemaType = types.GetSchemaType

// FormatSchemaName formats a schema name to follow KCL naming conventions
var FormatSchemaName = types.FormatSchemaName

// ExtractSchemaName extracts a schema name from a reference
var ExtractSchemaName = types.ExtractSchemaName

// SanitizePropertyName ensures the property name is valid in KCL
var SanitizePropertyName = utils.SanitizePropertyName

// GenerateKCLFilePath generates a valid file path for a KCL schema
var GenerateKCLFilePath = utils.GenerateKCLFilePath

// FormatLiteral formats a Go value as a KCL literal
var FormatLiteral = utils.FormatLiteral

// IsPropertyRequired checks if a property is required in a JSON Schema
var IsPropertyRequired = types.IsPropertyRequired

// GetPropRawSchema gets the raw schema for a property
var GetPropRawSchema = types.GetPropRawSchema

// GenerateConstraints generates KCL constraints for a property
var GenerateConstraints = validation.GenerateConstraints

// Rename getKCLType to resolveKCLType to avoid redeclaration
func resolveKCLType(schema map[string]interface{}) string {
	if schema == nil {
		return "any"
	}

	// Check if this is a reference to another schema
	if ref, ok := schema["$ref"].(string); ok {
		refParts := strings.Split(ref, "/")
		return refParts[len(refParts)-1]
	}

	// Get the schema type
	schemaType, ok := schema["type"].(string)
	if !ok {
		return "any"
	}

	// Map JSON Schema types to KCL types
	switch schemaType {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "[]"
	case "object":
		return "dict"
	default:
		return "any"
	}
}

// For compatibility with existing tests
var checkIfNeedsRegexImport = validation.CheckIfNeedsRegexImport
