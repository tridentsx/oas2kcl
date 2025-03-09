import regex

schema Validator:
    # String fields with constraints
    stringField: str
    patternField?: str
    emailField?: str
    
    # Number fields with constraints
    integerField: int
    exclusiveIntegerField?: int
    multipleOfIntegerField?: int
    enumIntegerField?: int
    
    # Array fields with constraints
    constrainedArray: [str]
    uniqueArray: [str]
    comboArray?: [str]

    check:
        # String validations
        len(stringField) >= 3, "stringField must be at least 3 characters"
        len(stringField) <= 50, "stringField must be at most 50 characters"
        
        patternField == None or regex.match("^[A-Z][a-z]+$", patternField), "patternField must start with uppercase and continue with lowercase"
        
        emailField == None or regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", emailField), "emailField must be a valid email address"
        
        # Number validations
        integerField >= 1, "integerField must be at least 1"
        integerField <= 100, "integerField must be at most 100"
        
        exclusiveIntegerField == None or exclusiveIntegerField >= 1, "exclusiveIntegerField must be at least 1"
        exclusiveIntegerField == None or exclusiveIntegerField < 10, "exclusiveIntegerField must be less than 10"
        
        multipleOfIntegerField == None or multipleOfIntegerField % 5 == 0, "multipleOfIntegerField must be a multiple of 5"
        
        enumIntegerField == None or enumIntegerField in [1, 2, 3, 5, 8, 13], "enumIntegerField must be one of [1, 2, 3, 5, 8, 13]"
        
        # Array validations
        constrainedArray != None, "constrainedArray is required"
        len(constrainedArray) >= 1, "constrainedArray must have at least 1 item"
        len(constrainedArray) <= 5, "constrainedArray must have at most 5 items"
        
        uniqueArray != None, "uniqueArray is required"
        len(uniqueArray) == len({item: None for item in uniqueArray}), "uniqueArray must contain unique items"
        
        comboArray == None or len(comboArray) >= 2, "comboArray must have at least 2 items"
        comboArray == None or len(comboArray) <= 10, "comboArray must have at most 10 items"
        comboArray == None or len(comboArray) == len({item: None for item in comboArray}), "comboArray must contain unique items" 