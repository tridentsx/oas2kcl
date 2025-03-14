package jsonschema

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ValidationTestCase represents a complete test case for validation
type ValidationTestCase struct {
	Name           string
	SchemaPath     string
	ValidInput     string
	InvalidInputs  map[string]string
	OutputDir      string
	SchemaName     string
	GeneratedFiles []string
}

// TestSchemaValidation runs validation tests for all test cases in the testdata folder
func TestSchemaValidation(t *testing.T) {
	// Skip if KCL is not installed
	if _, err := exec.LookPath("kcl"); err != nil {
		t.Skip("KCL not installed, skipping validation tests")
	}

	// Base testdata directory
	testdataDir := "testdata/validation"

	// Discover all test case directories
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testCaseName := entry.Name()

		// We now support pattern properties fully
		if testCaseName == "pattern_properties" {
			t.Logf("Running pattern properties test with full support")
		}

		t.Run(testCaseName, func(t *testing.T) {
			// Set up test case
			testCase := setupTestCase(t, testdataDir, testCaseName)

			// Run the test case
			runValidationTest(t, testCase)
		})
	}
}

// setupTestCase prepares a test case from a directory
func setupTestCase(t *testing.T, baseDir, testCaseName string) ValidationTestCase {
	testDir := filepath.Join(baseDir, testCaseName)
	inputDir := filepath.Join(testDir, "input")
	outputDir := filepath.Join(testDir, "output")

	// Create or clean output directory
	os.RemoveAll(outputDir)
	err := os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Failed to create output directory")

	// Find schema file
	schemaPath := filepath.Join(inputDir, "schema.json")
	_, err = os.Stat(schemaPath)
	require.NoError(t, err, "Schema file not found")

	// Find valid input
	validPath := filepath.Join(inputDir, "valid.json")
	_, err = os.Stat(validPath)
	require.NoError(t, err, "Valid input file not found")

	// Find invalid inputs
	invalidInputs := make(map[string]string)
	entries, err := os.ReadDir(inputDir)
	require.NoError(t, err, "Failed to read input directory")

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "invalid_") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			invalidInputs[name] = filepath.Join(inputDir, entry.Name())
		}
	}

	// Read schema to get schema name from title
	schemaBytes, err := os.ReadFile(schemaPath)
	require.NoError(t, err, "Failed to read schema file")

	var schema map[string]interface{}
	err = json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err, "Failed to parse schema JSON")

	schemaName := "Schema"
	if title, ok := schema["title"].(string); ok && title != "" {
		schemaName = title
	}

	return ValidationTestCase{
		Name:           testCaseName,
		SchemaPath:     schemaPath,
		ValidInput:     validPath,
		InvalidInputs:  invalidInputs,
		OutputDir:      outputDir,
		SchemaName:     schemaName,
		GeneratedFiles: []string{},
	}
}

// runValidationTest runs a validation test case
func runValidationTest(t *testing.T, testCase ValidationTestCase) {
	// Generate KCL schema
	schemaBytes := readFile(t, testCase.SchemaPath)

	// Create a generator
	var schema map[string]interface{}
	err := json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err, "Failed to parse schema JSON")

	generator := NewSchemaGenerator(schema, testCase.OutputDir)
	_, err = generator.GenerateKCLSchemas()
	require.NoError(t, err, "Failed to generate KCL schema")

	// Generate main.k file
	mainContent := fmt.Sprintf("package test\n\nimport %s\n", testCase.SchemaName)
	mainFilePath := filepath.Join(testCase.OutputDir, "main.k")
	err = os.WriteFile(mainFilePath, []byte(mainContent), 0644)
	require.NoError(t, err, "Failed to write main.k file")

	// Verify schema is generated
	schemaFilePath := filepath.Join(testCase.OutputDir, testCase.SchemaName+".k")
	_, err = os.Stat(schemaFilePath)
	require.NoError(t, err, "Schema file not generated: %s", schemaFilePath)

	// Verify main.k is generated
	_, err = os.Stat(mainFilePath)
	require.NoError(t, err, "Main file not generated")

	// Store generated files for cleanup
	testCase.GeneratedFiles = append(testCase.GeneratedFiles, schemaFilePath, mainFilePath)

	// Validate the valid input (should pass)
	valid, output, err := validateWithKCL(testCase.ValidInput, testCase.OutputDir, testCase.SchemaName)
	require.NoError(t, err, "Valid input failed validation: %s\nSchema: %s\nInput: %s", output, schemaFilePath, testCase.ValidInput)
	require.True(t, valid, "Valid input was incorrectly marked invalid")

	// Validate invalid inputs (should fail)
	for name, invalidInputPath := range testCase.InvalidInputs {
		valid, output, err := validateWithKCL(invalidInputPath, testCase.OutputDir, testCase.SchemaName)
		require.Error(t, err, "Invalid input '%s' unexpectedly passed validation", name)
		require.False(t, valid, "Invalid input '%s' was incorrectly marked valid", name)
		t.Logf("Invalid input '%s' correctly failed validation: %s", name, output)
	}
}

// readFile reads a file and returns its contents
func readFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read file: %s", path)
	return data
}

// validateWithKCL uses kcl vet to validate input against the schema
func validateWithKCL(inputPath, schemaPath, schemaName string) (bool, string, error) {
	mainKContent := fmt.Sprintf(`
import file
import json
import %s

data: %s = json.decode(file.read("%s"))
`, schemaName, schemaName, inputPath)

	tempMainPath := filepath.Join(schemaPath, "temp_main.k")
	err := os.WriteFile(tempMainPath, []byte(mainKContent), 0644)
	if err != nil {
		return false, "", fmt.Errorf("failed to write temporary main.k file: %w", err)
	}
	defer os.Remove(tempMainPath)

	cmd := exec.Command("kcl", "run", tempMainPath, schemaPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return false, string(output), fmt.Errorf("validation failed: %w", err)
	}

	return true, string(output), nil
}
