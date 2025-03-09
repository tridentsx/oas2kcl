package templates

// URITemplate returns a template for URI validation
func URITemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "uri",
		Description:    "URI string following RFC 3986",
		ValidationCode: `self.{property} == None or regex.match("^(https?|ftp)://[^\\s/$.?#].[^\\s]*$", self.{property}), "{property} must be a valid URI"`,
		Comments: []string{
			"# Format: uri",
			"# URI string following RFC 3986",
		},
		SchemaContent: `schema URI:
    """URI string validation.
    
    Validates strings to ensure they conform to URI format according to RFC 3986.
    """
    value: str
    
    check:
        value == None or regex.match("^(https?|ftp)://[^\\s/$.?#].[^\\s]*$", value), "must be a valid URI"
`,
	}
}
