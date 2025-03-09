// Package jsonschema implements the conversion of JSON Schema to KCL.
package jsonschema

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
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

// For compatibility with existing tests
func getKCLType(schema *jsonschema.Schema, rawSchema map[string]interface{}) string {
	// This function is called from tests with both parameters, but our implementation only needs rawSchema
	return types.GetKCLType(rawSchema)
}

// For compatibility with existing tests
var checkIfNeedsRegexImport = validation.CheckIfNeedsRegexImport
