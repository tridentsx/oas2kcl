schema ArrayValidation:
    typed_array?: [float]
    nested_array?: [ArrayValidation_nested_array_items]
    simple_array: [str]
    unique_array?: [str]

    check:
        all item in typed_array {item >= 0} if typed_array is not Undefined, "typed_array must be at least 0"
        all item in typed_array {item <= 100} if typed_array is not Undefined, "typed_array must be at most 100"
        len(simple_array) <= 5 if simple_array is not Undefined, "Array must have at most 5 items"
        len(simple_array) >= 1 if simple_array is not Undefined, "Array must have at least 1 items"
        isunique(unique_array) if unique_array is not Undefined, "Array items must be unique"


schema ArrayValidation_nested_array_items:
    id: str
    value?: float


