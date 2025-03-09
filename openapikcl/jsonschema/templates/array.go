package templates

import (
	"fmt"
)

// ArrayTemplate returns a template for array validation
func ArrayTemplate(propSchema map[string]interface{}) TypeTemplate {
	// Get the item type if available
	var itemType string
	if items, ok := propSchema["items"].(map[string]interface{}); ok {
		// This is a simple array with a single item type
		if itemTypeVal, ok := items["type"].(string); ok {
			switch itemTypeVal {
			case "string":
				itemType = "str"
			case "integer":
				itemType = "int"
			case "number":
				itemType = "float"
			case "boolean":
				itemType = "bool"
			case "array":
				itemType = "list"
			case "object":
				itemType = "dict"
			default:
				itemType = "any"
			}
		} else {
			itemType = "any"
		}
	} else {
		// Default to any for item type if not specified
		itemType = "any"
	}

	// Build validation code
	validationCode := buildArrayValidation(propSchema)

	typeName := "list"
	if itemType != "any" {
		typeName = fmt.Sprintf("list[%s]", itemType)
	}

	return TypeTemplate{
		TypeName:       typeName,
		FormatName:     "array",
		Description:    "Array with constraints",
		ValidationCode: validationCode,
		Comments:       buildArrayComments(propSchema),
		SchemaContent:  buildArraySchemaContent(propSchema, itemType),
	}
}

// buildArrayValidation builds validation code for arrays based on constraints
func buildArrayValidation(propSchema map[string]interface{}) string {
	validations := []string{}

	// minItems
	if minItems, ok := getInt(propSchema, "minItems"); ok {
		validations = append(validations, fmt.Sprintf("self.{property} == None or len(self.{property}) >= %d, \"{property} must have at least %d items\"", minItems, minItems))
	}

	// maxItems
	if maxItems, ok := getInt(propSchema, "maxItems"); ok {
		validations = append(validations, fmt.Sprintf("self.{property} == None or len(self.{property}) <= %d, \"{property} must have at most %d items\"", maxItems, maxItems))
	}

	// uniqueItems
	if uniqueItems, ok := getBool(propSchema, "uniqueItems"); ok && uniqueItems {
		validations = append(validations, "self.{property} == None or len(self.{property}) == len(set(self.{property})), \"{property} must contain unique items\"")
	}

	return buildFinalValidation(validations)
}

// buildArrayComments builds comments for array constraints
func buildArrayComments(propSchema map[string]interface{}) []string {
	comments := []string{}

	// minItems
	if minItems, ok := getInt(propSchema, "minItems"); ok {
		comments = append(comments, fmt.Sprintf("# Minimum items: %d", minItems))
	}

	// maxItems
	if maxItems, ok := getInt(propSchema, "maxItems"); ok {
		comments = append(comments, fmt.Sprintf("# Maximum items: %d", maxItems))
	}

	// uniqueItems
	if uniqueItems, ok := getBool(propSchema, "uniqueItems"); ok && uniqueItems {
		comments = append(comments, "# All items must be unique")
	}

	return comments
}

// buildArraySchemaContent builds schema content for arrays
func buildArraySchemaContent(propSchema map[string]interface{}, itemType string) string {
	// Build validation code for the schema
	validationCode := []string{}

	// minItems
	if minItems, ok := getInt(propSchema, "minItems"); ok {
		validationCode = append(validationCode, fmt.Sprintf("        value == None or len(value) >= %d, \"Array must have at least %d items\"", minItems, minItems))
	}

	// maxItems
	if maxItems, ok := getInt(propSchema, "maxItems"); ok {
		validationCode = append(validationCode, fmt.Sprintf("        value == None or len(value) <= %d, \"Array must have at most %d items\"", maxItems, maxItems))
	}

	// uniqueItems
	if uniqueItems, ok := getBool(propSchema, "uniqueItems"); ok && uniqueItems {
		validationCode = append(validationCode, "        value == None or len(value) == len(set(value)), \"Array must contain unique items\"")
	}

	// Only create schema content if we have validations
	if len(validationCode) == 0 {
		return ""
	}

	// Create a full type name for the schema value
	valueType := "list"
	if itemType != "any" {
		valueType = fmt.Sprintf("list[%s]", itemType)
	}

	return fmt.Sprintf(`schema Array:
    """Array validation.
    
    Validates arrays to ensure they conform to specified constraints.
    """
    value: %s
    
    check:
%s
`, valueType, buildFinalValidation(validationCode))
}

// Helper function to get integer values
func getInt(schema map[string]interface{}, key string) (int, bool) {
	if value, ok := schema[key]; ok {
		switch v := value.(type) {
		case int:
			return v, true
		case int64:
			return int(v), true
		case float64:
			return int(v), true
		}
	}
	return 0, false
}
