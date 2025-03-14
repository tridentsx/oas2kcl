import regex

schema StringValidation:
    # A schema for testing string validation constraints
    name: str
    # Min length: 3
    # Max length: 50

    email: EmailValidator
    # Format: email
    # Email address string

    pattern_field?: str
    # Regex pattern: ^[A-Z][a-z]+$

    enum_field?: str


    check:
        name != None
        email != None
        name == None or len(name) >= 3
        name == None or len(name) <= 50
        pattern_field == None or regex.match("^[A-Z][a-z]+$", pattern_field)