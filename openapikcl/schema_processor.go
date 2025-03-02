package openapikcl

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// flattenSchemaRef flattens a schema reference and its nested references
func (f *Flattener) flattenSchemaRef(ref *openapi3.SchemaRef) (*openapi3.SchemaRef, error) {
	if ref == nil {
		return nil, nil
	}

	// Handle direct reference
	if ref.Ref != "" {
		return f.resolveReference(ref.Ref)
	}

	// Create new schema to avoid modifying the original
	schema := *ref.Value
	flatSchema := &openapi3.SchemaRef{
		Value: &schema,
	}

	// Flatten properties
	if schema.Properties != nil {
		flatProps, err := f.flattenProperties(flatSchema.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten properties: %w", err)
		}
		flatSchema.Value.Properties = flatProps
	}

	// Flatten array items
	if schema.Items != nil {
		flatItems, err := f.flattenSchemaRef(schema.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten array items: %w", err)
		}
		flatSchema.Value.Items = flatItems
	}

	// Flatten allOf, oneOf, anyOf
	if err := f.flattenCompositeSchemas(flatSchema.Value); err != nil {
		return nil, err
	}

	// Process required properties
	if len(schema.Required) > 0 {
		// Make a copy of the required array to avoid modifying the original
		flatSchema.Value.Required = make([]string, len(schema.Required))
		copy(flatSchema.Value.Required, schema.Required)

		// Validate that all required properties exist
		for _, reqProp := range schema.Required {
			if _, ok := flatSchema.Value.Properties[reqProp]; !ok {
				return nil, fmt.Errorf("required property %q not found in schema", reqProp)
			}
		}
	}

	// Handle additional schema features
	if schema.AdditionalProperties.Has != nil {
		if schema.AdditionalProperties.Schema != nil {
			flatAdditionalProps, err := f.flattenSchemaRef(schema.AdditionalProperties.Schema)
			if err != nil {
				return nil, fmt.Errorf("failed to flatten additionalProperties: %w", err)
			}

			// Create a new AdditionalProperties with the flattened schema
			flatSchema.Value.AdditionalProperties = openapi3.AdditionalProperties{
				Has:    schema.AdditionalProperties.Has,
				Schema: flatAdditionalProps,
			}
		} else {
			// Just copy it as is if there's no schema
			flatSchema.Value.AdditionalProperties = schema.AdditionalProperties
		}
	}

	// Copy schema metadata
	copySchemaMetadata(&schema, flatSchema.Value)

	// Handle discriminator
	if schema.Discriminator != nil {
		// Validate discriminator property exists
		if schema.Properties == nil || schema.Properties[schema.Discriminator.PropertyName] == nil {
			return nil, fmt.Errorf("discriminator property %q not found in schema properties",
				schema.Discriminator.PropertyName)
		}

		// Copy discriminator
		flatSchema.Value.Discriminator = &openapi3.Discriminator{
			PropertyName: schema.Discriminator.PropertyName,
			Mapping:      make(map[string]string),
		}

		// Validate and copy mapping references
		for key, ref := range schema.Discriminator.Mapping {
			if _, err := f.resolveReference(ref); err != nil {
				return nil, fmt.Errorf("invalid discriminator mapping %q -> %q: %w", key, ref, err)
			}
			flatSchema.Value.Discriminator.Mapping[key] = ref
		}
	}

	return flatSchema, nil
}

// flattenProperties flattens the properties of a schema
func (f *Flattener) flattenProperties(schema *openapi3.Schema) (openapi3.Schemas, error) {
	flatProps := make(openapi3.Schemas)

	for name, prop := range schema.Properties {
		flatProp, err := f.flattenSchemaRef(prop)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten property %s: %w", name, err)
		}
		flatProps[name] = flatProp
	}

	return flatProps, nil
}

// flattenCompositeSchemas flattens allOf, oneOf, and anyOf schemas
func (f *Flattener) flattenCompositeSchemas(schema *openapi3.Schema) error {
	// Handle allOf
	if len(schema.AllOf) > 0 {
		// Flatten each schema in allOf
		flatAllOf := make([]*openapi3.SchemaRef, 0, len(schema.AllOf))
		for _, s := range schema.AllOf {
			flat, err := f.flattenSchemaRef(s)
			if err != nil {
				return fmt.Errorf("failed to flatten allOf schema: %w", err)
			}
			flatAllOf = append(flatAllOf, flat)
		}

		// Merge the flattened schemas
		merged, err := mergeSchemas(flatAllOf)
		if err != nil {
			return fmt.Errorf("failed to merge allOf schemas: %w", err)
		}

		// Copy merged properties to the parent schema
		if merged.Value.Properties != nil {
			if schema.Properties == nil {
				schema.Properties = make(openapi3.Schemas)
			}
			for name, prop := range merged.Value.Properties {
				schema.Properties[name] = prop
			}
		}

		// Merge required fields
		schema.Required = append(schema.Required, merged.Value.Required...)

		// Clear the allOf array since we've merged its contents
		schema.AllOf = nil
	}

	// Handle oneOf
	if len(schema.OneOf) > 0 {
		flatOneOf := make([]*openapi3.SchemaRef, 0, len(schema.OneOf))
		for _, s := range schema.OneOf {
			flat, err := f.flattenSchemaRef(s)
			if err != nil {
				return fmt.Errorf("failed to flatten oneOf schema: %w", err)
			}
			flatOneOf = append(flatOneOf, flat)
		}
		schema.OneOf = flatOneOf
	}

	// Handle anyOf
	if len(schema.AnyOf) > 0 {
		flatAnyOf := make([]*openapi3.SchemaRef, 0, len(schema.AnyOf))
		for _, s := range schema.AnyOf {
			flat, err := f.flattenSchemaRef(s)
			if err != nil {
				return fmt.Errorf("failed to flatten anyOf schema: %w", err)
			}
			flatAnyOf = append(flatAnyOf, flat)
		}
		schema.AnyOf = flatAnyOf
	}

	return nil
}

// copySchemaMetadata copies metadata from source schema to target schema
func copySchemaMetadata(source *openapi3.Schema, target *openapi3.Schema) {
	// Copy basic metadata
	target.Nullable = source.Nullable
	target.ReadOnly = source.ReadOnly
	target.WriteOnly = source.WriteOnly
	target.Deprecated = source.Deprecated
	target.Title = source.Title
	target.Description = source.Description
	target.Default = source.Default
	target.Example = source.Example

	// Copy validation constraints
	target.MinLength = source.MinLength
	target.MaxLength = source.MaxLength
	target.Pattern = source.Pattern
	target.Min = source.Min
	target.Max = source.Max
	target.ExclusiveMin = source.ExclusiveMin
	target.ExclusiveMax = source.ExclusiveMax
	target.MultipleOf = source.MultipleOf
	target.MinItems = source.MinItems
	target.MaxItems = source.MaxItems
	target.UniqueItems = source.UniqueItems
	target.Format = source.Format
}
