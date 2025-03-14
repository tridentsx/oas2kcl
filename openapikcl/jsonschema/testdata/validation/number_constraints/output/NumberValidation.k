schema NumberValidation:
    # A schema for testing number validation constraints
    integer_field: int

    number_field: float

    exclusive_range?: float

    multiple_of?: int

    enum_number?: float


    check:
        integer_field != None
        number_field != None