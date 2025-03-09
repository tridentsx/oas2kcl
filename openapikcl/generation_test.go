package openapikcl_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tridentsx/oas2kcl/openapikcl"
)

// Helper function to create types for testing
func createTypes(typeValue string) *openapi3.Types {
	t := openapi3.Types([]string{typeValue})
	return &t
}

// TestIntegratedGenerateKCL tests basic KCL generation for a simple schema
func TestIntegratedGenerateKCL(t *testing.T) {
	// Create a temporary directory for output
	tempDir, err := os.MkdirTemp("", "oas2kcl-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Simple JSON Schema
	jsonSchema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"name": {
				"type": "string"
			},
			"age": {
				"type": "integer"
			}
		}
	}`

	// Write the schema to a temporary file
	schemaPath := filepath.Join(tempDir, "schema.json")
	err = os.WriteFile(schemaPath, []byte(jsonSchema), 0644)
	require.NoError(t, err)

	// Generate KCL
	err = openapikcl.GenerateKCL(schemaPath, tempDir, "test", false)
	require.NoError(t, err)

	// Check if the output file exists
	assert.FileExists(t, filepath.Join(tempDir, "main.k"))
	assert.FileExists(t, filepath.Join(tempDir, "schema.k"))

	// Read the generated schema file
	schemaContent, err := os.ReadFile(filepath.Join(tempDir, "schema.k"))
	require.NoError(t, err)
	assert.Contains(t, string(schemaContent), "schema = schema {")
	assert.Contains(t, string(schemaContent), "name?: str")
	assert.Contains(t, string(schemaContent), "age?: int")
}

// TestIntegratedGenerateKCLFromOpenAPI tests KCL generation for OpenAPI schemas
func TestIntegratedGenerateKCLFromOpenAPI(t *testing.T) {
	// Create a temporary directory for output
	tempDir, err := os.MkdirTemp("", "oas2kcl-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Simple OpenAPI 3.0 document
	openAPI3 := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {},
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"id": {
							"type": "integer",
							"format": "int64"
						},
						"name": {
							"type": "string"
						},
						"email": {
							"type": "string",
							"format": "email"
						}
					}
				}
			}
		}
	}`

	// Write the schema to a temporary file
	schemaPath := filepath.Join(tempDir, "openapi.json")
	err = os.WriteFile(schemaPath, []byte(openAPI3), 0644)
	require.NoError(t, err)

	// Generate KCL
	err = openapikcl.GenerateKCL(schemaPath, tempDir, "test", false)
	require.NoError(t, err)

	// Check if the output files exist
	assert.FileExists(t, filepath.Join(tempDir, "main.k"))
	assert.FileExists(t, filepath.Join(tempDir, "User.k"))

	// Read the generated schema file
	schemaContent, err := os.ReadFile(filepath.Join(tempDir, "User.k"))
	require.NoError(t, err)
	assert.Contains(t, string(schemaContent), "User = schema {")
	assert.Contains(t, string(schemaContent), "id?: int")
	assert.Contains(t, string(schemaContent), "name?: str")
	assert.Contains(t, string(schemaContent), "email?: str")
}

// TestIntegratedTestCaseOutputs tests generating test case outputs
func TestIntegratedTestCaseOutputs(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping test case output generation in short mode")
	}

	// This test is primarily for development purposes
	// It will generate all test case outputs based on the inputs
	testCasesRoot := "testdata/oas"

	// Test directories can include input and any expected outputs
	testDirs, err := os.ReadDir(testCasesRoot)
	require.NoError(t, err)

	for _, dir := range testDirs {
		if !dir.IsDir() {
			continue
		}

		// Check if this is a test case directory
		testDir := filepath.Join(testCasesRoot, dir.Name())

		// Skip empty directories or directories without input
		inputFiles, err := os.ReadDir(filepath.Join(testDir, "input"))
		if err != nil || len(inputFiles) == 0 {
			continue
		}

		t.Run(dir.Name(), func(t *testing.T) {
			err := openapikcl.GenerateTestCaseOutput(testDir)
			assert.NoError(t, err)
		})
	}
}

// TestIntegratedSingleTestCase tests generating a single test case output
func TestIntegratedSingleTestCase(t *testing.T) {
	// Create a temporary test case directory
	tempDir, err := os.MkdirTemp("", "oas2kcl-testcase-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create input directory
	inputDir := filepath.Join(tempDir, "input")
	err = os.MkdirAll(inputDir, 0755)
	require.NoError(t, err)

	// Create a simple schema file
	schema := openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"Pet": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: createTypes("object"),
						Properties: openapi3.Schemas{
							"name": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: createTypes("string"),
								},
							},
							"breed": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: createTypes("string"),
								},
							},
						},
					},
				},
			},
		},
	}

	// Marshal to JSON
	schemaBytes, err := json.Marshal(schema)
	require.NoError(t, err)

	// Write schema to input directory
	err = os.WriteFile(filepath.Join(inputDir, "simple.json"), schemaBytes, 0644)
	require.NoError(t, err)

	// Generate output
	err = openapikcl.GenerateTestCaseOutput(tempDir)
	require.NoError(t, err)

	// Check if output directory and files exist
	assert.DirExists(t, filepath.Join(tempDir, "output"))
	assert.FileExists(t, filepath.Join(tempDir, "output", "Pet.k"))
	assert.FileExists(t, filepath.Join(tempDir, "output", "main.k"))

	// Read the Pet.k file
	petContent, err := os.ReadFile(filepath.Join(tempDir, "output", "Pet.k"))
	require.NoError(t, err)

	// Verify content
	assert.Contains(t, string(petContent), "Pet = schema {")
	assert.Contains(t, string(petContent), "name?: str")
	assert.Contains(t, string(petContent), "breed?: str")
}
