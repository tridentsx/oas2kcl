package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tridentsx/oas2kcl/openapikcl"
)

func main() {
	// Parse command-line arguments
	inputFile := flag.String("input", "", "Path to the input schema file (OpenAPI or JSON Schema)")
	outputDir := flag.String("output", "output", "Directory to output the generated KCL schemas")
	packageName := flag.String("package", "schema", "Name of the KCL package")
	generateValidator := flag.Bool("validator", false, "Generate validator schemas")
	flag.Parse()

	// Validate input file
	if *inputFile == "" {
		fmt.Println("Error: input file is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Fatalf("Error: failed to create output directory: %v", err)
		}
	}

	// Generate KCL schemas
	if err := openapikcl.GenerateKCL(*inputFile, *outputDir, *packageName, *generateValidator); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Successfully generated KCL schemas in %s\n", *outputDir)
}
