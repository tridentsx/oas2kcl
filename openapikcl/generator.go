package openapikcl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// GenerateKCLSchemas generates KCL schemas from a flattened OpenAPI spec
func GenerateKCLSchemas(doc *openapi3.T, outputDir string, packageName string) error {
	log.Print("starting KCL schema generation")

	if doc.Components == nil || doc.Components.Schemas == nil {
		log.Print("warning: no schemas found in OpenAPI components")
		return nil
	}

	// Create output directory based on packageName if outputDir is empty
	if outputDir == "" {
		outputDir = packageName
	}

	// Get schemas in deterministic order for consistent output
	schemaNames := collectSchemas(doc.Components.Schemas)
	log.Printf("processing %d schemas in order", len(schemaNames))

	// Track created schemas to avoid duplicates
	createdSchemas := make(map[string]bool)

	// Process each schema in order
	for _, name := range schemaNames {
		schema := doc.Components.Schemas[name]
		kclSchema, err := generateKCLSchema(name, schema, doc.Components.Schemas)
		if err != nil {
			return fmt.Errorf("failed to generate KCL schema for %s: %w", name, err)
		}

		// Generate the schema file
		if err := writeKCLSchemaFile(outputDir, name, kclSchema); err != nil {
			return fmt.Errorf("failed to write KCL schema for %s: %w", name, err)
		}

		createdSchemas[name] = true
	}

	log.Printf("completed generating KCL schemas for %d components", len(createdSchemas))
	return nil
}

// generateKCLSchema generates a KCL schema from an OpenAPI schema
func generateKCLSchema(name string, schema *openapi3.SchemaRef, allSchemas openapi3.Schemas) (string, error) {
	log.Printf("generating KCL schema for %s", name)

	if schema == nil || schema.Value == nil {
		return "", fmt.Errorf("schema is nil")
	}

	var builder strings.Builder

	// Check if any property has a pattern validation
	hasPatternValidation := schema.Value.Pattern != ""
	if !hasPatternValidation && schema.Value.Properties != nil {
		for _, propSchema := range schema.Value.Properties {
			if propSchema.Value != nil && propSchema.Value.Pattern != "" {
				hasPatternValidation = true
				break
			}
		}
	}

	// Add imports if needed
	if hasPatternValidation {
		// Add regex import for pattern matching
		builder.WriteString("import regex\n\n")
	}

	// Add schema documentation
	docString := FormatDocumentation(schema.Value)
	if docString != "" {
		builder.WriteString(docString)
	}

	// Begin schema definition
	schemaName := formatSchemaName(name)
	builder.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

	// Process inherited schemas (allOf)
	parentSchemas, err := processInheritance(schema.Value, allSchemas)
	if err != nil {
		return "", fmt.Errorf("error processing inheritance for %s: %w", name, err)
	}
	if len(parentSchemas) > 0 {
		// Add mixin for inherited schemas
		parents := make([]string, len(parentSchemas))
		for i, parent := range parentSchemas {
			parents[i] = formatSchemaName(parent)
		}
		builder.WriteString(fmt.Sprintf("    mixin [%s]\n", strings.Join(parents, ", ")))
	}

	// Add properties
	hasConstraints := false
	var constraints []string

	if schema.Value.Properties != nil {
		propCount := 0
		for fieldName, fieldSchema := range schema.Value.Properties {
			propCount++

			// Check if property is required
			isRequired := false
			for _, req := range schema.Value.Required {
				if req == fieldName {
					isRequired = true
					break
				}
			}

			// Get field documentation
			if fieldSchema.Value != nil {
				fieldDoc := FormatDocumentation(fieldSchema.Value)
				if fieldDoc != "" {
					// Indent the documentation for the field
					lines := strings.Split(fieldDoc, "\n")
					for _, line := range lines {
						if line != "" {
							builder.WriteString(fmt.Sprintf("    %s\n", line))
						}
					}
				}
			}

			// Generate field type
			fieldType, isComplexType := generateFieldType(fieldName, fieldSchema, isRequired)

			// Add ? suffix to field name if not required
			formattedFieldName := fieldName
			if !isRequired {
				formattedFieldName = fieldName + "?"
			}

			// Add the field type
			fieldDefinition := fmt.Sprintf("    %s: %s", formattedFieldName, fieldType)

			// Add default value if present
			if fieldSchema.Value != nil && fieldSchema.Value.Default != nil {
				// Format the default value based on its type
				var defaultStr string
				switch v := fieldSchema.Value.Default.(type) {
				case string:
					defaultStr = fmt.Sprintf(`"%s"`, v)
				case bool:
					// KCL uses capitalized True/False for boolean literals
					if v {
						defaultStr = "True"
					} else {
						defaultStr = "False"
					}
				case float64, float32, int, int64, int32:
					defaultStr = fmt.Sprintf("%v", v)
				default:
					// Skip complex default values or use appropriate serialization
					log.Printf("warning: complex default value for field %s not directly supported", fieldName)
					defaultStr = ""
				}

				if defaultStr != "" {
					fieldDefinition += fmt.Sprintf(" = %s", defaultStr)
				}
			}

			builder.WriteString(fieldDefinition + "\n")

			// Generate constraints
			if fieldSchema.Value != nil {
				fieldConstraints := GenerateConstraints(fieldSchema.Value, fieldName)
				if len(fieldConstraints) > 0 {
					hasConstraints = true
					for _, constraint := range fieldConstraints {
						constraints = append(constraints, fmt.Sprintf("    %s", constraint))
					}
				}

				// Add complex type validations if needed
				if isComplexType && fieldSchema.Value.AdditionalProperties.Has != nil {
					// Handle additionalProperties validation
					if fieldSchema.Value.AdditionalProperties.Schema != nil {
						// Additional validation specific to map types could be added here
					}
				}
			}
		}

		log.Printf("generated %d properties for schema %s", propCount, name)
	}

	// Add check block for constraints if needed
	if hasConstraints {
		builder.WriteString("\n    check:\n")
		for _, constraint := range constraints {
			builder.WriteString(fmt.Sprintf("        %s\n", constraint))
		}
	}

	// Add oneOf/anyOf handling
	if len(schema.Value.OneOf) > 0 || len(schema.Value.AnyOf) > 0 {
		builder.WriteString("\n    # This schema has alternative validation rules (oneOf/anyOf)\n")
		builder.WriteString("    # KCL doesn't directly support these OpenAPI constructs\n")

		// Add some commented guidance for oneOf
		if len(schema.Value.OneOf) > 0 {
			builder.WriteString("    # oneOf validation requires exactly one of the following schemas to be valid:\n")
			for i, oneOfSchema := range schema.Value.OneOf {
				schemaRef := extractSchemaName(oneOfSchema.Ref)
				builder.WriteString(fmt.Sprintf("    # Option %d: %s\n", i+1, schemaRef))
			}
		}

		// Add some commented guidance for anyOf
		if len(schema.Value.AnyOf) > 0 {
			builder.WriteString("    # anyOf validation requires at least one of the following schemas to be valid:\n")
			for i, anyOfSchema := range schema.Value.AnyOf {
				schemaRef := extractSchemaName(anyOfSchema.Ref)
				builder.WriteString(fmt.Sprintf("    # Option %d: %s\n", i+1, schemaRef))
			}
		}
	}

	return builder.String(), nil
}

// generateFieldType determines the appropriate KCL type for a field
func generateFieldType(fieldName string, fieldSchema *openapi3.SchemaRef, isRequired bool) (string, bool) {
	if fieldSchema == nil || fieldSchema.Value == nil {
		return "any", false
	}

	// Handle references
	if fieldSchema.Ref != "" {
		refName := extractSchemaName(fieldSchema.Ref)
		// Format the reference name as a KCL schema name
		formattedRef := formatSchemaName(refName)
		return formattedRef, false
	}

	isComplexType := false
	var fieldType string

	// Extract type information
	if fieldSchema.Value.Type != nil && len(*fieldSchema.Value.Type) > 0 {
		openAPIType := (*fieldSchema.Value.Type)[0]

		// Handle different types
		switch openAPIType {
		case "array":
			if fieldSchema.Value.Items != nil {
				// Get the item type
				itemType, _ := generateFieldType("item", fieldSchema.Value.Items, true)
				// For arrays of complex objects, just use any for now
				// KCL doesn't support inline complex object definitions in arrays
				if strings.Contains(itemType, "{") || strings.Contains(itemType, "}") {
					fieldType = "[any]"
				} else {
					fieldType = fmt.Sprintf("[%s]", itemType)
				}
			} else {
				fieldType = "[any]"
			}
		case "object":
			// Handle plain objects with properties
			if fieldSchema.Value.Properties != nil && len(fieldSchema.Value.Properties) > 0 {
				// KCL doesn't support inline object definitions with complex types
				// Just use dict for complex nested objects
				fieldType = "dict"
				isComplexType = true
			} else if fieldSchema.Value.AdditionalProperties.Has != nil {
				// Handle map types (objects with additionalProperties)
				if fieldSchema.Value.AdditionalProperties.Schema != nil {
					// Just use dict for maps
					fieldType = "dict"
				} else {
					fieldType = "dict"
				}
				isComplexType = true
			} else {
				fieldType = "dict"
			}
		default:
			// Use the basic type converter for primitive types
			fieldType = ConvertTypeToKCL(openAPIType, fieldSchema.Value.Format)
		}
	} else {
		// If no type is specified
		fieldType = "any"
	}

	return fieldType, isComplexType
}

// processInheritance handles allOf inheritance in OpenAPI schemas
func processInheritance(schema *openapi3.Schema, allSchemas openapi3.Schemas) ([]string, error) {
	var parents []string

	// Process allOf to extract parent schemas
	if len(schema.AllOf) > 0 {
		for _, allOfSchema := range schema.AllOf {
			// If it's a reference, add it as a parent
			if allOfSchema.Ref != "" {
				parent := extractSchemaName(allOfSchema.Ref)
				parents = append(parents, parent)
			}
			// If it has properties, these should be merged into the current schema
			// This is typically handled by the flattener
		}
	}

	return parents, nil
}

// formatSchemaName ensures the schema name follows KCL naming conventions
func formatSchemaName(name string) string {
	// UpperCamelCase for schema names in KCL
	parts := strings.Split(name, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// extractSchemaName extracts the schema name from a reference
func extractSchemaName(ref string) string {
	// Handle "#/components/schemas/Name" format
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

// writeKCLSchemaFile writes the KCL schema to a file
func writeKCLSchemaFile(outputDir, name, content string) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the file path
	filePath := filepath.Join(outputDir, name+".k")

	// Write the content to the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	log.Printf("Schema %s written to %s", name, filePath)
	return nil
}
