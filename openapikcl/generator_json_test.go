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
			assert.Contains(t, kclSchema, "tags?: [any]")
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
			assert.Contains(t, contentStr, "tags?: [any]")

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
		{"object", "{str:any}", "object type should convert to {str:any}"},
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

func TestJSONSchemaFormatValidation(t *testing.T) {
	// Test schema with various formats
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"email": {
				"type": "string",
				"format": "email"
			},
			"ipv4": {
				"type": "string",
				"format": "ipv4"
			},
			"ipv6": {
				"type": "string",
				"format": "ipv6"
			},
			"uri": {
				"type": "string",
				"format": "uri"
			},
			"date": {
				"type": "string",
				"format": "date"
			},
			"datetime": {
				"type": "string",
				"format": "date-time"
			},
			"uuid": {
				"type": "string",
				"format": "uuid"
			}
		}
	}`

	// Parse JSON Schema
	var schemaData map[string]interface{}
	err := json.Unmarshal([]byte(schemaStr), &schemaData)
	require.NoError(t, err)

	// Compile schema
	compiler := jsonschema.NewCompiler()
	schemaID := "test-format-schema"
	err = compiler.AddResource(schemaID, strings.NewReader(schemaStr))
	require.NoError(t, err)
	schema, err := compiler.Compile(schemaID)
	require.NoError(t, err)

	// Generate KCL from JSON Schema
	kclSchema, err := generateJSONSchemaToKCLWithDefaults("FormatSchema", schema, nil)
	require.NoError(t, err)

	// Debug - print the generated schema
	t.Logf("Generated KCL:\n%s", kclSchema)

	// Verify import
	assert.Contains(t, kclSchema, "import regex")

	// Verify format validations
	assert.Contains(t, kclSchema, "regex.match(email, r")
	assert.Contains(t, kclSchema, "regex.match(ipv4, r")
	assert.Contains(t, kclSchema, "regex.match(ipv6, r")
	assert.Contains(t, kclSchema, "regex.match(uri, r")
	assert.Contains(t, kclSchema, "regex.match(date, r")
	assert.Contains(t, kclSchema, "regex.match(datetime, r")
	assert.Contains(t, kclSchema, "regex.match(uuid, r")
}

func TestComplexSchemaGeneration(t *testing.T) {
	// Skip if running in CI without file access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_FILE_TESTS") != "" {
		t.Skip("Skipping test requiring file access in CI")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "kcl-complex-schema-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock complex schema instead of reading values.schema.json
	schemaData := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"title":   "CHA_Core_json_schema_definition",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":    "string",
				"title":   "Name",
				"default": "test-name",
				"examples": []interface{}{
					"test-name",
				},
			},
			"description": map[string]interface{}{
				"type":  "string",
				"title": "Description",
			},
			"version": map[string]interface{}{
				"type":    "string",
				"title":   "Version",
				"default": "1.0.0",
				"pattern": "^[0-9]+\\.[0-9]+\\.[0-9]+$",
			},
			"enabled": map[string]interface{}{
				"type":    "boolean",
				"title":   "Enabled",
				"default": true,
			},
			"replicas": map[string]interface{}{
				"type":    "integer",
				"title":   "Replicas",
				"default": 1,
				"minimum": 0,
				"maximum": 10,
			},
			"config": map[string]interface{}{
				"type":  "object",
				"title": "Config",
				"properties": map[string]interface{}{
					"timeout": map[string]interface{}{
						"type":    "integer",
						"title":   "Timeout",
						"default": 30,
					},
					"retries": map[string]interface{}{
						"type":    "integer",
						"title":   "Retries",
						"default": 3,
					},
				},
			},
			"tags": map[string]interface{}{
				"type":  "array",
				"title": "Tags",
				"items": map[string]interface{}{
					"type": "string",
				},
				"default": []interface{}{"tag1", "tag2"},
			},
		},
		"required": []interface{}{
			"name",
			"version",
		},
	}

	// Generate KCL schemas from the complex JSON Schema
	err = generateJSONSchemas(schemaData, tempDir, "test")
	require.NoError(t, err)

	// Check the generated files
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)

	// There should be at least two files (schema and main.k)
	assert.GreaterOrEqual(t, len(files), 2, "Should generate at least two files")

	// Log the generated files
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(tempDir, file.Name()))
		require.NoError(t, err)
		t.Logf("Generated file %s:\n%s", file.Name(), string(content))
	}
}

func TestCertManagerSchemaGeneration(t *testing.T) {
	// Skip if running in CI without file access
	if os.Getenv("CI") != "" && os.Getenv("SKIP_FILE_TESTS") != "" {
		t.Skip("Skipping test requiring file access in CI")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "kcl-certmanager-schema-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Read the certmanager schema file
	schemaPath := "testdata/jsonschema/certmanager.values.schema.json"
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read certmanager schema file: %v", err)
	}

	// Parse the schema
	var schemaData map[string]interface{}
	err = json.Unmarshal(data, &schemaData)
	require.NoError(t, err)

	// Generate KCL schemas from the certmanager JSON Schema
	t.Logf("Generating KCL schemas from certmanager schema")
	err = generateJSONSchemas(schemaData, tempDir, "certmanager")
	require.NoError(t, err)

	// Check the generated files
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)

	// Log the number of generated files
	t.Logf("Generated %d files", len(files))

	// There should be at least two files (schema and main.k)
	assert.GreaterOrEqual(t, len(files), 2, "Should generate at least two files")

	// Log the names of the generated files
	for _, file := range files {
		t.Logf("Generated file: %s", file.Name())
	}

	// Find the main schema file
	var mainSchemaFile os.DirEntry
	for _, file := range files {
		if file.Name() != "main.k" && file.Name() != "validation_test.k" && strings.HasSuffix(file.Name(), ".k") {
			mainSchemaFile = file
			break
		}
	}

	// Verify we found a main schema file
	require.NotNil(t, mainSchemaFile, "Could not find main schema file")
	t.Logf("Main schema file: %s", mainSchemaFile.Name())

	// Read the main schema file to check its content
	mainSchemaContent, err := os.ReadFile(filepath.Join(tempDir, mainSchemaFile.Name()))
	require.NoError(t, err)

	contentStr := string(mainSchemaContent)

	// Log a sample of the content
	if len(contentStr) > 500 {
		t.Logf("Main schema file content sample (first 500 chars):\n%s", contentStr[:500])
	} else {
		t.Logf("Main schema file content:\n%s", contentStr)
	}

	// Verify basic schema structure is correct
	assert.Contains(t, contentStr, "schema ")

	// Check for at least some property definitions
	assert.Contains(t, contentStr, "?: ")

	// Test KCL validation if available
	if isKCLAvailable() {
		// Run KCL against all files in the directory
		cmd := exec.Command("kcl", ".")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("KCL validation failed with error: %v", err)
			t.Logf("KCL output: %s", string(output))
			t.Fail()
		} else {
			t.Logf("KCL validation succeeded")
		}
	} else {
		t.Logf("KCL not available, skipping validation")
	}
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
