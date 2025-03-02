package openapikcl

// ConvertTypeToKCL maps OpenAPI types to KCL types
func ConvertTypeToKCL(oapiType, format string) string {
	switch oapiType {
	case "string":
		if format == "date-time" {
			return "datetime"
		}
		return "str"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	case "number":
		return "float"
	case "array":
		return "list"
	case "object":
		return "dict"
	default:
		return "any"
	}
}

