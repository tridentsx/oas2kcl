package openapikcl

import (
	"fmt"
	"log"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ConvertTypeToKCL maps OpenAPI types to KCL types
func ConvertTypeToKCL(oapiType, format string) string {
	log.Printf("converting OpenAPI type '%s' with format '%s' to KCL type", oapiType, format)

	var kclType string
	switch oapiType {
	case "string":
		switch format {
		case "date-time":
			kclType = "datetime"
		case "date":
			kclType = "date"
		case "email":
			kclType = "str" // KCL doesn't have a dedicated email type, but we'll add validation
		case "uuid":
			kclType = "str" // KCL doesn't have a dedicated UUID type, but we'll add validation
		case "uri":
			kclType = "str" // Same for URI
		default:
			kclType = "str"
		}
	case "integer":
		switch format {
		case "int32":
			kclType = "int"
		case "int64":
			kclType = "int"
		default:
			kclType = "int"
		}
	case "boolean":
		kclType = "bool"
	case "number":
		switch format {
		case "float":
			kclType = "float"
		case "double":
			kclType = "float"
		default:
			kclType = "float"
		}
	case "array":
		kclType = "list" // The element type will be handled separately
	case "object":
		kclType = "dict" // For generic objects, specific schema types will be handled differently
	default:
		log.Printf("warning: unknown type '%s', defaulting to 'any'", oapiType)
		kclType = "any"
	}

	log.Printf("mapped to KCL type: %s", kclType)
	return kclType
}

// GenerateConstraints creates KCL constraint expressions for a schema
func GenerateConstraints(schema *openapi3.Schema, fieldName string) []string {
	var constraints []string

	// Required validation is handled at the schema level

	// String constraints - MinLength is uint64 (non-pointer)
	if schema.MinLength > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", fieldName, schema.MinLength))
	}
	// MaxLength is *uint64 (pointer)
	if schema.MaxLength != nil && *schema.MaxLength > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", fieldName, *schema.MaxLength))
	}
	if schema.Pattern != "" {
		// KCL uses regex matching
		constraints = append(constraints, fmt.Sprintf("%s.match(r\"%s\")", fieldName, schema.Pattern))
	}

	// Numeric constraints
	if schema.Min != nil {
		if schema.ExclusiveMin {
			constraints = append(constraints, fmt.Sprintf("%s > %v", fieldName, *schema.Min))
		} else {
			constraints = append(constraints, fmt.Sprintf("%s >= %v", fieldName, *schema.Min))
		}
	}
	if schema.Max != nil {
		if schema.ExclusiveMax {
			constraints = append(constraints, fmt.Sprintf("%s < %v", fieldName, *schema.Max))
		} else {
			constraints = append(constraints, fmt.Sprintf("%s <= %v", fieldName, *schema.Max))
		}
	}
	if schema.MultipleOf != nil {
		// KCL doesn't have a direct way to check if a number is a multiple of another
		// But we can use a modulo check
		constraints = append(constraints, fmt.Sprintf("%s %% %v == 0", fieldName, *schema.MultipleOf))
	}

	// Array constraints - MinItems is uint64 (non-pointer)
	if schema.MinItems > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", fieldName, schema.MinItems))
	}
	// MaxItems is *uint64 (pointer)
	if schema.MaxItems != nil && *schema.MaxItems > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", fieldName, *schema.MaxItems))
	}
	if schema.UniqueItems {
		// KCL doesn't have a direct uniqueness check, but we can compare length with set length
		constraints = append(constraints, fmt.Sprintf("len(%s) == len(set(%s))", fieldName, fieldName))
	}

	// Enum validation
	if len(schema.Enum) > 0 {
		values := make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			// Format the enum value based on its type
			switch value := v.(type) {
			case string:
				values[i] = fmt.Sprintf("\"%s\"", value)
			default:
				values[i] = fmt.Sprintf("%v", value)
			}
		}
		constraints = append(constraints, fmt.Sprintf("%s in [%s]", fieldName, strings.Join(values, ", ")))
	}

	return constraints
}

// FormatDocumentation generates KCL documentation from OpenAPI schema
func FormatDocumentation(schema *openapi3.Schema) string {
	var doc strings.Builder

	if schema.Title != "" {
		doc.WriteString(fmt.Sprintf("# %s\n", schema.Title))
	}

	if schema.Description != "" {
		// Format multiline descriptions for KCL comment syntax
		lines := strings.Split(schema.Description, "\n")
		for _, line := range lines {
			doc.WriteString(fmt.Sprintf("# %s\n", line))
		}
	}

	if schema.Default != nil {
		doc.WriteString(fmt.Sprintf("# Default: %v\n", schema.Default))
	}

	if schema.Deprecated {
		doc.WriteString("# DEPRECATED\n")
	}

	if schema.ReadOnly {
		doc.WriteString("# ReadOnly: This field is read-only\n")
	}

	if schema.WriteOnly {
		doc.WriteString("# WriteOnly: This field is write-only\n")
	}

	return doc.String()
}
