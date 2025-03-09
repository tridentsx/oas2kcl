# Array Constraints in KCL Schemas

This directory contains examples of array constraints in KCL schemas, including:

- `minItems`: Enforces a minimum number of items in an array
- `maxItems`: Enforces a maximum number of items in an array
- `uniqueItems`: Requires all items in the array to be unique

## Generated Files

- `ArrayTest.k`: KCL schema with array constraints
- `Array.k`: Generic array validation schema
- `ArrayTestValidator`: Validator schema with constraints in a check block

## Implementation Details

The array constraints are implemented in the KCL schema using the `check` block with the following validations:

- For `minItems`: `len(array) >= minValue`
- For `maxItems`: `len(array) <= maxValue`
- For `uniqueItems`: `len(array) == len(set(array))`

## Array Types with Constraints

The test schema includes several array types with different constraints:

1. `simpleArray`: A basic array without constraints
2. `constrainedArray`: An array with min/max constraints
3. `uniqueArray`: An array that requires unique items
4. `comboArray`: An array with multiple constraints (min, max, and unique)
5. `numberArray`: An array of numbers with constraints
6. `integerArray`: An array of integers with constraints and unique items
7. `nestedArray`: An array containing arrays
8. `objectArray`: An array of objects

## Testing

To validate the generated KCL schema against test data, create JSON test files and use the following command:

```bash
# Generate KCL schema from JSON schema
go run main.go -input=examples/array_constraints.json -output=examples/output-arrays

# Validate the schema against test data
kcl vet test_data.json examples/output-arrays/ArrayTest.k -s ArrayTest
```

### Test Cases

The following test cases can be used to verify array constraints:

1. Valid array with all constraints satisfied
2. Array that violates the minimum items constraint
3. Array that violates the maximum items constraint
4. Array that violates the unique items constraint 