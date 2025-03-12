package templates

// URITemplate returns a template for URI validation
func URITemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "uri",
		Description:    "URI string",
		ValidationCode: `self.{property} == None or regex.match(self.{property}, "^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]*$"), "{property} must be a valid URI"`,
		Comments: []string{
			"# Format: uri",
			"# URI string according to RFC 3986",
		},
		SchemaContent: `schema URI:
    """URI string validation.
    
    Validates strings to ensure they conform to URI format according to RFC 3986.
    Requires a scheme (like http:, https:, ftp:) followed by a hierarchical or 
    opaque part.
    """
    value: str
    
    check:
        value == None or regex.match(value, "^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]*$"), "must be a valid URI"
`,
	}
}
