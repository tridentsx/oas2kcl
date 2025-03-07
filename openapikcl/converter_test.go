// converter_test.go
package openapikcl

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestConvertTypeToKCL(t *testing.T) {
	tests := []struct {
		name     string
		oapiType string
		format   string
		expected string
	}{
		{"String", "string", "", "str"},
		{"DateTime", "string", "date-time", "str"},
		{"Date", "string", "date", "str"},
		{"Email", "string", "email", "str"},
		{"UUID", "string", "uuid", "str"},
		{"URI", "string", "uri", "str"},
		{"Integer", "integer", "", "int"},
		{"Int32", "integer", "int32", "int"},
		{"Int64", "integer", "int64", "int"},
		{"Boolean", "boolean", "", "bool"},
		{"Number", "number", "", "float"},
		{"Float", "number", "float", "float"},
		{"Double", "number", "double", "float"},
		{"Array", "array", "", "list"},
		{"Object", "object", "", "{str:any}"},
		{"Unknown", "unknown", "", "any"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertTypeToKCL(tc.oapiType, tc.format)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGenerateConstraints(t *testing.T) {
	// Helper function to create uint64 pointer
	uintPtr := func(val uint64) *uint64 {
		return &val
	}

	// Helper function to create float64 pointer
	floatPtr := func(val float64) *float64 {
		return &val
	}

	tests := []struct {
		name         string
		schema       *openapi3.Schema
		fieldName    string
		expectedCons []string
	}{
		{
			name: "String Constraints",
			schema: &openapi3.Schema{
				MinLength: 3,
				MaxLength: uintPtr(50),
				Pattern:   "^[a-z]+$",
			},
			fieldName: "username",
			expectedCons: []string{
				"len(username) >= 3",
				"len(username) <= 50",
				"regex.match(username, r\"^[a-z]+$\")",
			},
		},
		{
			name: "Numeric Constraints",
			schema: &openapi3.Schema{
				Min:          floatPtr(5),
				Max:          floatPtr(100),
				ExclusiveMin: true,
				ExclusiveMax: false,
				MultipleOf:   floatPtr(5),
			},
			fieldName: "count",
			expectedCons: []string{
				"count > 5",
				"count <= 100",
				"count % 5 == 0",
			},
		},
		{
			name: "Array Constraints",
			schema: &openapi3.Schema{
				MinItems:    5,
				MaxItems:    uintPtr(20),
				UniqueItems: true,
			},
			fieldName: "tags",
			expectedCons: []string{
				"len(tags) >= 5",
				"len(tags) <= 20",
				"isunique(tags)",
			},
		},
		{
			name: "Enum Constraints",
			schema: &openapi3.Schema{
				Enum: []interface{}{"red", "green", "blue"},
			},
			fieldName: "color",
			expectedCons: []string{
				"color in [\"red\", \"green\", \"blue\"]",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			constraints := GenerateConstraints(tc.schema, tc.fieldName, false)

			// Check if all expected constraints are present
			for _, expected := range tc.expectedCons {
				assert.Contains(t, constraints, expected, "Missing expected constraint")
			}

			// Check if we don't have extra constraints
			assert.Equal(t, len(tc.expectedCons), len(constraints),
				"Number of constraints doesn't match expected count")
		})
	}
}

func TestFormatDocumentation(t *testing.T) {
	tests := []struct {
		name     string
		schema   *openapi3.Schema
		expected string
	}{
		{
			name: "Complete Documentation",
			schema: &openapi3.Schema{
				Title:       "User",
				Description: "A user of the system",
				Default:     "guest",
				Deprecated:  true,
				ReadOnly:    true,
			},
			expected: "# User\n# A user of the system\n# DEPRECATED\n# ReadOnly: This field is read-only\n",
		},
		{
			name: "Multiline Description",
			schema: &openapi3.Schema{
				Description: "Line 1\nLine 2\nLine 3",
			},
			expected: "# Line 1\n# Line 2\n# Line 3\n",
		},
		{
			name:     "Empty Schema",
			schema:   &openapi3.Schema{},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatDocumentation(tc.schema)
			assert.Equal(t, tc.expected, result)
		})
	}
}
