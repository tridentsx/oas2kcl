# Developer Guide: JSON Schema to KCL Translation

This document explains how OAS2KCL translates JSON Schema and OpenAPI specifications to KCL (Kubernetes Configuration Language) schemas. It covers the internal architecture, template system, and translation mechanisms.

## Architecture Overview

OAS2KCL follows a modular architecture with the following key components:

1. **Parser**: Reads and parses JSON Schema or OpenAPI specifications
2. **Generator**: Translates the parsed schema into KCL schemas
3. **Template System**: Provides templates for different schema types
4. **Validator Generator**: Creates validation code for schema constraints

## JSON Schema to KCL Translation

### Basic Type Mapping

| JSON Schema Type | KCL Type |
|------------------|----------|
| `string`         | `str`    |
| `integer`        | `int`    |
| `number`         | `float`  |
| `boolean`        | `bool`   |
| `array`          | `list`   |
| `object`         | Schema or `dict` |
| `null`           | `None`   |

### Object to Schema Translation

JSON Schema objects are translated to KCL schemas as follows:

1. The schema name is derived from the object's title or a generated name
2. Properties become schema attributes with appropriate types
3. Required properties have no optionality marker, optional properties use the `?` suffix
4. Property descriptions are preserved as comments
5. Nested objects may become nested schemas or inline dictionaries

Example:
```json
{
  "type": "object",
  "title": "Person",
  "properties": {
    "name": {
      "type": "string",
      "description": "The person's name"
    },
    "age": {
      "type": "integer"
    }
  },
  "required": ["name"]
}
```

Becomes:
```python
schema Person:
    # The person's name
    name: str
    age?: int
```

### Template System

The template system is used to generate KCL code for different schema types. Each type has a specialized template:

- `StringTemplate`: Handles string constraints
- `NumberTemplate` and `IntegerTemplate`: Handle numeric constraints
- `ArrayTemplate`: Manages array/list constraints
- `ObjectTemplate`: Deals with object/schema constraints
- `BooleanTemplate`: Manages boolean values

Templates are defined in the `openapikcl/jsonschema/templates` directory, with Go functions that build KCL strings with the appropriate syntax.

#### Template Selection

Templates are selected based on the schema type:
```go
func GetTemplateForType(schema map[string]interface{}, schemaName string) string {
    schemaType, _ := schema["type"].(string)
    
    switch schemaType {
    case "string":
        return templates.GetTemplateForStringType(schema, schemaName)
    case "number", "integer":
        return templates.GetTemplateForNumberType(schema, schemaName, schemaType)
    case "array":
        return templates.GetTemplateForArrayType(schema, schemaName)
    // ...and so on
    }
}
```

### Constraint Translation

#### String Constraints

| JSON Schema Constraint | KCL Implementation |
|------------------------|-------------------|
| `minLength`            | Validation check for string length |
| `maxLength`            | Validation check for string length |
| `pattern`              | Regex pattern validation using the `regex` module |
| `format`               | Format-specific validation (email, date, etc.) |
| `enum`                 | Literal union type or validation check |

Example implementation in `string.go`:
```go
func buildStringValidation(schema map[string]interface{}, attrName string) []string {
    validations := []string{}
    
    if minLength, ok := utils.GetIntValue(schema, "minLength"); ok {
        validations = append(validations, 
            fmt.Sprintf("%s == None or len(%s) >= %d, \"%s must be at least %d characters\"", 
                        attrName, attrName, minLength, attrName, minLength))
    }
    
    // Additional validations...
    
    return validations
}
```

#### Number Constraints

| JSON Schema Constraint | KCL Implementation |
|------------------------|-------------------|
| `minimum`              | Value comparison validation |
| `maximum`              | Value comparison validation |
| `exclusiveMinimum`     | Value comparison with exclusivity |
| `exclusiveMaximum`     | Value comparison with exclusivity |
| `multipleOf`           | Modulo operation validation |
| `enum`                 | Value inclusion validation |

Example implementation in `number.go`:
```go
func buildNumberValidation(schema map[string]interface{}, attrName, numType string) []string {
    validations := []string{}
    
    if minimum, ok := utils.GetFloatValue(schema, "minimum"); ok {
        validations = append(validations, 
            fmt.Sprintf("%s == None or %s >= %g, \"%s must be at least %g\"", 
                        attrName, attrName, minimum, attrName, minimum))
    }
    
    // Additional validations...
    
    return validations
}
```

#### Array Constraints

| JSON Schema Constraint | KCL Implementation |
|------------------------|-------------------|
| `minItems`             | Array length validation |
| `maxItems`             | Array length validation |
| `uniqueItems`          | Dictionary comprehension for uniqueness check |
| `items`                | Type and constraint checking for items |

Example implementation in `array.go`:
```go
func buildArrayValidation(schema map[string]interface{}, attrName string) []string {
    validations := []string{}
    
    if minItems, ok := utils.GetIntValue(schema, "minItems"); ok {
        validations = append(validations, 
            fmt.Sprintf("%s == None or len(%s) >= %d, \"%s must have at least %d items\"", 
                        attrName, attrName, minItems, attrName, minItems))
    }
    
    if uniqueItems, ok := utils.GetBoolValue(schema, "uniqueItems"); ok && uniqueItems {
        validations = append(validations,
            fmt.Sprintf("%s == None or len(%s) == len({str(item): None for item in %s}), \"%s must contain unique items\"",
                        attrName, attrName, attrName, attrName))
    }
    
    // Additional validations...
    
    return validations
}
```

#### Object Constraints

| JSON Schema Constraint | KCL Implementation |
|------------------------|-------------------|
| `required`             | Non-optional fields or validation checks |
| `properties`           | Schema attributes |
| `minProperties`        | Dictionary size validation |
| `maxProperties`        | Dictionary size validation |
| `patternProperties`    | Key pattern validation |
| `additionalProperties` | Additional key validation |

## Validator Schema Generation

Validator schemas implement KCL's validation capabilities to enforce JSON Schema constraints. They follow this pattern:

1. Import the base schema
2. Define a validator schema that inherits from the base schema
3. Add a check block with all constraint validations

Example:
```python
# Base schema
schema Person:
    name?: str
    age?: int

# Validator schema
schema PersonValidator(Person):
    check:
        self.name == None or len(self.name) >= 3, "name must be at least 3 characters"
        self.age == None or self.age >= 0, "age must be non-negative"
```

The validator generation logic is in `openapikcl/jsonschema/validation/validation.go`, which:

1. Walks through the schema properties
2. Identifies constraints for each property
3. Generates appropriate validation expressions
4. Assembles them into a comprehensive validator schema

## Advanced Features

### Reference Handling

JSON Schema references (`$ref`) are resolved by:

1. Identifying the reference target (local or remote)
2. Generating a schema for the reference target
3. Replacing the reference with the appropriate type or import

### Composition Keywords

Composition keywords (`allOf`, `anyOf`, `oneOf`, `not`) are handled by:

1. `allOf`: Merging the schemas together
2. `anyOf`/`oneOf`: Creating type unions or validation logic
3. `not`: Inverting validation constraints

## Implementation Notes

### Dictionary Comprehension for Uniqueness

To validate array uniqueness, KCL doesn't have a `set` function but can use dictionary comprehension:

```python
len(array) == len({str(item): None for item in array})
```

This creates a dictionary with items as keys (converting to string to handle non-hashable types), resulting in unique keys.

### Format Validation

Format validation is implemented using custom functions or regex patterns. For example, email validation:

```python
regex.match(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$', email)
```

## Testing

For testing our translation logic, we have a comprehensive test suite that:

1. Generates KCL schemas from test JSON Schema inputs
2. Runs the KCL linter to verify syntactic correctness
3. Tests validation against valid and invalid data samples
4. Ensures constraints are properly enforced

The tests are in the `examples/test_suite` directory, with scripts to automate validation.

## Custom Extensions

Custom extensions can be added by:

1. Defining a new template in the templates directory
2. Adding a handler function in the generator
3. Updating the template selection logic to recognize the extension

## Contributing

When adding new features, follow this workflow:

1. Add test cases for the new feature
2. Implement the feature in the appropriate template file
3. Update the generator to use the new template
4. Add validation logic for the new constraints
5. Document the new feature in this guide 