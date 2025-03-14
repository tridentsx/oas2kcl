import regex

schema ObjectValidation:
    # A schema for testing object validation constraints
    id: str
    # Regex pattern: ^[A-Z]{2}\d{4}$

    name: str
    # Min length: 3
    # Max length: 50

    address?: dict[str, any]

    tags?: [str]

    metadata?: dict


    check:
        id != None
        name != None
        name == None or len(name) >= 3
        name == None or len(name) <= 50
        id == None or regex.match("^[A-Z]{2}\\d{4}$", id)