import regex

schema Schema:
    email: str
    # email validation for email

    check:
        email == None or regex.match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$", email)

        email != None