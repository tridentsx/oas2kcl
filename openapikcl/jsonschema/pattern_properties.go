package jsonschema

import (
	"fmt"
	"strings"

	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/utils"
)

// PatternProperty represents a JSON Schema pattern property
type PatternProperty struct {
	Pattern     string
	Schema      map[string]interface{}
	Description string
	Type        string
}

// GeneratePatternPropertySchema generates a KCL schema for a pattern property
func GeneratePatternPropertySchema(schemaName string, patternProp PatternProperty) string {
	var schema strings.Builder

	// Generate schema header
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

	// Add description
	if patternProp.Description != "" {
		schema.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", patternProp.Description))
	}

	// Add value property with the appropriate type
	valueType := "str"
	if patternProp.Type != "" {
		switch patternProp.Type {
		case "string":
			valueType = "str"
		case "number":
			valueType = "float"
		case "integer":
			valueType = "int"
		case "boolean":
			valueType = "bool"
		case "object":
			valueType = "dict"
		case "array":
			valueType = "list"
		}
	}

	schema.WriteString(fmt.Sprintf("    key: str\n"))
	schema.WriteString(fmt.Sprintf("    value: %s\n\n", valueType))

	// Add check block for pattern validation
	schema.WriteString("    check:\n")

	// Translate the pattern to RE2 format
	re2Pattern := TranslateECMAToRE2(patternProp.Pattern)

	// Add pattern validation for the key
	schema.WriteString(fmt.Sprintf("        regex.match(key, r\"%s\"), \"key must match pattern %s\"\n",
		re2Pattern, patternProp.Pattern))

	// Add additional validations based on the schema
	if format, ok := utils.GetStringValue(patternProp.Schema, "format"); ok {
		schema.WriteString(generateFormatValidation(format, "value"))
	}

	if minLength, ok := utils.GetIntValue(patternProp.Schema, "minLength"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) >= %d, \"value must have a minimum length of %d\"\n",
			minLength, minLength))
	}

	if maxLength, ok := utils.GetIntValue(patternProp.Schema, "maxLength"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) <= %d, \"value must have a maximum length of %d\"\n",
			maxLength, maxLength))
	}

	if pattern, ok := utils.GetStringValue(patternProp.Schema, "pattern"); ok {
		re2ValuePattern := TranslateECMAToRE2(pattern)
		schema.WriteString(fmt.Sprintf("        value == None or regex.match(value, r\"%s\"), \"value must match pattern %s\"\n",
			re2ValuePattern, pattern))
	}

	if minimum, ok := utils.GetFloatValue(patternProp.Schema, "minimum"); ok && (patternProp.Type == "number" || patternProp.Type == "integer") {
		schema.WriteString(fmt.Sprintf("        value == None or value >= %g, \"value must be greater than or equal to %g\"\n",
			minimum, minimum))
	}

	if maximum, ok := utils.GetFloatValue(patternProp.Schema, "maximum"); ok && (patternProp.Type == "number" || patternProp.Type == "integer") {
		schema.WriteString(fmt.Sprintf("        value == None or value <= %g, \"value must be less than or equal to %g\"\n",
			maximum, maximum))
	}

	return schema.String()
}

// GeneratePatternPropertiesValidator generates a KCL schema that validates all pattern properties
func GeneratePatternPropertiesValidator(schemaName string, patternProps map[string]PatternProperty, imports *[]string) string {
	var schema strings.Builder

	// Generate schema header
	schema.WriteString(fmt.Sprintf("schema %sValidator:\n", schemaName))
	schema.WriteString("    \"\"\"Validates pattern properties for dynamic keys.\"\"\"\n")
	schema.WriteString("    data: dict\n\n")

	// Add check block
	schema.WriteString("    check:\n")

	// Add validation for each pattern property
	for pattern, prop := range patternProps {
		// Create a validator schema name for this pattern
		validatorName := fmt.Sprintf("%s_%s_Validator", schemaName, sanitizePatternForKCL(pattern))

		// Add import for the validator
		*imports = append(*imports, fmt.Sprintf("import %s", validatorName))

		// Add validation for this pattern
		schema.WriteString(fmt.Sprintf("        # Validate keys matching pattern: %s\n", pattern))
		schema.WriteString(fmt.Sprintf("        all k, v in data {\n"))
		schema.WriteString(fmt.Sprintf("            if regex.match(k, r\"%s\") {\n", TranslateECMAToRE2(pattern)))
		schema.WriteString(fmt.Sprintf("                %s {\n", validatorName))
		schema.WriteString(fmt.Sprintf("                    key = k\n"))
		schema.WriteString(fmt.Sprintf("                    value = v\n"))
		schema.WriteString(fmt.Sprintf("                }\n"))
		schema.WriteString(fmt.Sprintf("            }\n"))
		schema.WriteString(fmt.Sprintf("        }\n"))

		// Generate individual schema for this pattern property
		if prop.Type != "" {
			schema.WriteString(fmt.Sprintf("\n        # Schema for pattern: %s with type: %s\n", pattern, prop.Type))
		}
	}

	return schema.String()
}

// TranslateECMAToRE2 translates an ECMA regex pattern to RE2 format
func TranslateECMAToRE2(pattern string) string {
	// This is a simplified version - in a real implementation, you would use the
	// utils.TranslateECMAToGoRegex function or similar

	// Escape backslashes
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")

	// Escape double quotes
	pattern = strings.ReplaceAll(pattern, "\"", "\\\"")

	return pattern
}

// generateFormatValidation generates validation for a specific format
func generateFormatValidation(format, propName string) string {
	switch format {
	case "email":
		return fmt.Sprintf("        %s == None or regex.match(%s, r\"^[a-zA-Z0-9.!#$%%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$\"), \"%s must be a valid email address\"\n",
			propName, propName, propName)
	case "uri":
		return fmt.Sprintf("        %s == None or regex.match(%s, r\"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]*$\"), \"%s must be a valid URI\"\n",
			propName, propName, propName)
	case "date-time":
		return fmt.Sprintf("        %s == None or regex.match(%s, r\"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(?:\\.\\d+)?(?:Z|[+-]\\d{2}:\\d{2})$\"), \"%s must be a valid RFC 3339 date-time\"\n",
			propName, propName, propName)
	case "ipv4":
		return fmt.Sprintf("        %s == None or regex.match(%s, r\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\"), \"%s must be a valid IPv4 address\"\n",
			propName, propName, propName)
	case "uuid":
		return fmt.Sprintf("        %s == None or regex.match(%s, r\"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$\"), \"%s must be a valid UUID\"\n",
			propName, propName, propName)
	default:
		return ""
	}
}

// sanitizePatternForKCL converts a regex pattern to a valid KCL identifier
func sanitizePatternForKCL(pattern string) string {
	// Replace non-alphanumeric characters with underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, pattern)

	// Ensure it starts with a letter
	if len(result) > 0 && (result[0] >= '0' && result[0] <= '9') {
		result = "p_" + result
	}

	return result
}
