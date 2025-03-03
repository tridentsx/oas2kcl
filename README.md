# oas2kcl

A command-line tool to generate [KCL](https://kcl-lang.io/) schemas from OpenAPI specifications (supports OpenAPI 2.0/Swagger, 3.0, and 3.1).

## Overview

`oas2kcl` converts OpenAPI schemas into KCL (Kubernetes Configuration Language) schema files, enabling validation and configuration management using KCL for API specifications. The tool handles complex OpenAPI features including:

- Schema conversion with appropriate type mapping
- Constraint validation (min/max values, patterns, etc.)
- Schema references and inheritance
- Documentation generation

## Installation

### Prerequisites

- Go 1.23 or higher

### Build from Source

```bash
git clone https://github.com/tridentsx/oas2kcl.git
cd oas2kcl
go build -o oas2kcl cmd/main.go
```

### Using as a Module

```bash
go get github.com/tridentsx/oas2kcl
```

## Usage

### Basic Usage

Convert an OpenAPI specification file to KCL schemas:

```bash
oas2kcl -oas path/to/openapi.json -out schema.k
```

### Command Line Options

```
Usage:
  oas2kcl -oas openapi.json [options]

Options:
  -oas string        Path to the OpenAPI specification file (required)
  -out string        Optional output file for the generated KCL schema (.k)
  -package string    Package name for the generated KCL schema (default "schema")
  -skip-flatten      Skip flattening the OpenAPI spec
  -skip-remote       Skip remote references during flattening
  -max-depth int     Maximum depth for reference resolution (default 100)
```

## Features

- **Multiple OpenAPI versions support**: Compatible with OpenAPI 2.0 (Swagger), 3.0, and 3.1 soon
- **Schema flattening**: Resolves local and remote references
- **Type conversion**: Maps OpenAPI types to KCL types
- **Validation**: Generates KCL validation constraints from OpenAPI schemas
- **Documentation**: Preserves descriptions and examples from OpenAPI documents

## Example

Given an OpenAPI specification with a `Pet` schema:

```json
{
  "components": {
    "schemas": {
      "Pet": {
        "type": "object",
        "properties": {
          "id": {
            "type": "integer",
            "format": "int64",
            "description": "Pet ID"
          },
          "name": {
            "type": "string",
            "description": "Pet name"
          },
          "status": {
            "type": "string",
            "enum": ["available", "pending", "sold"],
            "description": "Pet status in the store"
          }
        },
        "required": ["name"]
      }
    }
  }
}
```

`oas2kcl` will generate a KCL schema like:

```python
# Pet represents a pet in the pet store
schema Pet:
    # Pet ID
    id?: int
    
    # Pet name
    name: str
    
    # Pet status in the store
    status?: "available" | "pending" | "sold"
```

## License

[Apache License 2.0](LICENSE)