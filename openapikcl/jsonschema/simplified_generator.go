package jsonschema

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// TreeBasedGenerator generates KCL schemas from a schema tree
type SimplifiedTreeBasedGenerator struct {
	OutputDir              string
	GeneratedFiles         map[string]bool
	SchemaRegistry         map[string]string
	processedNodes         map[string]bool
	ExtendedPrimitiveNodes map[string]bool //Or extended primitive nodes
}

// NewTreeBasedGenerator creates a new TreeBasedGenerator
func SimplifiedNewTreeBasedGenerator(outputDir string) *SimplifiedTreeBasedGenerator {
	return &SimplifiedTreeBasedGenerator{
		OutputDir:              outputDir,
		GeneratedFiles:         make(map[string]bool),
		SchemaRegistry:         make(map[string]string),
		processedNodes:         make(map[string]bool),
		ExtendedPrimitiveNodes: make(map[string]bool),
	}
}

type SchemaContent struct {
	Fields strings.Builder
	Check  strings.Builder
}

func NewSchemaContent() *SchemaContent {
	return &SchemaContent{
		Fields: strings.Builder{},
		Check:  strings.Builder{},
	}
}

type Kclschema struct {
	Imports   strings.Builder
	KCLSchema map[string]*SchemaContent
}

func NewSchema() *Kclschema {
	return &Kclschema{
		Imports:   strings.Builder{},
		KCLSchema: make(map[string]*SchemaContent),
	}
}

// GenerateSchemaTreeAndKCL parses a JSON schema, builds a schema tree, and generates KCL schemas
func SimplifiedGenerateSchemaTreeAndKCL(schemaBytes []byte, outputDir string, debugMode bool) error {
	// Parse the JSON schema
	var rawSchema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &rawSchema); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Determine the schema name from title or default to "Schema"
	// Determine the schema name from $id, title, or default to "Schema"
	schemaName := "Schema"
	if title, ok := rawSchema["title"].(string); ok && title != "" {
		schemaName = strings.ReplaceAll(strings.TrimSpace(title), " ", "_")
	} else if id, ok := rawSchema["$id"].(string); ok && id != "" {
		// Extract the last part of the path
		parts := strings.Split(id, "/")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			// Remove everything after first period (including the period)
			if dotIndex := strings.Index(lastPart, "."); dotIndex != -1 {
				lastPart = lastPart[:dotIndex]
			}
			// Capitalize first letter
			if len(lastPart) > 0 {
				schemaName = strings.ToUpper(lastPart[:1]) + lastPart[1:]
			}
		}
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
	generator := SimplifiedNewTreeBasedGenerator(outputDir)
	err = generator.SimplifiedGenerateKCLSchemasFromTree(tree)
	if err != nil {
		return fmt.Errorf("failed to generate KCL schemas: %w", err)
	}

	return nil
}

// GenerateKCLSchemasFromTree generates KCL schemas from a JSON schema tree
func (g *SimplifiedTreeBasedGenerator) SimplifiedGenerateKCLSchemasFromTree(tree *SchemaTreeNode) error {
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// First pass: identify all schemas that are primitive or extended primitive
	g.identifyNodeType(tree)

	//Create new instance of Kclschema
	kclschema := NewSchema()
	imports := make([]string, 0)
	g.determineImports(tree, &imports)
	if len(imports) > 0 {
		kclschema.Imports.WriteString(strings.Join(imports, "\n\n"))
		kclschema.Imports.WriteString("\n\n")
	}
	err := g.simplifiedGenerateSchemaFromNode(tree, kclschema)
	if err != nil {
		return err
	}
	// Output the content of kclschema in a single output file
	fileName := fmt.Sprintf("%s/schema.k", g.OutputDir)
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileName, err)
	}
	defer file.Close()

	// Write the imports once at the beginning
	file.WriteString(kclschema.Imports.String())

	// Create schemas for each schema content
	for _, schemaContent := range kclschema.KCLSchema {
		// Write the schema content
		file.WriteString(schemaContent.Fields.String())
		// file.WriteString("\n")

		// Write the check function
		file.WriteString(schemaContent.Check.String())
		file.WriteString("\n\n")
	}
	return nil

}

// generateSchemaFromNode generates a KCL schema from a tree node
func (g *SimplifiedTreeBasedGenerator) simplifiedGenerateSchemaFromNode(node *SchemaTreeNode, kclSchema *Kclschema) error {
	if node == nil {
		return nil
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
	//If it is primitive or extended primitive or an array or an object with no properties, its already added to its parent schema
	if g.ExtendedPrimitiveNodes[node.SchemaName] {
		return nil
	}

	// //Check if it is an array with primitive types
	// if node.Type == Array {
	// 	//Add a check if schema name key exists in map of kclschema. If not, initialize it
	// 	if _, ok := kclSchema.KCLSchema[node.SchemaName]; !ok {
	// 		optionalMarker := "?"
	// 		if required[node.SchemaName] {
	// 			optionalMarker = ""
	// 		}
	// 		//Create a new schema content if it doesn't exist
	// 		kclSchema.KCLSchema[node.SchemaName] = NewSchemaContent()
	// 		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s%s: [", optionalMarker, node.SchemaName))
	// 	}
	// 	//Add the property to the schema content
	// 	kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("    %s]\n", node.Items.SchemaName))
	// 	//Generate the schema for the property
	// 	g.simplifiedGenerateSchemaFromNode(node.Items, kclSchema)

	// 	return nil
	// }

	if g.isCompositeType(*node) {
		g.handleCompositeType(node, kclSchema, required)
	}

	//Add a check if schema name key exists in map of kclschema. If not, initialize it
	if _, ok := kclSchema.KCLSchema[node.SchemaName]; !ok {
		//Create a new schema content if it doesn't exist
		kclSchema.KCLSchema[node.SchemaName] = NewSchemaContent()
		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s:\n", node.SchemaName))
	}

	//Run a for loop on all properties on the object
	for propName, propNode := range node.Properties {
		optionalMarker := "?"
		if required[propName] {
			optionalMarker = ""
		}

		//Check if property is extended primitive
		if g.ExtendedPrimitiveNodes[propNode.SchemaName] {
			//Add the property to the schema content
			err := g.generatePrimitiveSchema(node, propName, propNode, kclSchema, required)
			if err != nil {
				return err
			}
		} else if propNode.Type == "object" {
			//Add the property to the schema content
			kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s: %s\n", propName, optionalMarker, propNode.SchemaName))
			//Generate the schema for the property
			g.simplifiedGenerateSchemaFromNode(propNode, kclSchema)
		} else if propNode.Type == "array" {
			kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s: [%s]\n", propName, optionalMarker, propNode.Items.SchemaName))
			if len(propNode.Constraints) != 0 {
				if kclSchema.KCLSchema[node.SchemaName].Check.Len() == 0 {
					kclSchema.KCLSchema[node.SchemaName].Check.WriteString("\n    check:\n")
				}
				for constraint, value := range propNode.Constraints {
					switch constraint {
					case "minItems":
						if minItems, ok := value.(float64); ok {
							kclSchema.KCLSchema[node.SchemaName].Check.WriteString(
								fmt.Sprintf("        len(%s) >= %d if %s, \"Array must have at least %d items\"\n",
									propName, int(minItems), propName, int(minItems)))
						}
					case "maxItems":
						if maxItems, ok := value.(float64); ok {
							kclSchema.KCLSchema[propNode.SchemaName].Check.WriteString(
								fmt.Sprintf("        len(%s) <= %d if %s, \"Array must have at most %d items\"\n",
									propName, int(maxItems), propName, int(maxItems)))
						}
					case "uniqueItems":
						if _, ok := value.(bool); ok {
							kclSchema.KCLSchema[propNode.SchemaName].Check.WriteString(
								fmt.Sprintf("        isunique(%s) if %s, \"Array items must be unique\"\n",
									propName, propName))
						}
					}
				}

			}
			g.simplifiedGenerateSchemaFromNode(propNode.Items, kclSchema)
		} else if g.isCompositeType(*propNode) {
			g.handleCompositeTypeProperies(propNode, propName, node, kclSchema, required)
		}
	}

	if node.AdditionalProperties != nil {

		if g.isExtendedPrimitiveType(*node.AdditionalProperties) {
			switch node.AdditionalProperties.Type {
			case "string":
				kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    [...str]: str\n")
			case "number":
				kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    [...str]: float\n")
			case "integer":
				kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    [...str]: int\n")
			case "boolean":
				kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    [...str]: bool\n")
			}
		} else if node.AdditionalProperties.Type == "array" {
			kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    [...str]: [" + node.AdditionalProperties.Items.SchemaName + "]\n")
			g.simplifiedGenerateSchemaFromNode(node.AdditionalProperties.Items, kclSchema)
		} else if node.AdditionalProperties.Type == "object" {
			kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    [...str]: " + node.AdditionalProperties.SchemaName + "\n")
			g.simplifiedGenerateSchemaFromNode(node.AdditionalProperties, kclSchema)
		}

	}
	return nil

}

// Helper Functions

func (g *SimplifiedTreeBasedGenerator) isCompositeType(node SchemaTreeNode) bool {
	nodeType := node.Type
	return nodeType == "anyOf" || nodeType == "allOf" || nodeType == "oneOf" || nodeType == "not"
}

// Add a fuction to check if nodetype is primitive/extended-primitive. So it doesn't need its own NeedsStandaloneSchemaObject
func (g *SimplifiedTreeBasedGenerator) isExtendedPrimitiveType(node SchemaTreeNode) bool {
	if g.isPrimitiveType(node) {
		return true
	}
	if node.Type == "array" {
		return g.isPrimitiveType(*node.Items)
	}
	if node.Type == "object" {
		return len(node.Properties) == 0 && node.AdditionalProperties == nil
	}
	return false
}

func (g *SimplifiedTreeBasedGenerator) isPrimitiveType(node SchemaTreeNode) bool {
	nodeType := node.Type
	return nodeType == "string" || nodeType == "number" || nodeType == "integer" || nodeType == "boolean" ||
		nodeType == "email" || nodeType == "date" || nodeType == "date-time" || nodeType == "uri" ||
		nodeType == "uuid" || nodeType == "hostname" || nodeType == "ipv4" || nodeType == "ipv6"
}

// identifyNodeType traverses the tree to identify which schemas need their own schemas. Lets call it NeedsStandaloneSchemaObject
func (g *SimplifiedTreeBasedGenerator) identifyNodeType(node *SchemaTreeNode) {
	if node == nil {
		return
	}

	// Check if this node is a primitive
	g.ExtendedPrimitiveNodes[node.SchemaName] = g.isExtendedPrimitiveType(*node)

	// Process child nodes
	switch node.Type {
	case Object:
		for _, propNode := range node.Properties {
			g.identifyNodeType(propNode)
		}
		// Process additional properties if present
		if node.AdditionalProperties != nil {
			g.identifyNodeType(node.AdditionalProperties)
		}
	case Array:
		g.identifyNodeType(node.Items)
	case AllOf, AnyOf, OneOf, Not, If, Then, Else:
		for _, subSchema := range node.SubSchemas {
			g.identifyNodeType(subSchema)
		}
	}
}

func (g *SimplifiedTreeBasedGenerator) addMixinSuffix(name string) string {
	if !strings.HasSuffix(name, "Mixin") {
		return name + "Mixin"
	}
	return name
}

func (g *SimplifiedTreeBasedGenerator) generatePrimitiveSchema(mainnode *SchemaTreeNode, nodeName string, node *SchemaTreeNode, kclSchema *Kclschema, required map[string]bool) error {
	// nodeName = formatSchemaName(nodeName)
	// Handle different node types
	optionalMarker := "?"
	if required[nodeName] {
		optionalMarker = ""
	}

	switch node.Type {

	case String:
		g.handleStringType(node, mainnode, nodeName, optionalMarker, kclSchema)

	case Number, Integer:
		var kclType string
		if node.Type == Integer {
			kclType = "int"
		} else {
			kclType = "float"
		}
		g.handleNumericType(node, mainnode, nodeName, optionalMarker, kclType, kclSchema)

	case Boolean:
		kclSchema.KCLSchema[mainnode.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s: %s\n", nodeName, optionalMarker, "bool"))

	// case AllOf, AnyOf, OneOf:
	// 	// Generate schemas for all subschemas first
	// 	for _, subSchema := range node.SubSchemas {
	// 		subFiles, err := g.generateSchemaFromNode(subSchema)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		createdFiles = append(createdFiles, subFiles...)
	// 	}

	// 	// Then generate the schema for this composition
	// 	schemaFile, err := g.generateCompositionSchema(node)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	createdFiles = append(createdFiles, schemaFile)
	// 	g.GeneratedFiles[nodeKey] = true

	// case Reference:
	// 	// Handling references - this might involve looking up the target
	// 	// and generating a schema for it
	// 	schemaFile, err := g.generateReferenceSchema(node)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	createdFiles = append(createdFiles, schemaFile)
	// 	g.GeneratedFiles[nodeKey] = true
	// }

	// return createdFiles, nil

	//This is case of array with primitive types
	case Array:
		kclSchema.KCLSchema[mainnode.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s: [%s]\n", nodeName, optionalMarker, convertKCLType(node.Items.Type)))
		if len(node.Constraints) != 0 {
			if kclSchema.KCLSchema[mainnode.SchemaName].Check.Len() == 0 {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString("\n    check:\n")
			}
			for constraint, value := range node.Constraints {
				switch constraint {
				case "minItems":
					if minItems, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        len(%s) >= %d if %s is not Undefined, \"Array must have at least %d items\"\n",
								nodeName, int(minItems), nodeName, int(minItems)))
					}
				case "maxItems":
					if maxItems, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        len(%s) <= %d if %s is not Undefined, \"Array must have at most %d items\"\n",
								nodeName, int(maxItems), nodeName, int(maxItems)))
					}
				case "uniqueItems":
					if _, ok := value.(bool); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        isunique(%s) if %s is not Undefined, \"Array items must be unique\"\n",
								nodeName, nodeName))
					}
				}
			}

		}
		switch node.Items.Type {
		case String:
			// all item in simple_array {len(item) > 3}
			// Skip if no constraints or format
			if len(node.Items.Constraints) == 0 && node.Items.Format == "" {
				return nil
			}

			// Initialize check block if needed
			if kclSchema.KCLSchema[mainnode.SchemaName].Check.Len() == 0 {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString("\n    check:\n")
			}
			// all item in typed_array {item > 0}
			for constraint, value := range node.Items.Constraints {
				switch constraint {
				case "minLength":
					if minLen, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item >= %d} if %s, \"String must be at least %d characters\"\n",
								nodeName, int(minLen), nodeName, int(minLen)))
					}
				case "maxLength":
					if maxLen, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item <= %d} if %s, \"String must be at most %d characters\"\n",
								nodeName, int(maxLen), nodeName, int(maxLen)))
					}
				case "pattern":
					if pattern, ok := value.(string); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {regex.match(item, r\"%s\")} if %s, \"String must match pattern %s\"\n",
								nodeName, pattern, nodeName, pattern))
					}
				case "enum":
					if enum, ok := value.([]interface{}); ok && len(enum) > 0 {
						enumValues := make([]string, len(enum))
						for i, e := range enum {
							enumValues[i] = fmt.Sprintf("%q", e)
						}
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item in [%s]} if %s \n",
								nodeName, strings.Join(enumValues, ", "), nodeName))
					}
				}
			}
		case Number, Integer:
			// Skip if no constraints
			if len(node.Items.Constraints) == 0 {
				return nil
			}

			// Initialize check block if needed
			if kclSchema.KCLSchema[mainnode.SchemaName].Check.Len() == 0 {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString("\n    check:\n")
			}

			for constraint, value := range node.Items.Constraints {
				switch constraint {
				case "minimum":
					if min, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item >= %g} if %s is not Undefined, \"%s must be at least %g\"\n",
								nodeName, min, nodeName, nodeName, min))
					}
				case "maximum":
					if max, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item <= %g} if %s is not Undefined, \"%s must be at most %g\"\n",
								nodeName, max, nodeName, nodeName, max))
					}
				case "exclusiveMinimum":
					if min, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item > %g} if %s is not Undefined, \"%s must be greater than %g\"\n",
								nodeName, min, nodeName, nodeName, min))
					}
				case "exclusiveMaximum":
					if max, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item < %g} if %s is not Undefined, \"%s must be less than %g\"\n",
								nodeName, max, nodeName, nodeName, max))
					}
				case "multipleOf":
					if multiple, ok := value.(float64); ok {
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item %% %g} == 0 if %s is not Undefined, \"%s must be a multiple of %g\"\n",
								nodeName, multiple, nodeName, nodeName, multiple))
					}
				case "enum":
					if enum, ok := value.([]interface{}); ok && len(enum) > 0 {
						enumValues := make([]string, len(enum))
						for i, e := range enum {
							enumValues[i] = fmt.Sprintf("%v", e)
						}
						kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
							fmt.Sprintf("        all item in %s {item in [%v]} if %s is not Undefined, \"%s must be one of: %s\"\n",
								nodeName, strings.Join(enumValues, ", "), nodeName, nodeName, strings.Join(enumValues, ", ")))
					}
				}
			}
		}
	case Object:
		kclSchema.KCLSchema[mainnode.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s: any\n", nodeName, optionalMarker))
	}

	return nil
}

func (g *SimplifiedTreeBasedGenerator) handleStringType(node *SchemaTreeNode, mainnode *SchemaTreeNode, nodeName string, optionalMarker string, kclSchema *Kclschema) {
	// Add string field
	kclSchema.KCLSchema[mainnode.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s: %s\n", nodeName, optionalMarker, "str"))

	// Skip if no constraints or format
	if len(node.Constraints) == 0 && node.Format == "" {
		return
	}

	// Initialize check block if needed
	if kclSchema.KCLSchema[mainnode.SchemaName].Check.Len() == 0 {
		kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString("\n    check:\n")
	}

	// Handle format-specific validation
	if node.Format != "" {
		g.handleStringFormat(node, mainnode, nodeName, kclSchema)
	}

	// Handle other constraints
	g.handleStringConstraints(node, mainnode, nodeName, kclSchema)
}

func (g *SimplifiedTreeBasedGenerator) handleStringFormat(node *SchemaTreeNode, mainnode *SchemaTreeNode, nodeName string, kclSchema *Kclschema) {
	switch node.Format {
	case "email":
		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		emailMessage := "Value must be a valid email address"
		kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
			fmt.Sprintf("        regex.match(%s, r\"%s\") if %s, \"%s\"\n",
				nodeName, emailPattern, nodeName, emailMessage))
	case "uri", "uri-template":
		uriPattern := `^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]`
		uriMessage := "Value must be a valid URI"
		kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
			fmt.Sprintf("        regex.match(%s, r\"%s\") if %s, \"%s\"\n",
				nodeName, uriPattern, nodeName, uriMessage))
	case "date-time":
		dateTimePattern := `^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):[0-5]\d:[0-5]\d(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$`
		dateTimeMessage := "Value must be a valid RFC 3339 date-time"
		dateValidation := "        value != None and value and datetime.validate(value[:10], \"%Y-%m-%d\"), \"Value contains an invalid date component\"\n"
		kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
			fmt.Sprintf("        regex.match(%s, r\"%s\") if %s, \"%s\"\n",
				nodeName, dateTimePattern, nodeName, dateTimeMessage))
		kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(dateValidation)
	case "ipv4":
		ipv4Pattern := `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
		ipv4Message := "Value must be a valid IPv4 address"
		kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
			fmt.Sprintf("        regex.match(%s, r\"%s\") if %s, \"%s\"\n",
				nodeName, ipv4Pattern, nodeName, ipv4Message))
	}
}

func (g *SimplifiedTreeBasedGenerator) handleStringConstraints(node *SchemaTreeNode, mainnode *SchemaTreeNode, nodeName string, kclSchema *Kclschema) {
	for constraint, value := range node.Constraints {
		switch constraint {
		case "minLength":
			if minLen, ok := value.(float64); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        len(%s) >= %d if %s is not Undefined, \"String must be at least %d characters\"\n",
						nodeName, int(minLen), nodeName, int(minLen)))
			}
		case "maxLength":
			if maxLen, ok := value.(float64); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        len(%s) <= %d if %s is not Undefined, \"String must be at most %d characters\"\n",
						nodeName, int(maxLen), nodeName, int(maxLen)))
			}
		case "pattern":
			if pattern, ok := value.(string); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        regex.match(%s, r\"%s\") if %s is not Undefined, \"String must match pattern %s\"\n",
						nodeName, pattern, nodeName, pattern))
			}
		case "enum":
			if enum, ok := value.([]interface{}); ok && len(enum) > 0 {
				enumValues := make([]string, len(enum))
				for i, e := range enum {
					enumValues[i] = fmt.Sprintf("%q", e)
				}
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        %s in [%s] if %s is not Undefined \n",
						nodeName, strings.Join(enumValues, ", "), nodeName))
			}
		}
	}
}

func (g *SimplifiedTreeBasedGenerator) handleNumericType(node *SchemaTreeNode, mainnode *SchemaTreeNode, nodeName string, optionalMarker string, kclType string, kclSchema *Kclschema) {
	// Add numeric field
	kclSchema.KCLSchema[mainnode.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s: %s\n", nodeName, optionalMarker, kclType))

	// Skip if no constraints
	if len(node.Constraints) == 0 {
		return
	}

	// Initialize check block if needed
	if kclSchema.KCLSchema[mainnode.SchemaName].Check.Len() == 0 {
		kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString("\n    check:\n")
	}

	g.handleNumericConstraints(node, mainnode, nodeName, kclSchema)
}

func (g *SimplifiedTreeBasedGenerator) handleNumericConstraints(node *SchemaTreeNode, mainnode *SchemaTreeNode, nodeName string, kclSchema *Kclschema) {
	for constraint, value := range node.Constraints {
		switch constraint {
		case "minimum":
			if min, ok := value.(float64); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        %s >= %g if %s is not Undefined, \"%s must be at least %g\"\n",
						nodeName, min, nodeName, nodeName, min))
			}
		case "maximum":
			if max, ok := value.(float64); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        %s <= %g if %s is not Undefined, \"%s must be at most %g\"\n",
						nodeName, max, nodeName, nodeName, max))
			}
		case "exclusiveMinimum":
			if min, ok := value.(float64); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        %s > %g if %s is not Undefined, \"%s must be greater than %g\"\n",
						nodeName, min, nodeName, nodeName, min))
			}
		case "exclusiveMaximum":
			if max, ok := value.(float64); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        %s < %g if %s is not Undefined, \"%s must be less than %g\"\n",
						nodeName, max, nodeName, nodeName, max))
			}
		case "multipleOf":
			if multiple, ok := value.(float64); ok {
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        %s %% %g == 0 if %s is not Undefined, \"%s must be a multiple of %g\"\n",
						nodeName, multiple, nodeName, nodeName, multiple))
			}
		case "enum":
			if enum, ok := value.([]interface{}); ok && len(enum) > 0 {
				enumValues := make([]string, len(enum))
				for i, e := range enum {
					enumValues[i] = fmt.Sprintf("%v", e)
				}
				kclSchema.KCLSchema[mainnode.SchemaName].Check.WriteString(
					fmt.Sprintf("        %s in [%v] if %s is not Undefined, \"%s must be one of: %s\"\n",
						nodeName, strings.Join(enumValues, ", "), nodeName, nodeName, strings.Join(enumValues, ", ")))
			}
		}
	}
}
func (g *SimplifiedTreeBasedGenerator) handleCompositeTypeProperies(node *SchemaTreeNode, nodeName string, parentNode *SchemaTreeNode, kclSchema *Kclschema, required map[string]bool) {
	optionalMarker := "?"
	if required[node.SchemaName] {
		optionalMarker = ""
	}
	switch node.Type {

	case "allOf":
		if _, ok := kclSchema.KCLSchema[parentNode.SchemaName]; !ok {
			kclSchema.KCLSchema[parentNode.SchemaName] = NewSchemaContent()
			kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s:\n", nodeName))
		}
		for _, subschema := range node.SubSchemas {
			subschema.SchemaName = g.addMixinSuffix(subschema.SchemaName)
			g.simplifiedGenerateSchemaFromNode(subschema, kclSchema)
		}
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s:\n", nodeName))
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString("    mixin [")
		for i, subschema := range node.SubSchemas {
			value := subschema.SchemaName
			if i < len(node.SubSchemas)-1 {
				value += ", "
			}
			kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(value)
		}
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString("]\n")

	case "anyOf":
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("%s:\n", nodeName))
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString("    any_of:\n")
		for _, subschema := range node.SubSchemas {
			kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("        - %s\n", subschema.SchemaName))
		}

	case "oneOf":
		if _, ok := kclSchema.KCLSchema[parentNode.SchemaName]; !ok {
			kclSchema.KCLSchema[parentNode.SchemaName] = NewSchemaContent()
			//There is no way to create schema with union operator. So just output schemas without top level schema
			// kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s:", nodeName))
		} else {
			//OneOf is in object property so schema is already initialized.
			kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("    %s%s:", nodeName, optionalMarker))
		}
		for i, subschema := range node.SubSchemas {
			unionOperator := ""
			if i < len(node.SubSchemas)-1 {
				unionOperator = "|"
			}
			if g.isExtendedPrimitiveType(*subschema) {
				switch subschema.Type {
				case "string":
					kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf(" str %s", unionOperator))
				case "integer":
					kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf(" int %s", unionOperator))
				case "number":
					kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf(" float %s", unionOperator))
				case "boolean":
					kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf(" bool %s", unionOperator))
				case "array":
					kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf(" [%s] %s", convertKCLType(subschema.Items.Type), unionOperator))
				}
			} else if subschema.Type == "array" {
				kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf(" [%s] %s", subschema.Items.SchemaName, unionOperator))
				g.simplifiedGenerateSchemaFromNode(subschema.Items, kclSchema)
			} else {
				kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf(" %s %s", subschema.SchemaName, unionOperator))
				g.simplifiedGenerateSchemaFromNode(subschema, kclSchema)
			}
		}
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString("\n")

	case "not":
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s:\n", node.SchemaName))
		kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString("    not:\n")
		for _, subschema := range node.SubSchemas {
			kclSchema.KCLSchema[parentNode.SchemaName].Fields.WriteString(fmt.Sprintf("        - %s\n", subschema.SchemaName))
		}
	}
}

func (g *SimplifiedTreeBasedGenerator) handleCompositeType(node *SchemaTreeNode, kclSchema *Kclschema, required map[string]bool) {
	switch node.Type {

	case "allOf":
		if _, ok := kclSchema.KCLSchema[node.SchemaName]; !ok {
			kclSchema.KCLSchema[node.SchemaName] = NewSchemaContent()
			kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s:\n", node.SchemaName))
		}
		for _, subschema := range node.SubSchemas {
			subschema.SchemaName = g.addMixinSuffix(subschema.SchemaName)
			g.simplifiedGenerateSchemaFromNode(subschema, kclSchema)
		}
		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    mixin [")
		for i, subschema := range node.SubSchemas {
			value := subschema.SchemaName
			if i < len(node.SubSchemas)-1 {
				value += ", "
			}
			kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(value)
		}
		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("]\n")

	case "anyOf":
		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("%s:\n", node.SchemaName))
		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("    any_of:\n")
		for _, subschema := range node.SubSchemas {
			kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("        - %s\n", subschema.SchemaName))
		}

	case "oneOf":
		if _, ok := kclSchema.KCLSchema[node.SchemaName]; !ok {
			kclSchema.KCLSchema[node.SchemaName] = NewSchemaContent()
		}
		for _, subschema := range node.SubSchemas {
			if g.isExtendedPrimitiveType(*subschema) {
				log.Default().Printf("subschema: %s is not supported in oneOf. Skipping creation!", subschema.SchemaName)
			} else if subschema.Type == "array" {
				log.Default().Printf("subschema: %s Array is not supported in oneOf. Will only create subschemas for items!", subschema.SchemaName)
				g.simplifiedGenerateSchemaFromNode(subschema.Items, kclSchema)
			} else {
				g.simplifiedGenerateSchemaFromNode(subschema, kclSchema)
			}
		}
		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString("\n")

	case "not":
		// Ensure the schema exists before accessing its fields
		if _, ok := kclSchema.KCLSchema[node.SchemaName]; !ok {
			kclSchema.KCLSchema[node.SchemaName] = NewSchemaContent()
		}
		
		// Write the schema declaration
		kclSchema.KCLSchema[node.SchemaName].Fields.WriteString(fmt.Sprintf("schema %s:\n", node.SchemaName))
		
		// Process the subschemas for the 'not' type
		if len(node.SubSchemas) > 0 {
			// In KCL, we can implement 'not' using a check block with a validation
			// that ensures the value doesn't match the negated schema
			kclSchema.KCLSchema[node.SchemaName].Check.WriteString("\n    check:\n")
			
			// For each subschema in the 'not', generate a validation that it doesn't match
			for _, subschema := range node.SubSchemas {
				// Generate the subschema first
				g.simplifiedGenerateSchemaFromNode(subschema, kclSchema)
				
				// Add a validation that ensures this value doesn't conform to the subschema
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString(
					fmt.Sprintf("        not is_%s(self), \"Value must not match schema %s\"\n", 
						strings.ToLower(subschema.SchemaName), subschema.SchemaName))
			}
			
			// Add helper functions to check against the negated schemas
			for _, subschema := range node.SubSchemas {
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString(
					fmt.Sprintf("\n    is_%s = lambda self -> bool {\n", strings.ToLower(subschema.SchemaName)))
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString(
					fmt.Sprintf("        schema = %s {}\n", subschema.SchemaName))
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString("        try:\n")
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString("            schema.check(self)\n")
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString("            return True  # Validation passed, so it matches\n")
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString("        except:\n")
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString("            return False  # Validation failed, so it doesn't match\n")
				kclSchema.KCLSchema[node.SchemaName].Check.WriteString("    }\n")
			}
		}
	}
}

// formatSchemaName formats a schema name for KCL
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

// determineImports determines the needed imports for a node
func (g *SimplifiedTreeBasedGenerator) determineImports(node *SchemaTreeNode, imports *[]string) {

	// Check for regex pattern or format that needs regex
	needsRegex := false

	if node.Type == String {
		if _, ok := node.Constraints["pattern"]; ok {
			needsRegex = true
		}

		if node.Format != "" {
			needsRegex = true
			if node.Format == "date-time" && !containsImport(*imports, "import datetime") {
				*imports = append(*imports, "import datetime")
			}
		}
	}
	if needsRegex && !containsImport(*imports, "import regex") {
		*imports = append(*imports, "import regex")
	}

	if node.Type == Object {
		for _, prop := range node.Properties {
			g.determineImports(prop, imports)
		}
	}
}

func containsImport(imports []string, importStr string) bool {
	for _, imp := range imports {
		if imp == importStr {
			return true
		}
	}
	return false
}

func convertKCLType(t NodeType) string {
	// Map JSON Schema types to KCL types
	switch t {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "[]"
	case "object":
		return "dict"
	default:
		return "any"
	}
}
