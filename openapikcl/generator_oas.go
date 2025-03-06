package openapikcl

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// generateOpenAPISchemas handles KCL generation from OpenAPI schemas
func generateOpenAPISchemas(doc *openapi3.T, outputDir string, packageName string, version OpenAPIVersion) error {
	log.Printf("processing OpenAPI schema (version: %s)", version)

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

	// Generate a main.k file that imports all schemas to handle circular dependencies
	if err := generateMainFile(outputDir, packageName, schemaNames); err != nil {
		return fmt.Errorf("failed to generate main.k file: %w", err)
	}

	log.Printf("successfully generated KCL schemas in %s", outputDir)
	return nil
}

// collectSchemas returns a sorted list of schema names from the components
func collectSchemas(schemas openapi3.Schemas) []string {
	var schemaNames []string
	for name := range schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)
	return schemaNames
}

// collectSchemaReferences recursively processes a schema to identify all referenced schemas
func collectSchemaReferences(schema *openapi3.SchemaRef, currentSchemaName string, allSchemas openapi3.Schemas, referencedSchemas map[string]bool) {
	if schema == nil || schema.Value == nil {
		return
	}

	// Check for direct reference
	if schema.Ref != "" {
		refName := extractSchemaName(schema.Ref)
		if refName != currentSchemaName && refName != "" {
			referencedSchemas[refName] = true
		}
		return
	}

	// Process properties
	for _, propSchema := range schema.Value.Properties {
		collectSchemaReferences(propSchema, currentSchemaName, allSchemas, referencedSchemas)
	}

	// Process array items
	if schema.Value.Items != nil {
		collectSchemaReferences(schema.Value.Items, currentSchemaName, allSchemas, referencedSchemas)
	}

	// Process AllOf, OneOf, AnyOf
	for _, subSchema := range schema.Value.AllOf {
		collectSchemaReferences(subSchema, currentSchemaName, allSchemas, referencedSchemas)
	}

	for _, subSchema := range schema.Value.OneOf {
		collectSchemaReferences(subSchema, currentSchemaName, allSchemas, referencedSchemas)
	}

	for _, subSchema := range schema.Value.AnyOf {
		collectSchemaReferences(subSchema, currentSchemaName, allSchemas, referencedSchemas)
	}
}

// GenerateKCLSchema generates a KCL schema from an OpenAPI schema
func GenerateKCLSchema(name string, schema *openapi3.SchemaRef, allSchemas openapi3.Schemas, version OpenAPIVersion, doc *openapi3.T) (string, error) {
	var sb strings.Builder

	// Import only regex, not other schemas - they are in the same directory
	sb.WriteString("# No schema imports needed - schemas in same directory\n\n")

	// Track referenced schemas that need to be imported
	referencedSchemas := make(map[string]bool)

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
	var referencedSchemaNames []string
	for refName := range referencedSchemas {
		referencedSchemaNames = append(referencedSchemaNames, refName)
	}
	sort.Strings(referencedSchemaNames)

	// We no longer need to import schemas since they're in the same directory
	// and we're using direct references

	// Add a newline after imports
	sb.WriteString("\n")

	// Add KCL schema definition
	sb.WriteString(fmt.Sprintf("schema %s:", name))

	// Add schema documentation if available
	if schema.Value.Description != "" || schema.Value.Title != "" {
		sb.WriteString("\n    ")
		sb.WriteString(FormatDocumentation(schema.Value))
	}

	// Add parent schemas if there are any
	if len(parents) > 0 {
		sb.WriteString(fmt.Sprintf("\n    mixin [%s]", strings.Join(parents, ", ")))
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

		sb.WriteString(fmt.Sprintf("\n    %s%s", documentation, fieldFormatted))

		// Collect constraints for later
		propConstraints := GenerateConstraints(propSchema.Value, propertyName, false)
		if propConstraints != nil && len(propConstraints) > 0 {
			sb.WriteString("\n    check:")
			for _, constraint := range propConstraints {
				sb.WriteString(fmt.Sprintf("\n        %s", constraint))
			}
		}

		propCount++
	}

	// If no properties, add a placeholder comment (not 'pass')
	if propCount == 0 {
		sb.WriteString("\n    # No properties defined")
	}

	log.Printf("generated %d properties for schema %s", propCount, name)
	return sb.String(), nil
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
				return formattedRef + "?", true, refName
			}
			// For non-required fields, the field name will already have a ? suffix,
			// so don't add another one to the type
			return formattedRef, true, refName
		}

		// For other references, use the schema name directly since we're not using imports anymore
		log.Printf("using reference %s for field %s", formattedRef, fieldName)
		return formattedRef, false, refName
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
	if schema == nil || len(schema.AllOf) == 0 {
		return nil, nil
	}

	var parents []string
	for _, parentRef := range schema.AllOf {
		if parentRef.Ref != "" {
			// Extract the parent schema name from the reference
			parentName := extractSchemaName(parentRef.Ref)
			if parentName != "" {
				parents = append(parents, formatSchemaName(parentName))
			}
		}
	}

	return parents, nil
}
