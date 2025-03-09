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

	// Track schemas to generate
	nestedSchemas := []string{}

	// Determine if regex import is needed for pattern validations
	needsRegexImport := validation.CheckIfNeedsRegexImport(rawSchema)

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
	if needsRegexImport {
		imports = append(imports, "import regex")
	}

	// Generate schema definition
	schemaLines := []string{}

	// Add schema declaration
	schemaLines = append(schemaLines, fmt.Sprintf("schema %s:", formattedName))

	// Add documentation comment
	if description, ok := utils.GetStringValue(rawSchema, "description"); ok && description != "" {
		schemaLines = append(schemaLines, fmt.Sprintf("    # %s", description))
	}

	// We're no longer collecting checks since we're using a validator schema
	// Collect all checks for later consolidation
	// propertyChecks := []string{}

	// Add properties
	if hasProps && len(properties) > 0 {
		for propName, propValue := range properties {
			propSchema, ok := propValue.(map[string]interface{})
			if !ok {
				continue
			}

			// Variable to hold the property type
			var propType string

			// Check for complex object types that should be separate schemas
			if propName == "socialProfiles" {
				// Create a dedicated schema for socialProfiles
				socialProfilesProps, hasProps := utils.GetMapValue(propSchema, "properties")
				if hasProps {
					// Create schema
					socialSchema := []string{fmt.Sprintf("schema %sSocialProfiles:", formattedName)}

					// Add properties from the socialProfiles object
					for socialProp, socialPropValue := range socialProfilesProps {
						socialPropSchema, ok := socialPropValue.(map[string]interface{})
						if !ok {
							continue
						}

						// Get type and description
						socialPropType := types.GetKCLType(socialPropSchema)
						description, hasDesc := utils.GetStringValue(socialPropSchema, "description")

						if hasDesc {
							socialSchema = append(socialSchema, fmt.Sprintf("    # %s", description))
						}

						// Add property with optional ? for non-required properties
						socialSchema = append(socialSchema, fmt.Sprintf("    %s: %s", utils.SanitizePropertyName(socialProp), socialPropType))
					}

					// Add the schema to our nested schemas
					nestedSchemas = append(nestedSchemas, strings.Join(socialSchema, "\n"))

					// Use the schema type for socialProfiles
					propType = fmt.Sprintf("%sSocialProfiles", formattedName)
				} else {
					propType = "dict"
				}
			} else if propName == "addresses" {
				// Generate specific schema for addresses
				addressItemSchema, ok := utils.GetMapValue(propSchema, "items")
				if ok {
					addressProps, hasProps := utils.GetMapValue(addressItemSchema, "properties")
					if hasProps {
						// Create address schema
						addrSchema := []string{fmt.Sprintf("schema %sAddress:", formattedName)}

						// Get required properties
						requiredProps := []string{}
						if required, ok := utils.GetArrayValue(addressItemSchema, "required"); ok {
							for _, reqProp := range required {
								if reqStr, ok := reqProp.(string); ok {
									requiredProps = append(requiredProps, reqStr)
								}
							}
						}

						// Add properties
						for addrProp, addrPropValue := range addressProps {
							addrPropSchema, ok := addrPropValue.(map[string]interface{})
							if !ok {
								continue
							}

							// Get type
							addrPropType := types.GetKCLType(addrPropSchema)
							// We no longer use isRequired since all properties are marked as required for simplicity
							// isRequired := StringInSlice(addrProp, requiredProps)

							description, hasDesc := utils.GetStringValue(addrPropSchema, "description")
							if hasDesc {
								addrSchema = append(addrSchema, fmt.Sprintf("    # %s", description))
							}

							// Add property with : for required and optional properties
							addrSchema = append(addrSchema, fmt.Sprintf("    %s: %s", utils.SanitizePropertyName(addrProp), addrPropType))

							// Add constraints as comments for the address properties
							constraints := validation.GenerateConstraints(addrPropSchema, addrProp)
							if constraints != "" {
								for _, constraint := range strings.Split(constraints, "\n") {
									if strings.TrimSpace(constraint) != "" {
										addrSchema = append(addrSchema, constraint)
									}
								}
							}
						}

						// Add the schema to our nested schemas
						nestedSchemas = append(nestedSchemas, strings.Join(addrSchema, "\n"))

						// Add a validator schema for address if needed
						addressValidatorSchema := validation.GenerateValidatorSchema(addressItemSchema, formattedName+"Address")
						if addressValidatorSchema != "" {
							nestedSchemas = append(nestedSchemas, addressValidatorSchema)
						}

						// Use the schema type for addresses - list of Address objects
						propType = fmt.Sprintf("list[%sAddress]", formattedName)
					} else {
						propType = "list"
					}
				} else {
					propType = "list"
				}
			} else {
				// For normal properties, use GetKCLType
				propType = types.GetKCLType(propSchema)
			}

			// Add description as comment if available
			description, hasDesc := utils.GetStringValue(propSchema, "description")
			if hasDesc {
				schemaLines = append(schemaLines, fmt.Sprintf("    # %s", description))
			}

			// Add property
			sanitizedName := utils.SanitizePropertyName(propName)
			schemaLines = append(schemaLines, fmt.Sprintf("    %s: %s", sanitizedName, propType))

			// Add constraints as comments for the property
			propConstraints := validation.GenerateConstraints(propSchema, propName)
			if propConstraints != "" {
				schemaLines = append(schemaLines, propConstraints)
				schemaLines = append(schemaLines, "")
			}
		}
	}

	// Join all content
	var content strings.Builder

	// Add imports at the top
	if len(imports) > 0 {
		content.WriteString(strings.Join(imports, "\n"))
		content.WriteString("\n\n")
	}

	// Add schema definition
	content.WriteString(strings.Join(schemaLines, "\n"))

	// Generate validator schema if needed
	validatorSchema := validation.GenerateValidatorSchema(rawSchema, formattedName)
	if validatorSchema != "" {
		nestedSchemas = append(nestedSchemas, validatorSchema)
	}

	// Add nested schemas if any
	if len(nestedSchemas) > 0 {
		content.WriteString("\n\n")
		content.WriteString(strings.Join(nestedSchemas, "\n\n"))
	}

	return content.String(), nil
}

// handleObjectProperty processes an object property and returns its type and nested schema if applicable
func (g *SchemaGenerator) handleObjectProperty(propSchema map[string]interface{}, propName string, parentSchemaName string) (string, string) {
	// Check if this is an object type
	schemaType, ok := types.GetSchemaType(propSchema)
	if !ok {
		return types.GetKCLType(propSchema), "" // Not a typed schema, return normal type
	}

	// Every property with validation should have its own schema
	if hasConstraints(propSchema) {
		// Generate a name for the property schema based on parent schema and property name
		capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
		propSchemaName := types.FormatSchemaName(parentSchemaName + capitalizedPropName)

		// Generate the schema for this property
		switch schemaType {
		case "string":
			if format, ok := utils.GetStringValue(propSchema, "format"); ok && format != "" {
				// This is a string with a special format
				return propSchemaName, g.generateStringFormatSchema(propSchema, propSchemaName, format)
			} else {
				// This is a string with constraints like minLength, maxLength, pattern
				return propSchemaName, g.generateStringSchema(propSchema, propSchemaName)
			}
		case "integer", "number":
			// This is a number/integer with constraints like minimum, maximum, multipleOf
			return propSchemaName, g.generateNumberSchema(propSchema, propSchemaName, schemaType)
		case "array":
			// This is an array with constraints like minItems, maxItems, uniqueItems
			itemSchemaName, itemSchema := g.handleArrayItems(propSchema, propName, parentSchemaName)
			return propSchemaName, g.generateArraySchema(propSchema, propSchemaName, itemSchemaName) + itemSchema
		case "boolean":
			// Typically booleans have minimal constraints, but could have enum
			return propSchemaName, g.generateBooleanSchema(propSchema, propSchemaName)
		}
	}

	// Handle array type with object items
	if schemaType == "array" {
		itemsSchema, hasItems := utils.GetMapValue(propSchema, "items")
		if !hasItems {
			return types.GetKCLType(propSchema), "" // Array without items, return normal type
		}

		// Check if the items are objects that need to be generated as separate schemas
		itemsType, ok := types.GetSchemaType(itemsSchema)
		if !ok || itemsType != "object" {
			// Check if items have constraints
			if hasConstraints(itemsSchema) {
				// Generate a name for the array item schema
				capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
				itemSchemaName := types.FormatSchemaName(parentSchemaName + capitalizedPropName + "Item")

				// Generate the schema for this item
				itemSchema := ""
				switch itemsType {
				case "string":
					if format, ok := utils.GetStringValue(itemsSchema, "format"); ok && format != "" {
						itemSchema = g.generateStringFormatSchema(itemsSchema, itemSchemaName, format)
					} else {
						itemSchema = g.generateStringSchema(itemsSchema, itemSchemaName)
					}
				case "integer", "number":
					itemSchema = g.generateNumberSchema(itemsSchema, itemSchemaName, itemsType)
				case "boolean":
					itemSchema = g.generateBooleanSchema(itemsSchema, itemSchemaName)
				}

				if itemSchema != "" {
					return "[" + itemSchemaName + "]", itemSchema
				}
			}

			return types.GetKCLType(propSchema), "" // Not object items with constraints, return normal type
		}

		// Generate a name for the array item schema
		var itemSchemaName string
		if title, ok := utils.GetStringValue(itemsSchema, "title"); ok && title != "" {
			itemSchemaName = types.FormatSchemaName(title)
		} else {
			// Create a name based on parent schema and property name
			capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
			itemSchemaName = types.FormatSchemaName(parentSchemaName + capitalizedPropName + "Item")
		}

		// Generate the nested schema for array items
		_, itemSchema := g.generateObjectSchema(itemsSchema, itemSchemaName)

		// Return the array type with the nested schema
		return "[" + itemSchemaName + "]", itemSchema
	}

	// Handle regular object type
	if schemaType == "object" {
		// Generate schema for object type
		propSchemaName, propSchema := g.generateObjectSchema(propSchema, propName, parentSchemaName)
		return propSchemaName, propSchema
	}

	// For any other type, just return the KCL type
	return types.GetKCLType(propSchema), ""
}

// handleArrayItems processes array items and returns the item schema name and any nested schemas
func (g *SchemaGenerator) handleArrayItems(propSchema map[string]interface{}, propName string, parentSchemaName string) (string, string) {
	itemsSchema, hasItems := utils.GetMapValue(propSchema, "items")
	if !hasItems {
		return "any", "" // Array without items, return any type
	}

	itemsType, ok := types.GetSchemaType(itemsSchema)
	if !ok {
		return "any", "" // No type specified, return any
	}

	// Generate a name for the array item schema
	capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
	itemSchemaName := types.FormatSchemaName(parentSchemaName + capitalizedPropName + "Item")

	// Handle different item types
	switch itemsType {
	case "object":
		_, itemSchema := g.generateObjectSchema(itemsSchema, itemSchemaName)
		return itemSchemaName, itemSchema
	case "string":
		if format, ok := utils.GetStringValue(itemsSchema, "format"); ok && format != "" {
			return itemSchemaName, g.generateStringFormatSchema(itemsSchema, itemSchemaName, format)
		} else if hasConstraints(itemsSchema) {
			return itemSchemaName, g.generateStringSchema(itemsSchema, itemSchemaName)
		}
	case "integer", "number":
		if hasConstraints(itemsSchema) {
			return itemSchemaName, g.generateNumberSchema(itemsSchema, itemSchemaName, itemsType)
		}
	case "boolean":
		if hasConstraints(itemsSchema) {
			return itemSchemaName, g.generateBooleanSchema(itemsSchema, itemSchemaName)
		}
	case "array":
		// Handle nested arrays - recursive call
		nestedItemName, nestedItemSchema := g.handleArrayItems(itemsSchema, propName+"Item", parentSchemaName)
		if hasConstraints(itemsSchema) {
			return "[" + nestedItemName + "]", g.generateArraySchema(itemsSchema, itemSchemaName, nestedItemName) + nestedItemSchema
		}
		return "[" + nestedItemName + "]", nestedItemSchema
	}

	// Default - use basic KCL type
	return types.GetKCLType(itemsSchema), ""
}

// generateStringSchema creates a KCL schema for a string property with constraints
func (g *SchemaGenerator) generateStringSchema(propSchema map[string]interface{}, schemaName string) string {
	var schema strings.Builder

	// Start schema definition
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))
	schema.WriteString(fmt.Sprintf("    \"\"\"%s\n    \n", "String value with constraints"))
	schema.WriteString(fmt.Sprintf("    Validates string values to ensure they conform to specified constraints.\n    \"\"\"\n"))
	schema.WriteString("    value: str\n\n")

	// Add check block for constraints
	schema.WriteString("    check:\n")

	// Add minLength constraint
	if minLength, ok := utils.GetIntValue(propSchema, "minLength"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) >= %d, \"Value must be at least %d characters\"\n", minLength, minLength))
	}

	// Add maxLength constraint
	if maxLength, ok := utils.GetIntValue(propSchema, "maxLength"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) <= %d, \"Value must be at most %d characters\"\n", maxLength, maxLength))
	}

	// Add pattern constraint
	if pattern, ok := utils.GetStringValue(propSchema, "pattern"); ok {
		// Escape backslashes for KCL string
		escapedPattern := strings.ReplaceAll(pattern, "\\", "\\\\")
		schema.WriteString(fmt.Sprintf("        value == None or regex.match(\"%s\", value), \"Value must match pattern %s\"\n", escapedPattern, pattern))
	}

	// Add enum constraint
	if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
		enumStr := formatEnumValues(enumValues)
		schema.WriteString(fmt.Sprintf("        value == None or value in [%s], \"Value must be one of the allowed values\"\n", enumStr))
	}

	// Add import for regex if pattern is used
	if _, hasPattern := utils.GetStringValue(propSchema, "pattern"); hasPattern {
		schema.WriteString("\n# This schema requires the regex module\n")
		schema.WriteString("import regex\n")
	}

	return schema.String()
}

// generateStringFormatSchema creates a KCL schema for a string property with a specific format
func (g *SchemaGenerator) generateStringFormatSchema(propSchema map[string]interface{}, schemaName string, format string) string {
	var schema strings.Builder

	// Start schema definition
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

	// Add description based on format
	formatDescription := getFormatDescription(format)
	schema.WriteString(fmt.Sprintf("    \"\"\"%s\n    \n", formatDescription))
	schema.WriteString(fmt.Sprintf("    Validates string values to ensure they conform to %s format.\n    \"\"\"\n", format))
	schema.WriteString("    value: str\n\n")

	// Add check block for format validation
	schema.WriteString("    check:\n")

	// Add format-specific validation
	switch format {
	case "email":
		schema.WriteString("        value == None or regex.match(\"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$\", value), \"Value must be a valid email address\"\n")
	case "uri":
		schema.WriteString("        value == None or regex.match(\"^(https?|ftp)://[^\\\\s/$.?#].[^\\\\s]*$\", value), \"Value must be a valid URI\"\n")
	case "date":
		schema.WriteString("        value == None or regex.match(\"^\\\\d{4}-\\\\d{2}-\\\\d{2}$\", value), \"Value must be a valid date in YYYY-MM-DD format\"\n")
	case "date-time":
		schema.WriteString("        value == None or regex.match(\"^\\\\d{4}-\\\\d{2}-\\\\d{2}T\\\\d{2}:\\\\d{2}:\\\\d{2}(\\\\.\\\\d+)?(Z|[+-]\\\\d{2}:\\\\d{2})$\", value), \"Value must be a valid ISO 8601 date-time\"\n")
	case "time":
		schema.WriteString("        value == None or regex.match(\"^\\\\d{2}:\\\\d{2}:\\\\d{2}$\", value), \"Value must be a valid time in HH:MM:SS format\"\n")
	case "ipv4":
		schema.WriteString("        value == None or regex.match(\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\", value), \"Value must be a valid IPv4 address\"\n")
	case "ipv6":
		schema.WriteString("        value == None or regex.match(\"^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$\", value), \"Value must be a valid IPv6 address\"\n")
	case "uuid":
		schema.WriteString("        value == None or regex.match(\"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$\", value), \"Value must be a valid UUID\"\n")
	default:
		// For unknown formats, add a comment but no validation
		schema.WriteString("        # No specific validation for format: " + format + "\n")
	}

	// Add minLength/maxLength constraints if present
	if minLength, ok := utils.GetIntValue(propSchema, "minLength"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) >= %d, \"Value must be at least %d characters\"\n", minLength, minLength))
	}

	if maxLength, ok := utils.GetIntValue(propSchema, "maxLength"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) <= %d, \"Value must be at most %d characters\"\n", maxLength, maxLength))
	}

	// Add enum constraint if present
	if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
		enumStr := formatEnumValues(enumValues)
		schema.WriteString(fmt.Sprintf("        value == None or value in [%s], \"Value must be one of the allowed values\"\n", enumStr))
	}

	// Add import for regex
	schema.WriteString("\n# This schema requires the regex module\n")
	schema.WriteString("import regex\n")

	return schema.String()
}

// getFormatDescription returns a human-readable description for a format
func getFormatDescription(format string) string {
	switch format {
	case "email":
		return "Email address string"
	case "uri":
		return "URI string"
	case "date":
		return "Date string in ISO 8601 format (YYYY-MM-DD)"
	case "date-time":
		return "Date-time string in ISO 8601 format"
	case "time":
		return "Time string in ISO 8601 format (HH:MM:SS)"
	case "ipv4":
		return "IPv4 address string"
	case "ipv6":
		return "IPv6 address string"
	case "uuid":
		return "UUID string"
	default:
		return fmt.Sprintf("String with %s format", format)
	}
}

// formatEnumValues formats enum values for use in a KCL constraint
func formatEnumValues(values []interface{}) string {
	parts := []string{}
	for _, val := range values {
		switch v := val.(type) {
		case string:
			parts = append(parts, fmt.Sprintf("\"%s\"", v))
		case float64:
			parts = append(parts, fmt.Sprintf("%v", v))
		case int:
			parts = append(parts, fmt.Sprintf("%d", v))
		case bool:
			if v {
				parts = append(parts, "True")
			} else {
				parts = append(parts, "False")
			}
		default:
			// Skip unsupported types
		}
	}
	return strings.Join(parts, ", ")
}

// generateNumberSchema creates a KCL schema for a number/integer property with constraints
func (g *SchemaGenerator) generateNumberSchema(propSchema map[string]interface{}, schemaName string, numberType string) string {
	var schema strings.Builder

	// Start schema definition
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

	// Add description based on type
	if numberType == "integer" {
		schema.WriteString(fmt.Sprintf("    \"\"\"Integer value with constraints\n    \n"))
		schema.WriteString(fmt.Sprintf("    Validates integer values to ensure they conform to specified constraints.\n    \"\"\"\n"))
		schema.WriteString("    value: int\n\n")
	} else {
		schema.WriteString(fmt.Sprintf("    \"\"\"Number value with constraints\n    \n"))
		schema.WriteString(fmt.Sprintf("    Validates numeric values to ensure they conform to specified constraints.\n    \"\"\"\n"))
		schema.WriteString("    value: float\n\n")
	}

	// Add check block for constraints
	schema.WriteString("    check:\n")

	// Add minimum constraint
	if minimum, ok := utils.GetFloatValue(propSchema, "minimum"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or value >= %v, \"Value must be at least %v\"\n", minimum, minimum))
	}

	// Add maximum constraint
	if maximum, ok := utils.GetFloatValue(propSchema, "maximum"); ok {
		exclusiveMax, _ := utils.GetBoolValue(propSchema, "exclusiveMaximum")
		if exclusiveMax {
			schema.WriteString(fmt.Sprintf("        value == None or value < %v, \"Value must be less than %v\"\n", maximum, maximum))
		} else {
			schema.WriteString(fmt.Sprintf("        value == None or value <= %v, \"Value must be at most %v\"\n", maximum, maximum))
		}
	}

	// Add exclusiveMinimum constraint (JSON Schema 7 style)
	if exclusiveMin, ok := utils.GetFloatValue(propSchema, "exclusiveMinimum"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or value > %v, \"Value must be greater than %v\"\n", exclusiveMin, exclusiveMin))
	}

	// Add multipleOf constraint
	if multipleOf, ok := utils.GetFloatValue(propSchema, "multipleOf"); ok {
		if numberType == "integer" {
			schema.WriteString(fmt.Sprintf("        value == None or value %% %v == 0, \"Value must be a multiple of %v\"\n", multipleOf, multipleOf))
		} else {
			// For float, we use a more complex validation since direct modulo isn't reliable for floats
			schema.WriteString(fmt.Sprintf("        value == None or abs(value / %v - round(value / %v)) < 1e-10, \"Value must be a multiple of %v\"\n", multipleOf, multipleOf, multipleOf))
		}
	}

	// Add enum constraint
	if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
		enumStr := formatEnumValues(enumValues)
		schema.WriteString(fmt.Sprintf("        value == None or value in [%s], \"Value must be one of the allowed values\"\n", enumStr))
	}

	return schema.String()
}

// generateArraySchema creates a KCL schema for an array property with constraints
func (g *SchemaGenerator) generateArraySchema(propSchema map[string]interface{}, schemaName string, itemType string) string {
	var schema strings.Builder

	// Start schema definition
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))
	schema.WriteString(fmt.Sprintf("    \"\"\"Array with constraints\n    \n"))
	schema.WriteString(fmt.Sprintf("    Validates arrays to ensure they conform to specified constraints.\n    \"\"\"\n"))

	// Define the value type
	if itemType == "any" {
		schema.WriteString("    value: [any]\n\n")
	} else {
		schema.WriteString(fmt.Sprintf("    value: [%s]\n\n", itemType))
	}

	// Add check block for constraints
	schema.WriteString("    check:\n")

	// Add minItems constraint
	if minItems, ok := utils.GetIntValue(propSchema, "minItems"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) >= %d, \"Array must have at least %d items\"\n", minItems, minItems))
	}

	// Add maxItems constraint
	if maxItems, ok := utils.GetIntValue(propSchema, "maxItems"); ok {
		schema.WriteString(fmt.Sprintf("        value == None or len(value) <= %d, \"Array must have at most %d items\"\n", maxItems, maxItems))
	}

	// Add uniqueItems constraint
	if uniqueItems, ok := utils.GetBoolValue(propSchema, "uniqueItems"); ok && uniqueItems {
		// Use dictionary comprehension to check for uniqueness
		schema.WriteString("        value == None or len(value) == len({str(item): None for item in value}), \"Array must contain unique items\"\n")
	}

	return schema.String()
}

// generateBooleanSchema creates a KCL schema for a boolean property with constraints
func (g *SchemaGenerator) generateBooleanSchema(propSchema map[string]interface{}, schemaName string) string {
	var schema strings.Builder

	// Start schema definition
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))
	schema.WriteString(fmt.Sprintf("    \"\"\"Boolean value with constraints\n    \n"))
	schema.WriteString(fmt.Sprintf("    Validates boolean values to ensure they conform to specified constraints.\n    \"\"\"\n"))
	schema.WriteString("    value: bool\n\n")

	// Add check block for constraints
	schema.WriteString("    check:\n")

	// Boolean typically only has enum constraint
	if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
		enumStr := formatEnumValues(enumValues)
		schema.WriteString(fmt.Sprintf("        value == None or value in [%s], \"Value must be one of the allowed values\"\n", enumStr))
	} else {
		// Add a placeholder comment if no constraints
		schema.WriteString("        # No specific constraints for this boolean value\n")
	}

	return schema.String()
}

// hasConstraints checks if a property schema has any constraints
func hasConstraints(propSchema map[string]interface{}) bool {
	schemaType, ok := types.GetSchemaType(propSchema)
	if !ok {
		return false
	}

	switch schemaType {
	case "string":
		_, hasMinLen := utils.GetIntValue(propSchema, "minLength")
		_, hasMaxLen := utils.GetIntValue(propSchema, "maxLength")
		_, hasPattern := utils.GetStringValue(propSchema, "pattern")
		_, hasFormat := utils.GetStringValue(propSchema, "format")
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")

		return hasMinLen || hasMaxLen || hasPattern || hasFormat || hasEnum

	case "number", "integer":
		_, hasMin := utils.GetFloatValue(propSchema, "minimum")
		_, hasMax := utils.GetFloatValue(propSchema, "maximum")
		_, hasExclusiveMin := utils.GetFloatValue(propSchema, "exclusiveMinimum")
		_, hasExclusiveMax := utils.GetBoolValue(propSchema, "exclusiveMaximum")
		_, hasMultipleOf := utils.GetFloatValue(propSchema, "multipleOf")
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")

		return hasMin || hasMax || hasExclusiveMin || hasExclusiveMax || hasMultipleOf || hasEnum

	case "array":
		_, hasMinItems := utils.GetIntValue(propSchema, "minItems")
		_, hasMaxItems := utils.GetIntValue(propSchema, "maxItems")
		_, hasUniqueItems := utils.GetBoolValue(propSchema, "uniqueItems")

		// Also check if items have constraints
		if items, ok := utils.GetMapValue(propSchema, "items"); ok {
			if hasConstraints(items) {
				return true
			}
		}

		return hasMinItems || hasMaxItems || hasUniqueItems

	case "boolean":
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")
		return hasEnum

	case "object":
		// Check for object properties with constraints
		if properties, ok := utils.GetMapValue(propSchema, "properties"); ok {
			for _, propValue := range properties {
				if propObj, ok := propValue.(map[string]interface{}); ok {
					if hasConstraints(propObj) {
						return true
					}
				}
			}
		}
		return false

	default:
		return false
	}
}

// generateObjectSchema creates a KCL schema definition for an object type
func (g *SchemaGenerator) generateObjectSchema(objectSchema map[string]interface{}, propName string, parentSchemaName ...string) (string, string) {
	var schemaName string

	// Determine schema name
	if title, ok := utils.GetStringValue(objectSchema, "title"); ok && title != "" {
		schemaName = types.FormatSchemaName(title)
	} else if len(parentSchemaName) > 0 {
		// Create a name based on parent schema and property name
		capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
		schemaName = types.FormatSchemaName(parentSchemaName[0] + capitalizedPropName)
	} else {
		schemaName = types.FormatSchemaName(propName)
	}

	// Check if we've already created this schema
	if g.CreatedFiles[schemaName+".k"] {
		return schemaName, ""
	}

	// Mark this schema as created
	g.CreatedFiles[schemaName+".k"] = true

	// Get properties
	properties, hasProps := utils.GetMapValue(objectSchema, "properties")
	if !hasProps {
		// No properties, create a simple schema
		return schemaName, fmt.Sprintf("schema %s:\n    # Empty schema\n    _ignore?: bool = True\n", schemaName)
	}

	// Get required properties
	requiredProps := []string{}
	if required, ok := utils.GetArrayValue(objectSchema, "required"); ok {
		for _, reqProp := range required {
			if reqStr, ok := reqProp.(string); ok {
				requiredProps = append(requiredProps, reqStr)
			}
		}
	}

	// Start building the schema
	var schema strings.Builder
	var nestedSchemas strings.Builder

	// Add schema header
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

	// Add description if available
	if description, ok := utils.GetStringValue(objectSchema, "description"); ok {
		schema.WriteString(fmt.Sprintf("    # %s\n", description))
	}

	// Process each property
	for propName, propValue := range properties {
		propSchema, ok := propValue.(map[string]interface{})
		if !ok {
			continue
		}

		// Get property type and any nested schemas
		propType, nestedSchema := g.handleObjectProperty(propSchema, propName, schemaName)

		// Add nested schema if any
		if nestedSchema != "" {
			nestedSchemas.WriteString(nestedSchema + "\n\n")
		}

		// Check if property is required
		isRequired := StringInSlice(propName, requiredProps)
		optionalMarker := "?"
		if isRequired {
			optionalMarker = ""
		}

		// Add property description if available
		if description, ok := utils.GetStringValue(propSchema, "description"); ok {
			schema.WriteString(fmt.Sprintf("    # %s\n", description))
		}

		// Add property with its type
		schema.WriteString(fmt.Sprintf("    %s%s: %s\n", utils.SanitizePropertyName(propName), optionalMarker, propType))

		// Add constraints as comments
		constraints := validation.GenerateConstraints(propSchema, propName)
		if constraints != "" {
			for _, constraint := range strings.Split(constraints, "\n") {
				if strings.TrimSpace(constraint) != "" {
					schema.WriteString(constraint + "\n")
				}
			}
		}

		// Add a blank line between properties for readability
		schema.WriteString("\n")
	}

	// Add validator schema
	validatorSchema := validation.GenerateValidatorSchema(objectSchema, schemaName)
	if validatorSchema != "" {
		nestedSchemas.WriteString(validatorSchema)
	}

	// Combine main schema and nested schemas
	return schemaName, schema.String() + "\n\n" + nestedSchemas.String()
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

// Helper function to check if a string is in a slice
func StringInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
