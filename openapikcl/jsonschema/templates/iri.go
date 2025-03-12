package templates

// IRITemplate returns a template for IRI validation
func IRITemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "iri",
		Description:    "IRI string",
		ValidationCode: `self.{property} == None or regex.match(self.{property}, "^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]*$"), "{property} must be a valid IRI"`,
		Comments: []string{
			"# Format: iri",
			"# IRI string (Internationalized URI)",
		},
		SchemaContent: `schema IRI:
    """IRI string validation.
    
    Validates strings to ensure they conform to IRI format.
    An IRI is similar to a URI but allows for non-ASCII characters.
    """
    value: str
    
    check:
        value == None or regex.match(value, "^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]*$"), "must be a valid IRI"
`,
	}
}
