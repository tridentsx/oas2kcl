package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"

	"github.com/tridentsx/oas2kcl/openapikcl"
)

func main() {
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.SetPrefix("openapi-to-kcl: ")

	// Define command-line flags
	schemaFile := flag.String("schema", "", "Path to the schema file (OpenAPI or JSON Schema)")
	outDir := flag.String("out", "", "Output directory for the generated KCL schemas")
	skipFlatten := flag.Bool("skip-flatten", false, "Skip flattening the OpenAPI spec")
	skipRemote := flag.Bool("skip-remote", false, "Skip remote references during flattening")
	maxDepth := flag.Int("max-depth", 100, "Maximum depth for reference resolution")
	packageName := flag.String("package", "schema", "Package name for the generated KCL schemas")
	flag.Parse()

	// Ensure a schema file is provided
	if *schemaFile == "" {
		log.Fatal("Missing required -schema flag. Usage:\n  openapi-to-kcl -schema schema.json -out output_dir")
	}

	// Process the schema file
	processSchema(*schemaFile, *outDir, *skipFlatten, *skipRemote, *maxDepth, *packageName)
}

// processSchema handles schema file conversion (either OpenAPI or JSON Schema)
func processSchema(schemaFile, outDir string, skipFlatten, skipRemote bool, maxDepth int, packageName string) {
	log.Printf("Processing schema from %s", schemaFile)

	// Read the schema file
	data, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	// Try to parse as JSON first, then fallback to YAML
	var rawSchema map[string]interface{}
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		// Try YAML if JSON parsing failed
		if err := yaml.Unmarshal(data, &rawSchema); err != nil {
			log.Fatalf("Failed to parse schema file: not a valid JSON or YAML: %v", err)
		}
	}

	// First, try to load as an OpenAPI schema
	var doc *openapi3.T
	var version openapikcl.OpenAPIVersion

	// Attempt to load as OpenAPI
	doc, version, err = openapikcl.LoadOpenAPISchema(schemaFile, openapikcl.LoadOptions{
		FlattenSpec: !skipFlatten,
		SkipRemote:  skipRemote,
		MaxDepth:    maxDepth,
	})

	// Create output directory if it doesn't exist
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	// Generate KCL schemas based on detected schema type
	err = openapikcl.GenerateKCLSchemas(doc, outDir, packageName, version, rawSchema)
	if err != nil {
		log.Fatalf("Failed to generate KCL schemas: %v", err)
	}

	log.Printf("Successfully generated KCL schemas in %s", outDir)
}
