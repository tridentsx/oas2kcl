import regex

schema StringValidation:
    pattern_field?: str
    enum_field?: str
    name: str
    email: str

    check:
        regex.match(pattern_field, r"^[A-Z][a-z]+$") if pattern_field is not Undefined, "String must match pattern ^[A-Z][a-z]+$"
        enum_field in ["option1", "option2", "option3"] if enum_field is not Undefined 
        len(name) >= 3 if name is not Undefined, "String must be at least 3 characters"
        len(name) <= 50 if name is not Undefined, "String must be at most 50 characters"
        regex.match(email, r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$") if email, "Value must be a valid email address"


