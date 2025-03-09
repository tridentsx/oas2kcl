import NumberTest as nt

schema Validator:
    simpleInteger: int
    exclusiveInteger: int
    comboInteger: int
    constrainedNumber: float
    exclusiveNumber: float
    multipleOfNumber: float
    comboNumber: float
    constrainedInteger: int
    multipleOfInteger: int
    enumInteger: int
    enumNumber: float
    simpleNumber: float

    check:
        constrainedInteger >= 1, "constrainedInteger must be at least 1"
        constrainedInteger <= 100, "constrainedInteger must be at most 100"
        exclusiveInteger >= 1, "exclusiveInteger must be at least 1"
        exclusiveInteger < 10, "exclusiveInteger must be less than 10"
        comboInteger >= 10, "comboInteger must be at least 10"
        comboInteger <= 100, "comboInteger must be at most 100"
        comboInteger % 5 == 0, "comboInteger must be a multiple of 5"
        comboInteger in [10, 25, 50, 75, 100], "comboInteger must be one of the allowed values"
        constrainedNumber >= 0.5, "constrainedNumber must be at least 0.5"
        constrainedNumber <= 99.5, "constrainedNumber must be at most 99.5"
        exclusiveNumber < 1, "exclusiveNumber must be less than 1"
        exclusiveNumber > 0, "exclusiveNumber must be greater than 0"
        abs(multipleOfNumber / 0.5 - round(multipleOfNumber / 0.5)) < 1e-10, "multipleOfNumber must be a multiple of 0.5"
        enumInteger in [1, 2, 3, 5, 8, 13], "enumInteger must be one of the allowed values"
        enumNumber in [0.5, 1, 1.5, 2, 2.5], "enumNumber must be one of the allowed values"
        multipleOfInteger % 5 == 0, "multipleOfInteger must be a multiple of 5" 