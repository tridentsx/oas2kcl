package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"

	"github.com/tridentsx/oas2kcl/openapikcl"
)

func main() {
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.SetPrefix("openapi-to-kcl: ")

	// Define command-line flags
	oasFile := flag.String("oas", "", "Path to the OpenAPI specification file")
	jsonFile := flag.String("json", "", "Path to the JSON Schema file")
	outFile := flag.String("out", "", "Optional output file for the generated KCL schema (.k)")
	skipFlatten := flag.Bool("skip-flatten", false, "Skip flattening the OpenAPI spec")
	skipRemote := flag.Bool("skip-remote", false, "Skip remote references during flattening")
	maxDepth := flag.Int("max-depth", 100, "Maximum depth for reference resolution")
	packageName := flag.String("package", "schema", "Package name for the generated KCL schema")
	flag.Parse()

	// Ensure only one of -oas or -json is provided
	if (*oasFile != "" && *jsonFile != "") || (*oasFile == "" && *jsonFile == "") {
		log.Fatal("Either -oas or -json must be provided, but not both. Usage:\n  openapi-to-kcl -oas openapi.json [-out schema.k]\n  OR\n  openapi-to-kcl -json schema.json [-out schema.k]")
	}

	// Processing based on the selected mode
	if *oasFile != "" {
		processOpenAPI(*oasFile, *outFile, *skipFlatten, *skipRemote, *maxDepth, *packageName)
	} else {
		processJSONSchema(*jsonFile, *outFile, *packageName)
	}
}

// processOpenAPI handles OpenAPI file conversion
func processOpenAPI(oasFile, outFile string, skipFlatten, skipRemote bool, maxDepth int, packageName string) {
	log.Printf("Loading OpenAPI schema from %s", oasFile)
	doc, version, err := openapikcl.LoadOpenAPISchema(oasFile, openapikcl.LoadOptions{
		FlattenSpec: !skipFlatten,
		SkipRemote:  skipRemote,
		MaxDepth:    maxDepth,
	})
	if err != nil {
		log.Fatalf("Failed to load OpenAPI schema: %v", err)
	}

	log.Printf("Detected OpenAPI version: %s", version)
	log.Print("OpenAPI schema validation successful")

	// Generate the KCL schema
	log.Print("Generating KCL schemas")
	err = openapikcl.GenerateKCLSchemas(doc, outFile, packageName, version)
	if err != nil {
		log.Fatalf("Failed to generate KCL schema: %v", err)
	}

	if outFile != "" {
		log.Printf("KCL schema written to %s", outFile)
	} else {
		log.Print("KCL schema generation complete")
	}
}

// processJSONSchema handles JSON Schema conversion
func processJSONSchema(jsonFile, outFile, packageName string) {
	log.Printf("Processing JSON Schema from %s", jsonFile)

	// Read the schema file
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON Schema file: %v", err)
	}

	// Try JSON first, then fallback to YAML
	var schemaData map[string]interface{}
	if json.Unmarshal(data, &schemaData) != nil {
		if yaml.Unmarshal(data, &schemaData) != nil {
			log.Fatal("Invalid JSON Schema: not a valid JSON or YAML file")
		}
	}

	// Validate schema using jsonschema package
	log.Print("Validating JSON Schema")
	compiler := jsonschema.NewCompiler()

	// Detect JSON Schema version from $schema field
	if val, ok := schemaData["$schema"]; ok {
		if schemaURL, valid := val.(string); valid {
			log.Printf("Detected JSON Schema version: %s", schemaURL)
		} else {
			log.Print("Warning: Could not determine JSON Schema version")
		}
	} else {
		log.Print("Warning: No $schema field detected; assuming latest draft")
	}

	// Compile schema
	err = compiler.AddResource(jsonFile, strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Failed to load JSON Schema: %v", err)
	}
	schema, err := compiler.Compile(jsonFile)
	if err != nil {
		log.Fatalf("JSON Schema validation failed: %v", err)
	}

	// Log validation success
	log.Print("JSON Schema validation successful")

	// TODO: Convert validated JSON Schema to KCL
	log.Print("Converting JSON Schema to KCL (not yet implemented)")

	if outFile != "" {
		log.Printf("Expected to write output to %s", outFile)
	} else {
		log.Print("JSON Schema processing completed")
	}
}
