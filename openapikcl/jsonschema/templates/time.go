package templates

// TimeTemplate returns a template for time validation
func TimeTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "time",
		Description:    "Time string in ISO 8601 format (HH:MM:SS)",
		ValidationCode: `self.{property} == None or regex.match("^\\d{2}:\\d{2}:\\d{2}$", self.{property}), "{property} must be a valid time in HH:MM:SS format"`,
		Comments: []string{
			"# Format: time",
			"# Time string in ISO 8601 format (HH:MM:SS)",
		},
		SchemaContent: `schema Time:
    """Time string validation.
    
    Validates strings to ensure they conform to ISO 8601 time format (HH:MM:SS).
    """
    value: str
    
    check:
        value == None or regex.match("^\\d{2}:\\d{2}:\\d{2}$", value), "must be a valid time in HH:MM:SS format"
`,
	}
}

// DurationTemplate returns a template for duration validation
func DurationTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "duration",
		Description:    "Duration string in ISO 8601 format (e.g., P1DT2H3M4S)",
		ValidationCode: `self.{property} == None or regex.match("^P(?:\\d+(?:\\.\\d+)?Y)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?W)?(?:\\d+(?:\\.\\d+)?D)?(?:T(?:\\d+(?:\\.\\d+)?H)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?S)?)?$", self.{property}), "{property} must be a valid ISO 8601 duration"`,
		Comments: []string{
			"# Format: duration",
			"# Duration string in ISO 8601 format (e.g., P1DT2H3M4S)",
		},
		SchemaContent: `schema Duration:
    """Duration string validation.
    
    Validates strings to ensure they conform to ISO 8601 duration format.
    """
    value: str
    
    check:
        value == None or regex.match("^P(?:\\d+(?:\\.\\d+)?Y)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?W)?(?:\\d+(?:\\.\\d+)?D)?(?:T(?:\\d+(?:\\.\\d+)?H)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?S)?)?$", value), "must be a valid ISO 8601 duration"
`,
	}
}
