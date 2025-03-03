package openapikcl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

// Add debugMode variable if it doesn't exist
var debugMode bool

// GenerateKCLSchemas generates KCL schemas from a flattened OpenAPI spec
func GenerateKCLSchemas(doc *openapi3.T, outputDir string, packageName string, version OpenAPIVersion) error {
	log.Printf("starting KCL schema generation (OpenAPI version: %s)", version)

	// Handle any version-specific preprocessing
	if IsSwaggerVersion(version) {
		// Any OpenAPI 2.0 specific processing
		HandleSwaggerSpecifics(version)
	}

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
		kclSchema, err := GenerateKCLSchema(name, schema, doc.Components.Schemas, version, doc)
		if err != nil {
			return fmt.Errorf("failed to generate KCL schema for %s: %w", name, err)
		}

		// Generate the schema file
		if err := writeKCLSchemaFile(outputDir, name, kclSchema); err != nil {
			return fmt.Errorf("failed to write KCL schema for %s: %w", name, err)
		}

		createdSchemas[name] = true
	}

	// Extract all schema names from doc.Components.Schemas
	var allSchemaNames []string
	for schemaName := range doc.Components.Schemas {
		allSchemaNames = append(allSchemaNames, schemaName)
	}

	// Generate a main.k file that imports all schemas to handle circular dependencies
	if err := generateMainK(outputDir, schemaNames, allSchemaNames); err != nil {
		return fmt.Errorf("failed to generate main.k file: %w", err)
	}

	log.Printf("completed generating KCL schemas for %d components", len(createdSchemas))
	return nil
}

// collectSchemaReferences recursively processes a schema to identify all referenced schemas
func collectSchemaReferences(schema *openapi3.SchemaRef, currentSchemaName string, allSchemas openapi3.Schemas, referencedSchemas map[string]bool) {
	// If this is a reference, add it to our map
	if schema.Ref != "" {
		// Extract schema name from reference
		parts := strings.Split(schema.Ref, "/")
		refName := parts[len(parts)-1]

		// Skip self-references or already processed references
		if refName != currentSchemaName && !referencedSchemas[refName] {
			referencedSchemas[refName] = true

			// Process the referenced schema to get its references too
			if refSchema, ok := allSchemas[refName]; ok {
				collectSchemaReferences(refSchema, currentSchemaName, allSchemas, referencedSchemas)
			}
		}
		return
	}

	if schema.Value == nil {
		return
	}

	// Process nested properties
	for _, propSchema := range schema.Value.Properties {
		collectSchemaReferences(propSchema, currentSchemaName, allSchemas, referencedSchemas)
	}

	// Process arrays - This is critical for array items that reference other schemas
	if schema.Value.Items != nil {
		collectSchemaReferences(schema.Value.Items, currentSchemaName, allSchemas, referencedSchemas)
	}

	// Process additionalProperties
	if schema.Value.AdditionalProperties.Has != nil && schema.Value.AdditionalProperties.Schema != nil {
		collectSchemaReferences(schema.Value.AdditionalProperties.Schema, currentSchemaName, allSchemas, referencedSchemas)
	}

	// Process allOf
	for _, subSchema := range schema.Value.AllOf {
		collectSchemaReferences(subSchema, currentSchemaName, allSchemas, referencedSchemas)
	}

	// Process oneOf
	for _, subSchema := range schema.Value.OneOf {
		collectSchemaReferences(subSchema, currentSchemaName, allSchemas, referencedSchemas)
	}

	// Process anyOf
	for _, subSchema := range schema.Value.AnyOf {
		collectSchemaReferences(subSchema, currentSchemaName, allSchemas, referencedSchemas)
	}
}

// GenerateKCLSchema creates a KCL schema from an OpenAPI schema and returns it as a string
func GenerateKCLSchema(name string, schema *openapi3.SchemaRef, allSchemas openapi3.Schemas, version OpenAPIVersion, doc *openapi3.T) (string, error) {
	log.Printf("generating KCL schema for %s (OpenAPI version: %s)", name, version)

	var builder strings.Builder
	var constraints []string

	// Track referenced schemas that need to be imported
	referencedSchemas := make(map[string]bool)

	// Add standard imports
	builder.WriteString("import regex\n")

	// Process properties
	var propertyNames []string
	for propertyName := range schema.Value.Properties {
		propertyNames = append(propertyNames, propertyName)
	}
	sort.Strings(propertyNames)

	// Add parent schemas if there are any
	parents, err := processInheritance(schema.Value, allSchemas)
	if err != nil {
		return "", err
	}

	// Add parents to referenced schemas
	for _, parent := range parents {
		if parent != name {
			referencedSchemas[parent] = true
		}
	}

	// Process properties to collect direct references
	for _, propertyName := range propertyNames {
		propSchema := schema.Value.Properties[propertyName]

		// Skip if this property is defined in a parent schema
		shouldSkip := false
		for _, parent := range parents {
			if parentSchema, ok := allSchemas[parent]; ok {
				if _, exists := parentSchema.Value.Properties[propertyName]; exists {
					shouldSkip = true
					break
				}
			}
		}
		if shouldSkip {
			continue
		}

		isRequired := false
		for _, required := range schema.Value.Required {
			if required == propertyName {
				isRequired = true
				break
			}
		}

		// Get the type and potential reference
		_, _, refType := generateFieldType(propertyName, propSchema, isRequired, name, doc)
		if refType != "" && refType != name {
			referencedSchemas[refType] = true
		}

		// Check for array items that reference schemas
		if propSchema.Value != nil && propSchema.Value.Items != nil {
			_, _, itemRefType := generateFieldType(propertyName+".items", propSchema.Value.Items, true, name, doc)
			if itemRefType != "" && itemRefType != name {
				referencedSchemas[itemRefType] = true
			}
		}
	}

	// Collect all schema references recursively - this is still useful for deeply nested references
	collectSchemaReferences(schema, name, allSchemas, referencedSchemas)

	// Add imports for all referenced schemas
	// var referencedSchemaNames []string
	// for refName := range referencedSchemas {
	// 	referencedSchemaNames = append(referencedSchemaNames, refName)
	// }
	// sort.Strings(referencedSchemaNames)

	// for _, refName := range referencedSchemaNames {
	// 	formattedRefName := formatSchemaName(refName)
	// 	builder.WriteString(fmt.Sprintf("import %s\n", formattedRefName))
	// }

	// Add a newline after imports
	builder.WriteString("\n")

	// Add KCL schema definition
	builder.WriteString(fmt.Sprintf("schema %s:", name))

	// Add schema documentation if available
	if schema.Value.Description != "" || schema.Value.Title != "" {
		builder.WriteString("\n    ")
		builder.WriteString(FormatDocumentation(schema.Value))
	}

	// Add parent schemas if there are any
	if len(parents) > 0 {
		builder.WriteString(fmt.Sprintf("\n    mixin [%s]", strings.Join(parents, ", ")))
	}

	// Process properties
	propCount := 0
	for _, propertyName := range propertyNames {
		propSchema := schema.Value.Properties[propertyName]

		// Skip if this property is defined in a parent schema
		shouldSkip := false
		for _, parent := range parents {
			if parentSchema, ok := allSchemas[parent]; ok {
				if _, exists := parentSchema.Value.Properties[propertyName]; exists {
					shouldSkip = true
					break
				}
			}
		}
		if shouldSkip {
			continue
		}

		isRequired := false
		for _, required := range schema.Value.Required {
			if required == propertyName {
				isRequired = true
				break
			}
		}

		kcltypeName, isCircular, _ := generateFieldType(propertyName, propSchema, isRequired, name, doc)

		// Add a comment about circular references if needed
		documentation := ""
		if propSchema.Value != nil && propSchema.Value.Description != "" {
			documentation = "# " + propSchema.Value.Description + "\n    "
		} else if isCircular {
			documentation = "# Circular reference to " + name + "\n    "
		}

		// Format the field with name and type
		fieldFormatted := propertyName
		if !isRequired {
			fieldFormatted += "?"
		}
		fieldFormatted += ": " + kcltypeName

		// Add default value if present
		if propSchema.Value != nil && propSchema.Value.Default != nil {
			// Format the default value based on its type
			var defaultStr string
			switch v := propSchema.Value.Default.(type) {
			case string:
				// For strings, check if it's an enum value
				if len(propSchema.Value.Enum) > 0 {
					// Verify the default is a valid enum value
					isValidEnum := false
					for _, enumVal := range propSchema.Value.Enum {
						if enumStr, ok := enumVal.(string); ok && enumStr == v {
							isValidEnum = true
							break
						}
					}
					if !isValidEnum {
						log.Printf("warning: default value '%s' is not a valid enum value for field %s", v, propertyName)
					}
				}
				defaultStr = fmt.Sprintf("\"%s\"", v)
			case float64:
				// Handle both integer and float defaults
				if v == float64(int64(v)) {
					// For integers, check if it's an enum value
					if len(propSchema.Value.Enum) > 0 {
						// Verify the default is a valid enum value
						isValidEnum := false
						for _, enumVal := range propSchema.Value.Enum {
							if enumFloat, ok := enumVal.(float64); ok && enumFloat == v {
								isValidEnum = true
								break
							}
						}
						if !isValidEnum {
							log.Printf("warning: default value '%d' is not a valid enum value for field %s", int64(v), propertyName)
						}
					}
					defaultStr = fmt.Sprintf("%d", int64(v))
				} else {
					defaultStr = fmt.Sprintf("%f", v)
				}
			case bool:
				// For booleans, check if it's an enum value
				if len(propSchema.Value.Enum) > 0 {
					// Verify the default is a valid enum value
					isValidEnum := false
					for _, enumVal := range propSchema.Value.Enum {
						if enumBool, ok := enumVal.(bool); ok && enumBool == v {
							isValidEnum = true
							break
						}
					}
					if !isValidEnum {
						log.Printf("warning: default value '%v' is not a valid enum value for field %s", v, propertyName)
					}
				}
				defaultStr = fmt.Sprintf("%v", v)
			default:
				defaultStr = fmt.Sprintf("%v", v)
			}
			fieldFormatted += " = " + defaultStr
		}

		builder.WriteString(fmt.Sprintf("\n    %s%s", documentation, fieldFormatted))

		// Collect constraints for later
		propConstraints := GenerateConstraints(propSchema.Value, propertyName, false)
		if propConstraints != nil && len(propConstraints) > 0 {
			constraints = append(constraints, propConstraints...)
		}

		propCount++
	}

	// If no properties, but we still need to handle various cases:
	// 1. Schema might have allOf/oneOf/anyOf
	// 2. Schema might be an enum
	// 3. Schema might be a primitive type with constraints
	// 4. Schema might be empty but referenced by other schemas
	if propCount == 0 {
		// Check if it's truly an empty schema or just lacks properties
		hasRefs := len(parents) > 0
		hasType := schema.Value.Type != nil && len(*schema.Value.Type) > 0
		hasEnum := len(schema.Value.Enum) > 0

		if hasRefs || hasType || hasEnum || schema.Value.Description != "" {
			// This is a valid schema without properties - keep it
			builder.WriteString("\n    # Schema without properties but with valid type or references")

			// If it's a schema with a specific type but no props, add type info
			if hasType {
				openAPIType := (*schema.Value.Type)[0]
				kclType := ConvertTypeToKCL(openAPIType, schema.Value.Format)
				builder.WriteString(fmt.Sprintf("\n    type: %q", kclType))
			}
		} else {
			builder.WriteString("\n    # No properties defined")
		}
	}

	// Add check block for constraints if we have any
	if len(constraints) > 0 {
		builder.WriteString("\n\n    check:")
		for _, constraint := range constraints {
			builder.WriteString(fmt.Sprintf("\n        %s", constraint))
		}
	}

	log.Printf("generated %d properties for schema %s", propCount, name)
	return builder.String(), nil
}

// generateFieldType determines the appropriate KCL type for a field
func generateFieldType(fieldName string, fieldSchema *openapi3.SchemaRef, isRequired bool, schemaName string, doc *openapi3.T) (string, bool, string) {
	if fieldSchema == nil || fieldSchema.Value == nil {
		return "any", false, ""
	}

	// Handle references
	if fieldSchema.Ref != "" {
		log.Printf("field %s has reference: %s", fieldName, fieldSchema.Ref)
		refName := extractSchemaName(fieldSchema.Ref)
		log.Printf("extracted reference name: %s", refName)

		// Format the reference name as a KCL schema name
		formattedRef := formatSchemaName(refName)
		log.Printf("formatted reference name: %s", formattedRef)

		// Check for self-reference (circular dependency)
		if refName == schemaName || formattedRef == schemaName {
			log.Printf("detected self-reference for field %s to schema %s", fieldName, schemaName)
			// For self-references, make them optional to break the cycle
			// But don't add the optional marker here if the field is already optional
			// KCL doesn't support double question marks like: field?: Type?
			if isRequired {
				// Only add the ? to the type for required fields
				return formattedRef + "?", false, refName
			}
			// For non-required fields, the field name will already have a ? suffix,
			// so don't add another one to the type
			return formattedRef, false, refName
		}

		// For other references, we need to handle how KCL imports work
		// When KCL imports a schema X and uses type X from it, it auto-prefixes it to X.X
		// So instead of returning formattedRef (which would result in duplicated names),
		// we return the rawRefName which will get properly qualified by KCL at import time
		log.Printf("using reference %s for field %s", formattedRef, fieldName)
		return refName, false, refName
	}

	isComplexType := false
	var fieldType string
	var refType string

	// Extract type information
	if fieldSchema.Value.Type != nil && len(*fieldSchema.Value.Type) > 0 {
		openAPIType := (*fieldSchema.Value.Type)[0]
		log.Printf("field %s has type: %s", fieldName, openAPIType)

		// Handle different types
		switch openAPIType {
		case "array":
			if fieldSchema.Value.Items != nil {
				log.Printf("processing array items for field %s", fieldName)
				// Get the item type - preserve references to other schemas
				itemType, _, refName := generateFieldType("item", fieldSchema.Value.Items, true, schemaName, doc)
				log.Printf("array item type for field %s: %s (ref: %s)", fieldName, itemType, refName)

				// For referenced schema types in arrays, we want to keep the reference
				// rather than using a generic type
				fieldType = fmt.Sprintf("[%s]", itemType)
				refType = refName // Pass along the reference type
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

	return fieldType, isComplexType, refType
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
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the file path
	filePath := filepath.Join(outputDir, name+".k")

	// Write the content to the file
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	log.Printf("Schema %s written to %s", name, filePath)
	return nil
}

// camelToSnake converts a camelCase string to snake_case
func camelToSnake(s string) string {
	var result strings.Builder
	for i, c := range s {
		if i > 0 && unicode.IsUpper(c) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(c))
	}
	return result.String()
}

// generateMainK creates a main.k file that imports all schemas for validation
func generateMainK(outputDir string, topLevelSchemas []string, allSchemas []string) error {
	var mainBuilder strings.Builder

	// Add comment header
	mainBuilder.WriteString("# This file is generated for KCL validation - DO NOT EDIT\n\n")

	// Add required standard imports
	mainBuilder.WriteString("import regex\n\n")

	// Import all schemas to ensure validation works properly
	// schemaMap := make(map[string]bool)
	// for _, schema := range allSchemas {
	// 	schemaMap[schema] = true
	// }

	// Add imports for all schemas
	// for _, schema := range allSchemas {
	// 	mainBuilder.WriteString(fmt.Sprintf("import %s\n", schema))
	// }
	// mainBuilder.WriteString("\n")

	// Create validation schema
	mainBuilder.WriteString("schema ValidationSchema:\n")

	// Add validation schema properties with optional field references to all schemas
	mainBuilder.WriteString("    # Validation schema to verify relationships between all generated schemas\n")

	// Add fields for all schemas
	for _, schema := range allSchemas {
		fieldName := camelToSnake(schema) + "_instance"
		mainBuilder.WriteString(fmt.Sprintf("    %s?: %s\n", fieldName, schema))
	}

	// Write the main.k file
	mainKPath := filepath.Join(outputDir, "main.k")
	if err := os.WriteFile(mainKPath, []byte(mainBuilder.String()), 0600); err != nil {
		return fmt.Errorf("failed to write main.k: %v", err)
	}

	if debugMode {
		fmt.Printf("  Main.k written to %s\n", mainKPath)
	}

	return nil
}

// GenerateTestMainK is a test helper function
func GenerateTestMainK(outputDir string, schemas []string) error {
	// For test cases, use the schemas as both top-level and all schemas
	return generateMainK(outputDir, schemas, schemas)
}

// extractSchemaReference extracts the reference type from a property in the original OpenAPI spec
func extractSchemaReference(doc *openapi3.T, schemaName string, propertyName string) string {
	// Check if the schema exists in the original document
	if doc.Components == nil || doc.Components.Schemas == nil {
		return ""
	}

	schema, ok := doc.Components.Schemas[schemaName]
	if !ok || schema.Value == nil {
		return ""
	}

	// Check if the property exists and has a reference
	prop, ok := schema.Value.Properties[propertyName]
	if !ok || prop.Value == nil {
		return ""
	}

	// If the property has a direct reference
	if prop.Ref != "" {
		return extractSchemaName(prop.Ref)
	}

	// If the property is an array, check if the items have a reference
	if prop.Value.Type != nil && len(*prop.Value.Type) > 0 && (*prop.Value.Type)[0] == "array" {
		if prop.Value.Items != nil && prop.Value.Items.Ref != "" {
			return extractSchemaName(prop.Value.Items.Ref)
		}
	}

	return ""
}
