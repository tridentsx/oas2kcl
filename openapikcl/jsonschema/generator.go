// This file is no longer needed as we are using the helpers package.
// All functionality has been moved to the helpers package and re-exported in jsonschema.go.
package jsonschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/types"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/utils"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/validation"
)

// SchemaGenerator converts JSON Schema to KCL schemas
type SchemaGenerator struct {
	RawSchema    map[string]interface{}
	OutputDir    string
	SchemaName   string
	Definitions  map[string]map[string]interface{}
	CreatedFiles map[string]bool
}

// NewSchemaGenerator creates a new SchemaGenerator
func NewSchemaGenerator(rawSchema map[string]interface{}, outputDir string) *SchemaGenerator {
	// Extract definitions
	defs := make(map[string]map[string]interface{})
	if defsMap, ok := utils.GetMapValue(rawSchema, "definitions"); ok {
		for name, schema := range defsMap {
			if schemaMap, ok := schema.(map[string]interface{}); ok {
				defs[name] = schemaMap
			}
		}
	}

	// Extract schema name from title or default to Schema
	schemaName := "Schema"
	if title, ok := utils.GetStringValue(rawSchema, "title"); ok && title != "" {
		schemaName = types.FormatSchemaName(title)
	} else if id, ok := utils.GetStringValue(rawSchema, "$id"); ok && id != "" {
		schemaName = types.FormatSchemaName(filepath.Base(id))
	}

	return &SchemaGenerator{
		RawSchema:    rawSchema,
		OutputDir:    outputDir,
		SchemaName:   schemaName,
		Definitions:  defs,
		CreatedFiles: make(map[string]bool),
	}
}

// GenerateKCLSchemas generates KCL schemas from a JSON Schema
func (g *SchemaGenerator) GenerateKCLSchemas() ([]string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	createdFiles := []string{}

	// Generate main schema
	mainSchemaContent, err := g.GenerateKCLSchema(g.RawSchema, g.SchemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate main schema: %w", err)
	}

	mainSchemaFile := filepath.Join(g.OutputDir, g.SchemaName+".k")
	if err := os.WriteFile(mainSchemaFile, []byte(mainSchemaContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write main schema file: %w", err)
	}
	createdFiles = append(createdFiles, mainSchemaFile)
	g.CreatedFiles[mainSchemaFile] = true

	// Generate schemas for definitions
	for name, defSchema := range g.Definitions {
		if g.CreatedFiles[filepath.Join(g.OutputDir, types.FormatSchemaName(name)+".k")] {
			continue // Skip if already created
		}

		schemaContent, err := g.GenerateKCLSchema(defSchema, name)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for definition %s: %w", name, err)
		}

		schemaFile := filepath.Join(g.OutputDir, types.FormatSchemaName(name)+".k")
		if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write schema file for definition %s: %w", name, err)
		}
		createdFiles = append(createdFiles, schemaFile)
		g.CreatedFiles[schemaFile] = true
	}

	return createdFiles, nil
}

// GenerateKCLSchema generates a KCL schema from a JSON Schema
func (g *SchemaGenerator) GenerateKCLSchema(rawSchema map[string]interface{}, schemaName string) (string, error) {
	// Format the schema name
	formattedName := types.FormatSchemaName(schemaName)

	// Track nested schema definitions that will be added after the main schema
	nestedSchemas := []string{}
	nestedSchemaTypes := map[string]string{} // Maps property name to schema type name

	// Build imports section
	imports := []string{}

	// Check for references to other schemas
	// We need to add imports for referenced schemas
	properties, hasProps := utils.GetMapValue(rawSchema, "properties")
	if hasProps {
		for _, propValue := range properties {
			propSchema, ok := propValue.(map[string]interface{})
			if !ok {
				continue
			}

			if ref, ok := utils.GetStringValue(propSchema, "$ref"); ok {
				refName := types.ExtractSchemaName(ref)
				// Skip self-references and standard imports
				if types.FormatSchemaName(refName) != formattedName && !strings.Contains(ref, "#/definitions/") {
					imports = append(imports, fmt.Sprintf("import %s", types.FormatSchemaName(refName)))
				}
			}
		}
	}

	// Check if regex import is needed
	if validation.CheckIfNeedsRegexImport(rawSchema) {
		imports = append(imports, "import regex")
	}

	// Build schema content
	lines := []string{}

	// Add imports if any
	if len(imports) > 0 {
		lines = append(lines, strings.Join(imports, "\n"))
		lines = append(lines, "")
	}

	// Add schema description
	description, hasDescription := utils.GetStringValue(rawSchema, "description")
	if hasDescription {
		lines = append(lines, fmt.Sprintf("# %s", description))
	}

	// Begin schema definition
	lines = append(lines, fmt.Sprintf("schema %s:", formattedName))

	// Add properties
	if hasProps && len(properties) > 0 {
		for propName, propValue := range properties {
			propSchema, ok := propValue.(map[string]interface{})
			if !ok {
				continue
			}

			// Handle nested object properties
			propType, nestedSchema := g.handleObjectProperty(propSchema, propName, formattedName)
			if nestedSchema != "" {
				nestedSchemas = append(nestedSchemas, nestedSchema)
				nestedSchemaTypes[propName] = propType
			} else {
				// Get property type for non-nested properties
				propType = types.GetKCLType(propSchema)
			}

			// Check if property is required
			required := types.IsPropertyRequired(rawSchema, propName)

			// Get property description
			propDescription, hasPropDesc := utils.GetStringValue(propSchema, "description")

			// Add property description as comment
			if hasPropDesc {
				lines = append(lines, fmt.Sprintf("    # %s", propDescription))
			}

			// Add property with type
			sanitizedName := utils.SanitizePropertyName(propName)
			if required {
				lines = append(lines, fmt.Sprintf("    %s: %s", sanitizedName, propType))
			} else {
				lines = append(lines, fmt.Sprintf("    %s?: %s", sanitizedName, propType))
			}

			// Add constraints for this property
			constraints := validation.GenerateConstraints(propSchema, propName)
			if constraints != "" {
				lines = append(lines, constraints)
			}
		}

		// Add validation for required properties at the schema level
		requiredChecks := validation.GenerateRequiredPropertyChecks(rawSchema)
		if requiredChecks != "" {
			lines = append(lines, "")
			lines = append(lines, requiredChecks)
		}
	} else {
		// If no properties, add a placeholder comment and pass statement
		lines = append(lines, "    # This schema has no properties defined")
		lines = append(lines, "    pass")
	}

	// Add nested schema definitions if any
	if len(nestedSchemas) > 0 {
		lines = append(lines, "")
		for _, nestedSchema := range nestedSchemas {
			lines = append(lines, nestedSchema)
		}
	}

	return strings.Join(lines, "\n"), nil
}

// handleObjectProperty processes an object property and returns its type and nested schema if applicable
func (g *SchemaGenerator) handleObjectProperty(propSchema map[string]interface{}, propName string, parentSchemaName string) (string, string) {
	// Check if this is an object type
	schemaType, ok := types.GetSchemaType(propSchema)
	if !ok || schemaType != "object" {
		return types.GetKCLType(propSchema), "" // Not an object, return normal type
	}

	// Check if property has nested properties
	properties, hasProps := utils.GetMapValue(propSchema, "properties")
	if !hasProps || len(properties) == 0 {
		return types.GetKCLType(propSchema), "" // Object without properties, return normal type
	}

	// Generate a name for the nested schema
	var nestedSchemaName string
	if title, ok := utils.GetStringValue(propSchema, "title"); ok && title != "" {
		nestedSchemaName = types.FormatSchemaName(title)
	} else {
		// Create a name based on parent schema and property name
		// Ensure property name has its first letter capitalized for the schema name
		capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
		nestedSchemaName = types.FormatSchemaName(parentSchemaName + capitalizedPropName)
	}

	// Generate the nested schema
	nestedSchema := fmt.Sprintf("schema %s:", nestedSchemaName)

	// Process properties of the nested object
	for childPropName, childPropValue := range properties {
		childPropSchema, ok := childPropValue.(map[string]interface{})
		if !ok {
			continue
		}

		// Get property type
		childPropType := types.GetKCLType(childPropSchema)

		// Check if property is required
		childRequired := types.IsPropertyRequired(propSchema, childPropName)

		// Get property description
		childDescription, hasChildDesc := utils.GetStringValue(childPropSchema, "description")

		// Add property description as comment if available
		if hasChildDesc {
			nestedSchema += fmt.Sprintf("\n    # %s", childDescription)
		}

		// Add property with appropriate optional marker
		sanitizedChildName := utils.SanitizePropertyName(childPropName)
		if childRequired {
			nestedSchema += fmt.Sprintf("\n    %s: %s", sanitizedChildName, childPropType)
		} else {
			nestedSchema += fmt.Sprintf("\n    %s?: %s", sanitizedChildName, childPropType)
		}

		// Add constraints for this property
		childConstraints := validation.GenerateConstraints(childPropSchema, childPropName)
		if childConstraints != "" {
			nestedSchema += childConstraints
		}
	}

	// Add required property checks for nested schema
	requiredChecks := validation.GenerateRequiredPropertyChecks(propSchema)
	if requiredChecks != "" {
		nestedSchema += fmt.Sprintf("\n\n    %s", strings.ReplaceAll(requiredChecks, "\n", "\n    "))
	}

	return nestedSchemaName, nestedSchema
}

// GenerateSchemas generates KCL schemas from a JSON Schema bytes
func GenerateSchemas(schemaBytes []byte, outputDir, packageName string) error {
	// Parse the schema as JSON
	var rawSchema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &rawSchema); err != nil {
		return err
	}

	// Create a schema generator
	generator := NewSchemaGenerator(rawSchema, outputDir)

	// Generate the schemas
	_, err := generator.GenerateKCLSchemas()
	return err
}
