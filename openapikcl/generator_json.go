package openapikcl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// generateJSONSchemas handles KCL generation from JSON Schema
func generateJSONSchemas(rawSchema map[string]interface{}, outputDir string, packageName string) error {
	log.Printf("processing JSON Schema")

	// Create output directory based on packageName if outputDir is empty
	if outputDir == "" {
		outputDir = packageName
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Find the root schema
	rootName := "Schema"
	if title, ok := rawSchema["title"].(string); ok && title != "" {
		rootName = formatSchemaName(title)
	}

	// Compile the JSON Schema
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json.Marshal(rawSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON schema: %w", err)
	}

	// Extract default values before compilation
	defaultValues := make(map[string]interface{})
	if props, ok := rawSchema["properties"].(map[string]interface{}); ok {
		for propName, propData := range props {
			if propObj, ok := propData.(map[string]interface{}); ok {
				if defaultVal, hasDefault := propObj["default"]; hasDefault {
					defaultValues[propName] = defaultVal
				}
			}
		}
	}

	// Add the schema to the compiler
	schemaID := "root-schema"
	err = compiler.AddResource(schemaID, bytes.NewReader(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile the schema
	schema, err := compiler.Compile(schemaID)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	// Generate KCL for the root schema
	kclSchema, err := generateJSONSchemaToKCLWithDefaults(rootName, schema, defaultValues)
	if err != nil {
		return fmt.Errorf("failed to generate KCL schema for %s: %w", rootName, err)
	}

	// Write the schema to a file
	schemaPath := filepath.Join(outputDir, rootName+".k")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	err = os.WriteFile(schemaPath, []byte(kclSchema), 0644)
	if err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	// Generate a simple main.k file
	mainContent := fmt.Sprintf(`# KCL schema generated from JSON Schema

# Import regex for pattern validation
import regex

schema ValidationSchema:
    # This schema can be used to validate instances
    # Example: myInstance: %s
    _ignore?: bool = True # Empty schema
`, rootName)
	mainPath := filepath.Join(outputDir, "main.k")
	err = os.WriteFile(mainPath, []byte(mainContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write main.k file: %w", err)
	}

	// Create a simple validation test file
	validationContent := fmt.Sprintf(`
# Basic validation test 
# Create an instance of %s schema
instance = {
    name = "Test User"
    age = 30
    isActive = True
    status = "active"
}

# Validate it against the %s schema
check_instance = %s {
    name = instance.name
    age = instance.age
    isActive = instance.isActive
    status = instance.status
}
`, rootName, rootName, rootName)
	validationPath := filepath.Join(outputDir, "validation_test.k")
	err = os.WriteFile(validationPath, []byte(validationContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write validation_test.k file: %w", err)
	}

	log.Printf("successfully generated KCL schema in %s", outputDir)
	return nil
}

// generateJSONSchemaToKCL converts a JSON Schema to KCL
func generateJSONSchemaToKCL(name string, schema *jsonschema.Schema) (string, error) {
	log.Printf("generating KCL schema for %s from JSON Schema", name)

	var builder strings.Builder
	var constraints []string

	// No schema imports needed - schemas in same directory
	builder.WriteString("# No schema imports needed - schemas in same directory\n\n")

	// Start schema definition
	builder.WriteString(fmt.Sprintf("schema %s:", name))

	// Add schema documentation if available
	if schema.Title != "" || schema.Description != "" {
		builder.WriteString("\n    ")
		if schema.Title != "" {
			builder.WriteString(fmt.Sprintf("# %s\n    ", schema.Title))
		}
		if schema.Description != "" {
			// Format multiline descriptions for KCL comment syntax
			lines := strings.Split(schema.Description, "\n")
			for _, line := range lines {
				builder.WriteString(fmt.Sprintf("# %s\n    ", line))
			}
		}
	}

	// Extract properties
	properties := make(map[string]*jsonschema.Schema)
	required := make(map[string]bool)

	// Extract properties from the schema
	if schema.Properties != nil {
		for name, propSchema := range schema.Properties {
			properties[name] = propSchema
		}
	}

	// Extract required properties
	for _, req := range schema.Required {
		required[req] = true
	}

	// Process properties in sorted order
	var propertyNames []string
	for propName := range properties {
		propertyNames = append(propertyNames, propName)
	}
	sort.Strings(propertyNames)

	propCount := 0
	for _, propName := range propertyNames {
		propSchema := properties[propName]
		isRequired := required[propName]

		// Generate the KCL type
		kclType := jsonSchemaTypeToKCL(propSchema)

		// Format the field
		fieldFormatted := propName
		if !isRequired {
			fieldFormatted += "?"
		}
		fieldFormatted += ": " + kclType

		// Add default value if present
		if propSchema.Default != nil {
			// Format the default value based on its type
			var defaultStr string
			switch v := propSchema.Default.(type) {
			case string:
				defaultStr = fmt.Sprintf("\"%s\"", v)
			case float64:
				// Handle both integer and float defaults
				if v == float64(int64(v)) {
					defaultStr = fmt.Sprintf("%d", int64(v))
				} else {
					defaultStr = fmt.Sprintf("%f", v)
				}
			case bool:
				defaultStr = fmt.Sprintf("%v", v)
			default:
				defaultStr = fmt.Sprintf("%v", v)
			}
			fieldFormatted += " = " + defaultStr
		}

		// Add documentation if available
		documentation := ""
		if propSchema.Description != "" {
			documentation = "# " + propSchema.Description + "\n    "
		}

		builder.WriteString(fmt.Sprintf("\n    %s%s", documentation, fieldFormatted))

		// Generate constraints for this property
		propConstraints := generateJSONSchemaConstraints(propSchema, propName)
		if len(propConstraints) > 0 {
			constraints = append(constraints, propConstraints...)
		}

		propCount++
	}

	// If no properties, add a placeholder comment
	if propCount == 0 {
		builder.WriteString("\n    # No properties defined")
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

// generateJSONSchemaToKCLWithDefaults converts a JSON Schema to KCL with supplied default values
func generateJSONSchemaToKCLWithDefaults(name string, schema *jsonschema.Schema, defaultValues map[string]interface{}) (string, error) {
	log.Printf("generating KCL schema for %s from JSON Schema with defaults", name)

	var builder strings.Builder
	var constraints []string

	// No schema imports needed - schemas in same directory
	builder.WriteString("# No schema imports needed - schemas in same directory\n\n")

	// Start schema definition
	builder.WriteString(fmt.Sprintf("schema %s:", name))

	// Add schema documentation if available
	if schema.Title != "" || schema.Description != "" {
		builder.WriteString("\n    ")
		if schema.Title != "" {
			builder.WriteString(fmt.Sprintf("# %s\n    ", schema.Title))
		}
		if schema.Description != "" {
			// Format multiline descriptions for KCL comment syntax
			lines := strings.Split(schema.Description, "\n")
			for _, line := range lines {
				builder.WriteString(fmt.Sprintf("# %s\n    ", line))
			}
		}
	}

	// Extract properties
	properties := make(map[string]*jsonschema.Schema)
	required := make(map[string]bool)

	// Extract properties from the schema
	if schema.Properties != nil {
		for name, propSchema := range schema.Properties {
			properties[name] = propSchema
		}
	}

	// Extract required properties
	for _, req := range schema.Required {
		required[req] = true
	}

	// Process properties in sorted order
	var propertyNames []string
	for propName := range properties {
		propertyNames = append(propertyNames, propName)
	}
	sort.Strings(propertyNames)

	propCount := 0
	for _, propName := range propertyNames {
		propSchema := properties[propName]
		isRequired := required[propName]

		// Generate the KCL type
		kclType := jsonSchemaTypeToKCL(propSchema)

		// Format the field
		fieldFormatted := propName
		if !isRequired {
			fieldFormatted += "?"
		}
		fieldFormatted += ": " + kclType

		// Add default value if present in our defaults map
		if defaultVal, hasDefault := defaultValues[propName]; hasDefault {
			// Format the default value based on its type
			var defaultStr string
			switch v := defaultVal.(type) {
			case string:
				defaultStr = fmt.Sprintf("\"%s\"", v)
			case float64:
				// Handle both integer and float defaults
				if v == float64(int64(v)) {
					defaultStr = fmt.Sprintf("%d", int64(v))
				} else {
					defaultStr = fmt.Sprintf("%f", v)
				}
			case bool:
				// Use KCL-style booleans (True/False with capital first letter)
				if v {
					defaultStr = "True"
				} else {
					defaultStr = "False"
				}
			default:
				defaultStr = fmt.Sprintf("%v", v)
			}
			fieldFormatted += " = " + defaultStr
		} else if propSchema.Default != nil {
			// Add default value if present in schema (though this shouldn't happen with jsonschema library)
			var defaultStr string
			switch v := propSchema.Default.(type) {
			case string:
				defaultStr = fmt.Sprintf("\"%s\"", v)
			case float64:
				if v == float64(int64(v)) {
					defaultStr = fmt.Sprintf("%d", int64(v))
				} else {
					defaultStr = fmt.Sprintf("%f", v)
				}
			case bool:
				if v {
					defaultStr = "True"
				} else {
					defaultStr = "False"
				}
			default:
				defaultStr = fmt.Sprintf("%v", v)
			}
			fieldFormatted += " = " + defaultStr
		}

		// Add documentation if available
		documentation := ""
		if propSchema.Description != "" {
			documentation = "# " + propSchema.Description + "\n    "
		}

		builder.WriteString(fmt.Sprintf("\n    %s%s", documentation, fieldFormatted))

		// Generate constraints for this property
		propConstraints := generateJSONSchemaConstraints(propSchema, propName)
		if len(propConstraints) > 0 {
			constraints = append(constraints, propConstraints...)
		}

		propCount++
	}

	// If no properties, add a placeholder comment
	if propCount == 0 {
		builder.WriteString("\n    # No properties defined")
	}

	// Add check block for constraints if we have any
	if len(constraints) > 0 {
		builder.WriteString("\n\n    check:")
		for _, constraint := range constraints {
			builder.WriteString(fmt.Sprintf("\n        %s", constraint))
		}
	}

	builder.WriteString("\n")
	log.Printf("generated %d properties for schema %s", propCount, name)
	return builder.String(), nil
}

// jsonSchemaTypeToKCL converts a JSON Schema type to a KCL type
func jsonSchemaTypeToKCL(schema *jsonschema.Schema) string {
	// Handle arrays
	if containsType(schema.Types, "array") {
		// Get the item type if specified
		if schema.Items != nil {
			// Type assertion for schema.Items
			if itemSchema, ok := schema.Items.(*jsonschema.Schema); ok {
				itemType := jsonSchemaTypeToKCL(itemSchema)
				return fmt.Sprintf("[%s]", itemType)
			}
			return "[any]"
		}
		return "[any]"
	}

	// Handle objects
	if containsType(schema.Types, "object") {
		// If it has properties, it's a complex object
		if len(schema.Properties) > 0 {
			return "dict"
		}
		return "dict"
	}

	// Handle primitive types
	if containsType(schema.Types, "string") {
		return "str"
	}
	if containsType(schema.Types, "integer") {
		return "int"
	}
	if containsType(schema.Types, "number") {
		return "float"
	}
	if containsType(schema.Types, "boolean") {
		return "bool"
	}
	if containsType(schema.Types, "null") {
		// KCL doesn't have a direct null type, but we could use nullable
		return "any"
	}

	// Handle multiple types (e.g., ["string", "null"])
	if len(schema.Types) > 1 {
		// For now, just return 'any' for multiple types
		return "any"
	}

	// Default to any if no type is specified
	return "any"
}

// containsType checks if a type is in the list of types
func containsType(types []string, typeToCheck string) bool {
	for _, t := range types {
		if t == typeToCheck {
			return true
		}
	}
	return false
}

// generateJSONSchemaConstraints creates KCL constraint expressions for a JSON Schema
func generateJSONSchemaConstraints(schema *jsonschema.Schema, fieldName string) []string {
	var constraints []string
	kclFieldRef := fieldName

	// String constraints
	if schema.MinLength > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", kclFieldRef, schema.MinLength))
	}
	if schema.MaxLength > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", kclFieldRef, schema.MaxLength))
	}
	if schema.Pattern != nil {
		// KCL uses regex matching
		patternStr := schema.Pattern.String()
		// Simplify the pattern if needed for KCL compatibility
		patternStr = strings.ReplaceAll(patternStr, "\\", "\\\\")
		constraints = append(constraints, fmt.Sprintf("regex.match(%s, r\"%s\")", kclFieldRef, patternStr))
	}

	// Numeric constraints - simplified to avoid complex expressions
	if schema.Minimum != nil {
		constraints = append(constraints, fmt.Sprintf("%s >= %v", kclFieldRef, schema.Minimum))
	}

	if schema.Maximum != nil {
		constraints = append(constraints, fmt.Sprintf("%s <= %v", kclFieldRef, schema.Maximum))
	}

	// Array constraints
	if schema.MinItems > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) >= %d", kclFieldRef, schema.MinItems))
	}
	if schema.MaxItems > 0 {
		constraints = append(constraints, fmt.Sprintf("len(%s) <= %d", kclFieldRef, schema.MaxItems))
	}
	if schema.UniqueItems {
		// Use isunique function in KCL
		constraints = append(constraints, fmt.Sprintf("isunique(%s)", kclFieldRef))
	}

	// Enum validation
	if len(schema.Enum) > 0 {
		values := make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			// Format the enum value based on its type
			switch value := v.(type) {
			case string:
				values[i] = fmt.Sprintf("\"%s\"", value)
			case bool:
				// Use KCL-style booleans
				if value {
					values[i] = "True"
				} else {
					values[i] = "False"
				}
			default:
				values[i] = fmt.Sprintf("%v", value)
			}
		}
		constraints = append(constraints, fmt.Sprintf("%s in [%s]", kclFieldRef, strings.Join(values, ", ")))
	}

	return constraints
}
