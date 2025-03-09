package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tridentsx/oas2kcl/openapikcl"
)

func main() {
	// Enable verbose logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Starting test...")

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "kcl-test-")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Created temp dir: %s\n", tempDir)

	// Create a simple JSON Schema file
	schemaContent := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"name": {
				"type": "string"
			},
			"age": {
				"type": "integer",
				"minimum": 0
			}
		},
		"required": ["name"]
	}`
	schemaFile := filepath.Join(tempDir, "schema.json")
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		log.Fatalf("Failed to write schema file: %v", err)
	}

	fmt.Printf("Created schema file: %s\n", schemaFile)

	// Generate KCL
	outputDir := filepath.Join(tempDir, "output")
	fmt.Printf("Output directory: %s\n", outputDir)

	fmt.Println("Generating KCL schemas...")
	if err := openapikcl.GenerateKCL(schemaFile, outputDir, "test", false); err != nil {
		log.Fatalf("Failed to generate KCL: %v", err)
	}

	// List the output directory
	fmt.Println("Listing output directory:")
	files, err := os.ReadDir(outputDir)
	if err != nil {
		log.Fatalf("Failed to read output directory: %v", err)
	}

	for _, file := range files {
		fmt.Printf("  - %s\n", file.Name())
	}

	// Check if main.k was created
	mainKPath := filepath.Join(outputDir, "main.k")
	if _, err := os.Stat(mainKPath); err != nil {
		log.Fatalf("main.k was not created: %v", err)
	}

	fmt.Println("Test passed! KCL schemas were generated successfully.")
}
