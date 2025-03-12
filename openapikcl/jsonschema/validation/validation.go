// Package validation provides validation-related functionality for JSON Schema to KCL conversion.
package validation

import (
	"fmt"
	"strings"

	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/types"
	"github.com/tridentsx/oas2kcl/openapikcl/jsonschema/utils"
)

// Schema represents a JSON Schema
type Schema struct {
	Type        string
	Properties  map[string]*Schema
	Items       *Schema
	Required    []string
	MinLength   *int
	MaxLength   *int
	Pattern     string
	Format      string
	Minimum     *float64
	Maximum     *float64
	MinItems    *int
	MaxItems    *int
	UniqueItems bool
	Enum        []interface{}
}

// SchemaImports tracks required imports for validation
type SchemaImports struct {
	NeedsDatetime bool
	NeedsNet      bool
	NeedsRegex    bool
	NeedsRuntime  bool
}

// formatDateTimeValidation generates consistent validation code for date-time format
// This ensures that date-time validation is the same regardless of whether it's a standalone property
// or part of an array
func formatDateTimeValidation(propNameOrItem string, propName string, isArrayItem bool, imports *SchemaImports) string {
	var content strings.Builder

	imports.NeedsRegex = true
	imports.NeedsDatetime = true

	if isArrayItem {
		// For array items, reference the DateTimeValidator schema
		content.WriteString(fmt.Sprintf("        # Validate all items in the array are valid RFC 3339 date-times\n"))
		content.WriteString(fmt.Sprintf("        all item in %s {\n", propName))
		content.WriteString(fmt.Sprintf("            # Use regex to check basic format with ranges\n"))
		content.WriteString(fmt.Sprintf("            regex.match(item, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\") and\n"))
		content.WriteString(fmt.Sprintf("            # Validate the date part\n"))
		content.WriteString(fmt.Sprintf("            item and datetime.validate(item[:10], \"%%Y-%%m-%%d\")\n"))
		content.WriteString(fmt.Sprintf("        } if %s, \"All items in %s must be valid RFC 3339 date-time\"\n", propName, propName))
	} else {
		// For standalone properties
		content.WriteString(fmt.Sprintf("        # RFC 3339 date-time validation for %s\n", propName))
		content.WriteString("        # Validates strings to ensure they conform to RFC 3339 date-time format.\n")
		content.WriteString(fmt.Sprintf("        # First check the format with a strict regex that enforces correct ranges\n"))
		content.WriteString(fmt.Sprintf("        regex.match(%s, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\") if %s != None, \"%s must be a valid RFC 3339 date-time\"\n", propNameOrItem, propNameOrItem, propName))
		content.WriteString(fmt.Sprintf("        # Then validate the date component to catch invalid dates like Feb 30\n"))
		content.WriteString(fmt.Sprintf("        %s != None and %s and datetime.validate(%s[:10], \"%%Y-%%m-%%d\"), \"%s contains an invalid date component\"\n", propNameOrItem, propNameOrItem, propNameOrItem, propName))
	}

	return content.String()
}

// generateSchemaBasedValidation creates validation logic for a specific format using dedicated schemas
// rather than embedding validation directly
func generateSchemaBasedValidation(format string, propName string) string {
	var content strings.Builder

	// Create header and import directives
	content.WriteString("\n# Generated schemas for format validation\n")

	switch format {
	case "date-time":
		content.WriteString("# DateTime validator schema\n")
		content.WriteString("schema DateTimeValidator:\n")
		content.WriteString("    value: str\n\n")
		content.WriteString("    check:\n")
		content.WriteString("        # First check format with regex that enforces correct ranges\n")
		content.WriteString("        regex.match(value, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\"), \"Value must be a valid RFC 3339 date-time\"\n")
		content.WriteString("        # Validate the date part to catch invalid dates like Feb 30\n")
		content.WriteString("        datetime.validate(value[:10], \"%Y-%m-%d\"), \"Value contains an invalid date component\"\n")
	case "email":
		content.WriteString("# Email validator schema\n")
		content.WriteString("schema EmailValidator:\n")
		content.WriteString("    value: str\n\n")
		content.WriteString("    check:\n")
		content.WriteString("        regex.match(value, r\"" + utils.CommonPatterns["email"] + "\"), \"Value must be a valid email address\"\n")
	case "uri":
		content.WriteString("# URI validator schema\n")
		content.WriteString("schema URIValidator:\n")
		content.WriteString("    value: str\n\n")
		content.WriteString("    check:\n")
		content.WriteString("        # URI format validation - checks for valid URI format with scheme\n")
		content.WriteString("        regex.match(value, r\"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]\"), \"Value must be a valid URI\"\n")
	case "uuid":
		content.WriteString("# UUID validator schema\n")
		content.WriteString("schema UUIDValidator:\n")
		content.WriteString("    value: str\n\n")
		content.WriteString("    check:\n")
		content.WriteString("        # UUID format validation - checks for valid UUID pattern\n")
		content.WriteString("        regex.match(value, r\"^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$\"), \"Value must be a valid UUID\"\n")
	case "ipv4":
		content.WriteString("# IPv4 validator schema\n")
		content.WriteString("schema IPv4Validator:\n")
		content.WriteString("    value: str\n\n")
		content.WriteString("    check:\n")
		content.WriteString("        # IPv4 format validation - checks for valid IPv4 address format\n")
		content.WriteString("        regex.match(value, r\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\"), \"Value must be a valid IPv4 address\"\n")
	}

	return content.String()
}

// generateTypeValidation generates validation code for a specific type and format
// This ensures consistent validation regardless of whether the type is used directly
// or as part of an array
func generateTypeValidation(typeName string, format string, propNameOrItem string, propName string, isArrayItem bool, imports *SchemaImports) string {
	var content strings.Builder

	switch format {
	case "email":
		imports.NeedsRegex = true
		if isArrayItem {
			content.WriteString(fmt.Sprintf("        all item in %s { regex.match(`%s`, item) } if %s, \"All items in %s must be valid email addresses\"\n",
				propName, utils.CommonPatterns["email"], propName, propName))
		} else {
			content.WriteString(fmt.Sprintf("        regex.match(%s, \"%s\") if %s, \"%s must be a valid email address\"\n",
				propNameOrItem, utils.CommonPatterns["email"], propNameOrItem, propName))
		}
	case "idn-email":
		imports.NeedsRegex = true
		if isArrayItem {
			content.WriteString(fmt.Sprintf("        all item in %s { regex.match(`%s`, item) } if %s, \"All items in %s must be valid internationalized email addresses\"\n",
				propName, utils.CommonPatterns["idn-email"], propName, propName))
		} else {
			content.WriteString(fmt.Sprintf("        regex.match(%s, \"%s\") if %s, \"%s must be a valid internationalized email address\"\n",
				propNameOrItem, utils.CommonPatterns["idn-email"], propNameOrItem, propName))
		}
	case "ipv4":
		imports.NeedsNet = true
		if isArrayItem {
			content.WriteString(fmt.Sprintf("        all item in %s { net.is_IPv4(item) } if %s, \"All items in %s must be valid IPv4 addresses\"\n",
				propName, propName, propName))
		} else {
			content.WriteString(fmt.Sprintf("        net.is_IPv4(%s) if %s, \"%s must be a valid IPv4 address\"\n",
				propNameOrItem, propNameOrItem, propName))
		}
	case "ipv6":
		imports.NeedsNet = true
		if isArrayItem {
			content.WriteString(fmt.Sprintf("        all item in %s { net.is_IP(item) } if %s, \"All items in %s must be valid IPv6 addresses\"\n",
				propName, propName, propName))
		} else {
			content.WriteString(fmt.Sprintf("        net.is_IP(%s) if %s, \"%s must be a valid IPv6 address\"\n",
				propNameOrItem, propNameOrItem, propName))
		}
	case "date-time":
		return formatDateTimeValidation(propNameOrItem, propName, isArrayItem, imports)
	case "date":
		imports.NeedsDatetime = true
		if isArrayItem {
			content.WriteString(fmt.Sprintf("        # Date validation for items in %s using datetime module\n", propName))
			content.WriteString("        # Validates strings to ensure they conform to YYYY-MM-DD format.\n")
			content.WriteString(fmt.Sprintf("        all item in %s { datetime.validate(item, \"%%Y-%%m-%%d\") } if %s, \"All items in %s must be valid dates in YYYY-MM-DD format\"\n",
				propName, propName, propName))
		} else {
			content.WriteString(fmt.Sprintf("        # Date validation for %s using datetime module\n", propName))
			content.WriteString("        # Validates strings to ensure they conform to YYYY-MM-DD format.\n")
			content.WriteString(fmt.Sprintf("        datetime.validate(%s, \"%%Y-%%m-%%d\") if %s != None, \"%s must be a valid date in YYYY-MM-DD format\"\n",
				propNameOrItem, propNameOrItem, propName))
		}
	case "time":
		imports.NeedsDatetime = true
		if isArrayItem {
			content.WriteString(fmt.Sprintf("        # Time validation for items in %s using datetime module\n", propName))
			content.WriteString("        # Validates strings to ensure they conform to HH:MM:SS format.\n")
			content.WriteString(fmt.Sprintf("        all item in %s { datetime.validate(item, \"%%H:%%M:%%S\") } if %s, \"All items in %s must be valid times in HH:MM:SS format\"\n",
				propName, propName, propName))
		} else {
			content.WriteString(fmt.Sprintf("        # Time validation for %s using datetime module\n", propName))
			content.WriteString("        # Validates strings to ensure they conform to HH:MM:SS format.\n")
			content.WriteString(fmt.Sprintf("        datetime.validate(%s, \"%%H:%%M:%%S\") if %s != None, \"%s must be a valid time in HH:MM:SS format\"\n",
				propNameOrItem, propNameOrItem, propName))
		}
		// Add other formats as needed...
	}

	return content.String()
}

// GenerateConstraints generates KCL constraints for a property
func GenerateConstraints(propSchema map[string]interface{}, propName string) string {
	var content strings.Builder
	var imports SchemaImports
	schemaType, _ := types.GetSchemaType(propSchema)

	content.WriteString("    check:\n")

	// String constraints
	if schemaType == "string" {
		// Min/max length
		if minLength, ok := utils.GetIntValue(propSchema, "minLength"); ok {
			content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have a minimum length of %d\"\n", propName, minLength, propName, propName, minLength))
		}
		if maxLength, ok := utils.GetIntValue(propSchema, "maxLength"); ok {
			content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have a maximum length of %d\"\n", propName, maxLength, propName, propName, maxLength))
		}

		// Pattern
		if pattern, ok := utils.GetStringValue(propSchema, "pattern"); ok {
			content.WriteString(fmt.Sprintf("        regex.match(%s, %s) if %s, \"%s must match pattern %s\"\n", propName, pattern, propName, propName, pattern))
		}

		// Format
		if format, ok := utils.GetStringValue(propSchema, "format"); ok {
			content.WriteString(generateTypeValidation(schemaType, format, propName, propName, false, &imports))
		}

		// Enum
		if enumValues, ok := utils.GetArrayValue(propSchema, "enum"); ok && len(enumValues) > 0 {
			enumStr := formatEnumList(enumValues)
			content.WriteString(fmt.Sprintf("        # Enum values: %s\n", formatEnumValues(enumValues)))
			content.WriteString(fmt.Sprintf("        %s == None or %s in %s, \"%s must be one of %s\"\n", propName, propName, enumStr, propName, formatEnumValues(enumValues)))
		}
	}

	// Number/integer constraints
	if schemaType == "number" || schemaType == "integer" {
		if minimum, ok := utils.GetFloatValue(propSchema, "minimum"); ok {
			content.WriteString(fmt.Sprintf("        %s == None or %s >= %v, \"%s must be greater than or equal to %v\"\n", propName, propName, minimum, propName, minimum))
		}
		if maximum, ok := utils.GetFloatValue(propSchema, "maximum"); ok {
			content.WriteString(fmt.Sprintf("        %s == None or %s <= %v, \"%s must be less than or equal to %v\"\n", propName, propName, maximum, propName, maximum))
		}
	}

	// Array constraints
	if schemaType == "array" {
		// Min/max items
		if minItems, ok := utils.GetIntValue(propSchema, "minItems"); ok {
			content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s != None, \"%s must have at least %d items\"\n", propName, minItems, propName, propName, minItems))
		}
		if maxItems, ok := utils.GetIntValue(propSchema, "maxItems"); ok {
			content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have at most %d items\"\n", propName, maxItems, propName, propName, maxItems))
		}

		// Unique items
		if uniqueItems, ok := utils.GetBoolValue(propSchema, "uniqueItems"); ok && uniqueItems {
			content.WriteString(fmt.Sprintf("        isunique(%s) if %s, \"%s must contain unique items\"\n", propName, propName, propName))
		}

		// Item constraints
		itemsSchema, ok := utils.GetMapValue(propSchema, "items")
		if ok {
			itemType, typeOk := types.GetSchemaType(itemsSchema)
			if typeOk && itemType == "string" {
				// Check for min/max length in items
				if minLength, ok := utils.GetIntValue(itemsSchema, "minLength"); ok {
					content.WriteString(fmt.Sprintf("        all item in %s { len(item) >= %d } if %s, \"All items in %s must have a minimum length of %d\"\n", propName, minLength, propName, propName, minLength))
				}
				if maxLength, ok := utils.GetIntValue(itemsSchema, "maxLength"); ok {
					content.WriteString(fmt.Sprintf("        all item in %s { len(item) <= %d } if %s, \"All items in %s must have a maximum length of %d\"\n", propName, maxLength, propName, propName, maxLength))
				}

				// Pattern validation for items
				if pattern, ok := utils.GetStringValue(itemsSchema, "pattern"); ok {
					content.WriteString(fmt.Sprintf("        all item in %s { regex.match(%s, item) } if %s, \"All items in %s must match pattern %s\"\n", propName, pattern, propName, propName, pattern))
				}

				// Format validation for items - use the same validation logic as standalone properties
				if format, ok := utils.GetStringValue(itemsSchema, "format"); ok {
					content.WriteString(generateTypeValidation(itemType, format, "item", propName, true, &imports))
				}

				// Enum for string items
				if enumValues, ok := utils.GetArrayValue(itemsSchema, "enum"); ok && len(enumValues) > 0 {
					enumStr := formatEnumList(enumValues)
					content.WriteString(fmt.Sprintf("        all item in %s { item in %s } if %s, \"All items in %s must be one of %s\"\n", propName, enumStr, propName, propName, formatEnumValues(enumValues)))
				}
			}

			// For numeric item types, add min/max constraints
			if typeOk && (itemType == "number" || itemType == "integer") {
				if minimum, ok := utils.GetFloatValue(itemsSchema, "minimum"); ok {
					content.WriteString(fmt.Sprintf("        all item in %s { item >= %v } if %s, \"All items in %s must be greater than or equal to %v\"\n", propName, minimum, propName, propName, minimum))
				}
				if maximum, ok := utils.GetFloatValue(itemsSchema, "maximum"); ok {
					content.WriteString(fmt.Sprintf("        all item in %s { item <= %v } if %s, \"All items in %s must be less than or equal to %v\"\n", propName, maximum, propName, propName, maximum))
				}
			}
		}
	}

	return content.String()
}

// GenerateValidatorSchema generates a KCL validator schema with constraints in a check block
func GenerateValidatorSchema(schema *Schema, schemaName string) (string, SchemaImports) {
	var imports SchemaImports
	var content strings.Builder
	var auxSchemas strings.Builder // For additional validator schemas

	// Add imports at file level first
	if schema.Type == "string" && schema.Format != "" {
		// Enable needed imports based on format
		switch schema.Format {
		case "date-time":
			imports.NeedsDatetime = true
			imports.NeedsRegex = true
		case "email", "uri", "uuid", "ipv4":
			imports.NeedsRegex = true
		}
	}

	// Need to check for array items with formats too
	if schema.Type == "array" && schema.Items != nil && schema.Items.Type == "string" && schema.Items.Format != "" {
		switch schema.Items.Format {
		case "date-time":
			imports.NeedsDatetime = true
			imports.NeedsRegex = true
		case "email", "uri", "uuid", "ipv4":
			imports.NeedsRegex = true
		}
	}

	// Include any other imports that might be needed based on properties
	for _, propSchema := range schema.Properties {
		if propSchema.Type == "string" && propSchema.Format != "" {
			switch propSchema.Format {
			case "date-time":
				imports.NeedsDatetime = true
				imports.NeedsRegex = true
			case "email", "uri", "uuid", "ipv4":
				imports.NeedsRegex = true
			}
		}

		// Check array items in properties
		if propSchema.Type == "array" && propSchema.Items != nil &&
			propSchema.Items.Type == "string" && propSchema.Items.Format != "" {
			switch propSchema.Items.Format {
			case "date-time":
				imports.NeedsDatetime = true
				imports.NeedsRegex = true
			case "email", "uri", "uuid", "ipv4":
				imports.NeedsRegex = true
			}
		}
	}

	// Avoid repeating imports by checking if we've already added them in the parent schema
	var existingImports string

	content.WriteString("\n\n")

	// Add file level imports
	if imports.NeedsRegex {
		content.WriteString("import regex\n")
		existingImports += "regex "
	}
	if imports.NeedsNet {
		content.WriteString("import net\n")
		existingImports += "net "
	}
	if imports.NeedsDatetime {
		content.WriteString("import datetime\n")
		existingImports += "datetime "
	}
	if imports.NeedsRuntime {
		content.WriteString("import runtime\n")
		existingImports += "runtime "
	}
	if imports.NeedsRegex || imports.NeedsNet || imports.NeedsDatetime || imports.NeedsRuntime {
		content.WriteString("\n")
	}

	// Generate specialized validator schemas for formats
	// First check if we need validator schemas for the top-level schema
	if schema.Type == "string" && schema.Format != "" {
		auxSchemas.WriteString(generateSchemaBasedValidation(schema.Format, schemaName))
		auxSchemas.WriteString("\n\n")
	}

	// Also check if we need validators for array items in the top-level schema
	if schema.Type == "array" && schema.Items != nil &&
		schema.Items.Type == "string" && schema.Items.Format != "" {
		auxSchemas.WriteString(generateSchemaBasedValidation(schema.Items.Format, schemaName+"Item"))
		auxSchemas.WriteString("\n\n")
	}

	// Check for properties in objects that need validator schemas
	if schema.Type == "object" && schema.Properties != nil {
		for propName, propSchema := range schema.Properties {
			// For string properties with format
			if propSchema.Type == "string" && propSchema.Format != "" {
				auxSchemas.WriteString(generateSchemaBasedValidation(propSchema.Format, propName))
				auxSchemas.WriteString("\n\n")
			}

			// For array properties with items that have format
			if propSchema.Type == "array" && propSchema.Items != nil &&
				propSchema.Items.Type == "string" && propSchema.Items.Format != "" {
				auxSchemas.WriteString(generateSchemaBasedValidation(propSchema.Items.Format, propName+"Item"))
				auxSchemas.WriteString("\n\n")
			}
		}
	}

	// Add auxiliary schemas to the content
	content.WriteString(auxSchemas.String())

	content.WriteString(fmt.Sprintf("# Validator schema for %s\n", schemaName))
	content.WriteString(fmt.Sprintf("schema %s:\n", schemaName))

	// Add properties based on schema type
	switch schema.Type {
	case "string":
		// If this is a format-based string that has a dedicated schema, reference it
		if schema.Format != "" {
			content.WriteString("    value: str\n\n")
		} else {
			content.WriteString("    value?: str\n\n")
		}
	case "integer":
		content.WriteString("    value?: int\n\n")
	case "number":
		content.WriteString("    value?: float\n\n")
	case "array":
		if schema.Items != nil {
			content.WriteString("    value?: [")
			switch schema.Items.Type {
			case "string":
				// Use validator schemas for specialized formats in arrays
				if schema.Items.Format != "" {
					switch schema.Items.Format {
					case "date-time":
						content.WriteString("DateTimeValidator")
					case "email":
						content.WriteString("EmailValidator")
					case "uri":
						content.WriteString("URIValidator")
					case "uuid":
						content.WriteString("UUIDValidator")
					case "ipv4":
						content.WriteString("IPv4Validator")
					default:
						content.WriteString("str")
					}
				} else {
					content.WriteString("str")
				}
			case "integer":
				content.WriteString("int")
			case "number":
				content.WriteString("float")
			case "boolean":
				content.WriteString("bool")
			case "object":
				// For object types, use a simple dict type as placeholder
				content.WriteString("dict")
			default:
				content.WriteString("any")
			}
			content.WriteString("] = []\n\n")
		}
	case "object":
		if schema.Properties != nil {
			// Add properties
			for propName, propSchema := range schema.Properties {
				isRequired := false
				for _, reqProp := range schema.Required {
					if reqProp == propName {
						isRequired = true
						break
					}
				}
				if isRequired {
					content.WriteString(fmt.Sprintf("    %s: ", propName))
				} else {
					content.WriteString(fmt.Sprintf("    %s?: ", propName))
				}
				switch propSchema.Type {
				case "string":
					// Handle special format properties by referencing the validator schema
					if propSchema.Format != "" {
						switch propSchema.Format {
						case "date-time":
							content.WriteString("DateTimeValidator")
						case "email":
							content.WriteString("EmailValidator")
						case "uri":
							content.WriteString("URIValidator")
						case "uuid":
							content.WriteString("UUIDValidator")
						case "ipv4":
							content.WriteString("IPv4Validator")
						default:
							content.WriteString("str")
						}
					} else {
						content.WriteString("str")
					}
				case "integer":
					content.WriteString("int")
				case "number":
					content.WriteString("float")
				case "boolean":
					content.WriteString("bool")
				case "array":
					content.WriteString("[")
					if propSchema.Items != nil {
						switch propSchema.Items.Type {
						case "string":
							// Use validator schemas for specialized formats in arrays
							if propSchema.Items.Format != "" {
								switch propSchema.Items.Format {
								case "date-time":
									content.WriteString("DateTimeValidator")
								case "email":
									content.WriteString("EmailValidator")
								case "uri":
									content.WriteString("URIValidator")
								case "uuid":
									content.WriteString("UUIDValidator")
								case "ipv4":
									content.WriteString("IPv4Validator")
								default:
									content.WriteString("str")
								}
							} else {
								content.WriteString("str")
							}
						case "integer":
							content.WriteString("int")
						case "number":
							content.WriteString("float")
						case "boolean":
							content.WriteString("bool")
						default:
							content.WriteString("any")
						}
					} else {
						content.WriteString("any")
					}
					content.WriteString("]")
					if propSchema.Type == "array" {
						content.WriteString(" = []")
					}
				case "object":
					content.WriteString("any")
				}
				content.WriteString("\n")
			}
			content.WriteString("\n    check:\n")
			// VALIDATED: Property validation - DO NOT CHANGE
			// Only validate properties that don't have specialized validator schemas
			for propName, propSchema := range schema.Properties {
				// Skip validation for properties using specialized validator schemas
				if propSchema.Type == "string" && propSchema.Format == "date-time" {
					continue
				}

				if propSchema.MinLength != nil {
					content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have a minimum length of %d\"\n", propName, *propSchema.MinLength, propName, propName, *propSchema.MinLength))
				}
				if propSchema.MaxLength != nil {
					content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have a maximum length of %d\"\n", propName, *propSchema.MaxLength, propName, propName, *propSchema.MaxLength))
				}
				if propSchema.Pattern != "" {
					imports.NeedsRegex = true
					goPattern := utils.TranslateECMAToGoRegex(propSchema.Pattern)
					content.WriteString(fmt.Sprintf("        regex.match(%s, %s) if %s, \"%s must match pattern %s\"\n", goPattern, propName, propName, propName, propSchema.Pattern))
				}
				if propSchema.Format != "" && propSchema.Format != "date-time" {
					content.WriteString(generateTypeValidation(propSchema.Type, propSchema.Format, propName, propName, false, &imports))
				}
				if propSchema.Type == "integer" || propSchema.Type == "number" {
					if propSchema.Minimum != nil {
						content.WriteString(fmt.Sprintf("        %s == None or %s >= %v, \"%s must be greater than or equal to %v\"\n", propName, propName, *propSchema.Minimum, propName, *propSchema.Minimum))
					}
					if propSchema.Maximum != nil {
						content.WriteString(fmt.Sprintf("        %s == None or %s <= %v, \"%s must be less than or equal to %v\"\n", propName, propName, *propSchema.Maximum, propName, *propSchema.Maximum))
					}
				}
				if propSchema.Type == "array" && propSchema.Items != nil {
					// Skip item validation for arrays of specialized validator schemas
					if propSchema.Items.Type == "string" && propSchema.Items.Format == "date-time" {
						// Only validate array-specific constraints like minItems, not item formats
						if propSchema.MinItems != nil {
							content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have at least %d items\"\n", propName, *propSchema.MinItems, propName, propName, *propSchema.MinItems))
						}
						if propSchema.MaxItems != nil {
							content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have at most %d items\"\n", propName, *propSchema.MaxItems, propName, propName, *propSchema.MaxItems))
						}
						if propSchema.UniqueItems {
							content.WriteString(fmt.Sprintf("        isunique(%s) if %s, \"%s must contain unique items\"\n", propName, propName, propName))
						}
						continue
					}

					if propSchema.MinItems != nil {
						content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have at least %d items\"\n", propName, *propSchema.MinItems, propName, propName, *propSchema.MinItems))
					}
					if propSchema.MaxItems != nil {
						content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have at most %d items\"\n", propName, *propSchema.MaxItems, propName, propName, *propSchema.MaxItems))
					}
					if propSchema.UniqueItems {
						content.WriteString(fmt.Sprintf("        isunique(%s) if %s, \"%s must contain unique items\"\n", propName, propName, propName))
					}
					if propSchema.Items.MinLength != nil {
						content.WriteString(fmt.Sprintf("        all item in %s { len(item) >= %d } if %s, \"All items in %s must have a minimum length of %d\"\n", propName, *propSchema.Items.MinLength, propName, propName, *propSchema.Items.MinLength))
					}
					if propSchema.Items.MaxLength != nil {
						content.WriteString(fmt.Sprintf("        all item in %s { len(item) <= %d } if %s, \"All items in %s must have a maximum length of %d\"\n", propName, *propSchema.Items.MaxLength, propName, propName, *propSchema.Items.MaxLength))
					}
					if propSchema.Items.Pattern != "" {
						imports.NeedsRegex = true
						goPattern := utils.TranslateECMAToGoRegex(propSchema.Items.Pattern)
						content.WriteString(fmt.Sprintf("        all item in %s { regex.match(%s, item) } if %s, \"All items in %s must match pattern %s\"\n", propName, goPattern, propName, propName, propSchema.Items.Pattern))
					}
					if propSchema.Items.Format != "" && propSchema.Items.Format != "date-time" {
						content.WriteString(generateTypeValidation(propSchema.Items.Type, propSchema.Items.Format, "item", propName, true, &imports))
					}
				}
				if schema.Enum != nil {
					enumValues := make([]string, len(schema.Enum))
					for i, v := range schema.Enum {
						enumValues[i] = fmt.Sprintf("%#v", v)
					}
					content.WriteString(fmt.Sprintf("        %s == None or %s in [%s], \"%s must be one of %s\"\n", propName, propName, strings.Join(enumValues, ", "), propName, strings.Join(enumValues, ", ")))
				}
			}
		}
	}

	// Always add a check block for validation logic
	content.WriteString("    check:\n")

	// Add validation logic based on schema type
	switch schema.Type {
	case "string":
		// If this is a format-based string that has a dedicated schema, add appropriate validation
		if schema.Format != "" {
			// Add imports needed by the validation
			switch schema.Format {
			case "date-time":
				imports.NeedsDatetime = true
				imports.NeedsRegex = true
				// Direct validation in the main schema
				content.WriteString("        # First check format with regex that enforces correct ranges\n")
				content.WriteString("        regex.match(value, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\") if value, \"Value must be a valid RFC 3339 date-time\"\n")
				content.WriteString("        # Validate the date part to catch invalid dates like Feb 30\n")
				content.WriteString("        datetime.validate(value[:10], \"%Y-%m-%d\") if value, \"Value contains an invalid date component\"\n")
			case "email":
				imports.NeedsRegex = true
				content.WriteString("        # Validate email format\n")
				content.WriteString("        regex.match(value, r\"" + utils.CommonPatterns["email"] + "\") if value, \"Value must be a valid email address\"\n")
			case "uri":
				imports.NeedsRegex = true
				content.WriteString("        # Validate URI format\n")
				content.WriteString("        regex.match(value, r\"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%%=~_|]\") if value, \"Value must be a valid URI\"\n")
			case "uuid":
				imports.NeedsRegex = true
				content.WriteString("        # Validate UUID format\n")
				content.WriteString("        regex.match(value, r\"^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$\") if value, \"Value must be a valid UUID\"\n")
			case "ipv4":
				imports.NeedsRegex = true
				content.WriteString("        # Validate IPv4 format\n")
				content.WriteString("        regex.match(value, r\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\") if value, \"Value must be a valid IPv4 address\"\n")
			}
		} else {
			// VALIDATED: String length validation - DO NOT CHANGE
			if schema.MinLength != nil {
				content.WriteString(fmt.Sprintf("        len(value) >= %d if value, \"value must have a minimum length of %d\"\n", *schema.MinLength, *schema.MinLength))
			}
			if schema.MaxLength != nil {
				content.WriteString(fmt.Sprintf("        len(value) <= %d if value, \"value must have a maximum length of %d\"\n", *schema.MaxLength, *schema.MaxLength))
			}
			// VALIDATED: Pattern validation - DO NOT CHANGE
			if schema.Pattern != "" {
				imports.NeedsRegex = true
				goPattern := utils.TranslateECMAToGoRegex(schema.Pattern)
				content.WriteString(fmt.Sprintf("        regex.match(value, %s) if value, \"value must match pattern %s\"\n", goPattern, schema.Pattern))
			}
			if schema.Format != "" && schema.Format != "date-time" {
				content.WriteString(generateTypeValidation("string", schema.Format, "value", "value", false, &imports))
			}
		}
	case "integer":
		// VALIDATED: Integer range validation - DO NOT CHANGE
		if schema.Minimum != nil {
			content.WriteString(fmt.Sprintf("        value == None or value >= %d, \"value must be greater than or equal to %d\"\n", int(*schema.Minimum), int(*schema.Minimum)))
		}
		if schema.Maximum != nil {
			content.WriteString(fmt.Sprintf("        value == None or value <= %d, \"value must be less than or equal to %d\"\n", int(*schema.Maximum), int(*schema.Maximum)))
		}
	case "number":
		// VALIDATED: Float range validation - DO NOT CHANGE
		if schema.Minimum != nil {
			content.WriteString(fmt.Sprintf("        value == None or value >= %f, \"value must be greater than or equal to %f\"\n", *schema.Minimum, *schema.Minimum))
		}
		if schema.Maximum != nil {
			content.WriteString(fmt.Sprintf("        value == None or value <= %f, \"value must be less than or equal to %f\"\n", *schema.Maximum, *schema.Maximum))
		}
	case "array":
		if schema.Items != nil {
			// VALIDATED: Array length validation - DO NOT CHANGE
			if schema.MinItems != nil {
				content.WriteString(fmt.Sprintf("        len(value) >= %d if value, \"value must have at least %d items\"\n", *schema.MinItems, *schema.MinItems))
			}
			if schema.MaxItems != nil {
				content.WriteString(fmt.Sprintf("        len(value) <= %d if value, \"value must have at most %d items\"\n", *schema.MaxItems, *schema.MaxItems))
			}
			if schema.UniqueItems {
				content.WriteString("        isunique(value) if value, \"value must contain unique items\"\n")
			}

			// For specialized types, add format-specific validation
			if schema.Items.Type == "string" && schema.Items.Format != "" {
				// Add format-specific validation for array items
				switch schema.Items.Format {
				case "date-time":
					imports.NeedsDatetime = true
					imports.NeedsRegex = true
					content.WriteString("        # Validate date-time format for all array items\n")
					content.WriteString("        all item in value { regex.match(item, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\") } if value, \"All items must be valid RFC 3339 date-time values\"\n")
					content.WriteString("        # Validate the date part to catch invalid dates like Feb 30\n")
					content.WriteString("        all item in value { datetime.validate(item[:10], \"%Y-%m-%d\") } if value, \"All items must have valid date components\"\n")
				case "email":
					imports.NeedsRegex = true
					content.WriteString("        # Validate email format for all array items\n")
					content.WriteString("        all item in value { regex.match(item, r\"" + utils.CommonPatterns["email"] + "\") } if value, \"All items must be valid email addresses\"\n")
				case "uri":
					imports.NeedsRegex = true
					content.WriteString("        # Validate URI format for all array items\n")
					content.WriteString("        all item in value { regex.match(item, r\"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%%=~_|]\") } if value, \"All items must be valid URIs\"\n")
				case "uuid":
					imports.NeedsRegex = true
					content.WriteString("        # Validate UUID format for all array items\n")
					content.WriteString("        all item in value { regex.match(item, r\"^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$\") } if value, \"All items must be valid UUIDs\"\n")
				case "ipv4":
					imports.NeedsRegex = true
					content.WriteString("        # Validate IPv4 format for all array items\n")
					content.WriteString("        all item in value { regex.match(item, r\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\") } if value, \"All items must be valid IPv4 addresses\"\n")
				}
			} else {
				// Regular validation for non-specialized formats
				if schema.Items.MinLength != nil {
					content.WriteString(fmt.Sprintf("        all item in value { len(item) >= %d } if value, \"All items must have a minimum length of %d\"\n", *schema.Items.MinLength, *schema.Items.MinLength))
				}
				if schema.Items.MaxLength != nil {
					content.WriteString(fmt.Sprintf("        all item in value { len(item) <= %d } if value, \"All items must have a maximum length of %d\"\n", *schema.Items.MaxLength, *schema.Items.MaxLength))
				}
				if schema.Items.Pattern != "" {
					imports.NeedsRegex = true
					goPattern := utils.TranslateECMAToGoRegex(schema.Items.Pattern)
					content.WriteString(fmt.Sprintf("        all item in value { regex.match(%s, item) } if value, \"All items must match pattern %s\"\n", goPattern, schema.Items.Pattern))
				}
				if schema.Items.Format != "" {
					content.WriteString(generateTypeValidation(schema.Items.Type, schema.Items.Format, "item", "value", true, &imports))
				}
			}
		}
	case "object":
		// VALIDATED: Property validation - DO NOT CHANGE
		// Only validate properties that don't have specialized validator schemas
		for propName, propSchema := range schema.Properties {
			// Handle special format properties
			if propSchema.Type == "string" && propSchema.Format != "" {
				switch propSchema.Format {
				case "date-time":
					imports.NeedsDatetime = true
					imports.NeedsRegex = true
					content.WriteString(fmt.Sprintf("        # Validate date-time format for %s\n", propName))
					content.WriteString(fmt.Sprintf("        regex.match(%s, r\"^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(\\.\\d+)?(Z|[+-]([01]\\d|2[0-3]):[0-5]\\d)$\") if %s, \"%s must be a valid RFC 3339 date-time\"\n", propName, propName, propName))
					content.WriteString(fmt.Sprintf("        # Validate the date part to catch invalid dates like Feb 30\n"))
					content.WriteString(fmt.Sprintf("        datetime.validate(%s[:10], \"%%Y-%%m-%%d\") if %s, \"%s contains an invalid date component\"\n", propName, propName, propName))
				case "email":
					imports.NeedsRegex = true
					content.WriteString(fmt.Sprintf("        # Validate email format for %s\n", propName))
					content.WriteString(fmt.Sprintf("        regex.match(%s, r\"%s\") if %s, \"%s must be a valid email address\"\n", propName, utils.CommonPatterns["email"], propName, propName))
				case "uri":
					imports.NeedsRegex = true
					content.WriteString(fmt.Sprintf("        # Validate URI format for %s\n", propName))
					content.WriteString(fmt.Sprintf("        regex.match(%s, r\"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%%=~_|]\") if %s, \"%s must be a valid URI\"\n", propName, propName, propName))
				case "uuid":
					imports.NeedsRegex = true
					content.WriteString(fmt.Sprintf("        # Validate UUID format for %s\n", propName))
					content.WriteString(fmt.Sprintf("        regex.match(%s, r\"^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$\") if %s, \"%s must be a valid UUID\"\n", propName, propName, propName))
				case "ipv4":
					imports.NeedsRegex = true
					content.WriteString(fmt.Sprintf("        # Validate IPv4 format for %s\n", propName))
					content.WriteString(fmt.Sprintf("        regex.match(%s, r\"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$\") if %s, \"%s must be a valid IPv4 address\"\n", propName, propName, propName))
				default:
					// For other formats, continue with regular validation
					if propSchema.MinLength != nil {
						content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have a minimum length of %d\"\n", propName, *propSchema.MinLength, propName, propName, *propSchema.MinLength))
					}
					if propSchema.MaxLength != nil {
						content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have a maximum length of %d\"\n", propName, *propSchema.MaxLength, propName, propName, *propSchema.MaxLength))
					}
					if propSchema.Pattern != "" {
						imports.NeedsRegex = true
						goPattern := utils.TranslateECMAToGoRegex(propSchema.Pattern)
						content.WriteString(fmt.Sprintf("        regex.match(%s, %s) if %s, \"%s must match pattern %s\"\n", propName, goPattern, propName, propName, propSchema.Pattern))
					}
				}
				continue
			}

			// Regular validation for non-specialized formats
			if propSchema.MinLength != nil {
				content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have a minimum length of %d\"\n", propName, *propSchema.MinLength, propName, propName, *propSchema.MinLength))
			}
			if propSchema.MaxLength != nil {
				content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have a maximum length of %d\"\n", propName, *propSchema.MaxLength, propName, propName, *propSchema.MaxLength))
			}
			if propSchema.Pattern != "" {
				imports.NeedsRegex = true
				goPattern := utils.TranslateECMAToGoRegex(propSchema.Pattern)
				content.WriteString(fmt.Sprintf("        regex.match(%s, %s) if %s, \"%s must match pattern %s\"\n", propName, goPattern, propName, propName, propSchema.Pattern))
			}
			if propSchema.Format != "" && propSchema.Format != "date-time" {
				content.WriteString(generateTypeValidation(propSchema.Type, propSchema.Format, propName, propName, false, &imports))
			}
			if propSchema.Type == "integer" || propSchema.Type == "number" {
				if propSchema.Minimum != nil {
					content.WriteString(fmt.Sprintf("        %s == None or %s >= %v, \"%s must be greater than or equal to %v\"\n", propName, propName, *propSchema.Minimum, propName, *propSchema.Minimum))
				}
				if propSchema.Maximum != nil {
					content.WriteString(fmt.Sprintf("        %s == None or %s <= %v, \"%s must be less than or equal to %v\"\n", propName, propName, *propSchema.Maximum, propName, *propSchema.Maximum))
				}
			}
			if propSchema.Type == "array" && propSchema.Items != nil {
				// Skip item validation for arrays of specialized validator schemas
				if propSchema.Items.Type == "string" && propSchema.Items.Format == "date-time" {
					// Only validate array-specific constraints like minItems, not item formats
					if propSchema.MinItems != nil {
						content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have at least %d items\"\n", propName, *propSchema.MinItems, propName, propName, *propSchema.MinItems))
					}
					if propSchema.MaxItems != nil {
						content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have at most %d items\"\n", propName, *propSchema.MaxItems, propName, propName, *propSchema.MaxItems))
					}
					if propSchema.UniqueItems {
						content.WriteString(fmt.Sprintf("        isunique(%s) if %s, \"%s must contain unique items\"\n", propName, propName, propName))
					}
					continue
				}

				if propSchema.MinItems != nil {
					content.WriteString(fmt.Sprintf("        len(%s) >= %d if %s, \"%s must have at least %d items\"\n", propName, *propSchema.MinItems, propName, propName, *propSchema.MinItems))
				}
				if propSchema.MaxItems != nil {
					content.WriteString(fmt.Sprintf("        len(%s) <= %d if %s, \"%s must have at most %d items\"\n", propName, *propSchema.MaxItems, propName, propName, *propSchema.MaxItems))
				}
				if propSchema.UniqueItems {
					content.WriteString(fmt.Sprintf("        isunique(%s) if %s, \"%s must contain unique items\"\n", propName, propName, propName))
				}
				if propSchema.Items.MinLength != nil {
					content.WriteString(fmt.Sprintf("        all item in %s { len(item) >= %d } if %s, \"All items in %s must have a minimum length of %d\"\n", propName, *propSchema.Items.MinLength, propName, propName, *propSchema.Items.MinLength))
				}
				if propSchema.Items.MaxLength != nil {
					content.WriteString(fmt.Sprintf("        all item in %s { len(item) <= %d } if %s, \"All items in %s must have a maximum length of %d\"\n", propName, *propSchema.Items.MaxLength, propName, propName, *propSchema.Items.MaxLength))
				}
				if propSchema.Items.Pattern != "" {
					imports.NeedsRegex = true
					goPattern := utils.TranslateECMAToGoRegex(propSchema.Items.Pattern)
					content.WriteString(fmt.Sprintf("        all item in %s { regex.match(%s, item) } if %s, \"All items in %s must match pattern %s\"\n", propName, goPattern, propName, propName, propSchema.Items.Pattern))
				}
				if propSchema.Items.Format != "" && propSchema.Items.Format != "date-time" {
					content.WriteString(generateTypeValidation(propSchema.Items.Type, propSchema.Items.Format, "item", propName, true, &imports))
				}
			}
			if schema.Enum != nil {
				enumValues := make([]string, len(schema.Enum))
				for i, v := range schema.Enum {
					enumValues[i] = fmt.Sprintf("%#v", v)
				}
				content.WriteString(fmt.Sprintf("        %s == None or %s in [%s], \"%s must be one of %s\"\n", propName, propName, strings.Join(enumValues, ", "), propName, strings.Join(enumValues, ", ")))
			}
		}
	}

	return content.String(), imports
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
		_, hasMultipleOf := utils.GetFloatValue(propSchema, "multipleOf")
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")

		return hasMin || hasMax || hasExclusiveMin || hasMultipleOf || hasEnum

	case "array":
		_, hasMinItems := utils.GetIntValue(propSchema, "minItems")
		_, hasMaxItems := utils.GetIntValue(propSchema, "maxItems")
		_, hasUniqueItems := utils.GetBoolValue(propSchema, "uniqueItems")

		return hasMinItems || hasMaxItems || hasUniqueItems

	case "boolean":
		_, hasEnum := utils.GetArrayValue(propSchema, "enum")
		return hasEnum

	default:
		return false
	}
}

// formatEnumValues formats enum values for use in a KCL constraint
func formatEnumValues(values []interface{}) string {
	parts := []string{}
	for _, val := range values {
		parts = append(parts, utils.FormatLiteral(val))
	}
	return strings.Join(parts, ", ")
}

// formatEnumList formats enum values for use in a KCL constraint
func formatEnumList(values []interface{}) string {
	parts := []string{}
	for _, val := range values {
		parts = append(parts, utils.FormatLiteral(val))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// CheckIfNeedsRegexImport recursively inspects a schema to determine if regex validation is needed
func CheckIfNeedsRegexImport(rawSchema map[string]interface{}) bool {
	// Check for pattern in the schema itself
	if pattern, ok := utils.GetStringValue(rawSchema, "pattern"); ok && pattern != "" {
		return true
	}

	// Check for format in the schema itself (many formats use regex validation)
	if format, ok := utils.GetStringValue(rawSchema, "format"); ok && format != "" {
		return true
	}

	// Check in properties
	if properties, ok := utils.GetMapValue(rawSchema, "properties"); ok {
		for _, propValue := range properties {
			if propSchema, ok := propValue.(map[string]interface{}); ok {
				if CheckIfNeedsRegexImport(propSchema) {
					return true
				}
			}
		}
	}

	// Check in array items
	if items, ok := utils.GetMapValue(rawSchema, "items"); ok {
		if CheckIfNeedsRegexImport(items) {
			return true
		}
	}

	// Check in definitions
	if definitions, ok := utils.GetMapValue(rawSchema, "definitions"); ok {
		for _, defValue := range definitions {
			if defSchema, ok := defValue.(map[string]interface{}); ok {
				if CheckIfNeedsRegexImport(defSchema) {
					return true
				}
			}
		}
	}

	return false
}

// GenerateRequiredPropertyChecks generates KCL check blocks for required properties
func GenerateRequiredPropertyChecks(schema map[string]interface{}) string {
	required, ok := utils.GetArrayValue(schema, "required")
	if !ok || len(required) == 0 {
		return ""
	}

	checks := []string{}
	for _, req := range required {
		propName, ok := req.(string)
		if !ok {
			continue
		}
		sanitizedName := utils.SanitizePropertyName(propName)
		checks = append(checks, "    if "+sanitizedName+" == None:")
		checks = append(checks, "        assert False, \""+propName+" is required\"")
	}

	if len(checks) > 0 {
		return "\n" + strings.Join(checks, "\n")
	}

	return ""
}
