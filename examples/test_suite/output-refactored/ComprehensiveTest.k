import regex

schema ComprehensiveTest:
    # A schema to test all implemented constraints
    # A string with length constraints
    stringField: str
    # Min length: 3
    # Max length: 50

    # A string with email format
    emailField: str
    # Format: email
    # Email address string

    # An integer with multipleOf constraint
    multipleOfIntegerField: int
    # Multiple of: 5

    # An integer with enum constraint
    enumIntegerField: int
    # Allowed values: 1, 2, 3, 5, 8, 13

    # A number with range constraints
    numberField: float
    # Minimum: 0.5
    # Maximum: 99.5

    # A number with exclusive range constraints
    exclusiveNumberField: float
    # Exclusive minimum: 0

    # A number with multipleOf constraint
    multipleOfNumberField: float
    # Multiple of: 0.5

    # An array with multiple constraints
    comboArray: list[str]
    # Minimum items: 2
    # Maximum items: 10
    # All items must be unique

    # A string with IPv4 format
    ipv4Field: str
    # Format: ipv4
    # IPv4 address string

    # An integer with exclusive range constraints
    exclusiveIntegerField: int
    # Minimum: 1

    # A number with enum constraint
    enumNumberField: float
    # Allowed values: 0.5, 1, 1.5, 2, 2.5

    # A nested object with properties
    nestedObject: dict
    # An array of objects
    objectArray: list
    # Minimum items: 1
    # Maximum items: 5

    # A string with URI format
    uriField: str
    # Format: uri
    # URI string following RFC 3986

    # An array with min/max constraints
    constrainedArray: list[str]
    # Minimum items: 1
    # Maximum items: 5

    # A string with pattern constraint
    patternField: str
    # Regex pattern: ^[A-Z][a-z]+$

    # An integer with range constraints
    integerField: int
    # Minimum: 1
    # Maximum: 100

    # A simple array without constraints
    simpleArray: list[str]
    # An array with unique items
    uniqueArray: list[str]
    # All items must be unique



schema Email:
    """Email address string validation.
    
    Validates strings to ensure they conform to email address format.
    """
    value: str
    
    check:
        value == None or regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", value), "must be a valid email address"


schema Array:
    """Array validation.
    
    Validates arrays to ensure they conform to specified constraints.
    """
    value: list[str]
    
    check:
        value == None or len(value) >= 2, "Array must have at least 2 items"
        value == None or len(value) <= 10, "Array must have at most 10 items"
        value == None or len(value) == len(set(value)), "Array must contain unique items"


schema IPv4Address:
    """IPv4 address string validation.
    
    Validates strings to ensure they conform to IPv4 address format.
    """
    value: str
    
    check:
        value == None or regex.match("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", value), "must be a valid IPv4 address"


schema URI:
    """URI string validation.
    
    Validates strings to ensure they conform to URI format according to RFC 3986.
    """
    value: str
    
    check:
        value == None or regex.match("^(https?|ftp)://[^\\s/$.?#].[^\\s]*$", value), "must be a valid URI"


schema Integer:
    """Integer value validation.
    
    Validates int values to ensure they conform to specified constraints.
    """
    value: int
    
    check:
        value == None or value % 5 == 0, "Value must be a multiple of 5"


schema Number:
    """Number value validation.
    
    Validates float values to ensure they conform to specified constraints.
    """
    value: float
    
    check:
        value == None or value >= 0.5, "Value must be at least 0.5"
        value == None or value <= 99.5, "Value must be at most 99.5"


# Validator schema for ComprehensiveTest
schema ComprehensiveTestValidator:
    self: ComprehensiveTest

    check:
        self.emailField == None or regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", self.emailField), "emailField must be a valid email address"
        self.multipleOfIntegerField == None or self.multipleOfIntegerField % 5 == 0, "multipleOfIntegerField must be a multiple of 5"
        self.enumIntegerField == None or self.enumIntegerField in [1, 2, 3, 5, 8, 13], "enumIntegerField must be one of the allowed values"
        self.numberField == None or self.numberField >= 0.5, "numberField must be at least 0.5"
        self.numberField == None or self.numberField <= 99.5, "numberField must be at most 99.5"
        self.exclusiveNumberField == None or self.exclusiveNumberField > 0, "exclusiveNumberField must be greater than 0"
        self.multipleOfNumberField == None or abs(self.multipleOfNumberField / 0.5 - round(self.multipleOfNumberField / 0.5)) < 1e-10, "multipleOfNumberField must be a multiple of 0.5"
        self.comboArray == None or len(self.comboArray) >= 2, "comboArray must have at least 2 items"
        self.comboArray == None or len(self.comboArray) <= 10, "comboArray must have at most 10 items"
        self.comboArray == None or len(self.comboArray) == len(set(self.comboArray)), "comboArray must contain unique items"
        self.ipv4Field == None or regex.match("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", self.ipv4Field), "ipv4Field must be a valid IPv4 address"
        self.exclusiveIntegerField == None or self.exclusiveIntegerField >= 1, "exclusiveIntegerField must be at least 1"
        self.enumNumberField == None or self.enumNumberField in [0.5, 1, 1.5, 2, 2.5], "enumNumberField must be one of the allowed values"
        self.objectArray == None or len(self.objectArray) >= 1, "objectArray must have at least 1 items"
        self.objectArray == None or len(self.objectArray) <= 5, "objectArray must have at most 5 items"
        self.uriField == None or regex.match("^(https?|ftp)://[^\\s/$.?#].[^\\s]*$", self.uriField), "uriField must be a valid URI"
        self.constrainedArray == None or len(self.constrainedArray) >= 1, "constrainedArray must have at least 1 items"
        self.constrainedArray == None or len(self.constrainedArray) <= 5, "constrainedArray must have at most 5 items"
        self.integerField == None or self.integerField >= 1, "integerField must be at least 1"
        self.integerField == None or self.integerField <= 100, "integerField must be at most 100"
        self.uniqueArray == None or len(self.uniqueArray) == len(set(self.uniqueArray)), "uniqueArray must contain unique items"