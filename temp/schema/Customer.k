import regex

schema Customer:
    # Circular reference to Customer
    address?: dict
    # Creation timestamp
    createdAt?: str
    # Customer email
    email: str
    # Customer first name
    firstName: str
    # Unique identifier
    id: str
    # Customer last name
    lastName: str
    # Phone number in E.164 format
    phoneNumber?: str
    # Last update timestamp
    updatedAt?: str

    check:
        regex.match(phoneNumber, r"^\+[1-9]\d{1,14}$")