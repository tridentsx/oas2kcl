// Package jsonschema implements the conversion of JSON Schema to KCL.
package jsonschema

import (
	"fmt"
	"strings"
)

// stringFormatRegexes defines regular expressions for validating string formats
// These patterns are used to validate string values when KCL doesn't have native format support
var stringFormatRegexes = map[string]string{
	// Date and time formats
	"date-time": `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?)(Z|[\+-]\d{2}:\d{2})?$`,
	"date":      `^\d{4}-\d{2}-\d{2}$`,
	"time":      `^\d{2}:\d{2}:\d{2}(?:\.\d+)?$`,
	"duration":  `^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`,

	// Email formats
	"email":     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
	"idn-email": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, // Simplified pattern

	// Hostname formats
	"hostname":     `^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`,
	"idn-hostname": `^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`, // Simplified

	// IP address formats
	"ipv4": `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
	"ipv6": `^(?:(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(?:ffff(?::0{1,4}){0,1}:){0,1}(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])|(?:[0-9a-fA-F]{1,4}:){1,4}:(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`,

	// URI formats
	"uri":           `^[a-zA-Z][a-zA-Z0-9+.-]*:[^\s]*$`,
	"uri-reference": `^(?:[a-zA-Z][a-zA-Z0-9+.-]*:)?(?://[^\s/$.?#].[^\s]*|[^\s/$.?#].[^\s]*)$`,
	"iri":           `^[a-zA-Z][a-zA-Z0-9+.-]*:[^\s]*$`,                                         // Simplified
	"iri-reference": `^(?:[a-zA-Z][a-zA-Z0-9+.-]*:)?(?://[^\s/$.?#].[^\s]*|[^\s/$.?#].[^\s]*)$`, // Simplified

	// Other formats
	"uuid":                  `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	"json-pointer":          `^(?:/(?:[^~/]|~0|~1)*)*$`,
	"relative-json-pointer": `^(?:0|[1-9][0-9]*)(?:/(?:[^~/]|~0|~1)*)*$`,
}

// GetFormatRegex returns a regular expression for validating a specific string format
// It returns an empty string if the format is not supported
func GetFormatRegex(format string) string {
	regex, ok := stringFormatRegexes[format]
	if ok {
		return regex
	}
	return ""
}

// FormatNeedsRegexValidation checks if a format requires regex validation in KCL
func FormatNeedsRegexValidation(format string) bool {
	_, ok := stringFormatRegexes[format]
	return ok
}

// GetFormatConstraint returns a KCL constraint expression for validating a string format
// For example, for "email" format, it returns: regex.match(self, r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$")
func GetFormatConstraint(fieldName, format string) string {
	regex := GetFormatRegex(format)
	if regex == "" {
		return ""
	}

	// Escape backslashes for string literal
	regex = strings.ReplaceAll(regex, `\`, `\\`)

	// Use 'self' for validation directly on the schema
	// or use the field name for validation within an object
	target := "self"
	if fieldName != "" {
		target = "self." + fieldName
	}

	return fmt.Sprintf(`regex.match(%s, r"%s")`, target, regex)
}
