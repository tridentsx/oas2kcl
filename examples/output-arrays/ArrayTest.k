schema ArrayTest:
    # A schema to test array constraints
    # An array of objects
    objectArray: list
    # Minimum items: 1
    # Maximum items: 5

    # A simple array without constraints
    simpleArray: list[str]
    # An array with min/max constraints
    constrainedArray: list[str]
    # Minimum items: 1
    # Maximum items: 5

    # An array with unique items
    uniqueArray: list[str]
    # All items must be unique

    # An array with multiple constraints
    comboArray: list[str]
    # Minimum items: 2
    # Maximum items: 10
    # All items must be unique

    # An array of numbers with constraints
    numberArray: list[float]
    # Minimum items: 1
    # Maximum items: 5

    # An array of integers with constraints
    integerArray: list[int]
    # Minimum items: 1
    # Maximum items: 10
    # All items must be unique

    # An array containing arrays
    nestedArray: list[list[str]]
    # Minimum items: 1
    # Maximum items: 3



schema Array:
    """Array validation.
    
    Validates arrays to ensure they conform to specified constraints.
    """
    value: list[str]
    
    check:
        value == None or len(value) >= 2, "Array must have at least 2 items"
        value == None or len(value) <= 10, "Array must have at most 10 items"
        value == None or len(value) == len(set(value)), "Array must contain unique items"


# Validator schema for ArrayTest
schema ArrayTestValidator:
    self: ArrayTest

    check:
        self.comboArray == None or len(self.comboArray) >= 2, "comboArray must have at least 2 items"
        self.comboArray == None or len(self.comboArray) <= 10, "comboArray must have at most 10 items"
        self.comboArray == None or len(self.comboArray) == len(set(self.comboArray)), "comboArray must contain unique items"
        self.numberArray == None or len(self.numberArray) >= 1, "numberArray must have at least 1 items"
        self.numberArray == None or len(self.numberArray) <= 5, "numberArray must have at most 5 items"
        self.integerArray == None or len(self.integerArray) >= 1, "integerArray must have at least 1 items"
        self.integerArray == None or len(self.integerArray) <= 10, "integerArray must have at most 10 items"
        self.integerArray == None or len(self.integerArray) == len(set(self.integerArray)), "integerArray must contain unique items"
        self.nestedArray == None or len(self.nestedArray) >= 1, "nestedArray must have at least 1 items"
        self.nestedArray == None or len(self.nestedArray) <= 3, "nestedArray must have at most 3 items"
        self.objectArray == None or len(self.objectArray) >= 1, "objectArray must have at least 1 items"
        self.objectArray == None or len(self.objectArray) <= 5, "objectArray must have at most 5 items"
        self.constrainedArray == None or len(self.constrainedArray) >= 1, "constrainedArray must have at least 1 items"
        self.constrainedArray == None or len(self.constrainedArray) <= 5, "constrainedArray must have at most 5 items"
        self.uniqueArray == None or len(self.uniqueArray) == len(set(self.uniqueArray)), "uniqueArray must contain unique items"