package openapikcl

import (
	"fmt"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// FlattenOptions configures the flattening process
type FlattenOptions struct {
	BaseDir    string // Base directory for relative file references
	MaxDepth   int    // Maximum depth for circular reference detection
	SkipRemote bool   // Skip remote references if true
}

// refContext tracks reference context
type refContext struct {
	schema *openapi3.SchemaRef
	source string // file or URL where the schema was loaded from
}

// Flattener handles the flattening of OpenAPI specs
type Flattener struct {
	opts       FlattenOptions
	seenRefs   map[string]bool
	depth      int
	httpClient *http.Client
	doc        *openapi3.T // Add this field to store the original document
	cache      map[string]refContext
	refPath    []string // Track reference resolution path
}

// NewFlattener creates a new Flattener instance
func NewFlattener(opts FlattenOptions, doc *openapi3.T) *Flattener {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = 100 // reasonable default
	}

	return &Flattener{
		opts:       opts,
		seenRefs:   make(map[string]bool),
		httpClient: &http.Client{},
		doc:        doc,
		cache:      make(map[string]refContext),
	}
}

// FlattenSpec flattens an OpenAPI specification by resolving all references
func (f *Flattener) FlattenSpec() (*openapi3.T, error) {
	log.Println("starting specification flattening process")

	if f.doc == nil {
		return nil, fmt.Errorf("no OpenAPI document provided")
	}

	if f.doc.Components == nil || f.doc.Components.Schemas == nil {
		log.Println("no schemas to flatten")
		return f.doc, nil
	}

	// Create a new document to store the flattened spec
	flatDoc := &openapi3.T{
		OpenAPI: f.doc.OpenAPI,
		Info:    f.doc.Info,
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	// Get a sorted list of schema names for deterministic processing
	schemaNames := collectSchemas(f.doc.Components.Schemas)
	log.Printf("processing schemas in order: %v", schemaNames)

	// Process schemas in sorted order
	for _, name := range schemaNames {
		schema := f.doc.Components.Schemas[name]
		log.Printf("flattening schema: %s", name)

		flatSchema, err := f.flattenSchemaRef(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten schema %s: %w", name, err)
		}

		flatDoc.Components.Schemas[name] = flatSchema
	}

	// Log information about all resolved references
	f.cacheReferences()

	log.Printf("flattened %d schemas successfully", len(flatDoc.Components.Schemas))
	return flatDoc, nil
}

// Close cleans up temporary resources
func (f *Flattener) Close() error {
	f.httpClient.CloseIdleConnections()
	f.cache = nil
	f.seenRefs = nil
	return nil
}
