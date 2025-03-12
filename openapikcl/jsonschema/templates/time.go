package templates

// TimeTemplate returns a template for time validation
func TimeTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "time",
		Description:    "Time string in HH:MM:SS format",
		ValidationCode: `self.{property} == None or datetime.validate(self.{property}, "%H:%M:%S"), "{property} must be a valid time in HH:MM:SS format"`,
		Comments: []string{
			"# Format: time",
			"# Time string in HH:MM:SS format",
		},
		SchemaContent: `schema Time:
    """Time string validation.
    
    Validates strings to ensure they conform to the time format (HH:MM:SS).
    """
    value: str
    
    check:
        value == None or datetime.validate(value, "%H:%M:%S"), "must be a valid time in HH:MM:SS format"
`,
	}
}

// DurationTemplate returns a template for duration validation
func DurationTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "duration",
		Description:    "Duration string in ISO 8601 format (e.g., P1DT2H3M4S)",
		ValidationCode: `self.{property} == None or regex.match(self.{property}, "^P(?:\\d+(?:\\.\\d+)?Y)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?W)?(?:\\d+(?:\\.\\d+)?D)?(?:T(?:\\d+(?:\\.\\d+)?H)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?S)?)?$"), "{property} must be a valid ISO 8601 duration"`,
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
        value == None or regex.match(value, "^P(?:\\d+(?:\\.\\d+)?Y)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?W)?(?:\\d+(?:\\.\\d+)?D)?(?:T(?:\\d+(?:\\.\\d+)?H)?(?:\\d+(?:\\.\\d+)?M)?(?:\\d+(?:\\.\\d+)?S)?)?$"), "must be a valid ISO 8601 duration"
`,
	}
}
