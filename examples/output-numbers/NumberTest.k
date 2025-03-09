schema NumberTest:
    # A schema to test number constraints
    # A simple integer without constraints
    simpleInteger: int
    # An integer with exclusive min/max constraints
    exclusiveInteger: int
    # Minimum: 1
    # Exclusive maximum: 10

    # An integer with multiple constraints
    comboInteger: int
    # Minimum: 10
    # Maximum: 100
    # Multiple of: 5
    # Allowed values: 10, 25, 50, 75, 100

    # A number with min/max constraints
    constrainedNumber: float
    # Minimum: 0.5
    # Maximum: 99.5

    # A number with exclusive constraints
    exclusiveNumber: float
    # Exclusive maximum: 1
    # Exclusive minimum: 0

    # A number that must be a multiple of 0.5
    multipleOfNumber: float
    # Multiple of: 0.5

    # A number with multiple constraints
    comboNumber: float
    # Minimum: 0
    # Exclusive maximum: 10
    # Multiple of: 0.25

    # An integer with min/max constraints
    constrainedInteger: int
    # Minimum: 1
    # Maximum: 100

    # An integer that must be a multiple of 5
    multipleOfInteger: int
    # Multiple of: 5

    # An integer with allowed values
    enumInteger: int
    # Allowed values: 1, 2, 3, 5, 8, 13

    # A simple floating-point number without constraints
    simpleNumber: float
    # A number with allowed values
    enumNumber: float
    # Allowed values: 0.5, 1, 1.5, 2, 2.5



schema Integer:
    """Integer value validation.
    
    Validates int values to ensure they conform to specified constraints.
    """
    value: int
    
    check:
        value == None or value >= 1, "Value must be at least 1"
        value == None or value < 10, "Value must be less than 10"


schema Number:
    """Number value validation.
    
    Validates float values to ensure they conform to specified constraints.
    """
    value: float
    
    check:
        value == None or value >= 0.5, "Value must be at least 0.5"
        value == None or value <= 99.5, "Value must be at most 99.5"


# Validator schema for NumberTest
schema NumberTestValidator:
    self: NumberTest

    check:
        self.enumInteger == None or self.enumInteger in [1, 2, 3, 5, 8, 13], "enumInteger must be one of the allowed values"
        self.enumNumber == None or self.enumNumber in [0.5, 1, 1.5, 2, 2.5], "enumNumber must be one of the allowed values"
        self.exclusiveInteger == None or self.exclusiveInteger >= 1, "exclusiveInteger must be at least 1"
        self.exclusiveInteger == None or self.exclusiveInteger < 10, "exclusiveInteger must be less than 10"
        self.comboInteger == None or self.comboInteger >= 10, "comboInteger must be at least 10"
        self.comboInteger == None or self.comboInteger <= 100, "comboInteger must be at most 100"
        self.comboInteger == None or self.comboInteger % 5 == 0, "comboInteger must be a multiple of 5"
        self.comboInteger == None or self.comboInteger in [10, 25, 50, 75, 100], "comboInteger must be one of the allowed values"
        self.constrainedNumber == None or self.constrainedNumber >= 0.5, "constrainedNumber must be at least 0.5"
        self.constrainedNumber == None or self.constrainedNumber <= 99.5, "constrainedNumber must be at most 99.5"
        self.exclusiveNumber == None or self.exclusiveNumber < 1, "exclusiveNumber must be less than 1"
        self.exclusiveNumber == None or self.exclusiveNumber > 0, "exclusiveNumber must be greater than 0"
        self.multipleOfNumber == None or abs(self.multipleOfNumber / 0.5 - round(self.multipleOfNumber / 0.5)) < 1e-10, "multipleOfNumber must be a multiple of 0.5"
        self.comboNumber == None or self.comboNumber >= 0, "comboNumber must be at least 0"
        self.comboNumber == None or self.comboNumber < 10, "comboNumber must be less than 10"
        self.comboNumber == None or abs(self.comboNumber / 0.25 - round(self.comboNumber / 0.25)) < 1e-10, "comboNumber must be a multiple of 0.25"
        self.constrainedInteger == None or self.constrainedInteger >= 1, "constrainedInteger must be at least 1"
        self.constrainedInteger == None or self.constrainedInteger <= 100, "constrainedInteger must be at most 100"
        self.multipleOfInteger == None or self.multipleOfInteger % 5 == 0, "multipleOfInteger must be a multiple of 5"