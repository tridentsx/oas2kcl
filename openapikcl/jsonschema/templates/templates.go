// Package templates provides reusable KCL schema templates for common format types.
package templates

import (
	"strings"
)

// TypeTemplate defines a template for a format type
type TypeTemplate struct {
	// Base KCL type (usually "str")
	TypeName string

	// The schema file content for this type
	SchemaContent string

	// Description of the type
	Description string

	// Validation code for the check block
	ValidationCode string

	// Comments to include with the property
	Comments []string

	// Format name (email, uri, etc.)
	FormatName string
}

// Registry of templates
var formatTemplates = map[string]func() TypeTemplate{
	// Date and time formats
	"date":      DateTemplate,
	"date-time": DateTimeTemplate,
	"time":      TimeTemplate,
	"duration":  DurationTemplate,

	// Email and hostname formats
	"email":        EmailTemplate,
	"idn-email":    IDNEmailTemplate,
	"hostname":     HostnameTemplate,
	"idn-hostname": IDNHostnameTemplate,

	// IP address formats
	"ipv4": IPv4Template,
	"ipv6": IPv6Template,

	// URI and IRI formats
	"uri":           URITemplate,
	"uri-reference": URIReferenceTemplate,
	"iri":           IRITemplate,
	"iri-reference": IRIReferenceTemplate,

	// JSON pointer formats
	"json-pointer":          JSONPointerTemplate,
	"relative-json-pointer": RelativeJSONPointerTemplate,

	// UUID format
	"uuid": UUIDTemplate,

	// Regex format
	"regex": RegexTemplate,
}

// GetTemplateForFormat returns a template for the given format
func GetTemplateForFormat(format string) *TypeTemplate {
	if templateFunc, ok := formatTemplates[format]; ok {
		template := templateFunc()
		return &template
	}
	return nil
}

// GetTemplateForNumberType returns a template for a number or integer type with constraints
func GetTemplateForNumberType(propSchema map[string]interface{}, schemaType string) *TypeTemplate {
	var template TypeTemplate

	if schemaType == "integer" {
		template = IntegerTemplate(propSchema)
	} else if schemaType == "number" {
		template = NumberTemplate(propSchema)
	} else {
		return nil
	}

	return &template
}

// GetTemplateForArrayType returns a template for an array type with constraints
func GetTemplateForArrayType(propSchema map[string]interface{}) *TypeTemplate {
	template := ArrayTemplate(propSchema)
	return &template
}

// GetComments returns formatted comments for the property
func (t *TypeTemplate) GetComments(indentation string) string {
	if len(t.Comments) == 0 {
		return ""
	}

	indentedComments := make([]string, len(t.Comments))
	for i, comment := range t.Comments {
		indentedComments[i] = indentation + comment
	}

	return strings.Join(indentedComments, "\n")
}

// GetValidation returns the formatted validation code for the property
func (t *TypeTemplate) GetValidation(propertyName string, indentation string) string {
	if t.ValidationCode == "" {
		return ""
	}

	// Replace {property} placeholder with actual property name
	validation := strings.Replace(t.ValidationCode, "{property}", propertyName, -1)

	// Indent each line
	lines := strings.Split(validation, "\n")
	for i, line := range lines {
		lines[i] = indentation + line
	}

	return strings.Join(lines, "\n")
}

// GetSchemaContent returns the content for the schema file
func (t *TypeTemplate) GetSchemaContent() string {
	return t.SchemaContent
}

// NeedsSeparateSchema returns true if this type should be in a separate schema file
func (t *TypeTemplate) NeedsSeparateSchema() bool {
	return t.SchemaContent != ""
}
