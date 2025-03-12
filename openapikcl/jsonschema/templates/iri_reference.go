package templates

// IRIReferenceTemplate returns a template for IRI reference validation
func IRIReferenceTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "iri-reference",
		Description:    "IRI reference string",
		ValidationCode: `self.{property} == None or regex.match(self.{property}, "^(?:[a-zA-Z][a-zA-Z0-9+.-]*:)?(?://[^\\s/$.?#].[^\\s]*|[^\\s/$.?#].[^\\s]*)$"), "{property} must be a valid IRI reference"`,
		Comments: []string{
			"# Format: iri-reference",
			"# IRI reference string (can be relative)",
		},
		SchemaContent: `schema IRIReference:
    """IRI reference string validation.
    
    Validates strings to ensure they conform to IRI reference format.
    Similar to URI reference but allows for internationalized characters.
    """
    value: str
    
    check:
        value == None or regex.match(value, "^(?:[a-zA-Z][a-zA-Z0-9+.-]*:)?(?://[^\\s/$.?#].[^\\s]*|[^\\s/$.?#].[^\\s]*)$"), "must be a valid IRI reference"
`,
	}
}
