package templates

// EmailTemplate returns a template for email validation
func EmailTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "email",
		Description:    "Email address string",
		ValidationCode: `self.{property} == None or regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", self.{property}), "{property} must be a valid email address"`,
		Comments: []string{
			"# Format: email",
			"# Email address string",
		},
		SchemaContent: `schema Email:
    """Email address string validation.
    
    Validates strings to ensure they conform to email address format.
    """
    value: str
    
    check:
        value == None or regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", value), "must be a valid email address"
`,
	}
}
