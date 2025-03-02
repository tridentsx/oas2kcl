package main

import (
	"flag"
	"fmt"
	"os"

    	"github.com/tridentsx/oas2kcl/openapikcl"
)

unc main() {
	// Define command-line flags
	oasFile := flag.String("oas", "", "Path to the OpenAPI specification file (required)")
	outFile := flag.String("out", "", "Optional output file for the generated KCL schema (.k)")
	flag.Parse()

	// Ensure the required flag is provided
	if *oasFile == "" {
		fmt.Println("Error: The -oas flag is required. Usage:")
		fmt.Println("  openapi-to-kcl -oas openapi.json [-out schema.k]")
		os.Exit(1)
	}

	// Load and validate the OpenAPI schema
	doc, err := openapikcl.LoadOpenAPISchema(*oasFile)
	if err != nil {
		fmt.Println("Error loading OpenAPI schema:", err)
		os.Exit(1)
	}

	fmt.Println("OpenAPI schema is valid!")

	// Generate the KCL schema
	kclOutput := openapikcl.GenerateKCLSchemas(doc)

	// Handle output: either write to a file or print to stdout
	if *outFile != "" {
		err := os.WriteFile(*outFile, []byte(kclOutput), 0644)
		if err != nil {
			fmt.Println("Error writing KCL schema to file:", err)
			os.Exit(1)
		}
		fmt.Println("KCL schema successfully written to", *outFile)
	} else {
		fmt.Println(kclOutput)
	}
}
t
