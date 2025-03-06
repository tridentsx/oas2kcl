package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
)

// DetectAPIType checks if a file is OpenAPI 2.0, 3.0, 3.1, or JSON Schema
func DetectAPIType(filename string) (string, error) {
	// Read file content
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Try to parse as JSON
	var jsonObj map[string]interface{}
	if json.Unmarshal(data, &jsonObj) != nil {
		// If JSON parsing fails, try YAML
		if yaml.Unmarshal(data, &jsonObj) != nil {
			return "", fmt.Errorf("file is neither valid JSON nor YAML")
		}
	}

	// Detect OpenAPI vs JSON Schema
	if val, ok := jsonObj["swagger"]; ok {
		if str, valid := val.(string); valid && str == "2.0" {
			return "OpenAPI 2.0", nil
		}
	}

	if val, ok := jsonObj["openapi"]; ok {
		if str, valid := val.(string); valid && strings.HasPrefix(str, "3.0") {
			return "OpenAPI 3.0", nil
		}
		if str, valid := val.(string); valid && strings.HasPrefix(str, "3.1") {
			return "OpenAPI 3.1", nil
		}
	}

	// Detect JSON Schema based on `$schema`
	if val, ok := jsonObj["$schema"]; ok {
		if schemaURL, valid := val.(string); valid {
			switch {
			case strings.Contains(schemaURL, "draft-04"):
				return "JSON Schema Draft 4", nil
			case strings.Contains(schemaURL, "draft-06"):
				return "JSON Schema Draft 6", nil
			case strings.Contains(schemaURL, "draft-07"):
				return "JSON Schema Draft 7", nil
			case strings.Contains(schemaURL, "2019-09"):
				return "JSON Schema 2019-09", nil
			case strings.Contains(schemaURL, "2020-12"):
				return "JSON Schema 2020-12", nil
			default:
				return "Unknown JSON Schema version", nil
			}
		}
	}

	return "Unknown format", nil
}
