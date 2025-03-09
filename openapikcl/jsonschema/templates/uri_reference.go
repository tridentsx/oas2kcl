package templates

// URIReferenceTemplate returns a template for URI reference validation
func URIReferenceTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "uri-reference",
		Description:    "URI Reference string",
		ValidationCode: `self.{property} == None or regex.match("^(?:(?:https?|ftp)://)?[^\\s/$.?#].[^\\s]*$", self.{property}), "{property} must be a valid URI reference"`,
		Comments: []string{
			"# Format: uri-reference",
			"# URI Reference string",
		},
		SchemaContent: `schema URIReference:
    """URI Reference string validation.
    
    Validates strings to ensure they conform to URI reference format.
    """
    value: str
    
    check:
        value == None or regex.match("^(?:(?:https?|ftp)://)?[^\\s/$.?#].[^\\s]*$", value), "must be a valid URI reference"
`,
	}
}

// IRITemplate returns a template for IRI validation
func IRITemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "iri",
		Description:    "Internationalized Resource Identifier",
		ValidationCode: `self.{property} == None or regex.match("^[^\\s]*$", self.{property}), "{property} must be a valid IRI"`,
		Comments: []string{
			"# Format: iri",
			"# Internationalized Resource Identifier",
			"# Note: Full IRI validation is complex and this is a simplified check",
		},
		SchemaContent: `schema IRI:
    """Internationalized Resource Identifier validation.
    
    Validates strings to ensure they conform to IRI format.
    Note: Full IRI validation is complex and this is a simplified check.
    """
    value: str
    
    check:
        # Full IRI validation would require more complex processing
        # This is a minimal check just ensuring there are no whitespace characters
        value == None or regex.match("^[^\\s]*$", value), "must be a valid IRI"
`,
	}
}

// IRIReferenceTemplate returns a template for IRI reference validation
func IRIReferenceTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "iri-reference",
		Description:    "Internationalized Resource Identifier Reference",
		ValidationCode: `self.{property} == None or regex.match("^[^\\s]*$", self.{property}), "{property} must be a valid IRI reference"`,
		Comments: []string{
			"# Format: iri-reference",
			"# Internationalized Resource Identifier Reference",
			"# Note: Full IRI reference validation is complex and this is a simplified check",
		},
		SchemaContent: `schema IRIReference:
    """Internationalized Resource Identifier Reference validation.
    
    Validates strings to ensure they conform to IRI reference format.
    Note: Full IRI reference validation is complex and this is a simplified check.
    """
    value: str
    
    check:
        # Full IRI reference validation would require more complex processing
        # This is a minimal check just ensuring there are no whitespace characters
        value == None or regex.match("^[^\\s]*$", value), "must be a valid IRI reference"
`,
	}
}
