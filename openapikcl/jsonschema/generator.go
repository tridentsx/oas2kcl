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
	RawSchema        map[string]interface{}
	OutputDir        string
	SchemaName       string
	Definitions      map[string]map[string]interface{}
	CreatedFiles     map[string]bool
	processedSchemas map[string]bool // Track which schemas have been processed to prevent circular references
}

// NewSchemaGenerator creates a new SchemaGenerator
func NewSchemaGenerator(rawSchema map[string]interface{}, outputDir string) *SchemaGenerator {
	// Create a schema generator
	generator := &SchemaGenerator{
		RawSchema:        rawSchema,
		OutputDir:        outputDir,
		CreatedFiles:     make(map[string]bool),
		processedSchemas: make(map[string]bool), // Initialize the map for tracking processed schemas
	}

	// Process any definitions
	if defsMap, ok := utils.GetMapValue(rawSchema, "definitions"); ok {
		defs := make(map[string]map[string]interface{})
		for name, schema := range defsMap {
			if schemaMap, ok := schema.(map[string]interface{}); ok {
				defs[name] = schemaMap
			}
		}
		generator.Definitions = defs
	}

	// Extract schema name from title or default to Schema
	schemaName := "Schema"
	if title, ok := utils.GetStringValue(rawSchema, "title"); ok && title != "" {
		schemaName = types.FormatSchemaName(title)
	} else if id, ok := utils.GetStringValue(rawSchema, "$id"); ok && id != "" {
		schemaName = types.FormatSchemaName(filepath.Base(id))
	}

	generator.SchemaName = schemaName

	return generator
}

// GenerateKCLSchemas generates KCL schemas from a JSON Schema
func (g *SchemaGenerator) GenerateKCLSchemas() ([]string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	createdFiles := []string{}

	// Generate main schema
	mainSchemaName, err := g.GenerateKCLSchema(g.RawSchema, g.SchemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate main schema: %w", err)
	}

	mainSchemaFile := filepath.Join(g.OutputDir, mainSchemaName+".k")
	createdFiles = append(createdFiles, mainSchemaFile)
	g.CreatedFiles[mainSchemaFile] = true

	// Generate schemas for definitions
	for name, defSchema := range g.Definitions {
		if g.CreatedFiles[filepath.Join(g.OutputDir, types.FormatSchemaName(name)+".k")] {
			continue // Skip if already created
		}

		schemaName, err := g.GenerateKCLSchema(defSchema, name)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for definition %s: %w", name, err)
		}

		schemaFile := filepath.Join(g.OutputDir, schemaName+".k")
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

	// Get required properties
	requiredProps := []string{}
	if requiredArray, ok := utils.GetArrayValue(rawSchema, "required"); ok {
		for _, reqProp := range requiredArray {
			if reqStr, ok := reqProp.(string); ok {
				requiredProps = append(requiredProps, reqStr)
			}
		}
	}

	// Add schema declaration
	schemaLines = append(schemaLines, fmt.Sprintf("schema %s:", formattedName))

	// Add documentation comment
	if description, ok := utils.GetStringValue(rawSchema, "description"); ok && description != "" {
		schemaLines = append(schemaLines, fmt.Sprintf("    # %s", description))
	}

	// We're no longer collecting checks since we're using a validator schema
	// Collect all checks for later consolidation
	// propertyChecks := []string{}

	// Track if we need to generate validator schemas
	needsEmailValidator := false
	needsURIValidator := false
	needsDateTimeValidator := false
	needsUUIDValidator := false
	needsIPv4Validator := false

	// Add properties
	if hasProps {
		var propLine string
		for propName, propValue := range properties {
			propSchema, ok := propValue.(map[string]interface{})
			if !ok {
				continue
			}

			// Determine if property is required
			isRequired := false
			for _, reqProp := range requiredProps {
				if reqProp == propName {
					isRequired = true
					break
				}
			}

			// Determine the property type
			var propType string
			schemaType, typeOk := types.GetSchemaType(propSchema)
			if !typeOk {
				propType = "str" // default to string for unknown types
			} else if schemaType == "array" {
				// Get item type for arrays
				items, hasItems := utils.GetMapValue(propSchema, "items")
				if hasItems {
					itemType, hasType := types.GetSchemaType(items)
					if !hasType {
						propType = "list"
					} else {
						// For specialized format items in arrays, use the corresponding validator
						if itemType == "string" {
							if format, hasFormat := utils.GetStringValue(items, "format"); hasFormat && format != "" {
								switch format {
								case "date-time":
									propType = "[DateTimeValidator]"
									needsDateTimeValidator = true
								case "email":
									propType = "[EmailValidator]"
									needsEmailValidator = true
								case "uri":
									propType = "[URIValidator]"
									needsURIValidator = true
								case "uuid":
									propType = "[UUIDValidator]"
									needsUUIDValidator = true
								case "ipv4":
									propType = "[IPv4Validator]"
									needsIPv4Validator = true
								default:
									propType = "[str]"
								}
							} else {
								propType = "[str]"
							}
						} else if itemType == "integer" {
							propType = "[int]"
						} else if itemType == "number" {
							propType = "[float]"
						} else if itemType == "boolean" {
							propType = "[bool]"
						} else if itemType == "object" {
							// For object types in arrays, we need to create a special schema
							// For simplicity, we'll just use dict here, but in a complete
							// implementation, we'd create a proper schema for the object type
							propType = "[dict]"
						} else {
							propType = "[any]"
						}
					}
				} else {
					propType = "list"
				}
			} else {
				// For non-array types, use the KCL type mapper
				propType = types.GetKCLType(propSchema)

				// For specialized format strings, use the corresponding validator
				if schemaType == "string" {
					if format, hasFormat := utils.GetStringValue(propSchema, "format"); hasFormat && format != "" {
						switch format {
						case "date-time":
							propType = "DateTimeValidator"
							needsDateTimeValidator = true
						case "email":
							propType = "EmailValidator"
							needsEmailValidator = true
						case "uri":
							propType = "URIValidator"
							needsURIValidator = true
						case "uuid":
							propType = "UUIDValidator"
							needsUUIDValidator = true
						case "ipv4":
							propType = "IPv4Validator"
							needsIPv4Validator = true
						default:
							propType = "str"
						}
					}
				}
			}

			// Add the property line
			var optionalMarker string
			if !isRequired {
				optionalMarker = "?"
			}
			propLine = fmt.Sprintf("    %s%s: %s", utils.SanitizePropertyName(propName), optionalMarker, propType)
			schemaLines = append(schemaLines, propLine)

			// Add property constraints as comments
			if schemaType == "string" {
				if minLength, ok := utils.GetIntValue(propSchema, "minLength"); ok {
					schemaLines = append(schemaLines, fmt.Sprintf("    # Min length: %d", minLength))
				}
				if maxLength, ok := utils.GetIntValue(propSchema, "maxLength"); ok {
					schemaLines = append(schemaLines, fmt.Sprintf("    # Max length: %d", maxLength))
				}
				if pattern, ok := utils.GetStringValue(propSchema, "pattern"); ok {
					schemaLines = append(schemaLines, fmt.Sprintf("    # Regex pattern: %s", pattern))
				}
				if format, ok := utils.GetStringValue(propSchema, "format"); ok {
					schemaLines = append(schemaLines, fmt.Sprintf("    # Format: %s", format))

					switch format {
					case "date-time":
						schemaLines = append(schemaLines, "    # Date-time string in ISO 8601 format")
					case "date":
						schemaLines = append(schemaLines, "    # Date string in ISO 8601 format (YYYY-MM-DD)")
					case "time":
						schemaLines = append(schemaLines, "    # Time string in ISO 8601 format (HH:MM:SS)")
					case "email":
						schemaLines = append(schemaLines, "    # Email address string")
					case "uri":
						schemaLines = append(schemaLines, "    # URI string following RFC 3986")
					case "hostname":
						schemaLines = append(schemaLines, "    # Hostname following RFC 1034")
					case "ipv4":
						schemaLines = append(schemaLines, "    # IPv4 address string")
					case "ipv6":
						schemaLines = append(schemaLines, "    # IPv6 address string")
					case "uuid":
						schemaLines = append(schemaLines, "    # UUID string representation")
					}
				}
			} else if schemaType == "array" {
				if minItems, ok := utils.GetIntValue(propSchema, "minItems"); ok {
					schemaLines = append(schemaLines, fmt.Sprintf("    # Min items: %d", minItems))
				}
				if maxItems, ok := utils.GetIntValue(propSchema, "maxItems"); ok {
					schemaLines = append(schemaLines, fmt.Sprintf("    # Max items: %d", maxItems))
				}
				if uniqueItems, ok := utils.GetBoolValue(propSchema, "uniqueItems"); ok && uniqueItems {
					schemaLines = append(schemaLines, "    # Unique items: true")
				}

				// Add item format if applicable
				items, hasItems := utils.GetMapValue(propSchema, "items")
				if hasItems {
					itemType, _ := types.GetSchemaType(items)
					if itemType == "string" {
						if format, ok := utils.GetStringValue(items, "format"); ok {
							schemaLines = append(schemaLines, fmt.Sprintf("    # Item format: %s", format))
						}
					}
				}
			}

			// Add a blank line after each property
			schemaLines = append(schemaLines, "")
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

	// Add validation checks if needed
	if hasValidationConstraints(rawSchema) || len(requiredProps) > 0 {
		content.WriteString("\n\n    check:")

		// Add validation for required properties
		for _, reqName := range requiredProps {
			sanitizedName := utils.SanitizePropertyName(reqName)
			content.WriteString(fmt.Sprintf("\n        %s != None", sanitizedName))
		}

		// Add validation for string properties
		for propName, propSchemaInterface := range properties {
			propSchema, ok := propSchemaInterface.(map[string]interface{})
			if !ok {
				continue
			}

			if schemaType, hasType := utils.GetStringValue(propSchema, "type"); hasType {
				sanitizedName := utils.SanitizePropertyName(propName)

				// String validation
				if schemaType == "string" {
					// Min/max length
					if minLength, ok := utils.GetIntValue(propSchema, "minLength"); ok {
						content.WriteString(fmt.Sprintf("\n        %s == None or len(%s) >= %d", sanitizedName, sanitizedName, minLength))
					}
					if maxLength, ok := utils.GetIntValue(propSchema, "maxLength"); ok {
						content.WriteString(fmt.Sprintf("\n        %s == None or len(%s) <= %d", sanitizedName, sanitizedName, maxLength))
					}
					// Pattern
					if pattern, ok := utils.GetStringValue(propSchema, "pattern"); ok {
						escapedPattern := strings.ReplaceAll(pattern, "\\", "\\\\")
						content.WriteString(fmt.Sprintf("\n        %s == None or regex.match(\"%s\", %s)", sanitizedName, escapedPattern, sanitizedName))
						imports = append(imports, "import regex")
					}
				}

				// Array validation
				if schemaType == "array" {
					// Min/max items
					if minItems, ok := utils.GetIntValue(propSchema, "minItems"); ok {
						content.WriteString(fmt.Sprintf("\n        %s == None or len(%s) >= %d", sanitizedName, sanitizedName, minItems))
					}
					if maxItems, ok := utils.GetIntValue(propSchema, "maxItems"); ok {
						content.WriteString(fmt.Sprintf("\n        %s == None or len(%s) <= %d", sanitizedName, sanitizedName, maxItems))
					}
					// Unique items
					if uniqueItems, ok := utils.GetBoolValue(propSchema, "uniqueItems"); ok && uniqueItems {
						content.WriteString(fmt.Sprintf("\n        %s == None or len(%s) == len({str(item): None for item in %s})", sanitizedName, sanitizedName, sanitizedName))
					}
				}
			}
		}
	}

	// Add nested schemas if any
	if len(nestedSchemas) > 0 {
		content.WriteString("\n\n")
		content.WriteString(strings.Join(nestedSchemas, "\n\n"))
	}

	// Generate validator schema if needed
	// Convert the raw schema to a validation.Schema object
	valSchema := convertToValidationSchema(rawSchema)
	validatorSchemaResult, validatorImports := validation.GenerateValidatorSchema(valSchema, formattedName)

	// Add any generated validator schemas
	if validatorSchemaResult != "" {
		nestedSchemas = append(nestedSchemas, validatorSchemaResult)
	}

	// Add imports based on schema requirements
	if validatorImports.NeedsRegex && !StringInSlice("import regex", imports) {
		imports = append(imports, "import regex")
	}
	if validatorImports.NeedsDatetime && !StringInSlice("import datetime", imports) {
		imports = append(imports, "import datetime")
	}
	if validatorImports.NeedsNet && !StringInSlice("import net", imports) {
		imports = append(imports, "import net")
	}

	// Write the main schema to a file
	schemaFilePath := filepath.Join(g.OutputDir, formattedName+".k")
	if err := os.WriteFile(schemaFilePath, []byte(content.String()), 0644); err != nil {
		return "", err
	}

	// Generate and write validator schemas if needed
	if needsEmailValidator {
		emailValidatorContent := `import regex

schema EmailValidator:
    value: str

    check:
        regex.match(value, r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$") if value, "Value must be a valid email address"
`
		emailValidatorPath := filepath.Join(g.OutputDir, "EmailValidator.k")
		if err := os.WriteFile(emailValidatorPath, []byte(emailValidatorContent), 0644); err != nil {
			return "", err
		}
	}

	if needsURIValidator {
		uriValidatorContent := `import regex

schema URIValidator:
    value: str

    check:
        # URI format validation - checks for valid URI format with scheme
        regex.match(value, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if value, "Value must be a valid URI"
`
		uriValidatorPath := filepath.Join(g.OutputDir, "URIValidator.k")
		if err := os.WriteFile(uriValidatorPath, []byte(uriValidatorContent), 0644); err != nil {
			return "", err
		}
	}

	if needsDateTimeValidator {
		dateTimeValidatorContent := `import regex
import datetime

schema DateTimeValidator:
    value: str

    check:
        # First check format with regex that enforces correct ranges
        regex.match(value, r"^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):[0-5]\d:[0-5]\d(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$") if value, "Value must be a valid RFC 3339 date-time"
        # Validate the date part to catch invalid dates like Feb 30
        value != None and value and datetime.validate(value[:10], "%Y-%m-%d"), "Value contains an invalid date component"
`
		dateTimeValidatorPath := filepath.Join(g.OutputDir, "DateTimeValidator.k")
		if err := os.WriteFile(dateTimeValidatorPath, []byte(dateTimeValidatorContent), 0644); err != nil {
			return "", err
		}
	}

	if needsUUIDValidator {
		uuidValidatorContent := `import regex

schema UUIDValidator:
    value: str

    check:
        # Validate UUID format
        regex.match(value, r"^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$") if value, "Value must be a valid UUID"
`
		uuidValidatorPath := filepath.Join(g.OutputDir, "UUIDValidator.k")
		if err := os.WriteFile(uuidValidatorPath, []byte(uuidValidatorContent), 0644); err != nil {
			return "", err
		}
	}

	if needsIPv4Validator {
		ipv4ValidatorContent := `import regex

schema IPv4Validator:
    value: str

    check:
        # Validate IPv4 format
        regex.match(value, r"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$") if value, "Value must be a valid IPv4 address"
`
		ipv4ValidatorPath := filepath.Join(g.OutputDir, "IPv4Validator.k")
		if err := os.WriteFile(ipv4ValidatorPath, []byte(ipv4ValidatorContent), 0644); err != nil {
			return "", err
		}
	}

	return formattedName, nil
}

// handleObjectProperty processes an object property and returns its type and nested schema if applicable
func (g *SchemaGenerator) handleObjectProperty(propSchema map[string]interface{}, propName string, parentSchemaName string) (string, string) {
	// Track processed schemas to avoid circular references
	if g.processedSchemas == nil {
		g.processedSchemas = make(map[string]bool)
	}

	// Check if this is an object type
	schemaType, ok := types.GetSchemaType(propSchema)
	if !ok {
		return types.GetKCLType(propSchema), "" // Not a typed schema, return normal type
	}

	// If this is an object type, generate a proper schema for it
	if schemaType == "object" {
		// Generate a name for the schema
		var schemaName string
		if title, ok := utils.GetStringValue(propSchema, "title"); ok && title != "" {
			schemaName = types.FormatSchemaName(title)
		} else {
			// Create a name based on parent schema and property name
			capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
			schemaName = types.FormatSchemaName(parentSchemaName + capitalizedPropName)
		}

		// Check for circular references
		if g.processedSchemas[schemaName] {
			return schemaName, ""
		}
		g.processedSchemas[schemaName] = true

		// Handle regular object type
		_, nestedSchema := g.generateObjectSchema(propSchema, propName, parentSchemaName)
		return schemaName, nestedSchema
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
			// For non-object items with constraints, recursively process
			if hasConstraints(itemsSchema) {
				capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]
				itemPropName := capitalizedPropName + "Item"
				itemSchemaName, itemSchema := g.handleObjectProperty(itemsSchema, itemPropName, parentSchemaName)

				// Return array of the nested type
				return "[" + itemSchemaName + "]", itemSchema
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

		// Check for circular references
		if g.processedSchemas[itemSchemaName] {
			return "[" + itemSchemaName + "]", ""
		}
		g.processedSchemas[itemSchemaName] = true

		// Generate the nested schema for array items
		_, itemSchema := g.generateObjectSchema(itemsSchema, itemSchemaName)

		// Return the array type with the nested schema
		return "[" + itemSchemaName + "]", itemSchema
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
	schema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

	formatDescription := getFormatDescription(format)
	schema.WriteString(fmt.Sprintf("    \"\"\"String with %s format\n    \n", formatDescription))
	schema.WriteString(fmt.Sprintf("    Validates strings to ensure they conform to %s format.\n    \"\"\"\n", formatDescription))
	schema.WriteString("    value: str\n\n")

	// Add validation
	schema.WriteString("    check:\n")

	// Add format-specific validation
	switch format {
	case "date-time":
		schema.WriteString("        # RFC 3339 date-time validation\n")
		schema.WriteString("        # First check format with regex that enforces correct ranges\n")
		schema.WriteString("        regex.match(value, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\") if value, \"Value must be a valid RFC 3339 date-time\"\n")
		schema.WriteString("        # Validate the date part to catch invalid dates like Feb 30\n")
		schema.WriteString("        value != None and value and datetime.validate(value[:10], \"%Y-%m-%d\"), \"Value contains an invalid date component\"\n")
	case "email":
		schema.WriteString("        # Email format validation\n")
		schema.WriteString("        regex.match(value, r\"" + utils.CommonPatterns["email"] + "\") if value, \"Value must be a valid email address\"\n")
	case "uri":
		schema.WriteString("        # URI format validation\n")
		schema.WriteString("        regex.match(value, r\"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]\") if value, \"Value must be a valid URI\"\n")
	case "uuid":
		schema.WriteString("        # UUID format validation\n")
		schema.WriteString("        regex.match(value, r\"^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$\") if value, \"Value must be a valid UUID\"\n")
	case "ipv4":
		schema.WriteString("        # IPv4 format validation\n")
		schema.WriteString("        regex.match(value, r\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\") if value, \"Value must be a valid IPv4 address\"\n")
	default:
		schema.WriteString("        # Generic format validation\n")
		schema.WriteString(fmt.Sprintf("        # This is a placeholder for %s format validation\n", format))
		schema.WriteString("        value != None, \"Value must not be None\"\n")
	}

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

	// Check if we've already created or processed this schema (prevents circular references)
	if g.CreatedFiles[schemaName+".k"] || g.processedSchemas[schemaName] {
		return schemaName, ""
	}

	// Mark this schema as created and processed
	g.CreatedFiles[schemaName+".k"] = true
	g.processedSchemas[schemaName] = true

	// Get properties
	properties, hasProps := utils.GetMapValue(objectSchema, "properties")
	if !hasProps {
		// No properties, create a properly structured empty schema
		var emptySchema strings.Builder

		// Add schema declaration
		emptySchema.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

		// Add documentation comment
		emptySchema.WriteString("    \"\"\"Empty schema with no defined properties.\n    \n")
		emptySchema.WriteString("    This schema represents an object with no specific properties defined.\n    \"\"\"\n")

		// Add a placeholder field to ensure the schema is valid KCL
		emptySchema.WriteString("    # This schema has no properties defined\n")
		emptySchema.WriteString("    _ignore?: bool = True\n")

		return schemaName, emptySchema.String()
	}

	// Get required properties
	requiredProps := []string{}
	if requiredArray, ok := utils.GetArrayValue(objectSchema, "required"); ok {
		for _, reqProp := range requiredArray {
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
		isRequired := false
		for _, req := range requiredProps {
			if req == propName {
				isRequired = true
				break
			}
		}

		optionalMarker := "?"
		if isRequired {
			optionalMarker = ""
		}

		// Add property description if available
		if description, ok := utils.GetStringValue(propSchema, "description"); ok {
			schema.WriteString(fmt.Sprintf("    # %s\n", description))
		}

		// Add property with its type
		sanitizedName := utils.SanitizePropertyName(propName)

		// If this is an object type, use the nested schema name
		if schemaType, hasType := utils.GetStringValue(propSchema, "type"); hasType && schemaType == "object" {
			// For nested objects, use the proper schema reference
			capitalizedPropName := strings.ToUpper(propName[0:1]) + propName[1:]

			// If title is defined, use it as the schema name
			var nestedSchemaName string
			if title, hasTitle := utils.GetStringValue(propSchema, "title"); hasTitle && title != "" {
				nestedSchemaName = types.FormatSchemaName(title)
			} else {
				nestedSchemaName = types.FormatSchemaName(schemaName + capitalizedPropName)
			}

			propType = nestedSchemaName
		} else if schemaType, hasType := utils.GetStringValue(propSchema, "type"); hasType && schemaType == "string" {
			// For string properties with format, use str directly instead of a reference
			if _, hasFormat := utils.GetStringValue(propSchema, "format"); hasFormat {
				propType = "str"
			}
		}

		schema.WriteString(fmt.Sprintf("    %s%s: %s\n", sanitizedName, optionalMarker, propType))

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

	// Convert the object schema to a validation.Schema object
	valSchema := convertToValidationSchema(objectSchema)
	validatorSchemaResult, _ := validation.GenerateValidatorSchema(valSchema, schemaName)

	// Add any generated validator schemas
	if validatorSchemaResult != "" {
		nestedSchemas.WriteString(validatorSchemaResult)
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

// Add this helper function
func hasValidationConstraints(schema map[string]interface{}) bool {
	// Check if there are required properties
	if required, ok := utils.GetArrayValue(schema, "required"); ok && len(required) > 0 {
		return true
	}

	properties, hasProps := utils.GetMapValue(schema, "properties")
	if !hasProps {
		return false
	}

	for _, propSchemaInterface := range properties {
		propSchema, ok := propSchemaInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Check for string constraints
		if schemaType, hasType := utils.GetStringValue(propSchema, "type"); hasType {
			if schemaType == "string" {
				if _, hasMinLength := utils.GetIntValue(propSchema, "minLength"); hasMinLength {
					return true
				}
				if _, hasMaxLength := utils.GetIntValue(propSchema, "maxLength"); hasMaxLength {
					return true
				}
				if _, hasPattern := utils.GetStringValue(propSchema, "pattern"); hasPattern {
					return true
				}
			}

			// Check for array constraints
			if schemaType == "array" {
				if _, hasMinItems := utils.GetIntValue(propSchema, "minItems"); hasMinItems {
					return true
				}
				if _, hasMaxItems := utils.GetIntValue(propSchema, "maxItems"); hasMaxItems {
					return true
				}
				if uniqueItems, hasUnique := utils.GetBoolValue(propSchema, "uniqueItems"); hasUnique && uniqueItems {
					return true
				}
			}
		}
	}

	return false
}

func getKCLType(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Check for format types
		if strings.Contains(v, "T") && strings.Contains(v, "Z") {
			return "str" // datetime format
		}
		if strings.Contains(v, ".") && len(strings.Split(v, ".")) == 4 {
			return "str" // ip format
		}
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
			return "str" // uri format
		}
		return "str"
	case float64:
		if v == float64(int64(v)) {
			return "int"
		}
		return "float"
	case bool:
		return "bool"
	case []interface{}:
		return "any"
	case map[string]interface{}:
		return "any"
	default:
		return "any"
	}
}

func (g *SchemaGenerator) generateConstraints(schema map[string]interface{}, indent string) string {
	var constraints strings.Builder

	// Handle format constraints
	if format, ok := schema["format"].(string); ok {
		switch format {
		case "date-time":
			constraints.WriteString(fmt.Sprintf("%scheck:\n%s    datetime.parse(value) != None\n", indent, indent))
		case "ipv4":
			constraints.WriteString(fmt.Sprintf("%scheck:\n%s    value.split('.') | len == 4\n%s    all v in value.split('.') { int(v) >= 0 and int(v) <= 255 }\n", indent, indent, indent))
		case "uri":
			constraints.WriteString(fmt.Sprintf("%scheck:\n%s    value.startswith('http://') or value.startswith('https://')\n", indent, indent))
		}
	}

	// Handle pattern constraints
	if pattern, ok := schema["pattern"].(string); ok {
		// Escape special characters in the pattern
		escapedPattern := strings.ReplaceAll(pattern, "\\", "\\\\")
		constraints.WriteString(fmt.Sprintf("%scheck:\n%s    value.match(r'%s') != None\n", indent, indent, escapedPattern))
	}

	// ... existing code ...

	return constraints.String()
}

// convertToValidationSchema converts a raw JSON schema to a validation.Schema object
func convertToValidationSchema(rawSchema map[string]interface{}) *validation.Schema {
	result := &validation.Schema{}

	// Set schema type
	if schemaType, ok := utils.GetStringValue(rawSchema, "type"); ok {
		result.Type = schemaType
	}

	// Set format if this is a string schema
	if result.Type == "string" {
		if format, ok := utils.GetStringValue(rawSchema, "format"); ok {
			result.Format = format
		}
	}

	// Set pattern if this is a string schema
	if result.Type == "string" {
		if pattern, ok := utils.GetStringValue(rawSchema, "pattern"); ok {
			result.Pattern = pattern
		}
	}

	// Set min/max length if this is a string schema
	if result.Type == "string" {
		if minLength, ok := utils.GetIntValue(rawSchema, "minLength"); ok {
			minLengthInt := int(minLength)
			result.MinLength = &minLengthInt
		}
		if maxLength, ok := utils.GetIntValue(rawSchema, "maxLength"); ok {
			maxLengthInt := int(maxLength)
			result.MaxLength = &maxLengthInt
		}
	}

	// Set min/max value if this is a number or integer schema
	if result.Type == "number" || result.Type == "integer" {
		if minimum, ok := utils.GetFloatValue(rawSchema, "minimum"); ok {
			result.Minimum = &minimum
		}
		if maximum, ok := utils.GetFloatValue(rawSchema, "maximum"); ok {
			result.Maximum = &maximum
		}
	}

	// Set array constraints if this is an array schema
	if result.Type == "array" {
		if minItems, ok := utils.GetIntValue(rawSchema, "minItems"); ok {
			minItemsInt := int(minItems)
			result.MinItems = &minItemsInt
		}
		if maxItems, ok := utils.GetIntValue(rawSchema, "maxItems"); ok {
			maxItemsInt := int(maxItems)
			result.MaxItems = &maxItemsInt
		}
		if uniqueItems, ok := utils.GetBoolValue(rawSchema, "uniqueItems"); ok {
			result.UniqueItems = uniqueItems
		}

		// Handle array items
		if items, ok := utils.GetMapValue(rawSchema, "items"); ok {
			result.Items = convertToValidationSchema(items)
		}
	}

	// Set enum values if present
	if enum, ok := utils.GetArrayValue(rawSchema, "enum"); ok {
		result.Enum = enum
	}

	// Set required properties
	if required, ok := utils.GetArrayValue(rawSchema, "required"); ok {
		result.Required = make([]string, 0, len(required))
		for _, r := range required {
			if str, ok := r.(string); ok {
				result.Required = append(result.Required, str)
			}
		}
	}

	// Handle object properties
	if result.Type == "object" {
		if properties, ok := utils.GetMapValue(rawSchema, "properties"); ok {
			result.Properties = make(map[string]*validation.Schema)
			for propName, propSchema := range properties {
				if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
					result.Properties[propName] = convertToValidationSchema(propSchemaMap)
				}
			}
		}
	}

	return result
}

// GenerateKCLSchemasWithTree generates KCL schemas from a JSON Schema using the tree-based approach
func GenerateKCLSchemasWithTree(rawSchema map[string]interface{}, outputDir string, schemaName string) error {
	// Build the schema tree
	tree, err := BuildSchemaTree(rawSchema, schemaName, nil)
	if err != nil {
		return fmt.Errorf("failed to build schema tree: %w", err)
	}

	// Generate KCL schemas from the tree
	generator := NewTreeBasedGenerator(outputDir)
	_, err = generator.GenerateKCLSchemasFromTree(tree)
	if err != nil {
		return fmt.Errorf("failed to generate KCL schemas: %w", err)
	}

	return nil
}
