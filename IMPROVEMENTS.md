# JSON Schema to KCL Generator Improvements

The current generator has several limitations when handling complex JSON Schema features. Here are the key areas that need improvement:

## 1. Handling `allOf` with Conditional Validation

- The current `handleAllOf` function doesn't properly extract and process `if/then/else` structures.
- Need to generate proper conditional validation checks using KCL's `if` statements.
- Required fields in conditional blocks should be validated only when conditions are met.

## 2. Handling `oneOf` Validation

- The current implementation doesn't properly identify discriminator properties.
- Need to:
  - Identify common discriminator properties (like `contactMethod`).
  - Generate existence checks for all required fields based on the discriminator value.
  - Generate property validation for each option.
  - Add mutual exclusivity checks when needed (e.g., if `contactMethod == "email"` then `phone` shouldn't exist).

## 3. Better Type Handling

- Improve field declarations to include more specific types when possible.
- Handle nested object types for properties like `configA` or `advancedConfig`.
- Handle array types with proper item type information.

## 4. Escape Sequences in Regex Patterns

- Current implementation has issues with escape sequences in regex patterns.
- Need to properly handle special characters in patterns like email validation.

## 5. Required Properties

- Improved handling of required fields, especially when they depend on conditions.
- Generate proper existence checks based on discriminator values or other conditional logic.

## Example of Proper Output

For reference, we've manually created proper KCL schemas for the advanced and composition test cases that demonstrate all the validations that should be generated:

1. `openapikcl/testdata/jsonschema/advanced/output/UserProfile.k`
2. `openapikcl/testdata/jsonschema/composition/output/CompositionTest.k`

These example files show how the generator should create comprehensive validation rules that match the JSON Schema constraints.

## Implementation Plan

1. Fix the regex escape sequence handling in the code.
2. Enhance the `handleAllOf` function to properly process `if/then/else` structures.
3. Improve the `handleOneOf` function to better identify discriminator properties and generate appropriate validation.
4. Add better support for nested type information.
5. Add comprehensive tests for each JSON Schema feature. 