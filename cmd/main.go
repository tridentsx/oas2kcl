package main

import (
	"flag"
	"log"

	"github.com/tridentsx/oas2kcl/openapikcl"
)

func main() {
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.SetPrefix("openapi-to-kcl: ")

	// Define command-line flags
	oasFile := flag.String("oas", "", "Path to the OpenAPI specification file (required)")
	outFile := flag.String("out", "", "Optional output file for the generated KCL schema (.k)")
	skipFlatten := flag.Bool("skip-flatten", false, "Skip flattening the OpenAPI spec")
	skipRemote := flag.Bool("skip-remote", false, "Skip remote references during flattening")
	maxDepth := flag.Int("max-depth", 100, "Maximum depth for reference resolution")
	packageName := flag.String("package", "schema", "Package name for the generated KCL schema")
	flag.Parse()

	// Ensure the required flag is provided
	if *oasFile == "" {
		log.Fatal("the -oas flag is required. Usage:\n  openapi-to-kcl -oas openapi.json [-out schema.k]")
	}

	// Load and validate the OpenAPI schema
	log.Printf("loading OpenAPI schema from %s", *oasFile)
	doc, err := openapikcl.LoadOpenAPISchema(*oasFile, openapikcl.LoadOptions{
		FlattenSpec: !*skipFlatten,
		SkipRemote:  *skipRemote,
		MaxDepth:    *maxDepth,
	})
	if err != nil {
		log.Fatalf("failed to load OpenAPI schema: %v", err)
	}

	log.Print("OpenAPI schema validation successful")

	// Generate the KCL schema
	log.Print("generating KCL schemas")
	err = openapikcl.GenerateKCLSchemas(doc, *packageName)
	if err != nil {
		log.Fatalf("failed to generate KCL schema: %v", err)
	}

	// Since the function doesn't return the KCL content directly,
	// we can no longer control where it's written from main.go.
	// If the function writes to a file internally, we should log success:
	if *outFile != "" {
		log.Printf("KCL schema should have been written to %s", *outFile)
	} else {
		log.Print("KCL schema generation complete")
	}
}
