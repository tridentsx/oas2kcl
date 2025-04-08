schema ArrayValidation:
    simple_array: [str]
    unique_array?: [str]
    typed_array?: [float]
    nested_array?: [ArrayValidation_nested_array_items]

    check:
        all item in typed_array {item >= 0} if typed_array is not Undefined, "typed_array must be at least 0"
        all item in typed_array {item <= 100} if typed_array is not Undefined, "typed_array must be at most 100"


schema ArrayValidation_nested_array_items:
    id: str
    value?: float


