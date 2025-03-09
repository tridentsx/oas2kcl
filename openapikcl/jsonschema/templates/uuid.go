package templates

// UUIDTemplate returns a template for UUID validation
func UUIDTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "uuid",
		Description:    "UUID string",
		ValidationCode: `self.{property} == None or regex.match("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", self.{property}), "{property} must be a valid UUID"`,
		Comments: []string{
			"# Format: uuid",
			"# UUID string representation",
		},
		SchemaContent: `schema UUID:
    """UUID string validation.
    
    Validates strings to ensure they conform to UUID format.
    """
    value: str
    
    check:
        value == None or regex.match("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", value), "must be a valid UUID"
`,
	}
}
