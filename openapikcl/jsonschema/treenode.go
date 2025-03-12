// Package jsonschema provides functionality for converting JSON Schema to KCL.
package jsonschema

import (
	"fmt"
)

// NodeType represents the type of a SchemaTreeNode
type NodeType string

const (
	// Object represents a JSON Schema object type
	Object NodeType = "object"
	// Array represents a JSON Schema array type
	Array NodeType = "array"
	// String represents a JSON Schema string type
	String NodeType = "string"
	// Number represents a JSON Schema number type
	Number NodeType = "number"
	// Integer represents a JSON Schema integer type
	Integer NodeType = "integer"
	// Boolean represents a JSON Schema boolean type
	Boolean NodeType = "boolean"
	// Null represents a JSON Schema null type
	Null NodeType = "null"

	// Composition types
	AllOf NodeType = "allOf"
	AnyOf NodeType = "anyOf"
	OneOf NodeType = "oneOf"
	Not   NodeType = "not"
	If    NodeType = "if"
	Then  NodeType = "then"
	Else  NodeType = "else"

	// Reference represents a reference to another schema
	Reference NodeType = "reference"
)

// SchemaTreeNode represents a node in the JSON Schema tree
type SchemaTreeNode struct {
	// Type of the node (object, array, string, etc.)
	Type NodeType

	// Name of the schema represented by this node
	SchemaName string

	// Original JSON Schema definition for this node
	RawSchema map[string]interface{}

	// Parent node (nil for root)
	Parent *SchemaTreeNode

	// For object nodes: child properties
	Properties map[string]*SchemaTreeNode

	// For object nodes: pattern properties
	PatternProperties map[string]*SchemaTreeNode

	// For array nodes: type of items
	Items *SchemaTreeNode

	// For reference nodes: the target schema name
	RefTarget string

	// For composition nodes: the subschemas
	SubSchemas []*SchemaTreeNode

	// Constraints from the schema
	Constraints map[string]interface{}

	// Metadata
	Description string
	Format      string
	Title       string
	Default     interface{}
}

// NewSchemaTreeNode creates a new SchemaTreeNode
func NewSchemaTreeNode(nodeType NodeType, schemaName string, rawSchema map[string]interface{}) *SchemaTreeNode {
	return &SchemaTreeNode{
		Type:              nodeType,
		SchemaName:        schemaName,
		RawSchema:         rawSchema,
		Properties:        make(map[string]*SchemaTreeNode),
		PatternProperties: make(map[string]*SchemaTreeNode),
		Constraints:       make(map[string]interface{}),
	}
}

// AddProperty adds a property to an object node
func (n *SchemaTreeNode) AddProperty(name string, property *SchemaTreeNode) error {
	if n.Type != Object {
		return fmt.Errorf("cannot add property to non-object node")
	}
	n.Properties[name] = property
	property.Parent = n
	return nil
}

// SetItems sets the items type for an array node
func (n *SchemaTreeNode) SetItems(items *SchemaTreeNode) error {
	if n.Type != Array {
		return fmt.Errorf("cannot set items on non-array node")
	}
	n.Items = items
	items.Parent = n
	return nil
}

// AddSubSchema adds a subschema to a composition node
func (n *SchemaTreeNode) AddSubSchema(subSchema *SchemaTreeNode) error {
	if n.Type != AllOf && n.Type != AnyOf && n.Type != OneOf && n.Type != Not {
		return fmt.Errorf("cannot add subschema to non-composition node")
	}
	n.SubSchemas = append(n.SubSchemas, subSchema)
	subSchema.Parent = n
	return nil
}

// SetRefTarget sets the reference target for a reference node
func (n *SchemaTreeNode) SetRefTarget(target string) error {
	if n.Type != Reference {
		return fmt.Errorf("cannot set reference target on non-reference node")
	}
	n.RefTarget = target
	return nil
}

// BuildSchemaTree builds a schema tree from a raw JSON Schema
func BuildSchemaTree(rawSchema map[string]interface{}, schemaName string, processedRefs map[string]bool) (*SchemaTreeNode, error) {
	if processedRefs == nil {
		processedRefs = make(map[string]bool)
	}

	// Check for $ref first
	if ref, ok := rawSchema["$ref"].(string); ok {
		refNode := NewSchemaTreeNode(Reference, schemaName, rawSchema)
		err := refNode.SetRefTarget(ref)
		if err != nil {
			return nil, err
		}
		return refNode, nil
	}

	// Check for composite schemas
	if allOf, ok := rawSchema["allOf"].([]interface{}); ok {
		allOfNode := NewSchemaTreeNode(AllOf, schemaName, rawSchema)
		for i, subSchema := range allOf {
			if subSchemaMap, ok := subSchema.(map[string]interface{}); ok {
				subNode, err := BuildSchemaTree(subSchemaMap, fmt.Sprintf("%s_allOf_%d", schemaName, i), processedRefs)
				if err != nil {
					return nil, err
				}
				err = allOfNode.AddSubSchema(subNode)
				if err != nil {
					return nil, err
				}
			}
		}
		return allOfNode, nil
	}

	if anyOf, ok := rawSchema["anyOf"].([]interface{}); ok {
		anyOfNode := NewSchemaTreeNode(AnyOf, schemaName, rawSchema)
		for i, subSchema := range anyOf {
			if subSchemaMap, ok := subSchema.(map[string]interface{}); ok {
				subNode, err := BuildSchemaTree(subSchemaMap, fmt.Sprintf("%s_anyOf_%d", schemaName, i), processedRefs)
				if err != nil {
					return nil, err
				}
				err = anyOfNode.AddSubSchema(subNode)
				if err != nil {
					return nil, err
				}
			}
		}
		return anyOfNode, nil
	}

	if oneOf, ok := rawSchema["oneOf"].([]interface{}); ok {
		oneOfNode := NewSchemaTreeNode(OneOf, schemaName, rawSchema)
		for i, subSchema := range oneOf {
			if subSchemaMap, ok := subSchema.(map[string]interface{}); ok {
				subNode, err := BuildSchemaTree(subSchemaMap, fmt.Sprintf("%s_oneOf_%d", schemaName, i), processedRefs)
				if err != nil {
					return nil, err
				}
				err = oneOfNode.AddSubSchema(subNode)
				if err != nil {
					return nil, err
				}
			}
		}
		return oneOfNode, nil
	}

	if not, ok := rawSchema["not"].(map[string]interface{}); ok {
		notNode := NewSchemaTreeNode(Not, schemaName, rawSchema)
		subNode, err := BuildSchemaTree(not, fmt.Sprintf("%s_not", schemaName), processedRefs)
		if err != nil {
			return nil, err
		}
		err = notNode.AddSubSchema(subNode)
		if err != nil {
			return nil, err
		}
		return notNode, nil
	}

	// Handle if-then-else
	hasIf := false
	if ifSchema, ok := rawSchema["if"].(map[string]interface{}); ok {
		hasIf = true
		ifNode := NewSchemaTreeNode(If, schemaName+"_if", rawSchema)
		subNode, err := BuildSchemaTree(ifSchema, fmt.Sprintf("%s_if", schemaName), processedRefs)
		if err != nil {
			return nil, err
		}
		err = ifNode.AddSubSchema(subNode)
		if err != nil {
			return nil, err
		}

		// Check for then clause
		if thenSchema, ok := rawSchema["then"].(map[string]interface{}); ok {
			thenNode := NewSchemaTreeNode(Then, schemaName+"_then", rawSchema)
			subNode, err := BuildSchemaTree(thenSchema, fmt.Sprintf("%s_then", schemaName), processedRefs)
			if err != nil {
				return nil, err
			}
			err = thenNode.AddSubSchema(subNode)
			if err != nil {
				return nil, err
			}
		}

		// Check for else clause
		if elseSchema, ok := rawSchema["else"].(map[string]interface{}); ok {
			elseNode := NewSchemaTreeNode(Else, schemaName+"_else", rawSchema)
			subNode, err := BuildSchemaTree(elseSchema, fmt.Sprintf("%s_else", schemaName), processedRefs)
			if err != nil {
				return nil, err
			}
			err = elseNode.AddSubSchema(subNode)
			if err != nil {
				return nil, err
			}
		}

		// If we have if-then-else, return an AllOf node containing all parts
		if hasIf {
			return ifNode, nil
		}
	}

	// Get the type from the schema
	var nodeType NodeType
	if typeVal, ok := rawSchema["type"].(string); ok {
		nodeType = NodeType(typeVal)
	} else if typeArr, ok := rawSchema["type"].([]interface{}); ok {
		// For multiple types, use the first one for now
		// We could handle this better later
		if len(typeArr) > 0 {
			if typeStr, ok := typeArr[0].(string); ok {
				nodeType = NodeType(typeStr)
			}
		}
	} else {
		// Default to object if no type specified
		nodeType = Object
	}

	// Create a node based on the type
	node := NewSchemaTreeNode(nodeType, schemaName, rawSchema)

	// Set metadata
	if title, ok := rawSchema["title"].(string); ok {
		node.Title = title
	}
	if description, ok := rawSchema["description"].(string); ok {
		node.Description = description
	}
	if format, ok := rawSchema["format"].(string); ok {
		node.Format = format
	}
	if defaultVal, ok := rawSchema["default"]; ok {
		node.Default = defaultVal
	}

	// Extract constraints
	for _, constraintKey := range []string{
		"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum",
		"minLength", "maxLength", "pattern",
		"minItems", "maxItems", "uniqueItems",
		"minProperties", "maxProperties",
		"enum", "const", "multipleOf",
	} {
		if val, ok := rawSchema[constraintKey]; ok {
			node.Constraints[constraintKey] = val
		}
	}

	// Process properties for object types
	if nodeType == Object {
		if props, ok := rawSchema["properties"].(map[string]interface{}); ok {
			for propName, propSchema := range props {
				if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
					propNode, err := BuildSchemaTree(propSchemaMap, fmt.Sprintf("%s_%s", schemaName, propName), processedRefs)
					if err != nil {
						return nil, err
					}
					err = node.AddProperty(propName, propNode)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		// Process pattern properties for object types
		if patternProps, ok := rawSchema["patternProperties"].(map[string]interface{}); ok {
			for pattern, propSchema := range patternProps {
				if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
					// Create a sanitized name for the pattern property schema
					patternName := fmt.Sprintf("%s_pattern_%s", schemaName, sanitizePatternName(pattern))
					propNode, err := BuildSchemaTree(propSchemaMap, patternName, processedRefs)
					if err != nil {
						return nil, err
					}
					// Store the original pattern in the constraints
					propNode.Constraints["pattern"] = pattern
					// Add the pattern property to the node
					node.PatternProperties[pattern] = propNode
				}
			}
		}
	}

	// Process items for array types
	if nodeType == Array {
		if items, ok := rawSchema["items"].(map[string]interface{}); ok {
			itemsNode, err := BuildSchemaTree(items, fmt.Sprintf("%s_items", schemaName), processedRefs)
			if err != nil {
				return nil, err
			}
			err = node.SetItems(itemsNode)
			if err != nil {
				return nil, err
			}
		}
	}

	return node, nil
}

// sanitizePatternName creates a valid KCL identifier from a regex pattern
func sanitizePatternName(pattern string) string {
	// Replace non-alphanumeric characters with underscores
	sanitized := ""
	for _, r := range pattern {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			sanitized += string(r)
		} else {
			sanitized += "_"
		}
	}

	// Ensure it's not empty
	if sanitized == "" {
		return "pattern"
	}

	// Ensure it doesn't start with a digit
	if sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "_" + sanitized
	}

	return sanitized
}
