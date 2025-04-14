schema NumberValidation:
    number_field: float
    exclusive_range?: float
    multiple_of?: int
    enum_number?: float
    integer_field: int

    check:
        number_field >= 0 if number_field is not Undefined, "number_field must be at least 0"
        number_field <= 1 if number_field is not Undefined, "number_field must be at most 1"
        exclusive_range > 0 if exclusive_range is not Undefined, "exclusive_range must be greater than 0"
        exclusive_range < 10 if exclusive_range is not Undefined, "exclusive_range must be less than 10"
        multiple_of % 5 == 0 if multiple_of is not Undefined, "multiple_of must be a multiple of 5"
        enum_number in [1, 2, 3] if enum_number is not Undefined, "enum_number must be one of: 1, 2, 3"
        integer_field >= 1 if integer_field is not Undefined, "integer_field must be at least 1"
        integer_field <= 100 if integer_field is not Undefined, "integer_field must be at most 100"


