package openapikcl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONSchemaComposition(t *testing.T) {
	tests := []struct {
		name            string
		jsonSchemaFile  string
		expectedOutputs []string
	}{
		{
			name:           "Test allOf Composition",
			jsonSchemaFile: "testdata/jsonschema/allof.json",
			expectedOutputs: []string{
				"schema Pet:",
				"age?: int",
				"breed?: str",
				"name?: str",
			},
		},
		{
			name:           "Test oneOf Composition",
			jsonSchemaFile: "testdata/jsonschema/oneof.json",
			expectedOutputs: []string{
				"schema Payment:",
				"type_value: str",
				"check:",
				"type_value in [\"credit_card\", \"bank_transfer\"]",
			},
		},
		{
			name:           "Test anyOf Composition",
			jsonSchemaFile: "testdata/jsonschema/anyof.json",
			expectedOutputs: []string{
				"schema Identifier:",
			},
		},
		{
			name:           "Test Nested Composition",
			jsonSchemaFile: "testdata/jsonschema/nested.json",
			expectedOutputs: []string{
				"schema ComplexSchema:",
				"id: str",
				"user?: any",
				"# oneOf validation using discriminator: type",
				"if user.type == individual:",
				"if user.type == organization:",
				"# Conditional validation: if region exists",
				"if user.user has region:",
			},
		},
		{
			name:           "Test If-Then-Else Condition",
			jsonSchemaFile: "testdata/jsonschema/ifthenelse.json",
			expectedOutputs: []string{
				"schema Shipping:",
				"country: str",
				"stateCode?: str",
				"zipCode?: str",
				"check:",
				"if country == \"US\":",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for output
			tempDir, err := os.MkdirTemp("", "kcl-test-")
			assert.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Generate KCL from JSON Schema
			err = GenerateKCL(tt.jsonSchemaFile, tempDir, "")
			assert.NoError(t, err)

			// Find the generated file
			generatedFiles, err := filepath.Glob(filepath.Join(tempDir, "*.k"))
			assert.NoError(t, err)
			assert.True(t, len(generatedFiles) > 0, "No KCL files were generated")

			// Check main schema file
			mainSchema := ""
			for _, file := range generatedFiles {
				if filepath.Base(file) != "main.k" && filepath.Base(file) != "validation_test.k" {
					// Read the file content
					content, err := os.ReadFile(file)
					assert.NoError(t, err)
					mainSchema = string(content)

					// Check that all expected outputs are in the content
					for _, expected := range tt.expectedOutputs {
						assert.Contains(t, mainSchema, expected)
					}
				}
			}

			t.Logf("Generated schema: %s", mainSchema)
		})
	}
}
