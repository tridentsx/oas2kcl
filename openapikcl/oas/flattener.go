// flattener.go
package oas

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// FlattenOptions defines options for flattening an OpenAPI schema
type FlattenOptions struct {
	BaseDir    string // base directory for resolving local file references
	MaxDepth   int    // maximum depth for flattening references
	SkipRemote bool   // skip remote references
}

// refContext is used to track context during reference resolution
type refContext struct {
	refPath        string
	currentSchemas map[string]bool
}

// Flattener is used to flatten an OpenAPI schema by resolving references
type Flattener struct {
	options      FlattenOptions
	doc          *openapi3.T
	usedRefs     map[string]bool
	err          error
	tempDirs     []string
	knownSchemas map[string]bool
}

// NewFlattener creates a new Flattener
func NewFlattener(opts FlattenOptions, doc *openapi3.T) *Flattener {
	if doc == nil {
		return nil
	}

	// Set default options
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 10 // Default to 10 levels of nesting
	}

	return &Flattener{
		options:      opts,
		doc:          doc,
		usedRefs:     make(map[string]bool),
		knownSchemas: make(map[string]bool),
	}
}

// FlattenSpec flattens the OpenAPI specification
func (f *Flattener) FlattenSpec() (*openapi3.T, error) {
	if f.doc == nil {
		return nil, errors.New("no OpenAPI document provided")
	}

	log.Print("flattening OpenAPI specification")

	// Process all schema references in components
	if f.doc.Components != nil && f.doc.Components.Schemas != nil {
		for schemaName, schemaRef := range f.doc.Components.Schemas {
			f.knownSchemas[schemaName] = true
			log.Printf("flattening schema: %s", schemaName)
			flattenedSchema, err := f.flattenSchemaRef(schemaRef, &refContext{
				refPath:        "#/components/schemas/" + schemaName,
				currentSchemas: make(map[string]bool),
			}, 0)
			if err != nil {
				return nil, fmt.Errorf("error flattening schema %s: %w", schemaName, err)
			}
			f.doc.Components.Schemas[schemaName] = flattenedSchema
		}
	}

	// Process all response references in components
	if f.doc.Components != nil && f.doc.Components.Responses != nil {
		// Similar logic for responses
	}

	// Process all parameter references in components
	if f.doc.Components != nil && f.doc.Components.Parameters != nil {
		// Similar logic for parameters
	}

	// Process all requestBody references in components
	if f.doc.Components != nil && f.doc.Components.RequestBodies != nil {
		// Similar logic for request bodies
	}

	// Process paths, operations, and their schemas
	if f.doc.Paths != nil {
		paths := f.doc.Paths.Map()
		for pathName, pathItem := range paths {
			log.Printf("flattening path: %s", pathName)
			for _, operation := range []*openapi3.Operation{
				pathItem.Connect, pathItem.Delete, pathItem.Get,
				pathItem.Head, pathItem.Options, pathItem.Patch,
				pathItem.Post, pathItem.Put, pathItem.Trace,
			} {
				if operation == nil {
					continue
				}

				// Flatten request body schemas
				if operation.RequestBody != nil && operation.RequestBody.Value != nil {
					for contentType, mediaType := range operation.RequestBody.Value.Content {
						if mediaType.Schema != nil {
							flattenedSchema, err := f.flattenSchemaRef(mediaType.Schema, &refContext{
								refPath:        "#/paths/" + pathName + "/requestBody/content/" + contentType + "/schema",
								currentSchemas: make(map[string]bool),
							}, 0)
							if err != nil {
								return nil, fmt.Errorf("error flattening request body schema: %w", err)
							}
							mediaType.Schema = flattenedSchema
						}
					}
				}

				// Flatten response schemas
				if operation.Responses != nil {
					responses := operation.Responses.Map()
					for statusCode, response := range responses {
						if response.Value == nil {
							continue
						}
						for contentType, mediaType := range response.Value.Content {
							if mediaType.Schema != nil {
								flattenedSchema, err := f.flattenSchemaRef(mediaType.Schema, &refContext{
									refPath:        "#/paths/" + pathName + "/responses/" + statusCode + "/content/" + contentType + "/schema",
									currentSchemas: make(map[string]bool),
								}, 0)
								if err != nil {
									return nil, fmt.Errorf("error flattening response schema: %w", err)
								}
								mediaType.Schema = flattenedSchema
							}
						}
					}
				}
			}
		}
	}

	return f.doc, nil
}

// flattenSchemaRef flattens a schema reference
func (f *Flattener) flattenSchemaRef(ref *openapi3.SchemaRef, ctx *refContext, depth int) (*openapi3.SchemaRef, error) {
	if ref == nil {
		return nil, nil
	}

	// Check if we've exceeded the maximum depth
	if depth > f.options.MaxDepth {
		return ref, fmt.Errorf("maximum reference depth exceeded (current: %d, max: %d)", depth, f.options.MaxDepth)
	}

	// If this is not a reference, process its sub-schemas
	if ref.Ref == "" {
		// Process properties
		if ref.Value != nil && ref.Value.Properties != nil {
			for propName, propSchema := range ref.Value.Properties {
				flattenedProp, err := f.flattenSchemaRef(propSchema, &refContext{
					refPath:        ctx.refPath + "/properties/" + propName,
					currentSchemas: copyStringBoolMap(ctx.currentSchemas),
				}, depth+1)
				if err != nil {
					return nil, err
				}
				ref.Value.Properties[propName] = flattenedProp
			}
		}

		// Process items for arrays
		if ref.Value != nil && ref.Value.Items != nil {
			flattenedItems, err := f.flattenSchemaRef(ref.Value.Items, &refContext{
				refPath:        ctx.refPath + "/items",
				currentSchemas: copyStringBoolMap(ctx.currentSchemas),
			}, depth+1)
			if err != nil {
				return nil, err
			}
			ref.Value.Items = flattenedItems
		}

		// Process allOf, oneOf, anyOf
		if ref.Value != nil && len(ref.Value.AllOf) > 0 {
			for i, schema := range ref.Value.AllOf {
				flattenedSchema, err := f.flattenSchemaRef(schema, &refContext{
					refPath:        ctx.refPath + fmt.Sprintf("/allOf/%d", i),
					currentSchemas: copyStringBoolMap(ctx.currentSchemas),
				}, depth+1)
				if err != nil {
					return nil, err
				}
				ref.Value.AllOf[i] = flattenedSchema
			}
		}

		if ref.Value != nil && len(ref.Value.OneOf) > 0 {
			for i, schema := range ref.Value.OneOf {
				flattenedSchema, err := f.flattenSchemaRef(schema, &refContext{
					refPath:        ctx.refPath + fmt.Sprintf("/oneOf/%d", i),
					currentSchemas: copyStringBoolMap(ctx.currentSchemas),
				}, depth+1)
				if err != nil {
					return nil, err
				}
				ref.Value.OneOf[i] = flattenedSchema
			}
		}

		if ref.Value != nil && len(ref.Value.AnyOf) > 0 {
			for i, schema := range ref.Value.AnyOf {
				flattenedSchema, err := f.flattenSchemaRef(schema, &refContext{
					refPath:        ctx.refPath + fmt.Sprintf("/anyOf/%d", i),
					currentSchemas: copyStringBoolMap(ctx.currentSchemas),
				}, depth+1)
				if err != nil {
					return nil, err
				}
				ref.Value.AnyOf[i] = flattenedSchema
			}
		}

		return ref, nil
	}

	// This is a reference, so we need to resolve it
	refURL := ref.Ref

	// Skip external references if configured to do so
	if f.options.SkipRemote && isRemoteRef(refURL) {
		return ref, nil
	}

	// Check for circular references
	if ctx.currentSchemas[refURL] {
		log.Printf("circular reference detected: %s", refURL)
		return ref, nil
	}

	// Mark this reference as being processed
	newCtx := &refContext{
		refPath:        refURL,
		currentSchemas: copyStringBoolMap(ctx.currentSchemas),
	}
	newCtx.currentSchemas[refURL] = true

	// Extract the schema name from the reference
	schemaName := getSchemaNameFromRef(refURL)
	if schemaName == "" {
		return ref, fmt.Errorf("invalid reference: %s", refURL)
	}

	// If this is an internal reference to a schema in components, and we've already seen it
	if strings.HasPrefix(refURL, "#/components/schemas/") && f.knownSchemas[schemaName] {
		// We don't need to flatten it again, just return the reference
		return ref, nil
	}

	// For simplicity, we'll only handle internal references in this implementation
	if isInternalRef(refURL) {
		// Try to find the schema in the components
		if f.doc.Components != nil && f.doc.Components.Schemas != nil {
			if schema, ok := f.doc.Components.Schemas[schemaName]; ok {
				// We found the schema, now flatten it
				flattenedSchema, err := f.flattenSchemaRef(schema, newCtx, depth+1)
				if err != nil {
					return nil, err
				}

				// Update the schema in components
				f.doc.Components.Schemas[schemaName] = flattenedSchema
				f.knownSchemas[schemaName] = true

				// Return a reference to the flattened schema
				return &openapi3.SchemaRef{Ref: refURL}, nil
			}
		}
	}

	// If we get here, it means we couldn't resolve the reference
	log.Printf("warning: unresolved reference: %s", refURL)
	return ref, nil
}

// Close cleans up any temporary resources
func (f *Flattener) Close() error {
	for _, dir := range f.tempDirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("warning: failed to remove temporary directory %s: %v", dir, err)
		}
	}
	f.tempDirs = nil
	return nil
}

// Helper functions

// isRemoteRef returns true if the reference is to a remote resource
func isRemoteRef(ref string) bool {
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://")
}

// isInternalRef returns true if the reference is internal to the document
func isInternalRef(ref string) bool {
	return strings.HasPrefix(ref, "#/")
}

// getSchemaNameFromRef extracts the schema name from a reference
func getSchemaNameFromRef(ref string) string {
	if strings.HasPrefix(ref, "#/components/schemas/") {
		return strings.TrimPrefix(ref, "#/components/schemas/")
	}
	if strings.HasPrefix(ref, "#/definitions/") {
		return strings.TrimPrefix(ref, "#/definitions/")
	}
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// copyStringBoolMap creates a copy of a map[string]bool
func copyStringBoolMap(m map[string]bool) map[string]bool {
	newMap := make(map[string]bool, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

// FlattenDocument is a utility function to flatten an OpenAPI document
func FlattenDocument(doc *openapi3.T, opts FlattenOptions) (*openapi3.T, error) {
	flattener := NewFlattener(opts, doc)
	defer flattener.Close()
	return flattener.FlattenSpec()
}
