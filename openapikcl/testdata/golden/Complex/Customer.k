schema Customer:
    mixin [BaseObject]
    
    # Customer first name
    firstName: str
    # Customer last name
    lastName: str
    # Customer email
    email: str
    address: Address | None
    # Phone number in E.164 format
    phoneNumber: str | None

    check:
        email.match(r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$")
        phoneNumber.match(r"^\+[1-9]\d{1,14}$")
