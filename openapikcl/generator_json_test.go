// generator_json_test.go
package openapikcl

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test various JSON Schema drafts
var jsonSchemaDrafts = []struct {
	name   string
	schema string
}{
	{
		name: "Draft-7",
		schema: `{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"properties": {
				"name": {
					"type": "string",
					"minLength": 1,
					"default": "default_name"
				},
				"age": {
					"type": "integer",
					"minimum": 0,
					"default": 25
				},
				"isActive": {
					"type": "boolean",
					"default": true
				},
				"status": {
					"type": "string",
					"enum": ["active", "inactive", "pending"],
					"default": "active"
				},
				"tags": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			},
			"required": ["name"]
		}`,
	},
	{
		name: "Draft-2019-09",
		schema: `{
			"$schema": "https://json-schema.org/draft/2019-09/schema",
			"type": "object",
			"properties": {
				"name": {
					"type": "string",
					"minLength": 1,
					"default": "default_name"
				},
				"age": {
					"type": "integer",
					"minimum": 0,
					"default": 25
				},
				"isActive": {
					"type": "boolean",
					"default": true
				},
				"status": {
					"type": "string",
					"enum": ["active", "inactive", "pending"],
					"default": "active"
				},
				"tags": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			},
			"required": ["name"]
		}`,
	},
}

func TestJSONSchemaToKCL(t *testing.T) {
	for _, draft := range jsonSchemaDrafts {
		t.Run(draft.name, func(t *testing.T) {
			// Parse JSON Schema
			var schemaData map[string]interface{}
			err := json.Unmarshal([]byte(draft.schema), &schemaData)
			require.NoError(t, err)

			// Extract default values before compilation
			defaultValues := make(map[string]interface{})
			if props, ok := schemaData["properties"].(map[string]interface{}); ok {
				for propName, propData := range props {
					if propObj, ok := propData.(map[string]interface{}); ok {
						if defaultVal, hasDefault := propObj["default"]; hasDefault {
							defaultValues[propName] = defaultVal
						}
					}
				}
			}

			// Compile schema
			compiler := jsonschema.NewCompiler()
			schemaID := "test-schema"
			err = compiler.AddResource(schemaID, strings.NewReader(draft.schema))
			require.NoError(t, err)
			schema, err := compiler.Compile(schemaID)
			require.NoError(t, err)

			// Generate KCL from JSON Schema, passing default values
			kclSchema, err := generateJSONSchemaToKCLWithDefaults("TestSchema", schema, defaultValues)
			require.NoError(t, err)

			// Verify the schema has the expected content
			assert.Contains(t, kclSchema, "# No schema imports needed - schemas in same directory")
			assert.Contains(t, kclSchema, "schema TestSchema:")

			// Check for fields with default values
			assert.Contains(t, kclSchema, "name: str = \"default_name\"")
			assert.Contains(t, kclSchema, "age?: int = 25")
			assert.Contains(t, kclSchema, "isActive?: bool = True")
			assert.Contains(t, kclSchema, "status?: str = \"active\"")
			assert.Contains(t, kclSchema, "tags?: [str]")
			assert.NotContains(t, kclSchema, "import regex")
		})
	}
}

func TestJSONSchemaGeneration(t *testing.T) {
	// Skip if running in CI without tempdir access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_TEMPDIR_TESTS") != "" {
		t.Skip("Skipping test requiring tempdir in CI")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "kcl-json-schema-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	for _, draft := range jsonSchemaDrafts {
		t.Run(draft.name, func(t *testing.T) {
			// Parse JSON Schema
			var schemaData map[string]interface{}
			err := json.Unmarshal([]byte(draft.schema), &schemaData)
			require.NoError(t, err)

			// Generate KCL schemas from JSON Schema
			err = generateJSONSchemas(schemaData, tempDir, "test")
			require.NoError(t, err)

			// Check if a schema file was created (Schema.k)
			schemaPath := filepath.Join(tempDir, "Schema.k")
			assert.True(t, fileExists(schemaPath), "Schema file should exist")

			// Read and verify basic content
			content, err := os.ReadFile(schemaPath)
			require.NoError(t, err)
			contentStr := string(content)

			// Debug - print the full content
			t.Logf("Generated Schema.k content:\n%s", contentStr)

			assert.Contains(t, contentStr, "schema Schema:")
			assert.Contains(t, contentStr, "name: str = \"default_name\"")
			assert.Contains(t, contentStr, "age?: int = 25")
			assert.Contains(t, contentStr, "isActive?: bool = True")
			assert.Contains(t, contentStr, "tags?: [str]")

			// Verify main.k was created
			mainKPath := filepath.Join(tempDir, "main.k")
			assert.True(t, fileExists(mainKPath))

			// Test KCL validation if available
			if isKCLAvailable() {
				// Run KCL against all files in the directory
				cmd := exec.Command("kcl", ".")
				cmd.Dir = tempDir
				output, err := cmd.CombinedOutput()
				assert.NoError(t, err, "KCL validation should succeed: %s", string(output))
			}
		})
	}
}

func TestJSONSchemaTypeConversion(t *testing.T) {
	tests := []struct {
		jsonType    string
		expectedKCL string
		description string
	}{
		{"string", "str", "string type should convert to str"},
		{"integer", "int", "integer type should convert to int"},
		{"number", "float", "number type should convert to float"},
		{"boolean", "bool", "boolean type should convert to bool"},
		{"array", "[any]", "array type without items should convert to [any]"},
		{"object", "dict", "object type should convert to dict"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Create a simple schema of the given type
			schemaStr := fmt.Sprintf(`{"type": "%s"}`, tt.jsonType)
			compiler := jsonschema.NewCompiler()
			schemaID := "test-type"
			err := compiler.AddResource(schemaID, strings.NewReader(schemaStr))
			require.NoError(t, err)
			schema, err := compiler.Compile(schemaID)
			require.NoError(t, err)

			// Test the type conversion
			kclType := jsonSchemaTypeToKCL(schema)
			assert.Equal(t, tt.expectedKCL, kclType)
		})
	}
}
