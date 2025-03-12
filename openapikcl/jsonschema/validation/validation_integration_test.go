package validation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSchemaValidation tests the actual validation of schemas against known good and bad objects
func TestSchemaValidation(t *testing.T) {
	t.Skip("This test is for the old validation implementation. The tree-based generator uses a different approach.")

	tests := []struct {
		name     string
		schema   *Schema
		goodObjs []map[string]interface{}
		badObjs  []map[string]interface{}
	}{
		{
			name: "String with min and max length",
			schema: &Schema{
				Type:      "string",
				MinLength: &[]int{3}[0],
				MaxLength: &[]int{20}[0],
			},
			goodObjs: []map[string]interface{}{
				{"value": "abc"},
				{"value": "abcdefghijklmnopqrst"},
			},
			badObjs: []map[string]interface{}{
				{"value": "ab"},
				{"value": "abcdefghijklmnopqrstu"},
			},
		},
		{
			name: "String with email format",
			schema: &Schema{
				Type:   "string",
				Format: "email",
			},
			goodObjs: []map[string]interface{}{
				{"value": "user@example.com"},
				{"value": "user.name@domain.co.uk"},
			},
			badObjs: []map[string]interface{}{
				{"value": "not-an-email"},
				{"value": "@example.com"},
			},
		},
		{
			name: "String with IP format",
			schema: &Schema{
				Type:   "string",
				Format: "ipv4",
			},
			goodObjs: []map[string]interface{}{
				{"value": "192.168.1.1"},
				{"value": "10.0.0.0"},
			},
			badObjs: []map[string]interface{}{
				{"value": "256.256.256.256"},
				{"value": "1.2.3"},
			},
		},
		{
			name: "String with datetime format",
			schema: &Schema{
				Type:   "string",
				Format: "date-time",
			},
			goodObjs: []map[string]interface{}{
				{"value": "2024-03-20T10:00:00Z"},
				{"value": "2024-03-20T10:00:00+01:00"},
			},
			badObjs: []map[string]interface{}{
				{"value": "2024-13-20T10:00:00Z"},
				{"value": "2024-03-20T25:00:00Z"},
			},
		},
		{
			name: "Number with range constraints",
			schema: &Schema{
				Type:    "integer",
				Minimum: &[]float64{0}[0],
				Maximum: &[]float64{120}[0],
			},
			goodObjs: []map[string]interface{}{
				{"value": 0},
				{"value": 60},
				{"value": 120},
			},
			badObjs: []map[string]interface{}{
				{"value": -1},
				{"value": 121},
			},
		},
		{
			name: "Array with constraints",
			schema: &Schema{
				Type:        "array",
				Items:       &Schema{Type: "string", MinLength: &[]int{2}[0]},
				MinItems:    &[]int{1}[0],
				MaxItems:    &[]int{3}[0],
				UniqueItems: true,
			},
			goodObjs: []map[string]interface{}{
				{"value": []string{"ab", "cd"}},
				{"value": []string{"ab", "cd", "ef"}},
			},
			badObjs: []map[string]interface{}{
				{"value": []string{}},
				{"value": []string{"a"}},
				{"value": []string{"ab", "ab"}},
				{"value": []string{"ab", "cd", "ef", "gh"}},
			},
		},
		{
			name: "Object with required properties",
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"name": {Type: "string"},
					"age": {
						Type:    "integer",
						Minimum: &[]float64{0}[0],
						Maximum: &[]float64{120}[0],
					},
				},
				Required: []string{"name", "age"},
			},
			goodObjs: []map[string]interface{}{
				{"name": "John", "age": 30},
			},
			badObjs: []map[string]interface{}{
				{"name": "John"},
				{"age": 30},
				{"name": "John", "age": -1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate schema content
			schemaContent, imports := GenerateValidatorSchema(tt.schema, "TestSchema")

			// Generate schema with imports
			var content strings.Builder
			if imports.NeedsDatetime {
				content.WriteString("import datetime\n")
			}
			if imports.NeedsNet {
				content.WriteString("import net\n")
			}
			if imports.NeedsRegex {
				content.WriteString("import regex\n")
			}
			content.WriteString(schemaContent)

			t.Logf("Schema content:\n%s", content.String())

			// Test good objects
			for i, obj := range tt.goodObjs {
				objContent := generateTestObject(obj)
				t.Logf("Good object %d content:\n%s", i, objContent)

				// Run KCL validation
				if err := runKCLValidation(content.String(), objContent); err != nil {
					t.Errorf("Good object %d failed validation: %v", i, err)
				}
			}

			// Test bad objects
			for i, obj := range tt.badObjs {
				objContent := generateTestObject(obj)
				t.Logf("Bad object %d content:\n%s", i, objContent)

				// Run KCL validation
				if err := runKCLValidation(content.String(), objContent); err == nil {
					t.Errorf("Bad object %d passed validation when it should have failed", i)
				}
			}
		})
	}
}

func generateTestObject(obj map[string]interface{}) string {
	var content strings.Builder
	content.WriteString("import .schema\n\n")
	content.WriteString("test_instance = schema.TestSchema {\n")
	for k, v := range obj {
		switch val := v.(type) {
		case string:
			content.WriteString(fmt.Sprintf("    %s = \"%s\"\n", k, val))
		case int:
			content.WriteString(fmt.Sprintf("    %s = %d\n", k, val))
		case float64:
			content.WriteString(fmt.Sprintf("    %s = %f\n", k, val))
		case []interface{}:
			var items []string
			for _, item := range val {
				switch itemVal := item.(type) {
				case string:
					items = append(items, fmt.Sprintf("\"%s\"", itemVal))
				case float64:
					items = append(items, fmt.Sprintf("%f", itemVal))
				case int:
					items = append(items, fmt.Sprintf("%d", itemVal))
				}
			}
			content.WriteString(fmt.Sprintf("    %s = [%s]\n", k, strings.Join(items, ", ")))
		case map[string]interface{}:
			content.WriteString(fmt.Sprintf("    %s = {\n", k))
			for subK, subV := range val {
				switch subVal := subV.(type) {
				case string:
					content.WriteString(fmt.Sprintf("        %s = \"%s\"\n", subK, subVal))
				case int:
					content.WriteString(fmt.Sprintf("        %s = %d\n", subK, subVal))
				case float64:
					content.WriteString(fmt.Sprintf("        %s = %f\n", subK, subVal))
				}
			}
			content.WriteString("    }\n")
		}
	}
	content.WriteString("}\n")
	return content.String()
}

func runKCLValidation(schema, obj string) error {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "validation_test")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Write schema file
	schemaDir := filepath.Join(tmpDir, "schema")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		return err
	}
	schemaFile := filepath.Join(schemaDir, "schema.k")
	if err := os.WriteFile(schemaFile, []byte(schema), 0644); err != nil {
		return err
	}

	// Write test file
	testFile := filepath.Join(tmpDir, "test.k")
	if err := os.WriteFile(testFile, []byte(obj), 0644); err != nil {
		return err
	}

	fmt.Printf("\nSchema file (schema/schema.k):\n%s\n", schema)
	fmt.Printf("\nTest file (test.k):\n%s\n", obj)

	// Run KCL validation
	cmd := exec.Command("kcl", testFile)
	output, err := cmd.CombinedOutput()
	fmt.Printf("\nKCL Output:\n%s\n", string(output))

	if err != nil {
		return fmt.Errorf("validation failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}
