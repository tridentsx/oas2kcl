package openapikcl

import (
	"path/filepath"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// Helper functions for reference resolution
func isLocalRef(ref string) bool {
	return ref[0] == '#'
}

func isFileRef(ref string) bool {
	return filepath.Ext(ref) != ""
}

func isURLRef(ref string) bool {
	return len(ref) > 7 && (ref[:7] == "http://" || ref[:8] == "https://")
}

// collectSchemas collects and returns a sorted list of schema names
func collectSchemas(schemas openapi3.Schemas) []string {
	keys := make([]string, 0, len(schemas))
	for k := range schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// mergeSchemas merges multiple schemas into a single schema
func mergeSchemas(schemas []*openapi3.SchemaRef) (*openapi3.SchemaRef, error) {
	if len(schemas) == 0 {
		return nil, nil
	}

	merged := &openapi3.Schema{}
	for _, schema := range schemas {
		if schema.Value == nil {
			continue
		}

		// Merge properties
		if merged.Properties == nil {
			merged.Properties = make(openapi3.Schemas)
		}
		for name, prop := range schema.Value.Properties {
			merged.Properties[name] = prop
		}

		// Merge required fields
		merged.Required = append(merged.Required, schema.Value.Required...)
	}

	return &openapi3.SchemaRef{Value: merged}, nil
}
