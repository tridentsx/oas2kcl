# Number Constraints in KCL Schemas

This directory contains examples of number constraints in KCL schemas, including:

- `minimum` and `maximum` for both integers and floats
- `exclusiveMinimum` and `exclusiveMaximum` for both integers and floats
- `multipleOf` for both integers and floats
- `enum` for both integers and floats

## Files

- `NumberTest.k`: The generated KCL schema with number constraints
- `validator.k`: A custom validator schema for testing number constraints
- Test JSON files:
  - `min_violation.json`: Violates the minimum constraint (constrainedInteger: 0)
  - `max_violation.json`: Violates the maximum constraint (constrainedInteger: 101)
  - `multiple_violation.json`: Violates the multipleOf constraint (comboInteger: 26, multipleOfInteger: 17)
  - `float_multiple_violation.json`: Violates the float multipleOf constraint (multipleOfNumber: 2.7)

## Testing

To test the constraints, use the `kcl vet` command:

```bash
# Test minimum constraint
kcl vet min_violation.json validator.k -s Validator

# Test maximum constraint
kcl vet max_violation.json validator.k -s Validator

# Test multipleOf constraint
kcl vet multiple_violation.json validator.k -s Validator

# Test float multipleOf constraint
kcl vet float_multiple_violation.json validator.k -s Validator
```

## Implementation Details

The number constraints are implemented in the KCL schema using the `check` block, which validates that:

- Integer values are within the specified range (minimum/maximum)
- Integer values are multiples of the specified value
- Integer values are in the allowed set of values
- Float values are within the specified range (minimum/maximum)
- Float values are multiples of the specified value
- Float values are in the allowed set of values

For float multipleOf validation, we use the formula:
```
abs(value / multipleOf - round(value / multipleOf)) < 1e-10
```

This accounts for floating-point precision issues when checking if a number is a multiple of another number. 