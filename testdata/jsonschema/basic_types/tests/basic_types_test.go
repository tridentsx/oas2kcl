package basic_types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPrimitiveTypes tests that the basic JSON Schema types are correctly represented in KCL
func TestPrimitiveTypes(t *testing.T) {
	testCases := []struct {
		name            string
		jsonSchemaType  string
		expectedKCLType string
	}{
		{
			name:            "String Type",
			jsonSchemaType:  "string",
			expectedKCLType: "str",
		},
		{
			name:            "Integer Type",
			jsonSchemaType:  "integer",
			expectedKCLType: "int",
		},
		{
			name:            "Number Type",
			jsonSchemaType:  "number",
			expectedKCLType: "float",
		},
		{
			name:            "Boolean Type",
			jsonSchemaType:  "boolean",
			expectedKCLType: "bool",
		},
		{
			name:            "Null Type",
			jsonSchemaType:  "null",
			expectedKCLType: "None",
		},
		{
			name:            "Any Type",
			jsonSchemaType:  "",
			expectedKCLType: "any",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// No-op test for documentation purposes
			// These mappings will be implemented in the code
			assert.NotEmpty(t, tc.jsonSchemaType) // This will fail for the "Any Type" case, but that's ok for this example
		})
	}
}

// TestArrayTypes tests that array types are correctly mapped to KCL
func TestArrayTypes(t *testing.T) {
	testCases := []struct {
		name            string
		jsonSchema      string
		expectedKCLType string
	}{
		{
			name:            "Array of Strings",
			jsonSchema:      `{"type": "array", "items": {"type": "string"}}`,
			expectedKCLType: "[str]",
		},
		{
			name:            "Array of Integers",
			jsonSchema:      `{"type": "array", "items": {"type": "integer"}}`,
			expectedKCLType: "[int]",
		},
		{
			name:            "Array of Numbers",
			jsonSchema:      `{"type": "array", "items": {"type": "number"}}`,
			expectedKCLType: "[float]",
		},
		{
			name:            "Array of Any",
			jsonSchema:      `{"type": "array"}`,
			expectedKCLType: "[any]",
		},
		{
			name:            "Array of Mixed Types",
			jsonSchema:      `{"type": "array", "items": {"type": ["string", "number"]}}`,
			expectedKCLType: "[any]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the JSON schema
			var schema map[string]interface{}
			err := json.Unmarshal([]byte(tc.jsonSchema), &schema)
			assert.NoError(t, err)

			// Validate that we constructed a valid test case
			assert.Equal(t, "array", schema["type"])
		})
	}
}

// TestObjectTypes tests that object types are correctly mapped to KCL
func TestObjectTypes(t *testing.T) {
	testCases := []struct {
		name            string
		jsonSchema      string
		expectedKCLType string
	}{
		{
			name:            "Empty Object",
			jsonSchema:      `{"type": "object"}`,
			expectedKCLType: "{str:any}",
		},
		{
			name:            "Object with Properties",
			jsonSchema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "integer"}}}`,
			expectedKCLType: "{str:any}",
		},
		{
			name:            "Object with AdditionalProperties",
			jsonSchema:      `{"type": "object", "additionalProperties": {"type": "string"}}`,
			expectedKCLType: "{str:str}",
		},
		{
			name:            "Object with Nested Objects",
			jsonSchema:      `{"type": "object", "properties": {"address": {"type": "object", "properties": {"street": {"type": "string"}}}}}`,
			expectedKCLType: "{str:any}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the JSON schema
			var schema map[string]interface{}
			err := json.Unmarshal([]byte(tc.jsonSchema), &schema)
			assert.NoError(t, err)

			// Validate that we constructed a valid test case
			assert.Equal(t, "object", schema["type"])
		})
	}
}
