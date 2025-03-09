import math
import NumberTest as nt

# Valid test case that should pass validation
valid_test = nt.NumberTest {
    simpleInteger: 5
    exclusiveInteger: 5
    comboInteger: 25
    constrainedNumber: 50.5
    exclusiveNumber: 0.5
    multipleOfNumber: 2.5
    comboNumber: 7.5
    constrainedInteger: 50
    multipleOfInteger: 15
    enumInteger: 5
    enumNumber: 1.5
    simpleNumber: 3.14
}

# Validate the valid test
valid_test_validator = nt.NumberTestValidator {
    self: valid_test
}

# Invalid test cases
# Violation of minimum constraint
test_min_violation = nt.NumberTest {
    simpleInteger: 5
    exclusiveInteger: 5
    comboInteger: 25
    constrainedNumber: 50.5
    exclusiveNumber: 0.5
    multipleOfNumber: 2.5
    comboNumber: 7.5
    constrainedInteger: 0  # Should be at least 1
    multipleOfInteger: 15
    enumInteger: 5
    enumNumber: 1.5
    simpleNumber: 3.14
}

# This should fail validation
test_min_validator = nt.NumberTestValidator {
    self: test_min_violation
}

# Violation of maximum constraint
test_max_violation = nt.NumberTest {
    simpleInteger: 5
    exclusiveInteger: 5
    comboInteger: 25
    constrainedNumber: 50.5
    exclusiveNumber: 0.5
    multipleOfNumber: 2.5
    comboNumber: 7.5
    constrainedInteger: 101  # Should be at most 100
    multipleOfInteger: 15
    enumInteger: 5
    enumNumber: 1.5
    simpleNumber: 3.14
}

# This should fail validation
test_max_validator = nt.NumberTestValidator {
    self: test_max_violation
}

# Violation of exclusive constraints
test_exclusive_violation = nt.NumberTest {
    simpleInteger: 5
    exclusiveInteger: 10  # Should be less than 10
    comboInteger: 25
    constrainedNumber: 50.5
    exclusiveNumber: 1.0  # Should be less than 1
    multipleOfNumber: 2.5
    comboNumber: 7.5
    constrainedInteger: 50
    multipleOfInteger: 15
    enumInteger: 5
    enumNumber: 1.5
    simpleNumber: 3.14
}

# This should fail validation
test_exclusive_validator = nt.NumberTestValidator {
    self: test_exclusive_violation
}

# Violation of multipleOf constraint
test_multiple_violation = nt.NumberTest {
    simpleInteger: 5
    exclusiveInteger: 5
    comboInteger: 26  # Should be a multiple of 5
    constrainedNumber: 50.5
    exclusiveNumber: 0.5
    multipleOfNumber: 2.5
    comboNumber: 7.5
    constrainedInteger: 50
    multipleOfInteger: 17  # Should be a multiple of 5
    enumInteger: 5
    enumNumber: 1.5
    simpleNumber: 3.14
}

# This should fail validation
test_multiple_validator = nt.NumberTestValidator {
    self: test_multiple_violation
}

# Violation of enum constraint
test_enum_violation = nt.NumberTest {
    simpleInteger: 5
    exclusiveInteger: 5
    comboInteger: 25
    constrainedNumber: 50.5
    exclusiveNumber: 0.5
    multipleOfNumber: 2.5
    comboNumber: 7.5
    constrainedInteger: 50
    multipleOfInteger: 15
    enumInteger: 4  # Should be one of 1, 2, 3, 5, 8, 13
    enumNumber: 1.5
    simpleNumber: 3.14
}

# This should fail validation
test_enum_validator = nt.NumberTestValidator {
    self: test_enum_violation
}

# Violation of float multipleOf constraint
test_float_multiple_violation = nt.NumberTest {
    simpleInteger: 5
    exclusiveInteger: 5
    comboInteger: 25
    constrainedNumber: 50.5
    exclusiveNumber: 0.5
    multipleOfNumber: 2.7  # Should be a multiple of 0.5
    comboNumber: 7.5
    constrainedInteger: 50
    multipleOfInteger: 15
    enumInteger: 5
    enumNumber: 1.5
    simpleNumber: 3.14
}

# This should fail validation
test_float_multiple_validator = nt.NumberTestValidator {
    self: test_float_multiple_violation
} 