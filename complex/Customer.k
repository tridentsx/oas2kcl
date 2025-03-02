schema Customer:
    # Phone number in E.164 format
    phoneNumber: str | None
    address: {"street": str | None, "city": str | None, "country": str | None, "postalCode": str | None, "state": str | None} | None
    # Customer email
    email: str
    # Creation timestamp
    createdAt: datetime | None
    # Unique identifier
    id: str
    # Last update timestamp
    updatedAt: datetime | None
    # Customer first name
    firstName: str
    # Customer last name
    lastName: str

    check:
            phoneNumber.match(r"^\+[1-9]\d{1,14}$")
