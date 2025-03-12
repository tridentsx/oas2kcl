// Package jsonschema provides functionality for converting JSON Schema to KCL.
package jsonschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/validation"
)

// TreeBasedGenerator generates KCL schemas from a schema tree
type TreeBasedGenerator struct {
	OutputDir       string
	GeneratedFiles  map[string]bool
	SchemaRegistry  map[string]string
	processedNodes  map[string]bool
	validatorNeeded map[string]bool
}

// NewTreeBasedGenerator creates a new TreeBasedGenerator
func NewTreeBasedGenerator(outputDir string) *TreeBasedGenerator {
	return &TreeBasedGenerator{
		OutputDir:       outputDir,
		GeneratedFiles:  make(map[string]bool),
		SchemaRegistry:  make(map[string]string),
		processedNodes:  make(map[string]bool),
		validatorNeeded: make(map[string]bool),
	}
}

// GenerateKCLSchemasFromTree generates KCL schemas from a JSON schema tree
func (g *TreeBasedGenerator) GenerateKCLSchemasFromTree(tree *SchemaTreeNode) ([]string, error) {
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// First pass: identify all schemas that need validator schemas
	g.identifyValidatorNeeds(tree)

	// Second pass: generate all schemas
	createdFiles := []string{}
	generatedFiles, err := g.generateSchemaFromNode(tree)
	if err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, generatedFiles...)

	return createdFiles, nil
}

// identifyValidatorNeeds traverses the tree to identify which schemas need validators
func (g *TreeBasedGenerator) identifyValidatorNeeds(node *SchemaTreeNode) {
	if node == nil {
		return
	}

	// Skip if already processed
	nodeKey := fmt.Sprintf("%s:%s", node.Type, node.SchemaName)
	if g.processedNodes[nodeKey] {
		return
	}
	g.processedNodes[nodeKey] = true

	// Check if this node needs a validator
	g.validatorNeeded[node.SchemaName] = g.nodeNeedsValidator(node)

	// Process child nodes
	switch node.Type {
	case Object:
		for _, propNode := range node.Properties {
			g.identifyValidatorNeeds(propNode)
		}
	case Array:
		g.identifyValidatorNeeds(node.Items)
	case AllOf, AnyOf, OneOf, Not, If, Then, Else:
		for _, subSchema := range node.SubSchemas {
			g.identifyValidatorNeeds(subSchema)
		}
	}
}

// nodeNeedsValidator checks if a node needs a validator schema
func (g *TreeBasedGenerator) nodeNeedsValidator(node *SchemaTreeNode) bool {
	// Check for constraints that would require validation
	if len(node.Constraints) > 0 {
		return true
	}

	// Specific format checks
	if node.Type == String && node.Format != "" {
		return true
	}

	// Object with required properties needs validation
	if node.Type == Object {
		if required, ok := node.RawSchema["required"].([]interface{}); ok && len(required) > 0 {
			return true
		}
	}

	// Array with constraints needs validation
	if node.Type == Array {
		if node.Items != nil && g.nodeNeedsValidator(node.Items) {
			return true
		}
	}

	return false
}

// generateSchemaFromNode generates a KCL schema from a tree node
func (g *TreeBasedGenerator) generateSchemaFromNode(node *SchemaTreeNode) ([]string, error) {
	if node == nil {
		return nil, nil
	}

	// Skip if already processed
	nodeKey := fmt.Sprintf("%s:%s", node.Type, node.SchemaName)
	if g.GeneratedFiles[nodeKey] {
		return nil, nil
	}
	g.GeneratedFiles[nodeKey] = true

	createdFiles := []string{}

	// Handle different node types
	switch node.Type {
	case Object:
		// Generate schemas for all properties first
		for _, propNode := range node.Properties {
			propFiles, err := g.generateSchemaFromNode(propNode)
			if err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, propFiles...)
		}

		// Then generate the schema for this object
		schemaFile, err := g.generateObjectSchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, schemaFile)

	case Array:
		// Generate schema for items first
		if node.Items != nil {
			itemFiles, err := g.generateSchemaFromNode(node.Items)
			if err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, itemFiles...)
		}

		// Then generate the schema for this array
		schemaFile, err := g.generateArraySchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, schemaFile)

	case String:
		schemaFile, err := g.generateStringSchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, schemaFile)

	case Number, Integer:
		schemaFile, err := g.generateNumberSchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, schemaFile)

	case Boolean:
		schemaFile, err := g.generateBooleanSchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, schemaFile)

	case AllOf, AnyOf, OneOf:
		// Generate schemas for all subschemas first
		for _, subSchema := range node.SubSchemas {
			subFiles, err := g.generateSchemaFromNode(subSchema)
			if err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, subFiles...)
		}

		// Then generate the schema for this composition
		schemaFile, err := g.generateCompositionSchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, schemaFile)

	case Reference:
		// Handling references - this might involve looking up the target
		// and generating a schema for it
		schemaFile, err := g.generateReferenceSchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, schemaFile)
	}

	// Generate validator schemas if needed
	if g.validatorNeeded[node.SchemaName] {
		validatorFile, err := g.generateValidatorSchema(node)
		if err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, validatorFile)
	}

	return createdFiles, nil
}

// generateObjectSchema generates a KCL schema for an object node
func (g *TreeBasedGenerator) generateObjectSchema(node *SchemaTreeNode) (string, error) {
	// Format schema name
	formattedName := node.SchemaName

	// Start with schema type schema
	schema := fmt.Sprintf("schema %s:\n", formattedName)

	// Add description if available
	if node.Description != "" {
		schema += fmt.Sprintf("    \"%s\"\n", node.Description)
	}

	// Check for required properties
	required := make(map[string]bool)
	if requiredProps, ok := node.RawSchema["required"].([]interface{}); ok {
		for _, prop := range requiredProps {
			if propName, ok := prop.(string); ok {
				required[propName] = true
			}
		}
	}

	// Track imports needed
	imports := []string{}
	needsRegexImport := false

	// Add properties
	for propName, propNode := range node.Properties {
		// Determine if property is required
		isRequired := required[propName]
		optionalMarker := "?"
		if isRequired {
			optionalMarker = ""
		}

		// Get property type
		propType := g.getNodeKCLType(propNode)

		// Add property to schema
		schema += fmt.Sprintf("    %s%s: %s\n", propName, optionalMarker, propType)
	}

	// Collect pattern properties
	patternProps := make(map[string]PatternProperty)
	for pattern, propNode := range node.PatternProperties {
		// Get property type from node
		nodeType := string(propNode.Type)

		// Create PatternProperty struct
		patternProp := PatternProperty{
			Pattern:     pattern,
			Schema:      propNode.RawSchema,
			Description: propNode.Description,
			Type:        nodeType,
		}

		patternProps[pattern] = patternProp
	}

	// Generate pattern property schemas if needed
	if len(patternProps) > 0 {
		// We need regex for pattern properties
		needsRegexImport = true

		// If we have pattern properties, generate the validator schema
		// Add a comment indicating pattern properties
		schema += "\n    # This schema has pattern properties that will be validated dynamically\n"

		// Generate individual validator schemas for each pattern property
		for pattern, prop := range patternProps {
			validatorName := fmt.Sprintf("%s_%s_Validator", formattedName, sanitizePatternName(pattern))

			// Generate the schema content
			patternSchema := GeneratePatternPropertySchema(validatorName, prop)

			// Write the schema to a file
			patternSchemaPath := filepath.Join(g.OutputDir, validatorName+".k")
			err := os.WriteFile(patternSchemaPath, []byte(patternSchema), 0644)
			if err != nil {
				return "", fmt.Errorf("failed to write pattern property schema: %v", err)
			}

			// Add to imports
			imports = append(imports, validatorName)
		}

		// Generate the main validator schema
		validatorName := fmt.Sprintf("%sPatternValidator", formattedName)
		validatorSchema := GeneratePatternPropertiesValidator(formattedName, patternProps, &imports)

		// Write the validator schema to a file
		validatorPath := filepath.Join(g.OutputDir, validatorName+".k")
		err := os.WriteFile(validatorPath, []byte(validatorSchema), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write pattern validator schema: %v", err)
		}

		// Add to imports
		imports = append(imports, validatorName)

		// Add validation code to check pattern properties
		if !strings.Contains(schema, "    check:") {
			schema += "\n    check:\n"
		}
		schema += fmt.Sprintf("        # Validate pattern properties\n")
		schema += fmt.Sprintf("        %sPatternValidator {\n", formattedName)
		schema += fmt.Sprintf("            data = __dict__\n")
		schema += fmt.Sprintf("        }\n")
	} else {
		// Use existing pattern property validation if no separate validators are generated
		if len(node.PatternProperties) > 0 {
			// We need regex for pattern properties
			needsRegexImport = true

			// Add a check block for pattern properties
			schema += "\n    # Pattern property validation\n"
			schema += "    check:\n"

			// Process each pattern property
			for pattern, propNode := range node.PatternProperties {
				// Get the property type
				propType := g.getNodeKCLType(propNode)

				// Add a check for this pattern property
				schema += fmt.Sprintf("        # Validate properties matching pattern: %s\n", pattern)
				schema += fmt.Sprintf("        all k, v in {k: v for k, v in __dict__ if regex.match(r\"%s\", k)} {\n", escapeRegexPattern(pattern))

				// Add type validation based on the property type
				if propType == "str" {
					schema += "            v is str\n"
				} else if propType == "int" {
					schema += "            v is int\n"
				} else if propType == "float" {
					schema += "            v is float\n"
				} else if propType == "bool" {
					schema += "            v is bool\n"
				} else if strings.HasPrefix(propType, "[") && strings.HasSuffix(propType, "]") {
					schema += "            v is list\n"
				} else {
					// For complex types, we'll just check it's a dict
					schema += "            v is dict\n"
				}

				// Add format validation if needed
				if format, ok := propNode.RawSchema["format"].(string); ok {
					switch format {
					case "email":
						schema += fmt.Sprintf("            v is str and regex.match(r\"^[a-zA-Z0-9.!#$%%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$\", v), \"Value must be a valid email address\"\n")
					case "date-time":
						schema += fmt.Sprintf("            v is str and regex.match(r\"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(?:\\.\\d+)?(?:Z|[+-]\\d{2}:\\d{2})$\", v), \"Value must be a valid date-time\"\n")
					case "uri":
						schema += fmt.Sprintf("            v is str and regex.match(r\"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]*$\", v), \"Value must be a valid URI\"\n")
					case "uuid":
						schema += fmt.Sprintf("            v is str and regex.match(r\"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$\", v), \"Value must be a valid UUID\"\n")
					case "ipv4":
						schema += fmt.Sprintf("            v is str and regex.match(r\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\", v), \"Value must be a valid IPv4 address\"\n")
					}
				}

				// Add pattern validation if needed
				if propPattern, ok := propNode.RawSchema["pattern"].(string); ok {
					schema += fmt.Sprintf("            v is str and regex.match(r\"%s\", v), \"Value must match pattern %s\"\n", escapeRegexPattern(propPattern), propPattern)
				}

				// Close the check block
				schema += "        }\n"
			}
		}
	}

	// Add validation checks for common constraints
	validationSchema := g.generateConstraintValidations(node, required)
	if validationSchema != "" {
		if !strings.Contains(schema, "    check:") {
			schema += "\n    check:\n"
		}
		schema += validationSchema
	}

	// Add imports if needed
	if needsRegexImport {
		schema = "import regex\n\n" + schema
	}

	// Add any additional imports
	for _, importName := range imports {
		schema = fmt.Sprintf("import %s\n", importName) + schema
	}

	return schema, nil
}

// generateConstraintValidations generates validation checks for common constraints
func (g *TreeBasedGenerator) generateConstraintValidations(node *SchemaTreeNode, required map[string]bool) string {
	// Check if there are constraints to validate
	if len(node.Constraints) == 0 {
		return ""
	}

	var validations strings.Builder

	// Handle minProperties constraint
	if minProps, ok := node.Constraints["minProperties"].(float64); ok {
		validations.WriteString(fmt.Sprintf("        len(__dict__) >= %d, \"Object must have at least %d properties\"\n", int(minProps), int(minProps)))
	}

	// Handle maxProperties constraint
	if maxProps, ok := node.Constraints["maxProperties"].(float64); ok {
		validations.WriteString(fmt.Sprintf("        len(__dict__) <= %d, \"Object must have at most %d properties\"\n", int(maxProps), int(maxProps)))
	}

	// Add additionalProperties validation if needed
	if additionalProps, ok := node.RawSchema["additionalProperties"].(bool); ok && !additionalProps {
		// Build a list of allowed property names
		allowedProps := make([]string, 0, len(node.Properties))
		for propName := range node.Properties {
			allowedProps = append(allowedProps, fmt.Sprintf("\"%s\"", propName))
		}

		// Add the check for additional properties
		validations.WriteString(fmt.Sprintf("        # Validate no additional properties\n"))
		validations.WriteString(fmt.Sprintf("        all k in __dict__ {\n"))
		if len(node.PatternProperties) > 0 {
			// If there are pattern properties, we need to check if the property matches any pattern
			validations.WriteString(fmt.Sprintf("            k in [%s]", strings.Join(allowedProps, ", ")))

			for pattern := range node.PatternProperties {
				validations.WriteString(fmt.Sprintf(" or regex.match(r\"%s\", k)", escapeRegexPattern(pattern)))
			}

			validations.WriteString(", \"Additional properties are not allowed\"\n")
		} else {
			// If there are no pattern properties, we can just check the list directly
			validations.WriteString(fmt.Sprintf("            k in [%s], \"Additional properties are not allowed\"\n", strings.Join(allowedProps, ", ")))
		}
		validations.WriteString(fmt.Sprintf("        }\n"))
	}

	return validations.String()
}

// escapeRegexPattern escapes special characters in a regex pattern for KCL
func escapeRegexPattern(pattern string) string {
	// Escape backslashes
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")

	// Escape double quotes
	pattern = strings.ReplaceAll(pattern, "\"", "\\\"")

	return pattern
}

// generateArraySchema generates a KCL schema for an array node
func (g *TreeBasedGenerator) generateArraySchema(node *SchemaTreeNode) (string, error) {
	var content strings.Builder
	formattedName := formatSchemaName(node.SchemaName)

	// Determine needed imports
	imports := g.determineImports(node)
	if len(imports) > 0 {
		content.WriteString(strings.Join(imports, "\n"))
		content.WriteString("\n\n")
	}

	// Schema declaration
	content.WriteString(fmt.Sprintf("schema %s:\n", formattedName))

	// Add description if available
	if node.Description != "" {
		content.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", node.Description))
	}

	// Define the array type
	var itemType string
	if node.Items != nil {
		itemType = g.getNodeKCLType(node.Items)
	} else {
		itemType = "any"
	}
	content.WriteString(fmt.Sprintf("    value: [%s]\n", itemType))

	// Add check block if there are constraints
	if len(node.Constraints) > 0 {
		content.WriteString("\n    check:\n")

		for constraint, value := range node.Constraints {
			switch constraint {
			case "minItems":
				if minItems, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        len(value) >= %d if value, \"Array must have at least %d items\"\n", int(minItems), int(minItems)))
				}
			case "maxItems":
				if maxItems, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        len(value) <= %d if value, \"Array must have at most %d items\"\n", int(maxItems), int(maxItems)))
				}
			case "uniqueItems":
				if uniqueItems, ok := value.(bool); ok && uniqueItems {
					content.WriteString("        isunique(value) if value, \"Array must have unique items\"\n")
				}
			}
		}
	}

	// Write the schema to a file
	schemaFilePath := filepath.Join(g.OutputDir, formattedName+".k")
	if err := os.WriteFile(schemaFilePath, []byte(content.String()), 0644); err != nil {
		return "", err
	}

	return schemaFilePath, nil
}

// generateStringSchema generates a KCL schema for a string node
func (g *TreeBasedGenerator) generateStringSchema(node *SchemaTreeNode) (string, error) {
	var content strings.Builder
	formattedName := formatSchemaName(node.SchemaName)

	// Determine needed imports
	imports := g.determineImports(node)
	if len(imports) > 0 {
		content.WriteString(strings.Join(imports, "\n"))
		content.WriteString("\n\n")
	}

	// Schema declaration
	content.WriteString(fmt.Sprintf("schema %s:\n", formattedName))

	// Add description if available
	if node.Description != "" {
		content.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", node.Description))
	}

	// Define the string type
	content.WriteString("    value: str\n")

	// Add check block if there are constraints
	if len(node.Constraints) > 0 || node.Format != "" {
		content.WriteString("\n    check:\n")

		// Format-specific validation
		if node.Format != "" {
			switch node.Format {
			case "email":
				content.WriteString("        regex.match(value, r\"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$\") if value, \"Value must be a valid email address\"\n")
			case "uri":
				content.WriteString("        regex.match(value, r\"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]\") if value, \"Value must be a valid URI\"\n")
			case "date-time":
				content.WriteString("        regex.match(value, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\") if value, \"Value must be a valid RFC 3339 date-time\"\n")
				content.WriteString("        value != None and value and datetime.validate(value[:10], \"%Y-%m-%d\"), \"Value contains an invalid date component\"\n")
			}
		}

		// Other constraints
		for constraint, value := range node.Constraints {
			switch constraint {
			case "minLength":
				if minLen, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        len(value) >= %d if value, \"String must be at least %d characters\"\n", int(minLen), int(minLen)))
				}
			case "maxLength":
				if maxLen, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        len(value) <= %d if value, \"String must be at most %d characters\"\n", int(maxLen), int(maxLen)))
				}
			case "pattern":
				if pattern, ok := value.(string); ok {
					content.WriteString(fmt.Sprintf("        regex.match(value, r\"%s\") if value, \"String must match pattern %s\"\n", pattern, pattern))
				}
			case "enum":
				if enum, ok := value.([]interface{}); ok && len(enum) > 0 {
					enumValues := make([]string, len(enum))
					for i, e := range enum {
						enumValues[i] = fmt.Sprintf("%v", e)
					}
					content.WriteString(fmt.Sprintf("        value in [%s] if value, \"String must be one of: %s\"\n",
						strings.Join(enumValues, ", "), strings.Join(enumValues, ", ")))
				}
			}
		}
	}

	// Write the schema to a file
	schemaFilePath := filepath.Join(g.OutputDir, formattedName+".k")
	if err := os.WriteFile(schemaFilePath, []byte(content.String()), 0644); err != nil {
		return "", err
	}

	return schemaFilePath, nil
}

// generateNumberSchema generates a KCL schema for a number/integer node
func (g *TreeBasedGenerator) generateNumberSchema(node *SchemaTreeNode) (string, error) {
	var content strings.Builder
	formattedName := formatSchemaName(node.SchemaName)

	// Type mapping
	var kclType string
	if node.Type == Integer {
		kclType = "int"
	} else {
		kclType = "float"
	}

	// Schema declaration
	content.WriteString(fmt.Sprintf("schema %s:\n", formattedName))

	// Add description if available
	if node.Description != "" {
		content.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", node.Description))
	}

	// Define the number type
	content.WriteString(fmt.Sprintf("    value: %s\n", kclType))

	// Add check block if there are constraints
	if len(node.Constraints) > 0 {
		content.WriteString("\n    check:\n")

		for constraint, value := range node.Constraints {
			switch constraint {
			case "minimum":
				if min, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        value >= %g if value, \"Value must be at least %g\"\n", min, min))
				}
			case "maximum":
				if max, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        value <= %g if value, \"Value must be at most %g\"\n", max, max))
				}
			case "exclusiveMinimum":
				if min, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        value > %g if value, \"Value must be greater than %g\"\n", min, min))
				}
			case "exclusiveMaximum":
				if max, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        value < %g if value, \"Value must be less than %g\"\n", max, max))
				}
			case "multipleOf":
				if multiple, ok := value.(float64); ok {
					content.WriteString(fmt.Sprintf("        value %% %g == 0 if value, \"Value must be a multiple of %g\"\n", multiple, multiple))
				}
			}
		}
	}

	// Write the schema to a file
	schemaFilePath := filepath.Join(g.OutputDir, formattedName+".k")
	if err := os.WriteFile(schemaFilePath, []byte(content.String()), 0644); err != nil {
		return "", err
	}

	return schemaFilePath, nil
}

// generateBooleanSchema generates a KCL schema for a boolean node
func (g *TreeBasedGenerator) generateBooleanSchema(node *SchemaTreeNode) (string, error) {
	var content strings.Builder
	formattedName := formatSchemaName(node.SchemaName)

	// Schema declaration
	content.WriteString(fmt.Sprintf("schema %s:\n", formattedName))

	// Add description if available
	if node.Description != "" {
		content.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", node.Description))
	}

	// Define the boolean type
	content.WriteString("    value: bool\n")

	// Write the schema to a file
	schemaFilePath := filepath.Join(g.OutputDir, formattedName+".k")
	if err := os.WriteFile(schemaFilePath, []byte(content.String()), 0644); err != nil {
		return "", err
	}

	return schemaFilePath, nil
}

// generateCompositionSchema generates a KCL schema for an allOf/anyOf/oneOf node
func (g *TreeBasedGenerator) generateCompositionSchema(node *SchemaTreeNode) (string, error) {
	var content strings.Builder
	formattedName := formatSchemaName(node.SchemaName)

	// Determine needed imports
	imports := g.determineImports(node)
	for _, subSchema := range node.SubSchemas {
		imports = append(imports, fmt.Sprintf("import %s", formatSchemaName(subSchema.SchemaName)))
	}
	if len(imports) > 0 {
		content.WriteString(strings.Join(imports, "\n"))
		content.WriteString("\n\n")
	}

	// Schema declaration
	content.WriteString(fmt.Sprintf("schema %s", formattedName))

	// Handle different composition types
	switch node.Type {
	case AllOf:
		// For allOf, we can use KCL's mixin functionality
		if len(node.SubSchemas) > 0 {
			mixins := make([]string, len(node.SubSchemas))
			for i, subSchema := range node.SubSchemas {
				mixins[i] = formatSchemaName(subSchema.SchemaName)
			}
			content.WriteString(fmt.Sprintf("(%s)", strings.Join(mixins, ", ")))
		}
	case AnyOf, OneOf:
		// For anyOf/oneOf, we use a union type in KCL
		content.WriteString(":\n")
		content.WriteString("    # This represents a union type in JSON Schema\n")
		content.WriteString("    # In KCL, we use a check block to validate against any of the schemas\n")
		content.WriteString("    # The data will need to validate against at least one of the schemas\n")
		content.WriteString("\n    check:\n")
		content.WriteString("        # At least one schema must validate\n")

		validations := make([]string, len(node.SubSchemas))
		for i, subSchema := range node.SubSchemas {
			schemaName := formatSchemaName(subSchema.SchemaName)
			validations[i] = fmt.Sprintf("is_valid_%s(self)", strings.ToLower(schemaName))
		}

		content.WriteString(fmt.Sprintf("        %s, \"Data must validate against at least one schema\"\n",
			strings.Join(validations, " or ")))

		// Add helper functions for validation
		content.WriteString("\n    # Helper functions to check validation against each schema\n")
		for _, subSchema := range node.SubSchemas {
			schemaName := formatSchemaName(subSchema.SchemaName)
			content.WriteString(fmt.Sprintf("    is_valid_%s = lambda self -> bool {\n", strings.ToLower(schemaName)))
			content.WriteString(fmt.Sprintf("        schema = %s {}\n", schemaName))
			content.WriteString("        try:\n")
			content.WriteString("            # Try to validate against this schema\n")
			content.WriteString("            return True\n")
			content.WriteString("        except:\n")
			content.WriteString("            return False\n")
			content.WriteString("    }\n")
		}
	}

	// Write the schema to a file
	schemaFilePath := filepath.Join(g.OutputDir, formattedName+".k")
	if err := os.WriteFile(schemaFilePath, []byte(content.String()), 0644); err != nil {
		return "", err
	}

	return schemaFilePath, nil
}

// generateReferenceSchema generates a KCL schema for a reference node
func (g *TreeBasedGenerator) generateReferenceSchema(node *SchemaTreeNode) (string, error) {
	// TODO: Implement reference resolution
	// For now, return empty to avoid errors
	return "", nil
}

// generateValidatorSchema generates a validator schema for a node
func (g *TreeBasedGenerator) generateValidatorSchema(node *SchemaTreeNode) (string, error) {
	// Convert to validation.Schema
	valSchema := convertNodeToValidationSchema(node)

	// Generate validator content - ignore the imports return value since we don't need it
	validatorSchema, _ := validation.GenerateValidatorSchema(valSchema, node.SchemaName+"Validator")

	// Write to file
	validatorFilePath := filepath.Join(g.OutputDir, node.SchemaName+"Validator.k")
	if err := os.WriteFile(validatorFilePath, []byte(validatorSchema), 0644); err != nil {
		return "", err
	}

	return validatorFilePath, nil
}

// determineImports determines the needed imports for a node
func (g *TreeBasedGenerator) determineImports(node *SchemaTreeNode) []string {
	imports := []string{}

	// Check for regex pattern or format that needs regex
	needsRegex := false

	if node.Type == String {
		if _, ok := node.Constraints["pattern"]; ok {
			needsRegex = true
		}

		if node.Format != "" {
			needsRegex = true
			if node.Format == "date-time" {
				imports = append(imports, "import datetime")
			}
		}
	}

	if needsRegex {
		imports = append(imports, "import regex")
	}

	return imports
}

// getNodeKCLType gets the KCL type for a node
func (g *TreeBasedGenerator) getNodeKCLType(node *SchemaTreeNode) string {
	switch node.Type {
	case Object:
		return formatSchemaName(node.SchemaName)
	case Array:
		var itemType string
		if node.Items != nil {
			itemType = g.getNodeKCLType(node.Items)
		} else {
			itemType = "any"
		}
		return fmt.Sprintf("[%s]", itemType)
	case String:
		if node.Format != "" {
			switch node.Format {
			case "email":
				return "EmailValidator"
			case "uri":
				return "URIValidator"
			case "date-time":
				return "DateTimeValidator"
			case "uuid":
				return "UUIDValidator"
			case "ipv4":
				return "IPv4Validator"
			default:
				return "str"
			}
		}
		return "str"
	case Integer:
		return "int"
	case Number:
		return "float"
	case Boolean:
		return "bool"
	case Null:
		return "None"
	default:
		return "any"
	}
}

// sanitizePropertyName sanitizes a property name for KCL
func sanitizePropertyName(name string) string {
	// Replace hyphens with underscores
	name = strings.ReplaceAll(name, "-", "_")

	// Ensure name doesn't start with a number
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}

	return name
}

// formatSchemaName formats a schema name for KCL
func formatSchemaName(name string) string {
	if name == "" {
		return "Schema"
	}

	// Convert to camel case
	parts := strings.Split(name, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			r := []rune(parts[i])
			r[0] = []rune(strings.ToUpper(string(r[0])))[0]
			parts[i] = string(r)
		}
	}

	return strings.Join(parts, "")
}

// convertNodeToValidationSchema converts a SchemaTreeNode to a validation.Schema
func convertNodeToValidationSchema(node *SchemaTreeNode) *validation.Schema {
	result := &validation.Schema{
		Type: string(node.Type),
	}

	// Set format if this is a string node
	if node.Type == String && node.Format != "" {
		result.Format = node.Format
	}

	// Set constraints
	for key, value := range node.Constraints {
		switch key {
		case "minLength":
			if minLength, ok := value.(float64); ok {
				minLengthInt := int(minLength)
				result.MinLength = &minLengthInt
			}
		case "maxLength":
			if maxLength, ok := value.(float64); ok {
				maxLengthInt := int(maxLength)
				result.MaxLength = &maxLengthInt
			}
		case "pattern":
			if pattern, ok := value.(string); ok {
				result.Pattern = pattern
			}
		case "minimum":
			if minimum, ok := value.(float64); ok {
				result.Minimum = &minimum
			}
		case "maximum":
			if maximum, ok := value.(float64); ok {
				result.Maximum = &maximum
			}
		case "minItems":
			if minItems, ok := value.(float64); ok {
				minItemsInt := int(minItems)
				result.MinItems = &minItemsInt
			}
		case "maxItems":
			if maxItems, ok := value.(float64); ok {
				maxItemsInt := int(maxItems)
				result.MaxItems = &maxItemsInt
			}
		case "uniqueItems":
			if uniqueItems, ok := value.(bool); ok {
				result.UniqueItems = uniqueItems
			}
		}
	}

	// Handle array items
	if node.Type == Array && node.Items != nil {
		result.Items = convertNodeToValidationSchema(node.Items)
	}

	// Handle object properties
	if node.Type == Object {
		result.Properties = make(map[string]*validation.Schema)
		for propName, propNode := range node.Properties {
			result.Properties[propName] = convertNodeToValidationSchema(propNode)
		}

		// Set required properties
		if required, ok := node.RawSchema["required"].([]interface{}); ok {
			result.Required = make([]string, 0, len(required))
			for _, r := range required {
				if str, ok := r.(string); ok {
					result.Required = append(result.Required, str)
				}
			}
		}
	}

	return result
}

// GenerateSchemaTreeAndKCL parses a JSON schema, builds a schema tree, and generates KCL schemas
func GenerateSchemaTreeAndKCL(schemaBytes []byte, outputDir string, debugMode bool) error {
	// Parse the JSON schema
	var rawSchema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &rawSchema); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Determine the schema name from title or default to "Schema"
	schemaName := "Schema"
	if title, ok := rawSchema["title"].(string); ok && title != "" {
		schemaName = title
	}

	// Build the schema tree
	tree, err := BuildSchemaTree(rawSchema, schemaName, nil)
	if err != nil {
		return fmt.Errorf("failed to build schema tree: %w", err)
	}

	// If debug mode is enabled, print the tree structure
	if debugMode {
		PrintSchemaTree(tree, 0)
	}

	// Generate KCL schemas from the tree
	generator := NewTreeBasedGenerator(outputDir)
	_, err = generator.GenerateKCLSchemasFromTree(tree)
	if err != nil {
		return fmt.Errorf("failed to generate KCL schemas: %w", err)
	}

	return nil
}

// PrintSchemaTree prints the schema tree for debugging purposes
func PrintSchemaTree(node *SchemaTreeNode, level int) {
	if node == nil {
		return
	}

	// Create indent based on level
	indent := strings.Repeat("  ", level)

	// Print node information
	fmt.Printf("%sNode: %s (Type: %s)\n", indent, node.SchemaName, node.Type)

	if node.Description != "" {
		fmt.Printf("%s  Description: %s\n", indent, node.Description)
	}

	if node.Format != "" {
		fmt.Printf("%s  Format: %s\n", indent, node.Format)
	}

	if len(node.Constraints) > 0 {
		fmt.Printf("%s  Constraints:\n", indent)
		for key, value := range node.Constraints {
			fmt.Printf("%s    %s: %v\n", indent, key, value)
		}
	}

	// Print properties for object nodes
	if node.Type == Object && len(node.Properties) > 0 {
		fmt.Printf("%s  Properties:\n", indent)
		for propName, propNode := range node.Properties {
			fmt.Printf("%s    %s:\n", indent, propName)
			PrintSchemaTree(propNode, level+2)
		}
	}

	// Print items for array nodes
	if node.Type == Array && node.Items != nil {
		fmt.Printf("%s  Items:\n", indent)
		PrintSchemaTree(node.Items, level+1)
	}

	// Print subschemas for composition nodes
	if (node.Type == AllOf || node.Type == AnyOf || node.Type == OneOf || node.Type == Not) && len(node.SubSchemas) > 0 {
		fmt.Printf("%s  SubSchemas:\n", indent)
		for i, subSchema := range node.SubSchemas {
			fmt.Printf("%s    [%d]:\n", indent, i)
			PrintSchemaTree(subSchema, level+2)
		}
	}

	// Print reference target for reference nodes
	if node.Type == Reference && node.RefTarget != "" {
		fmt.Printf("%s  RefTarget: %s\n", indent, node.RefTarget)
	}
}
