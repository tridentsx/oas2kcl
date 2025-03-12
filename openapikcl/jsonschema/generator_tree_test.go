package jsonschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSchemaTree(t *testing.T) {
	// Sample schema for testing
	sampleSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The person's name",
				"minLength":   2,
			},
			"age": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
			},
			"email": map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
			"addresses": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"street": map[string]interface{}{
							"type": "string",
						},
						"city": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
		"required": []interface{}{"name", "email"},
	}

	// Build the schema tree
	tree, err := BuildSchemaTree(sampleSchema, "Person", nil)
	if err != nil {
		t.Fatalf("Failed to build schema tree: %v", err)
	}

	// Verify the tree structure
	if tree.Type != Object {
		t.Errorf("Expected root node type to be Object, got %s", tree.Type)
	}

	if tree.SchemaName != "Person" {
		t.Errorf("Expected root node name to be Person, got %s", tree.SchemaName)
	}

	// Check properties
	if len(tree.Properties) != 4 {
		t.Errorf("Expected 4 properties, got %d", len(tree.Properties))
	}

	// Check name property
	nameNode, ok := tree.Properties["name"]
	if !ok {
		t.Fatalf("Expected name property")
	}
	if nameNode.Type != String {
		t.Errorf("Expected name property type to be String, got %s", nameNode.Type)
	}
	if nameNode.Description != "The person's name" {
		t.Errorf("Expected name description to be 'The person's name', got %s", nameNode.Description)
	}
	if minLength, exists := nameNode.Constraints["minLength"]; !exists {
		t.Errorf("Expected name to have minLength constraint")
	} else {
		var minLengthValue float64
		switch v := minLength.(type) {
		case int:
			minLengthValue = float64(v)
		case float64:
			minLengthValue = v
		default:
			t.Errorf("Expected minLength to be numeric, got %T", minLength)
		}

		if minLengthValue != 2 {
			t.Errorf("Expected name minLength to be 2, got %v", minLengthValue)
		}
	}

	// Check email property
	emailNode, ok := tree.Properties["email"]
	if !ok {
		t.Fatalf("Expected email property")
	}
	if emailNode.Type != String {
		t.Errorf("Expected email property type to be String, got %s", emailNode.Type)
	}
	if emailNode.Format != "email" {
		t.Errorf("Expected email format to be 'email', got %s", emailNode.Format)
	}

	// Check addresses property
	addressesNode, ok := tree.Properties["addresses"]
	if !ok {
		t.Fatalf("Expected addresses property")
	}
	if addressesNode.Type != Array {
		t.Errorf("Expected addresses property type to be Array, got %s", addressesNode.Type)
	}
	if addressesNode.Items == nil {
		t.Fatalf("Expected addresses items")
	}
	if addressesNode.Items.Type != Object {
		t.Errorf("Expected addresses items type to be Object, got %s", addressesNode.Items.Type)
	}
}

func TestGenerateKCLSchemasFromTree(t *testing.T) {
	// Only run this test if the TEST_GENERATE_FILES environment variable is set
	if os.Getenv("TEST_GENERATE_FILES") == "" {
		t.Skip("Skipping test that generates files")
	}

	// Sample schema for testing
	sampleSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The person's name",
				"minLength":   2,
			},
			"age": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
			},
			"email": map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
			"addresses": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"street": map[string]interface{}{
							"type": "string",
						},
						"city": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
		"required": []interface{}{"name", "email"},
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "kcl-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build the schema tree
	tree, err := BuildSchemaTree(sampleSchema, "Person", nil)
	if err != nil {
		t.Fatalf("Failed to build schema tree: %v", err)
	}

	// Generate KCL schemas from the tree
	generator := NewTreeBasedGenerator(tempDir)
	files, err := generator.GenerateKCLSchemasFromTree(tree)
	if err != nil {
		t.Fatalf("Failed to generate KCL schemas: %v", err)
	}

	// Verify the generated files
	if len(files) == 0 {
		t.Error("No files were generated")
	}

	// Check if the main schema file exists
	mainSchemaPath := filepath.Join(tempDir, "Person.k")
	if _, err := os.Stat(mainSchemaPath); os.IsNotExist(err) {
		t.Errorf("Expected main schema file %s to exist", mainSchemaPath)
	}

	// Optional: Print the generated schema file contents for debugging
	if content, err := os.ReadFile(mainSchemaPath); err == nil {
		t.Logf("Generated schema file contents:\n%s", string(content))
	}
}

func TestGenerateSchemaTreeAndKCL(t *testing.T) {
	// Only run this test if the TEST_GENERATE_FILES environment variable is set
	if os.Getenv("TEST_GENERATE_FILES") == "" {
		t.Skip("Skipping test that generates files")
	}

	// Sample schema for testing
	sampleSchema := map[string]interface{}{
		"type":  "object",
		"title": "Person",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The person's name",
				"minLength":   2,
			},
			"age": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
			},
			"email": map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
		},
		"required": []interface{}{"name", "email"},
	}

	// Convert schema to JSON bytes
	schemaBytes, err := json.Marshal(sampleSchema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "kcl-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate KCL schemas (with debug mode disabled)
	err = GenerateSchemaTreeAndKCL(schemaBytes, tempDir, false)
	if err != nil {
		t.Fatalf("Failed to generate KCL schemas: %v", err)
	}

	// Verify the generated files
	expectedFiles := []string{"Person.k", "Email.k"}
	for _, filename := range expectedFiles {
		path := filepath.Join(tempDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", path)
		}
	}

	// Optional: Print the generated schema file contents for debugging
	mainSchemaPath := filepath.Join(tempDir, "Person.k")
	if content, err := os.ReadFile(mainSchemaPath); err == nil {
		t.Logf("Generated main schema file contents:\n%s", string(content))
	}
}
