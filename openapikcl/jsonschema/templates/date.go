package templates

// DateTemplate returns a template for date validation
func DateTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "date",
		Description:    "Date string in ISO 8601 format (YYYY-MM-DD)",
		ValidationCode: `self.{property} == None or regex.match("^\\d{4}-\\d{2}-\\d{2}$", self.{property}), "{property} must be a valid date in YYYY-MM-DD format"`,
		Comments: []string{
			"# Format: date",
			"# Date string in ISO 8601 format (YYYY-MM-DD)",
		},
		SchemaContent: `schema Date:
    """Date string validation.
    
    Validates strings to ensure they conform to ISO 8601 date format (YYYY-MM-DD).
    """
    value: str
    
    check:
        value == None or regex.match("^\\d{4}-\\d{2}-\\d{2}$", value), "must be a valid date in YYYY-MM-DD format"
`,
	}
}

// DateTimeTemplate returns a template for date-time validation
func DateTimeTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "date-time",
		Description:    "Date-time string in ISO 8601 format",
		ValidationCode: `self.{property} == None or regex.match("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?(Z|[+-]\\d{2}:\\d{2})$", self.{property}), "{property} must be a valid ISO 8601 date-time"`,
		Comments: []string{
			"# Format: date-time",
			"# Date-time string in ISO 8601 format",
		},
		SchemaContent: `schema DateTime:
    """Date-time string validation.
    
    Validates strings to ensure they conform to ISO 8601 date-time format.
    """
    value: str
    
    check:
        value == None or regex.match("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?(Z|[+-]\\d{2}:\\d{2})$", value), "must be a valid ISO 8601 date-time"
`,
	}
}
