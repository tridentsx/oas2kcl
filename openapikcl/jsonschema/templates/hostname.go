package templates

// HostnameTemplate returns a template for hostname validation
func HostnameTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "hostname",
		Description:    "Hostname string",
		ValidationCode: `self.{property} == None or regex.match("^[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?)*$", self.{property}), "{property} must be a valid hostname"`,
		Comments: []string{
			"# Format: hostname",
			"# Hostname string following RFC 1123",
		},
		SchemaContent: `schema Hostname:
    """Hostname string validation.
    
    Validates strings to ensure they conform to hostname format according to RFC 1123.
    """
    value: str
    
    check:
        value == None or regex.match("^[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?)*$", value), "must be a valid hostname"
`,
	}
}
