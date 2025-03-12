package templates

// HostnameTemplate returns a template for hostname validation
func HostnameTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "hostname",
		Description:    "Hostname string",
		ValidationCode: `self.{property} == None or regex.match(self.{property}, "^[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?)*$"), "{property} must be a valid hostname"`,
		Comments: []string{
			"# Format: hostname",
			"# Hostname string compliant with RFC 1123",
		},
		SchemaContent: `schema Hostname:
    """Hostname string validation.
    
    Validates strings to ensure they conform to hostname format according to RFC 1123.
    Each label must start and end with a letter or digit, contain only letters, digits, or hyphens,
    and be at most 63 characters long.
    """
    value: str
    
    check:
        value == None or regex.match(value, "^[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?)*$"), "must be a valid hostname"
`,
	}
}
