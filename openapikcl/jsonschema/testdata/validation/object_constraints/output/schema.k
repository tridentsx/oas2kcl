import regex

schema ObjectValidation_address:
    street: str
    city: str
    zipCode?: str

    check:
        regex.match(zipCode, r"^\d{5}(-\d{4})?$") if zipCode is not Undefined, "String must match pattern ^\d{5}(-\d{4})?$"


schema ObjectValidation_metadata:
    [...str]: str


schema ObjectValidation:
    id: str
    name: str
    address?: ObjectValidation_address
    tags?: [str]
    metadata?: ObjectValidation_metadata

    check:
        regex.match(id, r"^[A-Z]{2}\d{4}$") if id is not Undefined, "String must match pattern ^[A-Z]{2}\d{4}$"
        len(name) >= 3 if name is not Undefined, "String must be at least 3 characters"
        len(name) <= 50 if name is not Undefined, "String must be at most 50 characters"


