package templates

// URIReferenceTemplate returns a template for URI reference validation
func URIReferenceTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "uri-reference",
		Description:    "URI reference string",
		ValidationCode: `self.{property} == None or regex.match(self.{property}, "^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]"), "{property} must be a valid URI reference"`,
		Comments: []string{
			"# Format: uri-reference",
			"# URI reference string",
		},
		SchemaContent: `schema URIReference:
    """URI reference string validation.
    
    Validates strings to ensure they conform to URI reference format.
    A URI Reference may be an absolute or relative URI.
    """
    value: str
    
    check:
        regex.match(value, "^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]"), "value must be a valid URI reference"
`,
	}
}
