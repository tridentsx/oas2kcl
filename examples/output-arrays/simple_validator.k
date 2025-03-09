schema Validator:
    simpleArray?: [str]
    constrainedArray: [str]
    uniqueArray: [str]
    comboArray?: [str]
    numberArray?: [float]
    integerArray?: [int]
    nestedArray?: [[str]]
    objectArray?: [any]

    check:
        # Min/Max constraints
        constrainedArray != None, "constrainedArray is required"
        len(constrainedArray) >= 1, "constrainedArray must have at least 1 item"
        len(constrainedArray) <= 5, "constrainedArray must have at most 5 items"
        
        # Unique items constraint
        uniqueArray != None, "uniqueArray is required"
        len(uniqueArray) == len({item: None for item in uniqueArray}), "uniqueArray must contain unique items"
        
        # Combo constraints
        comboArray == None or len(comboArray) >= 2, "comboArray must have at least 2 items"
        comboArray == None or len(comboArray) <= 10, "comboArray must have at most 10 items"
        comboArray == None or len(comboArray) == len({item: None for item in comboArray}), "comboArray must contain unique items"
        
        # Number array constraints
        numberArray == None or len(numberArray) >= 1, "numberArray must have at least 1 item"
        numberArray == None or len(numberArray) <= 5, "numberArray must have at most 5 items"
        
        # Integer array constraints
        integerArray == None or len(integerArray) >= 1, "integerArray must have at least 1 item"
        integerArray == None or len(integerArray) <= 10, "integerArray must have at most 10 items"
        integerArray == None or len(integerArray) == len({str(item): None for item in integerArray}), "integerArray must contain unique items"
        
        # Nested array constraints
        nestedArray == None or len(nestedArray) >= 1, "nestedArray must have at least 1 item"
        nestedArray == None or len(nestedArray) <= 3, "nestedArray must have at most 3 items"
        
        # Object array constraints
        objectArray == None or len(objectArray) >= 1, "objectArray must have at least 1 item"
        objectArray == None or len(objectArray) <= 5, "objectArray must have at most 5 items" 