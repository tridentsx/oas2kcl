package templates

// UUIDTemplate returns a template for UUID validation
func UUIDTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "uuid",
		Description:    "UUID string in standard format",
		ValidationCode: `self.{property} == None or regex.match(self.{property}, "^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"), "{property} must be a valid UUID"`,
		Comments: []string{
			"# Format: uuid",
			"# UUID string in standard format",
		},
		SchemaContent: `schema UUID:
    """UUID string validation.
    
    Validates strings to ensure they conform to UUID format.
    Pattern: xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx where M is 1-5 and N is 8,9,a,b
    """
    value: str
    
    check:
        value == None or regex.match(value, "^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"), "must be a valid UUID"
`,
	}
}
