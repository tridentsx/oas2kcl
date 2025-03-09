schema MinimalValidator:
    integerField: int
    numberField: float
    constrainedArray: [str]
    uniqueArray: [str]

    check:
        # Number validations
        integerField >= 1, "integerField must be at least 1"
        integerField <= 100, "integerField must be at most 100"
        
        numberField >= 0.5, "numberField must be at least 0.5"
        numberField <= 99.5, "numberField must be at most 99.5"
        
        # Array validations
        len(constrainedArray) >= 1, "constrainedArray must have at least 1 item"
        len(constrainedArray) <= 5, "constrainedArray must have at most 5 items"
        
        len(uniqueArray) == len({item: None for item in uniqueArray}), "uniqueArray must contain unique items" 