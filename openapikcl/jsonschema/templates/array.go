package templates

import (
	"fmt"
	"strings"
	"unicode"
)

// ArrayTemplate returns template for array type
func ArrayTemplate(propSchema map[string]interface{}) TypeTemplate {
	// Get type of items in the array
	itemType := "any"
	if items, ok := propSchema["items"].(map[string]interface{}); ok {
		// This is an array with a consistent type for all items
		if typeVal, ok := items["type"].(string); ok {
			switch typeVal {
			case "string":
				itemType = "str"
			case "integer":
				itemType = "int"
			case "number":
				itemType = "float"
			case "boolean":
				itemType = "bool"
			case "object":
				// If it's an object, it might be a schema reference or an inline schema
				if ref, ok := items["$ref"].(string); ok {
					// It's a reference, extract the name
					parts := strings.Split(ref, "/")
					if len(parts) > 0 {
						itemType = parts[len(parts)-1]
					}
				} else {
					// It's an inline schema, so use dict
					itemType = "dict"
				}
			default:
				itemType = "any"
			}
		}
	} else if _, ok := propSchema["items"].([]interface{}); ok {
		// This is a tuple-like array with different types for different positions
		// We don't fully support this in KCL, so we'll make it a list of any
		itemType = "any"
	}

	// Check if this is a tuple-like array (array with predefined items)
	if items, ok := propSchema["items"].([]interface{}); ok && len(items) > 0 {
		// Handle tuple array with object representation
		if shouldUseObjectForTuple(items) {
			return buildTupleObjectTemplate(propSchema, items)
		}
	}

	// Build type declaration
	typeName := fmt.Sprintf("[%s]", itemType)

	validationCode := buildArrayValidation(propSchema)
	schemaContent := buildArraySchemaContent(propSchema, itemType)
	description := ""
	if desc, ok := propSchema["description"].(string); ok {
		description = desc
	}
	comments := buildArrayComments(propSchema)

	return TypeTemplate{
		TypeName:       typeName,
		FormatName:     "array",
		Description:    description,
		ValidationCode: validationCode,
		Comments:       comments,
		SchemaContent:  schemaContent,
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
		validations = append(validations, "self.{property} == None or len(self.{property}) == len({v: v for v in self.{property}}), \"{property} must contain unique items\"")
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
		validationCode = append(validationCode, "        value == None or len(value) == len({v: v for v in value}), \"Array must contain unique items\"")
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

// Helper function to get boolean values

// Helper function to join validations and create a final validation string

// shouldUseObjectForTuple determines if we should use the object-based approach for tuples
func shouldUseObjectForTuple(items []interface{}) bool {
	// If any item has a description, it's a good candidate for object representation
	for _, item := range items {
		if itemObj, ok := item.(map[string]interface{}); ok {
			if _, hasDesc := itemObj["description"].(string); hasDesc {
				return true
			}
		}
	}

	// If there are more than 2 items, object approach is clearer
	if len(items) > 2 {
		return true
	}

	return false
}

// buildTupleObjectTemplate creates an object-based representation for tuple arrays
func buildTupleObjectTemplate(propSchema map[string]interface{}, items []interface{}) TypeTemplate {
	// Create a schema name based on the property name
	schemaName := "TupleObject"
	if name, ok := propSchema["title"].(string); ok && name != "" {
		schemaName = strings.Replace(name, " ", "", -1)
	}

	// Create object fields from tuple items
	var fieldDefs []string
	var fieldComments []string
	var fields []string

	for i, item := range items {
		if itemObj, ok := item.(map[string]interface{}); ok {
			fieldName := fmt.Sprintf("item%d", i)

			// Use description as field name if available
			if desc, ok := itemObj["description"].(string); ok && desc != "" {
				// Convert description to valid field name
				fieldName = strings.ToLower(strings.Replace(desc, " ", "_", -1))
				fieldName = strings.Replace(fieldName, ".", "_", -1)
				fieldName = strings.Replace(fieldName, "-", "_", -1)
				// Ensure first character is alphabetic
				if len(fieldName) > 0 && !unicode.IsLetter(rune(fieldName[0])) {
					fieldName = "f_" + fieldName
				}
			}

			fields = append(fields, fieldName)

			// Get the type for this item
			var itemType string
			if typeVal, ok := itemObj["type"].(string); ok {
				switch typeVal {
				case "string":
					itemType = "str"
				case "integer":
					itemType = "int"
				case "number":
					itemType = "float"
				case "boolean":
					itemType = "bool"
				case "object":
					// If it's an object, it might be a schema reference or an inline schema
					if ref, ok := itemObj["$ref"].(string); ok {
						// It's a reference, extract the name
						parts := strings.Split(ref, "/")
						if len(parts) > 0 {
							itemType = parts[len(parts)-1]
						}
					} else {
						// It's an inline schema, so use dict
						itemType = "dict"
					}
				default:
					itemType = "any"
				}
			} else {
				itemType = "any"
			}

			fieldDefs = append(fieldDefs, fmt.Sprintf("    %s: %s", fieldName, itemType))

			// Add comment if there's a description
			if desc, ok := itemObj["description"].(string); ok && desc != "" {
				fieldComments = append(fieldComments, fmt.Sprintf("    # %s", desc))
			}
		}
	}

	// Build the schema content
	content := fmt.Sprintf("schema %s:\n", schemaName)
	for i, def := range fieldDefs {
		if i < len(fieldComments) {
			content += fieldComments[i] + "\n"
		}
		content += def + "\n"
	}

	description := ""
	if desc, ok := propSchema["description"].(string); ok {
		description = desc
	}

	// Return the template
	return TypeTemplate{
		TypeName:       schemaName,
		FormatName:     "object",
		Description:    description,
		ValidationCode: "",
		Comments:       buildArrayComments(propSchema),
		SchemaContent:  content,
	}
}
