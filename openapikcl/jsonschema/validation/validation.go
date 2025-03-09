// Package validation provides validation-related functionality for JSON Schema to KCL conversion.
package validation

import (
	"fmt"
	"strings"

	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/types"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/utils"
)

// GenerateConstraints generates KCL constraints for a property
func GenerateConstraints(propSchema map[string]interface{}, propName string) string {
	constraints := []string{}
	indent := "    "
	schemaType, _ := types.GetSchemaType(propSchema)

	sanitizedName := utils.SanitizePropertyName(propName)

	// String constraints
	if schemaType == "string" {
		// Min/max length
		if minLength, ok := utils.GetIntValue(propSchema, "minLength"); ok {
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or len(%s) >= %d", indent, sanitizedName, sanitizedName, minLength))
		}
		if maxLength, ok := utils.GetIntValue(propSchema, "maxLength"); ok {
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or len(%s) <= %d", indent, sanitizedName, sanitizedName, maxLength))
		}

		// Pattern
		if pattern, ok := utils.GetStringValue(propSchema, "pattern"); ok {
			constraints = append(constraints, fmt.Sprintf("%s# Regex pattern: %s", indent, pattern))
			constraints = append(constraints, fmt.Sprintf("%s# KCL doesn't directly support regex validation without imports", indent))
		}

		// Format
		if format, ok := utils.GetStringValue(propSchema, "format"); ok {
			switch format {
			case "email":
				constraints = append(constraints, fmt.Sprintf("%s# TODO: Add email validation for %s", indent, sanitizedName))
			case "date-time":
				constraints = append(constraints, fmt.Sprintf("%s# TODO: Add date-time validation for %s", indent, sanitizedName))
			}
		}

		// Enum
		if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
			enumStr := formatEnumValues(enumValues)
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or %s in [%s]", indent, sanitizedName, sanitizedName, enumStr))
		}
	}

	// Number/integer constraints
	if schemaType == "number" || schemaType == "integer" {
		// Minimum/maximum
		if minimum, ok := utils.GetFloatValue(propSchema, "minimum"); ok {
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or %s >= %v", indent, sanitizedName, sanitizedName, minimum))
		}
		if maximum, ok := utils.GetFloatValue(propSchema, "maximum"); ok {
			if exclusiveMax, _ := utils.GetBoolValue(propSchema, "exclusiveMaximum"); exclusiveMax {
				constraints = append(constraints, fmt.Sprintf("%scheck %s == None or %s < %v", indent, sanitizedName, sanitizedName, maximum))
			} else {
				constraints = append(constraints, fmt.Sprintf("%scheck %s == None or %s <= %v", indent, sanitizedName, sanitizedName, maximum))
			}
		}

		// ExclusiveMinimum/ExclusiveMaximum as separate properties
		if exclusiveMin, ok := utils.GetFloatValue(propSchema, "exclusiveMinimum"); ok {
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or %s > %v", indent, sanitizedName, sanitizedName, exclusiveMin))
		}
		if exclusiveMax, ok := utils.GetFloatValue(propSchema, "exclusiveMaximum"); ok {
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or %s < %v", indent, sanitizedName, sanitizedName, exclusiveMax))
		}

		// MultipleOf
		if multipleOf, ok := utils.GetFloatValue(propSchema, "multipleOf"); ok {
			constraints = append(constraints, fmt.Sprintf("%s# Value should be multiple of %v", indent, multipleOf))
			constraints = append(constraints, fmt.Sprintf("%s# KCL doesn't directly support 'multipleOf' checks", indent))
		}
	}

	// Array constraints
	if schemaType == "array" {
		// MinItems/MaxItems
		if minItems, ok := utils.GetIntValue(propSchema, "minItems"); ok {
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or len(%s) >= %d", indent, sanitizedName, sanitizedName, minItems))
		}
		if maxItems, ok := utils.GetIntValue(propSchema, "maxItems"); ok {
			constraints = append(constraints, fmt.Sprintf("%scheck %s == None or len(%s) <= %d", indent, sanitizedName, sanitizedName, maxItems))
		}

		// UniqueItems
		if uniqueItems, ok := utils.GetBoolValue(propSchema, "uniqueItems"); ok && uniqueItems {
			constraints = append(constraints, fmt.Sprintf("%s# All items in array should be unique", indent))
			constraints = append(constraints, fmt.Sprintf("%s# KCL doesn't directly support unique items check", indent))
		}
	}

	if len(constraints) > 0 {
		return "\n" + strings.Join(constraints, "\n")
	}

	return ""
}

// formatEnumValues formats enum values for use in a KCL constraint
func formatEnumValues(values []interface{}) string {
	parts := make([]string, 0, len(values))
	for _, val := range values {
		parts = append(parts, utils.FormatLiteral(val))
	}
	return strings.Join(parts, ", ")
}

// CheckIfNeedsRegexImport checks if any property in the schema has a pattern constraint
func CheckIfNeedsRegexImport(rawSchema map[string]interface{}) bool {
	// Check if any string property has a pattern constraint
	properties, ok := utils.GetMapValue(rawSchema, "properties")
	if !ok {
		return false
	}

	for _, propValue := range properties {
		propSchema, ok := propValue.(map[string]interface{})
		if !ok {
			continue
		}

		schemaType, ok := types.GetSchemaType(propSchema)
		if !ok || schemaType != "string" {
			continue
		}

		if _, ok := utils.GetStringValue(propSchema, "pattern"); ok {
			return true
		}
	}

	return false
}

// GenerateRequiredPropertyChecks generates KCL validation checks for required properties
func GenerateRequiredPropertyChecks(schema map[string]interface{}) string {
	required, ok := utils.GetArrayValue(schema, "required")
	if !ok || len(required) == 0 {
		return ""
	}

	constraints := []string{}
	indent := "    "

	constraints = append(constraints, fmt.Sprintf("%scheck:", indent))

	for _, req := range required {
		reqStr, ok := req.(string)
		if !ok {
			continue
		}

		sanitizedName := utils.SanitizePropertyName(reqStr)
		constraints = append(constraints, fmt.Sprintf("%s    %s != None", indent, sanitizedName))
	}

	return strings.Join(constraints, "\n")
}
