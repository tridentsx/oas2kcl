import regex

schema ObjectValidation:
    # A schema for testing object validation constraints
    tags?: [str]
    metadata?: dict
    id: str
    # Regex pattern: ^[A-Z]{2}\d{4}$
    name: str
    # Min length: 3
    # Max length: 50
    address?: dict
    [str, any  in  {}]

    check:
        id != None
        name != None
        id == None or regex.match("^[A-Z]{2}\\d{4}$", id)
        name == None or len(name) >= 3
        name == None or len(name) <= 50

