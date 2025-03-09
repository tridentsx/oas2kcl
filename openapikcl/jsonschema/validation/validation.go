// Package validation provides validation-related functionality for JSON Schema to KCL conversion.
package validation

import (
	"fmt"
	"strings"

	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/templates"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/types"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/utils"
)

// GenerateConstraints generates KCL constraints for a property
func GenerateConstraints(propSchema map[string]interface{}, propName string) string {
	comments := []string{}
	indent := "    "
	schemaType, _ := types.GetSchemaType(propSchema)

	// sanitizedName not used since we're only generating comments
	// sanitizedName := utils.SanitizePropertyName(propName)

	// First check if there's a format template we can use
	if format, ok := utils.GetStringValue(propSchema, "format"); ok {
		template := templates.GetTemplateForFormat(format)
		if template != nil {
			// Return the template's comments
			comments = append(comments, template.GetComments(indent))
			return strings.Join(comments, "\n")
		}
	}

	// Check if we have a number template
	if schemaType == "number" || schemaType == "integer" {
		template := templates.GetTemplateForNumberType(propSchema, schemaType)
		if template != nil {
			comments = append(comments, template.GetComments(indent))
			return strings.Join(comments, "\n")
		}
	}

	// Check if we have an array template
	if schemaType == "array" {
		template := templates.GetTemplateForArrayType(propSchema)
		if template != nil {
			comments = append(comments, template.GetComments(indent))
			return strings.Join(comments, "\n")
		}
	}

	// String constraints
	if schemaType == "string" {
		// Min/max length
		if minLength, ok := utils.GetIntValue(propSchema, "minLength"); ok {
			comments = append(comments, indent+"# Min length: "+fmt.Sprintf("%d", minLength))
		}
		if maxLength, ok := utils.GetIntValue(propSchema, "maxLength"); ok {
			comments = append(comments, indent+"# Max length: "+fmt.Sprintf("%d", maxLength))
		}

		// Pattern
		if pattern, ok := utils.GetStringValue(propSchema, "pattern"); ok {
			comments = append(comments, indent+"# Regex pattern: "+pattern)
		}

		// Format - already handled at the top
		if format, ok := utils.GetStringValue(propSchema, "format"); ok {
			comments = append(comments, indent+"# Format: "+format)
		}

		// Enum
		if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
			enumStr := formatEnumValues(enumValues)
			comments = append(comments, indent+"# Allowed values: "+enumStr)
		}
	}

	// Number/integer constraints
	if schemaType == "number" || schemaType == "integer" {
		// No need for these now, they're handled by templates
	}

	// Array constraints
	if schemaType == "array" {
		// Min/max items
		if minItems, ok := utils.GetIntValue(propSchema, "minItems"); ok {
			comments = append(comments, indent+"# Min items: "+fmt.Sprintf("%d", minItems))
		}
		if maxItems, ok := utils.GetIntValue(propSchema, "maxItems"); ok {
			comments = append(comments, indent+"# Max items: "+fmt.Sprintf("%d", maxItems))
		}

		// Unique items
		if uniqueItems, ok := utils.GetBoolValue(propSchema, "uniqueItems"); ok && uniqueItems {
			comments = append(comments, indent+"# Unique items: true")
		}

		// Item constraints
		itemsSchema, ok := utils.GetMapValue(propSchema, "items")
		if ok {
			itemType, typeOk := types.GetSchemaType(itemsSchema)
			if typeOk && itemType == "string" {
				// Check for min/max length in items
				if minLength, ok := utils.GetIntValue(itemsSchema, "minLength"); ok {
					comments = append(comments, indent+"# Item min length: "+fmt.Sprintf("%d", minLength))
				}
				if maxLength, ok := utils.GetIntValue(itemsSchema, "maxLength"); ok {
					comments = append(comments, indent+"# Item max length: "+fmt.Sprintf("%d", maxLength))
				}

				// Pattern validation for items
				if pattern, ok := utils.GetStringValue(itemsSchema, "pattern"); ok {
					comments = append(comments, indent+"# Item pattern: "+pattern)
				}

				// Format validation for items
				if format, ok := utils.GetStringValue(itemsSchema, "format"); ok {
					comments = append(comments, indent+"# Item format: "+format)
				}

				// Enum for string items
				if enumValues, ok := utils.GetArrayValue(itemsSchema, "enum"); ok && len(enumValues) > 0 {
					enumStr := formatEnumValues(enumValues)
					comments = append(comments, indent+"# Item allowed values: "+enumStr)
				}
			}

			// For numeric item types, add min/max constraints
			if typeOk && (itemType == "number" || itemType == "integer") {
				// Min/max for numeric items
				if minimum, ok := utils.GetFloatValue(itemsSchema, "minimum"); ok {
					comments = append(comments, indent+"# Item minimum: "+fmt.Sprintf("%v", minimum))
				}
				if maximum, ok := utils.GetFloatValue(itemsSchema, "maximum"); ok {
					if exclusiveMax, _ := utils.GetBoolValue(itemsSchema, "exclusiveMaximum"); exclusiveMax {
						comments = append(comments, indent+"# Item exclusive maximum: "+fmt.Sprintf("%v", maximum))
					} else {
						comments = append(comments, indent+"# Item maximum: "+fmt.Sprintf("%v", maximum))
					}
				}
				if exclusiveMin, ok := utils.GetFloatValue(itemsSchema, "exclusiveMinimum"); ok {
					comments = append(comments, indent+"# Item exclusive minimum: "+fmt.Sprintf("%v", exclusiveMin))
				}

				// Enum for numeric items
				if enumValues, ok := utils.GetArrayValue(itemsSchema, "enum"); ok && len(enumValues) > 0 {
					enumStr := formatEnumValues(enumValues)
					comments = append(comments, indent+"# Item allowed values: "+enumStr)
				}
			}
		}
	}

	// Boolean constraints
	if schemaType == "boolean" {
		// Enum
		if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
			enumStr := formatEnumValues(enumValues)
			comments = append(comments, indent+"# Allowed values: "+enumStr)
		}
	}

	return strings.Join(comments, "\n")
}

// GenerateValidatorSchema generates a KCL validator schema with constraints in a check block
func GenerateValidatorSchema(schema map[string]interface{}, schemaName string) string {
	validatorName := schemaName + "Validator"
	validations := []string{}
	indent := "        "
	formats := make(map[string]bool)

	properties, hasProps := utils.GetMapValue(schema, "properties")
	if !hasProps || len(properties) == 0 {
		return ""
	}

	// Collect all the format types we need
	templateSchemas := make(map[string]string)
	numberSchemas := make(map[string]string)

	for propName, propValue := range properties {
		propSchema, ok := propValue.(map[string]interface{})
		if !ok {
			continue
		}

		// Skip if we don't have any constraints
		if !hasConstraints(propSchema) {
			continue
		}

		schemaType, _ := types.GetSchemaType(propSchema)
		sanitizedName := utils.SanitizePropertyName(propName)
		accessPath := "self." + sanitizedName

		// Check for format template first
		if format, ok := utils.GetStringValue(propSchema, "format"); ok {
			template := templates.GetTemplateForFormat(format)
			if template != nil {
				formats[format] = true
				validations = append(validations, template.GetValidation(sanitizedName, indent))

				// Store template schema content if it should be in a separate file
				if template.NeedsSeparateSchema() && template.GetSchemaContent() != "" {
					templateSchemas[format] = template.GetSchemaContent()
				}
				continue
			}
		}

		// Check for number template
		if schemaType == "number" || schemaType == "integer" {
			template := templates.GetTemplateForNumberType(propSchema, schemaType)
			if template != nil {
				validations = append(validations, template.GetValidation(sanitizedName, indent))

				// Store template schema content if needed
				if template.NeedsSeparateSchema() && template.GetSchemaContent() != "" {
					schemaKey := schemaType
					if schemaType == "number" {
						schemaKey = "Number"
					} else {
						schemaKey = "Integer"
					}

					// Only add if we don't already have this schema
					if _, exists := numberSchemas[schemaKey]; !exists {
						numberSchemas[schemaKey] = template.GetSchemaContent()
					}
				}
				continue
			}
		}

		// Check for array template
		if schemaType == "array" {
			template := templates.GetTemplateForArrayType(propSchema)
			if template != nil {
				validations = append(validations, template.GetValidation(sanitizedName, indent))

				// Store template schema content if needed
				if template.NeedsSeparateSchema() && template.GetSchemaContent() != "" {
					// Only add if we don't already have this schema
					if _, exists := templateSchemas["array"]; !exists {
						templateSchemas["array"] = template.GetSchemaContent()
					}
				}
				continue
			}
		}

		// Generate validations based on type - this is now only for types not handled by templates
		switch schemaType {
		case "array":
			// Min items
			if minItems, ok := utils.GetIntValue(propSchema, "minItems"); ok {
				validations = append(validations, indent+accessPath+" == None or len("+accessPath+") >= "+fmt.Sprintf("%d", minItems)+", \""+propName+" must have at least "+fmt.Sprintf("%d", minItems)+" items\"")
			}

			// Max items
			if maxItems, ok := utils.GetIntValue(propSchema, "maxItems"); ok {
				validations = append(validations, indent+accessPath+" == None or len("+accessPath+") <= "+fmt.Sprintf("%d", maxItems)+", \""+propName+" must have at most "+fmt.Sprintf("%d", maxItems)+" items\"")
			}

			// Unique items
			if uniqueItems, ok := utils.GetBoolValue(propSchema, "uniqueItems"); ok && uniqueItems {
				validations = append(validations, indent+accessPath+" == None or len("+accessPath+") == len({str(item): None for item in "+accessPath+"}), \""+propName+" must contain unique items\"")
			}

			// Item constraints
			itemsSchema, ok := utils.GetMapValue(propSchema, "items")
			if ok {
				itemType, typeOk := types.GetSchemaType(itemsSchema)
				if typeOk && itemType == "string" {
					// Check for min/max length in items
					if minLength, ok := utils.GetIntValue(itemsSchema, "minLength"); ok {
						validations = append(validations, indent+"# Item min length: "+fmt.Sprintf("%d", minLength))
					}
					if maxLength, ok := utils.GetIntValue(itemsSchema, "maxLength"); ok {
						validations = append(validations, indent+"# Item max length: "+fmt.Sprintf("%d", maxLength))
					}

					// Pattern validation for items
					if pattern, ok := utils.GetStringValue(itemsSchema, "pattern"); ok {
						validations = append(validations, indent+"# Item pattern: "+pattern)
					}

					// Format validation for items
					if format, ok := utils.GetStringValue(itemsSchema, "format"); ok {
						validations = append(validations, indent+"# Item format: "+format)
					}

					// Enum for string items
					if enumValues, ok := utils.GetArrayValue(itemsSchema, "enum"); ok && len(enumValues) > 0 {
						enumStr := formatEnumValues(enumValues)
						validations = append(validations, indent+"# Item allowed values: "+enumStr)
					}
				}

				// For numeric item types, add min/max constraints
				if typeOk && (itemType == "number" || itemType == "integer") {
					// Min/max for numeric items
					if minimum, ok := utils.GetFloatValue(itemsSchema, "minimum"); ok {
						validations = append(validations, indent+"# Item minimum: "+fmt.Sprintf("%v", minimum))
					}
					if maximum, ok := utils.GetFloatValue(itemsSchema, "maximum"); ok {
						if exclusiveMax, _ := utils.GetBoolValue(itemsSchema, "exclusiveMaximum"); exclusiveMax {
							validations = append(validations, indent+"# Item exclusive maximum: "+fmt.Sprintf("%v", maximum))
						} else {
							validations = append(validations, indent+"# Item maximum: "+fmt.Sprintf("%v", maximum))
						}
					}
					if exclusiveMin, ok := utils.GetFloatValue(itemsSchema, "exclusiveMinimum"); ok {
						validations = append(validations, indent+"# Item exclusive minimum: "+fmt.Sprintf("%v", exclusiveMin))
					}

					// Enum for numeric items
					if enumValues, ok := utils.GetArrayValue(itemsSchema, "enum"); ok && len(enumValues) > 0 {
						enumStr := formatEnumValues(enumValues)
						validations = append(validations, indent+"# Item allowed values: "+enumStr)
					}
				}
			}
		}
	}

	// Skip if no validations
	if len(validations) == 0 {
		return ""
	}

	// Build the validator schema
	var validatorSchema strings.Builder

	// Add any template schemas we need
	for _, schemaContent := range templateSchemas {
		validatorSchema.WriteString("\n")
		validatorSchema.WriteString(schemaContent)
		validatorSchema.WriteString("\n")
	}

	// Add any number schemas we need
	for _, schemaContent := range numberSchemas {
		validatorSchema.WriteString("\n")
		validatorSchema.WriteString(schemaContent)
		validatorSchema.WriteString("\n")
	}

	validatorSchema.WriteString("\n# Validator schema for " + schemaName + "\n")
	validatorSchema.WriteString("schema " + validatorName + ":\n")
	validatorSchema.WriteString("    self: " + schemaName + "\n\n")
	validatorSchema.WriteString("    check:\n")
	validatorSchema.WriteString(strings.Join(validations, "\n"))

	return validatorSchema.String()
}

// hasConstraints checks if a property schema has any constraints
func hasConstraints(propSchema map[string]interface{}) bool {
	schemaType, ok := types.GetSchemaType(propSchema)
	if !ok {
		return false
	}

	switch schemaType {
	case "string":
		_, hasMinLen := utils.GetIntValue(propSchema, "minLength")
		_, hasMaxLen := utils.GetIntValue(propSchema, "maxLength")
		_, hasPattern := utils.GetStringValue(propSchema, "pattern")
		_, hasFormat := utils.GetStringValue(propSchema, "format")
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")

		return hasMinLen || hasMaxLen || hasPattern || hasFormat || hasEnum

	case "number", "integer":
		_, hasMin := utils.GetFloatValue(propSchema, "minimum")
		_, hasMax := utils.GetFloatValue(propSchema, "maximum")
		_, hasExclusiveMin := utils.GetFloatValue(propSchema, "exclusiveMinimum")
		_, hasMultipleOf := utils.GetFloatValue(propSchema, "multipleOf")
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")

		return hasMin || hasMax || hasExclusiveMin || hasMultipleOf || hasEnum

	case "array":
		_, hasMinItems := utils.GetIntValue(propSchema, "minItems")
		_, hasMaxItems := utils.GetIntValue(propSchema, "maxItems")
		_, hasUniqueItems := utils.GetBoolValue(propSchema, "uniqueItems")

		return hasMinItems || hasMaxItems || hasUniqueItems

	case "boolean":
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")
		return hasEnum

	default:
		return false
	}
}

// formatEnumValues formats enum values for use in a KCL constraint
func formatEnumValues(values []interface{}) string {
	parts := []string{}
	for _, val := range values {
		parts = append(parts, utils.FormatLiteral(val))
	}
	return strings.Join(parts, ", ")
}

// formatEnumList formats enum values for use in a KCL constraint
func formatEnumList(values []interface{}) string {
	parts := []string{}
	for _, val := range values {
		parts = append(parts, utils.FormatLiteral(val))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// CheckIfNeedsRegexImport recursively inspects a schema to determine if regex validation is needed
func CheckIfNeedsRegexImport(rawSchema map[string]interface{}) bool {
	// Check for pattern in the schema itself
	if pattern, ok := utils.GetStringValue(rawSchema, "pattern"); ok && pattern != "" {
		return true
	}

	// Check for format in the schema itself (many formats use regex validation)
	if format, ok := utils.GetStringValue(rawSchema, "format"); ok && format != "" {
		return true
	}

	// Check in properties
	if properties, ok := utils.GetMapValue(rawSchema, "properties"); ok {
		for _, propValue := range properties {
			if propSchema, ok := propValue.(map[string]interface{}); ok {
				if CheckIfNeedsRegexImport(propSchema) {
					return true
				}
			}
		}
	}

	// Check in array items
	if items, ok := utils.GetMapValue(rawSchema, "items"); ok {
		if CheckIfNeedsRegexImport(items) {
			return true
		}
	}

	// Check in definitions
	if definitions, ok := utils.GetMapValue(rawSchema, "definitions"); ok {
		for _, defValue := range definitions {
			if defSchema, ok := defValue.(map[string]interface{}); ok {
				if CheckIfNeedsRegexImport(defSchema) {
					return true
				}
			}
		}
	}

	return false
}

// GenerateRequiredPropertyChecks generates KCL check blocks for required properties
func GenerateRequiredPropertyChecks(schema map[string]interface{}) string {
	required, ok := utils.GetArrayValue(schema, "required")
	if !ok || len(required) == 0 {
		return ""
	}

	checks := []string{}
	for _, req := range required {
		propName, ok := req.(string)
		if !ok {
			continue
		}
		sanitizedName := utils.SanitizePropertyName(propName)
		checks = append(checks, "    if "+sanitizedName+" == None:")
		checks = append(checks, "        assert False, \""+propName+" is required\"")
	}

	if len(checks) > 0 {
		return "\n" + strings.Join(checks, "\n")
	}

	return ""
}
