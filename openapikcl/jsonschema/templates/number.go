package templates

import (
	"fmt"
	"strings"
)

// NumberTemplate returns a template for generic number validation
func NumberTemplate(propSchema map[string]interface{}) TypeTemplate {
	// Build validation code
	validationCode := buildNumberValidation(propSchema, "float")

	return TypeTemplate{
		TypeName:       "float",
		FormatName:     "number",
		Description:    "Number value with constraints",
		ValidationCode: validationCode,
		Comments:       buildNumberComments(propSchema),
		SchemaContent:  buildNumberSchemaContent(propSchema, "float"),
	}
}

// IntegerTemplate returns a template for integer validation
func IntegerTemplate(propSchema map[string]interface{}) TypeTemplate {
	// Build validation code
	validationCode := buildNumberValidation(propSchema, "int")

	return TypeTemplate{
		TypeName:       "int",
		FormatName:     "integer",
		Description:    "Integer value with constraints",
		ValidationCode: validationCode,
		Comments:       buildNumberComments(propSchema),
		SchemaContent:  buildNumberSchemaContent(propSchema, "int"),
	}
}

// buildNumberValidation builds validation code for numbers based on constraints
func buildNumberValidation(propSchema map[string]interface{}, typeName string) string {
	validations := []string{}

	// Minimum
	if minimum, ok := getFloat(propSchema, "minimum"); ok {
		validations = append(validations, fmt.Sprintf("self.{property} == None or self.{property} >= %v, \"{property} must be at least %v\"", minimum, minimum))
	}

	// Maximum
	if maximum, ok := getFloat(propSchema, "maximum"); ok {
		exclusiveMax, _ := getBool(propSchema, "exclusiveMaximum")
		if exclusiveMax {
			validations = append(validations, fmt.Sprintf("self.{property} == None or self.{property} < %v, \"{property} must be less than %v\"", maximum, maximum))
		} else {
			validations = append(validations, fmt.Sprintf("self.{property} == None or self.{property} <= %v, \"{property} must be at most %v\"", maximum, maximum))
		}
	}

	// ExclusiveMinimum - in JSON Schema 7 this is a number, in older versions it's a boolean modifier for minimum
	if exclusiveMin, ok := getFloat(propSchema, "exclusiveMinimum"); ok {
		validations = append(validations, fmt.Sprintf("self.{property} == None or self.{property} > %v, \"{property} must be greater than %v\"", exclusiveMin, exclusiveMin))
	}

	// MultipleOf
	if multipleOf, ok := getFloat(propSchema, "multipleOf"); ok {
		if typeName == "int" {
			validations = append(validations, fmt.Sprintf("self.{property} == None or self.{property} %% %v == 0, \"{property} must be a multiple of %v\"", multipleOf, multipleOf))
		} else {
			// For float, we use a more complex validation since direct modulo isn't reliable for floats
			// Instead, we check if the division is close to an integer
			validations = append(validations, fmt.Sprintf("self.{property} == None or abs(self.{property} / %v - round(self.{property} / %v)) < 1e-10, \"{property} must be a multiple of %v\"", multipleOf, multipleOf, multipleOf))
		}
	}

	// Enum
	if enumValues, ok := getArray(propSchema, "enum"); ok && len(enumValues) > 0 {
		enumStr := formatNumberEnumValues(enumValues)
		validations = append(validations, fmt.Sprintf("self.{property} == None or self.{property} in [%s], \"{property} must be one of the allowed values\"", enumStr))
	}

	return buildFinalValidation(validations)
}

// buildNumberComments builds comments for number constraints
func buildNumberComments(propSchema map[string]interface{}) []string {
	comments := []string{}

	// Minimum
	if minimum, ok := getFloat(propSchema, "minimum"); ok {
		comments = append(comments, fmt.Sprintf("# Minimum: %v", minimum))
	}

	// Maximum
	if maximum, ok := getFloat(propSchema, "maximum"); ok {
		exclusiveMax, _ := getBool(propSchema, "exclusiveMaximum")
		if exclusiveMax {
			comments = append(comments, fmt.Sprintf("# Exclusive maximum: %v", maximum))
		} else {
			comments = append(comments, fmt.Sprintf("# Maximum: %v", maximum))
		}
	}

	// ExclusiveMinimum as a number
	if exclusiveMin, ok := getFloat(propSchema, "exclusiveMinimum"); ok {
		comments = append(comments, fmt.Sprintf("# Exclusive minimum: %v", exclusiveMin))
	}

	// MultipleOf
	if multipleOf, ok := getFloat(propSchema, "multipleOf"); ok {
		comments = append(comments, fmt.Sprintf("# Multiple of: %v", multipleOf))
	}

	// Enum
	if enumValues, ok := getArray(propSchema, "enum"); ok && len(enumValues) > 0 {
		enumStr := formatNumberEnumValues(enumValues)
		comments = append(comments, fmt.Sprintf("# Allowed values: %s", enumStr))
	}

	return comments
}

// buildNumberSchemaContent builds schema content for numbers
func buildNumberSchemaContent(propSchema map[string]interface{}, typeName string) string {
	// Build validation code for the schema
	validationCode := []string{}

	// Minimum
	if minimum, ok := getFloat(propSchema, "minimum"); ok {
		validationCode = append(validationCode, fmt.Sprintf("        value == None or value >= %v, \"Value must be at least %v\"", minimum, minimum))
	}

	// Maximum
	if maximum, ok := getFloat(propSchema, "maximum"); ok {
		exclusiveMax, _ := getBool(propSchema, "exclusiveMaximum")
		if exclusiveMax {
			validationCode = append(validationCode, fmt.Sprintf("        value == None or value < %v, \"Value must be less than %v\"", maximum, maximum))
		} else {
			validationCode = append(validationCode, fmt.Sprintf("        value == None or value <= %v, \"Value must be at most %v\"", maximum, maximum))
		}
	}

	// ExclusiveMinimum as a number
	if exclusiveMin, ok := getFloat(propSchema, "exclusiveMinimum"); ok {
		validationCode = append(validationCode, fmt.Sprintf("        value == None or value > %v, \"Value must be greater than %v\"", exclusiveMin, exclusiveMin))
	}

	// MultipleOf
	if multipleOf, ok := getFloat(propSchema, "multipleOf"); ok {
		if typeName == "int" {
			validationCode = append(validationCode, fmt.Sprintf("        value == None or value %% %v == 0, \"Value must be a multiple of %v\"", multipleOf, multipleOf))
		} else {
			// For float, we use a more complex validation since direct modulo isn't reliable for floats
			validationCode = append(validationCode, fmt.Sprintf("        value == None or abs(value / %v - round(value / %v)) < 1e-10, \"Value must be a multiple of %v\"", multipleOf, multipleOf, multipleOf))
		}
	}

	// Only create schema content if we have validations
	if len(validationCode) == 0 {
		return ""
	}

	// Capitalize the type name for schema name
	typeSchemeName := "Integer"
	if typeName == "float" {
		typeSchemeName = "Number"
	}

	return fmt.Sprintf(`schema %s:
    """%s value validation.
    
    Validates %s values to ensure they conform to specified constraints.
    """
    value: %s
    
    check:
%s
`, typeSchemeName, typeSchemeName, typeName, typeName, buildFinalValidation(validationCode))
}

// Helper function to join validations and create a final validation string
func buildFinalValidation(validations []string) string {
	if len(validations) == 0 {
		return ""
	}
	return strings.Join(validations, "\n")
}

// Helper functions to extract values from the schema
func getFloat(schema map[string]interface{}, key string) (float64, bool) {
	if value, ok := schema[key]; ok {
		switch v := value.(type) {
		case float64:
			return v, true
		case float32:
			return float64(v), true
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		}
	}
	return 0, false
}

func getBool(schema map[string]interface{}, key string) (bool, bool) {
	if value, ok := schema[key]; ok {
		if boolVal, ok := value.(bool); ok {
			return boolVal, true
		}
	}
	return false, false
}

func getArray(schema map[string]interface{}, key string) ([]interface{}, bool) {
	if value, ok := schema[key]; ok {
		if array, ok := value.([]interface{}); ok {
			return array, true
		}
	}
	return nil, false
}

// Helper function to format number enum values
func formatNumberEnumValues(values []interface{}) string {
	parts := []string{}
	for _, val := range values {
		switch v := val.(type) {
		case float64:
			parts = append(parts, fmt.Sprintf("%v", v))
		case int:
			parts = append(parts, fmt.Sprintf("%d", v))
		case int64:
			parts = append(parts, fmt.Sprintf("%d", v))
		}
	}
	return strings.Join(parts, ", ")
}
