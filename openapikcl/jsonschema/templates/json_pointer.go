package templates

// JSONPointerTemplate returns a template for JSON Pointer validation
func JSONPointerTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "json-pointer",
		Description:    "JSON Pointer (RFC 6901)",
		ValidationCode: `self.{property} == None or regex.match("^(|/[^/~]*(~[01][^/~]*)*)$", self.{property}), "{property} must be a valid JSON Pointer"`,
		Comments: []string{
			"# Format: json-pointer",
			"# JSON Pointer (RFC 6901)",
		},
		SchemaContent: `schema JSONPointer:
    """JSON Pointer validation (RFC 6901).
    
    Validates strings to ensure they conform to JSON Pointer format.
    """
    value: str
    
    check:
        value == None or regex.match("^(|/[^/~]*(~[01][^/~]*)*)$", value), "must be a valid JSON Pointer"
`,
	}
}

// RelativeJSONPointerTemplate returns a template for Relative JSON Pointer validation
func RelativeJSONPointerTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "relative-json-pointer",
		Description:    "Relative JSON Pointer",
		ValidationCode: `self.{property} == None or regex.match("^(0|[1-9][0-9]*)(#|(/(([^/~])|~[01])*)?)$", self.{property}), "{property} must be a valid Relative JSON Pointer"`,
		Comments: []string{
			"# Format: relative-json-pointer",
			"# Relative JSON Pointer",
		},
		SchemaContent: `schema RelativeJSONPointer:
    """Relative JSON Pointer validation.
    
    Validates strings to ensure they conform to Relative JSON Pointer format.
    """
    value: str
    
    check:
        value == None or regex.match("^(0|[1-9][0-9]*)(#|(/(([^/~])|~[01])*)?)$", value), "must be a valid Relative JSON Pointer"
`,
	}
}

// RegexTemplate returns a template for regex validation
func RegexTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "regex",
		Description:    "Regular expression",
		ValidationCode: `self.{property} == None or (try regex.match(self.{property}, ""), True catch e, False), "{property} must be a valid regular expression"`,
		Comments: []string{
			"# Format: regex",
			"# Regular expression pattern",
		},
		SchemaContent: `schema Regex:
    """Regular expression validation.
    
    Validates strings to ensure they are valid regular expressions.
    """
    value: str
    
    check:
        # Try to use the string as a regex pattern
        # If it throws an error, it's not a valid regex
        value == None or (try regex.match(value, ""), True catch e, False), "must be a valid regular expression"
`,
	}
}
