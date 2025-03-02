package openapikcl

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// resolveReference resolves an external reference
func (f *Flattener) resolveReference(ref string) (*openapi3.SchemaRef, error) {
	// Check for circular references with path
	if f.seenRefs[ref] {
		path := strings.Join(append(f.refPath, ref), " -> ")
		return nil, fmt.Errorf("circular reference detected: %s", path)
	}
	f.seenRefs[ref] = true
	f.refPath = append(f.refPath, ref)
	defer func() {
		delete(f.seenRefs, ref)
		f.refPath = f.refPath[:len(f.refPath)-1]
	}()

	// Check cache first
	if cached, ok := f.cache[ref]; ok {
		log.Printf("using cached reference for %s from %s", ref, cached.source)
		return cached.schema, nil
	}

	// Handle depth limit
	f.depth++
	if f.depth > f.opts.MaxDepth {
		return nil, fmt.Errorf("maximum reference depth exceeded: %d", f.opts.MaxDepth)
	}
	defer func() { f.depth-- }()

	log.Printf("resolving reference: %s", ref)

	// Handle different reference types
	switch {
	case isLocalRef(ref):
		return f.resolveLocalRef(ref)
	case isFileRef(ref):
		return f.resolveFileRef(ref)
	case isURLRef(ref):
		if f.opts.SkipRemote {
			log.Printf("skipping remote reference: %s", ref)
			return nil, nil
		}
		return f.resolveURLRef(ref)
	default:
		return nil, fmt.Errorf("unsupported reference format: %s", ref)
	}
}

// resolveLocalRef resolves references within the same document (#/components/...)
func (f *Flattener) resolveLocalRef(ref string) (*openapi3.SchemaRef, error) {
	if !strings.HasPrefix(ref, "#") {
		return nil, fmt.Errorf("invalid local reference format: %s", ref)
	}

	// Remove the '#' prefix
	path := strings.TrimPrefix(ref, "#")

	// Split the path into components
	components := strings.Split(path, "/")

	// Remove empty components
	var parts []string
	for _, part := range components {
		if part != "" {
			parts = append(parts, part)
		}
	}

	// Navigate through the document structure
	var current interface{} = f.doc
	for i, part := range parts {
		switch v := current.(type) {
		case *openapi3.T:
			switch part {
			case "components":
				current = v.Components
			default:
				return nil, fmt.Errorf("unsupported path component at position %d: %s", i, part)
			}
		case *openapi3.Components:
			switch part {
			case "schemas":
				current = v.Schemas
			default:
				return nil, fmt.Errorf("unsupported components section at position %d: %s", i, part)
			}
		case openapi3.Schemas:
			schema, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("schema not found: %s", part)
			}
			return schema, nil
		default:
			return nil, fmt.Errorf("unexpected type at position %d: %T", i, current)
		}
	}

	return nil, fmt.Errorf("invalid reference path: %s", ref)
}

// resolveFileRef resolves references to other files
func (f *Flattener) resolveFileRef(ref string) (*openapi3.SchemaRef, error) {
	// Split the reference into file path and fragment
	parts := strings.SplitN(ref, "#", 2)
	filePath := filepath.Join(f.opts.BaseDir, parts[0])

	log.Printf("resolving file reference: %s", filePath)

	// Load the referenced file
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load referenced file %s: %w", filePath, err)
	}

	// If there's a fragment, resolve it
	if len(parts) > 1 {
		fragment := "#" + parts[1]
		// Create a temporary flattener for the loaded document
		tempFlattener := NewFlattener(f.opts, doc)
		return tempFlattener.resolveLocalRef(fragment)
	}

	return nil, fmt.Errorf("file reference must include a fragment identifier: %s", ref)
}

// resolveURLRef resolves references to remote URLs
func (f *Flattener) resolveURLRef(ref string) (*openapi3.SchemaRef, error) {
	// Split the reference into URL and fragment
	parts := strings.SplitN(ref, "#", 2)
	url := parts[0]

	log.Printf("resolving URL reference: %s", url)

	// Make HTTP request
	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote reference %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Load the referenced document
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote reference %s: %w", url, err)
	}

	// If there's a fragment, resolve it
	if len(parts) > 1 {
		fragment := "#" + parts[1]
		// Create a temporary flattener for the loaded document
		tempFlattener := NewFlattener(f.opts, doc)
		return tempFlattener.resolveLocalRef(fragment)
	}

	return nil, fmt.Errorf("URL reference must include a fragment identifier: %s", ref)
}

// cacheReferences logs information about the cached references
func (f *Flattener) cacheReferences() {
	for ref, context := range f.cache {
		log.Printf("cached reference: %s from %s", ref, context.source)
	}
}
