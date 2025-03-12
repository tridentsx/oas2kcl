schema ArrayValidation:
    # A schema for testing array validation constraints
    simple_array: [str]
    # Min items: 1
    # Max items: 5
    unique_array?: [str]
    # Unique items: true
    typed_array?: [float]
    nested_array?: [dict]
    tuple_array?: list

    check:
        # Min items: 3
        # Max items: 3
        simple_array != None
        tuple_array == None or len(tuple_array) >= 3
        tuple_array == None or len(tuple_array) <= 3
        simple_array == None or len(simple_array) >= 1
        simple_array == None or len(simple_array) <= 5
        unique_array == None or len(unique_array) == len({str(item): None for item in unique_array})

