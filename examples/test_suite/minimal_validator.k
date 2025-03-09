schema Validator:
    integerField: int

    check:
        integerField >= 1, "integerField must be at least 1"
        integerField <= 100, "integerField must be at most 100" 