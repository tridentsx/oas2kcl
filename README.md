# OAS2KCL - OpenAPI and JSON Schema to KCL Converter

This tool converts OpenAPI specifications and JSON Schema files to KCL (Kubernetes Configuration Language) schemas.

## Features

- Converts OpenAPI 3.0 specifications to KCL schemas
- Converts JSON Schema files to KCL schemas
- Generates validator schemas with constraint validation
- Supports various constraint types:
  - String constraints: minLength, maxLength, pattern, format, enum
  - Number constraints: minimum, maximum, exclusiveMinimum, exclusiveMaximum, multipleOf, enum
  - Array constraints: minItems, maxItems, uniqueItems, items
  - Object constraints: required, properties, minProperties, maxProperties, patternProperties

## Documentation

- [Changes and Improvements](docs/CHANGES.md) - Recent updates and enhancements
- [Developer Guide](docs/DEVELOPERS.md) - Technical details on implementation

## Installation

```bash
git clone https://github.com/tridentsx/oas2kcl.git
cd oas2kcl
go build
```

## Usage

```bash
# Basic usage
go run main.go -input=<input-file> -output=<output-directory>

# Generate validator schemas
go run main.go -input=<input-file> -output=<output-directory> -validator

# Specify package name
go run main.go -input=<input-file> -output=<output-directory> -package=<package-name>
```

### Command-line Options

- `-input`: Path to the input schema file (OpenAPI or JSON Schema) (required)
- `-output`: Directory to output the generated KCL schemas (default: "output")
- `-package`: Name of the KCL package (default: "schema")
- `-validator`: Generate validator schemas (default: false)

## Testing

The repository includes comprehensive test cases for various constraint types:

```bash
# Run comprehensive tests
./run_comprehensive_tests.sh

# Clean up test files
./cleanup.sh
```

## Examples

### Converting a JSON Schema file

```bash
go run main.go -input=examples/string_constraints.json -output=examples/output-string
```

### Converting an OpenAPI specification

```bash
go run main.go -input=examples/openapi.yaml -output=examples/output-openapi
```

### Generating validator schemas

```bash
go run main.go -input=examples/number_constraints.json -output=examples/output-number -validator
```

## Validation

The generated KCL schemas can be used to validate JSON data using the KCL validator:

```bash
kcl vet <json-data-file> <kcl-schema-file> -s <schema-name>
```

## License

MIT
