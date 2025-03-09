import regex

schema Validator:
    # String fields
    stringField: str
    patternField?: str
    emailField?: str
    uriField?: str
    ipv4Field?: str
    
    # Number fields
    integerField: int
    exclusiveIntegerField?: int
    multipleOfIntegerField?: int
    enumIntegerField?: int
    
    numberField: float
    exclusiveNumberField?: float
    multipleOfNumberField?: float
    enumNumberField?: float
    
    # Array fields
    simpleArray?: [str]
    constrainedArray: [str]
    uniqueArray: [str]
    comboArray?: [str]
    
    # Object fields
    nestedObject: {str:}
    objectArray?: [{str:}]

    check:
        # String validations
        len(stringField) >= 3, "stringField must be at least 3 characters"
        len(stringField) <= 50, "stringField must be at most 50 characters"
        
        patternField == None or regex.match("^[A-Z][a-z]+$", patternField), "patternField must start with uppercase and continue with lowercase"
        
        emailField == None or regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", emailField), "emailField must be a valid email address"
        
        uriField == None or regex.match("^(https?|ftp)://[^\\s/$.?#].[^\\s]*$", uriField), "uriField must be a valid URI"
        
        ipv4Field == None or regex.match("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", ipv4Field), "ipv4Field must be a valid IPv4 address"
        
        # Number validations
        integerField >= 1, "integerField must be at least 1"
        integerField <= 100, "integerField must be at most 100"
        
        exclusiveIntegerField == None or exclusiveIntegerField >= 1, "exclusiveIntegerField must be at least 1"
        exclusiveIntegerField == None or exclusiveIntegerField < 10, "exclusiveIntegerField must be less than 10"
        
        multipleOfIntegerField == None or multipleOfIntegerField % 5 == 0, "multipleOfIntegerField must be a multiple of 5"
        
        enumIntegerField == None or enumIntegerField in [1, 2, 3, 5, 8, 13], "enumIntegerField must be one of [1, 2, 3, 5, 8, 13]"
        
        numberField >= 0.5, "numberField must be at least 0.5"
        numberField <= 99.5, "numberField must be at most 99.5"
        
        exclusiveNumberField == None or exclusiveNumberField > 0, "exclusiveNumberField must be greater than 0"
        exclusiveNumberField == None or exclusiveNumberField < 1, "exclusiveNumberField must be less than 1"
        
        multipleOfNumberField == None or abs(multipleOfNumberField / 0.5 - round(multipleOfNumberField / 0.5)) < 1e-10, "multipleOfNumberField must be a multiple of 0.5"
        
        enumNumberField == None or enumNumberField in [0.5, 1.0, 1.5, 2.0, 2.5], "enumNumberField must be one of [0.5, 1.0, 1.5, 2.0, 2.5]"
        
        # Array validations
        constrainedArray != None, "constrainedArray is required"
        len(constrainedArray) >= 1, "constrainedArray must have at least 1 item"
        len(constrainedArray) <= 5, "constrainedArray must have at most 5 items"
        
        uniqueArray != None, "uniqueArray is required"
        len(uniqueArray) == len({item: None for item in uniqueArray}), "uniqueArray must contain unique items"
        
        comboArray == None or len(comboArray) >= 2, "comboArray must have at least 2 items"
        comboArray == None or len(comboArray) <= 10, "comboArray must have at most 10 items"
        comboArray == None or len(comboArray) == len({item: None for item in comboArray}), "comboArray must contain unique items"
        
        # Object validations
        nestedObject != None, "nestedObject is required"
        nestedObject.name != None, "nestedObject.name is required"
        len(nestedObject.name) >= 1, "nestedObject.name must be at least 1 character"
        nestedObject.age == None or nestedObject.age >= 0, "nestedObject.age must be at least 0"
        
        # Object array validation - using foreach
        objectArray == None or all obj in objectArray {
            obj.name != None
            len(obj.name) >= 1
            obj.age == None or obj.age >= 0
        }, "All objects in objectArray must have a valid name and age" 