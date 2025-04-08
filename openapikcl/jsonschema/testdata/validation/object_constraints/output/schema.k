import regex

schema ObjectValidation:
    tags?: [str]
    metadata?: any
    id: str
    name: str
    address?: ObjectValidation_address

    check:
        regex.match(id, r"^[A-Z]{2}\d{4}$") if id is not Undefined, "String must match pattern ^[A-Z]{2}\d{4}$"
        len(name) >= 3 if name is not Undefined, "String must be at least 3 characters"
        len(name) <= 50 if name is not Undefined, "String must be at most 50 characters"


schema ObjectValidation_address:
    street: str
    city: str
    zipCode?: str

    check:
        regex.match(zipCode, r"^\d{5}(-\d{4})?$") if zipCode is not Undefined, "String must match pattern ^\d{5}(-\d{4})?$"


