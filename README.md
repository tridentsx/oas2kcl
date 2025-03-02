# oas2kcl
Generate KCL schema from an OpenAPI 3.0 schema

``` bash
openapi-to-kcl/
│── cmd/
│   └── main.go             # Entry point of the CLI application
│── openapikcl/             # Exported module for reuse
│   ├── loader.go           # Loads and validates OpenAPI schema
│   ├── converter.go        # Converts OpenAPI types to KCL types
│   ├── generator.go        # Generates KCL schemas from OpenAPI definitions
│── go.mod                  # Go module file
│── go.sum                  # Dependencies file
│── README.md               # Project documentation
```
