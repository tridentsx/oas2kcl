package templates

// IDNEmailTemplate returns a template for IDN email validation
func IDNEmailTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "idn-email",
		Description:    "Internationalized email address",
		ValidationCode: `self.{property} == None or regex.match("^[\\p{L}\\p{N}\\._%+-]+@[\\p{L}\\p{N}\\.-]+\\.[\\p{L}]{2,}$", self.{property}), "{property} must be a valid IDN email address"`,
		Comments: []string{
			"# Format: idn-email",
			"# Internationalized email address",
		},
		SchemaContent: `schema IDNEmail:
    """Internationalized email address validation.
    
    Validates strings to ensure they conform to internationalized email address format.
    """
    value: str
    
    check:
        # Basic pattern for internationalized email addresses
        # Full validation of IDN would require more complex processing
        value == None or regex.match("^[\\p{L}\\p{N}\\._%+-]+@[\\p{L}\\p{N}\\.-]+\\.[\\p{L}]{2,}$", value), "must be a valid IDN email address"
`,
	}
}

// IDNHostnameTemplate returns a template for IDN hostname validation
func IDNHostnameTemplate() TypeTemplate {
	return TypeTemplate{
		TypeName:       "str",
		FormatName:     "idn-hostname",
		Description:    "Internationalized hostname",
		ValidationCode: `self.{property} == None or regex.match("^[\\p{L}\\p{N}]([\\p{L}\\p{N}\\-]{0,61}[\\p{L}\\p{N}])?(\\.[\\p{L}\\p{N}]([\\p{L}\\p{N}\\-]{0,61}[\\p{L}\\p{N}])?)*$", self.{property}), "{property} must be a valid IDN hostname"`,
		Comments: []string{
			"# Format: idn-hostname",
			"# Internationalized hostname",
		},
		SchemaContent: `schema IDNHostname:
    """Internationalized hostname validation.
    
    Validates strings to ensure they conform to internationalized hostname format.
    """
    value: str
    
    check:
        # Basic pattern for internationalized hostnames
        # Full validation of IDN would require more complex processing
        value == None or regex.match("^[\\p{L}\\p{N}]([\\p{L}\\p{N}\\-]{0,61}[\\p{L}\\p{N}])?(\\.[\\p{L}\\p{N}]([\\p{L}\\p{N}\\-]{0,61}[\\p{L}\\p{N}])?)*$", value), "must be a valid IDN hostname"
`,
	}
}
