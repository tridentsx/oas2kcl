// Package utils provides utility functions for JSON Schema to KCL conversion.
package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// GetStringValue safely extracts a string value from a map
func GetStringValue(data map[string]interface{}, key string) (string, bool) {
	if value, ok := data[key]; ok {
		if strValue, isString := value.(string); isString {
			return strValue, true
		}
	}
	return "", false
}

// GetBoolValue safely extracts a boolean value from a map
func GetBoolValue(data map[string]interface{}, key string) (bool, bool) {
	if value, ok := data[key]; ok {
		if boolValue, isBool := value.(bool); isBool {
			return boolValue, true
		}
	}
	return false, false
}

// GetIntValue safely extracts an integer value from a map
func GetIntValue(data map[string]interface{}, key string) (int64, bool) {
	if value, ok := data[key]; ok {
		switch t := value.(type) {
		case int:
			return int64(t), true
		case int64:
			return t, true
		case float64:
			return int64(t), true
		}
	}
	return 0, false
}

// GetFloatValue safely extracts a float value from a map
func GetFloatValue(data map[string]interface{}, key string) (float64, bool) {
	if value, ok := data[key]; ok {
		switch t := value.(type) {
		case float64:
			return t, true
		case int:
			return float64(t), true
		case int64:
			return float64(t), true
		}
	}
	return 0, false
}

// GetMapValue safely extracts a map value from a map
func GetMapValue(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	if value, ok := data[key]; ok {
		if mapValue, isMap := value.(map[string]interface{}); isMap {
			return mapValue, true
		}
	}
	return nil, false
}

// GetArrayValue safely extracts an array value from a map
func GetArrayValue(data map[string]interface{}, key string) ([]interface{}, bool) {
	if value, ok := data[key]; ok {
		if arrayValue, isArray := value.([]interface{}); isArray {
			return arrayValue, true
		}
	}
	return nil, false
}

// FormatLiteral formats a Go value as a KCL literal
func FormatLiteral(value interface{}) string {
	if value == nil {
		return "None"
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case bool:
		if v {
			return "True"
		}
		return "False"
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		items := make([]string, 0, len(v))
		for _, item := range v {
			items = append(items, FormatLiteral(item))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	case map[string]interface{}:
		props := make([]string, 0, len(v))
		for key, val := range v {
			props = append(props, fmt.Sprintf(`"%s": %s`, key, FormatLiteral(val)))
		}
		return fmt.Sprintf("{%s}", strings.Join(props, ", "))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// SanitizePropertyName ensures the property name is valid in KCL
func SanitizePropertyName(name string) string {
	if name == "" {
		return "property"
	}

	// Replace special characters and spaces with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	sanitized := re.ReplaceAllString(name, "_")

	// Ensure it doesn't start with a number
	if match, _ := regexp.MatchString(`^[0-9]`, sanitized); match {
		sanitized = "_" + sanitized
	}

	// Handle reserved keywords in KCL
	reserved := map[string]string{
		"import": "import_",
		"schema": "schema_",
		"type":   "type_",
		"mixin":  "mixin_",
		"check":  "check_",
		"assert": "assert_",
		"if":     "if_",
		"elif":   "elif_",
		"else":   "else_",
		"for":    "for_",
		"in":     "in_",
	}

	if replacement, isReserved := reserved[sanitized]; isReserved {
		return replacement
	}

	return sanitized
}

// GenerateKCLFilePath generates a valid file path for a KCL schema
func GenerateKCLFilePath(outputDir string, schemaName string) string {
	fileName := schemaName + ".k"
	return filepath.Join(outputDir, fileName)
}
