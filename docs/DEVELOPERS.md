# Developer Guide: JSON Schema to KCL Translation

This document explains how OAS2KCL translates JSON Schema and OpenAPI specifications to KCL (Kubernetes Configuration Language) schemas. It covers the internal architecture, template system, and translation mechanisms.

## Architecture Overview

OAS2KCL follows a modular architecture with the following key components:

1. **Parser**: Reads and parses JSON Schema or OpenAPI specifications
2. **Schema Tree Builder**: Converts parsed schema into a tree structure
3. **Generator**: Translates the schema tree into KCL schemas
4. **Template System**: Provides templates for different schema types
5. **Validator Generator**: Creates validation code for schema constraints

## Schema Tree Structure

The core of the translation process is a tree structure that represents the JSON Schema. Each node in the tree corresponds to a schema or subschema that will be translated into a KCL schema.

### Tree Node Types

The `SchemaTreeNode` represents a single node in the schema tree, with the following key components:

1. **Node Type**: Indicates what kind of schema this represents (object, array, string, etc.)
2. **Schema Name**: The name that will be used for the generated KCL schema
3. **Raw Schema**: The original JSON Schema definition for this node
4. **Parent Node**: Reference to the parent node (null for the root)
5. **Properties**: For object nodes, maps property names to their schema nodes
6. **Items**: For array nodes, references the schema for array items
7. **Ref Target**: For reference nodes, indicates the referenced schema
8. **Sub-Schemas**: For composition nodes (allOf, anyOf, oneOf), contains child schemas
9. **Constraints**: Validation constraints from the original schema
10. **Metadata**: Description, format, title, and default values

### Node Type Hierarchy

The tree supports all JSON Schema types:
- Primitive types: `string`, `number`, `integer`, `boolean`, `null`
- Container types: `object`, `array`
- Composition types: `allOf`, `anyOf`, `oneOf`, `not`, `if`, `then`, `else`
- Reference type: `$ref` references

### Building the Tree

The `BuildSchemaTree` function constructs the tree by recursively processing the JSON Schema:

1. Start with the root schema
2. Check for references ($ref) and composition keywords first
3. Determine the node type based on the "type" property
4. Process properties for object types and items for array types
5. Extract constraints and metadata
6. Link the nodes together to form the tree

This process ensures that:
- Each distinct schema is represented exactly once
- References are properly tracked to prevent circular references
- The tree structure mirrors the logical structure of the JSON Schema

### Example Tree Structure

For a schema like:
```json
{
  "type": "object",
  "properties": {
    "name": { "type": "string", "minLength": 2 },
    "age": { "type": "integer", "minimum": 0 },
    "addresses": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "street": { "type": "string" },
          "city": { "type": "string" }
        }
      }
    }
  }
}
```

The resulting tree would have the structure:
```
PersonSchema (object)
├── name (string with minLength=2)
├── age (integer with minimum=0)
└── addresses (array)
    └── items (object)
        ├── street (string)
        └── city (string)
```

### Handling Special Cases

The tree structure handles all JSON Schema constructs:

1. **References**: Nodes of type `Reference` point to other schemas, allowing schema reuse while preventing circular references
2. **Multiple Types**: When a schema has multiple types (e.g., `"type": ["string", "null"]`), the first type is used
3. **Composition**: Schema composition (allOf, anyOf, oneOf) is represented as special nodes with sub-schemas
4. **Conditional Schemas**: If-then-else constructs are represented as specialized nodes

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

## Generator Implementation

The generator traverses the schema tree and generates KCL schemas:

1. **Tree Construction**: First, build the complete schema tree using `BuildSchemaTree`
2. **Schema Generation**: Recursively traverse the tree to generate KCL schemas
3. **File Management**: Write generated schemas to appropriate files
4. **Reference Resolution**: Track generated schemas to handle references properly

The main process flow is:

```go
func GenerateKCLSchemas(schema map[string]interface{}, outputDir string) error {
    // 1. Build schema tree
    tree, err := BuildSchemaTree(schema, "Root", nil)
    if err != nil {
        return err
    }
    
    // 2. Generate schemas from tree
    generator := NewGenerator(outputDir)
    return generator.GenerateSchemasFromTree(tree)
}
```

The generator maintains a registry of generated schemas to avoid duplicates and handle references:

```go
type Generator struct {
    OutputDir      string
    GeneratedFiles map[string]bool
    SchemaRegistry map[string]string
}
```

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

### Recursive Schema Tree Structure

When translating JSON Schema to KCL, a recursive tree structure should be used for type validation:

1. **Schema-Based Type System**: Each distinct type (including formatted types like `date-time`, `email`, etc.) should have its own dedicated schema with complete validation logic.

2. **Reference Instead of Duplication**: Instead of duplicating validation logic:
   - Arrays should reference their item type's schema: `[EmailSchema]` instead of embedding email validation
   - Object properties should reference their type's schema: `email: EmailSchema` 
   - Complex nested structures create a tree of schema references

3. **Validation Inheritance**: When a schema includes another schema, it inherits all validation logic, creating a proper inheritance hierarchy.

This approach ensures:
- Validation logic is defined exactly once (DRY principle)
- Changes to type validation automatically propagate throughout the schema tree
- The structure mirrors JSON Schema's reference system
- Validation is consistent for a given type regardless of where it appears

Example of the recursive structure:
```
PersonSchema
├── name: StringSchema (minLength=2)
├── email: EmailSchema (format=email)
└── addresses: [AddressSchema]
    └── AddressSchema
        ├── street: StringSchema (minLength=1)
        └── zipCode: ZipCodeSchema (pattern=^\d{5}$)
```

In this structure, every leaf node is a schema with its own validation. The actual array validation only needs to validate array-specific constraints (like minItems, maxItems, uniqueItems), while item-specific validation is handled by the referenced schema.

### Schema File Organization

KCL offers a powerful feature for organizing schemas across multiple files that simplifies the implementation of recursive schema structures:

1. **Multiple Files in Same Directory**: KCL schemas in the same directory can reference each other without explicit imports.

2. **Primitive Type Definitions**: Common validators for primitive types can be separated into dedicated files:
   ```
   schemas/
   ├── primitives/
   │   ├── datetime.k        # Contains DateTimeValidator with datetime/regex imports
   │   ├── email.k           # Contains EmailValidator with regex imports
   │   └── zipcode.k         # Contains ZipCodeValidator with regex imports
   ├── models/
   │   ├── person.k          # References primitives directly without imports
   │   └── address.k         # References primitives directly without imports
   └── api/
       └── endpoints.k       # References models directly
   ```

3. **Import Simplification**: This eliminates the need to manage duplicate imports at the top level of each schema file:
   - Each primitive validator file contains its own needed imports (e.g., `datetime`, `regex`)
   - Main schema files can reference these primitive validators directly
   - No need to track which imports are required for each validator

4. **Implementation Strategy**: When generating schemas:
   - Create one file per primitive type with its specific validation logic
   - Generate main schema files that reference these primitives
   - Place all files in the same output directory to enable automatic references

This organizational approach significantly simplifies the implementation of the recursive schema tree structure and produces cleaner, more maintainable KCL code.

### Dictionary Comprehension for Uniqueness

To validate array uniqueness, KCL doesn't have a `set` function but can use isunique() which returns boolean

```python
isunique(array)
```

This creates a dictionary with items as keys (converting to string to handle non-hashable types), resulting in unique keys.

### Format Validation

Format validation is implemented using custom functions or regex patterns. For example, email validation:

```python
regex.match(email, "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$")
```
where rehgex.match(str, regex) returns a boolean

## Testing

For testing our translation logic, we have a comprehensive test suite that:

1. Generates KCL schemas from test JSON Schema inputs
2. Runs the KCL linter to verify syntactic correctness
3. Tests validation against valid and invalid data samples
4. Ensures constraints are properly enforced

The tests are in the `examples/test_suite` directory, with scripts to automate validation.

## Testing Approach

The testing strategy is designed to validate the correctness of the KCL schemas generated from JSON Schema inputs. Rather than focusing on implementation details, we test that the generated schemas correctly validate JSON data according to the original constraints.

### Key Testing Principles

1. **Functionality over Implementation**: Test the functional correctness of generated schemas, not their exact text output. A schema can be written in multiple ways and still provide the same validation behavior.

2. **Avoid Fragile Text Matching**: Do NOT use exact text matching to validate the output of the generator. Instead, test the generated KCL schema's ability to correctly validate data.

3. **Use Known Good and Bad Objects**: For each schema, prepare:
   - Known good objects that should pass validation
   - Known bad objects that should fail validation for specific reasons

### Recommended Testing Flow

The recommended approach follows these steps:

1. Start with a test JSON Schema
2. Generate a KCL schema from the test JSON Schema
3. Test the KCL schema against known good objects (should pass)
4. Test the KCL schema against known bad objects (should fail)
5. Verify that validation errors match expected constraints

This approach tests what matters - the validation behavior - rather than the specific text formatting or structure of the generated KCL.

### Test Structure

The test framework follows a consistent directory structure:

```
openapikcl/jsonschema/testdata/
├── validation/
│   ├── string_constraints/
│   │   ├── input/
│   │   │   ├── schema.json          // The JSON Schema to convert
│   │   │   ├── valid.json           // Valid object that should pass validation
│   │   │   ├── invalid_reason1.json // Invalid object that should fail validation
│   │   │   └── invalid_reason2.json // Another invalid object
│   │   └── output/                  // Generated KCL schemas (cleared between test runs)
│   ├── number_constraints/
│   │   ├── input/
│   │   │   ├── schema.json
│   │   │   ├── valid.json
│   │   │   └── invalid_*.json
│   │   └── output/
│   ├── array_constraints/
│   │   └── ...
│   └── object_constraints/
│       └── ...
└── ...
```

### How Tests Work

The testing system:

1. Discovers test cases in the `testdata/validation` directory
2. For each test case:
   - Generates KCL schemas from the JSON Schema
   - Validates the "valid.json" (expecting success)
   - Validates each "invalid_*.json" (expecting failure)
   - Uses `kcl vet` to perform the actual validation

### Example Test Implementation

Here's a simplified example of how to implement a test:

```go
func TestSchemaValidation(t *testing.T) {
    // Test directory structure
    testDir := "testdata/validation/string_constraints"
    inputDir := filepath.Join(testDir, "input")
    outputDir := filepath.Join(testDir, "output")
    
    // Clear and recreate output directory
    os.RemoveAll(outputDir)
    os.MkdirAll(outputDir, 0755)
    
    // Read input schema
    schemaBytes, err := os.ReadFile(filepath.Join(inputDir, "schema.json"))
    if err != nil {
        t.Fatalf("Failed to read schema: %v", err)
    }
    
    // Generate KCL schema
    err = GenerateKCLSchemas(schemaBytes, outputDir)
    if err != nil {
        t.Fatalf("Failed to generate KCL schema: %v", err)
    }
    
    // Test valid.json (should pass validation)
    validJSON, _ := os.ReadFile(filepath.Join(inputDir, "valid.json"))
    err = validateWithKCL(outputDir, validJSON)
    if err != nil {
        t.Errorf("Validation failed for valid.json: %v", err)
    }
    
    // Find and test all invalid_*.json files (should fail validation)
    invalidFiles, _ := filepath.Glob(filepath.Join(inputDir, "invalid_*.json"))
    for _, invalidFile := range invalidFiles {
        invalidJSON, _ := os.ReadFile(invalidFile)
        err = validateWithKCL(outputDir, invalidJSON)
        if err == nil {
            t.Errorf("Validation unexpectedly passed for %s", filepath.Base(invalidFile))
        }
    }
}

// Helper function to validate JSON against KCL schema
func validateWithKCL(schemaDir string, jsonData []byte) error {
    // Implementation details for running kcl vet with the schema and JSON data
    // ...
}
```

### Advantages of This Approach

- **Resilience to Implementation Changes**: Tests won't break when the code structure changes.
- **Focuses on Actual Requirements**: Tests what users actually care about (correct validation).
- **Easier Maintenance**: No need to update tests when formatting or style changes.
- **Better Test Coverage**: Ensures the system works end-to-end from input to validation.
- **Multiple Valid Implementations**: Acknowledges that there can be multiple valid ways to generate the same functional KCL schema.

### Adding New Test Cases

To add a new test case:

1. Create a new directory under `testdata/validation/` with a descriptive name
2. Create an `input/` subdirectory
3. Add the following files to the `input/` directory:
   - `schema.json`: The JSON Schema to test
   - `valid.json`: A valid object that should pass validation
   - One or more `invalid_*.json` files: Objects that should fail validation, with descriptive names

For example, to test validation of a schema with nested objects:

```bash
mkdir -p openapikcl/jsonschema/testdata/validation/nested_objects/input
# Create schema.json with nested object definitions
# Create valid.json with a correctly structured object
# Create invalid_wrong_type.json, invalid_missing_nested.json, etc.
```

### Running Tests

To run all validation tests:

```bash
go test ./openapikcl/jsonschema -run TestSchemaValidation
```

To run a specific test case:

```bash
go test ./openapikcl/jsonschema -run TestSchemaValidation/string_constraints
```

### Integration with CI/CD

The validation tests are integrated into the CI/CD pipeline, ensuring that any changes to the code don't break the fundamental functionality of generating valid KCL schemas.

### Best Practices for Test Cases

1. **Test One Constraint Type Per Invalid File**: Each invalid JSON file should test a single validation constraint to make it clear what's failing.

2. **Use Descriptive Filenames**: Name invalid JSON files according to the constraint they violate (e.g., `invalid_minimum.json`, `invalid_pattern.json`).

3. **Test Required Properties**: Always include test cases for required vs. optional properties.

4. **Test Edge Cases**: Include tests for boundary values (e.g., exactly at the minimum, just above the maximum).

5. **Document Expected Behaviors**: Add a README.md to complex test cases explaining what each test verifies.

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